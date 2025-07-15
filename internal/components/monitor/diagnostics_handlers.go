package monitor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/aks-mcp/internal/azcli"
	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/tools"
)

// Control Plane Diagnostics Handler Functions

// HandleControlPlaneDiagnosticSettings checks diagnostic settings for AKS cluster
func HandleControlPlaneDiagnosticSettings(params map[string]interface{}, cfg *config.ConfigData) (string, error) {
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

	// Build cluster resource ID
	clusterResourceID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerService/managedClusters/%s",
		subscriptionID, resourceGroup, clusterName)

	// Execute Azure CLI command to get diagnostic settings
	executor := azcli.NewExecutor()
	args := []string{
		"monitor", "diagnostic-settings", "list",
		"--resource", clusterResourceID,
		"--output", "json",
	}

	cmdParams := map[string]interface{}{
		"command": "az " + strings.Join(args, " "),
	}

	result, err := executor.Execute(cmdParams, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get diagnostic settings: %w", err)
	}

	// Return raw JSON result from Azure CLI
	return result, nil
}

// HandleControlPlaneLogs queries specific control plane logs
func HandleControlPlaneLogs(params map[string]interface{}, cfg *config.ConfigData) (string, error) {
	// Extract and validate all parameters
	subscriptionID, _ := params["subscription_id"].(string)
	resourceGroup, _ := params["resource_group"].(string)
	clusterName, _ := params["cluster_name"].(string)
	logCategory, _ := params["log_category"].(string)
	startTime, _ := params["start_time"].(string)
	endTime, _ := params["end_time"].(string)
	maxRecords := getMaxRecords(params)
	logLevel, _ := params["log_level"].(string)

	// Validate parameters
	if err := validateControlPlaneLogsParams(params); err != nil {
		return "", err
	}

	// Get workspace GUID from diagnostic settings
	workspaceGUID, err := extractWorkspaceGUIDFromDiagnosticSettings(subscriptionID, resourceGroup, clusterName, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace GUID: %w", err)
	}

	// Build cluster resource ID for scoping
	clusterResourceID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerService/managedClusters/%s",
		subscriptionID, resourceGroup, clusterName)

	// Build safe KQL query scoped to this specific AKS cluster
	kqlQuery := buildSafeKQLQuery(logCategory, logLevel, maxRecords, clusterResourceID)

	// Calculate timespan for the query
	timespan, err := calculateTimespan(startTime, endTime)
	if err != nil {
		return "", fmt.Errorf("failed to calculate timespan: %w", err)
	}

	// Execute log query with properly quoted KQL
	executor := azcli.NewExecutor()
	
	// Build command string with proper quoting for the KQL query
	cmd := fmt.Sprintf("az monitor log-analytics query --workspace %s --analytics-query \"%s\" --timespan %s --output json",
		workspaceGUID, kqlQuery, timespan)

	cmdParams := map[string]interface{}{
		"command": cmd,
	}

	result, err := executor.Execute(cmdParams, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to query control plane logs: %w", err)
	}

	// Return raw JSON result from Azure CLI
	return result, nil
}

// Helper functions for control plane diagnostics

