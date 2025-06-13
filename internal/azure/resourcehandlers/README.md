# Azure Resource Handlers

This package provides handlers for Azure resource tools that retrieve information about Azure resources related to AKS clusters.

## Handlers

The following handlers are implemented:

### GetVNetInfoHandler

Retrieves information about the Virtual Network (VNet) used by an AKS cluster. This handler:

1. Gets the AKS cluster details
2. Extracts the VNet ID from the cluster using `resourcehelpers.GetVNetIDFromAKS`
3. Fetches the VNet details and converts to a `models.VNetInfo` structure
4. Returns the VNet information as JSON

### GetNSGInfoHandler

Retrieves information about the Network Security Group (NSG) used by an AKS cluster. This handler:

1. Gets the AKS cluster details
2. Extracts the NSG ID from the cluster using `resourcehelpers.GetNSGIDFromAKS`
3. Fetches the NSG details and converts to a `models.NSGInfo` structure
4. Returns the NSG information as JSON

### GetRouteTableInfoHandler

Retrieves information about the Route Table used by an AKS cluster. This handler:

1. Gets the AKS cluster details
2. Extracts the Route Table ID from the cluster using `resourcehelpers.GetRouteTableIDFromAKS`
3. Fetches the Route Table details and converts to a `models.RouteTableInfo` structure
4. Returns the Route Table information as JSON

### GetSubnetInfoHandler

Retrieves information about the Subnet used by an AKS cluster. This handler:

1. Gets the AKS cluster details
2. Extracts the Subnet ID from the cluster using `resourcehelpers.GetSubnetIDFromAKS`
3. Fetches the Subnet details and converts to a `models.SubnetInfo` structure
4. Returns the Subnet information as JSON

## Tool Registration

Each handler has a corresponding registration function that defines the tool parameters:

- `RegisterVNetInfoTool`
- `RegisterNSGInfoTool`
- `RegisterRouteTableInfoTool`
- `RegisterSubnetInfoTool`

All tools accept the following parameters:

- `subscription_id`: Azure Subscription ID
- `resource_group`: Azure Resource Group containing the AKS cluster
- `cluster_name`: Name of the AKS cluster

These tools are registered with the MCP server in the `registerAzureResourceTools` method in `server.go`.
