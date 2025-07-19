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

func TestValidateKQLQuery_DangerousKeywords(t *testing.T) {
	dangerousQueries := []string{
		"DELETE FROM requests",
		"DROP TABLE requests",
		"CREATE TABLE test",
		"ALTER TABLE requests",
		"INSERT INTO requests",
		"UPDATE requests SET duration = 0",
		"TRUNCATE TABLE requests",
	}

	for _, query := range dangerousQueries {
		t.Run(query, func(t *testing.T) {
			err := validateKQLQuery(query)
			if err == nil {
				t.Errorf("Expected error for dangerous query '%s', got nil", query)
			}
		})
	}
}

func TestValidateKQLQuery_ValidQueries(t *testing.T) {
	validQueries := []string{
		"requests | where timestamp > ago(1h) | limit 10",
		"dependencies | summarize count() by type",
		"exceptions | where timestamp > ago(1d)",
		"traces | project timestamp, message",
		"customevents | limit 100",
		"pageviews | where timestamp > ago(1h)",
		"union requests, dependencies",
		"let timeRange = ago(1h); requests | where timestamp > timeRange",
	}

	for _, query := range validQueries {
		t.Run(query, func(t *testing.T) {
			err := validateKQLQuery(query)
			if err != nil {
				t.Errorf("Expected no error for valid query '%s', got: %v", query, err)
			}
		})
	}
}

func TestValidateKQLQuery_InvalidTableNames(t *testing.T) {
	invalidQueries := []string{
		"invalid_table | limit 10",
		"some_random_table | where timestamp > ago(1h)",
		"fake_data | project *",
	}

	for _, query := range invalidQueries {
		t.Run(query, func(t *testing.T) {
			err := validateKQLQuery(query)
			if err == nil {
				t.Errorf("Expected error for invalid table query '%s', got nil", query)
			}
		})
	}
}

func TestRegisterAppInsightsQueryTool(t *testing.T) {
	tool := RegisterAppInsightsQueryTool()

	if tool.Name != "az_monitor_app_insights_query" {
		t.Errorf("Expected tool name 'az_monitor_app_insights_query', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected tool description to be non-empty")
	}

	// Check that we have input schema
	if tool.InputSchema.Properties == nil {
		t.Error("Expected tool to have input schema properties")
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
