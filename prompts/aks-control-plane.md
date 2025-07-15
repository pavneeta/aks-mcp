# AKS Control Plane Logs for AKS-MCP

## Overview
This document outlines the implementation plan for AKS control plane logs functionality that provides a safe, workflow-based approach to accessing AKS diagnostic logs:
- `aks_control_plane_diagnostic_settings`: Check if diagnostic settings are configured for an AKS cluster
- `aks_control_plane_log_categories`: List enabled control plane log categories
- `aks_control_plane_logs`: Query specific control plane log categories with safety constraints

## Requirements

### Workflow-Based Approach
1. **Check Diagnostic Settings**: Verify if AKS cluster has diagnostic settings configured
2. **List Log Categories**: Show which control plane log categories are enabled
3. **Query Logs Safely**: Query specific log categories with time range and size validation

### Azure CLI Commands
- **Diagnostic Settings**: `az monitor diagnostic-settings list --resource {cluster-resource-id}`
- **Log Analytics Workspace**: `az monitor log-analytics workspace show --workspace {workspace-id}`
- **Log Query**: `az logs query --workspace {workspace-id} --analytics-query "{kql-query}" --start-time {start} --end-time {end}`

### AKS Control Plane Log Categories
- `kube-apiserver`: Kubernetes API server logs
- `kube-audit`: Kubernetes audit logs
- `kube-audit-admin`: Admin audit logs
- `kube-controller-manager`: Controller manager logs
- `kube-scheduler`: Scheduler logs
- `cluster-autoscaler`: Cluster autoscaler logs
- `cloud-controller-manager`: Cloud controller manager logs

### Safety Constraints
- **Time Range Validation**: Maximum 7 days lookback, maximum 24-hour query window
- **Record Limits**: Default 100 records, maximum 1000 records per query
- **Query Validation**: Pre-built KQL queries to prevent injection
- **Workspace Validation**: Ensure workspace exists and is accessible

## Implementation Architecture

### 1. File Structure
Following existing patterns in the codebase:

```
internal/components/monitor/
├── handlers.go              # Add new control plane handlers
├── registry.go              # Add new tool registrations
└── controlplane_test.go     # Unit tests for control plane functionality
```

### 2. Core Components

#### 2.1 Data Structures
No custom data structures needed - return raw Azure ARM API responses as JSON strings. This follows the pattern of existing tools like the resource health tool that returns raw Azure CLI JSON output.

#### 2.2 Tool Registration (`registry.go` additions)
```go
// RegisterControlPlaneDiagnosticSettingsTool registers the diagnostic settings checker tool
func RegisterControlPlaneDiagnosticSettingsTool() mcp.Tool {
    return mcp.NewTool("aks_control_plane_diagnostic_settings",
        mcp.WithDescription("Check if AKS cluster has diagnostic settings configured and identify the Log Analytics workspace"),
        mcp.WithString("subscription_id",
            mcp.Required(),
            mcp.Description("Azure subscription ID"),
        ),
        mcp.WithString("resource_group",
            mcp.Required(),
            mcp.Description("Resource group name containing the AKS cluster"),
        ),
        mcp.WithString("cluster_name",
            mcp.Required(),
            mcp.Description("AKS cluster name"),
        ),
    )
}

// RegisterControlPlaneLogCategoriesTool registers the log categories listing tool
func RegisterControlPlaneLogCategoriesTool() mcp.Tool {
    return mcp.NewTool("aks_control_plane_log_categories",
        mcp.WithDescription("List enabled AKS control plane log categories from diagnostic settings"),
        mcp.WithString("subscription_id",
            mcp.Required(),
            mcp.Description("Azure subscription ID"),
        ),
        mcp.WithString("resource_group",
            mcp.Required(),
            mcp.Description("Resource group name containing the AKS cluster"),
        ),
        mcp.WithString("cluster_name",
            mcp.Required(),
            mcp.Description("AKS cluster name"),
        ),
    )
}

// RegisterControlPlaneLogsTool registers the logs querying tool
func RegisterControlPlaneLogsTool() mcp.Tool {
    return mcp.NewTool("aks_control_plane_logs",
        mcp.WithDescription("Query AKS control plane logs with safety constraints and time range validation"),
        mcp.WithString("subscription_id",
            mcp.Required(),
            mcp.Description("Azure subscription ID"),
        ),
        mcp.WithString("resource_group",
            mcp.Required(),
            mcp.Description("Resource group name containing the AKS cluster"),
        ),
        mcp.WithString("cluster_name",
            mcp.Required(),
            mcp.Description("AKS cluster name"),
        ),
        mcp.WithString("log_category",
            mcp.Required(),
            mcp.Description("Control plane log category (kube-apiserver, kube-audit, kube-controller-manager, kube-scheduler, cluster-autoscaler, cloud-controller-manager)"),
        ),
        mcp.WithString("start_time",
            mcp.Required(),
            mcp.Description("Start time in ISO 8601 format (max 7 days ago, e.g., '2025-07-14T00:00:00Z')"),
        ),
        mcp.WithString("end_time",
            mcp.Description("End time in ISO 8601 format (defaults to now, max 24 hours from start_time)"),
        ),
        mcp.WithInteger("max_records",
            mcp.Description("Maximum number of log records to return (default: 100, max: 1000)"),
        ),
        mcp.WithString("log_level",
            mcp.Description("Filter by log level (error, warning, info) - optional"),
        ),
    )
}
```

