package resourcehelpers

import (
	"context"
	"fmt"

	"github.com/Azure/aks-mcp/internal/azure"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
)

// GetRouteTableIDFromAKS attempts to find a route table associated with an AKS cluster.
// It first checks if a subnet is associated with the AKS cluster, then looks for a route table attached to that subnet.
// If no route table is found, it returns an empty string and no error (this is a valid state).
func GetRouteTableIDFromAKS(
	ctx context.Context,
	cluster *armcontainerservice.ManagedCluster,
	client *azure.AzureClient,
) (string, error) {
	// Ensure the cluster is valid
	if cluster == nil || cluster.Properties == nil {
		return "", fmt.Errorf("invalid cluster or cluster properties")
	}

	// Get subnet ID using the helper function which handles cases when VnetSubnetID is not set
	subnetID, err := GetSubnetIDFromAKS(ctx, cluster, client)
	if err != nil || subnetID == "" {
		return "", fmt.Errorf("no subnet found for AKS cluster: %v", err)
	}

	// Parse subnet ID to get subscription, resource group, vnet name and subnet name
	parsedSubnetID, err := arm.ParseResourceID(subnetID)
	if err != nil {
		return "", fmt.Errorf("failed to parse subnet ID: %v", err)
	}

	// Check if this is a subnet resource
	if parsedSubnetID.ResourceType.String() != "Microsoft.Network/virtualNetworks/subnets" {
		return "", fmt.Errorf("invalid subnet ID format: %s", subnetID)
	}

	// Get the subscription ID from the subnet ID
	subscriptionID := parsedSubnetID.SubscriptionID
	resourceGroup := parsedSubnetID.ResourceGroupName
	subnetName := parsedSubnetID.Name

	// Get VNet name from parent resource
	var vnetName string
	if parsedSubnetID.Parent != nil {
		vnetName = parsedSubnetID.Parent.Name
	} else {
		return "", fmt.Errorf("could not determine VNet name from subnet ID: %s", subnetID)
	}

	// Get subnet details to find attached route table
	clients, err := client.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return "", fmt.Errorf("failed to get clients for subscription %s: %v", subscriptionID, err)
	}

	subnet, err := clients.SubnetsClient.Get(ctx, resourceGroup, vnetName, subnetName, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get subnet details: %v", err)
	}

	// Check if the subnet has a route table attached
	if subnet.Properties == nil || subnet.Properties.RouteTable == nil || subnet.Properties.RouteTable.ID == nil {
		// No route table attached - this is a valid configuration state
		return "", nil
	}

	routeTableID := *subnet.Properties.RouteTable.ID

	return routeTableID, nil
}
