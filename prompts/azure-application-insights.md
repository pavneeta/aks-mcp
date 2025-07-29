# Azure Application Insights Tool for AKS-MCP

Implement Azure Application Insights monitoring capabilities for AKS clusters to track application performance and telemetry.

## Tool: `az_monitor_app_insights_query`

**Purpose**: Query Application Insights telemetry data for applications running in AKS clusters

**Parameters**:
- `subscription_id` (required): Azure subscription ID
- `resource_group` (required): Resource group name containing the Application Insights resource
- `app_insights_name` (required): Application Insights resource name
- `query` (required): KQL query to execute against Application Insights data
- `start_time` (optional): Start time for query in ISO 8601 format
- `end_time` (optional): End time for query (defaults to current time)
- `timespan` (optional): Query timespan (e.g., "PT1H", "P1D")

**Operations**:
- **query**: Execute KQL query against Application Insights telemetry data

## Implementation Steps

1. **Use existing executor** from `internal/azcli/executor.go` for Azure CLI commands
2. **Build Application Insights resource ID** from subscription, resource group, and app insights name
3. **Validate KQL queries** for safety and prevent injection
4. **Return raw JSON output** from Azure CLI commands

## Key Azure CLI Command

```bash
# Query Application Insights telemetry data
az monitor app-insights query \
  --app /subscriptions/{{ SUBSCRIPTION_ID }}/resourceGroups/{{ RESOURCE_GROUP_NAME }}/providers/Microsoft.Insights/components/{{ APP_INSIGHTS_NAME }} \
  --analytics-query "{{ KQL_QUERY }}" \
  --start-time {{ START_TIME }} \
  --output json
```

## Application Insights Telemetry Types
- **Requests**: HTTP requests to your applications
- **Dependencies**: Outbound calls to databases, APIs, services
- **Exceptions**: Application exceptions and errors
- **Traces**: Application logs and trace data
- **CustomEvents**: Custom events logged by applications

## Code Structure Requirements

- Extend existing `internal/components/monitor/handlers.go` with Application Insights handler
- Add tool registration to `internal/components/monitor/registry.go`
- Use existing `azcli.NewExecutor()` pattern for Azure CLI command execution
- Follow existing error handling patterns from advisor and network components

## Access Level Requirements
- **Readonly**: All operations (query)
- **Readwrite**: Same as readonly (querying is read-only)
- **Admin**: Same as readonly (no admin-specific query operations)

## Validation Requirements

- Validate required parameters: `subscription_id`, `resource_group`, `app_insights_name`, `query`
- Validate KQL query safety (prevent dangerous keywords: delete, drop, create, alter, insert, update)
- Ensure queries start with valid Application Insights table names
- Validate time format (RFC3339/ISO 8601) for start_time and end_time parameters

## Expected Integration

- Extend existing `internal/components/monitor/registry.go` with Application Insights commands
- Add handler functions to `internal/components/monitor/handlers.go`
- Follow existing error handling patterns from advisor and network components
- Use standard JSON output format
- Integrate with existing security validation

## Common KQL Query Examples

```kql
# Recent requests
requests | where timestamp > ago(1h) | limit 10

# Request performance over time
requests | where timestamp > ago(1h) | summarize avg(duration), count() by bin(timestamp, 5m)

# Error rate by operation
requests | where timestamp > ago(1h) | summarize total=count(), errors=countif(success == false) by operation_Name

# External dependency calls
dependencies | where timestamp > ago(1h) | summarize count(), avg(duration) by type, target

# Recent exceptions
exceptions | where timestamp > ago(1h) | project timestamp, type, method, outerMessage
```

## User Experience

**Example Usage**:
```bash
# Get recent requests
az_monitor_app_insights_query \
  --subscription-id 82d6efa7-b1b6-4aa0-ab12-d10788552670 \
  --resource-group thomas \
  --app-insights-name thomastest39-insights \
  --query "requests | where timestamp > ago(1h) | limit 10"

# Get dependency analysis with time range
az_monitor_app_insights_query \
  --subscription-id 82d6efa7-b1b6-4aa0-ab12-d10788552670 \
  --resource-group thomas \
  --app-insights-name thomastest39-insights \
  --query "dependencies | summarize count() by type, target" \
  --start-time 2025-07-17T00:00:00Z \
  --end-time 2025-07-18T00:00:00Z
```

## Tool Registration

- Register as MCP tool named `az_monitor_app_insights_query`
- Include all required and optional parameters with proper descriptions
- Follow existing tool registration patterns in `internal/components/monitor/registry.go`

## Success Criteria
- ✅ Execute KQL queries against Application Insights telemetry data
- ✅ Filter by time range using multiple time specification methods
- ✅ Return raw Azure CLI JSON output with telemetry data
- ✅ Validate KQL queries for safety and correctness
- ✅ Provide meaningful error messages for invalid parameters
- ✅ Integrate with existing MCP tool framework

## Implementation Priority
1. Basic KQL query execution with time filtering
2. Query validation and safety checks
3. Integration with existing monitoring tools
4. Performance optimization for common queries

## Error Handling
- Validate Application Insights resource ID format
- Handle Azure CLI authentication errors
- Validate KQL query syntax and safety
- Validate time range parameters
- Handle empty result sets gracefully
- Provide clear error messages for malformed queries

Generate the implementation following these high-level specifications and integrate with the existing `internal/components/monitor/` package structure.
