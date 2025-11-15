package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateListCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantErr       bool
		wantContains  []string
		wantNotEmpty  bool
	}{
		{
			name:         "list all templates",
			args:         []string{"template", "list"},
			wantErr:      false,
			wantContains: []string{"ecommerce", "saas", "healthcare", "finance"},
			wantNotEmpty: true,
		},
		{
			name:         "list with --help",
			args:         []string{"template", "list", "--help"},
			wantErr:      false,
			wantContains: []string{"List", "template"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewRootCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := out.String()

			if tt.wantNotEmpty {
				assert.NotEmpty(t, output, "Output should not be empty")
			}

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "Output should contain %q", want)
			}
		})
	}
}

func TestTemplateShowCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name:    "show ecommerce template",
			args:    []string{"template", "show", "ecommerce"},
			wantErr: false,
			wantContains: []string{
				"ecommerce",
				"E-commerce",
				"products",
				"customers",
				"orders",
			},
		},
		{
			name:    "show saas template",
			args:    []string{"template", "show", "saas"},
			wantErr: false,
			wantContains: []string{
				"saas",
				"tenants",
				"users",
				"subscriptions",
			},
		},
		{
			name:    "show healthcare template",
			args:    []string{"template", "show", "healthcare"},
			wantErr: false,
			wantContains: []string{
				"healthcare",
				"patients",
				"doctors",
				"appointments",
			},
		},
		{
			name:    "show finance template",
			args:    []string{"template", "show", "finance"},
			wantErr: false,
			wantContains: []string{
				"finance",
				"accounts",
				"transactions",
				"customers",
			},
		},
		{
			name:    "show non-existent template",
			args:    []string{"template", "show", "nonexistent"},
			wantErr: true,
		},
		{
			name:    "show without template name",
			args:    []string{"template", "show"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewRootCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := out.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "Output should contain %q", want)
			}
		})
	}
}

func TestTemplateExportCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name:    "export ecommerce template to stdout",
			args:    []string{"template", "export", "ecommerce"},
			wantErr: false,
			wantContains: []string{
				"\"version\"",
				"\"database\"",
				"\"tables\"",
			},
		},
		{
			name:    "export saas template to stdout",
			args:    []string{"template", "export", "saas"},
			wantErr: false,
			wantContains: []string{
				"\"version\"",
				"\"tables\"",
			},
		},
		{
			name:    "export non-existent template",
			args:    []string{"template", "export", "nonexistent"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewRootCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := out.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "Output should contain %q", want)
			}

			// Verify valid JSON
			assert.True(t, strings.HasPrefix(strings.TrimSpace(output), "{"), "Output should be valid JSON")
		})
	}
}

func TestTemplateParameterOverride(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name: "override row count parameter",
			args: []string{"template", "export", "ecommerce", "--param", "customers=500"},
			wantErr: false,
			wantContains: []string{
				"\"version\"",
				"\"customers\"",
			},
		},
		{
			name:    "invalid parameter format",
			args:    []string{"template", "export", "ecommerce", "--param", "invalid"},
			wantErr: true,
		},
		{
			name:    "unknown parameter",
			args:    []string{"template", "export", "ecommerce", "--param", "unknown=100"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.NewRootCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := out.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "Output should contain %q", want)
			}
		})
	}
}
