package diagnostics

import (
	"strings"
	"testing"

	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/security"
)

func TestGetWorkspaceGUID(t *testing.T) {
	tests := []struct {
		name                string
		workspaceResourceID string
		wantError           bool
		errorMsg            string
	}{
		{
			name:                "invalid resource ID format - too short",
			workspaceResourceID: "/invalid/format",
			wantError:           true,
			errorMsg:            "invalid workspace resource ID format",
		},
		{
			name:                "invalid resource ID format - missing parts",
			workspaceResourceID: "/subscriptions/test/resourceGroups/rg/providers/Microsoft.OperationalInsights",
			wantError:           true,
			errorMsg:            "invalid workspace resource ID format",
		},
		{
			name:                "valid resource ID format structure",
			workspaceResourceID: "/subscriptions/test/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/workspace",
			wantError:           true, // Will fail at Azure CLI execution level in test
		},
		{
			name:                "case insensitive resource ID parsing",
			workspaceResourceID: "/subscriptions/test/RESOURCEGROUPS/rg/providers/microsoft.operationalinsights/WORKSPACES/workspace",
			wantError:           true, // Will fail at Azure CLI execution level in test
		},
	}

	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getWorkspaceGUID(tt.workspaceResourceID, cfg)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGetWorkspaceGUID_ResourceIDParsing(t *testing.T) {
	// Test that we can properly extract resource group and workspace name from resource ID
	validResourceID := "/subscriptions/12345/resourceGroups/test-rg/providers/Microsoft.OperationalInsights/workspaces/test-workspace"

	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	// This will fail at Azure CLI execution but we can check that parsing doesn't fail immediately
	_, err := getWorkspaceGUID(validResourceID, cfg)

	// Should get an Azure CLI execution error, not a parsing error
	if err != nil && strings.Contains(err.Error(), "invalid workspace resource ID format") {
		t.Errorf("Expected Azure CLI execution error, got parsing error: %v", err)
	}
}

func TestExtractWorkspaceGUIDFromDiagnosticSettings_InvalidParameters(t *testing.T) {
	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	// This will fail at the diagnostic settings call, but we can test the error handling
	_, err := ExtractWorkspaceGUIDFromDiagnosticSettings("invalid", "invalid", "invalid", nil, cfg) // Pass nil Azure client for testing
	if err == nil {
		t.Error("Expected error for invalid parameters, got nil")
	}

	// Should fail at the Azure CLI execution level (could be timeout, permission, or other execution error)
	if !strings.Contains(err.Error(), "failed to get diagnostic settings") {
		t.Errorf("Expected diagnostic settings error, got: %v", err)
	}
}

func TestExtractWorkspaceGUIDFromDiagnosticSettings_JSONParsing(t *testing.T) {
	// Test JSON parsing logic with mock data
	// Note: This would require mocking the HandleControlPlaneDiagnosticSettings function
	// For now, we test the error case with invalid cluster details

	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	// Test with empty strings (should fail validation)
	_, err := ExtractWorkspaceGUIDFromDiagnosticSettings("", "", "", nil, cfg) // Pass nil Azure client for testing
	if err == nil {
		t.Error("Expected error for empty parameters, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get diagnostic settings") {
		t.Errorf("Expected diagnostic settings failure, got: %v", err)
	}
}

func TestGetWorkspaceGUID_EdgeCases(t *testing.T) {
	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	testCases := []struct {
		name        string
		resourceID  string
		expectError string
	}{
		{
			name:        "empty resource ID",
			resourceID:  "",
			expectError: "invalid workspace resource ID format",
		},
		{
			name:        "resource ID with no resource groups",
			resourceID:  "/subscriptions/test/providers/Microsoft.OperationalInsights/workspaces/workspace",
			expectError: "invalid workspace resource ID format",
		},
		{
			name:        "resource ID with no workspaces",
			resourceID:  "/subscriptions/test/resourceGroups/rg/providers/Microsoft.OperationalInsights",
			expectError: "invalid workspace resource ID format",
		},
		{
			name:        "resource ID with missing workspace name",
			resourceID:  "/subscriptions/test/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/",
			expectError: "could not extract resource group and workspace name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := getWorkspaceGUID(tc.resourceID, cfg)
			if err == nil {
				t.Errorf("Expected error for case '%s', got nil", tc.name)
				return
			}

			if !strings.Contains(err.Error(), tc.expectError) {
				t.Errorf("Expected error to contain '%s', got '%s'", tc.expectError, err.Error())
			}
		})
	}
}

func TestFindDiagnosticSettingForCategory(t *testing.T) {
	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	tests := []struct {
		name          string
		logCategory   string
		expectError   bool
		expectedError string
	}{
		{
			name:          "kube-apiserver category",
			logCategory:   "kube-apiserver",
			expectError:   true, // Will fail during Azure CLI execution in test environment
			expectedError: "",
		},
		{
			name:          "kube-audit category",
			logCategory:   "kube-audit",
			expectError:   true, // Will fail during Azure CLI execution in test environment
			expectedError: "",
		},
		{
			name:          "invalid category",
			logCategory:   "invalid-category",
			expectError:   true, // Will fail during Azure CLI execution in test environment
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := FindDiagnosticSettingForCategory("test-sub", "test-rg", "test-cluster", tt.logCategory, nil, cfg) // Pass nil Azure client for testing

			if tt.expectError && err == nil {
				t.Errorf("Expected error for category %s, got nil", tt.logCategory)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for category %s: %v", tt.logCategory, err)
			}
		})
	}
}

