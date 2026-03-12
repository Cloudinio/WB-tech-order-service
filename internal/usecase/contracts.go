package usecase

import (
	"context"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
)

type OrderRepository interface {
	GetByUID(ctx context.Context, orderUID string) (domain.Order, error)
}