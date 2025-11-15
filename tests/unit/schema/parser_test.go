package schema_test

import (
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseValid(t *testing.T) {
	t.Run("parse minimal valid schema", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"database": {
				"name": "testdb",
				"encoding": "UTF8",
				"locale": "en_US.utf8"
			},
			"tables": {
				"users": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "email", "type": "varchar(255)"}
					],
					"primary_key": ["id"],
					"row_count": 100
				}
			}
		}`

		s, err := schema.Parse(strings.NewReader(input))
		require.NoError(t, err)
		require.NotNil(t, s)

		assert.Equal(t, "1.0", s.Version)
		assert.Equal(t, "testdb", s.Database.Name)
		assert.Equal(t, "UTF8", s.Database.Encoding)
		assert.Len(t, s.Tables, 1)

		users := s.Tables["users"]
		require.NotNil(t, users)
		assert.Len(t, users.Columns, 2)
		assert.Equal(t, "id", users.Columns[0].Name)
		assert.Equal(t, "serial", users.Columns[0].Type)
		assert.Equal(t, 100, users.RowCount)
	})

	t.Run("parse schema with foreign keys", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"posts": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "user_id", "type": "integer"}
					],
					"foreign_keys": [{
						"columns": ["user_id"],
						"referenced_table": "users",
						"referenced_columns": ["id"],
						"on_delete": "CASCADE"
					}],
					"row_count": 500
				}
			}
		}`

		s, err := schema.Parse(strings.NewReader(input))
		require.NoError(t, err)

		posts := s.Tables["posts"]
		require.NotNil(t, posts)
		require.Len(t, posts.ForeignKeys, 1)

		fk := posts.ForeignKeys[0]
		assert.Equal(t, []string{"user_id"}, fk.Columns)
		assert.Equal(t, "users", fk.ReferencedTable)
		assert.Equal(t, []string{"id"}, fk.ReferencedColumns)
		assert.Equal(t, "CASCADE", fk.OnDelete)
	})

	t.Run("parse schema with constraints and indexes", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"products": {
					"columns": [
						{"name": "id", "type": "serial"},
						{"name": "sku", "type": "varchar(50)"},
						{"name": "price", "type": "decimal(10,2)"}
					],
					"unique_constraints": [{
						"name": "uq_products_sku",
						"columns": ["sku"]
					}],
					"check_constraints": [{
						"name": "ck_products_price",
						"expression": "price >= 0"
					}],
					"indexes": [{
						"name": "idx_products_price",
						"columns": ["price"],
						"type": "btree"
					}],
					"row_count": 1000
				}
			}
		}`

		s, err := schema.Parse(strings.NewReader(input))
		require.NoError(t, err)

		products := s.Tables["products"]
		require.NotNil(t, products)
		assert.Len(t, products.UniqueConstraints, 1)
		assert.Len(t, products.CheckConstraints, 1)
		assert.Len(t, products.Indexes, 1)

		assert.Equal(t, "uq_products_sku", products.UniqueConstraints[0].Name)
		assert.Equal(t, "ck_products_price", products.CheckConstraints[0].Name)
		assert.Equal(t, "idx_products_price", products.Indexes[0].Name)
	})

	t.Run("parse schema with sequences and custom types", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"sequences": {
				"order_id_seq": {
					"start": 1000,
					"increment": 1
				}
			},
			"custom_types": {
				"status_enum": {
					"kind": "enum",
					"definition": {
						"values": ["pending", "active", "completed"]
					}
				}
			},
			"tables": {}
		}`

		s, err := schema.Parse(strings.NewReader(input))
		require.NoError(t, err)

		require.Len(t, s.Sequences, 1)
		seq := s.Sequences["order_id_seq"]
		assert.Equal(t, int64(1000), seq.Start)
		assert.Equal(t, int64(1), seq.Increment)

		require.Len(t, s.CustomTypes, 1)
		ct := s.CustomTypes["status_enum"]
		assert.Equal(t, "enum", ct.Kind)
		require.NotNil(t, ct.Definition)
	})

	t.Run("parse schema with generator config", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"users": {
					"columns": [{
						"name": "status",
						"type": "varchar(20)",
						"generator": "weighted_enum",
						"generator_config": {
							"values": ["active", "inactive", "pending"],
							"weights": [80, 15, 5]
						}
					}],
					"row_count": 1000
				}
			}
		}`

		s, err := schema.Parse(strings.NewReader(input))
		require.NoError(t, err)

		users := s.Tables["users"]
		col := users.Columns[0]
		assert.Equal(t, "weighted_enum", col.GeneratorType)
		require.NotNil(t, col.GeneratorConfig)
		assert.Contains(t, col.GeneratorConfig, "values")
		assert.Contains(t, col.GeneratorConfig, "weights")
	})
}

func TestParseInvalid(t *testing.T) {
	t.Run("invalid JSON syntax", func(t *testing.T) {
		input := `{"version": "1.0", invalid json`
		s, err := schema.Parse(strings.NewReader(input))
		assert.Error(t, err)
		assert.Nil(t, s)
		assert.Contains(t, err.Error(), "parse")
	})

	t.Run("missing required version field", func(t *testing.T) {
		input := `{
			"database": {"name": "testdb"},
			"tables": {}
		}`
		s, err := schema.Parse(strings.NewReader(input))
		assert.Error(t, err)
		assert.Nil(t, s)
		assert.Contains(t, err.Error(), "version")
	})

	t.Run("missing required database field", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"tables": {}
		}`
		s, err := schema.Parse(strings.NewReader(input))
		assert.Error(t, err)
		assert.Nil(t, s)
		assert.Contains(t, err.Error(), "database")
	})

	t.Run("empty reader", func(t *testing.T) {
		s, err := schema.Parse(strings.NewReader(""))
		assert.Error(t, err)
		assert.Nil(t, s)
	})
}

func TestParseEdgeCases(t *testing.T) {
	t.Run("empty tables map", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {}
		}`
		s, err := schema.Parse(strings.NewReader(input))
		require.NoError(t, err)
		assert.NotNil(t, s.Tables)
		assert.Len(t, s.Tables, 0)
	})

	t.Run("nullable and default values", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"items": {
					"columns": [{
						"name": "description",
						"type": "text",
						"nullable": true,
						"default": "No description"
					}],
					"row_count": 50
				}
			}
		}`
		s, err := schema.Parse(strings.NewReader(input))
		require.NoError(t, err)

		items := s.Tables["items"]
		col := items.Columns[0]
		assert.True(t, col.Nullable)
		assert.Equal(t, "No description", col.DefaultValue)
	})

	t.Run("extensions field", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"extensions": ["uuid-ossp", "pgcrypto"],
			"tables": {}
		}`
		s, err := schema.Parse(strings.NewReader(input))
		require.NoError(t, err)
		assert.Equal(t, []string{"uuid-ossp", "pgcrypto"}, s.Extensions)
	})

	t.Run("column with comment", func(t *testing.T) {
		input := `{
			"version": "1.0",
			"database": {"name": "testdb"},
			"tables": {
				"tasks": {
					"columns": [{
						"name": "priority",
						"type": "integer",
						"comment": "Priority level: 1=high, 2=medium, 3=low"
					}],
					"row_count": 100
				}
			}
		}`
		s, err := schema.Parse(strings.NewReader(input))
		require.NoError(t, err)

		tasks := s.Tables["tasks"]
		col := tasks.Columns[0]
		assert.Equal(t, "Priority level: 1=high, 2=medium, 3=low", col.Comment)
	})
}