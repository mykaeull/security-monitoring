package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	appdb "security-monitoring/internal/db"
	"security-monitoring/internal/subdomainenum"
)

func main() {
	// Carrega variáveis do .env (em produção pode usar env do SO)
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool := appdb.MustConnect(ctx)
	defer pool.Close()

	repo := subdomainenum.NewPGRepository(pool)
	svc := subdomainenum.NewService(repo)

	mux := http.NewServeMux()
	subdomainenum.RegisterRoutes(mux, svc)

	log.Println("Servidor rodando em http://localhost:3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatal(err)
	}
}
