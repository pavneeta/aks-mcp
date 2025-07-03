package resourcehandlers

import (
	"testing"
)

func stringPtr(s string) *string {
	return &s
}

func TestExtractAKSParameters(t *testing.T) {
	tests := []struct {
		name           string
		params         map[string]interface{}
		expectedSubID  string
		expectedRG     string
		expectedCluster string
		expectError    bool
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"subscription_id": "sub-123",
				"resource_group":  "rg-test",
				"cluster_name":    "cluster-test",
			},
			expectedSubID:   "sub-123",
			expectedRG:      "rg-test",
			expectedCluster: "cluster-test",
			expectError:     false,
		},
		{
			name: "missing subscription_id",
			params: map[string]interface{}{
				"resource_group": "rg-test",
				"cluster_name":   "cluster-test",
			},
			expectError: true,
		},
		{
			name: "empty subscription_id",
			params: map[string]interface{}{
				"subscription_id": "",
				"resource_group":  "rg-test",
				"cluster_name":    "cluster-test",
			},
			expectError: true,
		},
		{
			name: "missing resource_group",
			params: map[string]interface{}{
				"subscription_id": "sub-123",
				"cluster_name":    "cluster-test",
			},
			expectError: true,
		},
		{
			name: "missing cluster_name",
			params: map[string]interface{}{
				"subscription_id": "sub-123",
				"resource_group":  "rg-test",
			},
			expectError: true,
		},
		{
			name: "invalid parameter types",
			params: map[string]interface{}{
				"subscription_id": 123,
				"resource_group":  "rg-test",
				"cluster_name":    "cluster-test",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subID, rg, clusterName, err := ExtractAKSParameters(tt.params)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if subID != tt.expectedSubID {
					t.Errorf("Expected subscription ID %s, got %s", tt.expectedSubID, subID)
				}
				if rg != tt.expectedRG {
					t.Errorf("Expected resource group %s, got %s", tt.expectedRG, rg)
				}
				if clusterName != tt.expectedCluster {
					t.Errorf("Expected cluster name %s, got %s", tt.expectedCluster, clusterName)
				}
			}
		})
	}
}