#### 2.2 Handlers (`handlers.go` additions)
```go
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

// HandleControlPlaneLogCategories lists enabled log categories
func HandleControlPlaneLogCategories(params map[string]interface{}, cfg *config.ConfigData) (string, error) {
    // First get diagnostic settings (raw JSON)
    diagnosticResult, err := HandleControlPlaneDiagnosticSettings(params, cfg)
    if err != nil {
        return "", err
    }

    // Parse only to extract enabled log categories, then return simplified JSON
    return extractAndFormatLogCategories(diagnosticResult)
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

    // Get workspace ID from diagnostic settings
    workspaceID, err := extractWorkspaceIDFromDiagnosticSettings(subscriptionID, resourceGroup, clusterName, cfg)
    if err != nil {
        return "", fmt.Errorf("failed to get workspace ID: %w", err)
    }

    // Build safe KQL query
    kqlQuery := buildSafeKQLQuery(logCategory, logLevel, maxRecords)

    // Execute log query
    executor := azcli.NewExecutor()
    args := []string{
        "logs", "query",
        "--workspace", workspaceID,
        "--analytics-query", kqlQuery,
        "--start-time", startTime,
        "--output", "json",
    }

    if endTime != "" {
        args = append(args, "--end-time", endTime)
    }

    cmdParams := map[string]interface{}{
        "command": "az " + strings.Join(args, " "),
    }

    result, err := executor.Execute(cmdParams, cfg)
    if err != nil {
        return "", fmt.Errorf("failed to query control plane logs: %w", err)
    }

    // Return raw JSON result from Azure CLI
    return result, nil
}

// Helper functions
func extractAndFormatLogCategories(diagnosticSettingsJSON string) (string, error) {
    // Parse diagnostic settings JSON to extract just the enabled log categories
    // Return simplified JSON with category names and enabled status
    // No custom structs - just JSON manipulation
}

func validateControlPlaneLogsParams(params map[string]interface{}) error {
    // Validate all parameters including time ranges, log categories, etc.
    // Same validation logic but no custom structs
}

func extractWorkspaceIDFromDiagnosticSettings(subscriptionID, resourceGroup, clusterName string, cfg *config.ConfigData) (string, error) {
    // Get diagnostic settings and extract workspace ID using JSON parsing
    // No custom structs needed
}

func buildSafeKQLQuery(category, logLevel string, maxRecords int) string {
    // Build pre-validated KQL queries to prevent injection
    baseQuery := fmt.Sprintf("AzureDiagnostics | where Category == '%s'", category)
    
    if logLevel != "" {
        baseQuery += fmt.Sprintf(" | where Level == '%s'", logLevel)
    }
    
    baseQuery += " | order by TimeGenerated desc"
    baseQuery += fmt.Sprintf(" | limit %d", maxRecords)
    
    return baseQuery
}

func getMaxRecords(params map[string]interface{}) int {
    if val, ok := params["max_records"].(float64); ok {
        if int(val) > 1000 {
            return 1000
        }
        if int(val) < 1 {
            return 100
        }
        return int(val)
    }
    return 100
}
```

