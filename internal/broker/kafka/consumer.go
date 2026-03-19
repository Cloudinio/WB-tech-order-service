package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
	"github.com/Cloudinio/wb-tech-order-service/internal/metrics"
	repopg "github.com/Cloudinio/wb-tech-order-service/internal/repository/postgres"
)

type Consumer struct {
	reader  *kafkago.Reader
	repo    OrderSaver
	cache   OrderCache
	metrics *metrics.Metrics
}

type OrderSaver interface {
	Save(ctx context.Context, order domain.Order) error
}

type OrderCache interface {
	Set(orderUID string, order domain.Order)
}

func NewConsumer(
	brokers []string,
	topic, groupID string,
	repo OrderSaver,
	cache OrderCache,
	metrics *metrics.Metrics,
) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader:  reader,
		repo:    repo,
		cache:   cache,
		metrics: metrics,
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		c.metrics.KafkaMessagesTotal.Inc()

		log.Printf("kafka message received: topic=%s partition=%d offset=%d",
			msg.Topic, msg.Partition, msg.Offset)

		var orderMsg OrderMessage
		if err := json.Unmarshal(msg.Value, &orderMsg); err != nil {
			c.metrics.KafkaInvalidMessagesTotal.Inc()
			log.Printf("invalid kafka message json: %v", err)
			continue
		}

		order, err := orderMsg.ToDomain()
		if err != nil {
			c.metrics.KafkaInvalidMessagesTotal.Inc()
			log.Printf("map kafka message to domain failed: %v", err)
			continue
		}

		if err := order.Validate(); err != nil {
			c.metrics.KafkaInvalidMessagesTotal.Inc()
			log.Printf("order validation failed: %v", err)
			continue
		}

		if err := c.repo.Save(ctx, order); err != nil {
			if errors.Is(err, repopg.ErrOrderDuplicated) {
				c.metrics.KafkaDuplicateMessages.Inc()
				log.Printf("duplicate order ignored: %s", order.OrderUID)
				c.cache.Set(order.OrderUID, order)
				continue
			}

			c.metrics.KafkaSaveErrorsTotal.Inc()
			log.Printf("save order failed: %v", err)
			continue
		}

		c.cache.Set(order.OrderUID, order)
		log.Printf("order saved and cached: %s", order.OrderUID)
	}
}