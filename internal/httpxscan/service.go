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

// Fonte de dados: qualquer componente que consiga listar hostnames
type HostnameSource interface {
	GetAllHostnames(ctx context.Context) ([]string, error)
}

type Service struct {
	src  HostnameSource
	repo HttpxRepository
	// Você pode expor configs aqui (threads, timeout, etc.) se quiser
}

func NewService(src HostnameSource, repo HttpxRepository) *Service {
	return &Service{src: src, repo: repo}
}

// Run busca hostnames no banco e roda o httpx salvando os resultados na tabela httpx.
func (s *Service) Run(ctx context.Context) error {
	// 1) Carrega hostnames do banco
	hosts, err := s.src.GetAllHostnames(ctx)
	if err != nil {
		return fmt.Errorf("get hostnames: %w", err)
	}
	if len(hosts) == 0 {
		// nada a fazer
		return nil
	}

	// 2) Configura verbosidade (opcional)
	gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)

	// 3) Monta opções do httpx
	opts := runner.Options{
		Methods:         "GET",
		InputTargetHost: goflags.StringSlice(hosts),
		Threads:         10,
		HttpApiEndpoint: "localhost:31234",
		TechDetect: true,
		OnResult: func(r runner.Result) {
			// Ignora entradas com erro, conforme seu requisito
			if r.Err != nil {
				return
			}
			rec := HttpxRecord{
				Host:         r.Host,
				Status:       strconv.Itoa(r.StatusCode), // sua coluna é TEXT
				Title:        r.Title,
				Location:     r.Location,
				URL:          r.URL,
				Technologies: r.Technologies,
			}
			// Salva cada resultado individualmente (simples). No futuro, podemos acumular e chamar SaveResults.
			_ = s.repo.SaveResult(ctx, rec)
		},
	}

	if err := opts.ValidateOptions(); err != nil {
		return fmt.Errorf("validate httpx options: %w", err)
	}

	// 4) Cria runner e executa
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
