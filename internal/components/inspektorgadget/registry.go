package inspektorgadget

import "github.com/mark3labs/mcp-go/mcp"

// =============================================================================
// Inspektor Gadget related Tool Registrations
// =============================================================================

// RegisterInspektorGadgetTool registers the inspektor-gadget tool to manage gadgets
func RegisterInspektorGadgetTool() mcp.Tool {
	return mcp.NewTool(
		"inspektor_gadget",
		mcp.WithDescription("Real-time observability tool for Azure Kubernetes Service (AKS) clusters, allowing users to manage gadgets for monitoring and debugging"),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("Action to perform on the gadget: "+
				runAction+" to run a gadget for a specific duration, "+
				startAction+" to start a gadget for continuous observation, "+
				stopAction+" to stop a running gadget, "+
				getResultsAction+" to retrieve results of a gadget run (only available before stopping the gadget), "+
				listGadgetsAction+" to list all running gadgets"+
				deployAction+" to deploy Inspektor Gadget, "+
				undeployAction+" to undeploy Inspektor Gadget"+
				isDeployedAction+" to check if Inspektor Gadget is deployed",
			),
			mcp.Enum(getActions()...),
		),
		mcp.WithObject("action_params",
			mcp.Description("Parameters for the action"),
			mcp.Required(),
			mcp.Properties(map[string]any{
				"gadget_name": map[string]any{
					"type":        "string",
					"description": "Name of the gadget to run/start",
					"enum":        getGadgetNames(),
				},
				"duration": map[string]any{
					"type":        "number",
					"description": "Duration in seconds to run the gadget",
					"default":     10,
				},
				"gadget_id": map[string]any{
					"type":        "string",
					"description": "ID of the gadget run to stop or get results for. This ID is returned when starting a gadget for continuous observation.",
				},
				"chart_version": map[string]any{
					"type":        "string",
					"description": "The version of the Inspektor Gadget Helm chart to deploy. Only set this if user explicitly wants to deploy a specific version",
				},
			}),
		),
		mcp.WithObject("filter_params",
			mcp.Description("Parameters to filter the data captured by the gadget"),
			mcp.Properties(
				mergeMaps(
					map[string]any{
						"namespace": map[string]any{
							"type":        "string",
							"description": "Kubernetes namespace",
						},
						"pod": map[string]any{
							"type":        "string",
							"description": "Kubernetes pod name",
						},
						"container": map[string]any{
							"type":        "string",
							"description": "Kubernetes container name",
						},
						"selector": map[string]any{
							"type":        "string",
							"description": "Label selector to filter pods by labels (e.g. key1=value1,key2=value2)",
						},
					},
					getGadgetParams(),
				),
			),
		),
	)
}
