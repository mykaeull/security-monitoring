package httpxscan

import (
	"context"
	"fmt"
	"strconv"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/httpx/runner"
)

type HostnameSource interface {
	GetAllHostnames(ctx context.Context) ([]string, error)
}

type Service struct {
	src  HostnameSource
	repo HttpxRepository
}

func NewService(src HostnameSource, repo HttpxRepository) *Service {
	return &Service{src: src, repo: repo}
}

func (s *Service) Run(ctx context.Context) error {
	hosts, err := s.src.GetAllHostnames(ctx)
	if err != nil {
		return fmt.Errorf("get hostnames: %w", err)
	}
	if len(hosts) == 0 {
		return nil
	}

	gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)

	opts := runner.Options{
		Methods:         "GET",
		InputTargetHost: goflags.StringSlice(hosts),
		Threads:         10,
		HttpApiEndpoint: "localhost:31234",
		TechDetect:      true,
		OnResult: func(r runner.Result) {
			if r.Err != nil {
				return
			}
			rec := HttpxRecord{
				Host:         r.Host,
				Status:       strconv.Itoa(r.StatusCode),
				Title:        r.Title,
				Location:     r.Location,
				URL:          r.URL,
				Technologies: r.Technologies,
			}
			_ = s.repo.SaveResult(ctx, rec)
		},
	}

	if err := opts.ValidateOptions(); err != nil {
		return fmt.Errorf("validate httpx options: %w", err)
	}

	httpxRunner, err := runner.New(&opts)
	if err != nil {
		return fmt.Errorf("new httpx runner: %w", err)
	}
	defer httpxRunner.Close()

	httpxRunner.RunEnumeration()
	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]HttpxRecord, error) {
	return s.repo.GetAll(ctx)
}
