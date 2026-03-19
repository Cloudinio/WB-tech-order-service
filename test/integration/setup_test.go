package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
	"github.com/Cloudinio/wb-tech-order-service/internal/repository/postgres"
)

func testPostgresDSN() string {
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		return dsn
	}

	return "postgres://orders_user:orders_pass@localhost:5432/orders_db?sslmode=disable"
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(testPostgresDSN())
	if err != nil {
		t.Skipf("skip integration test: bad postgres dsn: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Skipf("skip integration test: cannot create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skip integration test: postgres unavailable: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func cleanupOrder(t *testing.T, pool *pgxpool.Pool, orderUID string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _ = pool.Exec(ctx, `DELETE FROM items WHERE order_uid = $1`, orderUID)
	_, _ = pool.Exec(ctx, `DELETE FROM payments WHERE order_uid = $1`, orderUID)
	_, _ = pool.Exec(ctx, `DELETE FROM deliveries WHERE order_uid = $1`, orderUID)
	_, _ = pool.Exec(ctx, `DELETE FROM orders WHERE order_uid = $1`, orderUID)
}

func integrationOrder(uid string, createdAt time.Time) domain.Order {
	return domain.Order{
		OrderUID:    uid,
		TrackNumber: "track-" + uid,
		Entry:       "WBIL",
		Delivery: domain.Delivery{
			Name:    "Test User",
			Phone:   "+79999999999",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Lenina 1",
			Region:  "Moscow",
			Email:   "test@example.com",
		},
		Payment: domain.Payment{
			Transaction:  "tx-" + uid,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1000,
			PaymentDT:    createdAt.Unix(),
			Bank:         "alpha",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    0,
		},
		Items: []domain.Item{
			{
				ChrtID:      1,
				TrackNumber: "track-" + uid,
				Price:       1000,
				RID:         "rid-" + uid,
				Name:        "Item 1",
				Sale:        0,
				Size:        "L",
				TotalPrice:  1000,
				NmID:        2,
				Brand:       "Brand",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "customer-" + uid,
		DeliveryService:   "meest",
		ShardKey:          "1",
		SmID:              10,
		DateCreated:       createdAt.UTC(),
		OofShard:          "1",
	}
}

func newTestRepo(t *testing.T) (*pgxpool.Pool, *postgres.OrderRepository) {
	t.Helper()

	pool := newTestPool(t)
	repo := postgres.NewOrderRepository(pool)
	return pool, repo
}

func uniqueID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}