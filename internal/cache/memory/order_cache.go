package memory

import (
	"sync"
	"github.com/Cloudinio/wb-tech-order-service/internal/domain"
)

type OrderCache struct {
	mu    sync.RWMutex
	items map[string]domain.Order
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		items: make(map[string]domain.Order),
	}
}

func (c *OrderCache) Get(orderUID string) (domain.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, ok := c.items[orderUID]
	return order, ok
}

func (c *OrderCache) Set(orderUID string, order domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[orderUID] = order
}

func (c *OrderCache) Warmup(orders []domain.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, order := range orders {
		c.items[order.OrderUID] = order
	}
}