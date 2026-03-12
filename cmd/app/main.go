package main

import (
	"context"
	"log"
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
	log.Println("service started")
}