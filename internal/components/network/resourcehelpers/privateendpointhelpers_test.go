package resourcehelpers

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
)

// TestGetPrivateEndpointIDFromAKS tests the private endpoint ID extraction from an AKS cluster
func TestGetPrivateEndpointIDFromAKS(t *testing.T) {
	ctx := context.Background()

	t.Run("private cluster with EnablePrivateCluster true", func(t *testing.T) {
		enablePrivateCluster := true
		nodeResourceGroup := "MC_test-rg_test-cluster_eastus"

		cluster := &armcontainerservice.ManagedCluster{
			Properties: &armcontainerservice.ManagedClusterProperties{
				APIServerAccessProfile: &armcontainerservice.ManagedClusterAPIServerAccessProfile{
					EnablePrivateCluster: &enablePrivateCluster,
				},
				NodeResourceGroup: &nodeResourceGroup,
			},
		}

		// Since we can't mock the Azure client easily, we expect this to fail with a client error
		// but not with a "not a private cluster" error
		privateEndpointID, err := GetPrivateEndpointIDFromAKS(ctx, cluster, nil)

		// Should get an error because client is nil, but not because it's not a private cluster
		if err == nil {
			t.Error("Expected error for nil client")
		}
		if privateEndpointID != "" {
			t.Error("Expected empty private endpoint ID on error")
		}
		if err.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("public cluster with EnablePrivateCluster false", func(t *testing.T) {
		enablePrivateCluster := false
		nodeResourceGroup := "MC_test-rg_test-cluster_eastus"

		cluster := &armcontainerservice.ManagedCluster{
			Properties: &armcontainerservice.ManagedClusterProperties{
				APIServerAccessProfile: &armcontainerservice.ManagedClusterAPIServerAccessProfile{
					EnablePrivateCluster: &enablePrivateCluster,
				},
				NodeResourceGroup: &nodeResourceGroup,
			},
		}

		privateEndpointID, err := GetPrivateEndpointIDFromAKS(ctx, cluster, nil)

		// Should return empty string with no error for public cluster
		if err != nil {
			t.Errorf("Expected no error for public cluster, got %v", err)
		}
		if privateEndpointID != "" {
			t.Error("Expected empty private endpoint ID for public cluster")
		}
	})

	t.Run("cluster with nil APIServerAccessProfile", func(t *testing.T) {
		nodeResourceGroup := "MC_test-rg_test-cluster_eastus"

		cluster := &armcontainerservice.ManagedCluster{
			Properties: &armcontainerservice.ManagedClusterProperties{
				APIServerAccessProfile: nil,
				NodeResourceGroup:      &nodeResourceGroup,
			},
		}

		privateEndpointID, err := GetPrivateEndpointIDFromAKS(ctx, cluster, nil)

		// Should return empty string with no error for cluster without APIServerAccessProfile
		if err != nil {
			t.Errorf("Expected no error for cluster without APIServerAccessProfile, got %v", err)
		}
		if privateEndpointID != "" {
			t.Error("Expected empty private endpoint ID for cluster without APIServerAccessProfile")
		}
	})

	t.Run("cluster with nil EnablePrivateCluster", func(t *testing.T) {
		nodeResourceGroup := "MC_test-rg_test-cluster_eastus"

		cluster := &armcontainerservice.ManagedCluster{
			Properties: &armcontainerservice.ManagedClusterProperties{
				APIServerAccessProfile: &armcontainerservice.ManagedClusterAPIServerAccessProfile{
					EnablePrivateCluster: nil,
				},
				NodeResourceGroup: &nodeResourceGroup,
			},
		}

		privateEndpointID, err := GetPrivateEndpointIDFromAKS(ctx, cluster, nil)

		// Should return empty string with no error for cluster with nil EnablePrivateCluster
		if err != nil {
			t.Errorf("Expected no error for cluster with nil EnablePrivateCluster, got %v", err)
		}
		if privateEndpointID != "" {
			t.Error("Expected empty private endpoint ID for cluster with nil EnablePrivateCluster")
		}
	})

	t.Run("cluster with nil properties", func(t *testing.T) {
		cluster := &armcontainerservice.ManagedCluster{
			Properties: nil,
		}

		_, err := GetPrivateEndpointIDFromAKS(ctx, cluster, nil)
		if err == nil {
			t.Error("Expected error for cluster with nil properties")
		}
		if err.Error() != "invalid cluster or cluster properties" {
			t.Errorf("Expected 'invalid cluster or cluster properties' error, got %v", err)
		}
	})

	t.Run("nil cluster", func(t *testing.T) {
		_, err := GetPrivateEndpointIDFromAKS(ctx, nil, nil)
		if err == nil {
			t.Error("Expected error for nil cluster")
		}
		if err.Error() != "invalid cluster or cluster properties" {
			t.Errorf("Expected 'invalid cluster or cluster properties' error, got %v", err)
		}
	})

	t.Run("private cluster with nil NodeResourceGroup", func(t *testing.T) {
		enablePrivateCluster := true

		cluster := &armcontainerservice.ManagedCluster{
			Properties: &armcontainerservice.ManagedClusterProperties{
				APIServerAccessProfile: &armcontainerservice.ManagedClusterAPIServerAccessProfile{
					EnablePrivateCluster: &enablePrivateCluster,
				},
				NodeResourceGroup: nil,
			},
		}

		_, err := GetPrivateEndpointIDFromAKS(ctx, cluster, nil)
		if err == nil {
			t.Error("Expected error for private cluster with nil NodeResourceGroup")
		}
		if err.Error() != "node resource group not found for AKS cluster" {
			t.Errorf("Expected 'node resource group not found for AKS cluster' error, got %v", err)
		}
	})
}

// TestExtractSubscriptionIDFromCluster tests the subscription ID extraction from a cluster
func TestExtractSubscriptionIDFromCluster(t *testing.T) {
	t.Run("valid cluster ID", func(t *testing.T) {
		clusterID := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg/providers/Microsoft.ContainerService/managedClusters/test-cluster"

		cluster := &armcontainerservice.ManagedCluster{
			ID: &clusterID,
		}

		subscriptionID := extractSubscriptionIDFromCluster(cluster)
		expectedSubscriptionID := "12345678-1234-1234-1234-123456789012"

		if subscriptionID != expectedSubscriptionID {
			t.Errorf("Expected subscription ID %s, got %s", expectedSubscriptionID, subscriptionID)
		}
	})

	t.Run("cluster with nil ID", func(t *testing.T) {
		cluster := &armcontainerservice.ManagedCluster{
			ID: nil,
		}

		subscriptionID := extractSubscriptionIDFromCluster(cluster)
		if subscriptionID != "" {
			t.Errorf("Expected empty subscription ID for cluster with nil ID, got %s", subscriptionID)
		}
	})

	t.Run("cluster with invalid ID format", func(t *testing.T) {
		clusterID := "invalid-cluster-id"

		cluster := &armcontainerservice.ManagedCluster{
			ID: &clusterID,
		}

		subscriptionID := extractSubscriptionIDFromCluster(cluster)
		if subscriptionID != "" {
			t.Errorf("Expected empty subscription ID for invalid cluster ID, got %s", subscriptionID)
		}
	})

	t.Run("cluster with partial ID", func(t *testing.T) {
		clusterID := "/subscriptions/"

		cluster := &armcontainerservice.ManagedCluster{
			ID: &clusterID,
		}

		subscriptionID := extractSubscriptionIDFromCluster(cluster)
		if subscriptionID != "" {
			t.Errorf("Expected empty subscription ID for partial cluster ID, got %s", subscriptionID)
		}
	})
}
