package azcli

import (
	"strings"
	"testing"

	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/security"
)

func TestFleetExecutor_ValidateCombination(t *testing.T) {
	executor := NewFleetExecutor()

	tests := []struct {
		name      string
		operation string
		resource  string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid fleet list",
			operation: "list",
			resource:  "fleet",
			wantErr:   false,
		},
		{
			name:      "valid member create",
			operation: "create",
			resource:  "member",
			wantErr:   false,
		},
		{
			name:      "valid updaterun start",
			operation: "start",
			resource:  "updaterun",
			wantErr:   false,
		},
		{
			name:      "invalid operation for fleet",
			operation: "start",
			resource:  "fleet",
			wantErr:   true,
			errMsg:    "invalid operation 'start' for resource 'fleet'",
		},
		{
			name:      "invalid operation for updatestrategy",
			operation: "update",
			resource:  "updatestrategy",
			wantErr:   true,
			errMsg:    "invalid operation 'update' for resource 'updatestrategy'",
		},
		{
			name:      "invalid resource",
			operation: "list",
			resource:  "invalid",
			wantErr:   true,
			errMsg:    "invalid resource type: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.validateCombination(tt.operation, tt.resource)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateCombination() error = nil, wantErr %v", tt.wantErr)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateCombination() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateCombination() unexpected error = %v", err)
			}
		})
	}
}

func TestFleetExecutor_CheckAccessLevel(t *testing.T) {
	executor := NewFleetExecutor()

	tests := []struct {
		name        string
		operation   string
		resource    string
		accessLevel string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "readonly can list",
			operation:   "list",
			resource:    "fleet",
			accessLevel: "readonly",
			wantErr:     false,
		},
		{
			name:        "readonly can show",
			operation:   "show",
			resource:    "member",
			accessLevel: "readonly",
			wantErr:     false,
		},
		{
			name:        "readonly cannot create",
			operation:   "create",
			resource:    "fleet",
			accessLevel: "readonly",
			wantErr:     true,
			errMsg:      "requires readwrite or admin access level",
		},
		{
			name:        "readwrite can create",
			operation:   "create",
			resource:    "fleet",
			accessLevel: "readwrite",
			wantErr:     false,
		},
		{
			name:        "admin can delete",
			operation:   "delete",
			resource:    "updaterun",
			accessLevel: "admin",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.checkAccessLevel(tt.operation, tt.resource, tt.accessLevel)
			if tt.wantErr {
				if err == nil {
					t.Errorf("checkAccessLevel() error = nil, wantErr %v", tt.wantErr)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("checkAccessLevel() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("checkAccessLevel() unexpected error = %v", err)
			}
		})
	}
}

func TestFleetExecutor_GetCommandForValidation(t *testing.T) {
	executor := NewFleetExecutor()

	tests := []struct {
		name      string
		operation string
		resource  string
		args      string
		want      string
	}{
		{
			name:      "fleet list special case",
			operation: "list",
			resource:  "fleet",
			args:      "--resource-group myRG",
			want:      "az fleet list --resource-group myRG",
		},
		{
			name:      "member show with args",
			operation: "show",
			resource:  "member",
			args:      "--name myMember --fleet-name myFleet --resource-group myRG",
			want:      "az fleet member show --name myMember --fleet-name myFleet --resource-group myRG",
		},
		{
			name:      "updaterun create without args",
			operation: "create",
			resource:  "updaterun",
			args:      "",
			want:      "az fleet updaterun create",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executor.GetCommandForValidation(tt.operation, tt.resource, tt.args)
			if got != tt.want {
				t.Errorf("GetCommandForValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFleetExecutor_Execute(t *testing.T) {
	// Note: This test validates parameter extraction and command construction
	// but doesn't execute actual az commands

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"operation": "list",
				"resource":  "fleet",
				"args":      "--resource-group myRG",
			},
			wantErr: false,
		},
		{
			name: "missing operation",
			params: map[string]interface{}{
				"resource": "fleet",
				"args":     "--resource-group myRG",
			},
			wantErr: true,
			errMsg:  "operation parameter is required",
		},
		{
			name: "missing resource",
			params: map[string]interface{}{
				"operation": "list",
				"args":      "--resource-group myRG",
			},
			wantErr: true,
			errMsg:  "resource parameter is required",
		},
		{
			name: "missing args",
			params: map[string]interface{}{
				"operation": "list",
				"resource":  "fleet",
			},
			wantErr: true,
			errMsg:  "args parameter is required",
		},
		{
			name: "invalid combination",
			params: map[string]interface{}{
				"operation": "start",
				"resource":  "fleet",
				"args":      "",
			},
			wantErr: true,
			errMsg:  "invalid operation 'start' for resource 'fleet'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewFleetExecutor()
			cfg := &config.ConfigData{
				AccessLevel: "readwrite",
				SecurityConfig: &security.SecurityConfig{
					AccessLevel: "readwrite",
				},
			}

			_, err := executor.Execute(tt.params, cfg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Execute() error = nil, wantErr %v", tt.wantErr)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Execute() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
			// Note: We don't test successful execution as it would require mocking the az CLI
		})
	}
}