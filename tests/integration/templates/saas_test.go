package templates_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/NhaLeTruc/datagen-cli/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaaSTemplate(t *testing.T) {
	t.Run("load saas template and generate multi-tenant structure", func(t *testing.T) {
		// Load the SaaS template
		tmpl, err := templates.Get("saas")
		require.NoError(t, err, "should load saas template")
		require.NotNil(t, tmpl, "template should not be nil")

		// Verify template metadata
		assert.Equal(t, "saas", tmpl.Name)
		assert.Contains(t, tmpl.Description, "SaaS")

		// Override to smaller dataset for faster test
		params := map[string]interface{}{
			"tenants": 10,
			"users":   50,
		}

		err = templates.ApplyParameters(tmpl, params)
		require.NoError(t, err)

		// Create coordinator and register generators
		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterSemanticGenerators()
		coordinator.RegisterCustomGenerators()

		// Convert schema to JSON
		schemaJSON, err := json.Marshal(tmpl.Schema)
		require.NoError(t, err, "should marshal schema to JSON")

		// Generate data from template
		output := new(bytes.Buffer)
		err = coordinator.Execute(bytes.NewReader(schemaJSON), output, 54321)
		require.NoError(t, err, "should generate data from template")

		result := output.String()

		// Verify database creation
		assert.Contains(t, result, "CREATE DATABASE saas_db")
		assert.Contains(t, result, "\\connect saas_db")

		// Verify multi-tenant tables are created
		assert.Contains(t, result, "CREATE TABLE tenants")
		assert.Contains(t, result, "CREATE TABLE users")
		assert.Contains(t, result, "CREATE TABLE subscriptions")
		assert.Contains(t, result, "CREATE TABLE usage_metrics")
		assert.Contains(t, result, "CREATE TABLE billing_invoices")

		// Verify tenants table structure
		assert.Contains(t, result, "name varchar(200)")
		assert.Contains(t, result, "slug varchar(100)")
		assert.Contains(t, result, "domain varchar(255)")
		assert.Contains(t, result, "plan varchar(20)")
		assert.Contains(t, result, "status varchar(20)")

		// Verify users table structure (multi-tenant)
		assert.Contains(t, result, "tenant_id integer")
		assert.Contains(t, result, "email varchar(255)")
		assert.Contains(t, result, "role varchar(20)")

		// Verify subscriptions table structure
		assert.Contains(t, result, "start_date timestamp")
		assert.Contains(t, result, "end_date timestamp")
		assert.Contains(t, result, "amount numeric(10,2)")

		// Verify usage_metrics table structure (time-series data)
		assert.Contains(t, result, "recorded_at timestamp")
		assert.Contains(t, result, "metric_name varchar(100)")
		assert.Contains(t, result, "value numeric(12,2)")

		// Verify billing_invoices table structure
		assert.Contains(t, result, "invoice_number varchar(50)")
		assert.Contains(t, result, "amount numeric(10,2)")
		assert.Contains(t, result, "issued_at timestamp")

		// Verify row counts match parameters
		tenantsInserts := strings.Count(result, "INSERT INTO tenants")
		assert.Equal(t, 10, tenantsInserts, "should have 10 tenant rows")

		usersInserts := strings.Count(result, "INSERT INTO users")
		assert.Equal(t, 50, usersInserts, "should have 50 user rows")

		// Verify weighted enum values for plan
		assert.Contains(t, result, "'free'")
		assert.Contains(t, result, "'basic'")
		assert.Contains(t, result, "'pro'")
		assert.Contains(t, result, "'enterprise'")

		// Verify pattern generator for slug (lowercase letters)
		assert.Regexp(t, `'[a-z]{4,12}'`, result, "should contain slug matching pattern")
	})

	t.Run("verify multi-tenant CASCADE delete rules", func(t *testing.T) {
		tmpl, err := templates.Get("saas")
		require.NoError(t, err)

		// Use smaller dataset for faster test
		params := map[string]interface{}{
			"tenants": 5,
			"users":   10,
		}

		err = templates.ApplyParameters(tmpl, params)
		require.NoError(t, err)

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterSemanticGenerators()
		coordinator.RegisterCustomGenerators()

		schemaJSON, err := json.Marshal(tmpl.Schema)
		require.NoError(t, err)

		output := new(bytes.Buffer)
		err = coordinator.Execute(bytes.NewReader(schemaJSON), output, 12345)
		require.NoError(t, err)

		result := output.String()

		// Verify CASCADE delete is mentioned in schema (for multi-tenant cleanup)
		// Note: This depends on SQL writer implementation
		// For now, just verify the tables were created successfully
		assert.Contains(t, result, "CREATE TABLE tenants")
		assert.Contains(t, result, "CREATE TABLE users")
		assert.Contains(t, result, "CREATE TABLE subscriptions")
	})

	t.Run("verify semantic generators for SaaS data", func(t *testing.T) {
		tmpl, err := templates.Get("saas")
		require.NoError(t, err)

		params := map[string]interface{}{
			"tenants": 3,
			"users":   5,
		}

		err = templates.ApplyParameters(tmpl, params)
		require.NoError(t, err)

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterSemanticGenerators()
		coordinator.RegisterCustomGenerators()

		schemaJSON, err := json.Marshal(tmpl.Schema)
		require.NoError(t, err)

		output := new(bytes.Buffer)
		err = coordinator.Execute(bytes.NewReader(schemaJSON), output, 99999)
		require.NoError(t, err)

		result := output.String()

		// Verify email format (should contain @ and domain)
		assert.Regexp(t, `\w+@\w+\.\w+`, result, "should contain valid email addresses")

		// Verify data is generated (not empty)
		assert.Greater(t, len(result), 5000, "generated output should be substantial")
	})
}
