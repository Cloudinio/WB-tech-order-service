package memory

import (
	"testing"
	"time"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
)

func testOrder(uid string) domain.Order {
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
		DateCreated:       time.Now().UTC(),
		OofShard:          "1",
	}
}

func TestOrderCache_GetMissing(t *testing.T) {
	cache := NewOrderCache()

	_, ok := cache.Get("missing")
	if ok {
		t.Fatal("expected cache miss, got hit")
	}
}

func TestOrderCache_SetAndGet(t *testing.T) {
	cache := NewOrderCache()
	order := testOrder("order-1")

	cache.Set(order.OrderUID, order)

	got, ok := cache.Get(order.OrderUID)
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}

	if got.OrderUID != order.OrderUID {
		t.Fatalf("expected order uid %q, got %q", order.OrderUID, got.OrderUID)
	}
}

func TestOrderCache_Warmup(t *testing.T) {
	cache := NewOrderCache()

	orders := []domain.Order{
		testOrder("order-1"),
		testOrder("order-2"),
	}

	cache.Warmup(orders)

	for _, order := range orders {
		got, ok := cache.Get(order.OrderUID)
		if !ok {
			t.Fatalf("expected cache hit for %q, got miss", order.OrderUID)
		}
		if got.OrderUID != order.OrderUID {
			t.Fatalf("expected order uid %q, got %q", order.OrderUID, got.OrderUID)
		}
	}
}