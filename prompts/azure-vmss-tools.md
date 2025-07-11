# Azure VMSS Tools for AKS-MCP

Get detailed Virtual Machine Scale Set (VMSS) information for AKS node pools - provides low-level VMSS configuration details not available through standard AKS commands.

## Tools

### `get_vmss_info_by_node_pool`

**Purpose**: Get detailed VMSS configuration for a specific AKS node pool

**Parameters**:
- `subscription_id` (required): Azure subscription ID
- `resource_group` (required): Resource group containing the AKS cluster
- `cluster_name` (required): Name of the AKS cluster
- `node_pool_name` (required): Name of the specific node pool

**Returns**: Complete VMSS configuration including:
- VM instance configuration (size, OS image, extensions)
- Network configuration (subnets, load balancer settings)
- Scaling settings and upgrade policies
- Security settings and managed identity
- Storage configuration

### `get_all_vmss_by_cluster`

**Purpose**: Get VMSS information for all node pools in an AKS cluster

**Parameters**:
- `subscription_id` (required): Azure subscription ID
- `resource_group` (required): Resource group containing the AKS cluster
- `cluster_name` (required): Name of the AKS cluster

**Returns**: Array of VMSS configurations for each node pool:
- Summary with cluster name and total node pools count
- Per-node pool VMSS details or error messages
- Complete VMSS configuration for each successfully retrieved node pool

## Key Use Cases

1. **Troubleshooting Node Issues**: Get detailed VM configuration when nodes aren't behaving as expected
2. **Security Auditing**: Review VM extensions, security settings, and network configurations
3. **Performance Analysis**: Check VM sizes, storage types, and networking setup
4. **Compliance Checking**: Verify OS images, patches, and security configurations
5. **Resource Planning**: Understand current VM configurations for capacity planning

## What You Get vs Standard AKS Commands

**Standard `az aks nodepool show`** provides:
- Basic node pool settings (count, VM size, OS type)
- Kubernetes-level configuration
- High-level status information

**These VMSS tools provide**:
- Low-level Azure VM configuration
- Network interface details and IP configurations
- VM extensions and their settings
- Storage disk configuration and encryption
- Load balancer backend pool memberships
- Detailed OS and image information
- Scaling and upgrade policies

## Code Structure

### File Organization
```
internal/components/compute/
├── handlers.go           # Tool handlers for VMSS operations
├── registry.go          # Tool registration and MCP definitions
├── vmsshelpers.go       # Helper functions for VMSS operations
└── handlers_test.go     # Unit tests for handlers
```

### Tool Registration
Tools are registered in `internal/components/compute/registry.go`:
```go
func RegisterVMSSInfoByNodePoolTool() mcp.Tool {
    return mcp.NewTool(
        "get_vmss_info_by_node_pool",
        mcp.WithDescription("Get detailed VMSS configuration for a specific node pool"),
        mcp.WithString("subscription_id", mcp.Required()),
        mcp.WithString("resource_group", mcp.Required()),
        mcp.WithString("cluster_name", mcp.Required()),
        mcp.WithString("node_pool_name", mcp.Required()),
    )
}

func RegisterAllVMSSByClusterTool() mcp.Tool {
    return mcp.NewTool(
        "get_all_vmss_by_cluster",
        mcp.WithDescription("Get detailed VMSS configuration for all node pools"),
        mcp.WithString("subscription_id", mcp.Required()),
        mcp.WithString("resource_group", mcp.Required()),
        mcp.WithString("cluster_name", mcp.Required()),
    )
}
```

### Handler Implementation
Handlers in `internal/components/compute/handlers.go`:
```go
func GetVMSSInfoByNodePoolHandler(client *azureclient.AzureClient, cfg *config.ConfigData) tools.ResourceHandler {
    return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
        // Extract AKS parameters
        
        // Get node pool name
        
        // Get cluster details and VMSS information

        // Return JSON response
    })
}
```

### Helper Functions
VMSS helper functions in `internal/components/compute/vmsshelpers.go`:
```go
// GetVMSSIDFromNodePool - Find VMSS resource ID for a specific node pool
func GetVMSSIDFromNodePool(ctx context.Context, cluster *armcontainerservice.ManagedCluster, nodePoolName string, client *azureclient.AzureClient) (string, error)

// GetNodePoolsFromAKS - Extract all node pools from AKS cluster
func GetNodePoolsFromAKS(ctx context.Context, cluster *armcontainerservice.ManagedCluster, client *azureclient.AzureClient) ([]*armcontainerservice.ManagedClusterAgentPoolProfile, error)

// findVMSSForNodePool - Locate VMSS in node resource group by matching naming patterns
func findVMSSForNodePool(ctx context.Context, client *azureclient.AzureClient, subscriptionID, nodeResourceGroup, nodePoolName string) (string, error)
```
