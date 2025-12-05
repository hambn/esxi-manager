package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// ============================================================================
// CLI Output Formatting (internal/interfaces/cli only)
// ============================================================================
// These formatters are used ONLY by the CLI interface to display data
// in a human-readable format. The esxi layer returns JSON, and the CLI
// layer formats it for terminal display based on the RawJSON parameter.

// FormatAsTable formats a list of maps as a table with columns
// Input: []map[string]interface{} where each map is a row
// Output: Formatted table string
//
// Example:
//   data := []map[string]interface{}{
//     {"Name": "VM1", "Status": "Running"},
//     {"Name": "VM2", "Status": "Stopped"},
//   }
//   table := FormatAsTable(data, []string{"Name", "Status"})
func FormatAsTable(rows []map[string]interface{}, columns []string) string {
	if len(rows) == 0 {
		return "No data"
	}

	var buf strings.Builder
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	// Write header
	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, col)
	}
	fmt.Fprintln(w)

	// Write rows
	for _, row := range rows {
		for i, col := range columns {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			val := row[col]
			fmt.Fprint(w, formatValue(val))
		}
		fmt.Fprintln(w)
	}

	w.Flush()
	return buf.String()
}

// formatValue converts any value to a string for table display
func formatValue(val interface{}) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

// TruncateColumn truncates column values to a maximum width for display
func TruncateColumn(value string, maxWidth int) string {
	if len(value) <= maxWidth {
		return value
	}
	if maxWidth > 3 {
		return value[:maxWidth-3] + "..."
	}
	return value[:maxWidth]
}

// PadColumn pads a column value to a minimum width
func PadColumn(value string, minWidth int) string {
	if len(value) >= minWidth {
		return value
	}
	return value + strings.Repeat(" ", minWidth-len(value))
}

// AlignLeft aligns a string to the left within a given width
func AlignLeft(value string, width int) string {
	return PadColumn(value, width)
}

// AlignRight aligns a string to the right within a given width
func AlignRight(value string, width int) string {
	if len(value) >= width {
		return value
	}
	padding := width - len(value)
	return strings.Repeat(" ", padding) + value
}
