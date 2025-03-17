package simulator

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"dish-dispatcher/internal/config"
	"dish-dispatcher/internal/order"
	shelf "dish-dispatcher/internal/shelves"
)

// OrderData represents the structure of orders in the input JSON
type OrderData struct {
	Name      string  `json:"name"`
	Temp      string  `json:"temp"`
	ShelfLife float64 `json:"shelfLife"`
	DecayRate float64 `json:"decayRate"`
}

// Simulator manages the simulation of orders and deliveries
type Simulator struct {
	ShelfManager     *shelf.ShelfManager
	Config           *config.Config
	Orders           []OrderData
	stop             chan struct{}
	wg               sync.WaitGroup
	deliveryInterval time.Duration
	cleanupInterval  time.Duration
	statsMutex       sync.Mutex
	ordersProcessed  int // Track processed orders
	decayModifier    float64
}

// NewSimulator creates a new simulator with the given configuration
func NewSimulator(cfg *config.Config, ordersFile string) (*Simulator, error) {
	// Load orders from JSON file
	orders, err := loadOrdersFromFile(ordersFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load orders: %w", err)
	}

	shelfManager := shelf.NewShelfManager(
		cfg.HotShelfCapacity,
		cfg.ColdShelfCapacity,
		cfg.FrozenShelfCapacity,
		cfg.OverflowCapacity,
	)
	// Ensure decayModifier is set from config
	decayModifier := cfg.DecayModifier

	return &Simulator{
		ShelfManager:     shelfManager,
		Config:           cfg,
		Orders:           orders,
		stop:             make(chan struct{}),
		deliveryInterval: time.Millisecond * 500, // Check for deliveries every 500ms
		cleanupInterval:  time.Millisecond * 500, // Check for expired orders every 500ms
		decayModifier:    decayModifier,
	}, nil
}

// loadOrdersFromFile reads orders from a JSON file
func loadOrdersFromFile(filePath string) ([]OrderData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var orders []OrderData
	if err := json.NewDecoder(file).Decode(&orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *Simulator) generateOrders() {
	defer s.wg.Done()

	// Calculate interval between orders
	interval := time.Duration(1000.0/s.Config.OrdersPerSecond) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// If we still have orders to process
			if s.ordersProcessed < len(s.Orders) {
				s.createOrderFromList()

				// If this was the last order, wait a bit to allow
				// for deliveries and cleanup before stopping
				if s.ordersProcessed >= len(s.Orders) {
					// Give some time for delivery attempts and cleanup
					time.Sleep(10 * time.Second)
					fmt.Println("All orders have been processed!")
					close(s.stop)
				}
			}
		case <-s.stop:
			return
		}
	}
}

// Run starts the simulation
func (s *Simulator) Run() {
	fmt.Println("Starting simulation...")
	fmt.Printf("Configuration: Hot=%d, Cold=%d, Frozen=%d, Overflow=%d, Orders/sec=%.1f\n",
		s.Config.HotShelfCapacity,
		s.Config.ColdShelfCapacity,
		s.Config.FrozenShelfCapacity,
		s.Config.OverflowCapacity,
		s.Config.OrdersPerSecond)

	fmt.Printf("Total orders to process: %d\n", len(s.Orders))

	// Start order generator
	s.wg.Add(1)
	go s.generateOrders()

	// Start delivery processor
	s.wg.Add(1)
	go s.processDeliveries()

	// Start expired order cleanup
	s.wg.Add(1)
	go s.cleanupExpiredOrders()

	// Start stats reporter
	s.wg.Add(1)
	go s.reportStats()

	// If a duration is set, use that as a maximum time
	if s.Config.SimulationDuration > 0 {
		fmt.Printf("Maximum simulation time: %d seconds\n", s.Config.SimulationDuration)

		// Create a timer for the maximum duration
		durationTimer := time.NewTimer(time.Duration(s.Config.SimulationDuration) * time.Second)

		// Wait for either the simulation to end naturally or the max duration to expire
		select {
		case <-durationTimer.C:
			fmt.Println("Maximum simulation time reached!")
			close(s.stop)
		case <-s.stop:
			// The simulation ended on its own
		}
	} else {
		// Wait for the stop signal
		<-s.stop
	}

	s.wg.Wait()
	fmt.Println("Simulation completed!")
	s.printFinalStats()
}

