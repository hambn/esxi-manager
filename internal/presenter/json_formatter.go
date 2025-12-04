package presenter

import (
	"encoding/json"
	"fmt"
)

// FormatAsJSON formats any data structure as JSON with proper indentation
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

// FormatAsJSONIndent formats any data structure as JSON with custom indentation
func FormatAsJSONIndent(data interface{}, indent string) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonData), nil
}
