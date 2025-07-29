// Package azure provides Azure SDK integration for AKS MCP server.
package azureclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// SubscriptionClients contains Azure clients for a specific subscription.
type SubscriptionClients struct {
	SubscriptionID           string
	ContainerServiceClient   *armcontainerservice.ManagedClustersClient
	VNetClient               *armnetwork.VirtualNetworksClient
	SubnetsClient            *armnetwork.SubnetsClient
	RouteTableClient         *armnetwork.RouteTablesClient
	NSGClient                *armnetwork.SecurityGroupsClient
	LoadBalancerClient       *armnetwork.LoadBalancersClient
	PrivateEndpointsClient   *armnetwork.PrivateEndpointsClient
	VMSSClient               *armcompute.VirtualMachineScaleSetsClient
	VMSSVMsClient            *armcompute.VirtualMachineScaleSetVMsClient
	DiagnosticSettingsClient *armmonitor.DiagnosticSettingsClient
}

// AzureClient represents an Azure API client that can handle multiple subscriptions.
type AzureClient struct {
	// Map of subscription ID to clients for that subscription
	clientsMap map[string]*SubscriptionClients
	// Mutex to ensure thread safety when accessing the map
	mu sync.RWMutex
	// Shared credential for all clients
	credential *azidentity.DefaultAzureCredential
	// Cache for Azure resources
	cache *AzureCache
}

// NewAzureClient creates a new Azure client using default credentials and the provided configuration.
func NewAzureClient(cfg *config.ConfigData) (*AzureClient, error) {
	// Create a credential using DefaultAzureCredential
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %v", err)
	}

	return &AzureClient{
		clientsMap: make(map[string]*SubscriptionClients),
		credential: cred,
		cache:      NewAzureCache(cfg.CacheTimeout),
	}, nil
}

// GetOrCreateClientsForSubscription gets existing clients for a subscription or creates new ones.
func (c *AzureClient) GetOrCreateClientsForSubscription(subscriptionID string) (*SubscriptionClients, error) {
	// First try to get existing clients with a read lock
	c.mu.RLock()
	clients, exists := c.clientsMap[subscriptionID]
	c.mu.RUnlock()

	if exists {
		return clients, nil
	}

	// If no clients exist, create new ones with a write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check again in case another goroutine created the clients while we were waiting for the lock
	if clients, exists = c.clientsMap[subscriptionID]; exists {
		return clients, nil
	}

	// Create new clients for this subscription
	containerServiceClient, err := armcontainerservice.NewManagedClustersClient(subscriptionID, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create container service client for subscription %s: %v", subscriptionID, err)
	}

	vnetClient, err := armnetwork.NewVirtualNetworksClient(subscriptionID, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual network client for subscription %s: %v", subscriptionID, err)
	}

	routeTableClient, err := armnetwork.NewRouteTablesClient(subscriptionID, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create route table client for subscription %s: %v", subscriptionID, err)
	}

	nsgClient, err := armnetwork.NewSecurityGroupsClient(subscriptionID, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create network security group client for subscription %s: %v", subscriptionID, err)
	}

	subnetsClient, err := armnetwork.NewSubnetsClient(subscriptionID, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnets client for subscription %s: %v", subscriptionID, err)
	}

	loadBalancerClient, err := armnetwork.NewLoadBalancersClient(subscriptionID, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create load balancer client for subscription %s: %v", subscriptionID, err)
	}

	privateEndpointsClient, err := armnetwork.NewPrivateEndpointsClient(subscriptionID, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create private endpoints client for subscription %s: %v", subscriptionID, err)
	}

	vmssClient, err := armcompute.NewVirtualMachineScaleSetsClient(subscriptionID, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create VMSS client for subscription %s: %v", subscriptionID, err)
	}

	vmssVMsClient, err := armcompute.NewVirtualMachineScaleSetVMsClient(subscriptionID, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create VMSS VMs client for subscription %s: %v", subscriptionID, err)
	}

	diagnosticSettingsClient, err := armmonitor.NewDiagnosticSettingsClient(c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create diagnostic settings client for subscription %s: %v", subscriptionID, err)
	}

	// Create and store the clients
	clients = &SubscriptionClients{
		SubscriptionID:           subscriptionID,
		ContainerServiceClient:   containerServiceClient,
		VNetClient:               vnetClient,
		SubnetsClient:            subnetsClient,
		RouteTableClient:         routeTableClient,
		NSGClient:                nsgClient,
		LoadBalancerClient:       loadBalancerClient,
		PrivateEndpointsClient:   privateEndpointsClient,
		VMSSClient:               vmssClient,
		VMSSVMsClient:            vmssVMsClient,
		DiagnosticSettingsClient: diagnosticSettingsClient,
	}

	c.clientsMap[subscriptionID] = clients
	return clients, nil
}

