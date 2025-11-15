package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version information (set via -ldflags during build)
	Version   = "dev"
	GitCommit = ""
	BuildDate = ""
)

// VersionInfo holds version metadata
type VersionInfo struct {
	Version   string
	GitCommit string
	BuildDate string
	GoVersion string
}

// GetVersionInfo returns the current version information
func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
	}
}

// NewVersionCommand creates and returns the version command
func NewVersionCommand() *cobra.Command {
	var verbose bool
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the version number, build date, and other build information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := GetVersionInfo()

			if short {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", info.Version)
				return nil
			}

			if verbose {
				fmt.Fprintf(cmd.OutOrStdout(), "datagen version %s\n", info.Version)
				fmt.Fprintf(cmd.OutOrStdout(), "  Go version: %s\n", info.GoVersion)
				if info.GitCommit != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  Git commit: %s\n", info.GitCommit)
				}
				if info.BuildDate != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  Built: %s\n", info.BuildDate)
				}
				return nil
			}

			// Default output
			fmt.Fprintf(cmd.OutOrStdout(), "datagen version %s\n", info.Version)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "", false, "show detailed version information")
	cmd.Flags().BoolVarP(&short, "short", "s", false, "show only version number")

	return cmd
}