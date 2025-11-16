package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	AppConfig *Config
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize configuration
			cfg, err := InitConfig(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			AppConfig = cfg

			// Override verbose from flag if set
			if cmd.Flags().Changed("verbose") {
				AppConfig.Verbose = verbose
				// Set log level to debug when verbose
				if verbose {
					AppConfig.LogLevel = "debug"
				}
			}

			// Initialize logging system
			if err := InitLogging(AppConfig); err != nil {
				return fmt.Errorf("failed to initialize logging: %w", err)
			}

			// Log configuration load
			if cfgFile := GetConfigFilePath(); cfgFile != "" {
				LogConfigLoad(cfgFile, true)
			} else {
				LogConfigLoad("defaults", true)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is provided, show help
			return cmd.Help()
		},
	}

	// Global flags
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is .datagen.yaml)")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "", false, "enable verbose output")
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