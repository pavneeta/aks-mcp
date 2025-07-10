package monitor

import (
	"github.com/Azure/aks-mcp/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

// MonitorCommand defines a specific az monitor command to be registered as a tool
type MonitorCommand struct {
	Name        string
	Description string
	ArgsExample string // Example of command arguments
	Category    string // Category for the command (e.g., "metrics", "logs")
}

// RegisterMonitorCommand registers a specific az monitor command as an MCP tool
func RegisterMonitorCommand(cmd MonitorCommand) mcp.Tool {
	// Convert spaces to underscores for valid tool name
	commandName := cmd.Name
	validToolName := utils.ReplaceSpacesWithUnderscores(commandName)

	description := "Run " + cmd.Name + " command: " + cmd.Description + "."

	// Add example if available, with proper punctuation
	if cmd.ArgsExample != "" {
		description += "\nExample: `" + cmd.ArgsExample + "`"
	}

	return mcp.NewTool(validToolName,
		mcp.WithDescription(description),
		mcp.WithString("args",
			mcp.Required(),
			mcp.Description("Arguments for the `"+cmd.Name+"` command"),
		),
	)
}

// GetReadOnlyMonitorCommands returns all read-only az monitor commands
func GetReadOnlyMonitorCommands() []MonitorCommand {
	return []MonitorCommand{
		// Metrics commands
		{
			Name:        "az monitor metrics list",
			Description: "List the metric values for a resource",
			ArgsExample: "--resource /subscriptions/{subscription}/resourceGroups/{resourceGroup}/providers/Microsoft.Compute/virtualMachines/{vmName} --metric \"Percentage CPU\"",
			Category:    "metrics",
		},
		{
			Name:        "az monitor metrics list-definitions",
			Description: "List the metric definitions for a resource",
			ArgsExample: "--resource /subscriptions/{subscription}/resourceGroups/{resourceGroup}/providers/Microsoft.ContainerService/managedClusters/{clusterName}",
			Category:    "metrics",
		},
		{
			Name:        "az monitor metrics list-namespaces",
			Description: "List the metric namespaces for a resource",
			ArgsExample: "--resource /subscriptions/{subscription}/resourceGroups/{resourceGroup}/providers/Microsoft.ContainerService/managedClusters/{clusterName}",
			Category:    "metrics",
		},
	}
}

// GetReadWriteMonitorCommands returns all read-write az monitor commands
func GetReadWriteMonitorCommands() []MonitorCommand {
	return []MonitorCommand{}
}

// GetAdminMonitorCommands returns all admin az monitor commands
func GetAdminMonitorCommands() []MonitorCommand {
	return []MonitorCommand{}
}

// RegisterResourceHealthTool registers the Azure Resource Health monitoring tool
func RegisterResourceHealthTool() mcp.Tool {
	return mcp.NewTool("az_monitor_activity_log_resource_health",
		mcp.WithDescription("Retrieve resource health events for AKS clusters to monitor service availability and health status"),
		mcp.WithString("subscription_id",
			mcp.Required(),
			mcp.Description("Azure subscription ID"),
		),
		mcp.WithString("resource_group",
			mcp.Required(),
			mcp.Description("Resource group name containing the AKS cluster"),
		),
		mcp.WithString("cluster_name",
			mcp.Required(),
			mcp.Description("AKS cluster name"),
		),
		mcp.WithString("start_time",
			mcp.Required(),
			mcp.Description("Start date for health event query (ISO 8601 format, e.g., \"2025-01-01T00:00:00Z\")"),
		),
		mcp.WithString("end_time",
			mcp.Description("End date for health event query (defaults to current time)"),
		),
		mcp.WithString("status",
			mcp.Description("Filter by health status (Available, Unavailable, Degraded, Unknown)"),
		),
	)
}
