# Feature: Azure Resource Information Tools

## Description
This feature implements MCP (Model Context Protocol) server tools that enable AI assistants to retrieve detailed information about Azure networking resources associated with AKS clusters. It provides direct Azure SDK integration to access VNets, Network Security Groups, Route Tables, Subnets, and Load Balancers.

## How it Works
The server uses Azure SDK clients to query Azure Resource Manager APIs and retrieve detailed configuration information for networking resources that are associated with AKS clusters. Unlike CLI tools, these use direct API calls for faster and more structured responses.

## Architecture

### Core Components
- **Azure Client** - Multi-subscription Azure SDK client with caching capabilities
- **Resource Handlers** - Handlers for each resource type that process requests and return structured data
- **Resource Helpers** - Helper functions to discover resource relationships and extract resource IDs
- **Tool Registry** - MCP tool definitions for each resource type with parameter specifications
- **Caching Layer** - Resource caching for performance optimization

### Supported Resource Types

1. **Virtual Networks (VNet)**
   - Tool: `get_vnet_info`
   - Retrieves VNet configuration, address spaces, and subnets

2. **Network Security Groups (NSG)**
   - Tool: `get_nsg_info`
   - Retrieves NSG rules, security configurations, and associations

3. **Route Tables**
   - Tool: `get_route_table_info`
   - Retrieves routing rules and route configurations

4. **Subnets**
   - Tool: `get_subnet_info`
   - Retrieves subnet configuration, IP ranges, and associated resources

5. **Load Balancers**
   - Tool: `get_load_balancers_info`
   - Retrieves both external and internal load balancer configurations

## Requirements

### Prerequisites
- Valid Azure authentication (service principal, managed identity, or Azure CLI)
- Azure subscription access with appropriate read permissions
- Network Contributor or Reader role on the resources being queried

### Dependencies
- **Azure SDK for Go**: Required for Azure Resource Manager API interactions
- **Azure Identity**: Default Azure credential chain for authentication
- **MCP Go**: For MCP protocol implementation and tool definitions

## Implementation Details

### Resource Discovery Process
1. **Cluster Lookup**: First, retrieve the AKS cluster details using subscription ID, resource group, and cluster name
2. **Resource ID Extraction**: Extract networking resource IDs from the cluster configuration
3. **Resource Details Retrieval**: Query Azure APIs for detailed resource information
4. **Structured Response**: Return resource data in structured JSON format

### Caching Strategy
- **Multi-level Caching**: Client-level caching for Azure SDK clients and resource-level caching for API responses
- **Subscription-based Clients**: Separate Azure clients per subscription for better performance
- **Cache Invalidation**: Automatic cache refresh based on configurable time-to-live

### Error Handling
- **Resource Not Found**: Graceful handling when resources don't exist or aren't associated
- **Permission Errors**: Clear error messages for insufficient permissions
- **Network Timeouts**: Retry logic for transient network failures

## Tool Specifications

### Common Parameters
All resource tools require these parameters:
- `subscription_id` (required): Azure Subscription ID
- `resource_group` (required): Azure Resource Group containing the AKS cluster
- `cluster_name` (required): Name of the AKS cluster

