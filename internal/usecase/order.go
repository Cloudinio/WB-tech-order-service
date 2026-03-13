package usecase

import (
	"context"
	"log"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
)

type OrderService struct {
	repo  OrderRepository
	cache OrderCache
}

func NewOrderService(repo OrderRepository, cache OrderCache) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

func (s *OrderService) GetByUID(ctx context.Context, orderUID string) (domain.Order, error) {
	if order, ok := s.cache.Get(orderUID); ok {
		log.Printf("cache hit: %s", orderUID)
		return order, nil
	}

	log.Printf("cache miss: %s", orderUID)

	order, err := s.repo.GetByUID(ctx, orderUID)
	if err != nil {
		return domain.Order{}, err
	}

	s.cache.Set(orderUID, order)
	return order, nil
}