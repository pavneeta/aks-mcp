// Package azure provides Azure SDK integration for AKS MCP server.
package azure

// toString safely converts a potentially nil string pointer to a string.
func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
