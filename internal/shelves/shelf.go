package shelf

import (
	"sync"
	"time"

	"dish-dispatcher/internal/order"
)

type ShelfType string

const (
	HotShelf      ShelfType = "hot"
	ColdShelf     ShelfType = "cold"
	FrozenShelf   ShelfType = "frozen"
	OverflowShelf ShelfType = "overflow"
)

type Shelf struct {
	Type     ShelfType
	Capacity int
	mutex    sync.Mutex
	stats    ShelfStats
	Orders   map[string]*order.Order
}

type ShelfStats struct {
	OrdersAdded     int
	OrdersRemoved   int
	OrdersWasted    int
	OrdersDelivered int
	PeakUsage       int
}

func NewShelf(shelfType ShelfType, capacity int) *Shelf {
	return &Shelf{
		Type:     shelfType,
		Capacity: capacity,
		Orders:   make(map[string]*order.Order),
	}
}
func (s *Shelf) Size() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.Orders)
}

func (s *Shelf) IsFull() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.Orders) >= s.Capacity
}

func (s *Shelf) GetStats() ShelfStats {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.stats
}

func (s *Shelf) MarkOrderDelivered(orderID string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	order, exists := s.Orders[orderID]
	if !exists {
		return false
	}

	delete(s.Orders, orderID)
	order.DeliveredAt = time.Now()
	s.stats.OrdersDelivered++
	s.stats.OrdersRemoved++

	return true
}

func (s *Shelf) RemoveExpiredOrders() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for id, order := range s.Orders {
		if order.IsExpired(now) {
			delete(s.Orders, id)
			order.WastedAt = now
			s.stats.OrdersWasted++
			expiredCount++
		}
	}

	return expiredCount
}

func (s *Shelf) GetAllOrders() []*order.Order {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	orders := make([]*order.Order, 0, len(s.Orders))
	for _, order := range s.Orders {
		orders = append(orders, order)
	}

	return orders
}

func (s *Shelf) GetOrder(orderID string) *order.Order {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.Orders[orderID]
}

func (s *Shelf) RemoveOrder(orderID string) *order.Order {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	order, exists := s.Orders[orderID]
	if !exists {
		return nil
	}

	delete(s.Orders, orderID)
	s.stats.OrdersRemoved++

	return order
}

func (s *Shelf) AddOrder(order *order.Order) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.Orders) >= s.Capacity {
		return false
	}

	// Set placement time if not already set
	if order.PlacedOnShelfAt.IsZero() {
		order.PlacedOnShelfAt = time.Now()
	}

	// Update order current shelf
	order.CurrentShelfType = string(s.Type)

	// If we're moving to overflow shelf, track time
	if s.Type == OverflowShelf {
		// We only care about time on overflow shelf
		// This doesn't reset if the order moves back to a regular shelf
		if order.PlacedOnOverflow.IsZero() {
			order.PlacedOnOverflow = time.Now()
		}
	}

	s.Orders[order.ID] = order
	s.stats.OrdersAdded++

	// Update peak usage
	if len(s.Orders) > s.stats.PeakUsage {
		s.stats.PeakUsage = len(s.Orders)
	}

	return true
}

func (sm *ShelfManager) GetStats() map[string]interface{} {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	return map[string]interface{}{
		"hotShelf": map[string]interface{}{
			"capacity": sm.HotShelf.Capacity,
			"current":  sm.HotShelf.Size(),
			"stats":    sm.HotShelf.GetStats(),
		},
		"coldShelf": map[string]interface{}{
			"capacity": sm.ColdShelf.Capacity,
			"current":  sm.ColdShelf.Size(),
			"stats":    sm.ColdShelf.GetStats(),
		},
		"frozenShelf": map[string]interface{}{
			"capacity": sm.FrozenShelf.Capacity,
			"current":  sm.FrozenShelf.Size(),
			"stats":    sm.FrozenShelf.GetStats(),
		},
		"overflowShelf": map[string]interface{}{
			"capacity": sm.OverflowShelf.Capacity,
			"current":  sm.OverflowShelf.Size(),
			"stats":    sm.OverflowShelf.GetStats(),
		},
		"totalOrders": map[string]interface{}{
			"received":  sm.TotalOrdersReceived,
			"delivered": sm.TotalOrdersDelivered,
			"expired":   sm.TotalOrdersExpired,
			"wasted":    sm.TotalOrdersWasted,
		},
	}
}

func (sm *ShelfManager) GetAllOrders() []*order.Order {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	allOrders := make([]*order.Order, 0)
	allOrders = append(allOrders, sm.HotShelf.GetAllOrders()...)
	allOrders = append(allOrders, sm.ColdShelf.GetAllOrders()...)
	allOrders = append(allOrders, sm.FrozenShelf.GetAllOrders()...)
	allOrders = append(allOrders, sm.OverflowShelf.GetAllOrders()...)

	return allOrders
}

func (sm *ShelfManager) RemoveExpiredOrders() int {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	expiredCount := 0
	expiredCount += sm.HotShelf.RemoveExpiredOrders()
	expiredCount += sm.ColdShelf.RemoveExpiredOrders()
	expiredCount += sm.FrozenShelf.RemoveExpiredOrders()
	expiredCount += sm.OverflowShelf.RemoveExpiredOrders()

	sm.TotalOrdersExpired += expiredCount
	return expiredCount
}