func validateControlPlaneLogsParams(params map[string]interface{}) error {
	// Validate required parameters
	required := []string{"subscription_id", "resource_group", "cluster_name", "log_category", "start_time"}
	for _, param := range required {
		if value, ok := params[param].(string); !ok || value == "" {
			return fmt.Errorf("missing or invalid %s parameter", param)
		}
	}

	// Validate log category
	logCategory := params["log_category"].(string)
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

	valid := false
	for _, validCat := range validCategories {
		if logCategory == validCat {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log category: %s. Valid categories: %s", logCategory, strings.Join(validCategories, ", "))
	}

	// Validate time range
	startTime := params["start_time"].(string)
	if err := validateTimeRange(startTime, params); err != nil {
		return err
	}

	// Validate log level if provided
	if logLevel, ok := params["log_level"].(string); ok && logLevel != "" {
		validLevels := []string{"error", "warning", "info"}
		validLevel := false
		for _, level := range validLevels {
			if logLevel == level {
				validLevel = true
				break
			}
		}
		if !validLevel {
			return fmt.Errorf("invalid log level: %s. Valid levels: %s", logLevel, strings.Join(validLevels, ", "))
		}
	}

	return nil
}

func validateTimeRange(startTime string, params map[string]interface{}) error {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return fmt.Errorf("invalid start_time format, expected RFC3339 (ISO 8601): %w", err)
	}

	// Check if start time is not more than 7 days ago
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if start.Before(sevenDaysAgo) {
		return fmt.Errorf("start_time cannot be more than 7 days ago")
	}

	// Check if start time is in the future
	if start.After(time.Now()) {
		return fmt.Errorf("start_time cannot be in the future")
	}

	// Validate end time if provided
	if endTimeStr, ok := params["end_time"].(string); ok && endTimeStr != "" {
		end, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return fmt.Errorf("invalid end_time format, expected RFC3339 (ISO 8601): %w", err)
		}

		// Check if time range is not more than 24 hours
		if end.Sub(start) > 24*time.Hour {
			return fmt.Errorf("time range cannot exceed 24 hours")
		}

		if end.Before(start) {
			return fmt.Errorf("end_time must be after start_time")
		}

		if end.After(time.Now()) {
			return fmt.Errorf("end_time cannot be in the future")
		}
	}

	return nil
}

func extractWorkspaceGUIDFromDiagnosticSettings(subscriptionID, resourceGroup, clusterName string, cfg *config.ConfigData) (string, error) {
	// Get diagnostic settings
	params := map[string]interface{}{
		"subscription_id": subscriptionID,
		"resource_group":  resourceGroup,
		"cluster_name":    clusterName,
	}

	diagnosticResult, err := HandleControlPlaneDiagnosticSettings(params, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get diagnostic settings: %w", err)
	}

	// Parse to extract workspace ID
	var parsed interface{}
	if err := json.Unmarshal([]byte(diagnosticResult), &parsed); err != nil {
		return "", fmt.Errorf("failed to parse diagnostic settings JSON: %w", err)
	}

	// Handle both array and object formats
	var settings []interface{}
	
	// Check if it's an array (direct diagnostic settings response)
	if settingsArray, ok := parsed.([]interface{}); ok {
		settings = settingsArray
	} else if parsedObj, ok := parsed.(map[string]interface{}); ok {
		// Check if it's wrapped in a "value" property
		if value, ok := parsedObj["value"].([]interface{}); ok {
			settings = value
		}
	}

	// Extract workspace resource ID from the first diagnostic setting
	if len(settings) > 0 {
		if setting, ok := settings[0].(map[string]interface{}); ok {
			if workspaceResourceID, ok := setting["workspaceId"].(string); ok && workspaceResourceID != "" {
				// Extract workspace GUID from the workspace resource ID
				return getWorkspaceGUID(workspaceResourceID, cfg)
			}
		}
	}

	return "", fmt.Errorf("no Log Analytics workspace found in diagnostic settings")
}

func buildSafeKQLQuery(category, logLevel string, maxRecords int, clusterResourceID string) string {
	// Build pre-validated KQL queries to prevent injection, scoped to specific AKS cluster
	// Convert resource ID to uppercase as it's stored in uppercase in Log Analytics
	upperResourceID := strings.ToUpper(clusterResourceID)
	baseQuery := fmt.Sprintf("AzureDiagnostics | where Category == '%s' and ResourceId == '%s'", category, upperResourceID)

	if logLevel != "" {
		// Filter by log level embedded in the log message itself
		// Kubernetes logs use format like "I0715" (Info), "W0715" (Warning), "E0715" (Error)
		var levelPrefix string
		switch strings.ToLower(logLevel) {
		case "info":
			levelPrefix = "I"
		case "warning":
			levelPrefix = "W"  
		case "error":
			levelPrefix = "E"
		}
		if levelPrefix != "" {
			baseQuery += fmt.Sprintf(" | where log_s startswith '%s'", levelPrefix)
		}
	}

	baseQuery += " | order by TimeGenerated desc"
	baseQuery += fmt.Sprintf(" | limit %d", maxRecords)
	
	// Project only essential fields: log content, timestamp, and level
	baseQuery += " | project TimeGenerated, Level, log_s"

	return baseQuery
}

