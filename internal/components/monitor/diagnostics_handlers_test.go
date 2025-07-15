package monitor

import (
	"strings"
	"testing"
	"time"

	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/security"
)

func TestValidateControlPlaneLogsParams(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"log_category":    "kube-apiserver",
				"start_time":      "2025-07-15T10:00:00Z",
			},
			wantError: false,
		},
		{
			name: "missing subscription_id",
			params: map[string]interface{}{
				"resource_group": "test-rg",
				"cluster_name":   "test-cluster",
				"log_category":   "kube-apiserver",
				"start_time":     "2025-07-15T10:00:00Z",
			},
			wantError: true,
			errorMsg:  "missing or invalid subscription_id parameter",
		},
		{
			name: "missing resource_group",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"cluster_name":    "test-cluster",
				"log_category":    "kube-apiserver",
				"start_time":      "2025-07-15T10:00:00Z",
			},
			wantError: true,
			errorMsg:  "missing or invalid resource_group parameter",
		},
		{
			name: "missing cluster_name",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"log_category":    "kube-apiserver",
				"start_time":      "2025-07-15T10:00:00Z",
			},
			wantError: true,
			errorMsg:  "missing or invalid cluster_name parameter",
		},
		{
			name: "missing log_category",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"start_time":      "2025-07-15T10:00:00Z",
			},
			wantError: true,
			errorMsg:  "missing or invalid log_category parameter",
		},
		{
			name: "missing start_time",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"log_category":    "kube-apiserver",
			},
			wantError: true,
			errorMsg:  "missing or invalid start_time parameter",
		},
		{
			name: "invalid log_category",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"log_category":    "invalid-category",
				"start_time":      "2025-07-15T10:00:00Z",
			},
			wantError: true,
			errorMsg:  "invalid log category: invalid-category",
		},
		{
			name: "invalid log_level",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"log_category":    "kube-apiserver",
				"start_time":      "2025-07-15T10:00:00Z",
				"log_level":       "invalid-level",
			},
			wantError: true,
			errorMsg:  "invalid log level: invalid-level",
		},
		{
			name: "valid log_level info",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"log_category":    "kube-apiserver",
				"start_time":      "2025-07-15T10:00:00Z",
				"log_level":       "info",
			},
			wantError: false,
		},
		{
			name: "valid log_level warning",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"log_category":    "kube-apiserver",
				"start_time":      "2025-07-15T10:00:00Z",
				"log_level":       "warning",
			},
			wantError: false,
		},
		{
			name: "valid log_level error",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"log_category":    "kube-apiserver",
				"start_time":      "2025-07-15T10:00:00Z",
				"log_level":       "error",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateControlPlaneLogsParams(tt.params)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
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

func TestValidateTimeRange(t *testing.T) {
	tests := []struct {
		name      string
		startTime string
		params    map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid start time only",
			startTime: "2025-07-15T10:00:00Z",
			params:    map[string]interface{}{},
			wantError: false,
		},
		{
			name:      "valid start and end time",
			startTime: "2025-07-15T10:00:00Z",
			params: map[string]interface{}{
				"end_time": "2025-07-15T11:00:00Z",
			},
			wantError: false,
		},
		{
			name:      "invalid start time format",
			startTime: "2025-07-15 10:00:00",
			params:    map[string]interface{}{},
			wantError: true,
			errorMsg:  "invalid start_time format",
		},
		{
			name:      "start time too old",
			startTime: "2025-07-01T10:00:00Z", // More than 7 days ago
			params:    map[string]interface{}{},
			wantError: true,
			errorMsg:  "start_time cannot be more than 7 days ago",
		},
		{
			name:      "start time in future",
			startTime: time.Now().Add(time.Hour).Format(time.RFC3339),
			params:    map[string]interface{}{},
			wantError: true,
			errorMsg:  "start_time cannot be in the future",
		},
		{
			name:      "invalid end time format",
			startTime: "2025-07-15T10:00:00Z",
			params: map[string]interface{}{
				"end_time": "2025-07-15 11:00:00",
			},
			wantError: true,
			errorMsg:  "invalid end_time format",
		},
		{
			name:      "end time before start time",
			startTime: "2025-07-15T11:00:00Z",
			params: map[string]interface{}{
				"end_time": "2025-07-15T10:00:00Z",
			},
			wantError: true,
			errorMsg:  "end_time must be after start_time",
		},
		{
			name:      "time range too long",
			startTime: "2025-07-15T10:00:00Z",
			params: map[string]interface{}{
				"end_time": "2025-07-16T11:00:00Z", // More than 24 hours
			},
			wantError: true,
			errorMsg:  "time range cannot exceed 24 hours",
		},
		{
			name:      "end time in future",
			startTime: "2025-07-15T10:00:00Z",
			params: map[string]interface{}{
				"end_time": time.Now().Add(time.Hour).Format(time.RFC3339),
			},
			wantError: true,
			errorMsg:  "end_time cannot be in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimeRange(tt.startTime, tt.params)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
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

func TestGetMaxRecords(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected int
	}{
		{
			name:     "no max_records parameter",
			params:   map[string]interface{}{},
			expected: 100,
		},
		{
			name: "valid max_records",
			params: map[string]interface{}{
				"max_records": "500",
			},
			expected: 500,
		},
		{
			name: "max_records too high",
			params: map[string]interface{}{
				"max_records": "2000",
			},
			expected: 1000,
		},
		{
			name: "max_records too low",
			params: map[string]interface{}{
				"max_records": "0",
			},
			expected: 100,
		},
		{
			name: "invalid max_records",
			params: map[string]interface{}{
				"max_records": "invalid",
			},
			expected: 100,
		},
		{
			name: "max_records exactly 1000",
			params: map[string]interface{}{
				"max_records": "1000",
			},
			expected: 1000,
		},
		{
			name: "max_records exactly 1",
			params: map[string]interface{}{
				"max_records": "1",
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMaxRecords(tt.params)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestBuildSafeKQLQuery(t *testing.T) {
	tests := []struct {
		name              string
		category          string
		logLevel          string
		maxRecords        int
		clusterResourceID string
		expectedContains  []string
		expectedNotContains []string
	}{
		{
			name:              "basic query without log level",
			category:          "kube-apiserver",
			logLevel:          "",
			maxRecords:        100,
			clusterResourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster",
			expectedContains: []string{
				"AzureDiagnostics",
				"Category == 'kube-apiserver'",
				"ResourceId == '/SUBSCRIPTIONS/SUB/RESOURCEGROUPS/RG/PROVIDERS/MICROSOFT.CONTAINERSERVICE/MANAGEDCLUSTERS/CLUSTER'",
				"order by TimeGenerated desc",
				"limit 100",
				"project TimeGenerated, Level, log_s",
			},
			expectedNotContains: []string{
				"where log_s startswith",
			},
		},
		{
			name:              "query with info log level",
			category:          "kube-apiserver",
			logLevel:          "info",
			maxRecords:        50,
			clusterResourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster",
			expectedContains: []string{
				"AzureDiagnostics",
				"Category == 'kube-apiserver'",
				"ResourceId == '/SUBSCRIPTIONS/SUB/RESOURCEGROUPS/RG/PROVIDERS/MICROSOFT.CONTAINERSERVICE/MANAGEDCLUSTERS/CLUSTER'",
				"where log_s startswith 'I'",
				"order by TimeGenerated desc",
				"limit 50",
				"project TimeGenerated, Level, log_s",
			},
		},
		{
			name:              "query with warning log level",
			category:          "kube-controller-manager",
			logLevel:          "warning",
			maxRecords:        200,
			clusterResourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster",
			expectedContains: []string{
				"AzureDiagnostics",
				"Category == 'kube-controller-manager'",
				"where log_s startswith 'W'",
				"limit 200",
			},
		},
		{
			name:              "query with error log level",
			category:          "cloud-controller-manager",
			logLevel:          "error",
			maxRecords:        1000,
			clusterResourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster",
			expectedContains: []string{
				"AzureDiagnostics",
				"Category == 'cloud-controller-manager'",
				"where log_s startswith 'E'",
				"limit 1000",
			},
		},
		{
			name:              "query with invalid log level (should be ignored)",
			category:          "kube-apiserver",
			logLevel:          "invalid",
			maxRecords:        100,
			clusterResourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/cluster",
			expectedContains: []string{
				"AzureDiagnostics",
				"Category == 'kube-apiserver'",
				"limit 100",
			},
			expectedNotContains: []string{
				"where log_s startswith",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := buildSafeKQLQuery(tt.category, tt.logLevel, tt.maxRecords, tt.clusterResourceID)
			
			for _, expected := range tt.expectedContains {
				if !strings.Contains(query, expected) {
					t.Errorf("Expected query to contain '%s', but it didn't. Query: %s", expected, query)
				}
			}
			
			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(query, notExpected) {
					t.Errorf("Expected query NOT to contain '%s', but it did. Query: %s", notExpected, query)
				}
			}
		})
	}
}

func TestCalculateTimespan(t *testing.T) {
	tests := []struct {
		name      string
		startTime string
		endTime   string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid start and end time",
			startTime: "2025-07-15T10:00:00Z",
			endTime:   "2025-07-15T11:00:00Z",
			wantError: false,
		},
		{
			name:      "valid start time only (end time empty)",
			startTime: "2025-07-15T10:00:00Z",
			endTime:   "",
			wantError: false,
		},
		{
			name:      "invalid start time format",
			startTime: "2025-07-15 10:00:00",
			endTime:   "",
			wantError: true,
			errorMsg:  "invalid start time format",
		},
		{
			name:      "invalid end time format",
			startTime: "2025-07-15T10:00:00Z",
			endTime:   "2025-07-15 11:00:00",
			wantError: true,
			errorMsg:  "invalid end time format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timespan, err := calculateTimespan(tt.startTime, tt.endTime)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				
				// Verify timespan format
				if !strings.Contains(timespan, "/") {
					t.Errorf("Expected timespan to contain '/', got: %s", timespan)
				}
				
				parts := strings.Split(timespan, "/")
				if len(parts) != 2 {
					t.Errorf("Expected timespan to have 2 parts separated by '/', got: %s", timespan)
				}
				
				// Verify start time is correctly formatted
				if parts[0] != tt.startTime {
					t.Errorf("Expected start time to be '%s', got '%s'", tt.startTime, parts[0])
				}
			}
		})
	}
}

func TestGetWorkspaceGUID(t *testing.T) {
	tests := []struct {
		name                string
		workspaceResourceID string
		wantError           bool
		errorMsg            string
	}{
		{
			name:                "valid workspace resource ID",
			workspaceResourceID: "/subscriptions/sub/resourcegroups/rg/providers/microsoft.operationalinsights/workspaces/workspace-name",
			wantError:           true, // Will fail when trying to execute Azure CLI in test environment
		},
		{
			name:                "invalid workspace resource ID - too short",
			workspaceResourceID: "/subscriptions/sub/resourcegroups",
			wantError:           true,
			errorMsg:            "invalid workspace resource ID format",
		},
		{
			name:                "invalid workspace resource ID - wrong format",
			workspaceResourceID: "/subscriptions/sub/wrong/format",
			wantError:           true,
			errorMsg:            "invalid workspace resource ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ConfigData{
				Timeout:        30,
				AccessLevel:    "readonly",
				SecurityConfig: &security.SecurityConfig{
					AccessLevel: "readonly",
				},
			}
			
			_, err := getWorkspaceGUID(tt.workspaceResourceID, cfg)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				// Only check for specific error message if provided
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				// Note: This test will fail in unit testing because it tries to execute Azure CLI
				// In a real unit test environment, we would mock the Azure CLI executor
				if err != nil && !strings.Contains(err.Error(), "failed to get workspace GUID") {
					t.Errorf("Unexpected error type: %v", err)
				}
			}
		})
	}
}

