package pipeline_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicPipeline(t *testing.T) {
	t.Run("generate simple two-table schema", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {
				"name": "testdb",
				"encoding": "UTF8"
			},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "username", "type": "varchar(50)"},
						{"name": "active", "type": "boolean"}
					],
					"primary_key": ["id"],
					"row_count": 10
				},
				"posts": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "title", "type": "varchar(255)"},
						{"name": "content", "type": "text"},
						{"name": "created_at", "type": "timestamp"}
					],
					"primary_key": ["id"],
					"row_count": 20
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()

		input := strings.NewReader(schemaJSON)
		output := new(bytes.Buffer)

		err := coordinator.Execute(input, output, 42)
		require.NoError(t, err)

		result := output.String()

		// Verify database creation
		assert.Contains(t, result, "CREATE DATABASE testdb")
		assert.Contains(t, result, "\\connect testdb")

		// Verify table creation
		assert.Contains(t, result, "CREATE TABLE users")
		assert.Contains(t, result, "CREATE TABLE posts")

		// Verify columns
		assert.Contains(t, result, "id serial")
		assert.Contains(t, result, "username varchar(50)")
		assert.Contains(t, result, "active boolean")
		assert.Contains(t, result, "title varchar(255)")
		assert.Contains(t, result, "content text")
		assert.Contains(t, result, "created_at timestamp")

		// Verify primary keys
		assert.Contains(t, result, "PRIMARY KEY (id)")

		// Count INSERT statements (should have 10 for users, 20 for posts)
		insertCount := strings.Count(result, "INSERT INTO")
		assert.Equal(t, 30, insertCount, "should have 30 total INSERT statements")

		usersInserts := strings.Count(result, "INSERT INTO users")
		assert.Equal(t, 10, usersInserts, "should have 10 INSERT statements for users")

		postsInserts := strings.Count(result, "INSERT INTO posts")
		assert.Equal(t, 20, postsInserts, "should have 20 INSERT statements for posts")
	})

	t.Run("deterministic generation with seed", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"items": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "name", "type": "varchar(100)"}
					],
					"row_count": 5
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()

		// Generate twice with same seed
		output1 := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output1, 12345)
		require.NoError(t, err)

		output2 := new(bytes.Buffer)
		err = coordinator.Execute(strings.NewReader(schemaJSON), output2, 12345)
		require.NoError(t, err)

		// Should produce identical output
		assert.Equal(t, output1.String(), output2.String())
	})

	t.Run("different seeds produce different data", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"items": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "value", "type": "integer"}
					],
					"row_count": 10
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()

		// Generate with different seeds
		output1 := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output1, 100)
		require.NoError(t, err)

		output2 := new(bytes.Buffer)
		err = coordinator.Execute(strings.NewReader(schemaJSON), output2, 200)
		require.NoError(t, err)

		// Should produce different output (very unlikely to be identical)
		assert.NotEqual(t, output1.String(), output2.String())
	})

	t.Run("handle schema validation errors", func(t *testing.T) {
		invalidSchema := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"}
					],
					"foreign_keys": [{
						"columns": ["user_id"],
						"referenced_table": "nonexistent",
						"referenced_columns": ["id"]
					}],
					"row_count": 5
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()

		input := strings.NewReader(invalidSchema)
		output := new(bytes.Buffer)

		err := coordinator.Execute(input, output, 42)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("handle invalid JSON", func(t *testing.T) {
		invalidJSON := `{"version": "1.0", invalid json`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()

		input := strings.NewReader(invalidJSON)
		output := new(bytes.Buffer)

		err := coordinator.Execute(input, output, 42)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse")
	})
}

func TestSerialSequence(t *testing.T) {
	t.Run("serial columns increment correctly", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "name", "type": "varchar(50)"}
					],
					"row_count": 5
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()

		output := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Check that IDs are sequential
		assert.Contains(t, result, "VALUES (1,")
		assert.Contains(t, result, "VALUES (2,")
		assert.Contains(t, result, "VALUES (3,")
		assert.Contains(t, result, "VALUES (4,")
		assert.Contains(t, result, "VALUES (5,")
	})
}