#### 2.3 Resource Handler Functions
```go
// GetControlPlaneDiagnosticSettingsHandler returns handler for diagnostic settings tool
func GetControlPlaneDiagnosticSettingsHandler(cfg *config.ConfigData) tools.ResourceHandler {
    return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
        return HandleControlPlaneDiagnosticSettings(params, cfg)
    })
}

// GetControlPlaneLogCategoriesHandler returns handler for log categories tool
func GetControlPlaneLogCategoriesHandler(cfg *config.ConfigData) tools.ResourceHandler {
    return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
        return HandleControlPlaneLogCategories(params, cfg)
    })
}

// GetControlPlaneLogsHandler returns handler for logs querying tool
func GetControlPlaneLogsHandler(cfg *config.ConfigData) tools.ResourceHandler {
    return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
        return HandleControlPlaneLogs(params, cfg)
    })
}
```

### 3. Integration Points

#### 3.1 Server Registration
Update `internal/server/server.go` to register control plane tools:

```go
// Add to registerAzCommands method
func (s *Service) registerControlPlaneTools() {
    log.Println("Registering Control Plane tools...")

    // Register diagnostic settings tool
    log.Println("Registering control plane tool: aks_control_plane_diagnostic_settings")
    diagnosticTool := monitor.RegisterControlPlaneDiagnosticSettingsTool()
    s.mcpServer.AddTool(diagnosticTool, tools.CreateResourceHandler(monitor.GetControlPlaneDiagnosticSettingsHandler(s.cfg), s.cfg))

    // Register log categories tool
    log.Println("Registering control plane tool: aks_control_plane_log_categories")
    categoriesRool := monitor.RegisterControlPlaneLogCategoriesTool()
    s.mcpServer.AddTool(categoriesTool, tools.CreateResourceHandler(monitor.GetControlPlaneLogCategoriesHandler(s.cfg), s.cfg))

    // Register logs querying tool
    log.Println("Registering control plane tool: aks_control_plane_logs")
    logsTool := monitor.RegisterControlPlaneLogsTool()
    s.mcpServer.AddTool(logsTool, tools.CreateResourceHandler(monitor.GetControlPlaneLogsHandler(s.cfg), s.cfg))
}
```

### 4. Tool Definitions

#### 4.1 aks_control_plane_diagnostic_settings Tool
```json
{
  "name": "aks_control_plane_diagnostic_settings",
  "description": "Check if AKS cluster has diagnostic settings configured and identify the Log Analytics workspace",
  "inputSchema": {
    "type": "object",
    "properties": {
      "subscription_id": {
        "type": "string",
        "description": "Azure subscription ID"
      },
      "resource_group": {
        "type": "string",
        "description": "Resource group name containing the AKS cluster"
      },
      "cluster_name": {
        "type": "string",
        "description": "AKS cluster name"
      }
    },
    "required": ["subscription_id", "resource_group", "cluster_name"]
  }
}
```

#### 4.2 aks_control_plane_log_categories Tool
```json
{
  "name": "aks_control_plane_log_categories",
  "description": "List enabled AKS control plane log categories from diagnostic settings",
  "inputSchema": {
    "type": "object",
    "properties": {
      "subscription_id": {
        "type": "string",
        "description": "Azure subscription ID"
      },
      "resource_group": {
        "type": "string",
        "description": "Resource group name containing the AKS cluster"
      },
      "cluster_name": {
        "type": "string",
        "description": "AKS cluster name"
      }
    },
    "required": ["subscription_id", "resource_group", "cluster_name"]
  }
}
```

