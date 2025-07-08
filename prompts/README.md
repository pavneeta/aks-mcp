# AKS-MCP Prompts Folder

This folder includes all the prompts for the AKS-MCP server. These prompt files are designed for generating the functionality of the AKS-MCP (Azure Kubernetes Service - Model Context Protocol) server with AI assistants.


## Existing Files

### Current Prompt Files

- **`README.md`** - This file, describing the prompts folder and its contents
- **`monitoringservice.md`** - Feature requirements and implementation details for MCP monitoring service integration
- **`azure-diagnostics.md`** - Implementation specifications for Azure diagnostic and advisory tools (AppLens detectors, Resource Health, Azure Advisor)
- **`azure-advisor.md`** - Focused implementation specifications for Azure Advisor recommendations integration

### File Structure

```
prompts/
├── README.md                    # This documentation file
├── monitoringservice.md        # Monitoring service integration requirements
├── azure-diagnostics.md        # Azure diagnostics and advisory tools specifications
└── azure-advisor.md            # Azure Advisor recommendations integration
```

## AKS-MCP Server Capabilities

The prompts in this folder are designed to test and validate the following AKS-MCP server capabilities:

### Core AKS Operations
- Cluster information retrieval and management
- AKS cluster listing and discovery across subscriptions
- Cluster health status and monitoring
- Node and resource management

### Network Operations
- Virtual Network (VNet) information and configuration
- Subnet details and IP address management
- Network Security Group (NSG) rules and policies
- Route table information and network routing
- Load balancer and ingress configuration

### Monitoring and Observability
- Azure Monitor integration and dashboard access
- Log Analytics workspace queries and log retrieval
- Application Insights performance monitoring and tracing
- Alert management and notification systems
- Performance metrics collection and analysis
- Real-time monitoring and data visualization

### Diagnostics and Advisory
- AppLens detector integration and execution
- Resource Health event monitoring and access
- Azure Advisor recommendations retrieval
- Automated diagnostic workflows and reporting
- Proactive issue detection and remediation guidance

### Security and Access Control
- Access level validation (readonly, readwrite, admin)
- Security policy enforcement and validation
- Authentication and authorization testing
- Role-based access control (RBAC) verification

## Adding New Prompt Files

When adding new prompt files to this folder:

1. **Follow naming conventions**: Use descriptive, lowercase names with hyphens (e.g., `cluster-operations.md`)
2. **Include clear objectives**: Each file should specify what functionality it tests
3. **Provide expected behaviors**: Document what responses are expected for each prompt
4. **Add prerequisites**: List any required setup, permissions, or resources
5. **Update this README**: Add the new file to the "Existing Files" section above

