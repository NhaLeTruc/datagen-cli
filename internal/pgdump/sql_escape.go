package pgdump

import (
	"fmt"
	"strings"
	"time"
)

// EscapeIdentifier escapes a PostgreSQL identifier (table name, column name, etc.)
// following PostgreSQL quoting rules
func EscapeIdentifier(ident string) string {
	// If identifier is a reserved keyword or contains special characters,
	// quote it with double quotes
	if needsQuoting(ident) {
		// Escape any double quotes in the identifier by doubling them
		escaped := strings.ReplaceAll(ident, "\"", "\"\"")
		return fmt.Sprintf("\"%s\"", escaped)
	}
	return ident
}

// EscapeString escapes a string value for use in SQL
func EscapeString(s string) string {
	// Escape single quotes by doubling them
	// This is the standard SQL way to escape quotes
	escaped := strings.ReplaceAll(s, "'", "''")

	// Also escape backslashes if present (for PostgreSQL)
	escaped = strings.ReplaceAll(escaped, "\\", "\\\\")

	return escaped
}

// QuoteString quotes and escapes a string value for SQL
func QuoteString(s string) string {
	return fmt.Sprintf("'%s'", EscapeString(s))
}

// FormatValue formats any value for SQL, handling NULL, strings, numbers, booleans, and timestamps
func FormatValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case string:
		return QuoteString(v)
	case int:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case uint:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return fmt.Sprintf("%f", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case time.Time:
		// Format timestamp in PostgreSQL-compatible format
		return QuoteString(v.Format("2006-01-02 15:04:05"))
	default:
		// For other types, convert to string and quote
		return QuoteString(fmt.Sprintf("%v", v))
	}
}

// needsQuoting checks if an identifier needs to be quoted
func needsQuoting(ident string) bool {
	if len(ident) == 0 {
		return true
	}

	// Check if identifier starts with a digit
	if ident[0] >= '0' && ident[0] <= '9' {
		return true
	}

	// Check for uppercase letters or special characters
	for _, ch := range ident {
		if (ch >= 'A' && ch <= 'Z') ||
			!(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '_') {
			return true
		}
	}

	// Check if it's a reserved keyword (simplified list)
	reserved := map[string]bool{
		"select": true, "insert": true, "update": true, "delete": true,
		"from": true, "where": true, "join": true, "table": true,
		"create": true, "alter": true, "drop": true, "index": true,
		"user": true, "order": true, "group": true, "having": true,
	}

	if reserved[strings.ToLower(ident)] {
		return true
	}

	return false
}

// FormatValueList formats a list of values for use in VALUES clause
func FormatValueList(values []interface{}) string {
	formatted := make([]string, len(values))
	for i, val := range values {
		formatted[i] = FormatValue(val)
	}
	return strings.Join(formatted, ", ")
}

// FormatIdentifierList formats a list of identifiers (column names, etc.)
func FormatIdentifierList(idents []string) string {
	formatted := make([]string, len(idents))
	for i, ident := range idents {
		formatted[i] = EscapeIdentifier(ident)
	}
	return strings.Join(formatted, ", ")
}