// TestFindDiagnosticSettingForCategory_NegativeCases tests specific error conditions
func TestFindDiagnosticSettingForCategory_NegativeCases(t *testing.T) {
	// Note: These tests would ideally use mocking for HandleControlPlaneDiagnosticSettings
	// For now, we test the edge cases that we can reach with invalid parameters

	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	tests := []struct {
		name           string
		subscriptionID string
		resourceGroup  string
		clusterName    string
		logCategory    string
		expectedError  string
	}{
		{
			name:           "empty subscription ID",
			subscriptionID: "",
			resourceGroup:  "test-rg",
			clusterName:    "test-cluster",
			logCategory:    "kube-apiserver",
			expectedError:  "failed to get diagnostic settings",
		},
		{
			name:           "empty resource group",
			subscriptionID: "test-sub",
			resourceGroup:  "",
			clusterName:    "test-cluster",
			logCategory:    "kube-apiserver",
			expectedError:  "failed to get diagnostic settings",
		},
		{
			name:           "empty cluster name",
			subscriptionID: "test-sub",
			resourceGroup:  "test-rg",
			clusterName:    "",
			logCategory:    "kube-apiserver",
			expectedError:  "failed to get diagnostic settings",
		},
		{
			name:           "empty log category",
			subscriptionID: "test-sub",
			resourceGroup:  "test-rg",
			clusterName:    "test-cluster",
			logCategory:    "",
			expectedError:  "failed to get diagnostic settings", // Will fail before reaching the category logic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := FindDiagnosticSettingForCategory(tt.subscriptionID, tt.resourceGroup, tt.clusterName, tt.logCategory, nil, cfg)

			if err == nil {
				t.Errorf("Expected error for case '%s', got nil", tt.name)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

// TestFindDiagnosticSettingForCategory_JSONStructureEdgeCases tests JSON parsing edge cases
func TestFindDiagnosticSettingForCategory_JSONStructureEdgeCases(t *testing.T) {
	// Note: To fully test JSON parsing edge cases, we would need to mock HandleControlPlaneDiagnosticSettings
	// These tests focus on parameter validation and error propagation that we can test

	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	// Test with invalid characters that might cause JSON parsing issues
	invalidChars := []string{
		"test-cluster-with-unicode-\u0000",
		"test-cluster-with-newline-\n",
		"test-cluster-with-tab-\t",
	}

	for _, invalidCluster := range invalidChars {
		t.Run("cluster_name_with_special_chars", func(t *testing.T) {
			_, _, err := FindDiagnosticSettingForCategory("test-sub", "test-rg", invalidCluster, "kube-apiserver", nil, cfg)

			// Should get an error (likely from Azure CLI execution)
			if err == nil {
				t.Errorf("Expected error for cluster name with special characters, got nil")
			}

			// Should fail at diagnostic settings level, not JSON parsing
			if strings.Contains(err.Error(), "failed to parse diagnostic settings JSON") {
				t.Errorf("Unexpected JSON parsing error for input validation case: %v", err)
			}
		})
	}
}

// TestFindDiagnosticSettingForCategory_MissingWorkspaceScenarios tests workspace configuration issues
func TestFindDiagnosticSettingForCategory_MissingWorkspaceScenarios(t *testing.T) {
	// Note: These tests would benefit from mocking to inject specific JSON responses
	// For now, we test the error propagation paths we can reach

	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	// Test with parameters that should result in "no diagnostic setting found" error
	// This tests the final error condition when no matching category is found
	testCases := []struct {
		name        string
		logCategory string
	}{
		{
			name:        "non_existent_category",
			logCategory: "non-existent-log-category-12345",
		},
		{
			name:        "very_long_category_name",
			logCategory: "this-is-a-very-long-category-name-that-definitely-does-not-exist-in-any-diagnostic-setting",
		},
		{
			name:        "category_with_special_chars",
			logCategory: "kube-audit@#$%",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := FindDiagnosticSettingForCategory("test-sub", "test-rg", "test-cluster", tc.logCategory, nil, cfg)

			if err == nil {
				t.Errorf("Expected error for non-existent category '%s', got nil", tc.logCategory)
				return
			}

			// Should eventually result in "no diagnostic setting found" error
			// (after failing to get diagnostic settings from Azure CLI)
			expectedErrors := []string{
				"failed to get diagnostic settings",
				"no diagnostic setting found",
			}

			errorMatched := false
			for _, expectedError := range expectedErrors {
				if strings.Contains(err.Error(), expectedError) {
					errorMatched = true
					break
				}
			}

			if !errorMatched {
				t.Errorf("Expected error to contain one of %v, got '%s'", expectedErrors, err.Error())
			}
		})
	}
}

// TestFindDiagnosticSettingForCategory_ErrorPathCoverage tests specific error handling paths
func TestFindDiagnosticSettingForCategory_ErrorPathCoverage(t *testing.T) {
	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	// Test parameter validation edge cases
	paramTests := []struct {
		name           string
		subscriptionID string
		resourceGroup  string
		clusterName    string
		logCategory    string
		description    string
	}{
		{
			name:           "all_empty_parameters",
			subscriptionID: "",
			resourceGroup:  "",
			clusterName:    "",
			logCategory:    "",
			description:    "Tests handling of completely empty input",
		},
		{
			name:           "whitespace_only_parameters",
			subscriptionID: "   ",
			resourceGroup:  "\t",
			clusterName:    "\n",
			logCategory:    " \t\n ",
			description:    "Tests handling of whitespace-only input",
		},
		{
			name:           "very_long_parameters",
			subscriptionID: strings.Repeat("a", 1000),
			resourceGroup:  strings.Repeat("b", 1000),
			clusterName:    strings.Repeat("c", 1000),
			logCategory:    strings.Repeat("d", 1000),
			description:    "Tests handling of extremely long input parameters",
		},
		{
			name:           "special_characters_in_all_params",
			subscriptionID: "sub-with-@#$%^&*()",
			resourceGroup:  "rg-with-[]{}|\\",
			clusterName:    "cluster-with-<>?/",
			logCategory:    "category-with-+=~`",
			description:    "Tests handling of special characters that might break JSON",
		},
	}

	for _, tt := range paramTests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := FindDiagnosticSettingForCategory(tt.subscriptionID, tt.resourceGroup, tt.clusterName, tt.logCategory, nil, cfg)

			// All these cases should result in errors (either from parameter validation or Azure CLI execution)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.description)
				return
			}

			// The error should be related to diagnostic settings retrieval, not internal crashes
			if strings.Contains(err.Error(), "panic") || strings.Contains(err.Error(), "runtime error") {
				t.Errorf("Unexpected runtime error for %s: %v", tt.description, err)
			}
		})
	}
}

