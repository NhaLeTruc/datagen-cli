package schema_test

import (
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchema(t *testing.T) {
	t.Run("create valid schema", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{
				Name:     "testdb",
				Encoding: "UTF8",
				Locale:   "en_US.utf8",
			},
			Tables: make(map[string]*schema.Table),
		}

		assert.Equal(t, "1.0", s.Version)
		assert.Equal(t, "testdb", s.Database.Name)
		assert.Equal(t, "UTF8", s.Database.Encoding)
		assert.NotNil(t, s.Tables)
	})

	t.Run("schema with tables", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{
				Name: "testdb",
			},
			Tables: map[string]*schema.Table{
				"users": {
					Columns:  []*schema.Column{},
					RowCount: 100,
				},
			},
		}

		require.Contains(t, s.Tables, "users")
		assert.Equal(t, 100, s.Tables["users"].RowCount)
	})
}

func TestTable(t *testing.T) {
	t.Run("create table with columns", func(t *testing.T) {
		table := &schema.Table{
			Columns: []*schema.Column{
				{
					Name: "id",
					Type: "serial",
				},
				{
					Name: "email",
					Type: "varchar(255)",
				},
			},
			PrimaryKey: []string{"id"},
			RowCount:   1000,
		}

		assert.Len(t, table.Columns, 2)
		assert.Equal(t, "id", table.Columns[0].Name)
		assert.Equal(t, "serial", table.Columns[0].Type)
		assert.Contains(t, table.PrimaryKey, "id")
		assert.Equal(t, 1000, table.RowCount)
	})

	t.Run("table with foreign keys", func(t *testing.T) {
		table := &schema.Table{
			Columns: []*schema.Column{
				{Name: "id", Type: "serial"},
				{Name: "user_id", Type: "integer"},
			},
			ForeignKeys: []*schema.ForeignKey{
				{
					Columns:           []string{"user_id"},
					ReferencedTable:   "users",
					ReferencedColumns: []string{"id"},
					OnDelete:          "CASCADE",
				},
			},
		}

		require.Len(t, table.ForeignKeys, 1)
		fk := table.ForeignKeys[0]
		assert.Equal(t, []string{"user_id"}, fk.Columns)
		assert.Equal(t, "users", fk.ReferencedTable)
		assert.Equal(t, "CASCADE", fk.OnDelete)
	})

	t.Run("table with indexes", func(t *testing.T) {
		table := &schema.Table{
			Columns: []*schema.Column{
				{Name: "id", Type: "serial"},
				{Name: "email", Type: "varchar(255)"},
			},
			Indexes: []*schema.Index{
				{
					Columns: []string{"email"},
					Unique:  true,
					Type:    "btree",
				},
			},
		}

		require.Len(t, table.Indexes, 1)
		idx := table.Indexes[0]
		assert.True(t, idx.Unique)
		assert.Equal(t, "btree", idx.Type)
	})
}

func TestColumn(t *testing.T) {
	t.Run("create basic column", func(t *testing.T) {
		col := &schema.Column{
			Name:     "email",
			Type:     "varchar(255)",
			Nullable: false,
		}

		assert.Equal(t, "email", col.Name)
		assert.Equal(t, "varchar(255)", col.Type)
		assert.False(t, col.Nullable)
	})

	t.Run("column with default value", func(t *testing.T) {
		col := &schema.Column{
			Name:         "created_at",
			Type:         "timestamp",
			DefaultValue: "now()",
			Nullable:     false,
		}

		assert.Equal(t, "now()", col.DefaultValue)
	})

	t.Run("column with generator config", func(t *testing.T) {
		col := &schema.Column{
			Name:          "status",
			Type:          "varchar(20)",
			GeneratorType: "weighted_enum",
			GeneratorConfig: map[string]interface{}{
				"values": map[string]float64{
					"active":   0.8,
					"inactive": 0.2,
				},
			},
		}

		assert.Equal(t, "weighted_enum", col.GeneratorType)
		assert.NotNil(t, col.GeneratorConfig)
	})
}

func TestForeignKey(t *testing.T) {
	t.Run("create foreign key", func(t *testing.T) {
		fk := &schema.ForeignKey{
			Columns:           []string{"user_id"},
			ReferencedTable:   "users",
			ReferencedColumns: []string{"id"},
			OnDelete:          "CASCADE",
			OnUpdate:          "NO ACTION",
		}

		assert.Equal(t, "users", fk.ReferencedTable)
		assert.Equal(t, "CASCADE", fk.OnDelete)
		assert.Equal(t, "NO ACTION", fk.OnUpdate)
	})
}

func TestUniqueConstraint(t *testing.T) {
	t.Run("create unique constraint", func(t *testing.T) {
		uc := &schema.UniqueConstraint{
			Columns: []string{"email", "domain"},
			Name:    "unique_email_domain",
		}

		assert.Len(t, uc.Columns, 2)
		assert.Equal(t, "unique_email_domain", uc.Name)
	})
}

func TestCheckConstraint(t *testing.T) {
	t.Run("create check constraint", func(t *testing.T) {
		cc := &schema.CheckConstraint{
			Expression: "age >= 18",
			Name:       "valid_age",
		}

		assert.Equal(t, "age >= 18", cc.Expression)
		assert.Equal(t, "valid_age", cc.Name)
	})
}

func TestIndex(t *testing.T) {
	t.Run("create index", func(t *testing.T) {
		idx := &schema.Index{
			Columns: []string{"email"},
			Name:    "idx_email",
			Type:    "btree",
			Unique:  true,
		}

		assert.Equal(t, "idx_email", idx.Name)
		assert.True(t, idx.Unique)
	})
}

func TestSequence(t *testing.T) {
	t.Run("create sequence", func(t *testing.T) {
		seq := &schema.Sequence{
			Start:     1,
			Increment: 1,
			Cache:     1,
		}

		assert.Equal(t, int64(1), seq.Start)
		assert.Equal(t, int64(1), seq.Increment)
	})

	t.Run("sequence with min/max", func(t *testing.T) {
		minVal := int64(1)
		maxVal := int64(9999999)

		seq := &schema.Sequence{
			Start:     1,
			Increment: 1,
			MinValue:  &minVal,
			MaxValue:  &maxVal,
		}

		require.NotNil(t, seq.MinValue)
		require.NotNil(t, seq.MaxValue)
		assert.Equal(t, int64(1), *seq.MinValue)
		assert.Equal(t, int64(9999999), *seq.MaxValue)
	})
}

func TestCustomType(t *testing.T) {
	t.Run("create enum type", func(t *testing.T) {
		ct := &schema.CustomType{
			Kind: "enum",
			Definition: schema.EnumDefinition{
				Values: []string{"pending", "active", "inactive"},
			},
		}

		assert.Equal(t, "enum", ct.Kind)
		enumDef, ok := ct.Definition.(schema.EnumDefinition)
		require.True(t, ok)
		assert.Len(t, enumDef.Values, 3)
	})
}