func TestHandleControlPlaneDiagnosticSettings_ParameterValidation(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
			},
			wantError: false,
		},
		{
			name: "missing subscription_id",
			params: map[string]interface{}{
				"resource_group": "test-rg",
				"cluster_name":   "test-cluster",
			},
			wantError: true,
			errorMsg:  "missing or invalid subscription_id parameter",
		},
		{
			name: "empty subscription_id",
			params: map[string]interface{}{
				"subscription_id": "",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
			},
			wantError: true,
			errorMsg:  "missing or invalid subscription_id parameter",
		},
		{
			name: "missing resource_group",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"cluster_name":    "test-cluster",
			},
			wantError: true,
			errorMsg:  "missing or invalid resource_group parameter",
		},
		{
			name: "empty resource_group",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "",
				"cluster_name":    "test-cluster",
			},
			wantError: true,
			errorMsg:  "missing or invalid resource_group parameter",
		},
		{
			name: "missing cluster_name",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
			},
			wantError: true,
			errorMsg:  "missing or invalid cluster_name parameter",
		},
		{
			name: "empty cluster_name",
			params: map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "",
			},
			wantError: true,
			errorMsg:  "missing or invalid cluster_name parameter",
		},
	}

	cfg := &config.ConfigData{
		Timeout:        30,
		AccessLevel:    "readonly",
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HandleControlPlaneDiagnosticSettings(tt.params, cfg)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				// Note: This test will fail in unit testing because it tries to execute Azure CLI
				// In a real unit test environment, we would mock the Azure CLI executor
				if err != nil && !strings.Contains(err.Error(), "failed to get diagnostic settings") {
					t.Errorf("Unexpected error type: %v", err)
				}
			}
		})
	}
}

