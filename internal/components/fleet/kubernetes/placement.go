package kubernetes

import (
	"fmt"
	"os"
	"strings"

	"github.com/Azure/aks-mcp/internal/config"
)

// PlacementOperations handles Fleet placement CRD operations
type PlacementOperations struct {
	client *Client
}

// NewPlacementOperations creates a new placement operations handler
func NewPlacementOperations(client *Client) *PlacementOperations {
	if client == nil {
		panic("placement operations client cannot be nil")
	}
	return &PlacementOperations{
		client: client,
	}
}

// CreatePlacement creates a new ClusterResourcePlacement using kubectl
func (p *PlacementOperations) CreatePlacement(name, selector, policy string, cfg *config.ConfigData) (string, error) {
	// Create YAML manifest for ClusterResourcePlacement  
	var resourceSelectors string
	
	if selector != "" {
		// Parse selector: assume format like "app=nginx,env=prod"
		resourceSelectors = `
  resourceSelectors:
  - group: ""
    version: "v1"
    kind: "Namespace"
    labelSelector:
      matchLabels:`
		
		pairs := strings.Split(selector, ",")
		for _, pair := range pairs {
			parts := strings.Split(pair, "=")
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				resourceSelectors += fmt.Sprintf("\n        %s: \"%s\"", key, value)
			}
		}
	} else {
		// Default: select namespaces with a fleet label
		resourceSelectors = `
  resourceSelectors:
  - group: ""
    version: "v1"
    kind: "Namespace"
    labelSelector:
      matchLabels:
        fleet.azure.com/name: "default"`
	}

	manifest := fmt.Sprintf(`apiVersion: placement.kubernetes-fleet.io/v1beta1
kind: ClusterResourcePlacement
metadata:
  name: %s
spec:%s
  policy:
    placementType: %s`, name, resourceSelectors, policy)

	// Create a temporary file with the manifest
	tempFile, err := os.CreateTemp("", "placement-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFileName := tempFile.Name()
	defer os.Remove(tempFileName)

	// Write the manifest to the temp file
	if _, err := tempFile.WriteString(manifest); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("failed to write manifest to temp file: %w", err)
	}
	
	// Debug: print the generated manifest
	fmt.Printf("Generated YAML manifest:\n%s\n", manifest)
	
	// Close the file before using it
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Apply the manifest from the temp file
	command := fmt.Sprintf("apply -f %s", tempFileName)
	
	result, err := p.client.ExecuteKubectl(command, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create placement: %w", err)
	}

	return result, nil
}

// GetPlacement retrieves a ClusterResourcePlacement by name using kubectl
func (p *PlacementOperations) GetPlacement(name string, cfg *config.ConfigData) (string, error) {
	command := fmt.Sprintf("get clusterresourceplacement %s -o yaml", name)
	
	result, err := p.client.ExecuteKubectl(command, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get placement '%s': %w", name, err)
	}

	return result, nil
}

// ListPlacements lists all ClusterResourcePlacements using kubectl
func (p *PlacementOperations) ListPlacements(cfg *config.ConfigData) (string, error) {
	if p == nil {
		return "", fmt.Errorf("PlacementOperations is nil")
	}
	if p.client == nil {
		return "", fmt.Errorf("placement client is nil")
	}
	
	command := "get clusterresourceplacement -o wide"
	
	result, err := p.client.ExecuteKubectl(command, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to list placements: %w", err)
	}

	return result, nil
}

// DeletePlacement deletes a ClusterResourcePlacement by name using kubectl
func (p *PlacementOperations) DeletePlacement(name string, cfg *config.ConfigData) (string, error) {
	command := fmt.Sprintf("delete clusterresourceplacement %s", name)
	
	result, err := p.client.ExecuteKubectl(command, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to delete placement '%s': %w", name, err)
	}

	return result, nil
}

// ParsePlacementArgs parses command arguments for placement operations
func ParsePlacementArgs(args string) (map[string]string, error) {
	result := make(map[string]string)
	
	// Simple argument parser for --key value format
	parts := strings.Fields(args)
	for i := 0; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], "--") {
			key := strings.TrimPrefix(parts[i], "--")
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "--") {
				result[key] = parts[i+1]
				i++ // Skip the value
			}
		}
	}
	
	return result, nil
}