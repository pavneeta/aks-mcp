package diagnostics

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateControlPlaneLogsParams(tt.params)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeRange(tt.startTime, tt.params)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMaxRecords(tt.params)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestBuildSafeKQLQuery(t *testing.T) {
	tests := []struct {
		name             string
		category         string
		logLevel         string
		maxRecords       int
		clusterResourceID string
		expectedContains []string
	}{
		{
			name:             "basic query without log level",
			category:         "kube-apiserver",
			logLevel:         "",
			maxRecords:       100,
			clusterResourceID: "/subscriptions/test/resourcegroups/rg/providers/microsoft.containerservice/managedclusters/cluster",
			expectedContains: []string{
				"AzureDiagnostics",
				"where Category == 'kube-apiserver'",
				"limit 100",
				"project TimeGenerated, Level, log_s",
			},
		},
		{
			name:             "query with info log level",
			category:         "kube-apiserver",
			logLevel:         "info",
			maxRecords:       50,
			clusterResourceID: "/subscriptions/test/resourcegroups/rg/providers/microsoft.containerservice/managedclusters/cluster",
			expectedContains: []string{
				"where log_s startswith 'I'",
				"limit 50",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := BuildSafeKQLQuery(tt.category, tt.logLevel, tt.maxRecords, tt.clusterResourceID)
			for _, expected := range tt.expectedContains {
				if !strings.Contains(query, expected) {
					t.Errorf("Expected query to contain '%s', got '%s'", expected, query)
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
	}{
		{
			name:      "valid start and end time",
			startTime: "2025-07-15T10:00:00Z",
			endTime:   "2025-07-15T11:00:00Z",
			wantError: false,
		},
		{
			name:      "valid start time, empty end time",
			startTime: "2025-07-15T10:00:00Z",
			endTime:   "",
			wantError: false,
		},
		{
			name:      "invalid start time format",
			startTime: "invalid-time",
			endTime:   "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateTimespan(tt.startTime, tt.endTime)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

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
