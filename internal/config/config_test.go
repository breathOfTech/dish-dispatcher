package config_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"dish-dispatcher/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	assert.Equal(t, 20, cfg.HotShelfCapacity)
	assert.Equal(t, 20, cfg.ColdShelfCapacity)
	assert.Equal(t, 20, cfg.FrozenShelfCapacity)
	assert.Equal(t, 30, cfg.OverflowCapacity)
	assert.Equal(t, 2.0, cfg.OrdersPerSecond)
	assert.Equal(t, 300, cfg.SimulationDuration)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := config.LoadConfig("non_existent_file.json")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, config.DefaultConfig(), cfg)
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "config.json")
	assert.NoError(t, err)

	defer os.Remove(tempFile.Name())

	expectedConfig := &config.Config{
		HotShelfCapacity:    10,
		ColdShelfCapacity:   15,
		FrozenShelfCapacity: 25,
		OverflowCapacity:    40,
		OrdersPerSecond:     3.5,
		SimulationDuration:  600,
	}
	configData, err := json.Marshal(expectedConfig)
	assert.NoError(t, err)

	_, err = tempFile.Write(configData)
	assert.NoError(t, err)
	tempFile.Close()

	cfg, err := config.LoadConfig(tempFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, cfg)
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tempFile, err := os.CreateTemp("", "invalid_config.json")
	assert.NoError(t, err)

	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte("invalid json"))
	assert.NoError(t, err)
	tempFile.Close()

	cfg, err := config.LoadConfig(tempFile.Name())
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
