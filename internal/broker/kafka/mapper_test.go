package kafka

import (
	"testing"
	"time"
)

func validOrderMessage() OrderMessage {
	return OrderMessage{
		OrderUID:    "order-1",
		TrackNumber: "track-1",
		Entry:       "WBIL",
		Delivery: DeliveryMessage{
			Name:    "Test User",
			Phone:   "+79999999999",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Lenina 1",
			Region:  "Moscow",
			Email:   "test@example.com",
		},
		Payment: PaymentMessage{
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
		Items: []ItemMessage{
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
		DateCreated:       "2021-11-26T06:22:19Z",
		OofShard:          "1",
	}
}

func TestOrderMessageToDomain_Success(t *testing.T) {
	msg := validOrderMessage()

	order, err := msg.ToDomain()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if order.OrderUID != msg.OrderUID {
		t.Fatalf("expected order uid %q, got %q", msg.OrderUID, order.OrderUID)
	}

	if order.Payment.Transaction != msg.Payment.Transaction {
		t.Fatalf("expected transaction %q, got %q", msg.Payment.Transaction, order.Payment.Transaction)
	}

	if len(order.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(order.Items))
	}

	expectedTime, _ := time.Parse(time.RFC3339, msg.DateCreated)
	if !order.DateCreated.Equal(expectedTime) {
		t.Fatalf("expected date %v, got %v", expectedTime, order.DateCreated)
	}
}

func TestOrderMessageToDomain_InvalidDate(t *testing.T) {
	msg := validOrderMessage()
	msg.DateCreated = "not-a-date"

	_, err := msg.ToDomain()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}