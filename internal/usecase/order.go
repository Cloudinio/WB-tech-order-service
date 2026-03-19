package usecase

import (
	"context"

	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
	"github.com/Cloudinio/wb-tech-order-service/internal/metrics"
)

type OrderService struct {
	repo    OrderRepository
	cache   OrderCache
	metrics *metrics.Metrics
}

func NewOrderService(repo OrderRepository, cache OrderCache, metrics *metrics.Metrics) *OrderService {
	return &OrderService{
		repo:    repo,
		cache:   cache,
		metrics: metrics,
	}
}

func (s *OrderService) GetByUID(ctx context.Context, orderUID string) (domain.Order, error) {
	if order, ok := s.cache.Get(orderUID); ok {
		if s.metrics != nil {
			s.metrics.CacheHitsTotal.Inc()
		}
		return order, nil
	}

	if s.metrics != nil {
		s.metrics.CacheMissesTotal.Inc()
	}

	order, err := s.repo.GetByUID(ctx, orderUID)
	if err != nil {
		return domain.Order{}, err
	}

	s.cache.Set(orderUID, order)
	return order, nil
}