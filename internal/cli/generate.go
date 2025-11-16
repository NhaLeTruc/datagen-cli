package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NhaLeTruc/datagen-cli/internal/pgdump"
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
		validateOutput bool
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
  datagen generate -i schema.json -o dump.sql --jobs 8

  # Generate with SQL validation
  datagen generate -i schema.json -o dump.sql --validate-output`,
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

			// Validate output if requested (only for file output, not stdout)
			if validateOutput {
				if outputFile == "" || outputFile == "-" {
					LogWarn("Cannot validate output when writing to stdout (--validate-output requires --output <file>)")
				} else {
					if err := validateGeneratedSQL(outputFile); err != nil {
						return fmt.Errorf("output validation failed: %w", err)
					}
					if AppConfig != nil && AppConfig.Verbose {
						LogInfo("Output validation successful")
					}
				}
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
	cmd.Flags().BoolVar(&validateOutput, "validate-output", false, "validate generated SQL syntax using PostgreSQL parser (requires --output <file>)")
	cmd.Flags().StringVar(&templateName, "template", "", "use pre-built template (ecommerce, saas, healthcare, finance)")
	cmd.Flags().StringArrayVar(&templateParams, "param", []string{}, "override template parameters (format: key=value)")

	return cmd
}

// validateGeneratedSQL validates the SQL in the generated file
func validateGeneratedSQL(filePath string) error {
	// Read the generated SQL file
	sqlContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read output file: %w", err)
	}

	// Log validation start
	if AppConfig != nil && AppConfig.Verbose {
		LogInfof("Validating generated SQL in %s", filePath)
	}

	// Pre-process SQL to remove psql meta-commands and filter out COPY data
	cleanedSQL, err := preprocessSQLForValidation(string(sqlContent))
	if err != nil {
		return fmt.Errorf("failed to preprocess SQL: %w", err)
	}

	// Validate the SQL
	result := pgdump.ValidateSQLDetailed(cleanedSQL)

	// Check if validation succeeded
	if !result.Valid {
		// Log errors
		LogError("SQL validation failed", result.Errors[0])
		for i, err := range result.Errors {
			if i > 0 { // First error already logged above
				LogErrorf("Additional error %d: %v", i, err)
			}
		}
		return fmt.Errorf("generated SQL contains %d syntax error(s)", len(result.Errors))
	}

	// Log success with details
	if AppConfig != nil && AppConfig.Verbose {
		LogInfof("SQL validation passed: %d statement(s) validated", result.StatementCount)
	}

	return nil
}

// preprocessSQLForValidation removes psql meta-commands and COPY data sections
// that cannot be parsed by pg_query (which only parses actual SQL syntax)
func preprocessSQLForValidation(sql string) (string, error) {
	var cleaned bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(sql))
	inCopyData := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip psql meta-commands (start with backslash)
		if strings.HasPrefix(trimmedLine, "\\") {
			if AppConfig != nil && AppConfig.Verbose {
				LogDebugf("Skipping psql meta-command: %s", trimmedLine)
			}
			continue
		}

		// Detect COPY FROM stdin (start of COPY data section)
		if strings.Contains(strings.ToUpper(trimmedLine), "COPY") && strings.Contains(strings.ToUpper(line), "FROM STDIN") {
			// Keep the COPY command itself
			cleaned.WriteString(line)
			cleaned.WriteString("\n")
			inCopyData = true
			continue
		}

		// Detect end of COPY data (\.)
		if inCopyData && trimmedLine == "\\." {
			inCopyData = false
			continue
		}

		// Skip COPY data lines
		if inCopyData {
			continue
		}

		// Keep all other lines (including comments and SQL)
		cleaned.WriteString(line)
		cleaned.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading SQL: %w", err)
	}

	return cleaned.String(), nil
}