package azureclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
)

func TestAzureCache_SetAndGet(t *testing.T) {
	cache := NewAzureCache(5 * time.Minute)

	// Test setting and getting a value
	key := "test-key"
	value := "test-value"
	cache.Set(key, value)

	retrieved, found := cache.Get(key)
	if !found {
		t.Errorf("Expected to find cached value, but it was not found")
	}

	if retrieved != value {
		t.Errorf("Expected retrieved value to be %v, got %v", value, retrieved)
	}
}

func TestAzureCache_GetNonExistent(t *testing.T) {
	cache := NewAzureCache(5 * time.Minute)

	// Test getting a non-existent value
	retrieved, found := cache.Get("non-existent-key")
	if found {
		t.Errorf("Expected not to find cached value, but it was found")
	}

	if retrieved != nil {
		t.Errorf("Expected retrieved value to be nil, got %v", retrieved)
	}
}

func TestAzureCache_Expiration(t *testing.T) {
	// Create cache with very short timeout for testing
	cache := NewAzureCache(50 * time.Millisecond)

	key := "test-key"
	value := "test-value"
	cache.Set(key, value)

	// Value should be there immediately
	retrieved, found := cache.Get(key)
	if !found {
		t.Errorf("Expected to find cached value immediately after setting")
	}
	if retrieved != value {
		t.Errorf("Expected retrieved value to be %v, got %v", value, retrieved)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Value should be expired now
	retrieved, found = cache.Get(key)
	if found {
		t.Errorf("Expected cached value to be expired, but it was found")
	}
	if retrieved != nil {
		t.Errorf("Expected retrieved value to be nil after expiration, got %v", retrieved)
	}
}

func TestAzureCache_Clear(t *testing.T) {
	cache := NewAzureCache(5 * time.Minute)

	// Set multiple values
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Verify they exist
	if _, found := cache.Get("key1"); !found {
		t.Errorf("Expected key1 to be found before clear")
	}

	// Clear cache
	cache.Clear()

	// Verify they're gone
	if _, found := cache.Get("key1"); found {
		t.Errorf("Expected key1 to be gone after clear")
	}
	if _, found := cache.Get("key2"); found {
		t.Errorf("Expected key2 to be gone after clear")
	}
	if _, found := cache.Get("key3"); found {
		t.Errorf("Expected key3 to be gone after clear")
	}
}

func TestAzureCache_ComplexTypes(t *testing.T) {
	cache := NewAzureCache(5 * time.Minute)

	// Test with a complex Azure type
	cluster := &armcontainerservice.ManagedCluster{
		Name:     stringPtr("test-cluster"),
		Location: stringPtr("eastus"),
	}

	key := "cluster-key"
	cache.Set(key, cluster)

	retrieved, found := cache.Get(key)
	if !found {
		t.Errorf("Expected to find cached cluster")
	}

	retrievedCluster, ok := retrieved.(*armcontainerservice.ManagedCluster)
	if !ok {
		t.Errorf("Expected retrieved value to be a ManagedCluster")
	}

	if retrievedCluster.Name == nil || *retrievedCluster.Name != "test-cluster" {
		t.Errorf("Expected cluster name to be 'test-cluster', got %v", retrievedCluster.Name)
	}
}

func TestAzureCache_ConcurrentAccess(t *testing.T) {
	cache := NewAzureCache(5 * time.Minute)

	// Test concurrent writes and reads
	done := make(chan bool, 10)

	// Start multiple goroutines writing
	for i := 0; i < 5; i++ {
		go func(id int) {
			cache.Set(fmt.Sprintf("key-%d", id), fmt.Sprintf("value-%d", id))
			done <- true
		}(i)
	}

	// Start multiple goroutines reading
	for i := 0; i < 5; i++ {
		go func(id int) {
			cache.Get(fmt.Sprintf("key-%d", id))
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify some values exist
	if _, found := cache.Get("key-0"); !found {
		t.Errorf("Expected key-0 to exist after concurrent operations")
	}
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
