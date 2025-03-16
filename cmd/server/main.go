package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dish-dispatcher/internal/config"
	"dish-dispatcher/internal/simulator"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Path to configuration file")
	ordersFile := flag.String("orders", "orders.json", "Path to orders JSON file")
	flag.Parse()

	// Set random seed
	rand.Seed(time.Now().UnixNano())

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create simulator
	sim, err := simulator.NewSimulator(cfg, *ordersFile)
	if err != nil {
		fmt.Printf("Error creating simulator: %v\n", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run simulator in a separate goroutine if we need to handle interrupts
	done := make(chan struct{})

	go func() {
		sim.Run()
		close(done)
	}()

	// Wait for either:
	// 1. The simulation to finish naturally (if it has a duration)
	// 2. A keyboard interrupt
	select {
	case <-done:
		// Simulation finished naturally, just exit
		fmt.Println("Simulation completed successfully")
	case <-stop:
		fmt.Println("\nReceived interrupt signal, shutting down...")
		sim.Stop()
		fmt.Println("Shutdown complete")
	}
}
