package templates

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/NhaLeTruc/datagen-cli/internal/schema"
)

//go:embed ecommerce.json
var ecommerceJSON []byte

//go:embed saas.json
var saasJSON []byte

//go:embed healthcare.json
var healthcareJSON []byte

//go:embed finance.json
var financeJSON []byte

// Template represents a pre-built schema template
type Template struct {
	Name        string
	Description string
	Category    string
	Schema      *schema.Schema
	Parameters  map[string]TemplateParameter
}

// TemplateParameter defines a customizable template parameter
type TemplateParameter struct {
	Name        string
	Type        string // int, string, bool, string_array
	Default     interface{}
	Description string
}

// Available templates
var templates = map[string]*Template{
	"ecommerce": {
		Name:        "ecommerce",
		Description: "E-commerce platform with products, customers, orders, and reviews",
		Category:    "business",
		Parameters: map[string]TemplateParameter{
			"categories": {
				Name:        "categories",
				Type:        "int",
				Default:     50,
				Description: "Number of category records to generate",
			},
			"customers": {
				Name:        "customers",
				Type:        "int",
				Default:     1000,
				Description: "Number of customer records to generate",
			},
			"products": {
				Name:        "products",
				Type:        "int",
				Default:     500,
				Description: "Number of product records to generate",
			},
			"orders": {
				Name:        "orders",
				Type:        "int",
				Default:     2000,
				Description: "Number of order records to generate",
			},
			"order_items": {
				Name:        "order_items",
				Type:        "int",
				Default:     5000,
				Description: "Number of order item records to generate",
			},
			"reviews": {
				Name:        "reviews",
				Type:        "int",
				Default:     1500,
				Description: "Number of review records to generate",
			},
		},
	},
	"saas": {
		Name:        "saas",
		Description: "SaaS application with multi-tenant structure, subscriptions, and usage metrics",
		Category:    "business",
		Parameters: map[string]TemplateParameter{
			"tenants": {
				Name:        "tenants",
				Type:        "int",
				Default:     200,
				Description: "Number of tenant records to generate",
			},
			"users": {
				Name:        "users",
				Type:        "int",
				Default:     2000,
				Description: "Number of user records to generate",
			},
		},
	},
	"healthcare": {
		Name:        "healthcare",
		Description: "Healthcare system with patients, doctors, appointments, and medical records",
		Category:    "healthcare",
		Parameters: map[string]TemplateParameter{
			"patients": {
				Name:        "patients",
				Type:        "int",
				Default:     5000,
				Description: "Number of patient records to generate",
			},
			"doctors": {
				Name:        "doctors",
				Type:        "int",
				Default:     200,
				Description: "Number of doctor records to generate",
			},
			"appointments": {
				Name:        "appointments",
				Type:        "int",
				Default:     10000,
				Description: "Number of appointment records to generate",
			},
		},
	},
	"finance": {
		Name:        "finance",
		Description: "Financial system with accounts, transactions, investments, and portfolios",
		Category:    "finance",
		Parameters: map[string]TemplateParameter{
			"customers": {
				Name:        "customers",
				Type:        "int",
				Default:     10000,
				Description: "Number of customer records to generate",
			},
			"accounts": {
				Name:        "accounts",
				Type:        "int",
				Default:     15000,
				Description: "Number of account records to generate",
			},
			"transactions": {
				Name:        "transactions",
				Type:        "int",
				Default:     100000,
				Description: "Number of transaction records to generate",
			},
		},
	},
}

// List returns all available templates
func List() []*Template {
	result := make([]*Template, 0, len(templates))
	for _, t := range templates {
		result = append(result, t)
	}
	return result
}

// Get returns a template by name
func Get(name string) (*Template, error) {
	t, ok := templates[name]
	if !ok {
		return nil, fmt.Errorf("template %q not found", name)
	}

	// Parse schema from embedded JSON
	var data []byte
	switch name {
	case "ecommerce":
		data = ecommerceJSON
	case "saas":
		data = saasJSON
	case "healthcare":
		data = healthcareJSON
	case "finance":
		data = financeJSON
	default:
		return nil, fmt.Errorf("template %q not found", name)
	}

	var sch schema.Schema
	if err := json.Unmarshal(data, &sch); err != nil {
		return nil, fmt.Errorf("failed to parse template %q: %w", name, err)
	}

	// Clone template and attach schema
	result := *t
	result.Schema = &sch
	return &result, nil
}

// ApplyParameters applies parameter overrides to a template schema
func ApplyParameters(tmpl *Template, params map[string]interface{}) error {
	if tmpl.Schema == nil {
		return fmt.Errorf("template schema is nil")
	}

	for key, value := range params {
		param, ok := tmpl.Parameters[key]
		if !ok {
			return fmt.Errorf("unknown parameter %q", key)
		}

		// Apply parameter based on type
		switch param.Type {
		case "int":
			count, ok := value.(int)
			if !ok {
				return fmt.Errorf("parameter %q must be an integer", key)
			}

			// Update row count for matching table
			table, ok := tmpl.Schema.Tables[key]
			if ok {
				table.RowCount = count
			}

		case "string":
			// Not yet implemented
		case "bool":
			// Not yet implemented
		case "string_array":
			// Not yet implemented
		default:
			return fmt.Errorf("unsupported parameter type %q for %q", param.Type, key)
		}
	}

	return nil
}
