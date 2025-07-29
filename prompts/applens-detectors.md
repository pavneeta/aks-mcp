# AppLens Detectors for AKS MCP

## Overview
This document outlines the implementation plan for AKS detector functionality that provides:
- `list_detectors`: List all available AKS detectors
- `run_detector`: Execute a specific detector by name
- `run_detectors_by_category`: Execute all detectors in a specific category

## Requirements

### API Endpoints
- **List API**: `GET https://management.azure.com/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.ContainerService/managedClusters/{clusterName}/detectors?api-version=2024-08-01`
- **Run Detector API**: `GET https://management.azure.com/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/microsoft.containerservice/managedclusters/{clusterName}/detectors/{detectorName}?startTime={startTime}&endTime={endTime}&api-version=2024-08-01`

### Parameters
- `cluster_resource_id`: AKS cluster resource ID (format: `/subscriptions/{id}/resourceGroups/{rg}/providers/Microsoft.ContainerService/managedClusters/{name}`)
- `detector_name`: Detector name from list_detectors result
- `category`: Must be one of:
  - Best Practices
  - Cluster and Control Plane Availability and Performance
  - Connectivity Issues
  - Create, Upgrade, Delete and Scale
  - Deprecations
  - Identity and Security
  - Node Health
  - Storage
- `start_time`: Start time (within last 30 days)
- `end_time`: End time (within last 30 days, max 24h range)

### Caching Strategy
- Cache detector list results in memory using existing AzureCache
- Cache key: `detectors:list:{subscriptionId}:{resourceGroup}:{clusterName}`
- Cache TTL: Use default cache timeout from config

## Implementation Architecture

### 1. File Structure
Following existing patterns in the codebase:

```
internal/components/detectors/
├── handlers.go          # MCP tool handlers
├── registry.go          # Tool registration
├── types.go            # Data structures
├── client.go           # Azure API client wrapper
└── detectors_test.go   # Unit tests
```

### 2. Core Components

#### 2.1 Data Structures (`types.go`)
```go
// DetectorListResponse represents the API response for listing detectors
type DetectorListResponse struct {
    Value []Detector `json:"value"`
}

// Detector represents a single detector metadata
type Detector struct {
    ID         string             `json:"id"`
    Name       string             `json:"name"`
    Type       string             `json:"type"`
    Location   string             `json:"location"`
    Properties DetectorProperties `json:"properties"`
}

// DetectorProperties contains detector metadata
type DetectorProperties struct {
    Metadata DetectorMetadata `json:"metadata"`
    Status   DetectorStatus   `json:"status"`
}

// DetectorMetadata contains detector information
type DetectorMetadata struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Category    string `json:"category"`
    Description string `json:"description"`
    Type        string `json:"type"`
}

// DetectorStatus contains detector status
type DetectorStatus struct {
    Message  *string `json:"message"`
    StatusID int     `json:"statusId"`
}

// DetectorRunResponse represents the API response for running a detector
type DetectorRunResponse struct {
    ID         string                    `json:"id"`
    Name       string                    `json:"name"`
    Type       string                    `json:"type"`
    Location   string                    `json:"location"`
    Properties DetectorRunProperties    `json:"properties"`
}

// DetectorRunProperties contains detector run results
type DetectorRunProperties struct {
    Dataset  []DetectorDataset    `json:"dataset"`
    Metadata DetectorMetadata     `json:"metadata"`
    Status   DetectorStatus       `json:"status"`
}

// DetectorDataset represents detector output data
type DetectorDataset struct {
    RenderingProperties RenderingProperties `json:"renderingProperties"`
    Table              DetectorTable       `json:"table"`
}

// RenderingProperties defines how to display results
type RenderingProperties struct {
    Description *string `json:"description"`
    IsVisible   bool    `json:"isVisible"`
    Title       *string `json:"title"`
    Type        int     `json:"type"`
}

// DetectorTable contains tabular detector results
type DetectorTable struct {
    Columns   []DetectorColumn `json:"columns"`
    Rows      [][]interface{}  `json:"rows"`
    TableName string          `json:"tableName"`
}

// DetectorColumn defines table column metadata
type DetectorColumn struct {
    ColumnName string  `json:"columnName"`
    ColumnType *string `json:"columnType"`
    DataType   string  `json:"dataType"`
}
```