// Stop stops the simulation
func (s *Simulator) Stop() {
	close(s.stop)
	s.wg.Wait()
}

// createOrderFromList creates an order from the loaded list
func (s *Simulator) createOrderFromList() {
	orderData := s.Orders[s.ordersProcessed]
	modifiedDecayRate := orderData.DecayRate * s.decayModifier
	temp := order.Temperature(orderData.Temp)
	newOrder := order.NewOrder(orderData.Name, temp, orderData.ShelfLife, modifiedDecayRate)

	success := s.ShelfManager.PlaceOrder(newOrder)
	if success {
		fmt.Printf("📦 Order placed: %s (%s) - Shelf life: %.1fs, Decay rate: %.3f\n",
			newOrder.Name, newOrder.Temp, newOrder.ShelfLife, newOrder.DecayRate)
	} else {
		fmt.Printf("❌ Order wasted (no shelf space): %s (%s)\n", newOrder.Name, newOrder.Temp)
	}
	s.ordersProcessed++
}

// processLoadedOrders places all orders from the loaded list
func (s *Simulator) processLoadedOrders() {
	defer s.wg.Done()

	for _, orderData := range s.Orders {
		temp := order.Temperature(orderData.Temp)
		newOrder := order.NewOrder(orderData.Name, temp, orderData.ShelfLife, orderData.DecayRate)

		success := s.ShelfManager.PlaceOrder(newOrder)
		if success {
			fmt.Printf("📦 Order placed: %s (%s) - Shelf life: %.1fs, Decay rate: %.3f\n",
				newOrder.Name, newOrder.Temp, newOrder.ShelfLife, newOrder.DecayRate)
		} else {
			fmt.Printf("❌ Order wasted (no shelf space): %s (%s)\n", newOrder.Name, newOrder.Temp)
		}
	}
	close(s.stop) // Signal to stop after processing all orders
}

// processDeliveries simulates order deliveries
func (s *Simulator) processDeliveries() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.deliveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.attemptDeliveries()
		case <-s.stop:
			return
		}
	}
}

// attemptDeliveries attempts to deliver orders based on a probability
func (s *Simulator) attemptDeliveries() {
	// Get all orders
	allOrders := s.ShelfManager.GetAllOrders()
	if len(allOrders) == 0 {
		return
	}

	// For each order, there's a 30% chance it will be delivered in this cycle
	for _, order := range allOrders {
		//if rand.Float64() < 0.30 {
		// Introduce a random delay between 2 to 6 seconds before delivering the order
		randomDelay := time.Duration(rand.IntN(5)+2) * time.Second
		time.Sleep(randomDelay)

		if s.ShelfManager.DeliverOrder(order.ID) {
			fmt.Printf("🚚 Order delivered: %s (Value: %.2f)\n",
				order.Name, order.CalculateValue(time.Now()))
		}
		//}
	}
}

// cleanupExpiredOrders removes expired orders from shelves
func (s *Simulator) cleanupExpiredOrders() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			expired := s.ShelfManager.RemoveExpiredOrders()
			if expired > 0 {
				fmt.Printf("🗑️ Removed %d expired orders\n", expired)
			}
		case <-s.stop:
			return
		}
	}
}

// reportStats periodically reports simulation statistics
func (s *Simulator) reportStats() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.printCurrentStats()
		case <-s.stop:
			return
		}
	}
}

// printCurrentStats prints the current statistics of the simulation
func (s *Simulator) printCurrentStats() {
	stats := s.ShelfManager.GetStats()

	hotCurrent := stats["hotShelf"].(map[string]interface{})["current"].(int)
	coldCurrent := stats["coldShelf"].(map[string]interface{})["current"].(int)
	frozenCurrent := stats["frozenShelf"].(map[string]interface{})["current"].(int)
	overflowCurrent := stats["overflowShelf"].(map[string]interface{})["current"].(int)

	totalReceived := stats["totalOrders"].(map[string]interface{})["received"].(int)
	totalDelivered := stats["totalOrders"].(map[string]interface{})["delivered"].(int)
	totalWasted := stats["totalOrders"].(map[string]interface{})["wasted"].(int)
	totalExpired := stats["totalOrders"].(map[string]interface{})["expired"].(int)

	fmt.Println("\n📊 CURRENT SIMULATION STATS 📊")
	fmt.Println("------------------------------")
	fmt.Printf("Shelves: Hot=%d, Cold=%d, Frozen=%d, Overflow=%d\n",
		hotCurrent, coldCurrent, frozenCurrent, overflowCurrent)
	fmt.Printf("Orders: Received=%d, Delivered=%d, Wasted=%d, Expired=%d\n",
		totalReceived, totalDelivered, totalWasted, totalExpired)

	// Calculate percentages for better visibility
	deliveryRate := 0.0
	if totalReceived > 0 {
		deliveryRate = float64(totalDelivered) / float64(totalReceived) * 100
	}

	wasteRate := 0.0
	if totalReceived > 0 {
		wasteRate = float64(totalWasted+totalExpired) / float64(totalReceived) * 100
	}

	fmt.Printf("Delivery rate: %.1f%%, Waste rate: %.1f%%\n",
		deliveryRate, wasteRate)
	fmt.Println("------------------------------")
}

