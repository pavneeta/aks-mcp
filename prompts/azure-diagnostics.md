# AI Implementation Prompt: Azure Diagnostics and Advisory Tools for AKS-MCP

This prompt file provides specifications for implementing Azure diagnostic and advisory capabilities in the AKS-MCP server.

## Implementation Request

Generate code to add the following diagnostic and advisory tools to the AKS-MCP server:

1. **AppLens Detector Integration**
2. **Resource Health Event Access**
3. **Azure Advisor Recommendations**

## Required Functionality

### 1. AppLens Detector Tools

#### Tool: `invoke_applens_detector`
**Purpose**: Call and invoke AppLens detectors for AKS clusters

**Parameters**:
- `cluster_resource_id` (required): Full Azure resource ID of the AKS cluster
- `detector_name` (optional): Specific detector to run, if not provided, list available detectors
- `time_range` (optional): Time range for analysis (e.g., "24h", "7d", "30d")

**Expected Outputs**:
- List of available detectors with descriptions
- Detector execution results with findings and recommendations
- Severity levels and impact assessment
- Actionable remediation steps

**Implementation Requirements**:
- Use Azure Management SDK for AppLens API calls
- Handle authentication via Azure credential chain
- Support both listing detectors and executing specific detectors
- Parse and format detector results for readability
- Handle rate limiting and API quotas

#### Tool: `list_applens_detectors`
**Purpose**: List all available AppLens detectors for a cluster

**Parameters**:
- `cluster_resource_id` (required): Full Azure resource ID of the AKS cluster
- `category` (optional): Filter by detector category (performance, security, reliability)

**Expected Outputs**:
- Comprehensive list of available detectors
- Detector categories and descriptions
- Execution time estimates
- Prerequisites for each detector

### 2. Resource Health Event Tools

#### Tool: `get_resource_health_status`
**Purpose**: Access current resource health status for AKS clusters

**Parameters**:
- `resource_ids` (required): Array of Azure resource IDs (supports multiple clusters)
- `include_history` (optional): Boolean to include recent health events

**Expected Outputs**:
- Current health status (Available, Unavailable, Degraded, Unknown)
- Health summary with key metrics
- Active health issues and their impact
- Recommended actions for degraded health

#### Tool: `get_resource_health_events`
**Purpose**: Retrieve historical resource health events

**Parameters**:
- `resource_id` (required): Azure resource ID of the AKS cluster
- `start_time` (optional): Start time for historical query (ISO 8601 format)
- `end_time` (optional): End time for historical query (ISO 8601 format)
- `health_status_filter` (optional): Filter by health status types

**Expected Outputs**:
- Historical health events with timestamps
- Event duration and impact scope
- Root cause analysis when available
- Resolution status and time to resolution

**Implementation Requirements**:
- Use Azure Resource Health REST API
- Support filtering by time range and health status
- Handle large datasets with pagination
- Provide clear event categorization and severity

### 3. Azure Advisor Tools

#### Tool: `get_azure_advisor_recommendations`
**Purpose**: Access active Azure Advisor recommendations

**Parameters**:
- `subscription_id` (required): Azure subscription ID
- `resource_group` (optional): Filter by specific resource group
- `category` (optional): Filter by recommendation category (Cost, Performance, Security, Reliability)
- `severity` (optional): Filter by severity level (High, Medium, Low)

**Expected Outputs**:
- List of active recommendations with descriptions
- Severity levels and priority ranking
- Estimated impact and potential savings
- Implementation guidance and steps

#### Tool: `get_advisor_recommendation_details`
**Purpose**: Get detailed information about specific recommendations

**Parameters**:
- `recommendation_id` (required): Unique identifier for the recommendation
- `include_implementation_status` (optional): Include tracking of implementation progress

**Expected Outputs**:
- Detailed recommendation description
- Technical implementation steps
- Risk assessment and impact analysis
- Cost-benefit analysis where applicable

**Implementation Requirements**:
- Use Azure Advisor REST API
- Support filtering and querying capabilities
- Parse recommendation metadata and content
- Handle recommendation state changes and dismissals

## Technical Implementation Guidelines

### Authentication and Authorization
```go
// Use Azure SDK default credential chain
credential, err := azidentity.NewDefaultAzureCredential(nil)
if err != nil {
    return fmt.Errorf("failed to create Azure credential: %w", err)
}
```

