package security

import "strings"

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// AccessLevel controls the level of operations allowed (readonly, readwrite, admin)
	AccessLevel string
	// AllowedNamespaces is a comma-separated list of allowed Kubernetes namespaces
	AllowedNamespaces string
}

// NewSecurityConfig creates a new SecurityConfig instance
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AccessLevel:       "readwrite",
		AllowedNamespaces: "",
	}
}

// IsNamespaceAllowed checks if a namespace is allowed to be accessed
func (s *SecurityConfig) IsNamespaceAllowed(namespace string) bool {
	// If no restrictions are defined, allow all namespaces
	if s.AllowedNamespaces == "" {
		return true
	}

	// Check if the namespace is in the allowed list
	namespaces := strings.Split(s.AllowedNamespaces, ",")
	for _, ns := range namespaces {
		if strings.TrimSpace(ns) == namespace {
			return true
		}
	}

	return false
}
