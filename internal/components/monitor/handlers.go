package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/Azure/aks-mcp/internal/azcli"
	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/tools"
)

// HandleResourceHealthQuery handles the resource health query for AKS clusters
func HandleResourceHealthQuery(params map[string]interface{}, cfg *config.ConfigData) (string, error) {
	// Extract and validate parameters
	subscriptionID, ok := params["subscription_id"].(string)
	if !ok || subscriptionID == "" {
		return "", fmt.Errorf("missing or invalid subscription_id parameter")
	}

	resourceGroup, ok := params["resource_group"].(string)
	if !ok || resourceGroup == "" {
		return "", fmt.Errorf("missing or invalid resource_group parameter")
	}

	clusterName, ok := params["cluster_name"].(string)
	if !ok || clusterName == "" {
		return "", fmt.Errorf("missing or invalid cluster_name parameter")
	}

	startTime, ok := params["start_time"].(string)
	if !ok || startTime == "" {
		return "", fmt.Errorf("missing or invalid start_time parameter")
	}

	// Validate parameters
	if err := validateResourceHealthParams(params); err != nil {
		return "", err
	}

	// Build resource ID
	resourceID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerService/managedClusters/%s",
		subscriptionID, resourceGroup, clusterName)

	// Build Azure CLI command
	executor := azcli.NewExecutor()
	args := []string{
		"monitor", "activity-log", "list",
		"--resource-id", resourceID,
		"--start-time", startTime,
		"--query", "[?category.value=='ResourceHealth']",
		"--output", "json",
	}

	// Add end time if provided
	if endTime, ok := params["end_time"].(string); ok && endTime != "" {
		args = append(args, "--end-time", endTime)
	}

	// Add status filter if provided
	if status, ok := params["status"].(string); ok && status != "" {
		// Apply status filter in the query
		statusFilter := fmt.Sprintf("[?category.value=='ResourceHealth' && properties.currentHealthStatus=='%s']", status)
		args[len(args)-3] = statusFilter // Replace the query parameter
	}

	// Execute command
	cmdParams := map[string]interface{}{
		"command": "az " + strings.Join(args, " "),
	}

	result, err := executor.Execute(cmdParams, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to execute resource health query: %w", err)
	}

	// Return the raw JSON result from Azure CLI
	return result, nil
}

// validateResourceHealthParams validates the parameters for resource health queries
func validateResourceHealthParams(params map[string]interface{}) error {
	// Validate required parameters
	required := []string{"subscription_id", "resource_group", "cluster_name", "start_time"}
	for _, param := range required {
		if value, ok := params[param].(string); !ok || value == "" {
			return fmt.Errorf("missing or invalid %s parameter", param)
		}
	}

	// Validate time format
	startTime := params["start_time"].(string)
	if _, err := time.Parse(time.RFC3339, startTime); err != nil {
		return fmt.Errorf("invalid start_time format, expected RFC3339 (ISO 8601): %w", err)
	}

	// Validate end_time if provided
	if endTime, ok := params["end_time"].(string); ok && endTime != "" {
		if _, err := time.Parse(time.RFC3339, endTime); err != nil {
			return fmt.Errorf("invalid end_time format, expected RFC3339 (ISO 8601): %w", err)
		}
	}

	// Validate status if provided
	if status, ok := params["status"].(string); ok && status != "" {
		validStatuses := []string{"Available", "Unavailable", "Degraded", "Unknown"}
		valid := false
		for _, validStatus := range validStatuses {
			if status == validStatus {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid status parameter, must be one of: %s", strings.Join(validStatuses, ", "))
		}
	}

	return nil
}

// GetResourceHealthHandler returns a ResourceHandler for the resource health tool
func GetResourceHealthHandler(cfg *config.ConfigData) tools.ResourceHandler {
	return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		return HandleResourceHealthQuery(params, cfg)
	})
}
