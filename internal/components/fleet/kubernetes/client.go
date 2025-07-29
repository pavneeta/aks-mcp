package kubernetes

import (
	"fmt"

	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/k8s"
	"github.com/Azure/aks-mcp/internal/tools"
	"github.com/Azure/mcp-kubernetes/pkg/kubectl"
)

// Client wraps the mcp-kubernetes kubectl executor
type Client struct {
	executor tools.CommandExecutor
}

// NewClient creates a new Kubernetes client using mcp-kubernetes kubectl executor
func NewClient() (*Client, error) {
	// Create the mcp-kubernetes kubectl executor
	k8sExecutor := kubectl.NewExecutor()

	// Wrap it using the adapter to work with aks-mcp config
	wrappedExecutor := k8s.WrapK8sExecutor(k8sExecutor)

	return &Client{
		executor: wrappedExecutor,
	}, nil
}

// ExecuteKubectl executes a kubectl command
func (c *Client) ExecuteKubectl(command string, cfg *config.ConfigData) (string, error) {
	if c == nil {
		return "", fmt.Errorf("Client is nil")
	}
	if c.executor == nil {
		return "", fmt.Errorf("kubectl executor is nil")
	}

	params := map[string]interface{}{
		"command": command,
	}
	return c.executor.Execute(params, cfg)
}
