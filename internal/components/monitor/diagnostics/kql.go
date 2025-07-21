package diagnostics

import (
	"fmt"
	"strings"
	"time"
)

// Log level prefixes used in Kubernetes logs
var logLevelPrefixes = map[string]string{
	"info":    "I",
	"warning": "W",
	"error":   "E",
}

// Audit log categories that use different log level format
var auditCategories = map[string]bool{
	"kube-audit":       true,
	"kube-audit-admin": true,
}

// Resource-specific table mappings for AKS log categories
var resourceSpecificTables = map[string]string{
	"kube-audit":                        "AKSAudit",
	"kube-audit-admin":                  "AKSAuditAdmin", 
	"kube-apiserver":                    "AKSControlPlane",
	"kube-controller-manager":           "AKSControlPlane",
	"kube-scheduler":                    "AKSControlPlane",
	"cluster-autoscaler":                "AKSControlPlane",
	"cloud-controller-manager":          "AKSControlPlane",
	"guard":                             "AKSControlPlane",
	"csi-azuredisk-controller":          "AKSControlPlane",
	"csi-azurefile-controller":          "AKSControlPlane",
	"csi-snapshot-controller":           "AKSControlPlane",
}

// isAuditCategory checks if the given category is an audit log category
func isAuditCategory(category string) bool {
	return auditCategories[category]
}

// BuildSafeKQLQuery builds pre-validated KQL queries to prevent injection, scoped to specific AKS cluster
// Supports both Azure Diagnostics and Resource-specific destination tables
func BuildSafeKQLQuery(category, logLevel string, maxRecords int, clusterResourceID string, isResourceSpecific bool) string {
	var baseQuery string
	var actuallyUsingResourceSpecific bool
	
	if isResourceSpecific {
		// Use resource-specific table
		if tableName, exists := resourceSpecificTables[category]; exists {
			// For resource-specific tables, _ResourceId is stored in lowercase
			// Convert the resource ID to lowercase to match Azure's storage format
			lowerResourceID := strings.ToLower(clusterResourceID)
			baseQuery = fmt.Sprintf("%s | where _ResourceId == '%s'", tableName, lowerResourceID)
			actuallyUsingResourceSpecific = true
		} else {
			// Fallback to Azure Diagnostics table if no resource-specific mapping found
			// Azure Diagnostics uses uppercase ResourceId
			upperResourceID := strings.ToUpper(clusterResourceID)
			baseQuery = fmt.Sprintf("AzureDiagnostics | where Category == '%s' and ResourceId == '%s'", category, upperResourceID)
			actuallyUsingResourceSpecific = false
		}
	} else {
		// Use Azure Diagnostics table (legacy mode)
		// Azure Diagnostics ResourceId field is stored in uppercase
		upperResourceID := strings.ToUpper(clusterResourceID)
		baseQuery = fmt.Sprintf("AzureDiagnostics | where Category == '%s' and ResourceId == '%s'", category, upperResourceID)
		actuallyUsingResourceSpecific = false
	}

	// Add log level filtering for non-audit logs
	if logLevel != "" && !isAuditCategory(category) {
		if actuallyUsingResourceSpecific {
			// In resource-specific tables, log level is stored in the Level field as "INFO", "WARNING", "ERROR"
			// Convert the requested log level to the format used in resource-specific tables
			switch strings.ToLower(logLevel) {
			case "info":
				baseQuery += " | where Level == 'INFO'"
			case "warning":
				baseQuery += " | where Level == 'WARNING'"
			case "error":
				baseQuery += " | where Level == 'ERROR'"
			}
		} else {
			// For Azure Diagnostics, use the log_s prefix pattern
			// Kubernetes logs use format like "I0715" (Info), "W0715" (Warning), "E0715" (Error)
			if prefix, exists := logLevelPrefixes[strings.ToLower(logLevel)]; exists {
				baseQuery += fmt.Sprintf(" | where log_s startswith '%s'", prefix)
			}
		}
	}

	baseQuery += " | order by TimeGenerated desc"
	baseQuery += fmt.Sprintf(" | limit %d", maxRecords)

	// Project essential fields - adjust based on table type
	if actuallyUsingResourceSpecific {
		// Resource-specific tables have different field names based on the table type
		if tableName, exists := resourceSpecificTables[category]; exists {
			if tableName == "AKSAudit" || tableName == "AKSAuditAdmin" {
				// Audit tables have structured fields, no single message field
				baseQuery += " | project TimeGenerated, Level, AuditId, Stage, RequestUri, Verb, User"
			} else {
				// AKSControlPlane table has Message field
				baseQuery += " | project TimeGenerated, Category, Level, Message, PodName"
			}
		} else {
			// Fallback projection for unknown resource-specific tables
			baseQuery += " | project TimeGenerated, Level, Message"
		}
	} else {
		// Azure Diagnostics table fields
		baseQuery += " | project TimeGenerated, Level, log_s"
	}

	return baseQuery
}

// CalculateTimespan converts start/end times to Azure CLI timespan format
func CalculateTimespan(startTime, endTime string) (string, error) {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return "", fmt.Errorf("invalid start time format: %w", err)
	}

	var end time.Time
	if endTime != "" {
		end, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			return "", fmt.Errorf("invalid end time format: %w", err)
		}
	} else {
		// Default to current time if no end time specified
		end = time.Now()
	}

	// Azure CLI timespan format: start_time/end_time in ISO8601
	timespan := fmt.Sprintf("%s/%s", start.Format(time.RFC3339), end.Format(time.RFC3339))
	return timespan, nil
}
