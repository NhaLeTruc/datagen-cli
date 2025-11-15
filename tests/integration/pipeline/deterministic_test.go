package pipeline_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeterministicGeneration(t *testing.T) {
	t.Run("same seed produces identical output", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "email", "type": "varchar(255)"},
						{"name": "first_name", "type": "varchar(50)"},
						{"name": "age", "type": "integer"}
					],
					"primary_key": ["id"],
					"row_count": 10
				}
			}
		}`

		coordinator1 := pipeline.NewCoordinator()
		coordinator1.RegisterBasicGenerators()
		coordinator1.RegisterSemanticGenerators()

		output1 := new(bytes.Buffer)
		err := coordinator1.Execute(strings.NewReader(schemaJSON), output1, 12345)
		require.NoError(t, err)

		coordinator2 := pipeline.NewCoordinator()
		coordinator2.RegisterBasicGenerators()
		coordinator2.RegisterSemanticGenerators()

		output2 := new(bytes.Buffer)
		err = coordinator2.Execute(strings.NewReader(schemaJSON), output2, 12345)
		require.NoError(t, err)

		// Same seed produces byte-identical output
		assert.Equal(t, output1.String(), output2.String(), "Same seed should produce identical output")
		assert.Equal(t, output1.Len(), output2.Len(), "Output lengths should be identical")
	})

	t.Run("same seed with custom generators produces identical output", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"products": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "sku", "type": "varchar(20)", "generator_config": {"type": "pattern", "pattern": "[A-Z]{3}-\\d{4}"}},
						{"name": "status", "type": "varchar(20)", "generator_config": {"type": "weighted_enum", "weights": {"active": 0.8, "inactive": 0.2}}}
					],
					"primary_key": ["id"],
					"row_count": 20
				}
			}
		}`

		coordinator1 := pipeline.NewCoordinator()
		coordinator1.RegisterBasicGenerators()
		coordinator1.RegisterSemanticGenerators()

		output1 := new(bytes.Buffer)
		err := coordinator1.Execute(strings.NewReader(schemaJSON), output1, 99999)
		require.NoError(t, err)

		coordinator2 := pipeline.NewCoordinator()
		coordinator2.RegisterBasicGenerators()
		coordinator2.RegisterSemanticGenerators()

		output2 := new(bytes.Buffer)
		err = coordinator2.Execute(strings.NewReader(schemaJSON), output2, 99999)
		require.NoError(t, err)

		// Same seed with custom generators produces identical output
		assert.Equal(t, output1.String(), output2.String(), "Same seed with custom generators should produce identical output")
	})

	t.Run("same seed with multiple tables produces identical rows for each table", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "email", "type": "varchar(255)"}
					],
					"primary_key": ["id"],
					"row_count": 5
				},
				"posts": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "user_id", "type": "integer"},
						{"name": "title", "type": "varchar(200)"}
					],
					"primary_key": ["id"],
					"row_count": 15
				}
			}
		}`

		coordinator1 := pipeline.NewCoordinator()
		coordinator1.RegisterBasicGenerators()
		coordinator1.RegisterSemanticGenerators()

		output1 := new(bytes.Buffer)
		err := coordinator1.Execute(strings.NewReader(schemaJSON), output1, 54321)
		require.NoError(t, err)

		coordinator2 := pipeline.NewCoordinator()
		coordinator2.RegisterBasicGenerators()
		coordinator2.RegisterSemanticGenerators()

		output2 := new(bytes.Buffer)
		err = coordinator2.Execute(strings.NewReader(schemaJSON), output2, 54321)
		require.NoError(t, err)

		result1 := output1.String()
		result2 := output2.String()

		// Same seed with multiple tables produces identical output
		// (note: table ordering in Go maps is not deterministic, but data within each table should be)
		assert.Equal(t, result1, result2, "Same seed with multiple tables should produce identical output")

		// Verify both have the same number of inserts
		assert.Equal(t, strings.Count(result1, "INSERT INTO users"), strings.Count(result2, "INSERT INTO users"))
		assert.Equal(t, strings.Count(result1, "INSERT INTO posts"), strings.Count(result2, "INSERT INTO posts"))
	})
}