// TestFindDiagnosticSettingForCategory_ConcurrencyAndStress tests function under stress
func TestFindDiagnosticSettingForCategory_ConcurrencyAndStress(t *testing.T) {
	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	// Test that the function handles concurrent calls gracefully
	// This helps ensure there are no race conditions in error handling
	const numGoroutines = 10
	const callsPerGoroutine = 5

	errChan := make(chan error, numGoroutines*callsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < callsPerGoroutine; j++ {
				_, _, err := FindDiagnosticSettingForCategory(
					"test-sub",
					"test-rg",
					"test-cluster",
					"kube-apiserver",
					nil,
					cfg,
				)
				errChan <- err
			}
		}(i)
	}

	// Collect all results
	var errors []error
	for i := 0; i < numGoroutines*callsPerGoroutine; i++ {
		err := <-errChan
		if err != nil {
			errors = append(errors, err)
		}
	}

	// All calls should fail (due to test environment), but should fail gracefully
	if len(errors) != numGoroutines*callsPerGoroutine {
		t.Errorf("Expected all %d calls to fail in test environment, got %d failures",
			numGoroutines*callsPerGoroutine, len(errors))
	}

	// Check that errors are consistent and don't indicate race conditions
	for _, err := range errors {
		if strings.Contains(err.Error(), "panic") ||
			strings.Contains(err.Error(), "runtime error") ||
			strings.Contains(err.Error(), "concurrent map") {
			t.Errorf("Detected potential race condition or runtime error: %v", err)
		}
	}
}