#### 4.3 aks_control_plane_logs Tool
```json
{
  "name": "aks_control_plane_logs",
  "description": "Query AKS control plane logs with safety constraints and time range validation",
  "inputSchema": {
    "type": "object",
    "properties": {
      "subscription_id": {
        "type": "string",
        "description": "Azure subscription ID"
      },
      "resource_group": {
        "type": "string",
        "description": "Resource group name containing the AKS cluster"
      },
      "cluster_name": {
        "type": "string",
        "description": "AKS cluster name"
      },
      "log_category": {
        "type": "string",
        "enum": [
          "kube-apiserver",
          "kube-audit",
          "kube-audit-admin",
          "kube-controller-manager",
          "kube-scheduler",
          "cluster-autoscaler",
          "cloud-controller-manager"
        ],
        "description": "Control plane log category"
      },
      "start_time": {
        "type": "string",
        "description": "Start time in ISO 8601 format (max 7 days ago)"
      },
      "end_time": {
        "type": "string",
        "description": "End time in ISO 8601 format (max 24 hours from start_time)"
      },
      "max_records": {
        "type": "integer",
        "minimum": 1,
        "maximum": 1000,
        "default": 100,
        "description": "Maximum number of log records to return"
      },
      "log_level": {
        "type": "string",
        "enum": ["error", "warning", "info"],
        "description": "Filter by log level (optional)"
      }
    },
    "required": ["subscription_id", "resource_group", "cluster_name", "log_category", "start_time"]
  }
}
```

### 5. Implementation Flow

#### 5.1 Diagnostic Settings Check Flow
1. Parse cluster parameters (subscription, resource group, cluster name)
2. Construct cluster resource ID
3. Execute `az monitor diagnostic-settings list` command
4. Parse JSON response to extract workspace ID and enabled log categories
5. Return structured diagnostic settings information

#### 5.2 Log Categories List Flow
1. Call diagnostic settings check internally
2. Parse diagnostic settings response
3. Extract enabled log categories with descriptions
4. Return formatted list of available log categories

#### 5.3 Control Plane Logs Query Flow
1. Validate all input parameters (time ranges, log category, etc.)
2. Get workspace ID from diagnostic settings
3. Validate that requested log category is enabled
4. Build safe KQL query with constraints
5. Execute `az logs query` command
6. Parse and format results
7. Return log query results with metadata

### 6. Safety Validations

#### 6.1 Time Range Validation
```go
func validateTimeRange(startTime, endTime string) error {
    start, err := time.Parse(time.RFC3339, startTime)
    if err != nil {
        return fmt.Errorf("invalid start_time format: %w", err)
    }

    // Check if start time is not more than 7 days ago
    sevenDaysAgo := time.Now().AddDate(0, 0, -7)
    if start.Before(sevenDaysAgo) {
        return fmt.Errorf("start_time cannot be more than 7 days ago")
    }

    if endTime != "" {
        end, err := time.Parse(time.RFC3339, endTime)
        if err != nil {
            return fmt.Errorf("invalid end_time format: %w", err)
        }

        // Check if time range is not more than 24 hours
        if end.Sub(start) > 24*time.Hour {
            return fmt.Errorf("time range cannot exceed 24 hours")
        }

        if end.Before(start) {
            return fmt.Errorf("end_time must be after start_time")
        }
    }

    return nil
}
```

#### 6.2 Log Category Validation
```go
func validateLogCategory(category string, enabledCategories []string) error {
    validCategories := []string{
        "kube-apiserver",
        "kube-audit",
        "kube-audit-admin",
        "kube-controller-manager",
        "kube-scheduler",
        "cluster-autoscaler",
        "cloud-controller-manager",
    }

    // Check if category is valid
    valid := false
    for _, validCat := range validCategories {
        if category == validCat {
            valid = true
            break
        }
    }
    if !valid {
        return fmt.Errorf("invalid log category: %s", category)
    }

    // Check if category is enabled
    enabled := false
    for _, enabledCat := range enabledCategories {
        if category == enabledCat {
            enabled = true
            break
        }
    }
    if !enabled {
        return fmt.Errorf("log category '%s' is not enabled in diagnostic settings", category)
    }

    return nil
}
```

