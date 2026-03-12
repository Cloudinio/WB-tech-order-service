package postgres

import "github.com/Cloudinio/wb-tech-order-service/internal/domain"

import "time"

type orderRow struct {
	OrderUID          string
	TrackNumber       string
	Entry             string
	Locale            string
	InternalSignature string
	CustomerID        string
	DeliveryService   string
	ShardKey          string
	SmID              int
	DateCreated       time.Time
	OofShard          string
}

type deliveryRow struct {
	OrderUID string
	Name     string
	Phone    string
	Zip      string
	City     string
	Address  string
	Region   string
	Email    string
}

type paymentRow struct {
	Transaction  string
	OrderUID     string
	RequestID    string
	Currency     string
	Provider     string
	Amount       int
	PaymentDT    int64
	Bank         string
	DeliveryCost int
	GoodsTotal   int
	CustomFee    int
}

type itemRow struct {
	ID          int64
	OrderUID    string
	ChrtID      int64
	TrackNumber string
	Price       int
	RID         string
	Name        string
	Sale        int
	Size        string
	TotalPrice  int
	NmID        int64
	Brand       string
	Status      int
}

func buildOrder(or orderRow, dr deliveryRow, pr paymentRow, items []itemRow) domain.Order {
	return domain.Order{
		OrderUID:          or.OrderUID,
		TrackNumber:       or.TrackNumber,
		Entry:             or.Entry,
		Locale:            or.Locale,
		InternalSignature: or.InternalSignature,
		CustomerID:        or.CustomerID,
		DeliveryService:   or.DeliveryService,
		ShardKey:          or.ShardKey,
		SmID:              or.SmID,
		DateCreated:       or.DateCreated,
		OofShard:          or.OofShard,
		Delivery: domain.Delivery{
			Name:    dr.Name,
			Phone:   dr.Phone,
			Zip:     dr.Zip,
			City:    dr.City,
			Address: dr.Address,
			Region:  dr.Region,
			Email:   dr.Email,
		},
		Payment: domain.Payment{
			Transaction:  pr.Transaction,
			RequestID:    pr.RequestID,
			Currency:     pr.Currency,
			Provider:     pr.Provider,
			Amount:       pr.Amount,
			PaymentDT:    pr.PaymentDT,
			Bank:         pr.Bank,
			DeliveryCost: pr.DeliveryCost,
			GoodsTotal:   pr.GoodsTotal,
			CustomFee:    pr.CustomFee,
		},
		Items: itemsToDomain(items),
	}
}

func itemsToDomain(items []itemRow) []domain.Item {
	result := make([]domain.Item, 0, len(items))
	for _, it := range items {
		result = append(result, domain.Item{
			ChrtID:      it.ChrtID,
			TrackNumber: it.TrackNumber,
			Price:       it.Price,
			RID:         it.RID,
			Name:        it.Name,
			Sale:        it.Sale,
			Size:        it.Size,
			TotalPrice:  it.TotalPrice,
			NmID:        it.NmID,
			Brand:       it.Brand,
			Status:      it.Status,
		})
	}
	return result
}