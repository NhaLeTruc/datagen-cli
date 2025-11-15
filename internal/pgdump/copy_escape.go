package pgdump

import (
	"fmt"
	"strings"
	"time"
)

// EscapeCopyValue escapes a value for PostgreSQL COPY format
// COPY format uses backslash escapes for special characters
func EscapeCopyValue(val interface{}) string {
	if val == nil {
		return "\\N" // NULL representation in COPY format
	}

	switch v := val.(type) {
	case string:
		return escapeCopyString(v)
	case int, int32, int64, uint, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		if v {
			return "t" // true in COPY format
		}
		return "f" // false in COPY format
	case time.Time:
		// Format timestamp in PostgreSQL-compatible format
		return escapeCopyString(v.Format("2006-01-02 15:04:05"))
	default:
		// For other types, convert to string and escape
		return escapeCopyString(fmt.Sprintf("%v", v))
	}
}

// escapeCopyString escapes special characters in a string for COPY format
// According to PostgreSQL COPY format specification:
// - Backslash (\) → \\
// - Newline (\n) → \n
// - Carriage return (\r) → \r
// - Tab (\t) → \t
func escapeCopyString(s string) string {
	// Use strings.Builder for efficient string concatenation
	var result strings.Builder
	result.Grow(len(s) + 10) // Pre-allocate with some extra space for escapes

	for _, ch := range s {
		switch ch {
		case '\\':
			result.WriteString("\\\\")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		default:
			result.WriteRune(ch)
		}
	}

	return result.String()
}

// FormatCopyRow formats a complete row for COPY format
// Values are tab-separated
func FormatCopyRow(columns []string, row map[string]interface{}) string {
	values := make([]string, len(columns))
	for i, col := range columns {
		values[i] = EscapeCopyValue(row[col])
	}
	return strings.Join(values, "\t")
}