### 7. Error Handling

#### 7.1 Common Errors
- **No Diagnostic Settings**: Clear message that diagnostic settings need to be configured
- **Workspace Access**: Handle cases where workspace exists but user lacks access
- **Invalid Time Range**: Validate and provide helpful error messages
- **Category Not Enabled**: Inform user which categories are available
- **Query Timeout**: Handle Azure CLI timeouts gracefully

#### 7.2 Error Response Format
Return Azure CLI error messages directly, following the pattern of existing tools. No custom error structures needed.

### 8. Usage Examples

#### 8.1 Check Diagnostic Settings
```bash
# Check if cluster has diagnostic settings configured
aks_control_plane_diagnostic_settings \
  --subscription-id 12345678-1234-1234-1234-123456789012 \
  --resource-group my-rg \
  --cluster-name my-aks-cluster
```

#### 8.2 List Available Log Categories
```bash
# List enabled control plane log categories
aks_control_plane_log_categories \
  --subscription-id 12345678-1234-1234-1234-123456789012 \
  --resource-group my-rg \
  --cluster-name my-aks-cluster
```

#### 8.3 Query Control Plane Logs
```bash
# Query API server logs for the last hour
aks_control_plane_logs \
  --subscription-id 12345678-1234-1234-1234-123456789012 \
  --resource-group my-rg \
  --cluster-name my-aks-cluster \
  --log-category kube-apiserver \
  --start-time 2025-07-15T10:00:00Z \
  --end-time 2025-07-15T11:00:00Z \
  --max-records 500
```

### 9. Security Considerations

#### 9.1 Input Sanitization
- Validate all resource IDs and parameters
- Use parameterized KQL queries to prevent injection
- Sanitize log level and category parameters

#### 9.2 Access Control
- Require appropriate Azure RBAC permissions
- Use existing authentication patterns
- Log access attempts for audit purposes

#### 9.3 Rate Limiting
- Implement reasonable defaults for query frequency
- Respect Azure API rate limits
- Provide guidance on query optimization

### 10. Testing Strategy

#### 10.1 Unit Tests
- Test parameter validation
- Test KQL query building
- Test error handling scenarios
- Test time range validation

#### 10.2 Integration Tests
- Test against clusters with/without diagnostic settings
- Test with various log categories
- Test time range edge cases

### 11. Access Level Requirements

- **Readonly**: All tools available (diagnostic settings check, log categories list, log querying)
- **Readwrite**: Same as readonly (log access is inherently read-only)
- **Admin**: Same as readonly (no additional admin-specific log operations)

### 12. Documentation Updates

#### 12.1 README Updates
- Add control plane tools to available tools list
- Include usage examples and common scenarios
- Document required Azure permissions

#### 12.2 User Guide
- Step-by-step workflow documentation
- Troubleshooting common issues
- Best practices for log querying

## Implementation Timeline

1. **Phase 1**: Tool registration and basic handlers (registry.go updates, handlers.go additions)
2. **Phase 2**: Safety validations and helper functions (parameter validation, KQL query building)
3. **Phase 3**: Server integration and testing (server.go updates, unit tests)
4. **Phase 4**: Documentation and usage examples
5. **Phase 5**: Integration testing and refinements

## Dependencies

- Existing Azure CLI integration (`internal/azcli`)
- Monitor component infrastructure (`internal/components/monitor`)
- Tool registration patterns (`internal/tools`, `internal/server`)
- Configuration management (`internal/config`)
- Azure CLI with Log Analytics extension
- Appropriate Azure RBAC permissions for diagnostic settings and Log Analytics access
