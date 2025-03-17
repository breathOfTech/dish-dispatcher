package shelf_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"dish-dispatcher/internal/order"
	shelf "dish-dispatcher/internal/shelves"
)

func TestShelfManager_PlaceOrder(t *testing.T) {
	sm := shelf.NewShelfManager(1, 1, 1, 2)
	order1 := &order.Order{ID: "1", Temp: order.Hot}
	order2 := &order.Order{ID: "2", Temp: order.Cold}
	order3 := &order.Order{ID: "3", Temp: order.Frozen}
	order4 := &order.Order{ID: "4", Temp: order.Hot}
	order5 := &order.Order{ID: "5", Temp: order.Hot}

	assert.True(t, sm.PlaceOrder(order1))
	assert.True(t, sm.PlaceOrder(order2))
	assert.True(t, sm.PlaceOrder(order3))
	assert.True(t, sm.PlaceOrder(order4)) // Should go to overflow
	assert.True(t, sm.PlaceOrder(order5)) // Should also go to overflow

	// Overflow full, this order should be wasted
	order6 := &order.Order{ID: "6", Temp: order.Hot}
	assert.False(t, sm.PlaceOrder(order6))

	assert.Equal(t, 6, sm.TotalOrdersReceived)
	assert.Equal(t, 1, sm.TotalOrdersWasted)
}

func TestShelfManager_DeliverOrder(t *testing.T) {
	sm := shelf.NewShelfManager(2, 2, 2, 2)
	order1 := &order.Order{ID: "1", Temp: order.Hot}
	order2 := &order.Order{ID: "2", Temp: order.Cold}
	order3 := &order.Order{ID: "3", Temp: order.Frozen}

	sm.PlaceOrder(order1)
	sm.PlaceOrder(order2)
	sm.PlaceOrder(order3)

	assert.True(t, sm.DeliverOrder("1"))
	assert.True(t, sm.DeliverOrder("2"))
	assert.True(t, sm.DeliverOrder("3"))
	assert.False(t, sm.DeliverOrder("4")) // Non-existent order

	assert.Equal(t, 3, sm.TotalOrdersDelivered)
}

func TestShelfManager_OverflowHandling(t *testing.T) {
	sm := shelf.NewShelfManager(1, 1, 1, 1)
	order1 := &order.Order{ID: "1", Temp: order.Hot}
	order2 := &order.Order{ID: "2", Temp: order.Hot}
	order3 := &order.Order{ID: "3", Temp: order.Hot}

	assert.True(t, sm.PlaceOrder(order1))  // Goes to hot shelf
	assert.True(t, sm.PlaceOrder(order2))  // Goes to overflow
	assert.False(t, sm.PlaceOrder(order3)) // Should be wasted

	assert.Equal(t, 3, sm.TotalOrdersReceived)
	assert.Equal(t, 1, sm.TotalOrdersWasted)
}

func TestShelfManager_GetShelfForTemperature(t *testing.T) {
	sm := shelf.NewShelfManager(1, 1, 1, 1)
	assert.Equal(t, sm.HotShelf, sm.GetShelfForTemperature(order.Hot))
	assert.Equal(t, sm.ColdShelf, sm.GetShelfForTemperature(order.Cold))
	assert.Equal(t, sm.FrozenShelf, sm.GetShelfForTemperature(order.Frozen))
	assert.Nil(t, sm.GetShelfForTemperature(order.Temperature("invalid")))
}
