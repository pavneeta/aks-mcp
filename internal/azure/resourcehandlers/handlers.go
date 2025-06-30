// Package resourcehandlers provides handler functions for Azure resource tools.
package resourcehandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/aks-mcp/internal/azure"
	"github.com/Azure/aks-mcp/internal/azure/resourcehelpers"
	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/tools"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// GetVNetInfoHandler returns a handler for the get_vnet_info command
func GetVNetInfoHandler(client *azure.AzureClient, cache *azure.AzureCache, cfg *config.ConfigData) tools.CommandExecutor {
	return tools.CommandExecutorFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		// Extract parameters
		subID, ok := params["subscription_id"].(string)
		if !ok || subID == "" {
			return "", fmt.Errorf("missing or invalid subscription_id parameter")
		}

		rg, ok := params["resource_group"].(string)
		if !ok || rg == "" {
			return "", fmt.Errorf("missing or invalid resource_group parameter")
		}

		clusterName, ok := params["cluster_name"].(string)
		if !ok || clusterName == "" {
			return "", fmt.Errorf("missing or invalid cluster_name parameter")
		}

		// Get the cluster details
		ctx := context.Background()
		cluster, err := GetClusterDetails(ctx, client, subID, rg, clusterName)
		if err != nil {
			return "", fmt.Errorf("failed to get cluster details: %v", err)
		}

		// Get the VNet ID from the cluster
		vnetID, err := resourcehelpers.GetVNetIDFromAKS(ctx, cluster, client, cache)
		if err != nil {
			return "", fmt.Errorf("failed to get VNet ID: %v", err)
		}

		// Get the VNet details using the resource ID
		vnetInterface, err := client.GetResourceByID(ctx, vnetID)
		if err != nil {
			return "", fmt.Errorf("failed to get VNet details: %v", err)
		}

		vnet, ok := vnetInterface.(*armnetwork.VirtualNetwork)
		if !ok {
			return "", fmt.Errorf("unexpected resource type returned for VNet")
		}

		// Return the VNet details directly as JSON
		resultJSON, err := json.MarshalIndent(vnet, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal VNet info to JSON: %v", err)
		}

		return string(resultJSON), nil
	})
}

// GetNSGInfoHandler returns a handler for the get_nsg_info command
func GetNSGInfoHandler(client *azure.AzureClient, cache *azure.AzureCache, cfg *config.ConfigData) tools.CommandExecutor {
	return tools.CommandExecutorFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		// Extract parameters
		subID, ok := params["subscription_id"].(string)
		if !ok || subID == "" {
			return "", fmt.Errorf("missing or invalid subscription_id parameter")
		}

		rg, ok := params["resource_group"].(string)
		if !ok || rg == "" {
			return "", fmt.Errorf("missing or invalid resource_group parameter")
		}

		clusterName, ok := params["cluster_name"].(string)
		if !ok || clusterName == "" {
			return "", fmt.Errorf("missing or invalid cluster_name parameter")
		}

		// Get the cluster details
		ctx := context.Background()
		cluster, err := GetClusterDetails(ctx, client, subID, rg, clusterName)
		if err != nil {
			return "", fmt.Errorf("failed to get cluster details: %v", err)
		}

		// Get the NSG ID from the cluster
		nsgID, err := resourcehelpers.GetNSGIDFromAKS(ctx, cluster, client, cache)
		if err != nil {
			return "", fmt.Errorf("failed to get NSG ID: %v", err)
		}

		// Get the NSG details using the resource ID
		nsgInterface, err := client.GetResourceByID(ctx, nsgID)
		if err != nil {
			return "", fmt.Errorf("failed to get NSG details: %v", err)
		}

		nsg, ok := nsgInterface.(*armnetwork.SecurityGroup)
		if !ok {
			return "", fmt.Errorf("unexpected resource type returned for NSG")
		}

		// Return the NSG details directly as JSON
		resultJSON, err := json.MarshalIndent(nsg, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal NSG info to JSON: %v", err)
		}

		return string(resultJSON), nil
	})
}

