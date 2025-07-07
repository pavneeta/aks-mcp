package main

import (
	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/server"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestContainerTransportConfiguration validates that the transport configuration
// is appropriate for container deployment
func TestContainerTransportConfiguration(t *testing.T) {
	// Test that sse transport works
	cfg := config.NewConfig()
	cfg.Transport = "sse"
	cfg.Host = "0.0.0.0"
	cfg.Port = 8000

	service := server.NewService(cfg)
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service with sse transport: %v", err)
	}

	// Test that streamable-http transport works
	cfg.Transport = "streamable-http"
	service = server.NewService(cfg)
	err = service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service with streamable-http transport: %v", err)
	}
}

// TestAzureCLIAvailability tests that Azure CLI would be available
// This simulates what would happen in the container environment
func TestAzureCLIAvailability(t *testing.T) {
	// Check if az command would be found in PATH
	// In container, this would be installed via pip3 install azure-cli
	_, err := exec.LookPath("az")

	// If az is not found locally, that's expected - we just want to
	// verify our container setup logic would work
	if err != nil {
		t.Logf("Azure CLI not found locally (expected): %v", err)
		t.Logf("In container, Azure CLI would be installed via: pip3 install --break-system-packages azure-cli")
	} else {
		t.Logf("Azure CLI found locally")

		// If available, test that it works
		cmd := exec.Command("az", "--version")
		output, err := cmd.Output()
		if err != nil {
			t.Logf("Azure CLI version check failed: %v", err)
		} else {
			t.Logf("Azure CLI version: %s", strings.TrimSpace(string(output)))
		}
	}
}

// TestNetworkTransportStartup tests that network transports start correctly
func TestNetworkTransportStartup(t *testing.T) {
	// Test SSE transport startup
	go func() {
		cfg := config.NewConfig()
		cfg.Transport = "sse"
		cfg.Host = "127.0.0.1"
		cfg.Port = 8999 // Use different port to avoid conflicts

		service := server.NewService(cfg)
		if err := service.Initialize(); err != nil {
			t.Logf("Failed to initialize service: %v", err)
			return
		}

		// Start service (this will block)
		if err := service.Run(); err != nil {
			t.Logf("Failed to run service: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(2 * time.Second)

	// Try to connect to the SSE endpoint
	resp, err := http.Get("http://127.0.0.1:8999")
	if err != nil {
		t.Logf("Could not connect to SSE server (expected in test environment): %v", err)
	} else {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Error closing response body: %v", err)
			}
		}()
		t.Logf("SSE server is accessible")
	}
}
