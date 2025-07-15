package azcli

import (
	"fmt"
	"strings"

	"github.com/Azure/aks-mcp/internal/config"
)

// FleetExecutor handles structured fleet command execution
type FleetExecutor struct {
	*AzExecutor
}

// NewFleetExecutor creates a new fleet command executor
func NewFleetExecutor() *FleetExecutor {
	return &FleetExecutor{
		AzExecutor: NewExecutor(),
	}
}

// Execute processes structured fleet commands
func (e *FleetExecutor) Execute(params map[string]interface{}, cfg *config.ConfigData) (string, error) {
	// Extract structured parameters
	operation, ok := params["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation parameter is required and must be a string")
	}

	resource, ok := params["resource"].(string)
	if !ok {
		return "", fmt.Errorf("resource parameter is required and must be a string")
	}

	args, ok := params["args"].(string)
	if !ok {
		return "", fmt.Errorf("args parameter is required and must be a string")
	}

	// Validate operation/resource combination
	if err := e.validateCombination(operation, resource); err != nil {
		return "", err
	}

	// Construct the full command
	command := fmt.Sprintf("az fleet %s %s", resource, operation)
	if operation == "list" && resource == "fleet" {
		// Special case: "az fleet list" without resource in between
		command = "az fleet list"
	}

	// Check access level
	if err := e.checkAccessLevel(operation, resource, cfg.AccessLevel); err != nil {
		return "", err
	}

	// Build full command with args
	fullCommand := command
	if args != "" {
		fullCommand = fmt.Sprintf("%s %s", command, args)
	}

	// Create params for the base executor
	execParams := map[string]interface{}{
		"command": fullCommand,
	}

	// Execute using the base executor
	return e.AzExecutor.Execute(execParams, cfg)
}

// validateCombination validates if the operation/resource combination is valid
func (e *FleetExecutor) validateCombination(operation, resource string) error {
	validCombinations := map[string][]string{
		"fleet":          {"list", "show", "create", "update", "delete"},
		"member":         {"list", "show", "create", "update", "delete"},
		"updaterun":      {"list", "show", "create", "start", "stop", "delete"},
		"updatestrategy": {"list", "show", "create", "delete"},
	}

	validOps, exists := validCombinations[resource]
	if !exists {
		return fmt.Errorf("invalid resource type: %s", resource)
	}

	for _, validOp := range validOps {
		if operation == validOp {
			return nil
		}
	}

	return fmt.Errorf("invalid operation '%s' for resource '%s'. Valid operations: %s",
		operation, resource, strings.Join(validOps, ", "))
}

// checkAccessLevel ensures the operation is allowed for the current access level
func (e *FleetExecutor) checkAccessLevel(operation, resource string, accessLevel string) error {
	// Read-only operations are allowed for all access levels
	readOnlyOps := []string{"list", "show"}
	for _, op := range readOnlyOps {
		if operation == op {
			return nil
		}
	}

	// Write operations require readwrite or admin access
	if accessLevel == "readonly" {
		return fmt.Errorf("operation '%s' requires readwrite or admin access level, current level is readonly", operation)
	}

	// All operations are allowed for readwrite and admin
	return nil
}

// GetCommandForValidation returns the constructed command for security validation
func (e *FleetExecutor) GetCommandForValidation(operation, resource, args string) string {
	command := fmt.Sprintf("az fleet %s %s", resource, operation)
	if operation == "list" && resource == "fleet" {
		command = "az fleet list"
	}

	if args != "" {
		command = fmt.Sprintf("%s %s", command, args)
	}

	return command
}