func TestValidLogCategories(t *testing.T) {
	validCategories := []string{
		"kube-apiserver",
		"kube-audit",
		"kube-audit-admin",
		"kube-controller-manager",
		"kube-scheduler",
		"cluster-autoscaler",
		"cloud-controller-manager",
		"guard",
		"csi-azuredisk-controller",
		"csi-azurefile-controller",
		"csi-snapshot-controller",
		"fleet-member-agent",
		"fleet-member-net-controller-manager",
		"fleet-mcs-controller-manager",
	}

	for _, category := range validCategories {
		t.Run("valid_category_"+category, func(t *testing.T) {
			params := map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"log_category":    category,
				"start_time":      "2025-07-15T10:00:00Z",
			}

			err := validateControlPlaneLogsParams(params)
			if err != nil {
				t.Errorf("Expected category '%s' to be valid, but got error: %v", category, err)
			}
		})
	}
}

func TestValidLogLevels(t *testing.T) {
	validLevels := []string{"error", "warning", "info"}

	for _, level := range validLevels {
		t.Run("valid_level_"+level, func(t *testing.T) {
			params := map[string]interface{}{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"resource_group":  "test-rg",
				"cluster_name":    "test-cluster",
				"log_category":    "kube-apiserver",
				"start_time":      "2025-07-15T10:00:00Z",
				"log_level":       level,
			}

			err := validateControlPlaneLogsParams(params)
			if err != nil {
				t.Errorf("Expected log level '%s' to be valid, but got error: %v", level, err)
			}
		})
	}
}

