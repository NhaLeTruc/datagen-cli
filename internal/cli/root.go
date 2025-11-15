package cli

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

// NewRootCommand creates and returns the root command
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datagen",
		Short: "Generate PostgreSQL dump files from JSON schemas",
		Long: `datagen is a CLI tool that generates realistic PostgreSQL dump files
from simple JSON schema definitions.

It supports generating test data with semantic types, custom patterns,
and deterministic seeds for reproducible datasets.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}

	// Global flags
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is .datagen.yaml)")
	cmd.Flags().BoolP("version", "v", false, "show version information")

	// Add subcommands
	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(NewGenerateCommand())
	cmd.AddCommand(NewValidateCommand())
	cmd.AddCommand(newTemplateCmd())

	return cmd
}

// NewRootCmd is an alias for NewRootCommand for backward compatibility
func NewRootCmd() *cobra.Command {
	return NewRootCommand()
}