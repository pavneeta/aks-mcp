package security

import (
	"strings"
	"testing"
)

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name        string
		accessLevel string
		command     string
		wantErr     bool
	}{
		{
			name:        "ReadOnly_ReadCommand_ShouldSucceed",
			accessLevel: "readonly",
			command:     "az aks show --name myCluster --resource-group myRG",
			wantErr:     false,
		},
		{
			name:        "ReadOnly_WriteCommand_ShouldFail",
			accessLevel: "readonly",
			command:     "az aks create --name myCluster --resource-group myRG",
			wantErr:     true,
		},
		{
			name:        "ReadWrite_ReadCommand_ShouldSucceed",
			accessLevel: "readwrite",
			command:     "az aks show --name myCluster --resource-group myRG",
			wantErr:     false,
		},
		{
			name:        "ReadWrite_WriteCommand_ShouldSucceed",
			accessLevel: "readwrite",
			command:     "az aks create --name myCluster --resource-group myRG",
			wantErr:     false,
		},
		{
			name:        "Admin_ReadCommand_ShouldSucceed",
			accessLevel: "admin",
			command:     "az aks show --name myCluster --resource-group myRG",
			wantErr:     false,
		},
		{
			name:        "Admin_WriteCommand_ShouldSucceed",
			accessLevel: "admin",
			command:     "az aks create --name myCluster --resource-group myRG",
			wantErr:     false,
		},
		{
			name:        "Admin_AdminCommand_ShouldSucceed",
			accessLevel: "admin",
			command:     "az aks get-credentials --name myCluster --resource-group myRG",
			wantErr:     false,
		},
		{
			name:        "PartialCommand_ShouldMatch",
			accessLevel: "readonly",
			command:     "az version",
			wantErr:     false,
		},
		{
			name:        "AccountCommands_ShouldWork",
			accessLevel: "readonly",
			command:     "az account list",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secConfig := &SecurityConfig{
				AccessLevel: tt.accessLevel,
			}
			validator := NewValidator(secConfig)
			err := validator.ValidateCommand(tt.command, CommandTypeAz)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsReadOperation(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		allowList []string
		want      bool
	}{
		{
			name:      "ExactMatch",
			command:   "az aks show --name myCluster",
			allowList: []string{"az aks show", "az aks list"},
			want:      true,
		},
		{
			name:      "NoMatch",
			command:   "az aks create --name myCluster",
			allowList: []string{"az aks show", "az aks list"},
			want:      false,
		},
		{
			name:      "PrefixMatch",
			command:   "az account list",
			allowList: []string{"az account", "az aks show"},
			want:      true,
		},
		{
			name:      "TwoWordCommand",
			command:   "az version",
			allowList: []string{"az version", "az help"},
			want:      true,
		},
		{
			name:      "NonAzCommand",
			command:   "kubectl get pods",
			allowList: []string{"az aks show", "az aks list"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(&SecurityConfig{})
			if got := validator.isReadOperation(tt.command, tt.allowList); got != tt.want {
				t.Errorf("isReadOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessLevelValidation(t *testing.T) {
	// Setup common allowed operations
	readOps := []string{"az aks show", "az aks list"}

	tests := []struct {
		name        string
		accessLevel string
		command     string
		wantErr     bool
	}{
		{
			name:        "ReadOnly_ReadCommand",
			accessLevel: "readonly",
			command:     "az aks show --name myCluster",
			wantErr:     false,
		},
		{
			name:        "ReadOnly_WriteCommand",
			accessLevel: "readonly",
			command:     "az aks create --name myCluster",
			wantErr:     true,
		},
		{
			name:        "ReadWrite_ReadCommand",
			accessLevel: "readwrite",
			command:     "az aks show --name myCluster",
			wantErr:     false,
		},
		{
			name:        "ReadWrite_WriteCommand",
			accessLevel: "readwrite",
			command:     "az aks create --name myCluster",
			wantErr:     false,
		},
		{
			name:        "Admin_ReadCommand",
			accessLevel: "admin",
			command:     "az aks show --name myCluster",
			wantErr:     false,
		},
		{
			name:        "Admin_WriteCommand",
			accessLevel: "admin",
			command:     "az aks create --name myCluster",
			wantErr:     false,
		},
		{
			name:        "Unknown_ReadCommand",
			accessLevel: "unknown",
			command:     "az aks show --name myCluster",
			wantErr:     false, // Default to readwrite behavior
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secConfig := &SecurityConfig{
				AccessLevel: tt.accessLevel,
			}
			validator := NewValidator(secConfig)
			err := validator.validateAccessLevel(tt.command, readOps)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAccessLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsReadOperation_HelpFlags(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "create with --help should be read-only",
			command:  "az aks create --help",
			expected: true,
		},
		{
			name:     "delete with -h should be read-only",
			command:  "az aks delete -h",
			expected: true,
		},
		{
			name:     "nodepool add with --help should be read-only",
			command:  "az aks nodepool add --help",
			expected: true,
		},
		{
			name:     "command ending with -h should be read-only",
			command:  "az aks create -h",
			expected: true,
		},
		{
			name:     "command with -h in middle should be read-only",
			command:  "az aks create -h --name test",
			expected: true,
		},
		{
			name:     "command with --help in middle should be read-only",
			command:  "az aks nodepool delete --help --cluster-name test",
			expected: true,
		},
		{
			name:     "command with -h as part of argument value should not be read-only",
			command:  "az aks create --name cluster-h --resource-group rg",
			expected: false,
		},
		{
			name:     "command with help substring in argument should not be read-only",
			command:  "az aks create --name helpful-cluster --resource-group rg",
			expected: false,
		},
	}

	validator := NewValidator(&SecurityConfig{})
	// Use minimal allowed operations for testing
	allowedOps := []string{"az aks show", "az aks list"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isReadOperation(tt.command, allowedOps)
			if result != tt.expected {
				t.Errorf("isReadOperation(%q) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestIsReadOperation_TrustedAccessCommands(t *testing.T) {
	validator := NewValidator(&SecurityConfig{})

	// Get the actual read operations from the validator
	readOps := AzReadOperations

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "trustedaccess rolebinding list should be read-only",
			command:  "az aks trustedaccess rolebinding list",
			expected: true,
		},
		{
			name:     "trustedaccess rolebinding list with args should be read-only",
			command:  "az aks trustedaccess rolebinding list --cluster-name test --resource-group rg",
			expected: true,
		},
		{
			name:     "trustedaccess rolebinding show should be read-only",
			command:  "az aks trustedaccess rolebinding show",
			expected: true,
		},
		{
			name:     "trustedaccess rolebinding show with args should be read-only",
			command:  "az aks trustedaccess rolebinding show --cluster-name test --name binding",
			expected: true,
		},
		{
			name:     "trustedaccess rolebinding create should not be read-only",
			command:  "az aks trustedaccess rolebinding create --cluster-name test",
			expected: false,
		},
		{
			name:     "trustedaccess rolebinding delete should not be read-only",
			command:  "az aks trustedaccess rolebinding delete --cluster-name test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isReadOperation(tt.command, readOps)
			if result != tt.expected {
				t.Errorf("isReadOperation(%q) = %v, expected %v", tt.command, result, tt.expected)

				// Debug: let's see what base command is being extracted
				cmdParts := strings.Fields(tt.command)
				var baseCommand string
				if len(cmdParts) >= 3 && cmdParts[0] == CommandTypeAz {
					baseCommand = strings.Join(cmdParts[:3], " ")
				}
				t.Logf("Extracted base command: %q", baseCommand)

				// Check if it's in the allowed operations
				found := false
				for _, allowed := range readOps {
					if baseCommand == allowed || strings.HasPrefix(baseCommand, allowed) {
						found = true
						t.Logf("Matched against allowed operation: %q", allowed)
						break
					}
				}
				if !found {
					t.Logf("No match found in allowed operations")
				}
			}
		})
	}
}

func TestIsReadOperation_LongCommands(t *testing.T) {
	validator := NewValidator(&SecurityConfig{})
	readOps := AzReadOperations

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "check-network outbound should be read-only",
			command:  "az aks check-network outbound --name test --resource-group rg",
			expected: true,
		},
		{
			name:     "nodepool get-upgrades should be read-only",
			command:  "az aks nodepool get-upgrades --cluster-name test --name pool1",
			expected: true,
		},
		{
			name:     "addon list should be read-only",
			command:  "az aks addon list --name test --resource-group rg",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isReadOperation(tt.command, readOps)
			if result != tt.expected {
				t.Errorf("isReadOperation(%q) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestValidateCommand_WithHelpFlags(t *testing.T) {
	tests := []struct {
		name        string
		accessLevel string
		command     string
		expectError bool
	}{
		{
			name:        "readonly mode allows write commands with --help",
			accessLevel: "readonly",
			command:     "az aks create --help",
			expectError: false,
		},
		{
			name:        "readonly mode allows write commands with -h",
			accessLevel: "readonly",
			command:     "az aks delete -h",
			expectError: false,
		},
		{
			name:        "readonly mode allows nodepool commands with --help",
			accessLevel: "readonly",
			command:     "az aks nodepool add --help",
			expectError: false,
		},
		{
			name:        "readonly mode blocks write commands without help",
			accessLevel: "readonly",
			command:     "az aks create --name test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(&SecurityConfig{AccessLevel: tt.accessLevel})
			err := validator.ValidateCommand(tt.command, CommandTypeAz)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none for command: %q", tt.command)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v for command: %q", err, tt.command)
			}
		})
	}
}

func TestSpecificTrustedAccessFix(t *testing.T) {
	validator := NewValidator(&SecurityConfig{})
	readOps := AzReadOperations

	// Test the specific case that was failing
	command := "az aks trustedaccess rolebinding list --cluster-name test"
	result := validator.isReadOperation(command, readOps)

	if !result {
		t.Errorf("Expected trustedaccess rolebinding list to be read-only, but got false")

		// Debug output
		cmdParts := strings.Fields(command)
		t.Logf("Command parts: %v", cmdParts)

		for _, allowed := range readOps {
			if strings.Contains(allowed, "trustedaccess") {
				t.Logf("Found allowed trustedaccess command: %q", allowed)
				allowedParts := strings.Fields(allowed)
				t.Logf("Allowed parts: %v", allowedParts)
			}
		}
	}
}
