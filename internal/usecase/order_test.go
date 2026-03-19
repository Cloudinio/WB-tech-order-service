package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
)

type fakeOrderRepository struct {
	order      domain.Order
	err        error
	getByUIDN  int
}

func (f *fakeOrderRepository) GetByUID(ctx context.Context, orderUID string) (domain.Order, error) {
	f.getByUIDN++
	if f.err != nil {
		return domain.Order{}, f.err
	}
	return f.order, nil
}

func (f *fakeOrderRepository) ListRecent(ctx context.Context, limit, offset int) ([]domain.Order, error) {
	return nil, nil
}

func (f *fakeOrderRepository) Save(ctx context.Context, order domain.Order) error {
	return nil
}

type fakeOrderCache struct {
	items map[string]domain.Order
}

func newFakeOrderCache() *fakeOrderCache {
	return &fakeOrderCache{
		items: make(map[string]domain.Order),
	}
}

func (f *fakeOrderCache) Get(orderUID string) (domain.Order, bool) {
	order, ok := f.items[orderUID]
	return order, ok
}

func (f *fakeOrderCache) Set(orderUID string, order domain.Order) {
	f.items[orderUID] = order
}

func (f *fakeOrderCache) Warmup(orders []domain.Order) {
	for _, order := range orders {
		f.items[order.OrderUID] = order
	}
}

func serviceTestOrder() domain.Order {
	return domain.Order{
		OrderUID:    "order-1",
		TrackNumber: "track-1",
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
			Transaction:  "tx-1",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1000,
			PaymentDT:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 100,
			GoodsTotal:   900,
			CustomFee:    0,
		},
		Items: []domain.Item{
			{
				ChrtID:      1,
				TrackNumber: "track-1",
				Price:       1000,
				RID:         "rid-1",
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
		CustomerID:        "customer-1",
		DeliveryService:   "meest",
		ShardKey:          "1",
		SmID:              10,
		DateCreated:       time.Now().UTC(),
		OofShard:          "1",
	}
}

func TestOrderService_GetByUID_CacheHit(t *testing.T) {
	order := serviceTestOrder()

	repo := &fakeOrderRepository{order: order}
	cache := newFakeOrderCache()
	cache.Set(order.OrderUID, order)

	svc := NewOrderService(repo, cache, nil)

	got, err := svc.GetByUID(context.Background(), order.OrderUID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got.OrderUID != order.OrderUID {
		t.Fatalf("expected order uid %q, got %q", order.OrderUID, got.OrderUID)
	}

	if repo.getByUIDN != 0 {
		t.Fatalf("expected repo calls 0, got %d", repo.getByUIDN)
	}
}

func TestOrderService_GetByUID_CacheMiss(t *testing.T) {
	order := serviceTestOrder()

	repo := &fakeOrderRepository{order: order}
	cache := newFakeOrderCache()

	svc := NewOrderService(repo, cache, nil)

	got, err := svc.GetByUID(context.Background(), order.OrderUID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got.OrderUID != order.OrderUID {
		t.Fatalf("expected order uid %q, got %q", order.OrderUID, got.OrderUID)
	}

	if repo.getByUIDN != 1 {
		t.Fatalf("expected repo calls 1, got %d", repo.getByUIDN)
	}

	_, ok := cache.Get(order.OrderUID)
	if !ok {
		t.Fatal("expected order to be stored in cache")
	}
}

func TestOrderService_GetByUID_RepoError(t *testing.T) {
	expectedErr := errors.New("repo error")

	repo := &fakeOrderRepository{err: expectedErr}
	cache := newFakeOrderCache()

	svc := NewOrderService(repo, cache, nil)

	_, err := svc.GetByUID(context.Background(), "order-1")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if repo.getByUIDN != 1 {
		t.Fatalf("expected repo calls 1, got %d", repo.getByUIDN)
	}
}