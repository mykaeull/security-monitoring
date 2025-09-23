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

func (s *Service) Run(ctx context.Context) ([]byte, error) {
	return s.RunWith(ctx, Request{})
}

func (s *Service) RunWith(ctx context.Context, req Request) ([]byte, error) {
	var (
		hosts []string
		err   error
	)

	if len(req.Domains) > 0 {
		hosts = req.Domains
	} else {
		hosts, err = s.src.GetAllDomains(ctx)
		if err != nil {
			return nil, fmt.Errorf("get domains: %w", err)
		}
	}

	hosts = uniqLowerTrim(hosts)
	if len(hosts) == 0 {
		return []byte(`{"records":[],"meta":{}}`), nil
	}

	apiKey := os.Getenv("SECURITYTRAILS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SECURITYTRAILS_API_KEY não configurada no ambiente")
	}

	formatted := quoteJoin(hosts)

	timestamp := time.Now().Add(-1 * 2 * time.Hour).Unix()

	url := "https://api.securitytrails.com/v1/search/list"

	headers := map[string]string{
		"Content-Type": "application/json",
		"APIKEY":       apiKey,
	}

	payloadMap := map[string]any{
		"query": fmt.Sprintf("(apex_domain IN (%s)) AND first_seen > %d", formatted, timestamp),
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("criar request: %w", err)
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executar request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("SecurityTrails status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func quoteJoin(items []string) string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		x := strings.ToLower(strings.TrimSpace(it))
		if x == "" {
			continue
		}
		out = append(out, fmt.Sprintf("\"%s\"", x))
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
