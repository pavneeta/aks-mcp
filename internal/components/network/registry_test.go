package network

import (
	"testing"

	"github.com/Azure/aks-mcp/internal/azureclient"
	"github.com/Azure/aks-mcp/internal/config"
)

func TestRegisterVNetInfoTool(t *testing.T) {
	tool := RegisterVNetInfoTool()

	if tool.Name != "get_vnet_info" {
		t.Errorf("Expected tool name 'get_vnet_info', got %s", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected tool description to be set")
	}
}

func TestRegisterNSGInfoTool(t *testing.T) {
	tool := RegisterNSGInfoTool()

	if tool.Name != "get_nsg_info" {
		t.Errorf("Expected tool name 'get_nsg_info', got %s", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected tool description to be set")
	}
}

func TestRegisterRouteTableInfoTool(t *testing.T) {
	tool := RegisterRouteTableInfoTool()

	if tool.Name != "get_route_table_info" {
		t.Errorf("Expected tool name 'get_route_table_info', got %s", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected tool description to be set")
	}
}

func TestRegisterSubnetInfoTool(t *testing.T) {
	tool := RegisterSubnetInfoTool()

	if tool.Name != "get_subnet_info" {
		t.Errorf("Expected tool name 'get_subnet_info', got %s", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Expected tool description to be set")
	}
}

func TestGetVNetInfoHandler(t *testing.T) {
	mockClient := &azureclient.AzureClient{}
	cfg := &config.ConfigData{}

	handler := GetVNetInfoHandler(mockClient, cfg)

	if handler == nil {
		t.Error("Expected handler to be non-nil")
	}
}

func TestGetNSGInfoHandler(t *testing.T) {
	mockClient := &azureclient.AzureClient{}
	cfg := &config.ConfigData{}

	handler := GetNSGInfoHandler(mockClient, cfg)

	if handler == nil {
		t.Error("Expected handler to be non-nil")
	}
}

func TestGetRouteTableInfoHandler(t *testing.T) {
	mockClient := &azureclient.AzureClient{}
	cfg := &config.ConfigData{}

	handler := GetRouteTableInfoHandler(mockClient, cfg)

	if handler == nil {
		t.Error("Expected handler to be non-nil")
	}
}

func TestGetSubnetInfoHandler(t *testing.T) {
	mockClient := &azureclient.AzureClient{}
	cfg := &config.ConfigData{}

	handler := GetSubnetInfoHandler(mockClient, cfg)

	if handler == nil {
		t.Error("Expected handler to be non-nil")
	}
}
