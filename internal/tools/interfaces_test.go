package tools

import (
	"testing"

	"github.com/Azure/aks-mcp/internal/config"
)

func TestResourceHandlerInterface(t *testing.T) {
	// Test that ResourceHandlerFunc implements ResourceHandler
	var handler ResourceHandler = ResourceHandlerFunc(func(params map[string]interface{}, cfg *config.ConfigData) (string, error) {
		return "test result", nil
	})

	cfg := config.NewConfig()
	params := make(map[string]interface{})

	result, err := handler.Handle(params, cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "test result" {
		t.Errorf("Expected 'test result', got: %s", result)
	}
}

func TestCommandExecutorStillWorks(t *testing.T) {
	// Test that existing CommandExecutor interface still works
	var executor CommandExecutor = CommandExecutorFunc(func(params map[string]interface{}, cfg *config.ConfigData) (string, error) {
		return "command result", nil
	})

	cfg := config.NewConfig()
	params := make(map[string]interface{})

	result, err := executor.Execute(params, cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "command result" {
		t.Errorf("Expected 'command result', got: %s", result)
	}
}
