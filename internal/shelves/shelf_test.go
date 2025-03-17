package shelf_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"dish-dispatcher/internal/order"
	shelf "dish-dispatcher/internal/shelves"
)

func TestShelf_AddOrder(t *testing.T) {
	s := shelf.NewShelf(shelf.HotShelf, 2)
	o := order.NewOrder("Burger", order.Hot, 300, 0.5)

	added := s.AddOrder(o)
	assert.True(t, added)
	assert.Equal(t, 1, s.Size())
	assert.Equal(t, "hot", o.CurrentShelfType)
}

func TestShelf_IsFull(t *testing.T) {
	s := shelf.NewShelf(shelf.ColdShelf, 1)
	o1 := order.NewOrder("IceCream", order.Cold, 300, 0.2)
	o2 := order.NewOrder("Juice", order.Cold, 300, 0.2)

	s.AddOrder(o1)
	assert.True(t, s.IsFull())
	assert.False(t, s.AddOrder(o2))
}

func TestShelf_RemoveOrder(t *testing.T) {
	s := shelf.NewShelf(shelf.FrozenShelf, 2)
	o := order.NewOrder("FrozenPizza", order.Frozen, 300, 0.1)

	s.AddOrder(o)
	removed := s.RemoveOrder(o.ID)
	assert.NotNil(t, removed)
	assert.Equal(t, 0, s.Size())
}

func TestShelf_MarkOrderDelivered(t *testing.T) {
	s := shelf.NewShelf(shelf.HotShelf, 2)
	o := order.NewOrder("Pasta", order.Hot, 300, 0.3)

	s.AddOrder(o)
	marked := s.MarkOrderDelivered(o.ID)
	assert.True(t, marked)
	assert.Equal(t, 0, s.Size())
	assert.False(t, o.DeliveredAt.IsZero())
}

func TestShelf_RemoveExpiredOrders(t *testing.T) {
	s := shelf.NewShelf(shelf.OverflowShelf, 2)
	o := order.NewOrder("ExpiredSoup", order.Hot, 1, 1.0)
	o.PlacedOnShelfAt = time.Now().Add(-5 * time.Second) // Expired order

	s.AddOrder(o)
	expiredCount := s.RemoveExpiredOrders()
	assert.Equal(t, 1, expiredCount)
	assert.Equal(t, 0, s.Size())
}

func TestShelf_GetAllOrders(t *testing.T) {
	s := shelf.NewShelf(shelf.ColdShelf, 3)
	o1 := order.NewOrder("Salad", order.Cold, 300, 0.2)
	o2 := order.NewOrder("Soda", order.Cold, 300, 0.2)

	s.AddOrder(o1)
	s.AddOrder(o2)

	orders := s.GetAllOrders()
	assert.Equal(t, 2, len(orders))
}
