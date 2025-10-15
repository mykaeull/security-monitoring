package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type HostnameSource interface {
	GetAllDomains(ctx context.Context) ([]string, error)
}

type Service struct {
	src HostnameSource
}

func NewService(src HostnameSource) *Service {
	return &Service{src: src}
}

// Request representa a entrada opcional de domínios.
// Se Domains vier vazio, será usado s.src.GetAllDomains(ctx).

// Run mantém compatibilidade com a assinatura original.
func (s *Service) Run(ctx context.Context) ([]byte, error) {
	return s.RunWith(ctx, Request{})
}

func (s *Service) RunWith(ctx context.Context, req Request) ([]byte, error) {
	var (
		hosts []string
		err   error
	)

	// Fonte dos domínios
	if len(req.Domains) > 0 {
		hosts = req.Domains
	} else {
		hosts, err = s.src.GetAllDomains(ctx)
		if err != nil {
			return nil, fmt.Errorf("get domains: %w", err)
		}
	}

	// Normaliza (lower/trim) e remove duplicados
	hosts = uniqLowerTrim(hosts)
	if len(hosts) == 0 {
		return []byte(`{"records":[],"meta":{}}`), nil
	}

	// Chave da API
	apiKey := os.Getenv("SECURITYTRAILS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SECURITYTRAILS_API_KEY não configurada no ambiente")
	}

	// Janela de tempo (últimas 2h)
	timestamp := time.Now().Add(-1 * 2 * time.Hour).Unix()

	// Endpoint e headers fixos
	const url = "https://api.securitytrails.com/v1/search/list"
	headers := map[string]string{
		"Content-Type": "application/json",
		"APIKEY":       apiKey,
	}

	// --------- Requisições em lotes de 300 -----------
	const batchSize = 300
	type apiResp struct {
		Records []json.RawMessage `json:"records"`
		Meta    map[string]any    `json:"meta"`
		Extra   map[string]any    `json:"-"` // ignorado; placeholder se quiser capturar mais tarde
	}

	var (
		allRecords []json.RawMessage
		batches    = 0
	)

	for i := 0; i < len(hosts); i += batchSize {
		end := i + batchSize
		if end > len(hosts) {
			end = len(hosts)
		}
		batch := hosts[i:end]
		batches++

		formatted := quoteJoin(batch)
		payloadMap := map[string]any{
			"query": fmt.Sprintf("(apex_domain IN (%s)) AND first_seen > %d", formatted, timestamp),
		}
		payload, err := json.Marshal(payloadMap)
		if err != nil {
			return nil, fmt.Errorf("marshal body (lote %d): %w", batches, err)
		}

		// Contexto por requisição (60s) herdando o ctx pai
		reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("criar request (lote %d): %w", batches, err)
		}
		for k, v := range headers {
			httpReq.Header.Set(k, v)
		}

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("executar request (lote %d): %w", batches, err)
		}
		func() {
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				err = fmt.Errorf("SecurityTrails status=%d body=%s", resp.StatusCode, string(respBody))
				return
			}

			var parsed apiResp
			if uerr := json.Unmarshal(respBody, &parsed); uerr != nil {
				err = fmt.Errorf("unmarshal resposta (lote %d): %w", batches, uerr)
				return
			}

			// Acumula todos os records
			if len(parsed.Records) > 0 {
				allRecords = append(allRecords, parsed.Records...)
			}
		}()
		if err != nil {
			return nil, err
		}
	}

	// Monta a resposta final mesclada
	out := map[string]any{
		"records": allRecords,
		"meta": map[string]any{
			"batches": batches,
			"total":   len(allRecords),
			// Dica: se quiser, adicione "since_unix": timestamp
		},
	}

	final, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal resposta final: %w", err)
	}
	return final, nil
}

func quoteJoin(items []string) string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		x := strings.ToLower(strings.TrimSpace(it))
		if x == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%q", x))
	}
	return strings.Join(out, ", ")
}

func uniqLowerTrim(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	res := make([]string, 0, len(items))
	for _, it := range items {
		x := strings.ToLower(strings.TrimSpace(it))
		if x == "" {
			continue
		}
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		res = append(res, x)
	}
	return res
}
