package httpxscan

import (
	"context"
	"fmt"
	"strings"
	"log"

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
	src HostnameSource
	// você pode parametrizar threads, método, etc. no futuro
}

func NewService(src HostnameSource) *Service {
	return &Service{src: src}
}

// Run busca hostnames no banco, roda o httpx e imprime os resultados no terminal.
func (s *Service) Run(ctx context.Context) error {
	// 1) Carrega hostnames do banco
	hosts, err := s.src.GetAllHostnames(ctx)
	if err != nil {
		return fmt.Errorf("get hostnames: %w", err)
	}
	if len(hosts) == 0 {
		log.Println("[httpx] Nenhum hostname encontrado na tabela subdomain")
		return nil
	}

	// 2) Configura verbosidade (opcional)
	gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)

	// 3) Monta opções do httpx
	opts := runner.Options{
		Methods:         "GET",
		InputTargetHost: goflags.StringSlice(hosts),
		Threads:         10, // ajuste se quiser
		HttpApiEndpoint: "localhost:31234", // conforme seu pedido
		TechDetect: true,
		OnResult: func(r runner.Result) {
			if r.Err != nil {
				// fmt.Printf("[Err] %s: %s\n", r.Input, r.Err)
				return
			}
			// fmt.Printf("[OK] %s -> %s (%d)\n", r.Input, r.Host, r.StatusCode)
			fmt.Printf("Input: %s\n", r.Input)
			fmt.Printf("Host: %s\n", r.Host)
			fmt.Printf("Status: %d\n", r.StatusCode)
			fmt.Printf("Title: %s\n", r.Title)
			fmt.Printf("FinalURL: %s\n", r.FinalURL)
			fmt.Printf("Location: %s\n", r.Location)
			fmt.Printf("URL: %s\n", r.URL)

			// Technologies pode vir vazia; cheque antes de indexar
			if len(r.Technologies) > 0 {
				// Juntar array em uma string legível
				fmt.Printf("Technologies: %s\n", strings.Join(r.Technologies, ", "))
			} else {
				fmt.Println("Technologies: -")
			}

    		fmt.Println("===============================================")
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
