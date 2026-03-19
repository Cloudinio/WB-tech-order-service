package main

import (
	"context"
	"errors"
	"log"
	nethttp "net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Cloudinio/wb-tech-order-service/internal/broker/kafka"
	"github.com/Cloudinio/wb-tech-order-service/internal/cache/memory"
	"github.com/Cloudinio/wb-tech-order-service/internal/config"
	"github.com/Cloudinio/wb-tech-order-service/internal/repository/postgres"
	transporthttp "github.com/Cloudinio/wb-tech-order-service/internal/transport/http"
	"github.com/Cloudinio/wb-tech-order-service/internal/usecase"
	"github.com/Cloudinio/wb-tech-order-service/internal/metrics"
)

func main() {
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool, err := postgres.NewPool(rootCtx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	appMetrics := metrics.New()

	repo := postgres.NewOrderRepository(pool)
	cache := memory.NewOrderCache()

	if err := warmupCache(rootCtx, repo, cache, 100); err != nil {
		log.Fatalf("warmup cache: %v", err)
	}

	service := usecase.NewOrderService(repo, cache, appMetrics)
	handler := transporthttp.NewHandler(service)
	router := transporthttp.NewRouter(handler, appMetrics)

	consumer := kafka.NewConsumer(
		cfg.KafkaBrokersList(),
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
		repo,
		cache,
		appMetrics,
	)

	server := &nethttp.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Println("kafka consumer started")
		if err := consumer.Run(rootCtx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("kafka consumer stopped with error: %v", err)
			return
		}

		log.Println("kafka consumer stopped")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Printf("http server started on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			log.Printf("http server stopped with error: %v", err)
			return
		}

		log.Println("http server stopped")
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	sig := <-stop
	log.Printf("shutdown signal received: %v", sig)

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("http server shutdown error: %v", err)
	} else {
		log.Println("http server shutdown completed")
	}

	if err := consumer.Close(); err != nil {
		log.Printf("kafka consumer close error: %v", err)
	} else {
		log.Println("kafka consumer close completed")
	}

	wg.Wait()
	log.Println("application stopped")
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