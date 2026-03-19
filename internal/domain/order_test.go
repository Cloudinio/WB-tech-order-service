package domain

import (
	"errors"
	"testing"
	"time"
)

func validOrder() Order {
	return Order{
		OrderUID:    "order-1",
		TrackNumber: "track-1",
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    "Test User",
			Phone:   "+79999999999",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Lenina 1",
			Region:  "Moscow",
			Email:   "test@example.com",
		},
		Payment: Payment{
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
		Items: []Item{
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

func TestOrderValidate_Success(t *testing.T) {
	order := validOrder()

	err := order.Validate()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestOrderValidate_EmptyOrderUID(t *testing.T) {
	order := validOrder()
	order.OrderUID = ""

	err := order.Validate()
	if !errors.Is(err, ErrEmptyOrderUID) {
		t.Fatalf("expected ErrEmptyOrderUID, got %v", err)
	}
}

func TestOrderValidate_EmptyTrackNumber(t *testing.T) {
	order := validOrder()
	order.TrackNumber = ""

	err := order.Validate()
	if !errors.Is(err, ErrEmptyTrackNumber) {
		t.Fatalf("expected ErrEmptyTrackNumber, got %v", err)
	}
}

func TestOrderValidate_EmptyTransaction(t *testing.T) {
	order := validOrder()
	order.Payment.Transaction = ""

	err := order.Validate()
	if !errors.Is(err, ErrEmptyTransaction) {
		t.Fatalf("expected ErrEmptyTransaction, got %v", err)
	}
}

func TestOrderValidate_EmptyItems(t *testing.T) {
	order := validOrder()
	order.Items = nil

	err := order.Validate()
	if !errors.Is(err, ErrEmptyItems) {
		t.Fatalf("expected ErrEmptyItems, got %v", err)
	}
}