#### 2.2 Client Wrapper (`client.go`)
```go
// DetectorClient wraps Azure API calls with caching
type DetectorClient struct {
    azClient *azureclient.AzureClient
    cache    *azureclient.AzureCache
}

// NewDetectorClient creates a new detector client
func NewDetectorClient(azClient *azureclient.AzureClient) *DetectorClient

// ListDetectors lists all detectors for a cluster with caching
func (c *DetectorClient) ListDetectors(ctx context.Context, subscriptionID, resourceGroup, clusterName string) (*DetectorListResponse, error)

// RunDetector executes a specific detector
func (c *DetectorClient) RunDetector(ctx context.Context, subscriptionID, resourceGroup, clusterName, detectorName, startTime, endTime string) (*DetectorRunResponse, error)

// GetDetectorsByCategory filters detectors by category from cached list
func (c *DetectorClient) GetDetectorsByCategory(ctx context.Context, subscriptionID, resourceGroup, clusterName, category string) ([]Detector, error)
```

#### 2.3 Handlers (`handlers.go`)
```go
// GetListDetectorsHandler returns handler for list_detectors tool
func GetListDetectorsHandler(azClient *azureclient.AzureClient, cfg *config.ConfigData) tools.ResourceHandler

// GetRunDetectorHandler returns handler for run_detector tool
func GetRunDetectorHandler(azClient *azureclient.AzureClient, cfg *config.ConfigData) tools.ResourceHandler

// GetRunDetectorsByCategoryHandler returns handler for run_detectors_by_category tool
func GetRunDetectorsByCategoryHandler(azClient *azureclient.AzureClient, cfg *config.ConfigData) tools.ResourceHandler

// HandleListDetectors implements the list_detectors functionality
func HandleListDetectors(params map[string]interface{}, client *DetectorClient) (string, error)

// HandleRunDetector implements the run_detector functionality
func HandleRunDetector(params map[string]interface{}, client *DetectorClient) (string, error)

// HandleRunDetectorsByCategory implements the run_detectors_by_category functionality
func HandleRunDetectorsByCategory(params map[string]interface{}, client *DetectorClient) (string, error)
```

#### 2.4 Tool Registration (`registry.go`)
```go
// RegisterListDetectorsTool registers the list_detectors MCP tool
func RegisterListDetectorsTool() *mcp.Tool

// RegisterRunDetectorTool registers the run_detector MCP tool
func RegisterRunDetectorTool() *mcp.Tool

// RegisterRunDetectorsByCategoryTool registers the run_detectors_by_category MCP tool
func RegisterRunDetectorsByCategoryTool() *mcp.Tool
```

### 3. Integration Points

#### 3.1 Azure Client Extension
Extend `internal/azureclient/client.go` to add detector-specific HTTP client:

```go
// Add to SubscriptionClients struct
type SubscriptionClients struct {
    // ... existing fields
    HTTPClient *http.Client  // For custom API calls
}

// Add method to AzureClient
func (c *AzureClient) MakeDetectorAPICall(ctx context.Context, url string, subscriptionID string) (*http.Response, error)
```

#### 3.2 Server Registration
Update `internal/server/server.go` to register detector tools:

```go
// Add to registerAzureResourceTools method
func (s *Service) registerDetectorTools(azClient *azureclient.AzureClient) {
    log.Println("Registering Detector tools...")

    // Register list detectors tool
    log.Println("Registering detector tool: list_detectors")
    listTool := detectors.RegisterListDetectorsTool()
    s.mcpServer.AddTool(listTool, tools.CreateResourceHandler(detectors.GetListDetectorsHandler(azClient, s.cfg), s.cfg))

    // Register run detector tool
    log.Println("Registering detector tool: run_detector")
    runTool := detectors.RegisterRunDetectorTool()
    s.mcpServer.AddTool(runTool, tools.CreateResourceHandler(detectors.GetRunDetectorHandler(azClient, s.cfg), s.cfg))

    // Register run detectors by category tool
    log.Println("Registering detector tool: run_detectors_by_category")
    categoryTool := detectors.RegisterRunDetectorsByCategoryTool()
    s.mcpServer.AddTool(categoryTool, tools.CreateResourceHandler(detectors.GetRunDetectorsByCategoryHandler(azClient, s.cfg), s.cfg))
}
```

### 4. Tool Definitions

#### 4.1 list_detectors Tool
```json
{
  "name": "list_detectors",
  "description": "List all available AKS cluster detectors",
  "inputSchema": {
    "type": "object",
    "properties": {
      "cluster_resource_id": {
        "type": "string",
        "description": "AKS cluster resource ID"
      }
    },
    "required": ["cluster_resource_id"]
  }
}
```

