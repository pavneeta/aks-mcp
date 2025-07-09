package azureclient

import (
	"testing"
	"time"

	"github.com/Azure/aks-mcp/internal/config"
)

func TestNewAzureClientWithConfigurableTimeout(t *testing.T) {
	// Test default timeout
	cfg := config.NewConfig()
	client, err := NewAzureClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create Azure client: %v", err)
	}

	if client.cache.defaultTimeout != cfg.CacheTimeout {
		t.Errorf("Expected cache timeout to be %v, got %v", cfg.CacheTimeout, client.cache.defaultTimeout)
	}

	if client.cache.defaultTimeout != 1*time.Minute {
		t.Errorf("Expected default cache timeout to be 1 minute, got %v", client.cache.defaultTimeout)
	}

	// Test custom timeout
	customCfg := &config.ConfigData{
		CacheTimeout: 5 * time.Minute,
	}
	customClient, err := NewAzureClient(customCfg)
	if err != nil {
		t.Fatalf("Failed to create Azure client with custom config: %v", err)
	}

	if customClient.cache.defaultTimeout != customCfg.CacheTimeout {
		t.Errorf("Expected cache timeout to be %v, got %v", customCfg.CacheTimeout, customClient.cache.defaultTimeout)
	}

	if customClient.cache.defaultTimeout != 5*time.Minute {
		t.Errorf("Expected custom cache timeout to be 5 minutes, got %v", customClient.cache.defaultTimeout)
	}
}
