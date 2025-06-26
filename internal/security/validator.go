package security

import (
	"strings"
)

// Command type constants
const (
	CommandTypeAz = "az"
)

var (
	// AzReadOperations defines az operations that don't modify state
	AzReadOperations = []string{
		// Cluster information commands
		"az aks show",
		"az aks list",
		"az aks get-versions",
		"az aks get-upgrades",
		"az aks check-acr",
		"az aks check-network outbound",
		"az aks browse",

		// Addon commands
		"az aks addon list",
		"az aks addon show",

		// Nodepool commands
		"az aks nodepool list",
		"az aks nodepool show",
		"az aks nodepool get-upgrades",

		// Operation and snapshot commands
		"az aks operation",
		"az aks snapshot list",
		"az aks snapshot show",

		// Trusted access commands
		"az aks trustedaccess rolebinding list",
		"az aks trustedaccess rolebinding show",

		// Other read operations
		"az aks install-cli",
		// "az aks get-credentials", // Commented out as it may require special handling

		// Account management commands
		"az account list",
		"az account set",
		"az login",

		// Other general commands
		"az find",
		"az version",
		"az help",
		"az config",
		"az group list",
		"az group show",
		"az resource list",
		"az resource show",
	}
)

// Validator handles validation of commands against security configuration
type Validator struct {
	secConfig *SecurityConfig
}

// NewValidator creates a new Validator instance with the given security configuration
func NewValidator(secConfig *SecurityConfig) *Validator {
	return &Validator{
		secConfig: secConfig,
	}
}

// ValidationError represents a security validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// getReadOperationsList returns the appropriate list of read operations based on command type
func (v *Validator) getReadOperationsList(commandType string) []string {
	switch commandType {
	case CommandTypeAz:
		return AzReadOperations
	default:
		return []string{}
	}
}

// ValidateCommand validates a command against all security settings
// The command parameter should be the full command string (e.g., "az aks show --name myCluster")
// AzReadOperations should now contain full command prefixes with "az" included
func (v *Validator) ValidateCommand(command, commandType string) error {
	readOperations := v.getReadOperationsList(commandType)

	// Check access level restrictions
	if err := v.validateAccessLevel(command, readOperations); err != nil {
		return err
	}

	return nil
}

// validateAccessLevel validates if a command is allowed based on the current access level
func (v *Validator) validateAccessLevel(command string, readOperations []string) error {
	// Check if this is a read operation
	isReadOperation := v.isReadOperation(command, readOperations)

	// Handle restrictions based on access level
	switch v.secConfig.AccessLevel {
	case "readonly":
		if !isReadOperation {
			return &ValidationError{Message: "Error: Cannot execute write operations in read-only mode"}
		}
	case "readwrite":
		// All read and write operations are allowed, but not admin operations
		// Admin operations are handled separately by not registering those commands
	case "admin":
		// All operations are allowed
	default:
		// Default to readwrite behavior for unknown access levels
		// This could alternatively return an error for invalid access levels
	}

	return nil
}

// isReadOperation checks if a command is a read operation
func (v *Validator) isReadOperation(command string, allowedOperations []string) bool {
	// Check if the command contains help flags - these are always read-only
	if strings.Contains(command, "--help") || strings.Contains(command, " -h ") || strings.HasSuffix(command, " -h") {
		return true
	}

	// Normalize command by removing any options/arguments
	// This extracts the base command like "az aks show" from "az aks show --name myCluster"
	cmdParts := strings.Fields(command)
	
	if len(cmdParts) == 0 || cmdParts[0] != CommandTypeAz {
		return false
	}

	// For az commands, we need to handle various command structures:
	// - "az version" (2 parts)
	// - "az aks show" (3 parts)
	// - "az aks check-network outbound" (4 parts)
	// - "az aks trustedaccess rolebinding list" (5 parts)
	// - "az aks nodepool get-upgrades" (4 parts)
	
	// We'll try to match the longest possible command first by checking against allowed operations
	for _, allowed := range allowedOperations {
		allowedParts := strings.Fields(allowed)
		
		// Skip if the allowed operation has more parts than our command
		if len(allowedParts) > len(cmdParts) {
			continue
		}
		
		// Check if the command starts with this allowed operation
		match := true
		for i, allowedPart := range allowedParts {
			if i >= len(cmdParts) || cmdParts[i] != allowedPart {
				match = false
				break
			}
		}
		
		if match {
			return true
		}
	}

	return false
}
