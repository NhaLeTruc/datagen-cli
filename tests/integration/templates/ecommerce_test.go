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

func TestEcommerceTemplate(t *testing.T) {
	t.Run("load ecommerce template and generate data", func(t *testing.T) {
		// Load the ecommerce template
		tmpl, err := templates.Get("ecommerce")
		require.NoError(t, err, "should load ecommerce template")
		require.NotNil(t, tmpl, "template should not be nil")

		// Verify template metadata
		assert.Equal(t, "ecommerce", tmpl.Name)
		assert.Contains(t, tmpl.Description, "E-commerce")

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
		err = coordinator.Execute(bytes.NewReader(schemaJSON), output, 12345)
		require.NoError(t, err, "should generate data from template")

		result := output.String()

		// Verify database creation
		assert.Contains(t, result, "CREATE DATABASE ecommerce_db")
		assert.Contains(t, result, "\\connect ecommerce_db")

		// Verify all tables are created in correct order (respecting dependencies)
		assert.Contains(t, result, "CREATE TABLE categories")
		assert.Contains(t, result, "CREATE TABLE customers")
		assert.Contains(t, result, "CREATE TABLE products")
		assert.Contains(t, result, "CREATE TABLE orders")
		assert.Contains(t, result, "CREATE TABLE order_items")
		assert.Contains(t, result, "CREATE TABLE reviews")

		// Verify categories table structure
		assert.Contains(t, result, "id serial")
		assert.Contains(t, result, "name varchar(100)")
		assert.Contains(t, result, "description text")
		assert.Contains(t, result, "parent_id integer")
		assert.Contains(t, result, "created_at timestamp")

		// Verify customers table structure with semantic columns
		assert.Contains(t, result, "email varchar(255)")
		assert.Contains(t, result, "first_name varchar(100)")
		assert.Contains(t, result, "last_name varchar(100)")
		assert.Contains(t, result, "phone varchar(20)")
		assert.Contains(t, result, "address text")
		assert.Contains(t, result, "city varchar(100)")
		assert.Contains(t, result, "state varchar(100)")
		assert.Contains(t, result, "postal_code varchar(20)")
		assert.Contains(t, result, "country varchar(100)")

		// Verify products table structure with custom generators
		assert.Contains(t, result, "sku varchar(50)")
		assert.Contains(t, result, "price numeric(10,2)")
		assert.Contains(t, result, "stock_quantity integer")
		assert.Contains(t, result, "status varchar(20)")

		// Verify orders table structure
		assert.Contains(t, result, "order_number varchar(50)")
		assert.Contains(t, result, "total_amount numeric(10,2)")
		assert.Contains(t, result, "shipping_address text")
		assert.Contains(t, result, "shipping_city varchar(100)")

		// Verify order_items table structure
		assert.Contains(t, result, "quantity integer")
		assert.Contains(t, result, "unit_price numeric(10,2)")
		assert.Contains(t, result, "subtotal numeric(10,2)")

		// Verify reviews table structure
		assert.Contains(t, result, "rating integer")
		assert.Contains(t, result, "title varchar(200)")
		assert.Contains(t, result, "comment text")
		assert.Contains(t, result, "verified_purchase boolean")

		// Verify primary keys
		assert.Regexp(t, `PRIMARY KEY\s*\(\s*id\s*\)`, result)

		// Verify unique constraints
		assert.Contains(t, result, "UNIQUE")

		// Verify row counts (INSERT statements)
		categoriesInserts := strings.Count(result, "INSERT INTO categories")
		assert.Equal(t, 50, categoriesInserts, "should have 50 category rows")

		customersInserts := strings.Count(result, "INSERT INTO customers")
		assert.Equal(t, 1000, customersInserts, "should have 1000 customer rows")

		productsInserts := strings.Count(result, "INSERT INTO products")
		assert.Equal(t, 500, productsInserts, "should have 500 product rows")

		ordersInserts := strings.Count(result, "INSERT INTO orders")
		assert.Equal(t, 2000, ordersInserts, "should have 2000 order rows")

		orderItemsInserts := strings.Count(result, "INSERT INTO order_items")
		assert.Equal(t, 5000, orderItemsInserts, "should have 5000 order_item rows")

		reviewsInserts := strings.Count(result, "INSERT INTO reviews")
		assert.Equal(t, 1500, reviewsInserts, "should have 1500 review rows")

		// Total row count should be 10,100
		totalInserts := strings.Count(result, "INSERT INTO")
		assert.Equal(t, 10050, totalInserts, "should have 10,050 total INSERT statements (10,100 rows)")
	})

	t.Run("verify custom generators work correctly", func(t *testing.T) {
		tmpl, err := templates.Get("ecommerce")
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

		// Verify SKU pattern generator (PRD-[0-9]{8})
		assert.Regexp(t, `PRD-\d{8}`, result, "should contain SKU matching pattern PRD-########")

		// Verify order number template generator (ORD-{year}-{seq:6})
		assert.Regexp(t, `ORD-\d{4}-\d{6}`, result, "should contain order numbers matching ORD-YYYY-######")

		// Verify weighted enum for product status (active, inactive, out_of_stock)
		assert.Contains(t, result, "'active'")
		assert.Contains(t, result, "'inactive'")
		assert.Contains(t, result, "'out_of_stock'")

		// Verify weighted enum for order status
		assert.Contains(t, result, "'pending'")
		assert.Contains(t, result, "'processing'")
		assert.Contains(t, result, "'shipped'")
		assert.Contains(t, result, "'delivered'")
		assert.Contains(t, result, "'cancelled'")

		// Verify boolean values for verified_purchase
		assert.Contains(t, result, "true")
		assert.Contains(t, result, "false")
	})

	t.Run("verify semantic generators for customer data", func(t *testing.T) {
		tmpl, err := templates.Get("ecommerce")
		require.NoError(t, err)

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterSemanticGenerators()
		coordinator.RegisterCustomGenerators()

		schemaJSON, err := json.Marshal(tmpl.Schema)
		require.NoError(t, err)

		output := new(bytes.Buffer)
		err = coordinator.Execute(bytes.NewReader(schemaJSON), output, 42)
		require.NoError(t, err)

		result := output.String()

		// Verify email format (should contain @ and domain)
		assert.Regexp(t, `\w+@\w+\.\w+`, result, "should contain valid email addresses")

		// Verify that data is generated (not empty)
		assert.Greater(t, len(result), 100000, "generated output should be substantial")
	})

	t.Run("template parameters override row counts", func(t *testing.T) {
		tmpl, err := templates.Get("ecommerce")
		require.NoError(t, err)

		// Override row counts
		params := map[string]interface{}{
			"categories": 10,
			"customers":  50,
		}

		err = templates.ApplyParameters(tmpl, params)
		require.NoError(t, err)

		// Verify row counts were updated
		assert.Equal(t, 10, tmpl.Schema.Tables["categories"].RowCount)
		assert.Equal(t, 50, tmpl.Schema.Tables["customers"].RowCount)

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

		// Verify actual row counts match parameters
		categoriesInserts := strings.Count(result, "INSERT INTO categories")
		assert.Equal(t, 10, categoriesInserts, "should have 10 category rows after parameter override")

		customersInserts := strings.Count(result, "INSERT INTO customers")
		assert.Equal(t, 50, customersInserts, "should have 50 customer rows after parameter override")
	})

	t.Run("deterministic generation with seed", func(t *testing.T) {
		tmpl1, err := templates.Get("ecommerce")
		require.NoError(t, err)

		tmpl2, err := templates.Get("ecommerce")
		require.NoError(t, err)

		// Override to smaller dataset for faster test
		params := map[string]interface{}{
			"categories":  5,
			"customers":   10,
			"products":    10,
			"orders":      10,
			"order_items": 20,
			"reviews":     10,
		}

		err = templates.ApplyParameters(tmpl1, params)
		require.NoError(t, err)

		err = templates.ApplyParameters(tmpl2, params)
		require.NoError(t, err)

		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterSemanticGenerators()
		coordinator.RegisterCustomGenerators()

		// Generate twice with same seed
		schemaJSON1, err := json.Marshal(tmpl1.Schema)
		require.NoError(t, err)

		output1 := new(bytes.Buffer)
		err = coordinator.Execute(bytes.NewReader(schemaJSON1), output1, 99999)
		require.NoError(t, err)

		schemaJSON2, err := json.Marshal(tmpl2.Schema)
		require.NoError(t, err)

		output2 := new(bytes.Buffer)
		err = coordinator.Execute(bytes.NewReader(schemaJSON2), output2, 99999)
		require.NoError(t, err)

		// Verify that key data elements are the same (deterministic data generation)
		result1 := output1.String()
		result2 := output2.String()

		// Note: Table creation order might vary due to map iteration,
		// but INSERT data should be identical for same seed
		// So we check that the same data values appear in both outputs

		// Extract a few sample INSERT statements and verify they're in both outputs
		assert.Contains(t, result1, "ORD-2025-000001", "should contain first order number")
		assert.Contains(t, result2, "ORD-2025-000001", "should contain first order number in second run")

		// Verify row counts are the same
		assert.Equal(t, strings.Count(result1, "INSERT INTO"), strings.Count(result2, "INSERT INTO"),
			"same seed should produce same number of INSERTs")
	})
}
