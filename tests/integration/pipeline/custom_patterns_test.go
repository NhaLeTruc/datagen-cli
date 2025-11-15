package pipeline_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomPatterns(t *testing.T) {
	t.Run("weighted enum distribution in real schema", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"orders": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "status", "type": "varchar(20)", "generator_config": {
							"type": "weighted_enum",
							"weights": {
								"completed": 0.70,
								"pending": 0.20,
								"cancelled": 0.10
							}
						}}
					],
					"row_count": 100
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterCustomGenerators()

		output := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Count occurrences of each status
		completedCount := strings.Count(result, "'completed'")
		pendingCount := strings.Count(result, "'pending'")
		cancelledCount := strings.Count(result, "'cancelled'")

		// Verify distribution roughly matches weights (±15% tolerance for 100 samples)
		assert.InDelta(t, 70, completedCount, 15, "completed should be ~70%")
		assert.InDelta(t, 20, pendingCount, 10, "pending should be ~20%")
		assert.InDelta(t, 10, cancelledCount, 10, "cancelled should be ~10%")
	})

	t.Run("pattern generator creates valid formats", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"products": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "sku", "type": "varchar(50)", "generator_config": {
							"type": "pattern",
							"pattern": "[A-Z]{3}-\\d{4}"
						}}
					],
					"row_count": 5
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterCustomGenerators()

		output := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Verify SKU pattern appears
		assert.Regexp(t, `'[A-Z]{3}-\d{4}'`, result)
	})

	t.Run("template generator creates sequential IDs", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"invoices": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "invoice_number", "type": "varchar(50)", "generator_config": {
							"type": "template",
							"template": "INV-{{year}}-{{seq:5}}"
						}}
					],
					"row_count": 3
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterCustomGenerators()

		output := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Verify invoice numbers with year and sequence
		assert.Regexp(t, `'INV-\d{4}-00001'`, result)
		assert.Regexp(t, `'INV-\d{4}-00002'`, result)
		assert.Regexp(t, `'INV-\d{4}-00003'`, result)
	})

	t.Run("integer range generator respects bounds", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"ratings": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "score", "type": "integer", "generator_config": {
							"type": "integer_range",
							"min": 1,
							"max": 5
						}}
					],
					"row_count": 10
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterCustomGenerators()

		output := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Verify all scores are 1-5
		assert.Contains(t, result, "INSERT INTO ratings")
		// Scores should not contain 0 or 6+
		assert.NotContains(t, result, ", 0)")
		assert.NotContains(t, result, ", 6)")
	})

	t.Run("timeseries generator creates sequential timestamps", func(t *testing.T) {
		schemaJSON := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"events": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "event_time", "type": "timestamp", "generator_config": {
							"type": "timeseries",
							"start": "2024-01-01T00:00:00Z",
							"end": "2024-01-02T00:00:00Z",
							"interval": "1h",
							"pattern": "uniform"
						}}
					],
					"row_count": 5
				}
			}
		}`

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterCustomGenerators()

		output := new(bytes.Buffer)
		err := coordinator.Execute(strings.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Verify timestamps from 2024-01-01
		assert.Contains(t, result, "2024-01-01")
		assert.Contains(t, result, "INSERT INTO events")
	})
}