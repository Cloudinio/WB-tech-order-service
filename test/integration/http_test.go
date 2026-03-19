package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Cloudinio/wb-tech-order-service/internal/cache/memory"
	"github.com/Cloudinio/wb-tech-order-service/internal/metrics"
	transporthttp "github.com/Cloudinio/wb-tech-order-service/internal/transport/http"
	"github.com/Cloudinio/wb-tech-order-service/internal/usecase"
)

func TestHTTP_GetOrderByUID(t *testing.T) {
	pool, repo := newTestRepo(t)

	orderUID := uniqueID("http-get")
	order := integrationOrder(orderUID, time.Now())

	cleanupOrder(t, pool, orderUID)
	t.Cleanup(func() { cleanupOrder(t, pool, orderUID) })

	ctx := context.Background()
	if err := repo.Save(ctx, order); err != nil {
		t.Fatalf("save order: %v", err)
	}

	cache := memory.NewOrderCache()
	appMetrics := metrics.New()
	service := usecase.NewOrderService(repo, cache, appMetrics)
	handler := transporthttp.NewHandler(service)
	router := transporthttp.NewRouter(handler, appMetrics)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/orders/" + orderUID)
	if err != nil {
		t.Fatalf("http get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var got transporthttp.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.OrderUID != orderUID {
		t.Fatalf("expected order uid %q, got %q", orderUID, got.OrderUID)
	}
}

func TestHTTP_GetOrderByUID_NotFound(t *testing.T) {
	_, repo := newTestRepo(t)

	cache := memory.NewOrderCache()
	appMetrics := metrics.New()
	service := usecase.NewOrderService(repo, cache, appMetrics)
	handler := transporthttp.NewHandler(service)
	router := transporthttp.NewRouter(handler, appMetrics)

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/orders/" + uniqueID("missing-http"))
	if err != nil {
		t.Fatalf("http get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}