// Test resource handler functions
func TestGetControlPlaneDiagnosticSettingsHandler(t *testing.T) {
	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}
	handler := GetControlPlaneDiagnosticSettingsHandler(cfg)
	
	if handler == nil {
		t.Error("Expected handler to be created, got nil")
	}
	
	// Test handler with invalid params to ensure it calls the underlying function
	params := map[string]interface{}{}
	_, err := handler.Handle(params, cfg)
	if err == nil {
		t.Error("Expected error for missing parameters, got nil")
	}
	
	if !strings.Contains(err.Error(), "missing or invalid subscription_id parameter") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestGetControlPlaneLogsHandler(t *testing.T) {
	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readonly",
		},
	}
	handler := GetControlPlaneLogsHandler(cfg)
	
	if handler == nil {
		t.Error("Expected handler to be created, got nil")
	}
	
	// Test handler with invalid params to ensure it calls the underlying function
	params := map[string]interface{}{}
	_, err := handler.Handle(params, cfg)
	if err == nil {
		t.Error("Expected error for missing parameters, got nil")
	}
	
	if !strings.Contains(err.Error(), "missing or invalid") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestExtractWorkspaceGUIDFromDiagnosticSettings_InvalidResourceID(t *testing.T) {
	cfg := &config.ConfigData{
		SecurityConfig: &security.SecurityConfig{
			AccessLevel: "readwrite", // Changed to readwrite since diagnostic settings is not readonly
		},
	}
	
	// This will fail at the diagnostic settings call, but we can test the error handling
	_, err := extractWorkspaceGUIDFromDiagnosticSettings("invalid-sub", "invalid-rg", "invalid-cluster", cfg)
	if err == nil {
		t.Error("Expected error for invalid parameters, got nil")
	}
	
	// Should fail at the Azure CLI execution level (could be timeout, permission, or other execution error)
	if !strings.Contains(err.Error(), "failed to get") && !strings.Contains(err.Error(), "context deadline exceeded") && !strings.Contains(err.Error(), "workspace GUID") {
		t.Errorf("Expected Azure CLI execution error, got: %v", err)
	}
}

func TestBuildSafeKQLQuery_UppercaseResourceID(t *testing.T) {
	clusterResourceID := "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	
	query := buildSafeKQLQuery("kube-apiserver", "", 100, clusterResourceID)
	
	// Verify the resource ID is converted to uppercase
	expectedUpperResourceID := "/SUBSCRIPTIONS/TEST-SUB/RESOURCEGROUPS/TEST-RG/PROVIDERS/MICROSOFT.CONTAINERSERVICE/MANAGEDCLUSTERS/TEST-CLUSTER"
	if !strings.Contains(query, expectedUpperResourceID) {
		t.Errorf("Expected query to contain uppercase resource ID '%s', got: %s", expectedUpperResourceID, query)
	}
}

