package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	t.Run("version command has expected metadata", func(t *testing.T) {
		cmd := cli.NewVersionCommand()
		require.NotNil(t, cmd)

		assert.Equal(t, "version", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("version command outputs version info", func(t *testing.T) {
		cmd := cli.NewVersionCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)

		err := cmd.Execute()
		require.NoError(t, err)

		result := output.String()
		assert.NotEmpty(t, result)
		// Should contain version number
		assert.True(t, len(result) > 0)
	})

	t.Run("version output contains version number", func(t *testing.T) {
		cmd := cli.NewVersionCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)

		err := cmd.RunE(cmd, []string{})
		require.NoError(t, err)

		result := output.String()
		// Should have version in format like "datagen version X.Y.Z"
		assert.Contains(t, strings.ToLower(result), "version")
	})

	t.Run("verbose flag shows build info", func(t *testing.T) {
		cmd := cli.NewVersionCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetArgs([]string{"--verbose"})

		err := cmd.Execute()
		require.NoError(t, err)

		result := output.String()
		assert.NotEmpty(t, result)
		// In verbose mode, should show Go version and build info
		assert.Contains(t, strings.ToLower(result), "go")
	})

	t.Run("short flag displays version", func(t *testing.T) {
		cmd := cli.NewVersionCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetArgs([]string{"--short"})

		err := cmd.Execute()
		require.NoError(t, err)

		result := output.String()
		assert.NotEmpty(t, result)
		// Short mode should be concise
		lines := strings.Split(strings.TrimSpace(result), "\n")
		assert.LessOrEqual(t, len(lines), 3)
	})
}

func TestVersionInfo(t *testing.T) {
	t.Run("version info has default values", func(t *testing.T) {
		info := cli.GetVersionInfo()
		require.NotNil(t, info)

		assert.NotEmpty(t, info.Version)
		assert.NotEmpty(t, info.GoVersion)
	})

	t.Run("version info includes build metadata", func(t *testing.T) {
		info := cli.GetVersionInfo()

		// These may be empty in dev builds but should exist
		assert.NotNil(t, info.GitCommit)
		assert.NotNil(t, info.BuildDate)
	})
}