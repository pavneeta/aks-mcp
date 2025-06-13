// Package resourcehandlers provides handler functions for Azure resource tools.
package resourcehandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/aks-mcp/internal/azure"
	"github.com/Azure/aks-mcp/internal/azure/resourcehelpers"
	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/models"
	"github.com/Azure/aks-mcp/internal/tools"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
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

		// Parse the VNet ID
		vnetResourceID, err := azure.ParseResourceID(vnetID)
		if err != nil {
			return "", fmt.Errorf("failed to parse VNet ID: %v", err)
		}

		// Get the VNet details
		clients, err := client.GetOrCreateClientsForSubscription(vnetResourceID.SubscriptionID)
		if err != nil {
			return "", fmt.Errorf("failed to get clients for subscription %s: %v", vnetResourceID.SubscriptionID, err)
		}

		vnet, err := clients.VNetClient.Get(ctx, vnetResourceID.ResourceGroup, vnetResourceID.ResourceName, nil)
		if err != nil {
			return "", fmt.Errorf("failed to get VNet details: %v", err)
		}

		// Convert to our model
		vnetInfo := models.VNetInfo{
			Name:              *vnet.Name,
			ResourceGroup:     vnetResourceID.ResourceGroup,
			Location:          *vnet.Location,
			ID:                *vnet.ID,
			ResourceGUID:      *vnet.Properties.ResourceGUID,
			ProvisioningState: string(*vnet.Properties.ProvisioningState),
		}

		if vnet.Properties.AddressSpace != nil && vnet.Properties.AddressSpace.AddressPrefixes != nil {
			vnetInfo.AddressSpace = make([]string, len(vnet.Properties.AddressSpace.AddressPrefixes))
			for i, prefix := range vnet.Properties.AddressSpace.AddressPrefixes {
				vnetInfo.AddressSpace[i] = *prefix
			}
		}

		// Include subnet information
		if vnet.Properties.Subnets != nil {
			vnetInfo.Subnets = make([]models.SubnetInfo, 0, len(vnet.Properties.Subnets))
			for _, subnet := range vnet.Properties.Subnets {
				subnetInfo := models.SubnetInfo{
					Name:              *subnet.Name,
					ID:                *subnet.ID,
					AddressPrefix:     *subnet.Properties.AddressPrefix,
					ProvisioningState: string(*subnet.Properties.ProvisioningState),
				}
				
				// Add optional properties
				if subnet.Properties.NetworkSecurityGroup != nil && subnet.Properties.NetworkSecurityGroup.ID != nil {
					subnetInfo.NetworkSecurityGroup = *subnet.Properties.NetworkSecurityGroup.ID
				}
				
				if subnet.Properties.RouteTable != nil && subnet.Properties.RouteTable.ID != nil {
					subnetInfo.RouteTable = *subnet.Properties.RouteTable.ID
				}
				
				vnetInfo.Subnets = append(vnetInfo.Subnets, subnetInfo)
			}
		}

		// Parse the tags
		if vnet.Tags != nil {
			vnetInfo.Tags = make(map[string]string)
			for k, v := range vnet.Tags {
				if v != nil {
					vnetInfo.Tags[k] = *v
				} else {
					vnetInfo.Tags[k] = ""
				}
			}
		}

		// Return the result as JSON
		resultJSON, err := json.MarshalIndent(vnetInfo, "", "  ")
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

		// Parse the NSG ID
		nsgResourceID, err := azure.ParseResourceID(nsgID)
		if err != nil {
			return "", fmt.Errorf("failed to parse NSG ID: %v", err)
		}

		// Get the NSG details
		clients, err := client.GetOrCreateClientsForSubscription(nsgResourceID.SubscriptionID)
		if err != nil {
			return "", fmt.Errorf("failed to get clients for subscription %s: %v", nsgResourceID.SubscriptionID, err)
		}

		nsg, err := clients.NSGClient.Get(ctx, nsgResourceID.ResourceGroup, nsgResourceID.ResourceName, nil)
		if err != nil {
			return "", fmt.Errorf("failed to get NSG details: %v", err)
		}

		// Convert to our model
		nsgInfo := models.NSGInfo{
			Name:              *nsg.Name,
			ResourceGroup:     nsgResourceID.ResourceGroup,
			Location:          *nsg.Location,
			ID:                *nsg.ID,
			ProvisioningState: string(*nsg.Properties.ProvisioningState),
		}

		// Parse security rules
		if nsg.Properties.SecurityRules != nil {
			nsgInfo.SecurityRules = make([]models.NSGRule, 0, len(nsg.Properties.SecurityRules))
			for _, rule := range nsg.Properties.SecurityRules {
				secRule := models.NSGRule{
					Name:                     *rule.Name,
					ID:                       *rule.ID,
					Priority:                 *rule.Properties.Priority,
					Direction:                string(*rule.Properties.Direction),
					Access:                   string(*rule.Properties.Access),
					Protocol:                 string(*rule.Properties.Protocol),
					SourcePortRange:          *rule.Properties.SourcePortRange,
					DestinationPortRange:     *rule.Properties.DestinationPortRange,
					SourceAddressPrefix:      *rule.Properties.SourceAddressPrefix,
					DestinationAddressPrefix: *rule.Properties.DestinationAddressPrefix,
				}
				
				nsgInfo.SecurityRules = append(nsgInfo.SecurityRules, secRule)
			}
		}

		// Parse the tags
		if nsg.Tags != nil {
			nsgInfo.Tags = make(map[string]string)
			for k, v := range nsg.Tags {
				if v != nil {
					nsgInfo.Tags[k] = *v
				} else {
					nsgInfo.Tags[k] = ""
				}
			}
		}

		// Return the result as JSON
		resultJSON, err := json.MarshalIndent(nsgInfo, "", "  ")
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

		// Parse the RouteTable ID
		rtResourceID, err := azure.ParseResourceID(rtID)
		if err != nil {
			return "", fmt.Errorf("failed to parse RouteTable ID: %v", err)
		}

		// Get the RouteTable details
		clients, err := client.GetOrCreateClientsForSubscription(rtResourceID.SubscriptionID)
		if err != nil {
			return "", fmt.Errorf("failed to get clients for subscription %s: %v", rtResourceID.SubscriptionID, err)
		}

		rt, err := clients.RouteTableClient.Get(ctx, rtResourceID.ResourceGroup, rtResourceID.ResourceName, nil)
		if err != nil {
			return "", fmt.Errorf("failed to get RouteTable details: %v", err)
		}

		// Convert to our model
		rtInfo := models.RouteTableInfo{
			Name:              *rt.Name,
			ResourceGroup:     rtResourceID.ResourceGroup,
			Location:          *rt.Location,
			ID:                *rt.ID,
			ProvisioningState: string(*rt.Properties.ProvisioningState),
		}

		// Parse routes
		if rt.Properties.Routes != nil {
			rtInfo.Routes = make([]models.RouteInfo, 0, len(rt.Properties.Routes))
			for _, route := range rt.Properties.Routes {
				r := models.RouteInfo{
					Name:                *route.Name,
					ID:                  *route.ID,
					AddressPrefix:       *route.Properties.AddressPrefix,
					NextHopType:         string(*route.Properties.NextHopType),
					ProvisioningState:   string(*route.Properties.ProvisioningState),
				}
				
				if route.Properties.NextHopIPAddress != nil {
					r.NextHopIPAddress = *route.Properties.NextHopIPAddress
				}
				
				rtInfo.Routes = append(rtInfo.Routes, r)
			}
		}

		// Parse the tags
		if rt.Tags != nil {
			rtInfo.Tags = make(map[string]string)
			for k, v := range rt.Tags {
				if v != nil {
					rtInfo.Tags[k] = *v
				} else {
					rtInfo.Tags[k] = ""
				}
			}
		}

		// Return the result as JSON
		resultJSON, err := json.MarshalIndent(rtInfo, "", "  ")
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

		// Parse the Subnet ID
		subnetResourceID, err := azure.ParseResourceID(subnetID)
		if err != nil {
			return "", fmt.Errorf("failed to parse Subnet ID: %v", err)
		}

		// For subnets, the vnet name is ResourceName and subnet name is SubResourceName
		clients, err := client.GetOrCreateClientsForSubscription(subnetResourceID.SubscriptionID)
		if err != nil {
			return "", fmt.Errorf("failed to get clients for subscription %s: %v", subnetResourceID.SubscriptionID, err)
		}

		subnet, err := clients.SubnetsClient.Get(ctx, subnetResourceID.ResourceGroup, subnetResourceID.ResourceName, subnetResourceID.SubResourceName, nil)
		if err != nil {
			return "", fmt.Errorf("failed to get Subnet details: %v", err)
		}

	// Convert to our model
	subnetInfo := models.SubnetInfo{
		Name:              *subnet.Name,
		ID:                *subnet.ID,
		AddressPrefix:     *subnet.Properties.AddressPrefix,
		ProvisioningState: string(*subnet.Properties.ProvisioningState),
	}

		// Add optional properties
		if subnet.Properties.NetworkSecurityGroup != nil && subnet.Properties.NetworkSecurityGroup.ID != nil {
			subnetInfo.NetworkSecurityGroup = *subnet.Properties.NetworkSecurityGroup.ID
		}
		
		if subnet.Properties.RouteTable != nil && subnet.Properties.RouteTable.ID != nil {
			subnetInfo.RouteTable = *subnet.Properties.RouteTable.ID
		}

		// Return the result as JSON
		resultJSON, err := json.MarshalIndent(subnetInfo, "", "  ")
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
