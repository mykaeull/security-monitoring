package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	appdb "security-monitoring/internal/db"
	"security-monitoring/internal/domain"
	"security-monitoring/internal/httpxscan"
	"security-monitoring/internal/katana"
	"security-monitoring/internal/monitor"
	"security-monitoring/internal/nuclei"
	"security-monitoring/internal/subdomainenum"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool := appdb.MustConnect(ctx)
	defer pool.Close()

	// --- Domain feature ---
	domainRepo := domain.NewPGRepository(pool)
	domainSvc := domain.NewService(domainRepo)

	// --- SubdomainEnum feature ---
	subRepo := subdomainenum.NewPGRepository(pool)
	subSvc := subdomainenum.NewService(subRepo)

	// --- httpx feature ---
	httpxRepo := httpxscan.NewPGHttpxRepository(pool)
	httpxSvc := httpxscan.NewService(subRepo, httpxRepo)

	// --- katana feature ---
	katanaRepo := katana.NewPGRepository(pool)
	katanaSvc := katana.NewKatanaService(httpxRepo, katanaRepo)

	// --- Monitor feature ---
	monitorSvc := monitor.NewService(domainSvc)

	// --- Nuclei feature ---
	nucleiRepo := nuclei.NewPGRepository(pool)
	nucleiSvc := nuclei.NewNucleiService(httpxRepo, nucleiRepo)

	// --- Router ---
	mux := http.NewServeMux()

	domain.RegisterRoutes(mux, domainSvc)
	subdomainenum.RegisterRoutes(mux, subSvc)
	httpxscan.RegisterRoutes(mux, httpxSvc)
	monitor.RegisterRoutes(mux, monitorSvc)
	katana.RegisterRoutes(mux, katanaSvc)
	nuclei.RegisterRoutes(mux, nucleiSvc)

	log.Println("Servidor rodando em http://localhost:3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatal(err)
	}
}
