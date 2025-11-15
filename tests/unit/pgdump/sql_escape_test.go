package pgdump_test

import (
	"testing"
	"time"

	"github.com/NhaLeTruc/datagen-cli/internal/pgdump"
	"github.com/stretchr/testify/assert"
)

func TestEscapeIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase identifier",
			input:    "users",
			expected: "users",
		},
		{
			name:     "identifier with uppercase",
			input:    "Users",
			expected: "\"Users\"",
		},
		{
			name:     "identifier with special characters",
			input:    "user_name",
			expected: "user_name",
		},
		{
			name:     "identifier starting with number",
			input:    "1users",
			expected: "\"1users\"",
		},
		{
			name:     "reserved keyword",
			input:    "select",
			expected: "\"select\"",
		},
		{
			name:     "reserved keyword user",
			input:    "user",
			expected: "\"user\"",
		},
		{
			name:     "identifier with double quotes",
			input:    "user\"name",
			expected: "\"user\"\"name\"",
		},
		{
			name:     "identifier with spaces",
			input:    "user name",
			expected: "\"user name\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pgdump.EscapeIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with single quote",
			input:    "O'Reilly",
			expected: "O''Reilly",
		},
		{
			name:     "string with multiple quotes",
			input:    "It's a \"test\"",
			expected: "It''s a \"test\"",
		},
		{
			name:     "string with backslash",
			input:    "C:\\Users\\test",
			expected: "C:\\\\Users\\\\test",
		},
		{
			name:     "string with newline",
			input:    "line1\nline2",
			expected: "line1\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pgdump.EscapeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuoteString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "'hello'",
		},
		{
			name:     "string with quote",
			input:    "O'Reilly",
			expected: "'O''Reilly'",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pgdump.QuoteString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "NULL value",
			input:    nil,
			expected: "NULL",
		},
		{
			name:     "string value",
			input:    "test",
			expected: "'test'",
		},
		{
			name:     "string with quote",
			input:    "O'Reilly",
			expected: "'O''Reilly'",
		},
		{
			name:     "integer",
			input:    42,
			expected: "42",
		},
		{
			name:     "int32",
			input:    int32(123),
			expected: "123",
		},
		{
			name:     "int64",
			input:    int64(999),
			expected: "999",
		},
		{
			name:     "float32",
			input:    float32(3.14),
			expected: "3.140000",
		},
		{
			name:     "float64",
			input:    float64(2.71828),
			expected: "2.718280",
		},
		{
			name:     "boolean true",
			input:    true,
			expected: "TRUE",
		},
		{
			name:     "boolean false",
			input:    false,
			expected: "FALSE",
		},
		{
			name:     "timestamp",
			input:    time.Date(2023, 1, 15, 10, 30, 45, 0, time.UTC),
			expected: "'2023-01-15 10:30:45'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pgdump.FormatValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValueList(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected string
	}{
		{
			name:     "simple values",
			input:    []interface{}{1, "test", true},
			expected: "1, 'test', TRUE",
		},
		{
			name:     "with NULL",
			input:    []interface{}{1, nil, "value"},
			expected: "1, NULL, 'value'",
		},
		{
			name:     "single value",
			input:    []interface{}{42},
			expected: "42",
		},
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pgdump.FormatValueList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatIdentifierList(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "simple identifiers",
			input:    []string{"id", "name", "email"},
			expected: "id, name, email",
		},
		{
			name:     "with reserved keyword",
			input:    []string{"id", "user", "email"},
			expected: "id, \"user\", email",
		},
		{
			name:     "with uppercase",
			input:    []string{"id", "Name", "email"},
			expected: "id, \"Name\", email",
		},
		{
			name:     "single identifier",
			input:    []string{"users"},
			expected: "users",
		},
		{
			name:     "empty list",
			input:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pgdump.FormatIdentifierList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
