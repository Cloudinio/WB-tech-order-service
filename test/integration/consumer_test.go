package integration

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/Cloudinio/wb-tech-order-service/internal/broker/kafka"
	"github.com/Cloudinio/wb-tech-order-service/internal/cache/memory"
	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
	"github.com/Cloudinio/wb-tech-order-service/internal/metrics"
)

func testKafkaBrokers() []string {
	if brokers := os.Getenv("TEST_KAFKA_BROKERS"); brokers != "" {
		return []string{brokers}
	}
	return []string{"localhost:9092"}
}

func waitForOrder(t *testing.T, repo interface {
	GetByUID(ctx context.Context, orderUID string) (domain.Order, error)
}, orderUID string, timeout time.Duration) error {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := repo.GetByUID(context.Background(), orderUID)
		if err == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return context.DeadlineExceeded
}

func TestKafkaConsumer_Run(t *testing.T) {
	pool, repo := newTestRepo(t)

	orderUID := uniqueID("consumer")
	order := integrationOrder(orderUID, time.Now())

	cleanupOrder(t, pool, orderUID)
	t.Cleanup(func() { cleanupOrder(t, pool, orderUID) })

	msg := kafka.OrderMessage{
		OrderUID:    order.OrderUID,
		TrackNumber: order.TrackNumber,
		Entry:       order.Entry,
		Delivery: kafka.DeliveryMessage{
			Name:    order.Delivery.Name,
			Phone:   order.Delivery.Phone,
			Zip:     order.Delivery.Zip,
			City:    order.Delivery.City,
			Address: order.Delivery.Address,
			Region:  order.Delivery.Region,
			Email:   order.Delivery.Email,
		},
		Payment: kafka.PaymentMessage{
			Transaction:  order.Payment.Transaction,
			RequestID:    order.Payment.RequestID,
			Currency:     order.Payment.Currency,
			Provider:     order.Payment.Provider,
			Amount:       order.Payment.Amount,
			PaymentDT:    order.Payment.PaymentDT,
			Bank:         order.Payment.Bank,
			DeliveryCost: order.Payment.DeliveryCost,
			GoodsTotal:   order.Payment.GoodsTotal,
			CustomFee:    order.Payment.CustomFee,
		},
		Items: []kafka.ItemMessage{
			{
				ChrtID:      order.Items[0].ChrtID,
				TrackNumber: order.Items[0].TrackNumber,
				Price:       order.Items[0].Price,
				RID:         order.Items[0].RID,
				Name:        order.Items[0].Name,
				Sale:        order.Items[0].Sale,
				Size:        order.Items[0].Size,
				TotalPrice:  order.Items[0].TotalPrice,
				NmID:        order.Items[0].NmID,
				Brand:       order.Items[0].Brand,
				Status:      order.Items[0].Status,
			},
		},
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		ShardKey:          order.ShardKey,
		SmID:              order.SmID,
		DateCreated:       order.DateCreated.Format(time.RFC3339),
		OofShard:          order.OofShard,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal kafka message: %v", err)
	}

	appMetrics := metrics.New()
	cache := memory.NewOrderCache()
	groupID := uniqueID("test-group")
	topic := "orders"

	consumer := kafka.NewConsumer(
		testKafkaBrokers(),
		topic,
		groupID,
		repo,
		cache,
		appMetrics,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer consumer.Close()

	go func() {
		_ = consumer.Run(ctx)
	}()

	writer := &kafkago.Writer{
		Addr:     kafkago.TCP(testKafkaBrokers()...),
		Topic:    topic,
		Balancer: &kafkago.LeastBytes{},
	}
	defer writer.Close()

	writeCtx, writeCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer writeCancel()

	if err := writer.WriteMessages(writeCtx, kafkago.Message{
		Key:   []byte(orderUID),
		Value: payload,
	}); err != nil {
		t.Skipf("skip integration test: kafka unavailable or topic missing: %v", err)
	}

	if err := waitForOrder(t, repo, orderUID, 10*time.Second); err != nil {
		t.Fatalf("order was not saved by consumer: %v", err)
	}

	got, err := repo.GetByUID(context.Background(), orderUID)
	if err != nil {
		t.Fatalf("get order after consume: %v", err)
	}

	if got.OrderUID != orderUID {
		t.Fatalf("expected order uid %q, got %q", orderUID, got.OrderUID)
	}
}