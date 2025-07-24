package kubernetes

import (
	"fmt"

	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/tools"
)

// MockExecutor is a mock implementation of CommandExecutor for testing
type MockExecutor struct {
	ExecuteFunc func(params map[string]any, cfg *config.ConfigData) (string, error)
}

func (m *MockExecutor) Execute(params map[string]any, cfg *config.ConfigData) (string, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(params, cfg)
	}
	return "", fmt.Errorf("mock not implemented")
}

// Verify MockExecutor implements tools.CommandExecutor
var _ tools.CommandExecutor = (*MockExecutor)(nil)
