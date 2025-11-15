package pipeline_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedVariation(t *testing.T) {
	t.Run("different seeds produce different output", func(t *testing.T) {
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
		err := coordinator1.Execute(strings.NewReader(schemaJSON), output1, 11111)
		require.NoError(t, err)

		coordinator2 := pipeline.NewCoordinator()
		coordinator2.RegisterBasicGenerators()
		coordinator2.RegisterSemanticGenerators()

		output2 := new(bytes.Buffer)
		err = coordinator2.Execute(strings.NewReader(schemaJSON), output2, 22222)
		require.NoError(t, err)

		// Different seeds produce different output
		assert.NotEqual(t, output1.String(), output2.String(), "Different seeds should produce different output")

		// But both should have the same structure (same number of INSERTs)
		insertCount1 := strings.Count(output1.String(), "INSERT INTO users")
		insertCount2 := strings.Count(output2.String(), "INSERT INTO users")
		assert.Equal(t, insertCount1, insertCount2, "Both should generate same number of rows")
		assert.Equal(t, 10, insertCount1, "Should have exactly 10 inserts")
	})

	t.Run("different seeds with weighted distributions maintain distribution", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"orders": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "status", "type": "varchar(20)", "generator_config": {"type": "weighted_enum", "weights": {"completed": 0.70, "pending": 0.20, "cancelled": 0.10}}}
					],
					"primary_key": ["id"],
					"row_count": 100
				}
			}
		}`

		// Test with seed1
		coordinator1 := pipeline.NewCoordinator()
		coordinator1.RegisterBasicGenerators()

		output1 := new(bytes.Buffer)
		err := coordinator1.Execute(strings.NewReader(schemaJSON), output1, 33333)
		require.NoError(t, err)

		result1 := output1.String()
		completed1 := strings.Count(result1, "'completed'")
		pending1 := strings.Count(result1, "'pending'")
		cancelled1 := strings.Count(result1, "'cancelled'")

		// Test with seed2
		coordinator2 := pipeline.NewCoordinator()
		coordinator2.RegisterBasicGenerators()

		output2 := new(bytes.Buffer)
		err = coordinator2.Execute(strings.NewReader(schemaJSON), output2, 44444)
		require.NoError(t, err)

		result2 := output2.String()
		completed2 := strings.Count(result2, "'completed'")
		pending2 := strings.Count(result2, "'pending'")
		cancelled2 := strings.Count(result2, "'cancelled'")

		// Different seeds produce different data
		assert.NotEqual(t, result1, result2, "Different seeds should produce different data")

		// But distributions should be similar (within ±15% tolerance for 100 samples)
		assert.InDelta(t, 70, completed1, 15, "Seed 1: completed should be ~70% ±15%")
		assert.InDelta(t, 20, pending1, 15, "Seed 1: pending should be ~20% ±15%")
		assert.InDelta(t, 10, cancelled1, 15, "Seed 1: cancelled should be ~10% ±15%")

		assert.InDelta(t, 70, completed2, 15, "Seed 2: completed should be ~70% ±15%")
		assert.InDelta(t, 20, pending2, 15, "Seed 2: pending should be ~20% ±15%")
		assert.InDelta(t, 10, cancelled2, 15, "Seed 2: cancelled should be ~10% ±15%")

		// Total should be 100
		assert.Equal(t, 100, completed1+pending1+cancelled1, "Seed 1 total should be 100")
		assert.Equal(t, 100, completed2+pending2+cancelled2, "Seed 2 total should be 100")
	})

	t.Run("zero seed is valid and produces consistent output", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"test": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "value", "type": "integer"}
					],
					"primary_key": ["id"],
					"row_count": 5
				}
			}
		}`

		coordinator1 := pipeline.NewCoordinator()
		coordinator1.RegisterBasicGenerators()

		output1 := new(bytes.Buffer)
		err := coordinator1.Execute(strings.NewReader(schemaJSON), output1, 0)
		require.NoError(t, err)

		coordinator2 := pipeline.NewCoordinator()
		coordinator2.RegisterBasicGenerators()

		output2 := new(bytes.Buffer)
		err = coordinator2.Execute(strings.NewReader(schemaJSON), output2, 0)
		require.NoError(t, err)

		// Zero seed is valid and produces consistent output
		assert.Equal(t, output1.String(), output2.String(), "Zero seed should be valid and produce consistent output")
	})

	t.Run("negative seed is valid and produces consistent output", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"test": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "value", "type": "integer"}
					],
					"primary_key": ["id"],
					"row_count": 5
				}
			}
		}`

		coordinator1 := pipeline.NewCoordinator()
		coordinator1.RegisterBasicGenerators()

		output1 := new(bytes.Buffer)
		err := coordinator1.Execute(strings.NewReader(schemaJSON), output1, -12345)
		require.NoError(t, err)

		coordinator2 := pipeline.NewCoordinator()
		coordinator2.RegisterBasicGenerators()

		output2 := new(bytes.Buffer)
		err = coordinator2.Execute(strings.NewReader(schemaJSON), output2, -12345)
		require.NoError(t, err)

		// Negative seed is valid and produces consistent output
		assert.Equal(t, output1.String(), output2.String(), "Negative seed should be valid and produce consistent output")
	})
}