// GetRouteTableInfoHandler returns a handler for the get_route_table_info command
func GetRouteTableInfoHandler(client *azure.AzureClient, cache *azure.AzureCache, cfg *config.ConfigData) tools.CommandExecutor {
	return tools.CommandExecutorFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		// Extract parameters
		subID, ok := params["subscription_id"].(string)
		if !ok || subID == "" {
			return "", fmt.Errorf("missing or invalid subscription_id parameter")
		}

		rg, ok := params["resource_group"].(string)
		if !ok || rg == "" {
			return "", fmt.Errorf("missing or invalid resource_group parameter")
		}

		clusterName, ok := params["cluster_name"].(string)
		if !ok || clusterName == "" {
			return "", fmt.Errorf("missing or invalid cluster_name parameter")
		}

		// Get the cluster details
		ctx := context.Background()
		cluster, err := GetClusterDetails(ctx, client, subID, rg, clusterName)
		if err != nil {
			return "", fmt.Errorf("failed to get cluster details: %v", err)
		}

		// Get the RouteTable ID from the cluster
		rtID, err := resourcehelpers.GetRouteTableIDFromAKS(ctx, cluster, client, cache)
		if err != nil {
			return "", fmt.Errorf("failed to get RouteTable ID: %v", err)
		}

		// Get the RouteTable details using the resource ID
		rtInterface, err := client.GetResourceByID(ctx, rtID)
		if err != nil {
			return "", fmt.Errorf("failed to get RouteTable details: %v", err)
		}

		rt, ok := rtInterface.(*armnetwork.RouteTable)
		if !ok {
			return "", fmt.Errorf("unexpected resource type returned for RouteTable")
		}

		// Return the RouteTable details directly as JSON
		resultJSON, err := json.MarshalIndent(rt, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal RouteTable info to JSON: %v", err)
		}

		return string(resultJSON), nil
	})
}

// GetSubnetInfoHandler returns a handler for the get_subnet_info command
func GetSubnetInfoHandler(client *azure.AzureClient, cache *azure.AzureCache, cfg *config.ConfigData) tools.CommandExecutor {
	return tools.CommandExecutorFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		// Extract parameters
		subID, ok := params["subscription_id"].(string)
		if !ok || subID == "" {
			return "", fmt.Errorf("missing or invalid subscription_id parameter")
		}

		rg, ok := params["resource_group"].(string)
		if !ok || rg == "" {
			return "", fmt.Errorf("missing or invalid resource_group parameter")
		}

		clusterName, ok := params["cluster_name"].(string)
		if !ok || clusterName == "" {
			return "", fmt.Errorf("missing or invalid cluster_name parameter")
		}

		// Get the cluster details
		ctx := context.Background()
		cluster, err := GetClusterDetails(ctx, client, subID, rg, clusterName)
		if err != nil {
			return "", fmt.Errorf("failed to get cluster details: %v", err)
		}

		// Get the Subnet ID from the cluster
		subnetID, err := resourcehelpers.GetSubnetIDFromAKS(ctx, cluster, client, cache)
		if err != nil {
			return "", fmt.Errorf("failed to get Subnet ID: %v", err)
		}

		// Get the Subnet details using the resource ID
		subnetInterface, err := client.GetResourceByID(ctx, subnetID)
		if err != nil {
			return "", fmt.Errorf("failed to get Subnet details: %v", err)
		}

		subnet, ok := subnetInterface.(*armnetwork.Subnet)
		if !ok {
			return "", fmt.Errorf("unexpected resource type returned for Subnet")
		}

		// Return the Subnet details directly as JSON
		resultJSON, err := json.MarshalIndent(subnet, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal Subnet info to JSON: %v", err)
		}

		return string(resultJSON), nil
	})
}

// GetClusterDetails gets the details of an AKS cluster
func GetClusterDetails(ctx context.Context, client *azure.AzureClient, subscriptionID, resourceGroup, clusterName string) (*armcontainerservice.ManagedCluster, error) {
	// Get the cluster directly using the Azure client
	return client.GetAKSCluster(ctx, subscriptionID, resourceGroup, clusterName)
}
