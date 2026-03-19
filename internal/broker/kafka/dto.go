package kafka

import (
	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
	"time"
)

type OrderMessage struct {
	OrderUID          string            `json:"order_uid"`
	TrackNumber       string            `json:"track_number"`
	Entry             string            `json:"entry"`
	Delivery          DeliveryMessage   `json:"delivery"`
	Payment           PaymentMessage    `json:"payment"`
	Items             []ItemMessage     `json:"items"`
	Locale            string            `json:"locale"`
	InternalSignature string            `json:"internal_signature"`
	CustomerID        string            `json:"customer_id"`
	DeliveryService   string            `json:"delivery_service"`
	ShardKey          string            `json:"shardkey"`
	SmID              int               `json:"sm_id"`
	DateCreated       string            `json:"date_created"`
	OofShard          string            `json:"oof_shard"`
}

type DeliveryMessage struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type PaymentMessage struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type ItemMessage struct {
	ChrtID      int64  `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int64  `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func (m OrderMessage) ToDomain() (domain.Order, error) {
	createdAt, err := time.Parse(time.RFC3339, m.DateCreated)
	if err != nil {
		return domain.Order{}, err
	}

	items := make([]domain.Item, 0, len(m.Items))
	for _, item := range m.Items {
		items = append(items, domain.Item{
			ChrtID:      item.ChrtID,
			TrackNumber: item.TrackNumber,
			Price:       item.Price,
			RID:         item.RID,
			Name:        item.Name,
			Sale:        item.Sale,
			Size:        item.Size,
			TotalPrice:  item.TotalPrice,
			NmID:        item.NmID,
			Brand:       item.Brand,
			Status:      item.Status,
		})
	}

	return domain.Order{
		OrderUID:          m.OrderUID,
		TrackNumber:       m.TrackNumber,
		Entry:             m.Entry,
		Delivery: domain.Delivery{
			Name:    m.Delivery.Name,
			Phone:   m.Delivery.Phone,
			Zip:     m.Delivery.Zip,
			City:    m.Delivery.City,
			Address: m.Delivery.Address,
			Region:  m.Delivery.Region,
			Email:   m.Delivery.Email,
		},
		Payment: domain.Payment{
			Transaction:  m.Payment.Transaction,
			RequestID:    m.Payment.RequestID,
			Currency:     m.Payment.Currency,
			Provider:     m.Payment.Provider,
			Amount:       m.Payment.Amount,
			PaymentDT:    m.Payment.PaymentDT,
			Bank:         m.Payment.Bank,
			DeliveryCost: m.Payment.DeliveryCost,
			GoodsTotal:   m.Payment.GoodsTotal,
			CustomFee:    m.Payment.CustomFee,
		},
		Items:             items,
		Locale:            m.Locale,
		InternalSignature: m.InternalSignature,
		CustomerID:        m.CustomerID,
		DeliveryService:   m.DeliveryService,
		ShardKey:          m.ShardKey,
		SmID:              m.SmID,
		DateCreated:       createdAt,
		OofShard:          m.OofShard,
	}, nil
}