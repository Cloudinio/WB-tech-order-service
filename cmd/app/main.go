package main

import (
	"context"
	"log"
	nethttp "net/http"
	transporthttp "github.com/Cloudinio/wb-tech-order-service/internal/transport/http"
	"github.com/Cloudinio/wb-tech-order-service/internal/config"
	"github.com/Cloudinio/wb-tech-order-service/internal/repository/postgres"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	log.Println("postgres connected")

	repo := postgres.NewOrderRepository(pool)
	handler := transporthttp.NewHandler(repo)
	router := transporthttp.NewRouter(handler)

	addr := ":" + cfg.AppPort
	log.Printf("http server started on %s", addr)

	if err := nethttp.ListenAndServe(addr, router); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}