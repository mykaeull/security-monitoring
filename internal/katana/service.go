package katana

import (
	"context"
	"encoding/json"
	"math"
	"strings" // 👈 precisar para EqualFold
	"time"

	"security-monitoring/internal/httpxscan"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/katana/pkg/engine/standard"
	"github.com/projectdiscovery/katana/pkg/output"
	"github.com/projectdiscovery/katana/pkg/types"
)

type KatanaService struct {
	src httpxscan.HttpxRepository
	dst KatanaRepository
}

func NewKatanaService(src httpxscan.HttpxRepository, dst KatanaRepository) *KatanaService {
	return &KatanaService{src: src, dst: dst}
}

// (Opcional) Log bonito sem campos pesados (raw/body)
func MarshalWithoutRaw(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var any interface{}
	if err := json.Unmarshal(b, &any); err != nil {
		return nil, err
	}
	any = stripRaw(any)
	return json.MarshalIndent(any, "", "  ")
}

func stripRaw(i interface{}) interface{} {
	switch t := i.(type) {
	case map[string]interface{}:
		delete(t, "raw")
		delete(t, "body")
		for k, v := range t {
			t[k] = stripRaw(v)
		}
		return t
	case []interface{}:
		for idx, v := range t {
			t[idx] = stripRaw(v)
		}
		return t
	default:
		return i
	}
}

// helper: pega header de forma case-insensitive
func getHeaderCI(h map[string]string, name string) string {
	if h == nil {
		return ""
	}
	// tentativas diretas comuns
	if v, ok := h[name]; ok {
		return v
	}
	if v, ok := h[strings.ToLower(name)]; ok {
		return v
	}
	if v, ok := h[strings.Title(strings.ToLower(name))]; ok {
		return v
	}
	// varre com EqualFold
	for k, v := range h {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	return ""
}

func (s *KatanaService) Run(ctx context.Context) error {
	opts := &types.Options{
		MaxDepth:     3,
		FieldScope:   "rdn",
		BodyReadSize: math.MaxInt,
		Timeout:      10,
		Concurrency:  10,
		Parallelism:  10,
		Delay:        0,
		RateLimit:    150,
		Strategy:     "depth-first",
		OnResult: func(res output.Result) {
			// Log sem raw/body (opcional)
			if b, err := MarshalWithoutRaw(res); err == nil {
				gologger.Info().Msgf("Katana result:\n%s", string(b))
			} else {
				gologger.Warning().Msgf("Erro ao serializar resultado: %v", err)
			}

			rec := STRecord{}
			rec.OccurredAt = time.Now().UTC()
			rec.Request.Method = res.Request.Method
			rec.Request.Endpoint = res.Request.URL

			// defaults
			var status int
			var length int64
			var ctype string

			if res.Response != nil {
				status = res.Response.StatusCode
				length = res.Response.ContentLength
				ctype = getHeaderCI(res.Response.Headers, "Content-Type")
			}

			rec.Response.StatusCode = status
			rec.Response.ContentLength = length
			rec.Response.ContentType = ctype

			gologger.Info().Msgf(
				"DEBUG -> status=%d length=%d type=%q url=%s",
				rec.Response.StatusCode, rec.Response.ContentLength, rec.Response.ContentType, rec.Request.Endpoint,
			)

			if err := s.dst.SaveRecords(ctx, []STRecord{rec}); err != nil {
				gologger.Warning().Msgf("Falha ao salvar (%s): %v", rec.Request.Endpoint, err)
			}
		},
	}

	crawlerOptions, err := types.NewCrawlerOptions(opts)
	if err != nil {
		return err
	}
	defer crawlerOptions.Close()

	crawler, err := standard.New(crawlerOptions)
	if err != nil {
		return err
	}
	defer crawler.Close()

	records, err := s.src.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, r := range records {
		if r.URL == "" {
			continue
		}
		if err := crawler.Crawl(r.URL); err != nil {
			gologger.Warning().Msgf("Could not crawl %s: %s", r.URL, err.Error())
		}
	}

	return nil
}
