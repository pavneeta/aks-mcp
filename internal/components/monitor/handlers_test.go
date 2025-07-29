package monitor

import (
	"testing"
)

func TestHandleAppInsightsQuery_ValidParameters(t *testing.T) {
	params := map[string]interface{}{
		"subscription_id":   "test-subscription",
		"resource_group":    "test-resource-group",
		"app_insights_name": "test-app-insights",
		"query":             "requests | where timestamp > ago(1h) | limit 10",
	}

	// This would normally execute the Azure CLI command, but since we don't have
	// Azure CLI available in test, we just validate that parameters are processed correctly
	err := validateAppInsightsParams(params)
	if err != nil {
		t.Errorf("Expected no error for valid parameters, got: %v", err)
	}
}

func TestHandleAppInsightsQuery_MissingParameters(t *testing.T) {
	testCases := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name: "missing subscription_id",
			params: map[string]interface{}{
				"resource_group":    "test-rg",
				"app_insights_name": "test-ai",
				"query":             "requests | limit 10",
			},
		},
		{
			name: "missing resource_group",
			params: map[string]interface{}{
				"subscription_id":   "test-sub",
				"app_insights_name": "test-ai",
				"query":             "requests | limit 10",
			},
		},
		{
			name: "missing app_insights_name",
			params: map[string]interface{}{
				"subscription_id": "test-sub",
				"resource_group":  "test-rg",
				"query":           "requests | limit 10",
			},
		},
		{
			name: "missing query",
			params: map[string]interface{}{
				"subscription_id":   "test-sub",
				"resource_group":    "test-rg",
				"app_insights_name": "test-ai",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAppInsightsParams(tc.params)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestValidateAppInsightsParams_TimeValidation(t *testing.T) {
	testCases := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid RFC3339 start_time",
			params: map[string]interface{}{
				"subscription_id":   "test-sub",
				"resource_group":    "test-rg",
				"app_insights_name": "test-ai",
				"query":             "requests | limit 10",
				"start_time":        "2025-01-01T00:00:00Z",
			},
			expectError: false,
		},
		{
			name: "invalid start_time format",
			params: map[string]interface{}{
				"subscription_id":   "test-sub",
				"resource_group":    "test-rg",
				"app_insights_name": "test-ai",
				"query":             "requests | limit 10",
				"start_time":        "invalid-time",
			},
			expectError: true,
		},
		{
			name: "valid timespan",
			params: map[string]interface{}{
				"subscription_id":   "test-sub",
				"resource_group":    "test-rg",
				"app_insights_name": "test-ai",
				"query":             "requests | limit 10",
				"timespan":          "PT1H",
			},
			expectError: false,
		},
		{
			name: "invalid timespan format",
			params: map[string]interface{}{
				"subscription_id":   "test-sub",
				"resource_group":    "test-rg",
				"app_insights_name": "test-ai",
				"query":             "requests | limit 10",
				"timespan":          "1hour",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAppInsightsParams(tc.params)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tc.name, err)
			}
		})
	}
}
