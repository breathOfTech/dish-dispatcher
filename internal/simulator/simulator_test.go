package simulator

import (
	"os"
	"sync"
	"testing"
	"time"

	"dish-dispatcher/internal/config"
	shelf "dish-dispatcher/internal/shelves"
)

func setupTestSimulator(t *testing.T) *Simulator {
	cfg := &config.Config{
		HotShelfCapacity:    5,
		ColdShelfCapacity:   5,
		FrozenShelfCapacity: 5,
		OverflowCapacity:    10,
		OrdersPerSecond:     1,
		SimulationDuration:  0,
	}

	orders := []OrderData{
		{Name: "Burger", Temp: "hot", ShelfLife: 300, DecayRate: 0.5},
		{Name: "Ice Cream", Temp: "frozen", ShelfLife: 200, DecayRate: 0.2},
	}

	s := &Simulator{
		ShelfManager:     shelf.NewShelfManager(cfg.HotShelfCapacity, cfg.ColdShelfCapacity, cfg.FrozenShelfCapacity, cfg.OverflowCapacity),
		Config:           cfg,
		Orders:           orders,
		stop:             make(chan struct{}),
		deliveryInterval: 500 * time.Millisecond,
		cleanupInterval:  2 * time.Second,
	}

	return s
}

func TestOrderPlacement(t *testing.T) {
	s := setupTestSimulator(t)
	s.createOrderFromList()

	if s.ordersProcessed != 1 {
		t.Errorf("Expected 1 order to be processed, got %d", s.ordersProcessed)
	}

	orders := s.ShelfManager.GetAllOrders()
	if len(orders) == 0 {
		t.Errorf("Expected order to be placed on a shelf, but none found")
	}
}

// func TestOrderDelivery(t *testing.T) {
// 	s := setupTestSimulator(t)
// 	s.createOrderFromList()

// 	s.wg.Add(1)
// 	go func() {
// 		defer s.wg.Done()
// 		s.processDeliveries()
// 	}()

// 	time.Sleep(1 * time.Second)
// 	s.attemptDeliveries()
// 	s.wg.Wait()

// 	ordersAfter := len(s.ShelfManager.GetAllOrders())
// 	if ordersAfter > 0 {
// 		t.Errorf("Expected some orders to be delivered, but %d orders remain", ordersAfter)
// 	}
// }

func TestLoadOrdersFromFile(t *testing.T) {
	data := `[ {"name": "Pizza", "temp": "hot", "shelfLife": 600, "decayRate": 0.3} ]`
	file, err := os.CreateTemp("", "orders.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	if _, err := file.Write([]byte(data)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	file.Close()

	orders, err := loadOrdersFromFile(file.Name()) // Ensure correct function reference
	if err != nil {
		t.Fatalf("Failed to load orders from file: %v", err)
	}

	if len(orders) != 1 || orders[0].Name != "Pizza" {
		t.Errorf("Expected one order with name Pizza, got %+v", orders)
	}
}

func TestSimulator_Run(t *testing.T) {
	s := setupTestSimulator(t)
	s.Config.SimulationDuration = 2 // Set short duration for testing

	// Start the simulator in a separate goroutine
	go func() {
		s.Run()
	}()

	// Wait for the simulation to complete
	time.Sleep(3 * time.Second) // Simulate enough time for the simulation to run

	if s.ordersProcessed == 0 {
		t.Errorf("Expected some orders to be processed, but none were")
	}
}

func TestSimulator_Stop(t *testing.T) {
	s := setupTestSimulator(t)

	// Use a WaitGroup to track when the simulator is done processing
	var wg sync.WaitGroup

	// Start the simulator in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Run() // Run the simulator
	}()

	// Sleep for a short period to give the simulator time to process some orders
	time.Sleep(1 * time.Second)

	// Stop the simulator early
	s.Stop()

	// Wait for the simulator to finish processing
	wg.Wait()

	// Ensure that some orders were processed before stopping
	if s.ordersProcessed == 0 {
		t.Errorf("Expected orders to be processed before stopping, but none were")
	}
}
