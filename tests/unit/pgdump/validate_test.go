package pgdump_test

import (
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pgdump"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSQL(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		wantValid bool
		wantErr   bool
	}{
		{
			name: "valid CREATE TABLE",
			sql: `CREATE TABLE users (
				id SERIAL PRIMARY KEY,
				email VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "valid INSERT statement",
			sql: `INSERT INTO users (id, email, created_at) VALUES
				(1, 'user1@example.com', '2024-01-15 10:30:00'),
				(2, 'user2@example.com', '2024-01-15 10:31:00');`,
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "valid multiple statements",
			sql: `CREATE TABLE posts (id SERIAL PRIMARY KEY, title TEXT);
				INSERT INTO posts (id, title) VALUES (1, 'First Post');`,
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "valid ALTER TABLE",
			sql: `ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'active';`,
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "valid CREATE INDEX",
			sql: `CREATE INDEX idx_users_email ON users (email);`,
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "valid COPY statement",
			sql: `COPY users (id, email) FROM stdin;`,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "invalid SQL - missing parenthesis",
			sql:       "CREATE TABLE users (id SERIAL PRIMARY KEY;",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "invalid SQL - syntax error",
			sql:       "CREATE TABEL users (id SERIAL);", // TABEL instead of TABLE
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "invalid SQL - missing comma",
			sql:       "INSERT INTO users (id email) VALUES (1, 'test@example.com');",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "invalid SQL - incomplete statement",
			sql:       "CREATE TABLE",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "empty SQL",
			sql:       "",
			wantValid: true, // Empty SQL is technically valid (no errors)
			wantErr:   false,
		},
		{
			name:      "whitespace only",
			sql:       "   \n\t  ",
			wantValid: true, // Whitespace only is valid (no errors)
			wantErr:   false,
		},
		{
			name:      "comments only",
			sql:       "-- This is a comment\n/* Another comment */",
			wantValid: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := pgdump.ValidateSQL(tt.sql)

			if tt.wantErr {
				require.Error(t, err, "ValidateSQL() should return error")
				assert.False(t, valid, "SQL should be invalid when error occurs")
			} else {
				require.NoError(t, err, "ValidateSQL() should not return error")
				assert.Equal(t, tt.wantValid, valid, "SQL validity mismatch")
			}
		})
	}
}

func TestValidateSQLStatements(t *testing.T) {
	tests := []struct {
		name       string
		statements []string
		wantValid  bool
		wantErr    bool
	}{
		{
			name: "all valid statements",
			statements: []string{
				"CREATE TABLE users (id SERIAL PRIMARY KEY);",
				"CREATE TABLE posts (id SERIAL PRIMARY KEY, user_id INTEGER REFERENCES users(id));",
				"INSERT INTO users (id) VALUES (1);",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "mixed valid and invalid statements",
			statements: []string{
				"CREATE TABLE users (id SERIAL PRIMARY KEY);",
				"CREATE TABEL posts (id SERIAL);", // Invalid
				"INSERT INTO users (id) VALUES (1);",
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "single invalid statement",
			statements: []string{
				"CREATE TABLE users (id SERIAL PRIMARY KEY",
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name:       "empty statements list",
			statements: []string{},
			wantValid:  true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := pgdump.ValidateSQLStatements(tt.statements)

			if tt.wantErr {
				require.Error(t, err, "ValidateSQLStatements() should return error")
				assert.False(t, valid, "Statements should be invalid when error occurs")
			} else {
				require.NoError(t, err, "ValidateSQLStatements() should not return error")
				assert.Equal(t, tt.wantValid, valid, "Statements validity mismatch")
			}
		})
	}
}

func TestGetValidationErrors(t *testing.T) {
	tests := []struct {
		name           string
		sql            string
		wantErrCount   int
		wantErrContains string
	}{
		{
			name:           "invalid CREATE TABLE",
			sql:            "CREATE TABLE users (id SERIAL PRIMARY KEY",
			wantErrCount:   1,
			wantErrContains: "syntax error",
		},
		{
			name:           "invalid INSERT",
			sql:            "INSERT INTO users (id email) VALUES (1, 'test');",
			wantErrCount:   1,
			wantErrContains: "syntax error",
		},
		{
			name:         "valid SQL",
			sql:          "CREATE TABLE users (id SERIAL PRIMARY KEY);",
			wantErrCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := pgdump.GetValidationErrors(tt.sql)

			assert.Len(t, errors, tt.wantErrCount, "Number of validation errors mismatch")

			if tt.wantErrCount > 0 && tt.wantErrContains != "" {
				found := false
				for _, err := range errors {
					if assert.Contains(t, err.Error(), tt.wantErrContains) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error message containing %q not found", tt.wantErrContains)
			}
		})
	}
}
