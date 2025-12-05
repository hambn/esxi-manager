package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ============================================================================
// JSON Formatting (for internal/esxi output - ONLY format used in esxi layer)
// ============================================================================

// FormatAsJSON formats any data structure as JSON with proper indentation
// Used by ALL esxi commands for output - this is the ONLY output format from esxi layer
func FormatAsJSON(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonData), nil
}

// FormatAsJSONCompact formats any data structure as JSON without indentation
func FormatAsJSONCompact(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonData), nil
}

// ============================================================================
// Value Formatters (used by internal/esxi commands for formatting data values)
// ============================================================================

// FormatBytes converts bytes to human-readable format with max 3 digits
// Examples: 512B, 1.5KB, 256MB, 1.23GB, 4.56TB
func FormatBytes(bytes int64) string {
	if bytes < 0 {
		return "0B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	value := float64(bytes)

	for i, unit := range units {
		if value < 1024.0 || i == len(units)-1 {
			if value < 10 {
				return fmt.Sprintf("%.2f%s", value, unit)
			} else if value < 100 {
				return fmt.Sprintf("%.1f%s", value, unit)
			} else {
				return fmt.Sprintf("%.0f%s", value, unit)
			}
		}
		value /= 1024.0
	}
	return fmt.Sprintf("%.0f%s", value, units[len(units)-1])
}

// FormatPercentage formats a percentage value
func FormatPercentage(percent float64) string {
	return fmt.Sprintf("%.1f%%", percent)
}

// FormatBool converts boolean to Yes/No (for display purposes)
func FormatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen > 3 {
		return s[:maxLen-3] + "..."
	}
	return s[:maxLen]
}

// PadString pads a string to a minimum length
func PadString(s string, minLen int) string {
	if len(s) >= minLen {
		return s
	}
	return s + strings.Repeat(" ", minLen-len(s))
}

// JoinStrings joins strings with a separator, skipping empty strings
func JoinStrings(separator string, strs ...string) string {
	var nonEmpty []string
	for _, s := range strs {
		if strings.TrimSpace(s) != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return strings.Join(nonEmpty, separator)
}
