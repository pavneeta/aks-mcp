package resourcehelpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/aks-mcp/internal/azureclient"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
)

// GetPrivateEndpointIDFromAKS attempts to find the private endpoint associated with an AKS cluster.
// It looks for a private endpoint named "kube-apiserver" in the cluster's node resource group.
// Returns empty string if no private endpoint is found (indicating a non-private cluster).
func GetPrivateEndpointIDFromAKS(
	ctx context.Context,
	cluster *armcontainerservice.ManagedCluster,
	client *azureclient.AzureClient,
) (string, error) {
	// Ensure the cluster is valid
	if cluster == nil || cluster.Properties == nil {
		return "", fmt.Errorf("invalid cluster or cluster properties")
	}

	// Check if cluster has APIServerAccessProfile indicating private cluster
	if cluster.Properties.APIServerAccessProfile == nil ||
		cluster.Properties.APIServerAccessProfile.EnablePrivateCluster == nil ||
		!*cluster.Properties.APIServerAccessProfile.EnablePrivateCluster {
		// Not a private cluster, return empty string (not an error)
		return "", nil
	}

	// Get the node resource group
	if cluster.Properties.NodeResourceGroup == nil {
		return "", fmt.Errorf("node resource group not found for AKS cluster")
	}
	nodeResourceGroup := *cluster.Properties.NodeResourceGroup

	// Get subscription ID from cluster ID
	subscriptionID := extractSubscriptionIDFromCluster(cluster)
	if subscriptionID == "" {
		return "", fmt.Errorf("unable to extract subscription ID from cluster")
	}

	// Look for the "kube-apiserver" private endpoint in the node resource group
	return findPrivateEndpointInNodeResourceGroup(ctx, client, subscriptionID, nodeResourceGroup)
}

// findPrivateEndpointInNodeResourceGroup looks for the "kube-apiserver" private endpoint
// in the specified node resource group
func findPrivateEndpointInNodeResourceGroup(
	ctx context.Context,
	client *azureclient.AzureClient,
	subscriptionID string,
	nodeResourceGroup string,
) (string, error) {
	// Get clients for the subscription
	clients, err := client.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return "", fmt.Errorf("failed to get clients for subscription %s: %v", subscriptionID, err)
	}

	// List private endpoints in the node resource group
	pager := clients.PrivateEndpointsClient.NewListPager(nodeResourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list private endpoints in resource group %s: %v", nodeResourceGroup, err)
		}

		for _, pe := range page.Value {
			if pe.Name != nil && pe.ID != nil {
				peName := *pe.Name
				peID := *pe.ID

				// Look for the "kube-apiserver" private endpoint
				if peName == "kube-apiserver" {
					return peID, nil
				}
			}
		}
	}

	// No private endpoint found
	return "", nil
}

// getSubscriptionFromCluster extracts the subscription ID from the cluster's ID
func extractSubscriptionIDFromCluster(cluster *armcontainerservice.ManagedCluster) string {
	if cluster.ID == nil {
		return ""
	}

	parts := strings.Split(*cluster.ID, "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "subscriptions" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
