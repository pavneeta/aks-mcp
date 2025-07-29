# Azure Fleet Tools for AKS-MCP

Implement Azure Kubernetes Service Fleet management capabilities for AKS-MCP.

## Overview

This component provides Azure Fleet command-line tools for managing AKS Fleet resources, including fleet creation, member management, update runs, and update strategies through the Azure CLI.

## Supported Commands

### Generic Fleet Tool

#### `az_fleet`
**Purpose**: Execute Azure Fleet commands for AKS Fleet management with structured parameters

**Parameters**:
- `operation` (required): The operation to perform (list, show, create, update, delete, start, stop, get-credentials)
- `resource` (required): The resource type to operate on (fleet, member, updaterun, updatestrategy, placement)
- `args` (required): Additional arguments for the command

**Example Usage**:
```json
{
  "operation": "list",
  "resource": "fleet",
  "args": "--resource-group myResourceGroup"
}

{
  "operation": "show",
  "resource": "fleet", 
  "args": "--name myFleet --resource-group myResourceGroup"
}

{
  "operation": "create",
  "resource": "fleet",
  "args": "--name myFleet --resource-group myResourceGroup --location eastus"
}

{
  "operation": "create",
  "resource": "clusterresourceplacement",
  "args": "--name nginx --policy PickAll --selector app=nginx"
}

{
  "operation": "get-credentials",
  "resource": "fleet",
  "args": "--name myFleet --resource-group myResourceGroup"
}
```

## Fleet Command Categories

### Read-Only Commands (Available at all access levels)

#### Fleet Information
- `az fleet list` - List all fleets
- `az fleet show` - Show details of a specific fleet

#### Fleet Member Information  
- `az fleet member list` - List all members of a fleet
- `az fleet member show` - Show details of a specific fleet member

#### Update Run Information
- `az fleet updaterun list` - List all update runs for a fleet
- `az fleet updaterun show` - Show details of a specific update run

#### Update Strategy Information
- `az fleet updatestrategy list` - List all update strategies for a fleet
- `az fleet updatestrategy show` - Show details of a specific update strategy

#### Fleet Credentials
- `az fleet get-credentials` - Get kubeconfig credentials for a fleet

#### Kubernetes ClusterResourcePlacement Operations (Cluster-scoped CRDs)
- `clusterresourceplacement list` - List all ClusterResourcePlacements
- `clusterresourceplacement show` - Show details of a specific ClusterResourcePlacement
- `clusterresourceplacement get` - Get a specific ClusterResourcePlacement (alias for show)

### Read-Write Commands (Available at readwrite and admin access levels)

#### Fleet Management
- `az fleet create` - Create a new fleet
- `az fleet update` - Update a fleet
- `az fleet delete` - Delete a fleet

#### Fleet Member Management
- `az fleet member create` - Add a member to a fleet
- `az fleet member update` - Update a fleet member
- `az fleet member delete` - Remove a member from a fleet

#### Update Run Management
- `az fleet updaterun create` - Create a new update run
- `az fleet updaterun start` - Start an update run
- `az fleet updaterun stop` - Stop an update run
- `az fleet updaterun delete` - Delete an update run

#### Update Strategy Management
- `az fleet updatestrategy create` - Create a new update strategy
- `az fleet updatestrategy delete` - Delete an update strategy

#### Kubernetes ClusterResourcePlacement Management (Cluster-scoped CRDs)
- `clusterresourceplacement create` - Create a new ClusterResourcePlacement
- `clusterresourceplacement delete` - Delete a ClusterResourcePlacement

### Admin Commands (Available only at admin access level)
- Currently no admin-only fleet commands defined
- Admin users get all readwrite commands by default

## Implementation Details

### File Organization
```
internal/components/fleet/
├── registry.go          # Fleet tool registration and command definitions
├── registry_test.go     # Unit tests for the registry
└── kubernetes/
    ├── client.go        # Kubernetes client wrapper using k8s adapter
    └── placement.go     # ClusterResourcePlacement operations

internal/azcli/
├── fleet_executor.go    # Unified fleet command executor (Azure CLI + Kubernetes)
└── fleet_executor_test.go # Unit tests for the fleet executor

internal/k8s/
└── adapter.go           # Config adapter for mcp-kubernetes integration
```

### Tool Registration
A single generic fleet tool is registered in the MCP server:
- **Generic Tool**: `az_fleet` - Accepts structured parameters: operation, resource, and args
- **Access Control**: Commands are validated against the configured access level through security validation
- **Execution**: Uses the unified `azcli.NewFleetExecutor()` for both Azure CLI and Kubernetes operations
- **Kubernetes Integration**: Placement operations are routed to Kubernetes API through mcp-kubernetes adapter
- **Config Conversion**: Uses `internal/k8s/adapter.go` to convert aks-mcp config to mcp-kubernetes format

