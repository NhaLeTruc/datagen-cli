package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommand(t *testing.T) {
	t.Run("validate command has expected metadata", func(t *testing.T) {
		cmd := cli.NewValidateCommand()
		require.NotNil(t, cmd)

		assert.Equal(t, "validate", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("validate valid schema from stdin", func(t *testing.T) {
		validSchema := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "email", "type": "varchar(255)"}
					],
					"primary_key": ["id"],
					"row_count": 100
				}
			}
		}`

		cmd := cli.NewValidateCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetIn(strings.NewReader(validSchema))
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		require.NoError(t, err)

		result := output.String()
		assert.Contains(t, strings.ToLower(result), "valid")
		assert.NotContains(t, strings.ToLower(result), "error")
	})

	t.Run("validate invalid schema shows errors", func(t *testing.T) {
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
					"row_count": 100
				}
			}
		}`

		cmd := cli.NewValidateCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetIn(strings.NewReader(invalidSchema))
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		require.Error(t, err)

		result := output.String()
		assert.Contains(t, strings.ToLower(result), "error")
		assert.Contains(t, result, "nonexistent")
	})

	t.Run("validate schema with missing row_count", func(t *testing.T) {
		invalidSchema := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"}
					],
					"row_count": 0
				}
			}
		}`

		cmd := cli.NewValidateCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetIn(strings.NewReader(invalidSchema))
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		require.Error(t, err)

		result := output.String()
		assert.Contains(t, result, "row_count")
		assert.Contains(t, result, "greater than 0")
	})

	t.Run("validate schema with circular dependencies", func(t *testing.T) {
		circularSchema := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"table_a": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "b_id", "type": "integer"}
					],
					"foreign_keys": [{
						"columns": ["b_id"],
						"referenced_table": "table_b",
						"referenced_columns": ["id"]
					}],
					"row_count": 10
				},
				"table_b": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "a_id", "type": "integer"}
					],
					"foreign_keys": [{
						"columns": ["a_id"],
						"referenced_table": "table_a",
						"referenced_columns": ["id"]
					}],
					"row_count": 10
				}
			}
		}`

		cmd := cli.NewValidateCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetIn(strings.NewReader(circularSchema))
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		require.Error(t, err)

		result := output.String()
		assert.Contains(t, strings.ToLower(result), "circular")
	})

	t.Run("validate with json output format", func(t *testing.T) {
		validSchema := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"}
					],
					"row_count": 10
				}
			}
		}`

		cmd := cli.NewValidateCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetIn(strings.NewReader(validSchema))
		cmd.SetArgs([]string{"--format", "json"})

		err := cmd.Execute()
		require.NoError(t, err)

		result := output.String()
		assert.Contains(t, result, "{")
		assert.Contains(t, result, "valid")
	})

	t.Run("validate with json output format shows errors", func(t *testing.T) {
		invalidSchema := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "invalidtype"}
					],
					"row_count": 10
				}
			}
		}`

		cmd := cli.NewValidateCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetIn(strings.NewReader(invalidSchema))
		cmd.SetArgs([]string{"--format", "json"})

		err := cmd.Execute()
		require.Error(t, err)

		result := output.String()
		assert.Contains(t, result, "{")
		assert.Contains(t, strings.ToLower(result), "error")
	})

	t.Run("validate schema from file", func(t *testing.T) {
		// This test is skipped as it requires file I/O
		// In practice, file validation is tested via integration tests
		t.Skip("File-based validation tested in integration tests")
	})

	t.Run("validate command has input flag", func(t *testing.T) {
		cmd := cli.NewValidateCommand()
		require.NotNil(t, cmd.Flags().Lookup("input"))
	})

	t.Run("validate command has format flag", func(t *testing.T) {
		cmd := cli.NewValidateCommand()
		require.NotNil(t, cmd.Flags().Lookup("format"))
	})
}
