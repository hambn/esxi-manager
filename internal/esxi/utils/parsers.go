package utils

import (
	"strings"
)

// ParseMultiWordName parses a field that may contain spaces, followed by fixed-width fields.
// Splits the line by whitespace, then reconstructs the name from the fields before the fixed-width columns.
//
// Example:
//   fields := strings.Fields("Management Network vSwitch0 1 0")
//   name := ParseMultiWordName(fields, 3)  // "Management Network"
//   // Last 3 fields are: vSwitch0 (vswitch), 1 (activeClients), 0 (vlanID)
func ParseMultiWordName(fields []string, fixedFieldCount int) string {
	if len(fields) <= fixedFieldCount {
		if len(fields) > 0 {
			return fields[0]
		}
		return ""
	}
	return strings.Join(fields[:len(fields)-fixedFieldCount], " ")
}

// ParseMultiLineKeyValue parses output with multi-line format where indentation indicates hierarchy.
// Common pattern in esxcli output:
//   Name1
//     Key1: Value1
//     Key2: Value2
//   Name2
//     Key1: Value1
//
// The parser calls recordFunc whenever a new unindented line is found (new record),
// and parses indented lines as key-value pairs, calling parseFunc for each pair.
//
// Example:
//   lines := []string{
//     "vSwitch0",
//     "  Num Ports: 1536",
//     "  MTU: 1500",
//     "vSwitch1",
//     "  Num Ports: 128",
//   }
//   var switches []Switch
//   currentSwitch := &Switch{}
//   ParseMultiLineKeyValue(lines,
//     func() { switches = append(switches, *currentSwitch) },
//     func(key, val string) {
//       if key == "Num Ports" {
//         currentSwitch.NumPorts = val
//       }
//     },
//   )
func ParseMultiLineKeyValue(
	lines []string,
	recordFunc func(),
	parseFunc func(key, value string),
) {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check if line is indented (continuation of current record)
		isIndented := strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")

		if !isIndented {
			// New record found, call record function for previous record
			recordFunc()
		} else {
			// Parse indented line as key-value pair
			if strings.Contains(trimmed, ":") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					parseFunc(key, value)
				}
			}
		}
	}
	// Don't forget the last record
	recordFunc()
}

// ParseKeyValueLine parses a single line with "Key: Value" format.
// Returns the key and value, trimmed of whitespace.
//
// Example:
//   key, val, ok := ParseKeyValueLine("MTU: 1500")
//   // key: "MTU", val: "1500", ok: true
func ParseKeyValueLine(line string) (key, value string, ok bool) {
	if !strings.Contains(line, ":") {
		return "", "", false
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

// IsLineIndented checks if a line is indented (starts with space or tab).
// Useful for detecting line hierarchy in structured output.
func IsLineIndented(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}

// SkipHeaderLines skips the first n non-empty lines (typically headers and separators).
// Returns slice of lines starting from the first data line.
//
// Example:
//   lines := []string{"Name", "----", "VM1", "VM2"}
//   dataLines := SkipHeaderLines(lines, 2)
//   // dataLines: ["VM1", "VM2"]
func SkipHeaderLines(lines []string, headerLineCount int) []string {
	if headerLineCount >= len(lines) {
		return []string{}
	}
	return lines[headerLineCount:]
}

// TrimAndSplit splits a string by newlines, trims whitespace, and filters empty lines.
func TrimAndSplit(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
