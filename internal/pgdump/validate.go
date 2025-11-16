package pgdump

import (
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

// ValidateSQL validates a SQL string using PostgreSQL's actual parser.
// Returns true if the SQL is valid, false otherwise.
// Returns an error if validation fails with details about the syntax error.
func ValidateSQL(sql string) (bool, error) {
	// Empty or whitespace-only SQL is considered valid
	if strings.TrimSpace(sql) == "" {
		return true, nil
	}

	// Parse the SQL using pg_query
	_, err := pg_query.Parse(sql)
	if err != nil {
		return false, fmt.Errorf("SQL validation failed: %w", err)
	}

	return true, nil
}

// ValidateSQLStatements validates multiple SQL statements.
// Returns true if all statements are valid, false otherwise.
// Returns an error on the first invalid statement encountered.
func ValidateSQLStatements(statements []string) (bool, error) {
	// Empty list is considered valid
	if len(statements) == 0 {
		return true, nil
	}

	for i, stmt := range statements {
		valid, err := ValidateSQL(stmt)
		if err != nil {
			return false, fmt.Errorf("statement %d validation failed: %w", i+1, err)
		}
		if !valid {
			return false, fmt.Errorf("statement %d is invalid", i+1)
		}
	}

	return true, nil
}

// GetValidationErrors returns all validation errors for a SQL string.
// Returns an empty slice if the SQL is valid.
// This is useful for collecting all errors instead of failing on the first one.
func GetValidationErrors(sql string) []error {
	// Empty or whitespace-only SQL has no errors
	if strings.TrimSpace(sql) == "" {
		return []error{}
	}

	// Try to parse the SQL
	_, err := pg_query.Parse(sql)
	if err != nil {
		return []error{err}
	}

	return []error{}
}

// ValidateSQLWithDetails validates SQL and returns detailed error information.
// Returns the parse tree on success, or detailed error information on failure.
func ValidateSQLWithDetails(sql string) (*pg_query.ParseResult, error) {
	// Empty or whitespace-only SQL
	if strings.TrimSpace(sql) == "" {
		return nil, nil
	}

	// Parse the SQL using pg_query
	result, err := pg_query.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("SQL syntax error: %w", err)
	}

	return result, nil
}

// ValidationResult contains the results of SQL validation
type ValidationResult struct {
	Valid      bool
	Errors     []error
	StatementCount int
}

// ValidateSQLDetailed performs detailed validation and returns comprehensive results.
func ValidateSQLDetailed(sql string) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Errors:     []error{},
		StatementCount: 0,
	}

	// Empty or whitespace-only SQL
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return result
	}

	// Parse the SQL
	parseResult, err := pg_query.Parse(sql)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err)
		return result
	}

	// Count statements
	if parseResult != nil && parseResult.Stmts != nil {
		result.StatementCount = len(parseResult.Stmts)
	}

	return result
}

// ValidateAndNormalize validates SQL and returns a normalized version.
// Normalization includes:
// - Consistent formatting
// - Removal of unnecessary whitespace
// - Standard capitalization (optional)
func ValidateAndNormalize(sql string) (string, error) {
	// Empty or whitespace-only SQL
	if strings.TrimSpace(sql) == "" {
		return "", nil
	}

	// First validate
	_, err := pg_query.Parse(sql)
	if err != nil {
		return "", fmt.Errorf("cannot normalize invalid SQL: %w", err)
	}

	// Normalize using pg_query
	normalized, err := pg_query.Normalize(sql)
	if err != nil {
		// If normalization fails, return original SQL if it's valid
		return sql, nil
	}

	return normalized, nil
}
