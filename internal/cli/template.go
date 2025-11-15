package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/NhaLeTruc/datagen-cli/internal/templates"
	"github.com/spf13/cobra"
)

var (
	templateParams []string
)

// newTemplateCmd creates the template command
func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage pre-built schema templates",
		Long: `Manage pre-built schema templates for common use cases.

Templates provide ready-to-use database schemas for common scenarios like
e-commerce, SaaS applications, healthcare systems, and financial services.

Available subcommands:
  list   - List all available templates
  show   - Show template details
  export - Export template as JSON schema`,
		Example: `  # List all templates
  datagen template list

  # Show details for ecommerce template
  datagen template show ecommerce

  # Export template to file
  datagen template export ecommerce > schema.json

  # Export with custom parameters
  datagen template export ecommerce --param customers=5000 --param orders=10000`,
	}

	cmd.AddCommand(newTemplateListCmd())
	cmd.AddCommand(newTemplateShowCmd())
	cmd.AddCommand(newTemplateExportCmd())

	return cmd
}

// newTemplateListCmd creates the list subcommand
func newTemplateListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available templates",
		Long:  `List all available pre-built schema templates with their descriptions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tmplList := templates.List()

			// Sort by name for consistent output
			sort.Slice(tmplList, func(i, j int) bool {
				return tmplList[i].Name < tmplList[j].Name
			})

			cmd.Println("Available Templates:")
			cmd.Println()

			for _, tmpl := range tmplList {
				cmd.Printf("  %-15s %s\n", tmpl.Name, tmpl.Description)
			}

			cmd.Println()
			cmd.Println("Use 'datagen template show <name>' to see template details")
			cmd.Println("Use 'datagen template export <name>' to export template as JSON")

			return nil
		},
	}
}

// newTemplateShowCmd creates the show subcommand
func newTemplateShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <template-name>",
		Short: "Show template details",
		Long:  `Show detailed information about a specific template including tables and parameters.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]

			tmpl, err := templates.Get(templateName)
			if err != nil {
				return fmt.Errorf("failed to get template: %w", err)
			}

			cmd.Printf("Template: %s\n", tmpl.Name)
			cmd.Printf("Description: %s\n", tmpl.Description)
			cmd.Printf("Category: %s\n", tmpl.Category)
			cmd.Println()

			if tmpl.Schema != nil {
				cmd.Println("Tables:")
				// Sort table names for consistent output
				tableNames := make([]string, 0, len(tmpl.Schema.Tables))
				for name := range tmpl.Schema.Tables {
					tableNames = append(tableNames, name)
				}
				sort.Strings(tableNames)

				for _, name := range tableNames {
					table := tmpl.Schema.Tables[name]
					cmd.Printf("  %-20s %d rows, %d columns\n", name, table.RowCount, len(table.Columns))
				}
				cmd.Println()
			}

			if len(tmpl.Parameters) > 0 {
				cmd.Println("Parameters:")
				// Sort parameter names for consistent output
				paramNames := make([]string, 0, len(tmpl.Parameters))
				for name := range tmpl.Parameters {
					paramNames = append(paramNames, name)
				}
				sort.Strings(paramNames)

				for _, name := range paramNames {
					param := tmpl.Parameters[name]
					cmd.Printf("  %-15s %s (default: %v)\n", param.Name, param.Description, param.Default)
				}
				cmd.Println()
				cmd.Println("Use --param flag to override parameters when exporting")
			}

			return nil
		},
	}
}

// newTemplateExportCmd creates the export subcommand
func newTemplateExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <template-name>",
		Short: "Export template as JSON schema",
		Long: `Export a template as a JSON schema that can be used with the generate command.

Parameters can be customized using the --param flag.`,
		Example: `  # Export to stdout
  datagen template export ecommerce

  # Export to file
  datagen template export ecommerce > schema.json

  # Override parameters
  datagen template export ecommerce --param customers=5000 --param products=1000`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]

			tmpl, err := templates.Get(templateName)
			if err != nil {
				return fmt.Errorf("failed to get template: %w", err)
			}

			// Parse and apply parameters
			params, err := parseTemplateParams(templateParams)
			if err != nil {
				return fmt.Errorf("failed to parse parameters: %w", err)
			}

			if len(params) > 0 {
				if err := templates.ApplyParameters(tmpl, params); err != nil {
					return fmt.Errorf("failed to apply parameters: %w", err)
				}
			}

			// Marshal schema to JSON
			data, err := json.MarshalIndent(tmpl.Schema, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal schema: %w", err)
			}

			cmd.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&templateParams, "param", []string{}, "Override template parameters (format: key=value)")

	return cmd
}

// parseTemplateParams parses --param flags into a map
func parseTemplateParams(params []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid parameter format %q, expected key=value", param)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Try to parse as integer first
		if intVal, err := strconv.Atoi(value); err == nil {
			result[key] = intVal
			continue
		}

		// Try to parse as boolean
		if boolVal, err := strconv.ParseBool(value); err == nil {
			result[key] = boolVal
			continue
		}

		// Default to string
		result[key] = value
	}

	return result, nil
}