func getMaxRecords(params map[string]interface{}) int {
	if val, ok := params["max_records"].(string); ok && val != "" {
		if recordsInt, err := strconv.Atoi(val); err == nil {
			if recordsInt > 1000 {
				return 1000
			}
			if recordsInt < 1 {
				return 100
			}
			return recordsInt
		}
	}
	return 100
}

// calculateTimespan converts start/end times to Azure CLI timespan format
func calculateTimespan(startTime, endTime string) (string, error) {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return "", fmt.Errorf("invalid start time format: %w", err)
	}
	
	var end time.Time
	if endTime != "" {
		end, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			return "", fmt.Errorf("invalid end time format: %w", err)
		}
	} else {
		// Default to current time if no end time specified
		end = time.Now()
	}
	
	// Azure CLI timespan format: start_time/end_time in ISO8601
	timespan := fmt.Sprintf("%s/%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
	return timespan, nil
}

// getWorkspaceGUID extracts the workspace GUID from a workspace resource ID
func getWorkspaceGUID(workspaceResourceID string, cfg *config.ConfigData) (string, error) {
	// Parse the workspace resource ID to extract resource group and workspace name
	// Format: /subscriptions/{sub}/resourcegroups/{rg}/providers/microsoft.operationalinsights/workspaces/{workspace-name}
	parts := strings.Split(workspaceResourceID, "/")
	if len(parts) < 8 {
		return "", fmt.Errorf("invalid workspace resource ID format: %s", workspaceResourceID)
	}
	
	var resourceGroup, workspaceName string
	for i, part := range parts {
		if strings.ToLower(part) == "resourcegroups" && i+1 < len(parts) {
			resourceGroup = parts[i+1]
		}
		if strings.ToLower(part) == "workspaces" && i+1 < len(parts) {
			workspaceName = parts[i+1]
		}
	}
	
	if resourceGroup == "" || workspaceName == "" {
		return "", fmt.Errorf("could not extract resource group and workspace name from: %s", workspaceResourceID)
	}
	
	// Query the workspace to get its GUID (customerId)
	executor := azcli.NewExecutor()
	cmd := fmt.Sprintf("az monitor log-analytics workspace show --resource-group %s --workspace-name %s --query customerId --output tsv", resourceGroup, workspaceName)
	
	cmdParams := map[string]interface{}{
		"command": cmd,
	}
	
	result, err := executor.Execute(cmdParams, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace GUID: %w", err)
	}
	
	// The result should be the workspace GUID, trim any whitespace
	workspaceGUID := strings.TrimSpace(result)
	if workspaceGUID == "" {
		return "", fmt.Errorf("empty workspace GUID returned for workspace: %s", workspaceName)
	}
	
	return workspaceGUID, nil
}

// Resource handler functions for control plane diagnostics tools

// GetControlPlaneDiagnosticSettingsHandler returns handler for diagnostic settings tool
func GetControlPlaneDiagnosticSettingsHandler(cfg *config.ConfigData) tools.ResourceHandler {
	return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		return HandleControlPlaneDiagnosticSettings(params, cfg)
	})
}

// GetControlPlaneLogsHandler returns handler for logs querying tool
func GetControlPlaneLogsHandler(cfg *config.ConfigData) tools.ResourceHandler {
	return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		return HandleControlPlaneLogs(params, cfg)
	})
}