### Fleet Command Structure
Fleet commands are organized using the `FleetCommand` structure for documentation:
```go
type FleetCommand struct {
    Name        string // Full Azure CLI command name
    Description string // Human-readable description
    ArgsExample string // Example arguments
}
```

### Integration with Server
The fleet tool is registered in `internal/server/server.go`:
```go
// Register generic az fleet tool with structured parameters (available at all access levels)
log.Println("Registering az fleet tool: az_fleet")
fleetTool := fleet.RegisterFleet()
s.mcpServer.AddTool(fleetTool, tools.CreateToolHandler(azcli.NewFleetExecutor(), s.cfg))
```

## Access Level Requirements

### Readonly Access
- ✅ `az fleet list` - List fleets
- ✅ `az fleet show` - Show fleet details
- ✅ `az fleet get-credentials` - Get fleet kubeconfig
- ✅ `az fleet member list` - List fleet members
- ✅ `az fleet member show` - Show fleet member details
- ✅ `az fleet updaterun list` - List update runs
- ✅ `az fleet updaterun show` - Show update run details
- ✅ `az fleet updatestrategy list` - List update strategies
- ✅ `az fleet updatestrategy show` - Show update strategy details
- ✅ `clusterresourceplacement list` - List ClusterResourcePlacements
- ✅ `clusterresourceplacement show/get` - Show ClusterResourcePlacement details

### Readwrite Access
- Inherits all readonly commands
- ✅ `az fleet create/update/delete` - Fleet management
- ✅ `az fleet member create/update/delete` - Member management
- ✅ `az fleet updaterun create/start/stop/delete` - Update run management
- ✅ `az fleet updatestrategy create/delete` - Update strategy management
- ✅ `clusterresourceplacement create/delete` - ClusterResourcePlacement management

### Admin Access
- Inherits all readwrite commands
- Currently no additional admin-specific fleet commands

## Common Use Cases

### Fleet Management
Create and manage AKS fleets:
```bash
# Create a new fleet
az fleet create --name myFleet --resource-group myResourceGroup --location eastus

# List all fleets
az fleet list --resource-group myResourceGroup

# Show fleet details
az fleet show --name myFleet --resource-group myResourceGroup

# Get fleet kubeconfig credentials
az fleet get-credentials --name myFleet --resource-group myResourceGroup
```

### Fleet Member Management
Add and manage AKS clusters in a fleet:
```bash
# Add a cluster to the fleet
az fleet member create --name myMember --fleet-name myFleet --resource-group myResourceGroup --member-cluster-id /subscriptions/xxx/resourceGroups/xxx/providers/Microsoft.ContainerService/managedClusters/myCluster

# List fleet members
az fleet member list --fleet-name myFleet --resource-group myResourceGroup

# Remove a cluster from the fleet
az fleet member delete --name myMember --fleet-name myFleet --resource-group myResourceGroup --yes
```

### Update Run Management
Coordinate updates across fleet members:
```bash
# Create an update run
az fleet updaterun create --name myUpdateRun --fleet-name myFleet --resource-group myResourceGroup --upgrade-type Full --kubernetes-version 1.28.0

# Start the update run
az fleet updaterun start --name myUpdateRun --fleet-name myFleet --resource-group myResourceGroup

# Monitor update run progress
az fleet updaterun show --name myUpdateRun --fleet-name myFleet --resource-group myResourceGroup
```

### Update Strategy Management
Define update strategies for fleet operations:
```bash
# Create an update strategy
az fleet updatestrategy create --name myStrategy --fleet-name myFleet --resource-group myResourceGroup --stages stage1

# List update strategies
az fleet updatestrategy list --fleet-name myFleet --resource-group myResourceGroup
```

### Kubernetes ClusterResourcePlacement Management
Manage ClusterResourcePlacements for fleet workload distribution:
```bash
# Create a ClusterResourcePlacement for nginx workloads
az_fleet with parameters: {
  "operation": "create",
  "resource": "clusterresourceplacement",
  "args": "--name nginx --policy PickAll --selector app=nginx"
}

# List all ClusterResourcePlacements
az_fleet with parameters: {
  "operation": "list",
  "resource": "clusterresourceplacement",
  "args": ""
}

# Show ClusterResourcePlacement details
az_fleet with parameters: {
  "operation": "show",
  "resource": "clusterresourceplacement",
  "args": "--name nginx"
}

# Delete a ClusterResourcePlacement
az_fleet with parameters: {
  "operation": "delete",
  "resource": "clusterresourceplacement",
  "args": "--name nginx"
}
```

