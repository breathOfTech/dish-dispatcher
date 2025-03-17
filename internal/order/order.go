package order

import (
	"fmt"
	"time"
)

// Temperature type for order temperature
type Temperature string

// Temperature constants
const (
	Hot    Temperature = "hot"
	Cold   Temperature = "cold"
	Frozen Temperature = "frozen"
)

// Order represents a food order in the system
type Order struct {
	ID        string
	Name      string
	Temp      Temperature
	ShelfLife float64 // in seconds
	DecayRate float64
	CreatedAt time.Time

	// Runtime tracking
	PlacedOnShelfAt  time.Time
	PlacedOnOverflow time.Time
	CurrentShelfType string
	WastedAt         time.Time
	DeliveredAt      time.Time
}

func NewOrder(name string, temp Temperature, shelfLife float64, decayRate float64) *Order {
	return &Order{
		ID:        fmt.Sprintf("%s-%d", name, time.Now().UnixNano()),
		Name:      name,
		Temp:      temp,
		ShelfLife: shelfLife,
		DecayRate: decayRate,
		CreatedAt: time.Now(),
	}
}

// Value decay formula: (shelf_life - decay_rate * elapsedTime) / shelf_life
func (o *Order) CalculateValue(now time.Time) float64 {
	// If the order hasn't been placed on a shelf yet, its value is 1.0
	if o.PlacedOnShelfAt.IsZero() {
		return 1.0
	}

	var elapsedTime float64
	var decayAmount float64

	if o.PlacedOnOverflow.IsZero() {
		// Order is on a primary shelf
		elapsedTime = now.Sub(o.PlacedOnShelfAt).Seconds()
		decayAmount = o.DecayRate * elapsedTime
	} else {
		// Order is on the overflow shelf
		elapsedTimePrimary := o.PlacedOnOverflow.Sub(o.PlacedOnShelfAt).Seconds()
		decayAmountPrimary := o.DecayRate * elapsedTimePrimary

		elapsedTimeOverflow := now.Sub(o.PlacedOnOverflow).Seconds()
		decayAmountOverflow := o.DecayRate * elapsedTimeOverflow // Assuming the same decay rate on overflow

		decayAmount = decayAmountPrimary + decayAmountOverflow
	}

	remainingShelfLife := o.ShelfLife - decayAmount

	if remainingShelfLife <= 0 {
		return 0.0
	}

	return remainingShelfLife / o.ShelfLife
}

func (o *Order) CalculateValueV2(now time.Time) float64 {
	// If the order hasn't been placed on a shelf yet, its value is 1.0
	if o.PlacedOnShelfAt.IsZero() {
		return 1.0
	}

	orderAge := now.Sub(o.PlacedOnShelfAt).Seconds()
	shelfDecayModifier := 1.0 // Default for primary shelves

	if !o.PlacedOnOverflow.IsZero() {
		// Order is on overflow
		overflowAge := now.Sub(o.PlacedOnOverflow).Seconds()
		primaryAge := o.PlacedOnOverflow.Sub(o.PlacedOnShelfAt).Seconds()
		orderAge = primaryAge + overflowAge
		shelfDecayModifier = 2.0 // Overflow decay modifier
	}

	decayAmount := o.DecayRate * orderAge * shelfDecayModifier
	remainingShelfLife := o.ShelfLife - orderAge - decayAmount

	if remainingShelfLife <= 0 {
		return 0.0
	}

	return remainingShelfLife / o.ShelfLife
}

func (o *Order) IsExpired(now time.Time) bool {
	return o.CalculateValue(now) <= 0
}

func (o *Order) String() string {
	return fmt.Sprintf("Order{ID: %s, Name: %s, Temp: %s, Value: %.2f}",
		o.ID, o.Name, o.Temp, o.CalculateValue(time.Now()))
}
