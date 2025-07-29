package inspektorgadget

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/Azure/aks-mcp/internal/command"
	"github.com/Azure/aks-mcp/internal/config"
	"github.com/Azure/aks-mcp/internal/tools"
)

// =============================================================================
// Inspektor Gadget Handler
// =============================================================================

var ErrNotDeployed = fmt.Errorf("inspektor gadget is not deployed, please deploy it first e.g. using 'inspektor_gadget' (action: deploy) tool (requires 'readwrite' or 'admin' access level)")

// InspektorGadgetHandler returns a handler to manage gadgets
func InspektorGadgetHandler(mgr GadgetManager, cfg *config.ConfigData) tools.ResourceHandler {
	return tools.ResourceHandlerFunc(func(params map[string]interface{}, _ *config.ConfigData) (string, error) {
		ctx := context.Background()

		// Validate action parameter
		action, ok := params["action"].(string)
		if !ok || action == "" {
			return "", fmt.Errorf("missing 'action' parameter, must be a non-empty string")
		}
		if !isValidAction(action) {
			validActions := getActions()
			return "", fmt.Errorf("invalid action: %s, expected one of %v", action, validActions)
		}

		// Check if Inspektor Gadget is deployed
		deployed, _, err := mgr.IsDeployed(ctx)
		if err != nil {
			return "", fmt.Errorf("checking Inspektor Gadget deployment: %w", err)
		}
		if !deployed && !slices.Contains(getLifecycleActions(), action) {
			return "", ErrNotDeployed
		}

		// Initialize action/filter parameters if not provided
		actionParams, ok := params["action_params"].(map[string]interface{})
		if !ok {
			actionParams = map[string]interface{}{}
		}
		filterParams, ok := params["filter_params"].(map[string]interface{})
		if !ok {
			filterParams = map[string]interface{}{}
		}

		// validate filter parameters
		for k := range filterParams {
			if !isValidFilterParamKey(k) {
				return "", fmt.Errorf("invalid filter parameter key: %s, must be one of %v", k, getFilterParamKeys())
			}
		}

		switch action {
		case runAction:
			return handleRunAction(ctx, mgr, actionParams, filterParams, cfg)
		case startAction:
			return handleStartAction(ctx, mgr, actionParams, filterParams, cfg)
		case stopAction:
			return handleStopAction(ctx, mgr, actionParams, cfg)
		case getResultsAction:
			return handleGetResultsAction(ctx, mgr, actionParams, cfg)
		case listGadgetsAction:
			return handleListGadgetsAction(ctx, mgr, cfg)
		case isDeployedAction, undeployAction, deployAction:
			return handleLifecycleAction(deployed, action, actionParams, cfg)
		}

		return "", fmt.Errorf("unsupported action: %s", action)
	})
}

func handleRunAction(ctx context.Context, mgr GadgetManager, actionParams map[string]interface{}, filterParams map[string]interface{}, cfg *config.ConfigData) (string, error) {
	gadgetName, ok := actionParams["gadget_name"].(string)
	if !ok || gadgetName == "" {
		return "", fmt.Errorf("invalid or missing 'gadget_name' parameter in 'run' action, must be a non-empty string")
	}

	gadget, ok := getGadgetByName(gadgetName)
	if !ok {
		return "", fmt.Errorf("invalid or unsupported gadget name: %s", gadgetName)
	}

	duration, ok := actionParams["duration"].(float64)
	if !ok || duration <= 0 {
		duration = 10
	}

	// TODO: Use GetGadgetInfo to validate gadgetParams to ensure compatibility with different gadget versions
	gadgetParams, err := prepareCommonParams(filterParams, cfg)
	if err != nil {
		return "", fmt.Errorf("preparing common parameters: %w", err)
	}
	if gadget.ParamsFunc != nil {
		gadget.ParamsFunc(filterParams, gadgetParams)
	}
	// set map-fetch-interval to half of the timeout to limit the volume of data fetched
	dur := time.Duration(duration) * time.Second
	gadgetParams[paramFetchInterval] = (dur / 2).String()

	resp, err := mgr.RunGadget(ctx, gadget.Image, gadgetParams, dur)
	if err != nil {
		return "", fmt.Errorf("running gadget: %w", err)
	}
	return resp, nil
}

