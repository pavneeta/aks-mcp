# Feature: Azure Resource Caching System

## Description
This feature implements a comprehensive in-memory caching system for Azure resources that significantly improves performance by reducing Azure API calls and providing faster response times for frequently accessed resources. The caching system is integrated into the Azure client and provides automatic cache management with configurable expiration times.

## How it Works
The caching system intercepts Azure SDK calls and stores resource information in memory with time-based expiration. When a resource is requested, the system first checks the cache before making an API call to Azure. This reduces latency, minimizes Azure API rate limiting, and improves overall server performance.

## Architecture

### Core Components
- **Azure Cache** - Generic in-memory cache implementation with time-based expiration
- **Azure Client Integration** - Cache integration with Azure SDK clients for transparent caching
- **Configuration Management** - Configurable cache timeout and expiration settings
- **Thread Safety** - Concurrent access protection with read-write mutexes
- **Automatic Expiration** - Time-based cache invalidation and cleanup

## Cached Resource Types

### Azure Kubernetes Service Resources
- **AKS Clusters**: Complete cluster configuration and status
- **Node Pools**: Node pool details and configuration
- **Cluster Credentials**: Authentication and access information

### Networking Resources
- **Virtual Networks (VNets)**: VNet configuration and address spaces
- **Subnets**: Subnet details and IP allocations
- **Network Security Groups (NSGs)**: Security rules and associations
- **Route Tables**: Routing configuration and rules
- **Load Balancers**: Load balancer configuration and backend pools

### Resource Metadata
- **Resource Hierarchies**: Parent-child relationships between resources
- **Resource IDs**: Azure resource identifiers and references
- **Resource States**: Current operational state of resources

## Cache Key Strategy

### Hierarchical Key Structure
Cache keys follow a structured pattern for easy management and retrieval:

```
Format: resource:type:subscription:resourcegroup:name

Examples:
- "resource:cluster:12345678-1234-1234-1234-123456789012:myRG:myCluster"
- "resource:vnet:12345678-1234-1234-1234-123456789012:networkRG:myVNet"
- "resource:nsg:12345678-1234-1234-1234-123456789012:aksRG:myNSG"
- "resource:routetable:12345678-1234-1234-1234-123456789012:aksRG:myRT"
```

### Key Benefits
- **Predictable Structure**: Easy to construct and understand
- **Collision Avoidance**: Unique keys across all Azure subscriptions
- **Scope Isolation**: Resources isolated by subscription and resource group
- **Type Organization**: Clear resource type identification
