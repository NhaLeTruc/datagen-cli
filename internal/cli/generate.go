package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/NhaLeTruc/datagen-cli/internal/templates"
	"github.com/spf13/cobra"
)

// NewGenerateCommand creates the generate command
func NewGenerateCommand() *cobra.Command {
	var (
		inputFile      string
		outputFile     string
		seed           int64
		templateName   string
		templateParams []string
		format         string
		jobs           int
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate PostgreSQL dump from JSON schema",
		Long: `Generate a PostgreSQL dump file from a JSON schema definition.

The schema defines tables, columns, and data generation rules.
Output is a SQL file compatible with PostgreSQL.

You can use a pre-built template with --template or provide a custom schema with --input.`,
		Example: `  # Generate from custom schema (SQL format by default)
  datagen generate -i schema.json -o dump.sql

  # Generate from template
  datagen generate --template ecommerce -o dump.sql

  # Generate with COPY format
  datagen generate -i schema.json -o dump.sql --format copy

  # Generate from template with custom parameters
  datagen generate --template saas --param tenants=500 -o dump.sql

  # Generate with deterministic seed
  datagen generate -i schema.json -o dump.sql --seed 12345

  # Generate with parallel workers
  datagen generate -i schema.json -o dump.sql --jobs 8`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate flags: must have either input or template, but not both
			if templateName != "" && inputFile != "" {
				return fmt.Errorf("cannot specify both --input and --template")
			}
			if templateName == "" && inputFile == "" {
				return fmt.Errorf("must specify either --input or --template")
			}

			// Validate format
			if format == "" {
				format = "sql" // Default format
			}
			validFormats := map[string]bool{"sql": true, "copy": true}
			if !validFormats[format] {
				return fmt.Errorf("invalid format %q, must be one of: sql, copy", format)
			}

			// Open input (stdin, file, or template)
			var input *os.File
			var err error

			if templateName != "" {
				// Load template
				tmpl, err := templates.Get(templateName)
				if err != nil {
					return fmt.Errorf("failed to get template: %w", err)
				}

				// Parse and apply parameters
				params, err := parseTemplateParams(templateParams)
				if err != nil {
					return fmt.Errorf("failed to parse template parameters: %w", err)
				}

				if len(params) > 0 {
					if err := templates.ApplyParameters(tmpl, params); err != nil {
						return fmt.Errorf("failed to apply template parameters: %w", err)
					}
				}

				// Convert template schema to JSON
				schemaJSON, err := json.MarshalIndent(tmpl.Schema, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal template schema: %w", err)
				}

				// Create a bytes reader as input
				input = os.NewFile(0, "template")
				defer input.Close()

				// We need to create a temp file for the schema
				tmpFile, err := os.CreateTemp("", "datagen-template-*.json")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()

				if _, err := tmpFile.Write(schemaJSON); err != nil {
					return fmt.Errorf("failed to write template schema: %w", err)
				}

				if _, err := tmpFile.Seek(0, 0); err != nil {
					return fmt.Errorf("failed to seek temp file: %w", err)
				}

				input = tmpFile
			} else if inputFile == "" || inputFile == "-" {
				input = os.Stdin
			} else {
				input, err = os.Open(inputFile)
				if err != nil {
					return fmt.Errorf("failed to open input file: %w", err)
				}
				defer input.Close()
			}

			// Open output (stdout or file)
			var output *os.File
			if outputFile == "" || outputFile == "-" {
				output = os.Stdout
			} else {
				output, err = os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer output.Close()
			}

			// Determine number of workers to use
			// Priority: flag > config > default (4)
			workerCount := jobs
			if workerCount == 0 {
				// Use config value if flag not set
				if AppConfig != nil {
					workerCount = AppConfig.Workers
				} else {
					workerCount = 4 // Default fallback
				}
			}

			// Validate worker count
			if workerCount < 1 {
				workerCount = 1
			} else if workerCount > 100 {
				return fmt.Errorf("jobs must be between 1 and 100, got %d", workerCount)
			}

			// Log configuration
			if AppConfig != nil && AppConfig.Verbose {
				LogDebugf("Using %d worker(s) for data generation", workerCount)
			}

			// Create coordinator and register generators
			coordinator := pipeline.NewCoordinator()
			coordinator.RegisterBasicGenerators()
			coordinator.RegisterSemanticGenerators()

			// Execute pipeline with format
			// Note: Worker pool support will be added in future enhancement
			if err := coordinator.ExecuteWithFormat(input, output, seed, format); err != nil {
				return fmt.Errorf("generation failed: %w", err)
			}

			// Print success message to stderr (so it doesn't mix with output)
			if outputFile != "" && outputFile != "-" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Successfully generated dump to %s\n", outputFile)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "input schema file (default: stdin)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "output SQL file (default: stdout)")
	cmd.Flags().Int64VarP(&seed, "seed", "s", 0, "random seed for deterministic generation")
	cmd.Flags().StringVarP(&format, "format", "f", "sql", "output format: sql (INSERT statements), copy (COPY format)")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", 0, "number of parallel workers (default: from config or 4)")
	cmd.Flags().StringVar(&templateName, "template", "", "use pre-built template (ecommerce, saas, healthcare, finance)")
	cmd.Flags().StringArrayVar(&templateParams, "param", []string{}, "override template parameters (format: key=value)")

	return cmd
}