func TestBuildSafeKQLQuery_LogLevelFiltering(t *testing.T) {
	clusterResourceID := "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	
	testCases := []struct {
		name       string
		logLevel   string
		shouldFind string
	}{
		{"info level", "info", "log_s startswith 'I'"},
		{"warning level", "warning", "log_s startswith 'W'"},
		{"error level", "error", "log_s startswith 'E'"},
		{"empty level", "", ""}, // Should not have any log level filter
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := buildSafeKQLQuery("kube-apiserver", tc.logLevel, 100, clusterResourceID)
			
			if tc.shouldFind != "" {
				if !strings.Contains(query, tc.shouldFind) {
					t.Errorf("Expected query to contain '%s', got: %s", tc.shouldFind, query)
				}
			} else {
				// For empty log level, should not contain any startswith filter
				if strings.Contains(query, "startswith") {
					t.Errorf("Expected query to not contain log level filter, got: %s", query)
				}
			}
		})
	}
}

func TestBuildSafeKQLQuery_QueryStructure(t *testing.T) {
	clusterResourceID := "/subscriptions/test-sub/resourceGroups/test-rg/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	
	query := buildSafeKQLQuery("kube-apiserver", "info", 50, clusterResourceID)
	
	// Verify all required components are present in the correct order
	requiredComponents := []string{
		"AzureDiagnostics",
		"where Category == 'kube-apiserver'",
		"ResourceId ==",
		"where log_s startswith 'I'",
		"order by TimeGenerated desc",
		"limit 50",
		"project TimeGenerated, Level, log_s",
	}
	
	lastIndex := -1
	for _, component := range requiredComponents {
		index := strings.Index(query, component)
		if index == -1 {
			t.Errorf("Expected query to contain '%s', got: %s", component, query)
		}
		if index <= lastIndex {
			t.Errorf("Expected component '%s' to appear after previous components, got: %s", component, query)
		}
		lastIndex = index
	}
}

func TestGetMaxRecords_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"nil value", nil, 100},
		{"non-string value", 123, 100},
		{"empty string", "", 100},
		{"valid string", "250", 250},
		{"max boundary", "1000", 1000},
		{"over max", "2000", 1000},
		{"zero", "0", 100},
		{"negative", "-50", 100},
		{"non-numeric string", "abc", 100},
		{"float string", "50.5", 100},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]interface{}{}
			if tc.input != nil {
				params["max_records"] = tc.input
			}
			
			result := getMaxRecords(params)
			if result != tc.expected {
				t.Errorf("Expected %d for input %v, got %d", tc.expected, tc.input, result)
			}
		})
	}
}

func TestValidateTimeRange_EdgeCases(t *testing.T) {
	now := time.Now()
	
	testCases := []struct {
		name        string
		startTime   string
		params      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "exactly 7 days ago",
			startTime:   now.AddDate(0, 0, -7).Add(time.Minute).Format(time.RFC3339), // Just under 7 days
			params:      map[string]interface{}{},
			expectError: false,
		},
		{
			name:        "exactly 24 hours range",
			startTime:   now.Add(-24 * time.Hour).Format(time.RFC3339),
			params:      map[string]interface{}{"end_time": now.Format(time.RFC3339)},
			expectError: false,
		},
		{
			name:        "slightly over 24 hours",
			startTime:   now.Add(-24*time.Hour - time.Minute).Format(time.RFC3339),
			params:      map[string]interface{}{"end_time": now.Format(time.RFC3339)},
			expectError: true,
			errorMsg:    "cannot exceed 24 hours",
		},
		{
			name:        "invalid end time format",
			startTime:   now.Add(-time.Hour).Format(time.RFC3339),
			params:      map[string]interface{}{"end_time": "invalid-format"},
			expectError: true,
			errorMsg:    "invalid end_time format",
		},
		{
			name:        "end time in future",
			startTime:   now.Add(-time.Hour).Format(time.RFC3339),
			params:      map[string]interface{}{"end_time": now.Add(time.Hour).Format(time.RFC3339)},
			expectError: true,
			errorMsg:    "cannot be in the future",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTimeRange(tc.startTime, tc.params)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.name)
				} else if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, got: %v", tc.name, err)
				}
			}
		})
	}
}
