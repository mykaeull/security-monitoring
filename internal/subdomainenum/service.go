package subdomainenum

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Service interface {
	Enumerate(ctx context.Context, apexDomain string) (status int, body []byte, contentType string, err error)
}

type DefaultService struct {
	httpClient *http.Client
	repo       Repository
}

func NewService(repo Repository) *DefaultService {
	return &DefaultService{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		repo:       repo,
	}
}

func (s *DefaultService) Enumerate(ctx context.Context, apexDomain string) (int, []byte, string, error) {
	apiKey := os.Getenv("SECURITYTRAILS_API_KEY")
	if apiKey == "" {
		return http.StatusInternalServerError, nil, "application/json", fmt.Errorf("missing SECURITYTRAILS_API_KEY")
	}

	// Monta a query no formato esperado pela SecurityTrails
	payload := map[string]string{
		"query": fmt.Sprintf("apex_domain = '%s'", apexDomain),
	}
	b, _ := json.Marshal(payload)

	// Endpoint com query params fixos
	url := fmt.Sprintf("https://api.securitytrails.com/v1/domains/list?apikey=%s&include_ips=true&scroll=true", apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return http.StatusInternalServerError, nil, "application/json", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return http.StatusBadGateway, nil, "application/json", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.StatusBadGateway, nil, "application/json", err
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/json"
	}

	var parsed STListResponse
	if err := json.Unmarshal(respBody, &parsed); err == nil && len(parsed.Records) > 0 {
		_ = s.repo.SaveRecords(ctx, parsed.Records)
	}

	return resp.StatusCode, respBody, ct, nil
}

