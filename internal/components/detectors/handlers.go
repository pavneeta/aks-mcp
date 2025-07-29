package detectors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/aks-mcp/internal/azureclient"
	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/tools"
)

// =============================================================================
// Detector-related Handlers
// =============================================================================

// GetListDetectorsHandler returns handler for list_detectors tool
func GetListDetectorsHandler(azClient *azureclient.AzureClient, cfg *config.ConfigData) tools.ResourceHandler {
	return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		return HandleListDetectors(params, NewDetectorClient(azClient))
	})
}

// GetRunDetectorHandler returns handler for run_detector tool
func GetRunDetectorHandler(azClient *azureclient.AzureClient, cfg *config.ConfigData) tools.ResourceHandler {
	return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		return HandleRunDetector(params, NewDetectorClient(azClient))
	})
}

// GetRunDetectorsByCategoryHandler returns handler for run_detectors_by_category tool
func GetRunDetectorsByCategoryHandler(azClient *azureclient.AzureClient, cfg *config.ConfigData) tools.ResourceHandler {
	return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		return HandleRunDetectorsByCategory(params, NewDetectorClient(azClient))
	})
}

// =============================================================================
// Handler Implementation Functions
// =============================================================================

// HandleListDetectors implements the list_detectors functionality
func HandleListDetectors(params map[string]interface{}, client *DetectorClient) (string, error) {
	// Extract cluster resource ID
	clusterResourceID, ok := params["cluster_resource_id"].(string)
	if !ok || clusterResourceID == "" {
		return "", fmt.Errorf("missing or invalid cluster_resource_id parameter")
	}

	// Parse resource ID
	subscriptionID, resourceGroup, clusterName, err := azureclient.ParseAKSResourceID(clusterResourceID)
	if err != nil {
		return "", fmt.Errorf("failed to parse cluster resource ID: %v", err)
	}

	// List detectors
	ctx := context.Background()
	detectors, err := client.ListDetectors(ctx, subscriptionID, resourceGroup, clusterName)
	if err != nil {
		return "", fmt.Errorf("failed to list detectors: %v", err)
	}

	// Return as JSON
	resultJSON, err := json.MarshalIndent(detectors, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal detectors to JSON: %v", err)
	}

	return string(resultJSON), nil
}

// HandleRunDetector implements the run_detector functionality
func HandleRunDetector(params map[string]interface{}, client *DetectorClient) (string, error) {
	// Extract cluster resource ID
	clusterResourceID, ok := params["cluster_resource_id"].(string)
	if !ok || clusterResourceID == "" {
		return "", fmt.Errorf("missing or invalid cluster_resource_id parameter")
	}

	// Extract detector name
	detectorName, ok := params["detector_name"].(string)
	if !ok || detectorName == "" {
		return "", fmt.Errorf("missing or invalid detector_name parameter")
	}

	// Extract start time
	startTime, ok := params["start_time"].(string)
	if !ok || startTime == "" {
		return "", fmt.Errorf("missing or invalid start_time parameter")
	}

	// Extract end time
	endTime, ok := params["end_time"].(string)
	if !ok || endTime == "" {
		return "", fmt.Errorf("missing or invalid end_time parameter")
	}

	// Validate time format and constraints
	if err := validateTimeParameters(startTime, endTime); err != nil {
		return "", fmt.Errorf("invalid time parameters: %v", err)
	}

	// Parse resource ID
	subscriptionID, resourceGroup, clusterName, err := azureclient.ParseAKSResourceID(clusterResourceID)
	if err != nil {
		return "", fmt.Errorf("failed to parse cluster resource ID: %v", err)
	}

	// Run detector
	ctx := context.Background()
	result, err := client.RunDetector(ctx, subscriptionID, resourceGroup, clusterName, detectorName, startTime, endTime)
	if err != nil {
		return "", fmt.Errorf("failed to run detector: %v", err)
	}

	// Return as JSON
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal detector result to JSON: %v", err)
	}

	return string(resultJSON), nil
}

// HandleRunDetectorsByCategory implements the run_detectors_by_category functionality
func HandleRunDetectorsByCategory(params map[string]interface{}, client *DetectorClient) (string, error) {
	// Extract cluster resource ID
	clusterResourceID, ok := params["cluster_resource_id"].(string)
	if !ok || clusterResourceID == "" {
		return "", fmt.Errorf("missing or invalid cluster_resource_id parameter")
	}

	// Extract category
	category, ok := params["category"].(string)
	if !ok || category == "" {
		return "", fmt.Errorf("missing or invalid category parameter")
	}

	// Extract start time
	startTime, ok := params["start_time"].(string)
	if !ok || startTime == "" {
		return "", fmt.Errorf("missing or invalid start_time parameter")
	}

	// Extract end time
	endTime, ok := params["end_time"].(string)
	if !ok || endTime == "" {
		return "", fmt.Errorf("missing or invalid end_time parameter")
	}

	// Validate time parameters
	if err := validateTimeParameters(startTime, endTime); err != nil {
		return "", fmt.Errorf("invalid time parameters: %v", err)
	}

	// Validate category
	if err := validateCategory(category); err != nil {
		return "", fmt.Errorf("invalid category: %v", err)
	}

	// Parse resource ID
	subscriptionID, resourceGroup, clusterName, err := azureclient.ParseAKSResourceID(clusterResourceID)
	if err != nil {
		return "", fmt.Errorf("failed to parse cluster resource ID: %v", err)
	}

	// Run detectors by category
	ctx := context.Background()
	results, err := client.RunDetectorsByCategory(ctx, subscriptionID, resourceGroup, clusterName, category, startTime, endTime)
	if err != nil {
		return "", fmt.Errorf("failed to run detectors by category: %v", err)
	}

	// Create response with metadata
	response := map[string]interface{}{
		"category":        category,
		"detectors_count": len(results),
		"results":         results,
	}

	// Return as JSON
	resultJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal detector results to JSON: %v", err)
	}

	return string(resultJSON), nil
}

// =============================================================================
// Validation Helper Functions
// =============================================================================

// validateTimeParameters validates start and end time parameters
func validateTimeParameters(startTime, endTime string) error {
	// Parse times
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return fmt.Errorf("invalid start_time format, expected ISO 8601 (RFC3339): %v", err)
	}

	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return fmt.Errorf("invalid end_time format, expected ISO 8601 (RFC3339): %v", err)
	}

	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Check if times are within last 30 days
	if start.Before(thirtyDaysAgo) {
		return fmt.Errorf("start_time must be within the last 30 days")
	}

	if end.Before(thirtyDaysAgo) {
		return fmt.Errorf("end_time must be within the last 30 days")
	}

	// Check if end time is after start time
	if end.Before(start) || end.Equal(start) {
		return fmt.Errorf("end_time must be after start_time")
	}

	// Check if duration is max 24 hours
	if end.Sub(start) > 24*time.Hour {
		return fmt.Errorf("time range cannot exceed 24 hours")
	}

	return nil
}

// validateCategory validates the category parameter
func validateCategory(category string) error {
	validCategories := []string{
		"Best Practices",
		"Cluster and Control Plane Availability and Performance",
		"Connectivity Issues",
		"Create, Upgrade, Delete and Scale",
		"Deprecations",
		"Identity and Security",
		"Node Health",
		"Storage",
	}

	for _, valid := range validCategories {
		if strings.EqualFold(category, valid) {
			return nil
		}
	}

	return fmt.Errorf("invalid category '%s', must be one of: %v", category, validCategories)
}
