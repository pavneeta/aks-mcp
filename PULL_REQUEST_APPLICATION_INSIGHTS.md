# Azure Application Insights Tool for AKS-MCP

## 🎯 **Overview**
This PR adds a comprehensive specification for Azure Application Insights integration to the AKS-MCP server. The tool enables AI assistants to query Application Insights telemetry data for applications running in AKS clusters, providing deep observability and performance monitoring capabilities.

## 🚀 **Features Specified**

### **Core Functionality**
- ✅ **Application Insights Query Tool**: Execute KQL queries against Application Insights telemetry data
- ✅ **Multiple Time Filters**: Support start_time, end_time, and timespan parameters for flexible time range queries
- ✅ **KQL Query Validation**: Security validation to prevent dangerous operations and injection attacks
- ✅ **Raw JSON Output**: Returns standard Azure CLI JSON output for consistent data format
- ✅ **Azure CLI Integration**: Uses existing `azcli.NewExecutor()` pattern for command execution

### **Telemetry Data Access**
- ✅ **Requests**: HTTP request performance and success metrics
- ✅ **Dependencies**: External service calls and database connections
- ✅ **Exceptions**: Application errors and exception tracking
- ✅ **Traces**: Application logs and diagnostic trace data
- ✅ **CustomEvents**: Custom telemetry logged by applications

### **Security & Validation**
- ✅ **Parameter Validation**: Required fields validation for subscription, resource group, app name, and query
- ✅ **KQL Safety**: Prevents dangerous keywords (delete, drop, create, alter, insert, update)
- ✅ **Table Whitelisting**: Ensures queries start with valid Application Insights table names
- ✅ **Time Format Validation**: RFC3339 (ISO 8601) format validation for time parameters

## 📋 **Tool Specification**

### **Tool Name**: `az_monitor_app_insights_query`

### **Parameters**:
- `subscription_id` (required): Azure subscription ID
- `resource_group` (required): Resource group containing the Application Insights resource
- `app_insights_name` (required): Application Insights resource name
- `query` (required): KQL query to execute against telemetry data
- `start_time` (optional): Start time for query in ISO 8601 format
- `end_time` (optional): End time for query (defaults to current time)
- `timespan` (optional): Query timespan (e.g., "PT1H", "P1D")

### **Azure CLI Command**:
```bash
az monitor app-insights query \
  --app /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Insights/components/{app} \
  --analytics-query "{kql-query}" \
  --start-time {time} \
  --output json
```

## 📊 **KQL Query Examples**

### **Performance Monitoring**
```kql
# Recent requests
requests | where timestamp > ago(1h) | limit 10

# Request performance over time
requests | where timestamp > ago(1h) | summarize avg(duration), count() by bin(timestamp, 5m)

# Error rate by operation
requests | where timestamp > ago(1h) | summarize total=count(), errors=countif(success == false) by operation_Name
```

### **Dependency Analysis**
```kql
# External dependency calls
dependencies | where timestamp > ago(1h) | summarize count(), avg(duration) by type, target
```

### **Exception Tracking**
```kql
# Recent exceptions
exceptions | where timestamp > ago(1h) | project timestamp, type, method, outerMessage
```

## 🔧 **Usage Examples**

### **Basic Query**
```bash
az_monitor_app_insights_query \
  --subscription-id 82d6efa7-b1b6-4aa0-ab12-d10788552670 \
  --resource-group thomas \
  --app-insights-name thomastest39-insights \
  --query "requests | where timestamp > ago(1h) | limit 10"
```

### **Time Range Query**
```bash
az_monitor_app_insights_query \
  --subscription-id 82d6efa7-b1b6-4aa0-ab12-d10788552670 \
  --resource-group thomas \
  --app-insights-name thomastest39-insights \
  --query "dependencies | summarize count() by type, target" \
  --start-time 2025-07-17T00:00:00Z \
  --end-time 2025-07-18T00:00:00Z
```

## 🏗️ **Integration Architecture**

### **File Organization**
- **Specification**: `prompts/azure-application-insights.md` - Complete tool specification
- **Handler Integration**: Extend `internal/components/monitor/handlers.go`
- **Tool Registration**: Extend `internal/components/monitor/registry.go`
- **Testing**: Add `appinsights_test.go` for unit tests

### **Integration Points**
- ✅ **Azure CLI Executor**: Uses existing `azcli.NewExecutor()` pattern
- ✅ **Monitor Component**: Integrates with existing monitor infrastructure
- ✅ **Security Validation**: Follows established validation patterns
- ✅ **Error Handling**: Uses consistent error handling from other tools

## 🔒 **Security & Access Control**

### **Access Levels**
- **Readonly**: All query operations (Application Insights querying is inherently read-only)
- **Readwrite**: Same as readonly (no write operations for telemetry data)
- **Admin**: Same as readonly (no admin-specific query operations)

### **Security Measures**
- **Query Validation**: Prevents dangerous SQL/KQL operations
- **Table Restrictions**: Only allows queries against valid Application Insights tables
- **Parameter Sanitization**: Validates all input parameters
- **Azure RBAC**: Respects existing Azure role-based access control

## 📈 **Benefits**

### **For AKS Monitoring**
- **Application Performance**: Monitor request latency, throughput, and error rates
- **Dependency Tracking**: Analyze external service calls and database performance
- **Error Analysis**: Track exceptions and diagnose application issues
- **Custom Telemetry**: Query custom events and metrics logged by applications

### **For AI Assistants**
- **Rich Query Capabilities**: Execute complex KQL queries for deep insights
- **Time-based Analysis**: Flexible time range filtering for historical analysis
- **Performance Correlation**: Correlate application metrics with infrastructure data
- **Troubleshooting Support**: Query exception data and traces for issue resolution

## ✅ **Implementation Roadmap**

### **Phase 1: Core Implementation**
1. Implement basic KQL query execution with time filtering
2. Add parameter validation and safety checks
3. Integrate with existing monitor component structure

### **Phase 2: Enhanced Features**
1. Add advanced query validation and optimization
2. Implement caching for common queries
3. Add comprehensive error handling and logging

### **Phase 3: Integration & Testing**
1. Complete integration with MCP server
2. Add comprehensive unit tests
3. Performance optimization and documentation

## 🔄 **Integration with Existing Tools**

This Application Insights tool complements existing AKS-MCP tools:

1. **Resource Health**: Correlate application issues with infrastructure health
2. **Azure Monitor**: Combine infrastructure metrics with application telemetry
3. **Azure Advisor**: Act on performance recommendations using telemetry insights
4. **Network Tools**: Analyze network-related dependency issues

## 🎯 **Success Criteria**
- ✅ Execute KQL queries against Application Insights telemetry data
- ✅ Filter by time range using multiple time specification methods
- ✅ Return raw Azure CLI JSON output with telemetry data
- ✅ Validate KQL queries for safety and correctness
- ✅ Provide meaningful error messages for invalid parameters
- ✅ Integrate seamlessly with existing MCP tool framework

## 📝 **Next Steps**

1. **Review Specification**: Review the prompt file for completeness and accuracy
2. **Implementation Planning**: Plan the implementation phases and timeline
3. **Code Development**: Begin implementation following the specification
4. **Testing Strategy**: Develop comprehensive testing approach
5. **Documentation**: Create user documentation and examples

This specification provides a solid foundation for implementing comprehensive Application Insights monitoring capabilities in the AKS-MCP server, enabling rich application observability for AKS workloads.

---

**Ready for implementation! 🚀**