#### 4.2 run_detector Tool
```json
{
  "name": "run_detector",
  "description": "Run a specific AKS detector",
  "inputSchema": {
    "type": "object",
    "properties": {
      "cluster_resource_id": {
        "type": "string",
        "description": "AKS cluster resource ID"
      },
      "detector_name": {
        "type": "string",
        "description": "Name of the detector to run"
      },
      "start_time": {
        "type": "string",
        "description": "Start time in ISO format (within last 30 days)"
      },
      "end_time": {
        "type": "string",
        "description": "End time in ISO format (within last 30 days, max 24h from start)"
      }
    },
    "required": ["cluster_resource_id", "detector_name", "start_time", "end_time"]
  }
}
```

#### 4.3 run_detectors_by_category Tool
```json
{
  "name": "run_detectors_by_category",
  "description": "Run all detectors in a specific category",
  "inputSchema": {
    "type": "object",
    "properties": {
      "cluster_resource_id": {
        "type": "string",
        "description": "AKS cluster resource ID"
      },
      "category": {
        "type": "string",
        "enum": [
          "Best Practices",
          "Cluster and Control Plane Availability and Performance",
          "Connectivity Issues",
          "Create, Upgrade, Delete and Scale",
          "Deprecations",
          "Identity and Security",
          "Node Health",
          "Storage"
        ],
        "description": "Detector category to run"
      },
      "start_time": {
        "type": "string",
        "description": "Start time in ISO format (within last 30 days)"
      },
      "end_time": {
        "type": "string",
        "description": "End time in ISO format (within last 30 days, max 24h from start)"
      }
    },
    "required": ["cluster_resource_id", "category", "start_time", "end_time"]
  }
}
```

### 5. Implementation Flow

#### 5.1 list_detectors Flow
1. Parse `cluster_resource_id` to extract subscription, resource group, cluster name
2. Check cache for detector list using key: `detectors:list:{subscription}:{resourceGroup}:{cluster}`
3. If not cached, make HTTP GET request to Azure Management API
4. Parse response into `DetectorListResponse` struct
5. Cache the response with default TTL
6. Return JSON-formatted detector list

#### 5.2 run_detector Flow
1. Parse `cluster_resource_id` and validate time parameters
2. Construct detector API URL with query parameters
3. Make HTTP GET request with Authorization header
4. Parse response into `DetectorRunResponse` struct
5. Return JSON-formatted detector results

#### 5.3 run_detectors_by_category Flow
1. Call `GetDetectorsByCategory()` to get filtered detector list
2. For each detector in category, call `RunDetector()`
3. Aggregate all results into a single response
4. Return combined JSON results

### 6. Error Handling

#### 6.1 Input Validation
- Validate cluster resource ID format
- Validate time range (within 30 days, max 24h duration)
- Validate category against allowed values
- Validate detector name exists in cached list

#### 6.2 API Error Handling
- Handle Azure authentication errors
- Handle rate limiting (429) with retry logic
- Handle not found (404) for invalid clusters/detectors
- Handle timeout errors with appropriate messages

#### 6.3 Cache Error Handling
- Graceful fallback when cache is unavailable
- Cache invalidation on API errors
- Logging for cache operations

### 7. Security Considerations

#### 7.1 Authentication
- Use existing `DefaultAzureCredential` pattern
- Ensure proper RBAC permissions for detector APIs
- Add User-Agent header: "AKS-MCP"

#### 7.2 Input Sanitization
- Validate all input parameters
- URL-encode query parameters
- Prevent injection attacks in resource IDs

### 8. Testing Strategy

#### 8.1 Unit Tests
- Test detector list parsing
- Test detector run result parsing
- Test cache operations
- Test input validation
- Test error scenarios

#### 8.2 Integration Tests
- Test against live Azure API (with test clusters)
- Test caching behavior
- Test concurrent access

### 9. Documentation Updates

#### 9.1 README Updates
- Add detector tools to tool list
- Include usage examples
- Document required Azure permissions

#### 9.2 New Documentation
- Create `prompts/aks-detectors.md` with detailed usage guide
- Include example outputs
- Document troubleshooting steps

## Implementation Timeline

1. **Phase 1**: Core types and client wrapper (types.go, client.go)
2. **Phase 2**: Handler implementations (handlers.go)
3. **Phase 3**: Tool registration and server integration (registry.go, server.go updates)
4. **Phase 4**: Testing and documentation
5. **Phase 5**: Error handling refinements and edge cases

## Dependencies

- Existing Azure client and cache infrastructure
- Azure Management API access permissions
- Go HTTP client for custom API calls
- JSON marshaling/unmarshaling for API responses