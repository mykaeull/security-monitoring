package nuclei

import (
	"context"
	"fmt"
	"log"

	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"security-monitoring/internal/httpxscan"
)

type NucleiService struct {
	repo httpxscan.HttpxRepository
}

func NewNucleiService(repo httpxscan.HttpxRepository) *NucleiService {
	return &NucleiService{repo: repo}
}

func (s *NucleiService) Run(ctx context.Context) error {
	records, err := s.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("get records from db: %w", err)
	}

	var urls []string
	for _, rec := range records {
		urls = append(urls, rec.URL)
	}

	if len(urls) == 0 {
		log.Println("[nuclei] Nenhuma URL encontrada para scan")
		return nil
	}

	ne, err := nuclei.NewNucleiEngineCtx(ctx,
		nuclei.WithTemplateFilters(nuclei.TemplateFilters{Tags: []string{"oast"}}),
	)
	if err != nil {
		return fmt.Errorf("nuclei engine initialization failed: %w", err)
	}
	defer ne.Close()

	ne.LoadTargets(urls, false)

	err = ne.ExecuteWithCallback(func(result *output.ResultEvent) {
		// Imprime os resultados no terminal
		fmt.Printf("URL: %s\n", result.URL)
		fmt.Printf("Template: %s\n", result.Template)
		fmt.Printf("Info: %s\n", result.Info)
		fmt.Printf("Matched: %s\n", result.Matched)
		fmt.Println("=====================================")
	})

	if err != nil {
		return fmt.Errorf("nuclei execution failed: %w", err)
	}

	return nil
}