func handleStartAction(ctx context.Context, mgr GadgetManager, actionParams map[string]interface{}, filterParams map[string]interface{}, cfg *config.ConfigData) (string, error) {
	gadgetName, ok := actionParams["gadget_name"].(string)
	if !ok || gadgetName == "" {
		return "", fmt.Errorf("invalid or missing 'gadget_name' parameter in 'start' action, must be a non-empty string")
	}

	gadget, ok := getGadgetByName(gadgetName)
	if !ok {
		return "", fmt.Errorf("invalid or unsupported gadget name: %s", gadgetName)
	}

	// TODO: Use GetGadgetInfo to validate gadgetParams to ensure compatibility with different gadget versions
	gadgetParams, err := prepareCommonParams(filterParams, cfg)
	if err != nil {
		return "", fmt.Errorf("preparing common parameters: %w", err)
	}
	if gadget.ParamsFunc != nil {
		gadget.ParamsFunc(filterParams, gadgetParams)
	}

	var filterParamsStr string
	for k, v := range filterParams {
		if filterParamsStr != "" {
			filterParamsStr += ","
		}
		filterParamsStr += fmt.Sprintf("%s=%v", k, v)
	}
	tags := []string{
		"gadgetName=" + gadgetName,
		"filterParams=" + filterParamsStr,
		"namespaces=" + getNamespace(gadgetParams),
	}
	id, err := mgr.StartGadget(ctx, gadget.Image, gadgetParams, tags)
	if err != nil {
		return "", fmt.Errorf("starting gadget: %w", err)
	}
	return fmt.Sprintf("Gadget started with ID: %s", id), nil
}

func handleStopAction(ctx context.Context, mgr GadgetManager, actionParams map[string]interface{}, cfg *config.ConfigData) (string, error) {
	id, ok := actionParams["gadget_id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("invalid or missing 'gadget_id' parameter in 'stop' action, must be a non-empty string")
	}

	if err := validateGadgetAccess(ctx, mgr, id, cfg); err != nil {
		return "", err
	}

	err := mgr.StopGadget(ctx, id)
	if err != nil {
		return "", fmt.Errorf("stopping gadget: %w", err)
	}
	return fmt.Sprintf("Gadget with ID %s stopped successfully", id), nil
}

func handleGetResultsAction(ctx context.Context, mgr GadgetManager, actionParams map[string]interface{}, cfg *config.ConfigData) (string, error) {
	id, ok := actionParams["gadget_id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("invalid or missing 'gadget_id' parameter in 'get_results' action, must be a non-empty string")
	}

	if err := validateGadgetAccess(ctx, mgr, id, cfg); err != nil {
		return "", err
	}

	results, err := mgr.GetResults(ctx, id)
	if err != nil {
		return "", fmt.Errorf("getting gadget results: %w", err)
	}
	return results, nil
}

func handleListGadgetsAction(ctx context.Context, mgr GadgetManager, cfg *config.ConfigData) (string, error) {
	gs, err := mgr.ListGadgets(ctx)
	if err != nil {
		return "", fmt.Errorf("listing gadgets: %w", err)
	}

	// Filter gadgets based on namespace access
	if cfg.SecurityConfig != nil {
		filteredGadgets := make([]*GadgetInstance, 0)
		for _, gadget := range gs {
			if isGadgetAccessAllowed(gadget, cfg) {
				filteredGadgets = append(filteredGadgets, gadget)
			}
		}
		gs = filteredGadgets
	}

	if len(gs) == 0 {
		return "No gadgets are currently running", nil
	}
	JSONData, err := json.Marshal(gs)
	if err != nil {
		return "", fmt.Errorf("marshalling gadget list to JSON: %w", err)
	}
	return string(JSONData), nil
}

func handleLifecycleAction(deployed bool, action string, actionParams map[string]interface{}, cfg *config.ConfigData) (string, error) {
	// TODO: use security.Validator once helm readwrite/admin operations are implemented
	if !cfg.SecurityConfig.IsNamespaceAllowed(inspektorGadgetChartNamespace) {
		return "", fmt.Errorf("namespace %s is not allowed by security policy", inspektorGadgetChartNamespace)
	}
	if (cfg.AccessLevel != "readwrite" && cfg.AccessLevel != "admin") && (!slices.Contains(getReadonlyLifecycleActions(), action)) {
		return "", fmt.Errorf("action %q requires 'readwrite' or 'admin' access level, current access level is '%s'", action, cfg.AccessLevel)
	}

	switch action {
	case isDeployedAction:
		if deployed {
			return "inspektor gadget is deployed", nil
		}
		return "inspektor gadget is not deployed", nil
	case undeployAction:
		if !deployed {
			return "inspektor gadget is not deployed", nil
		}
		return handleUndeployAction(cfg)
	case deployAction:
		if deployed {
			return "inspektor gadget is already deployed", nil
		}
		return handleDeployAction(actionParams, cfg)
	}

	return "", fmt.Errorf("unsupported lifecycle action %q, must be one of %v", action, getLifecycleActions())
}

