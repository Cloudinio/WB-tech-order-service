package domain

import "time"

type Order struct {
	OrderUID          string
	TrackNumber       string
	Entry             string
	Delivery          Delivery
	Payment           Payment
	Items             []Item
	Locale            string
	InternalSignature string
	CustomerID        string
	DeliveryService   string
	ShardKey          string
	SmID              int
	DateCreated       time.Time
	OofShard          string
}

type Delivery struct {
	Name    string
	Phone   string
	Zip     string
	City    string
	Address string
	Region  string
	Email   string
}

type Payment struct {
	Transaction  string
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

type Item struct {
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

func (o Order) Validate() error {
	if o.OrderUID == "" {
		return ErrEmptyOrderUID
	}

	if o.TrackNumber == "" {
		return ErrEmptyTrackNumber
	}

	if o.Payment.Transaction == "" {
		return ErrEmptyTransaction
	}

	if len(o.Items) == 0 {
		return ErrEmptyItems
	}

	return nil
}