// GetAKSCluster retrieves information about the specified AKS cluster.
func (c *AzureClient) GetAKSCluster(ctx context.Context, subscriptionID, resourceGroup, clusterName string) (*armcontainerservice.ManagedCluster, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("resource:cluster:%s:%s:%s", subscriptionID, resourceGroup, clusterName)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		if cluster, ok := cached.(*armcontainerservice.ManagedCluster); ok {
			return cluster, nil
		}
	}

	clients, err := c.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.ContainerServiceClient.Get(ctx, resourceGroup, clusterName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get AKS cluster: %v", err)
	}

	cluster := &resp.ManagedCluster
	// Store in cache
	c.cache.Set(cacheKey, cluster)

	return cluster, nil
}

// GetVirtualNetwork retrieves information about the specified virtual network.
func (c *AzureClient) GetVirtualNetwork(ctx context.Context, subscriptionID, resourceGroup, vnetName string) (*armnetwork.VirtualNetwork, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("resource:vnet:%s:%s:%s", subscriptionID, resourceGroup, vnetName)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		if vnet, ok := cached.(*armnetwork.VirtualNetwork); ok {
			return vnet, nil
		}
	}

	clients, err := c.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.VNetClient.Get(ctx, resourceGroup, vnetName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual network: %v", err)
	}

	vnet := &resp.VirtualNetwork
	// Store in cache
	c.cache.Set(cacheKey, vnet)

	return vnet, nil
}

// GetRouteTable retrieves information about the specified route table.
func (c *AzureClient) GetRouteTable(ctx context.Context, subscriptionID, resourceGroup, routeTableName string) (*armnetwork.RouteTable, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("resource:routetable:%s:%s:%s", subscriptionID, resourceGroup, routeTableName)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		if routeTable, ok := cached.(*armnetwork.RouteTable); ok {
			return routeTable, nil
		}
	}

	clients, err := c.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.RouteTableClient.Get(ctx, resourceGroup, routeTableName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get route table: %v", err)
	}

	routeTable := &resp.RouteTable
	// Store in cache
	c.cache.Set(cacheKey, routeTable)

	return routeTable, nil
}

// GetNetworkSecurityGroup retrieves information about the specified network security group.
func (c *AzureClient) GetNetworkSecurityGroup(ctx context.Context, subscriptionID, resourceGroup, nsgName string) (*armnetwork.SecurityGroup, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("resource:nsg:%s:%s:%s", subscriptionID, resourceGroup, nsgName)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		if nsg, ok := cached.(*armnetwork.SecurityGroup); ok {
			return nsg, nil
		}
	}

	clients, err := c.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.NSGClient.Get(ctx, resourceGroup, nsgName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get network security group: %v", err)
	}

	nsg := &resp.SecurityGroup
	// Store in cache
	c.cache.Set(cacheKey, nsg)

	return nsg, nil
}

// GetSubnet retrieves information about the specified subnet in a virtual network.
func (c *AzureClient) GetSubnet(ctx context.Context, subscriptionID, resourceGroup, vnetName, subnetName string) (*armnetwork.Subnet, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("resource:subnet:%s:%s:%s:%s", subscriptionID, resourceGroup, vnetName, subnetName)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		if subnet, ok := cached.(*armnetwork.Subnet); ok {
			return subnet, nil
		}
	}

	clients, err := c.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.SubnetsClient.Get(ctx, resourceGroup, vnetName, subnetName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subnet: %v", err)
	}

	subnet := &resp.Subnet
	// Store in cache
	c.cache.Set(cacheKey, subnet)

	return subnet, nil
}

// GetLoadBalancer retrieves information about the specified load balancer.
func (c *AzureClient) GetLoadBalancer(ctx context.Context, subscriptionID, resourceGroup, lbName string) (*armnetwork.LoadBalancer, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("resource:loadbalancer:%s:%s:%s", subscriptionID, resourceGroup, lbName)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		if lb, ok := cached.(*armnetwork.LoadBalancer); ok {
			return lb, nil
		}
	}

	clients, err := c.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.LoadBalancerClient.Get(ctx, resourceGroup, lbName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get load balancer: %v", err)
	}

	lb := &resp.LoadBalancer
	// Store in cache
	c.cache.Set(cacheKey, lb)

	return lb, nil
}

// GetPrivateEndpoint retrieves information about the specified private endpoint.
func (c *AzureClient) GetPrivateEndpoint(ctx context.Context, subscriptionID, resourceGroup, peName string) (*armnetwork.PrivateEndpoint, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("resource:privateendpoint:%s:%s:%s", subscriptionID, resourceGroup, peName)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		if pe, ok := cached.(*armnetwork.PrivateEndpoint); ok {
			return pe, nil
		}
	}

	clients, err := c.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.PrivateEndpointsClient.Get(ctx, resourceGroup, peName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get private endpoint: %v", err)
	}

	pe := &resp.PrivateEndpoint
	// Store in cache
	c.cache.Set(cacheKey, pe)

	return pe, nil
}