## Error Handling

The fleet tools leverage enhanced error handling infrastructure:
- Azure CLI authentication errors are handled gracefully
- Invalid fleet or member names return descriptive error messages
- Network connectivity issues are properly reported
- Malformed arguments are validated before execution
- Access level violations are caught by security validation
- **Parameter Validation**: Operation/resource combinations are validated (e.g., 'start' is only valid for 'updaterun')
- **Access Level Enforcement**: Read-only operations (list, show) are allowed at all access levels, while write operations require readwrite or admin access

## Security and Access Control

Fleet commands are subject to the same security validation as other Azure CLI commands:
- **Command Validation**: All fleet commands must pass security validation before execution
- **Access Level Control**: Commands are filtered based on readonly/readwrite/admin access levels
- **Binary Validation**: Only `az fleet` commands are allowed for execution through the fleet tool
- **Timeout Protection**: Commands have configurable execution timeouts

## Integration Examples

### Using with Claude/AI Assistants
```
"Please show me all fleets in my production resource group"

This would translate to:
az_fleet with parameters: {
  "operation": "list",
  "resource": "fleet", 
  "args": "--resource-group production"
}

"Add the cluster 'web-cluster' to my fleet 'prod-fleet'"

This would translate to:
az_fleet with parameters: {
  "operation": "create",
  "resource": "member",
  "args": "--name web-cluster --fleet-name prod-fleet --resource-group production --member-cluster-id /subscriptions/.../managedClusters/web-cluster"
}

"Create a ClusterResourcePlacement to deploy nginx to all clusters with the app=frontend label"

This would translate to:
az_fleet with parameters: {
  "operation": "create",
  "resource": "clusterresourceplacement",
  "args": "--name nginx-placement --policy PickAll --selector app=frontend"
}
```

## Requirements

### Prerequisites
- Azure CLI installed and accessible in PATH
- kubectl installed and accessible in PATH (for placement operations)
- Valid Azure authentication (via `az login` or service principal)
- Valid Kubernetes authentication (kubeconfig configured for fleet cluster)
- Appropriate Azure permissions for fleet operations
- AKS Fleet preview features enabled in your subscription

### Dependencies
- **Azure CLI**: The `az` command-line tool with fleet extension
- **kubectl**: For Kubernetes placement operations
- **Fleet Extension**: `az extension add --name fleet` (if not already installed)
- **mcp-kubernetes**: Integration library for Kubernetes operations
- **k8s Adapter**: Config conversion between aks-mcp and mcp-kubernetes
- **Security Validation**: All commands are validated against the configured access level
- **Shell Execution**: Commands are executed through secure shell process handling

## Testing

Comprehensive unit tests cover:
- Tool registration functionality
- Command validation
- Fleet command structure validation
- Integration with MCP framework

Run tests with:
```bash
go test -v ./internal/components/fleet/...
go test -v ./internal/azcli/...
```

## Configuration

### Access Level Configuration
Set the access level when starting the server:
```bash
./aks-mcp --access-level readonly    # Only read operations
./aks-mcp --access-level readwrite   # Read and write operations  
./aks-mcp --access-level admin       # All operations
```

### Timeout Configuration
Configure command execution timeout:
```bash
./aks-mcp --timeout 600  # 10 minutes timeout (default)
```

## Future Enhancements

### Planned Features
- Fleet-specific resource management
- Advanced update orchestration
- Cross-region fleet management
- Integration with GitOps workflows
- Fleet monitoring and alerting
- **Enhanced Parameter Validation**: More granular validation of command arguments
- **Smart Defaults**: Automatic parameter inference based on context
- **Batch Operations**: Support for bulk fleet operations
- **Extended Placement Operations**: Support for more complex ClusterResourcePlacement configurations
- **Unified Config Management**: Further integration between Azure and Kubernetes configuration

## Best Practices

### Fleet Design
- Use descriptive fleet names that reflect their purpose
- Group clusters by environment, region, or application
- Plan update strategies to minimize service disruption
- Monitor update runs for completion and errors

### Member Management
- Ensure clusters meet fleet requirements before adding
- Use consistent tagging across fleet members
- Regularly review and update fleet membership
- Test updates on staging fleets before production

### Update Management
- Create update strategies for controlled rollouts
- Test Kubernetes version compatibility before fleet updates
- Monitor cluster health during update runs
- Plan rollback strategies for failed updates