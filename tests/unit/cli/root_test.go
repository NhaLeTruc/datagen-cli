package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	t.Run("root command has expected metadata", func(t *testing.T) {
		cmd := cli.NewRootCommand()
		require.NotNil(t, cmd)

		assert.Equal(t, "datagen", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("help flag displays usage", func(t *testing.T) {
		cmd := cli.NewRootCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"--help"})

		err := cmd.Execute()
		require.NoError(t, err)

		result := output.String()
		assert.Contains(t, result, "datagen")
		assert.Contains(t, result, "Usage:")
		assert.Contains(t, result, "Flags:")
	})

	t.Run("version flag is available", func(t *testing.T) {
		cmd := cli.NewRootCommand()

		versionFlag := cmd.Flags().Lookup("version")
		require.NotNil(t, versionFlag)
		assert.Equal(t, "bool", versionFlag.Value.Type())
	})

	t.Run("config flag is available", func(t *testing.T) {
		cmd := cli.NewRootCommand()

		configFlag := cmd.PersistentFlags().Lookup("config")
		require.NotNil(t, configFlag)
		assert.Equal(t, "string", configFlag.Value.Type())
	})

	t.Run("running without subcommand shows help", func(t *testing.T) {
		cmd := cli.NewRootCommand()
		output := new(bytes.Buffer)
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		// Should not error, just show help
		require.NoError(t, err)

		result := output.String()
		assert.Contains(t, strings.ToLower(result), "usage")
	})
}

func TestRootCommandFlags(t *testing.T) {
	t.Run("config flag accepts file path", func(t *testing.T) {
		cmd := cli.NewRootCommand()
		cmd.SetArgs([]string{"--config", "/path/to/config.yaml"})

		err := cmd.ParseFlags([]string{"--config", "/path/to/config.yaml"})
		require.NoError(t, err)

		configPath, err := cmd.Flags().GetString("config")
		require.NoError(t, err)
		assert.Equal(t, "/path/to/config.yaml", configPath)
	})

	t.Run("short flag -c works for config", func(t *testing.T) {
		cmd := cli.NewRootCommand()

		err := cmd.ParseFlags([]string{"-c", "config.yaml"})
		require.NoError(t, err)

		configPath, err := cmd.Flags().GetString("config")
		require.NoError(t, err)
		assert.Equal(t, "config.yaml", configPath)
	})

	t.Run("version flag is boolean", func(t *testing.T) {
		cmd := cli.NewRootCommand()

		err := cmd.ParseFlags([]string{"--version"})
		require.NoError(t, err)

		version, err := cmd.Flags().GetBool("version")
		require.NoError(t, err)
		assert.True(t, version)
	})

	t.Run("short flag -v works for version", func(t *testing.T) {
		cmd := cli.NewRootCommand()

		err := cmd.ParseFlags([]string{"-v"})
		require.NoError(t, err)

		version, err := cmd.Flags().GetBool("version")
		require.NoError(t, err)
		assert.True(t, version)
	})
}