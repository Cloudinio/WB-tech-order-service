package usecase

import (
	"context"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
)

type OrderRepository interface {
	GetByUID(ctx context.Context, orderUID string) (domain.Order, error)
	ListRecent(ctx context.Context, limit, offset int) ([]domain.Order, error)
	Save(ctx context.Context, order domain.Order) error
}

type OrderCache interface {
	Get(orderUID string) (domain.Order, bool)
	Set(orderUID string, order domain.Order)
	Warmup(orders []domain.Order)
}

type OrderGetter interface {
	GetByUID(ctx context.Context, orderUID string) (domain.Order, error)
}