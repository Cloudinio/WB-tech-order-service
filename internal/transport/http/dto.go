package http

import (
	"time"
	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
)

type OrderResponse struct {
	OrderUID          string           `json:"order_uid"`
	TrackNumber       string           `json:"track_number"`
	Entry             string           `json:"entry"`
	Delivery          DeliveryResponse `json:"delivery"`
	Payment           PaymentResponse  `json:"payment"`
	Items             []ItemResponse   `json:"items"`
	Locale            string           `json:"locale"`
	InternalSignature string           `json:"internal_signature"`
	CustomerID        string           `json:"customer_id"`
	DeliveryService   string           `json:"delivery_service"`
	ShardKey          string           `json:"shardkey"`
	SmID              int              `json:"sm_id"`
	DateCreated       time.Time        `json:"date_created"`
	OofShard          string           `json:"oof_shard"`
}

type DeliveryResponse struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type PaymentResponse struct {
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

type ItemResponse struct {
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

func NewOrderResponse(order domain.Order) OrderResponse {
	items := make([]ItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, ItemResponse{
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

	return OrderResponse{
		OrderUID:    order.OrderUID,
		TrackNumber: order.TrackNumber,
		Entry:       order.Entry,
		Delivery: DeliveryResponse{
			Name:    order.Delivery.Name,
			Phone:   order.Delivery.Phone,
			Zip:     order.Delivery.Zip,
			City:    order.Delivery.City,
			Address: order.Delivery.Address,
			Region:  order.Delivery.Region,
			Email:   order.Delivery.Email,
		},
		Payment: PaymentResponse{
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
		Items:             items,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		ShardKey:          order.ShardKey,
		SmID:              order.SmID,
		DateCreated:       order.DateCreated,
		OofShard:          order.OofShard,
	}
}