package domain

import (
	"context"
	"strings"
)

type Service interface {
	AddDomains(ctx context.Context, domains []string) (inserted, skipped int, details []string, err error)
	GetAllDomains(ctx context.Context) ([]string, error)
}

type DefaultService struct {
	repo Repository
}

func NewService(repo Repository) *DefaultService {
	return &DefaultService{repo: repo}
}

func (s *DefaultService) AddDomains(ctx context.Context, domains []string) (int, int, []string, error) {
	if domains == nil {
		domains = []string{}
	}

	clean := make([]string, 0, len(domains))
	seen := make(map[string]struct{}, len(domains))

	for _, d := range domains {
		x := strings.ToLower(strings.TrimSpace(d))
		if x == "" {
			continue
		}
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		clean = append(clean, x)
	}

	inserted, skipped := 0, 0
	details := make([]string, 0, len(clean))

	for _, d := range clean {
		ok, err := s.repo.InsertDomain(ctx, d)
		if err != nil {
			return 0, 0, nil, err
		}
		if ok {
			inserted++
			details = append(details, "inserted: "+d)
		} else {
			skipped++
			details = append(details, "skipped (exists): "+d)
		}
	}

	return inserted, skipped, details, nil
}

func (s *DefaultService) GetAllDomains(ctx context.Context) ([]string, error) {
	return s.repo.GetAllDomains(ctx)
}