func handleDeployAction(actionParams map[string]interface{}, cfg *config.ConfigData) (string, error) {
	chartVersion, ok := actionParams["chart_version"].(string)
	if !ok || chartVersion == "" {
		chartVersion = getChartVersion()
	}
	chartUrl := fmt.Sprintf("%s:%s", inspektorGadgetChartURL, chartVersion)
	helmArgs := fmt.Sprintf("install %s -n %s --create-namespace %s", inspektorGadgetChartRelease, inspektorGadgetChartNamespace, chartUrl)
	process := command.NewShellProcess("helm", cfg.Timeout)
	return process.Run(helmArgs)
}

func handleUndeployAction(cfg *config.ConfigData) (string, error) {
	helmArgs := fmt.Sprintf("uninstall %s -n %s", inspektorGadgetChartRelease, inspektorGadgetChartNamespace)
	process := command.NewShellProcess("helm", cfg.Timeout)
	return process.Run(helmArgs)
}

func prepareCommonParams(filterParams map[string]interface{}, cfg *config.ConfigData) (map[string]string, error) {
	// We need to ensure that the security policy allows the namespace
	ns, ok := filterParams["namespace"].(string)
	if ok && ns != "" && cfg.SecurityConfig != nil && !cfg.SecurityConfig.IsNamespaceAllowed(ns) {
		return nil, fmt.Errorf("namespace %q is not allowed by security policy", ns)
	}

	// If the namespace is provided, use it; otherwise, use the allowed namespaces or all namespaces
	// depending on the security policy
	params := make(map[string]string)
	if ns != "" {
		params[paramNamespace] = ns
	} else if cfg.SecurityConfig != nil && cfg.SecurityConfig.AllowedNamespaces != "" {
		params[paramNamespace] = cfg.SecurityConfig.AllowedNamespaces
	} else {
		params[paramAllNamespaces] = "true"
	}

	if pod, ok := filterParams["pod"].(string); ok && pod != "" {
		params[paramPod] = pod
	}

	if container, ok := filterParams["container"].(string); ok && container != "" {
		params[paramContainer] = container
	}

	if selector, ok := filterParams["selector"].(string); ok && selector != "" {
		params[paramSelector] = selector
	}

	return params, nil
}

func getNamespace(gadgetParams map[string]string) string {
	if ns, ok := gadgetParams[paramNamespace]; ok && ns != "" {
		return ns
	}
	if allNs, ok := gadgetParams[paramAllNamespaces]; ok && allNs == "true" {
		return ""
	}
	return ""
}

// validateGadgetAccess checks if the user has access to a specific gadget based on namespace restrictions
func validateGadgetAccess(ctx context.Context, mgr GadgetManager, gadgetID string, cfg *config.ConfigData) error {
	if cfg.SecurityConfig == nil {
		return nil // No security restrictions
	}

	gadgetInstances, err := mgr.ListGadgets(ctx)
	if err != nil {
		return fmt.Errorf("failed to list gadgets for validation: %w", err)
	}

	for _, gadget := range gadgetInstances {
		if gadget.ID == gadgetID {
			if !isGadgetAccessAllowed(gadget, cfg) {
				return fmt.Errorf("access denied: gadget %s is running in namespaces not allowed by security policy", gadgetID)
			}
			return nil
		}
	}

	return fmt.Errorf("gadget with ID %s not found", gadgetID)
}

// isGadgetAccessAllowed checks if a gadget is accessible based on namespace restrictions
func isGadgetAccessAllowed(gadget *GadgetInstance, cfg *config.ConfigData) bool {
	if cfg.SecurityConfig == nil {
		return true
	}

	// If the gadget has no namespace restrictions (runs in all namespaces),
	// only allow access if security config allows all namespaces
	if len(gadget.Namespaces) == 0 {
		return cfg.SecurityConfig.AllowedNamespaces == ""
	}

	// Check if all gadget namespaces are allowed
	for _, ns := range gadget.Namespaces {
		if ns != "" && !cfg.SecurityConfig.IsNamespaceAllowed(ns) {
			return false
		}
	}

	return true
}