### Error Handling
- Implement comprehensive error handling for API failures
- Provide meaningful error messages for permission issues
- Handle service outages and rate limiting gracefully
- Log diagnostic information for troubleshooting

### Data Processing
- Parse and format API responses for readability
- Implement caching for frequently accessed data
- Support real-time and historical data queries
- Provide data aggregation and correlation capabilities

### Integration with MCP Framework
- Follow existing MCP tool patterns in the codebase
- Integrate with current authentication and configuration systems
- Support all access levels (readonly, readwrite, admin)
- Maintain consistent error handling and logging

## Code Structure Requirements

### File Organization
```
internal/azure/
├── applens/
│   ├── client.go          # AppLens API client
│   ├── detectors.go       # Detector management
│   └── types.go           # AppLens data types
├── resourcehealth/
│   ├── client.go          # Resource Health API client
│   ├── events.go          # Health event handling
│   └── types.go           # Resource Health data types
└── advisor/
    ├── client.go          # Azure Advisor API client
    ├── recommendations.go # Recommendation handling
    └── types.go           # Advisor data types
```

### Tool Registration
```go
// Add to internal/server/server.go
func (s *Server) registerDiagnosticTools() {
    s.registerTool("invoke_applens_detector", s.handleAppLensDetector)
    s.registerTool("list_applens_detectors", s.handleListAppLensDetectors)
    s.registerTool("get_resource_health_status", s.handleResourceHealthStatus)
    s.registerTool("get_resource_health_events", s.handleResourceHealthEvents)
    s.registerTool("get_azure_advisor_recommendations", s.handleAdvisorRecommendations)
    s.registerTool("get_advisor_recommendation_details", s.handleAdvisorDetails)
}
```

### Configuration Support
- Add configuration options for API endpoints and timeouts
- Support custom authentication methods
- Allow configuration of default time ranges and filters
- Enable/disable specific diagnostic tools based on access level

## Testing Requirements

### Unit Tests
- Test each tool with various input parameters
- Mock Azure API responses for consistent testing
- Validate error handling and edge cases
- Test authentication and authorization scenarios

### Integration Tests
- Test with real Azure resources (in test environment)
- Validate API integration and data parsing
- Test performance with large datasets
- Verify cross-tool data correlation

### Example Test Cases
```go
func TestAppLensDetectorInvocation(t *testing.T) {
    // Test invoking specific detector
    // Test listing available detectors
    // Test error handling for invalid clusters
}

func TestResourceHealthEvents(t *testing.T) {
    // Test current health status retrieval
    // Test historical event queries
    // Test filtering and pagination
}

func TestAzureAdvisorRecommendations(t *testing.T) {
    // Test recommendation retrieval
    // Test filtering by category and severity
    // Test detailed recommendation access
}
```

## Documentation Requirements

### Tool Documentation
- Provide comprehensive tool descriptions
- Include parameter specifications and examples
- Document expected outputs and formats
- Include troubleshooting guides

### API Documentation
- Document Azure API endpoints used
- Include authentication requirements
- Provide rate limiting and quota information
- Include service availability considerations

## Success Criteria

### Functional Requirements
- ✅ Successfully invoke AppLens detectors and retrieve results
- ✅ Access current and historical Resource Health events
- ✅ Retrieve Azure Advisor recommendations with severity levels
- ✅ Provide actionable insights and recommendations
- ✅ Handle errors and edge cases gracefully

### Performance Requirements
- ✅ Respond to diagnostic queries within reasonable time (< 30s)
- ✅ Handle multiple concurrent requests efficiently
- ✅ Cache frequently accessed data appropriately
- ✅ Scale with cluster count and data volume

### Security Requirements
- ✅ Implement proper Azure authentication and authorization
- ✅ Respect Azure RBAC and subscription boundaries
- ✅ Protect sensitive diagnostic information
- ✅ Log security events and access attempts

### Integration Requirements
- ✅ Seamlessly integrate with existing AKS-MCP architecture
- ✅ Follow established code patterns and conventions
- ✅ Support all configured access levels
- ✅ Maintain backward compatibility

## Implementation Priority

1. **Phase 1**: Basic AppLens detector invocation
2. **Phase 2**: Resource Health event access
3. **Phase 3**: Azure Advisor recommendation retrieval
4. **Phase 4**: Advanced filtering and correlation features
5. **Phase 5**: Performance optimization and caching

Generate the implementation code following these specifications, ensuring robust error handling, comprehensive testing, and clear documentation.