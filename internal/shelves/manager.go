package shelf

import (
	"sync"
	"time"

	"dish-dispatcher/internal/order"
)

type ShelfManager struct {
	HotShelf      *Shelf
	ColdShelf     *Shelf
	FrozenShelf   *Shelf
	OverflowShelf *Shelf
	mutex         sync.Mutex

	TotalOrdersReceived  int
	TotalOrdersDelivered int
	TotalOrdersExpired   int
	TotalOrdersWasted    int
}

func NewShelfManager(hotCapacity, coldCapacity, frozenCapacity, overflowCapacity int) *ShelfManager {
	return &ShelfManager{
		HotShelf:      NewShelf(HotShelf, hotCapacity),
		ColdShelf:     NewShelf(ColdShelf, coldCapacity),
		FrozenShelf:   NewShelf(FrozenShelf, frozenCapacity),
		OverflowShelf: NewShelf(OverflowShelf, overflowCapacity),
	}
}

func (sm *ShelfManager) GetShelfForTemperature(temp order.Temperature) *Shelf {
	switch temp {
	case order.Hot:
		return sm.HotShelf
	case order.Cold:
		return sm.ColdShelf
	case order.Frozen:
		return sm.FrozenShelf
	default:
		return nil

	}
}

func (sm *ShelfManager) PlaceOrder(order *order.Order) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.TotalOrdersReceived++

	primaryShelf := sm.GetShelfForTemperature(order.Temp)
	if primaryShelf == nil {
		sm.TotalOrdersWasted++
		return false
	}
	if primaryShelf.AddOrder(order) {
		return true
	}
	if sm.OverflowShelf.AddOrder(order) {
		order.PlacedOnOverflow = time.Now()
		return true
	}
	sm.TotalOrdersWasted++
	order.WastedAt = time.Now()
	return false
}

func (sm *ShelfManager) DeliverOrder(orderID string) bool {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Try to find and deliver the order from any shelf
	if sm.deliverFromShelf(sm.HotShelf, orderID) ||
		sm.deliverFromShelf(sm.ColdShelf, orderID) ||
		sm.deliverFromShelf(sm.FrozenShelf, orderID) ||
		sm.deliverFromShelf(sm.OverflowShelf, orderID) {
		sm.TotalOrdersDelivered++
		return true
	}

	return false
}

func (sm *ShelfManager) deliverFromShelf(shelf *Shelf, orderID string) bool {
	if order := shelf.GetOrder(orderID); order != nil {
		return shelf.MarkOrderDelivered(orderID)
	}
	return false
}
