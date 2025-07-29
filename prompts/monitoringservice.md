Feature: Add MCP (Monitoring Control Plane) server
// Description:
// Implement a service (MCP server) that connects to any attached monitoring services for a given AKS cluster.
// The server should expose APIs or a UI to perform the following tasks:
//
// 1. Read and query Azure Log Analytics workspaces linked to the cluster:
// - Retrieve control plane logs
// - Query audit logs
// - Fetch historical logs for nodes and pods
//
// 2. Read and visualize metrics from Managed Prometheus (AMP):
// - Access Prometheus scrape endpoint via Azure Monitor
// - Display basic dashboard visualizations (e.g., CPU, memory, network)
//
// 3. Access and query Application Insights:
// - Read distributed trace data
// - Enable filtering by operation name, request ID, service name, etc.
//
// Requirements:
// - Use Azure SDKs (Go or Python preferred)
// - Support authentication via kubeconfig or managed identity
// - Implement minimal RESTful API to trigger each of the above

//
// Goal:
// Provide a dev-friendly mcp implementation for AKS clusters with access to log/metric/trace data via attached services.