package presenter

import (
	"fmt"
	"strings"
)

// FormatBytes converts bytes to human-readable format (B, KB, MB, GB, TB)
func FormatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(bytes)

	for _, unit := range units {
		if size < 1024.0 {
			if unit == "B" {
				return fmt.Sprintf("%.0f %s", size, unit)
			}
			return fmt.Sprintf("%.2f %s", size, unit)
		}
		size /= 1024.0
	}

	return fmt.Sprintf("%.2f TB", size*1024)
}

// FormatPercentage formats a percentage value
func FormatPercentage(percent float64) string {
	return fmt.Sprintf("%.1f%%", percent)
}

// FormatBool converts boolean to Yes/No
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

// SanitizeForJSON escapes special characters for JSON output
func SanitizeForJSON(s string) string {
	// Go's json.Marshal handles this, but this is here for reference
	return s
}

// Format types of output
type FormatType string

const (
	FormatJSON   FormatType = "json"
	FormatYAML   FormatType = "yaml"
	FormatTable  FormatType = "table"
	FormatCSV    FormatType = "csv"
)

// IsValidFormat checks if the format string is valid
func IsValidFormat(format string) bool {
	switch FormatType(format) {
	case FormatJSON, FormatYAML, FormatTable, FormatCSV:
		return true
	default:
		return false
	}
}

// ValidFormats returns list of valid formats
func ValidFormats() []string {
	return []string{
		string(FormatJSON),
		string(FormatYAML),
		string(FormatTable),
		string(FormatCSV),
	}
}
