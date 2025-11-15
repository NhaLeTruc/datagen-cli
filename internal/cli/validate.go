package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/NhaLeTruc/datagen-cli/internal/schema"
	"github.com/spf13/cobra"
)

// NewValidateCommand creates the validate command
func NewValidateCommand() *cobra.Command {
	var inputFile string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a JSON schema without generating data",
		Long: `Validate a JSON schema file for correctness.

This command checks:
  - Schema structure and required fields
  - Table and column definitions
  - Foreign key relationships
  - Circular dependencies
  - PostgreSQL type validity

No data is generated - only the schema is validated.

Examples:
  # Validate schema from stdin
  cat schema.json | datagen validate

  # Validate schema from file
  datagen validate --input schema.json

  # Get JSON output
  datagen validate --input schema.json --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get input reader (stdin or file)
			var input io.Reader
			if inputFile != "" {
				f, err := os.Open(inputFile)
				if err != nil {
					return fmt.Errorf("failed to open input file: %w", err)
				}
				defer f.Close()
				input = f
			} else {
				input = cmd.InOrStdin()
			}

			// Parse schema
			sch, err := schema.Parse(input)
			if err != nil {
				return formatValidationOutput(cmd, nil, []error{err}, outputFormat)
			}

			// Validate schema
			errs := schema.Validate(sch)
			if len(errs) > 0 {
				return formatValidationOutput(cmd, sch, errs, outputFormat)
			}

			// Success
			if outputFormat == "json" {
				return outputValidationJSON(cmd, sch, nil)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Schema is valid\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Database: %s\n", sch.Database.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "  Tables: %d\n", len(sch.Tables))

			// Calculate total row count
			totalRows := 0
			for _, table := range sch.Tables {
				totalRows += table.RowCount
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Total rows to generate: %d\n", totalRows)

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input schema file (default: stdin)")
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text, json")

	return cmd
}

// formatValidationOutput formats validation errors for output
func formatValidationOutput(cmd *cobra.Command, sch *schema.Schema, errs []error, format string) error {
	if format == "json" {
		return outputValidationJSON(cmd, sch, errs)
	}

	// Text format
	fmt.Fprintf(cmd.OutOrStdout(), "✗ Schema validation failed\n\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Errors found:\n")
	for i, err := range errs {
		fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s\n", i+1, err.Error())
	}

	return fmt.Errorf("validation failed with %d error(s)", len(errs))
}

// outputValidationJSON outputs validation results as JSON
func outputValidationJSON(cmd *cobra.Command, sch *schema.Schema, errs []error) error {
	result := struct {
		Valid    bool     `json:"valid"`
		Errors   []string `json:"errors,omitempty"`
		Database string   `json:"database,omitempty"`
		Tables   int      `json:"tables,omitempty"`
	}{
		Valid: len(errs) == 0,
	}

	if len(errs) > 0 {
		result.Errors = make([]string, len(errs))
		for i, err := range errs {
			result.Errors[i] = err.Error()
		}
	}

	if sch != nil {
		result.Database = sch.Database.Name
		result.Tables = len(sch.Tables)
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON output: %w", err)
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}