// printFinalStats prints the final statistics when the simulation ends
func (s *Simulator) printFinalStats() {
	stats := s.ShelfManager.GetStats()

	// Get shelf stats
	hotStats := stats["hotShelf"].(map[string]interface{})["stats"].(shelf.ShelfStats)
	coldStats := stats["coldShelf"].(map[string]interface{})["stats"].(shelf.ShelfStats)
	frozenStats := stats["frozenShelf"].(map[string]interface{})["stats"].(shelf.ShelfStats)
	overflowStats := stats["overflowShelf"].(map[string]interface{})["stats"].(shelf.ShelfStats)

	// Get total numbers
	totalReceived := stats["totalOrders"].(map[string]interface{})["received"].(int)
	totalDelivered := stats["totalOrders"].(map[string]interface{})["delivered"].(int)
	totalWasted := stats["totalOrders"].(map[string]interface{})["wasted"].(int)
	totalExpired := stats["totalOrders"].(map[string]interface{})["expired"].(int)

	fmt.Println("\n🎯 FINAL SIMULATION RESULTS 🎯")
	fmt.Println("===============================")

	fmt.Println("📦 ORDERS:")
	fmt.Printf("  Total received: %d\n", totalReceived)
	fmt.Printf("  Total delivered: %d (%.1f%%)\n",
		totalDelivered, float64(totalDelivered)/float64(totalReceived)*100)
	fmt.Printf("  Total wasted: %d (%.1f%%)\n",
		totalWasted, float64(totalWasted)/float64(totalReceived)*100)
	fmt.Printf("  Total expired: %d (%.1f%%)\n",
		totalExpired, float64(totalExpired)/float64(totalReceived)*100)

	fmt.Println("\n🔥 HOT SHELF:")
	fmt.Printf("  Orders added: %d\n", hotStats.OrdersAdded)
	fmt.Printf("  Orders delivered: %d\n", hotStats.OrdersDelivered)
	fmt.Printf("  Orders wasted: %d\n", hotStats.OrdersWasted)
	fmt.Printf("  Peak usage: %d\n", hotStats.PeakUsage)

	fmt.Println("\n❄️ COLD SHELF:")
	fmt.Printf("  Orders added: %d\n", coldStats.OrdersAdded)
	fmt.Printf("  Orders delivered: %d\n", coldStats.OrdersDelivered)
	fmt.Printf("  Orders wasted: %d\n", coldStats.OrdersWasted)
	fmt.Printf("  Peak usage: %d\n", coldStats.PeakUsage)

	fmt.Println("\n🧊 FROZEN SHELF:")
	fmt.Printf("  Orders added: %d\n", frozenStats.OrdersAdded)
	fmt.Printf("  Orders delivered: %d\n", frozenStats.OrdersDelivered)
	fmt.Printf("  Orders wasted: %d\n", frozenStats.OrdersWasted)
	fmt.Printf("  Peak usage: %d\n", frozenStats.PeakUsage)

	fmt.Println("\n♻️ OVERFLOW SHELF:")
	fmt.Printf("  Orders added: %d\n", overflowStats.OrdersAdded)
	fmt.Printf("  Orders delivered: %d\n", overflowStats.OrdersDelivered)
	fmt.Printf("  Orders wasted: %d\n", overflowStats.OrdersWasted)
	fmt.Printf("  Peak usage: %d\n", overflowStats.PeakUsage)

	fmt.Println("===============================")
}
