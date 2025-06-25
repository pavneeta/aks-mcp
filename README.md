# AKS-MCP

The AKS-MCP is a Model Context Protocol (MCP) server that enables AI assistants to interact with Azure Kubernetes Service (AKS) clusters. It serves as a bridge between AI tools (like GitHub Copilot, Claude, and other MCP-compatible AI assistants) and AKS, translating natural language requests into AKS operations and returning the results in a format the AI tools can understand.

It allows AI tools to:

- Operate (CRUD) AKS resources
- Retrieve details related to AKS clusters (VNets, Subnets, NSGs, Route Tables, etc.)

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

You can also specify a specific AKS cluster to work with:

```json
{
  "mcpServers": {
    "aks": {
      "command": "<path of binary aks-mcp>",
      "args": [
        "--transport", "stdio",
        "--resource-id", "/subscriptions/your-subscription-id/resourceGroups/your-resource-group/providers/Microsoft.ContainerService/managedClusters/your-cluster-name"
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
        "--aks-resource-id",
        "/subscriptions/your-subscription-id/resourceGroups/your-resource-group/providers/Microsoft.ContainerService/managedClusters/your-cluster-name"
      ]
    }
  }
}
```

### Options

Command line arguments:

```sh
Usage of ./aks-mcp:
  --access-level string      Access level for tools (read, readwrite, admin) (default "read")
  --address string           Address to listen on when using transport SSE or streamable-http (default "localhost:8080")
  --aks-resource-id string   AKS Resource ID (optional), set this when using single cluster mode
  -t, --transport string     Transport type (stdio, sse or streamable-http) (default "stdio")
```

**Environment variables:**
- Standard Azure authentication environment variables are supported (`AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SUBSCRIPTION_ID`)

## Usage

Ask any questions about your AKS clusters in your AI client, for example:

```
List all my AKS clusters in my subscription xxx.

What is the network configuration of my AKS cluster?

Show me the network security groups associated with my cluster.
```

## Available Tools

The AKS-MCP server provides the following tools for interacting with AKS clusters:

<details>
<summary>Cluster Tools</summary>

- `get_cluster_info`: Get detailed information about an AKS cluster
- `list_aks_clusters`: List all AKS clusters in a subscription and optional resource group
</details>

<details>
<summary>Network Tools</summary>

- `get_vnet_info`: Get information about the VNet used by the AKS cluster
- `get_subnet_info`: Get information about the subnets used by the AKS cluster
- `get_route_table_info`: Get information about the route tables used by the AKS cluster
- `get_nsg_info`: Get information about the network security groups used by the AKS cluster
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
