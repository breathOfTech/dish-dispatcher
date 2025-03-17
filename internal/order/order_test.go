package order_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"dish-dispatcher/internal/order"
)

func TestNewOrder(t *testing.T) {
	o := order.NewOrder("Burger", order.Hot, 300, 0.5)

	assert.NotEmpty(t, o.ID)
	assert.Equal(t, "Burger", o.Name)
	assert.Equal(t, order.Hot, o.Temp)
	assert.Equal(t, 300.0, o.ShelfLife)
	assert.Equal(t, 0.5, o.DecayRate)
	assert.False(t, o.CreatedAt.IsZero())
}

func TestCalculateValue(t *testing.T) {
	o := order.NewOrder("Pizza", order.Hot, 300, 0.5)
	testTime := o.CreatedAt.Add(100 * time.Second)

	o.PlacedOnShelfAt = o.CreatedAt

	value := o.CalculateValue(testTime)
	expectedValue := (300 - (0.5 * 100)) / 300 // (300 - 50) / 300
	assert.InDelta(t, expectedValue, value, 0.01)
}

func TestCalculateValue_OverflowShelf(t *testing.T) {
	o := order.NewOrder("Fries", order.Hot, 300, 0.5)
	o.PlacedOnShelfAt = o.CreatedAt
	o.PlacedOnOverflow = o.CreatedAt.Add(50 * time.Second)
	testTime := o.CreatedAt.Add(150 * time.Second)

	value := o.CalculateValue(testTime)
	expectedValue := (300 - (0.5 * 50) - (0.5 * 100)) / 300 // (300 - 25 - 50) / 300
	assert.InDelta(t, expectedValue, value, 0.01)
}

func TestIsExpired(t *testing.T) {
	o := order.NewOrder("Ice Cream", order.Frozen, 100, 1.0)
	o.PlacedOnShelfAt = o.CreatedAt
	testTime := o.CreatedAt.Add(150 * time.Second)

	assert.True(t, o.IsExpired(testTime))
}

func TestString(t *testing.T) {
	o := order.NewOrder("Salad", order.Cold, 200, 0.2)
	assert.Contains(t, o.String(), "Salad")
	assert.Contains(t, o.String(), "cold")
}