// GetPrivateEndpointByID retrieves information about the specified private endpoint using its resource ID.
func (c *AzureClient) GetPrivateEndpointByID(ctx context.Context, privateEndpointID string) (*armnetwork.PrivateEndpoint, error) {
	parsedID, err := arm.ParseResourceID(privateEndpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private endpoint ID: %v", err)
	}

	// Validate that the resource type is actually a private endpoint
	if parsedID.ResourceType.String() != "Microsoft.Network/privateEndpoints" {
		return nil, fmt.Errorf("invalid resource type: expected Microsoft.Network/privateEndpoints, got %s", parsedID.ResourceType.String())
	}

	return c.GetPrivateEndpoint(ctx, parsedID.SubscriptionID, parsedID.ResourceGroupName, parsedID.Name)
}

// GetVMSS retrieves information about the specified VMSS.
func (c *AzureClient) GetVMSS(ctx context.Context, subscriptionID, resourceGroup, vmssName string) (*armcompute.VirtualMachineScaleSet, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("resource:vmss:%s:%s:%s", subscriptionID, resourceGroup, vmssName)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		if vmss, ok := cached.(*armcompute.VirtualMachineScaleSet); ok {
			return vmss, nil
		}
	}

	clients, err := c.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.VMSSClient.Get(ctx, resourceGroup, vmssName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get VMSS: %v", err)
	}

	vmss := &resp.VirtualMachineScaleSet
	// Store in cache
	c.cache.Set(cacheKey, vmss)

	return vmss, nil
}

// Helper methods for working with resource IDs

// GetResourceByID retrieves a resource by its full Azure resource ID.
// It parses the ID, determines the resource type, and calls the appropriate method.
func (c *AzureClient) GetResourceByID(ctx context.Context, resourceID string) (interface{}, error) {
	// Parse the resource ID
	parsed, err := arm.ParseResourceID(resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource ID: %v", err)
	}

	// Based on the resource type, call the appropriate method
	switch parsed.ResourceType.String() {
	case "Microsoft.ContainerService/managedClusters":
		return c.GetAKSCluster(ctx, parsed.SubscriptionID, parsed.ResourceGroupName, parsed.Name)
	case "Microsoft.Network/virtualNetworks":
		return c.GetVirtualNetwork(ctx, parsed.SubscriptionID, parsed.ResourceGroupName, parsed.Name)
	case "Microsoft.Network/routeTables":
		return c.GetRouteTable(ctx, parsed.SubscriptionID, parsed.ResourceGroupName, parsed.Name)
	case "Microsoft.Network/networkSecurityGroups":
		return c.GetNetworkSecurityGroup(ctx, parsed.SubscriptionID, parsed.ResourceGroupName, parsed.Name)
	case "Microsoft.Network/loadBalancers":
		return c.GetLoadBalancer(ctx, parsed.SubscriptionID, parsed.ResourceGroupName, parsed.Name)
	case "Microsoft.Network/virtualNetworks/subnets":
		// For subnets, we need the VNet name from parent and subnet name
		if parsed.Parent != nil {
			return c.GetSubnet(ctx, parsed.SubscriptionID, parsed.ResourceGroupName, parsed.Parent.Name, parsed.Name)
		}
		return nil, fmt.Errorf("invalid subnet resource ID format: %s", resourceID)
	case "Microsoft.Compute/virtualMachineScaleSets":
		return c.GetVMSS(ctx, parsed.SubscriptionID, parsed.ResourceGroupName, parsed.Name)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", parsed.ResourceType)
	}
}

// GetCache returns the Azure cache instance
func (c *AzureClient) GetCache() *AzureCache {
	return c.cache
}

// GetDiagnosticSettings retrieves diagnostic settings for the specified resource.
func (c *AzureClient) GetDiagnosticSettings(ctx context.Context, subscriptionID, resourceURI string) ([]*armmonitor.DiagnosticSettingsResource, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("resource:diagnosticsettings:%s:%s", subscriptionID, resourceURI)

	// Check cache first
	if cached, found := c.cache.Get(cacheKey); found {
		if settings, ok := cached.([]*armmonitor.DiagnosticSettingsResource); ok {
			return settings, nil
		}
	}

	clients, err := c.GetOrCreateClientsForSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	pager := clients.DiagnosticSettingsClient.NewListPager(resourceURI, nil)
	var diagnosticSettings []*armmonitor.DiagnosticSettingsResource

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get diagnostic settings: %v", err)
		}
		diagnosticSettings = append(diagnosticSettings, page.Value...)
	}

	// Store in cache
	c.cache.Set(cacheKey, diagnosticSettings)

	return diagnosticSettings, nil
}
