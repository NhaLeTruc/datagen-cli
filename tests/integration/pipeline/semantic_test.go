package pipeline_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemanticGeneration(t *testing.T) {
	t.Run("semantic generators for user schema", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "email", "type": "varchar(255)"},
						{"name": "phone", "type": "varchar(20)"},
						{"name": "first_name", "type": "varchar(50)"},
						{"name": "last_name", "type": "varchar(50)"},
						{"name": "city", "type": "varchar(100)"},
						{"name": "created_at", "type": "timestamp"}
					],
					"primary_key": ["id"],
					"row_count": 3
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterSemanticGenerators()

		output := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Verify schema structure
		assert.Contains(t, result, "CREATE TABLE users")
		assert.Contains(t, result, "email varchar(255)")
		assert.Contains(t, result, "phone varchar(20)")

		// Verify semantic data (emails should contain @)
		assert.Contains(t, result, "@")

		// Should have 3 inserts
		insertCount := strings.Count(result, "INSERT INTO users")
		assert.Equal(t, 3, insertCount)
	})

	t.Run("semantic generators are deterministic", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"contacts": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "email", "type": "varchar(255)"},
						{"name": "full_name", "type": "varchar(100)"}
					],
					"row_count": 5
				}
			}
		}`

		coordinator1 := pipeline.NewCoordinator()
		coordinator1.RegisterBasicGenerators()
		coordinator1.RegisterSemanticGenerators()

		output1 := new(bytes.Buffer)
		err := coordinator1.Execute(strings.NewReader(schemaJSON), output1, 999)
		require.NoError(t, err)

		coordinator2 := pipeline.NewCoordinator()
		coordinator2.RegisterBasicGenerators()
		coordinator2.RegisterSemanticGenerators()

		output2 := new(bytes.Buffer)
		err = coordinator2.Execute(strings.NewReader(schemaJSON), output2, 999)
		require.NoError(t, err)

		// Same seed produces identical output
		assert.Equal(t, output1.String(), output2.String())
	})

	t.Run("address related semantic fields", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"locations": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "street_address", "type": "text"},
						{"name": "city", "type": "varchar(100)"},
						{"name": "country", "type": "varchar(100)"},
						{"name": "postal_code", "type": "varchar(20)"}
					],
					"row_count": 2
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterSemanticGenerators()

		output := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Verify table structure
		assert.Contains(t, result, "CREATE TABLE locations")
		assert.Contains(t, result, "street_address text")

		// Should have 2 inserts
		insertCount := strings.Count(result, "INSERT INTO locations")
		assert.Equal(t, 2, insertCount)
	})

	t.Run("fallback to basic generators for unknown columns", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"products": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "email", "type": "varchar(255)"},
						{"name": "price", "type": "integer"},
						{"name": "active", "type": "boolean"}
					],
					"row_count": 2
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterSemanticGenerators()

		output := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Verify both semantic and basic types work
		assert.Contains(t, result, "@") // email
		assert.Contains(t, result, "VALUES (1,") // serial
		assert.Regexp(t, `(?i)(true|false)`, result) // boolean (case-insensitive)
	})
}