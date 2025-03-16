package config

import (
	"encoding/json"
	"os"
)

// ShelfConfig contains configuration for a shelf
type ShelfConfig struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

// Config contains all configuration parameters for the simulation
type Config struct {
	HotShelfCapacity    int     `json:"hotShelfCapacity"`
	ColdShelfCapacity   int     `json:"coldShelfCapacity"`
	FrozenShelfCapacity int     `json:"frozenShelfCapacity"`
	OverflowCapacity    int     `json:"overflowCapacity"`
	OrdersPerSecond     float64 `json:"ordersPerSecond"`
	SimulationDuration  int     `json:"simulationDuration"` // in seconds, 0 means run indefinitely
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		HotShelfCapacity:    20,
		ColdShelfCapacity:   20,
		FrozenShelfCapacity: 20,
		OverflowCapacity:    30,
		OrdersPerSecond:     2.0,
		SimulationDuration:  300, // 5 minutes by default
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	// Start with default config
	config := DefaultConfig()

	// Try to open and parse the config file
	file, err := os.Open(path)
	if err != nil {
		// If file doesn't exist, return default config
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}
	defer file.Close()

	// Parse JSON into config struct
	if err := json.NewDecoder(file).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
