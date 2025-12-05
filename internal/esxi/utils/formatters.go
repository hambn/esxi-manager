package utils

import "fmt"

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
