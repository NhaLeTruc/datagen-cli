package cli

import (
	"fmt"
	"os"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/spf13/cobra"
)

// NewGenerateCommand creates the generate command
func NewGenerateCommand() *cobra.Command {
	var (
		inputFile  string
		outputFile string
		seed       int64
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate PostgreSQL dump from JSON schema",
		Long: `Generate a PostgreSQL dump file from a JSON schema definition.

The schema defines tables, columns, and data generation rules.
Output is a SQL file compatible with PostgreSQL.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open input (stdin or file)
			var input *os.File
			var err error
			if inputFile == "" || inputFile == "-" {
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

			// Create coordinator and register generators
			coordinator := pipeline.NewCoordinator()
			coordinator.RegisterBasicGenerators()

			// Execute pipeline
			if err := coordinator.Execute(input, output, seed); err != nil {
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

	return cmd
}