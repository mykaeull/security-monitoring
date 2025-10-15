// internal/nuclei/service.go
package nuclei

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"security-monitoring/internal/httpxscan"

	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
)

type NucleiService struct {
	httpxRepo  httpxscan.HttpxRepository
	nucleiRepo NucleiRepository
}

func NewNucleiService(httpxRepo httpxscan.HttpxRepository, nucleiRepo NucleiRepository) *NucleiService {
	return &NucleiService{
		httpxRepo:  httpxRepo,
		nucleiRepo: nucleiRepo,
	}
}

func (s *NucleiService) Run(ctx context.Context) error {
	// ✅ Engine sem filtro de template → vai rodar TODOS os templates disponíveis
	ne, err := nuclei.NewNucleiEngineCtx(ctx)
	if err != nil {
		return fmt.Errorf("criar engine: %w", err)
	}
	defer ne.Close()

	// Carregar URLs
	records, err := s.httpxRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("buscar records: %w", err)
	}

	var urls []string
	for _, r := range records {
		if r.URL != "" {
			urls = append(urls, r.URL)
		}
	}
	if len(urls) == 0 {
		return nil
	}

	ne.LoadTargets(urls, false)

	// Callback mínimo e seguro
	err = ne.ExecuteWithCallback(func(res *output.ResultEvent) {
		defer func() {
			if r := recover(); r != nil {
				_ = r
				_ = debug.Stack()
			}
		}()

		if res == nil {
			return
		}

		var rec STRecord
		rec.TemplateID = res.TemplateID
		rec.OccurredAt = res.Timestamp
		rec.Info.Name = res.Info.Name
		rec.Info.Description = res.Info.Description

		refs := res.Info.Reference.ToSlice()
		if refs == nil {
			refs = []string{}
		}
		rec.Info.Reference = refs

		rec.Info.Severity = res.Info.SeverityHolder.Severity.String()

		var cves []string
		func() {
			defer func() { _ = recover() }()
			cves = res.Info.Classification.CVEID.ToSlice()
		}()
		if cves == nil {
			cves = []string{}
		}
		rec.Info.Classification.CVEID = cves

		var cwes []string
		func() {
			defer func() { _ = recover() }()
			cwes = res.Info.Classification.CWEID.ToSlice()
		}()
		if cwes == nil {
			cwes = []string{}
		}
		rec.Info.Classification.CWEID = cwes

		rec.Type = res.Type
		rec.Host = res.Host
		rec.Port = fmt.Sprint(res.Port)
		rec.Scheme = res.Scheme
		rec.URL = res.URL
		rec.MatchedAt = res.Matched
		rec.IP = res.IP
		rec.CurlCommand = res.CURLCommand

		if res.Interaction != nil {
			rec.Interaction.Protocol = res.Interaction.Protocol
			rec.Interaction.UniqueID = res.Interaction.UniqueID
			rec.Interaction.FullID = res.Interaction.FullId
			rec.Interaction.QType = res.Interaction.QType
			rec.Interaction.RawRequest = res.Interaction.RawRequest
			rec.Interaction.RawResponse = res.Interaction.RawResponse
			rec.Interaction.RemoteAddress = res.Interaction.RemoteAddress
			rec.Interaction.Timestamp = res.Interaction.Timestamp
		}

		_ = json.Marshal
		_ = s.nucleiRepo.SaveRecords(ctx, []STRecord{rec})
	})

	if err != nil {
		return fmt.Errorf("executar nuclei: %w", err)
	}

	return nil
}
