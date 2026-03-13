package main

import (
	"context"
	"log"
	nethttp "net/http"
	"github.com/Cloudinio/wb-tech-order-service/internal/cache/memory"
	"github.com/Cloudinio/wb-tech-order-service/internal/config"
	"github.com/Cloudinio/wb-tech-order-service/internal/repository/postgres"
	transporthttp "github.com/Cloudinio/wb-tech-order-service/internal/transport/http"
	"github.com/Cloudinio/wb-tech-order-service/internal/usecase"
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

	repo := postgres.NewOrderRepository(pool)
	cache := memory.NewOrderCache()

	if err := warmupCache(ctx, repo, cache, 100); err != nil {
		log.Fatalf("warmup cache: %v", err)
	}

	service := usecase.NewOrderService(repo, cache)
	handler := transporthttp.NewHandler(service)
	router := transporthttp.NewRouter(handler)

	addr := ":" + cfg.AppPort
	log.Printf("http server started on %s", addr)

	if err := nethttp.ListenAndServe(addr, router); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}

func warmupCache(ctx context.Context, repo *postgres.OrderRepository, cache *memory.OrderCache, batchSize int) error {
	offset := 0

	for {
		orders, err := repo.ListRecent(ctx, batchSize, offset)
		if err != nil {
			return err
		}

		if len(orders) == 0 {
			break
		}

		cache.Warmup(orders)
		log.Printf("cache warmup: loaded %d orders", len(orders))

		offset += batchSize
	}

	return nil
}