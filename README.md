# AKS-MCP

The AKS-MCP is a Model Context Protocol (MCP) server that enables AI assistants to interact with Azure Kubernetes Service (AKS) clusters. It serves as a bridge between AI tools (like GitHub Copilot, Claude, and other MCP-compatible AI assistants) and AKS, translating natural language requests into AKS operations and returning the results in a format the AI tools can understand.

It allows AI tools to:

- Operate (CRUD) AKS resources
- Retrieve details related to AKS clusters (VNets, Subnets, NSGs, Route Tables, etc.)
- Manage Azure Fleet operations for multi-cluster scenarios

## How it works

AKS-MCP connects to Azure using the Azure SDK and provides a set of tools that AI assistants can use to interact with AKS resources. It leverages the Model Context Protocol (MCP) to facilitate this communication, enabling AI tools to make API calls to Azure and interpret the responses.

## How to install

### Local

<details>
<summary>Install prerequisites</summary>

1. Set up [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) and authenticate
```bash
az login
```
</details>

<br/>

Configure your MCP servers in supported AI clients like [GitHub Copilot](https://github.com/features/copilot), [Claude](https://claude.ai/), or other MCP-compatible clients:

```json
{
  "mcpServers": {
    "aks": {
      "command": "<path of binary aks-mcp>",
      "args": [
        "--transport", "stdio"
      ]
    }
  }
}
```

### GitHub Copilot Configuration in VS Code

For GitHub Copilot in VS Code, configure the MCP server in your `.vscode/mcp.json` file:

```json
{
  "servers": {
    "aks-mcp-server": {
      "type": "stdio",
      "command": "<path of binary aks-mcp>",
      "args": [
        "--transport", "stdio"
      ]
    }
  }
}
```

### Options

Command line arguments:

```sh
Usage of ./aks-mcp:
      --access-level string       Access level (readonly, readwrite, admin) (default "readonly")
      --additional-tools string   Comma-separated list of additional Kubernetes tools to support (kubectl is always enabled). Available: helm,cilium,inspektor-gadget
      --allow-namespaces string   Comma-separated list of allowed Kubernetes namespaces (empty means all namespaces)
      --host string               Host to listen for the server (only used with transport sse or streamable-http) (default "127.0.0.1")
      --port int                  Port to listen for the server (only used with transport sse or streamable-http) (default 8000)
      --timeout int               Timeout for command execution in seconds, default is 600s (default 600)
      --transport string          Transport mechanism to use (stdio, sse or streamable-http) (default "stdio")
```

**Environment variables:**
- Standard Azure authentication environment variables are supported (`AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SUBSCRIPTION_ID`)

## Development

### Building from Source

This project includes a Makefile for convenient development, building, and testing. To see all available targets:

```bash
make help
```

#### Quick Start

```bash
# Build the binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format and lint code
make check

# Build for all platforms
make release
```

#### Common Development Tasks

```bash
# Install dependencies
make deps

# Build and run with --help
make run

# Clean build artifacts
make clean

# Install binary to GOBIN
make install
```

#### Docker

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run
```

### Manual Build

If you prefer to build without the Makefile:

```bash
go build -o aks-mcp ./cmd/aks-mcp
```

## Usage

Ask any questions about your AKS clusters in your AI client, for example:

```
List all my AKS clusters in my subscription xxx.

What is the network configuration of my AKS cluster?

Show me the network security groups associated with my cluster.

Create a new Azure Fleet named prod-fleet in eastus region.

List all members in my fleet.

Create a placement to deploy nginx workloads to clusters with app=frontend label.

Show me all ClusterResourcePlacements in my fleet.
```

## Available Tools

The AKS-MCP server provides consolidated tools for interacting with AKS clusters. These tools have been designed to provide comprehensive functionality through unified interfaces:

<details>
<summary>AKS Cluster Management</summary>

**Tool:** `az_aks_operations`

Unified tool for managing Azure Kubernetes Service (AKS) clusters and related operations.

**Available Operations:**
- **Read-Only** (all access levels):
  - `show`: Show cluster details
  - `list`: List clusters in subscription/resource group
  - `get-versions`: Get available Kubernetes versions
  - `check-network`: Perform outbound network connectivity check
  - `nodepool-list`: List node pools in cluster
  - `nodepool-show`: Show node pool details
  - `account-list`: List Azure subscriptions

- **Read-Write** (`readwrite`/`admin` access levels):
  - `create`: Create new cluster
  - `delete`: Delete cluster
  - `scale`: Scale cluster node count
  - `update`: Update cluster configuration
  - `upgrade`: Upgrade Kubernetes version
  - `nodepool-add`: Add node pool to cluster
  - `nodepool-delete`: Delete node pool
  - `nodepool-scale`: Scale node pool
  - `nodepool-upgrade`: Upgrade node pool
  - `account-set`: Set active subscription
  - `login`: Azure authentication

- **Admin-Only** (`admin` access level):
  - `get-credentials`: Get cluster credentials for kubectl access

</details>

<details>
<summary>Network Resource Management</summary>

**Tool:** `az_network_resources`

Unified tool for getting Azure network resource information used by AKS clusters.

**Available Resource Types:**
- `all`: Get information about all network resources
- `vnet`: Virtual Network information
- `subnet`: Subnet information  
- `nsg`: Network Security Group information
- `route_table`: Route Table information
- `load_balancer`: Load Balancer information
- `private_endpoint`: Private endpoint information

</details>

<details>
<summary>Monitoring and Diagnostics</summary>

**Tool:** `az_monitoring`

Unified tool for Azure monitoring and diagnostics operations for AKS clusters.

**Available Operations:**
- `metrics`: List metric values for resources
- `resource_health`: Retrieve resource health events for AKS clusters
- `app_insights`: Execute KQL queries against Application Insights telemetry data
- `diagnostics`: Check if AKS cluster has diagnostic settings configured
- `control_plane_logs`: Query AKS control plane logs with safety constraints and time range validation

</details>

<details>
<summary>Compute Resources</summary>

**Tool:** `get_aks_vmss_info`
- Get detailed VMSS configuration for node pools in the AKS cluster

**Tool:** `az_vmss_run-command_invoke` *(readwrite/admin only)*
- Execute commands on Virtual Machine Scale Set instances

</details>

<details>
<summary>Fleet Management</summary>

**Tool:** `az_fleet`

Comprehensive Azure Fleet management for multi-cluster scenarios.

**Available Operations:**
- **Fleet Operations**: list, show, create, update, delete, get-credentials
- **Member Operations**: list, show, create, update, delete
- **Update Run Operations**: list, show, create, start, stop, delete
- **Update Strategy Operations**: list, show, create, delete
- **ClusterResourcePlacement Operations**: list, show, get, create, delete

Supports both Azure Fleet management and Kubernetes ClusterResourcePlacement CRD operations.

</details>

<details>
<summary>Diagnostic Detectors</summary>

**Tool:** `list_detectors`
- List all available AKS cluster detectors

**Tool:** `run_detector`
- Run a specific AKS diagnostic detector

**Tool:** `run_detectors_by_category`
- Run all detectors in a specific category
- **Categories**: Best Practices, Cluster and Control Plane Availability and Performance, Connectivity Issues, Create/Upgrade/Delete and Scale, Deprecations, Identity and Security, Node Health, Storage

</details>

<details>
<summary>Azure Advisor</summary>

**Tool:** `az_advisor_recommendation`

Retrieve and manage Azure Advisor recommendations for AKS clusters.

**Available Operations:**
- `list`: List recommendations with filtering options
- `report`: Generate recommendation reports
- **Filter Options**: resource_group, cluster_names, category (Cost, HighAvailability, Performance, Security), severity (High, Medium, Low)

</details>

<details>
<summary>Kubernetes Tools</summary>

*Note: kubectl commands are available with all access levels. Additional tools require explicit enablement via `--additional-tools`*

**kubectl Commands (Read-Only):**
- `kubectl_get`, `kubectl_describe`, `kubectl_explain`, `kubectl_logs`
- `kubectl_api-resources`, `kubectl_api-versions`, `kubectl_diff`
- `kubectl_cluster-info`, `kubectl_top`, `kubectl_events`, `kubectl_auth`

**kubectl Commands (Read-Write/Admin):**
- `kubectl_create`, `kubectl_delete`, `kubectl_apply`, `kubectl_expose`, `kubectl_run`
- `kubectl_set`, `kubectl_rollout`, `kubectl_scale`, `kubectl_autoscale`
- `kubectl_label`, `kubectl_annotate`, `kubectl_patch`, `kubectl_replace`
- `kubectl_cp`, `kubectl_exec`, `kubectl_cordon`, `kubectl_uncordon`
- `kubectl_drain`, `kubectl_taint`, `kubectl_certificate`

**Additional Tools (Optional):**
- `helm`: Helm package manager (requires `--additional-tools helm`)
- `cilium`: Cilium CLI for eBPF networking (requires `--additional-tools cilium`)

</details>

<details>
<summary>Real-time Observability</summary>

**Tool:** `inspektor_gadget` *(requires `--additional-tools inspektor-gadget`)*

Real-time observability tool for Azure Kubernetes Service (AKS) clusters using eBPF.

**Available Actions:**
- `deploy`: Deploy Inspektor Gadget to cluster
- `undeploy`: Remove Inspektor Gadget from cluster
- `is_deployed`: Check deployment status
- `run`: Run one-shot gadgets
- `start`: Start continuous gadgets
- `stop`: Stop running gadgets
- `get_results`: Retrieve gadget results
- `list_gadgets`: List available gadgets

**Available Gadgets:**
- `observe_dns`: Monitor DNS requests and responses
- `observe_tcp`: Monitor TCP connections
- `observe_file_open`: Monitor file system operations
- `observe_process_execution`: Monitor process execution
- `observe_signal`: Monitor signal delivery
- `observe_system_calls`: Monitor system calls
- `top_file`: Top files by I/O operations
- `top_tcp`: Top TCP connections by traffic

</details>

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft
trademarks or logos is subject to and must follow
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.