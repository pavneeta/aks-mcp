# AKS Private Endpoint for AKS-MCP

## Overview
This document outlines the implementation plan for AKS private endpoint functionality that provides:
- `get_private_endpoint_info`: Get information about the private endpoint used by private AKS clusters

## Requirements

### Private Endpoint Behavior
- Only private AKS clusters have private endpoints configured
- The private endpoint is located in the cluster's node resource group
- The private endpoint is named `kube-apiserver`
- For non-private clusters, the tool should return a clear message indicating no private endpoint is configured

### Azure Resource Details
- **Resource Type**: `Microsoft.Network/privateEndpoints`
- **Location**: Node resource group of the AKS cluster
- **Naming Convention**: `kube-apiserver`
- **Related Resources**: 
  - Network Interface Cards (NICs) attached to the private endpoint
  - Private DNS zone groups for DNS resolution
  - Subnet association for the private endpoint

### Tool Functionality
The tool should return comprehensive private endpoint information including:
- Basic private endpoint properties (ID, name, location, provisioning state)
- Network interface details (private IP addresses, subnet associations)
- Private DNS zone group information
- Connection details to the AKS API server
- Subnet information where the private endpoint is deployed

## Implementation Architecture

### 1. File Structure
Following existing patterns in the codebase:

```
internal/components/network/
├── handlers.go              # Add GetPrivateEndpointInfoHandler
├── registry.go              # Add RegisterPrivateEndpointInfoTool
├── registry_test.go         # Add tests for private endpoint tool registration
├── handlers_test.go         # Add tests for private endpoint handler
└── resourcehelpers/
    ├── privateendpointhelpers.go       # New file for private endpoint logic
    └── privateendpointhelpers_test.go  # Unit tests for helper functions
```

### 2. Core Components

#### 2.1 Data Structures
No custom data structures needed - return raw Azure ARM API responses as JSON strings. This follows the pattern of existing network tools that return Azure SDK responses directly.

#### 2.2 Tool Registration
Add new tool registration function following existing patterns in `registry.go`.

#### 2.3 Handler Implementation
Add new handler function in `handlers.go` that:
- Validates input parameters
- Retrieves AKS cluster information
- Detects if cluster is private or public
- Returns appropriate response for each scenario

#### 2.4 Helper Functions
Create helper functions in `resourcehelpers/privateendpointhelpers.go` to:
- Extract subscription ID from cluster resource ID
- Check if cluster has private endpoint enabled
- Search for private endpoint in node resource group
- Handle edge cases and error scenarios

### 3. Integration Points

#### 3.1 Azure Client Extensions
Extend the Azure client to support private endpoints by:
- Adding PrivateEndpointsClient to SubscriptionClients struct
- Adding methods to retrieve private endpoint details
- Implementing proper caching for private endpoint resources

#### 3.2 Server Registration
Update the server registration to include the new private endpoint tool alongside other network tools.

### 4. Testing Strategy

#### 4.1 Unit Tests
Create comprehensive unit tests for:
- Helper function validation with various cluster configurations
- Handler parameter validation and error handling
- Tool registration verification

#### 4.2 Integration Testing
Test with real AKS clusters:
- Private clusters (should return private endpoint details)
- Public clusters (should return appropriate message)
- Invalid/non-existent clusters (should return proper errors)

### 5. Error Handling and Edge Cases

#### 5.1 Cluster Type Detection
- **Private Cluster**: Return private endpoint details as JSON
- **Public Cluster**: Return message indicating no private endpoint (not an error)
- **Invalid Cluster**: Return appropriate error messages

#### 5.2 Access Control
- Follow existing access level patterns (readonly access is sufficient)
- Validate subscription and resource group access through Azure RBAC

#### 5.3 Caching Strategy
- Leverage existing Azure client caching mechanisms
- Cache private endpoint lookups similar to other network resources

### 6. Usage Examples

#### 6.1 Private Cluster Response
Returns detailed private endpoint information including network interfaces, private link service connections, and subnet associations.

#### 6.2 Public Cluster Response
Returns a clear message indicating the cluster is not configured as a private cluster.

### 7. Security Considerations

- Follow existing authentication and authorization patterns
- Ensure proper subscription-level access validation
- No additional security risks beyond existing network tools
- Private endpoint information is read-only and follows Azure RBAC

## Summary

This implementation provides a comprehensive solution for AKS private endpoint discovery and information retrieval. The tool seamlessly integrates with the existing network tools architecture, following established patterns for Azure resource access, caching, and error handling.

The implementation successfully handles both private and public cluster scenarios, providing clear and useful responses for each case while maintaining consistency with the existing codebase patterns.
