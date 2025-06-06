// Package models contains data models for Azure resources used in the AKS MCP server.
package models

// AKSClusterSummary represents essential information about an AKS cluster.
// It provides a lightweight representation of the cluster for listing operations.
type AKSClusterSummary struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Location          string `json:"location"`
	ResourceGroup     string `json:"resourceGroup"`
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
	ProvisioningState string `json:"provisioningState,omitempty"`
	AgentPoolCount    int    `json:"agentPoolCount"`
}
