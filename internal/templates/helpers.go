package templates

import (
	"fmt"
	"strconv"
)

// formatBytes formats bytes into human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatBytes is a template function wrapper
func FormatBytes(bytes int64) string {
	return formatBytes(bytes)
}

// Itoa converts int to string for templates
func Itoa(i int) string {
	return strconv.Itoa(i)
}
