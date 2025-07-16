package diagnostics

import (
	"strings"
	"testing"
	"time"
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
			errorMsg:  "time range cannot exceed 24h0m0s",
		},
		{
			name:      "end time in future",
			startTime: time.Now().Add(-time.Hour).Format(time.RFC3339), // 1 hour ago
			params: map[string]interface{}{
				"end_time": time.Now().Add(time.Hour).Format(time.RFC3339), // 1 hour in future
			},
			wantError: true,
			errorMsg:  "end_time cannot be in the future",
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
		{
			name: "negative max_records",
			params: map[string]interface{}{
				"max_records": "-10",
			},
			expected: 100, // Should default to 100 for invalid negative values
		},
		{
			name: "zero max_records",
			params: map[string]interface{}{
				"max_records": "0",
			},
			expected: 100, // Should default to 100 for zero
		},
		{
			name: "float max_records",
			params: map[string]interface{}{
				"max_records": "150.5",
			},
			expected: 100, // Should default to 100 for invalid float
		},
		{
			name: "empty string max_records",
			params: map[string]interface{}{
				"max_records": "",
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

			err := ValidateControlPlaneLogsParams(params)
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

			err := ValidateControlPlaneLogsParams(params)
			if err != nil {
				t.Errorf("Expected log level '%s' to be valid, but got error: %v", level, err)
			}
		})
	}
}

func TestValidateTimeRange_EdgeCases(t *testing.T) {
	// Test the edge case where end_time equals start_time
	now := time.Now()
	sameTime := now.Format(time.RFC3339)

	err := ValidateTimeRange(sameTime, map[string]interface{}{
		"end_time": sameTime,
	})

	// This should be valid (same time is allowed)
	if err != nil {
		t.Errorf("Expected same start and end time to be valid, got error: %v", err)
	}

	// Test close to the 7-day boundary (6 days and 23 hours ago)
	almostSevenDaysAgo := time.Now().AddDate(0, 0, -6).Add(-23 * time.Hour).Format(time.RFC3339)
	err = ValidateTimeRange(almostSevenDaysAgo, map[string]interface{}{})

	// This should be valid (less than 7 days ago)
	if err != nil {
		t.Errorf("Expected time less than 7 days ago to be valid, got error: %v", err)
	}

	// Test exactly at the 24-hour boundary
	start := "2025-07-15T10:00:00Z"
	exactly24HoursLater := "2025-07-16T10:00:00Z"

	err = ValidateTimeRange(start, map[string]interface{}{
		"end_time": exactly24HoursLater,
	})

	// This should be valid (exactly 24 hours)
	if err != nil {
		t.Errorf("Expected exactly 24-hour range to be valid, got error: %v", err)
	}
}
