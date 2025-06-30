// Package azure provides Azure SDK integration for AKS MCP server.
package azure

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// SubscriptionClients contains Azure clients for a specific subscription.
type SubscriptionClients struct {
	SubscriptionID         string
	ContainerServiceClient *armcontainerservice.ManagedClustersClient
	VNetClient             *armnetwork.VirtualNetworksClient
	SubnetsClient          *armnetwork.SubnetsClient
	RouteTableClient       *armnetwork.RouteTablesClient
	NSGClient              *armnetwork.SecurityGroupsClient
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

	// Create and store the clients
	clients = &SubscriptionClients{
		SubscriptionID:         subscriptionID,
		ContainerServiceClient: containerServiceClient,
		VNetClient:             vnetClient,
		SubnetsClient:          subnetsClient,
		RouteTableClient:       routeTableClient,
		NSGClient:              nsgClient,
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

// Helper methods for working with resource IDs

// GetResourceByID retrieves a resource by its full Azure resource ID.
// It parses the ID, determines the resource type, and calls the appropriate method.
func (c *AzureClient) GetResourceByID(ctx context.Context, resourceID string) (interface{}, error) {
	// Parse the resource ID
	parsed, err := ParseResourceID(resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource ID: %v", err)
	}

	// Based on the resource type, call the appropriate method
	switch parsed.ResourceType {
	case ResourceTypeAKSCluster:
		return c.GetAKSCluster(ctx, parsed.SubscriptionID, parsed.ResourceGroup, parsed.ResourceName)
	case ResourceTypeVirtualNetwork:
		return c.GetVirtualNetwork(ctx, parsed.SubscriptionID, parsed.ResourceGroup, parsed.ResourceName)
	case ResourceTypeRouteTable:
		return c.GetRouteTable(ctx, parsed.SubscriptionID, parsed.ResourceGroup, parsed.ResourceName)
	case ResourceTypeSecurityGroup:
		return c.GetNetworkSecurityGroup(ctx, parsed.SubscriptionID, parsed.ResourceGroup, parsed.ResourceName)
	case ResourceTypeSubnet:
		return c.GetSubnet(ctx, parsed.SubscriptionID, parsed.ResourceGroup, parsed.ResourceName, parsed.SubResourceName)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", parsed.ResourceType)
	}
}
