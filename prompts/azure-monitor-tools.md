# Azure Monitor Tools for AKS-MCP

Implement Azure Monitor capabilities for AKS clusters and related Azure resources.

## Overview

This component provides Azure Monitor command-line tools for retrieving metrics, managing diagnostic settings, and monitoring Azure resources through the Azure CLI.

## Supported Commands

### Metrics Commands

#### `az_monitor_metrics_list`
**Purpose**: List the metric values for a resource

**Parameters**:
- `args` (required): Arguments for the `az monitor metrics list` command

**Example Usage**:
```bash
--resource /subscriptions/{subscription}/resourceGroups/{resourceGroup}/providers/Microsoft.Compute/virtualMachines/{vmName} --metric "Percentage CPU"
```

#### `az_monitor_metrics_list-definitions`
**Purpose**: List the metric definitions for a resource

**Parameters**:
- `args` (required): Arguments for the `az monitor metrics list-definitions` command

**Example Usage**:
```bash
--resource /subscriptions/{subscription}/resourceGroups/{resourceGroup}/providers/Microsoft.ContainerService/managedClusters/{clusterName}
```

#### `az_monitor_metrics_list-namespaces`
**Purpose**: List the metric namespaces for a resource

**Parameters**:
- `args` (required): Arguments for the `az monitor metrics list-namespaces` command

**Example Usage**:
```bash
--resource /subscriptions/{subscription}/resourceGroups/{resourceGroup}/providers/Microsoft.ContainerService/managedClusters/{clusterName}
```

## Implementation Details

### File Organization
```
internal/components/monitor/
├── registry.go        # Command registration and tool definitions
└── registry_test.go   # Unit tests for the registry
```

### Tool Registration
Tools are automatically registered in the MCP server based on access level:
- **Read-only**: All metric listing commands
- **Read-write**: Currently empty (placeholder for future features)
- **Admin**: Currently empty (placeholder for future features)

### Command Structure
Each monitor command follows the `MonitorCommand` structure:
```go
type MonitorCommand struct {
    Name        string // Full Azure CLI command name
    Description string // Human-readable description
    ArgsExample string // Example arguments
    Category    string // Command category (e.g., "metrics")
}
```

### Integration with Server
The monitor commands are registered in `internal/server/server.go`:
```go
// Register read-only az monitor commands (available at all access levels)
for _, cmd := range monitor.GetReadOnlyMonitorCommands() {
    log.Println("Registering az monitor command:", cmd.Name)
    azTool := monitor.RegisterMonitorCommand(cmd)
    commandExecutor := azcli.CreateCommandExecutorFunc(cmd.Name)
    s.mcpServer.AddTool(azTool, tools.CreateToolHandler(commandExecutor, s.cfg))
}
```

## Access Level Requirements

### Readonly Access
- ✅ `az monitor metrics list` - List metric values
- ✅ `az monitor metrics list-definitions` - List metric definitions  
- ✅ `az monitor metrics list-namespaces` - List metric namespaces

### Readwrite Access
- Inherits all readonly commands
- 🔄 *Future: Additional monitoring configuration commands*

### Admin Access
- Inherits all readwrite commands
- 🔄 *Future: Advanced monitoring management commands*

## Common Use Cases

### AKS Cluster Monitoring
Monitor AKS cluster performance and health:
```bash
# Get cluster metrics
az monitor metrics list --resource /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ContainerService/managedClusters/{cluster}

# List available metrics for AKS cluster
az monitor metrics list-definitions --resource /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ContainerService/managedClusters/{cluster}

az monitor metrics list --resource /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ContainerService/managedClusters/{cluster} --metric apiserver_cpu_usage_percentage   --interval PT1M   --aggregation Average   --output table
```



## Error Handling

The monitor tools leverage the existing error handling infrastructure:
- Azure CLI authentication errors are handled gracefully
- Invalid resource IDs return descriptive error messages
- Network connectivity issues are properly reported
- Malformed arguments are validated before execution

## Future Enhancements

### Planned Read-Write Features
- Diagnostic settings management
- Alert rule configuration
- Action group management
- Log Analytics workspace operations

### Planned Admin Features
- Advanced monitoring configuration
- Custom metric definitions
- Cross-subscription monitoring setup

## Testing

Comprehensive unit tests cover:
- Command registration functionality
- Tool creation and validation
- Command structure validation
- Integration with MCP framework

Run tests with:
```bash
go test -v ./internal/components/monitor/...
```

## Integration Examples

### Using with Claude/AI Assistants
```
"Please show me the CPU metrics for my AKS cluster named 'prod-cluster' in resource group 'production'"

This would translate to:
az_monitor_metrics_list with args: "--resource /subscriptions/{sub}/resourceGroups/production/providers/Microsoft.ContainerService/managedClusters/prod-cluster --metric 'CPU Usage'"
```

## Dependencies

- Azure CLI (`az` command) must be installed and configured
- Valid Azure authentication (service principal or user login)
- Appropriate RBAC permissions for target resources
- Network connectivity to Azure endpoints
