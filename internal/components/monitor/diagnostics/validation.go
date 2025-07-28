package diagnostics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/aks-mcp/internal/components/common"
)

// Constants for diagnostics configuration
const (
	MaxLogRetentionDays   = 7
	MaxQueryRangeDuration = 24 * time.Hour
	DefaultMaxRecords     = 100
	MaxAllowedRecords     = 1000
)

// ValidateControlPlaneLogsParams validates all parameters for control plane logs query
func ValidateControlPlaneLogsParams(params map[string]interface{}) error {
	// Validate AKS parameters using common helper
	_, _, _, err := common.ExtractAKSParameters(params)
	if err != nil {
		return err
	}

	// Validate remaining required parameters
	required := []string{"log_category", "start_time"}
	for _, param := range required {
		if value, ok := params[param].(string); !ok || value == "" {
			return fmt.Errorf("missing or invalid %s parameter", param)
		}
	}

	// Validate log category
	logCategory := params["log_category"].(string)
	validCategories := []string{
		"kube-apiserver",
		"kube-audit",
		"kube-audit-admin",
		"kube-controller-manager",
		"kube-scheduler",
		"cluster-autoscaler",
		"cloud-controller-manager",
		"guard",
		"csi-azuredisk-controller",
		"csi-azurefile-controller",
		"csi-snapshot-controller",
		"fleet-member-agent",
		"fleet-member-net-controller-manager",
		"fleet-mcs-controller-manager",
	}

	valid := false
	for _, validCat := range validCategories {
		if logCategory == validCat {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log category: %s. Valid categories: %s", logCategory, strings.Join(validCategories, ", "))
	}

	// Validate time range
	startTime := params["start_time"].(string)
	if err := ValidateTimeRange(startTime, params); err != nil {
		return err
	}

	// Validate log level if provided
	if logLevel, ok := params["log_level"].(string); ok && logLevel != "" {
		validLevels := []string{"error", "warning", "info"}
		validLevel := false
		for _, level := range validLevels {
			if logLevel == level {
				validLevel = true
				break
			}
		}
		if !validLevel {
			return fmt.Errorf("invalid log level: %s. Valid levels: %s", logLevel, strings.Join(validLevels, ", "))
		}
	}

	return nil
}

// ValidateTimeRange validates start and end time parameters
func ValidateTimeRange(startTime string, params map[string]interface{}) error {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return fmt.Errorf("invalid start_time format, expected RFC3339 (ISO 8601): %w", err)
	}

	// // Check if start time is not more than the maximum retention period
	// maxRetentionAgo := time.Now().AddDate(0, 0, -MaxLogRetentionDays)
	// if start.Before(maxRetentionAgo) {
	// 	return fmt.Errorf("start_time cannot be more than %d days ago", MaxLogRetentionDays)
	// }

	// Check if start time is in the future
	if start.After(time.Now()) {
		return fmt.Errorf("start_time cannot be in the future")
	}

	// Validate end time if provided
	if endTimeStr, ok := params["end_time"].(string); ok && endTimeStr != "" {
		end, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return fmt.Errorf("invalid end_time format, expected RFC3339 (ISO 8601): %w", err)
		}

		// Check if time range exceeds maximum query duration
		if end.Sub(start) > MaxQueryRangeDuration {
			return fmt.Errorf("time range cannot exceed %v", MaxQueryRangeDuration)
		}

		if end.Before(start) {
			return fmt.Errorf("end_time must be after start_time")
		}

		if end.After(time.Now()) {
			return fmt.Errorf("end_time cannot be in the future")
		}
	}

	return nil
}

// GetMaxRecords extracts and validates the max_records parameter
func GetMaxRecords(params map[string]interface{}) int {
	if val, ok := params["max_records"].(string); ok && val != "" {
		if recordsInt, err := strconv.Atoi(val); err == nil {
			if recordsInt > MaxAllowedRecords {
				return MaxAllowedRecords
			}
			if recordsInt < 1 {
				return DefaultMaxRecords
			}
			return recordsInt
		}
	}
	return DefaultMaxRecords
}
