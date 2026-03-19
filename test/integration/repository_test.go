package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Cloudinio/wb-tech-order-service/internal/repository/postgres"
)

func TestOrderRepository_SaveAndGetByUID(t *testing.T) {
	pool, repo := newTestRepo(t)

	orderUID := uniqueID("repo-save-get")
	order := integrationOrder(orderUID, time.Now())

	cleanupOrder(t, pool, orderUID)
	t.Cleanup(func() { cleanupOrder(t, pool, orderUID) })

	ctx := context.Background()

	if err := repo.Save(ctx, order); err != nil {
		t.Fatalf("save order: %v", err)
	}

	got, err := repo.GetByUID(ctx, orderUID)
	if err != nil {
		t.Fatalf("get order by uid: %v", err)
	}

	if got.OrderUID != order.OrderUID {
		t.Fatalf("expected order uid %q, got %q", order.OrderUID, got.OrderUID)
	}

	if got.TrackNumber != order.TrackNumber {
		t.Fatalf("expected track number %q, got %q", order.TrackNumber, got.TrackNumber)
	}

	if got.Payment.Transaction != order.Payment.Transaction {
		t.Fatalf("expected transaction %q, got %q", order.Payment.Transaction, got.Payment.Transaction)
	}

	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got.Items))
	}
}

func TestOrderRepository_GetByUID_NotFound(t *testing.T) {
	_, repo := newTestRepo(t)

	_, err := repo.GetByUID(context.Background(), uniqueID("missing"))
	if !errors.Is(err, postgres.ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestOrderRepository_Save_Duplicate(t *testing.T) {
	pool, repo := newTestRepo(t)

	orderUID := uniqueID("repo-dup")
	order := integrationOrder(orderUID, time.Now())

	cleanupOrder(t, pool, orderUID)
	t.Cleanup(func() { cleanupOrder(t, pool, orderUID) })

	ctx := context.Background()

	if err := repo.Save(ctx, order); err != nil {
		t.Fatalf("first save: %v", err)
	}

	err := repo.Save(ctx, order)
	if !errors.Is(err, postgres.ErrOrderDuplicated) {
		t.Fatalf("expected ErrOrderDuplicated, got %v", err)
	}
}

func TestOrderRepository_ListRecent(t *testing.T) {
	pool, repo := newTestRepo(t)

	oldUID := uniqueID("recent-old")
	newUID := uniqueID("recent-new")

	oldOrder := integrationOrder(oldUID, time.Now().Add(-2*time.Hour))
	newOrder := integrationOrder(newUID, time.Now())

	cleanupOrder(t, pool, oldUID)
	cleanupOrder(t, pool, newUID)
	t.Cleanup(func() {
		cleanupOrder(t, pool, oldUID)
		cleanupOrder(t, pool, newUID)
	})

	ctx := context.Background()

	if err := repo.Save(ctx, oldOrder); err != nil {
		t.Fatalf("save old order: %v", err)
	}
	if err := repo.Save(ctx, newOrder); err != nil {
		t.Fatalf("save new order: %v", err)
	}

	orders, err := repo.ListRecent(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list recent: %v", err)
	}

	foundOld := -1
	foundNew := -1
	for i, o := range orders {
		if o.OrderUID == oldUID {
			foundOld = i
		}
		if o.OrderUID == newUID {
			foundNew = i
		}
	}

	if foundOld == -1 || foundNew == -1 {
		t.Fatalf("expected both test orders in result, got old=%d new=%d", foundOld, foundNew)
	}

	if foundNew > foundOld {
		t.Fatalf("expected newer order before older order, got new=%d old=%d", foundNew, foundOld)
	}
}