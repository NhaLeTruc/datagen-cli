package schema_test

import (
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSchema(t *testing.T) {
	t.Run("valid minimal schema", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{
				Name: "testdb",
			},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
						{Name: "email", Type: "varchar(255)"},
					},
					RowCount: 100,
				},
			},
		}

		errs := schema.Validate(s)
		assert.Empty(t, errs)
	})

	t.Run("valid schema with foreign keys", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
					},
					RowCount: 100,
				},
				"posts": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
						{Name: "user_id", Type: "integer"},
					},
					ForeignKeys: []*schema.ForeignKey{
						{
							Columns:           []string{"user_id"},
							ReferencedTable:   "users",
							ReferencedColumns: []string{"id"},
						},
					},
					RowCount: 500,
				},
			},
		}

		errs := schema.Validate(s)
		assert.Empty(t, errs)
	})
}

func TestValidateInvalidTypes(t *testing.T) {
	t.Run("invalid column type", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "invalid_type"},
					},
					RowCount: 100,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "invalid_type")
		assert.Contains(t, errs[0].Error(), "users")
	})

	t.Run("empty column name", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "", Type: "integer"},
					},
					RowCount: 100,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "column name")
	})

	t.Run("empty column type", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: ""},
					},
					RowCount: 100,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "column type")
	})
}

func TestValidateForeignKeys(t *testing.T) {
	t.Run("foreign key to non-existent table", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"posts": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
						{Name: "user_id", Type: "integer"},
					},
					ForeignKeys: []*schema.ForeignKey{
						{
							Columns:           []string{"user_id"},
							ReferencedTable:   "users",
							ReferencedColumns: []string{"id"},
						},
					},
					RowCount: 100,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "users")
		assert.Contains(t, errs[0].Error(), "non-existent")
	})

	t.Run("foreign key to non-existent column", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
					},
					RowCount: 100,
				},
				"posts": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
						{Name: "user_id", Type: "integer"},
					},
					ForeignKeys: []*schema.ForeignKey{
						{
							Columns:           []string{"user_id"},
							ReferencedTable:   "users",
							ReferencedColumns: []string{"uuid"},
						},
					},
					RowCount: 500,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "uuid")
		assert.Contains(t, errs[0].Error(), "non-existent")
	})

	t.Run("foreign key column does not exist", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
					},
					RowCount: 100,
				},
				"posts": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
					},
					ForeignKeys: []*schema.ForeignKey{
						{
							Columns:           []string{"user_id"},
							ReferencedTable:   "users",
							ReferencedColumns: []string{"id"},
						},
					},
					RowCount: 500,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "user_id")
	})
}

func TestValidateCircularDependencies(t *testing.T) {
	t.Run("circular dependency between two tables", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
						{Name: "main_post_id", Type: "integer", Nullable: true},
					},
					ForeignKeys: []*schema.ForeignKey{
						{
							Columns:           []string{"main_post_id"},
							ReferencedTable:   "posts",
							ReferencedColumns: []string{"id"},
						},
					},
					RowCount: 100,
				},
				"posts": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
						{Name: "user_id", Type: "integer"},
					},
					ForeignKeys: []*schema.ForeignKey{
						{
							Columns:           []string{"user_id"},
							ReferencedTable:   "users",
							ReferencedColumns: []string{"id"},
						},
					},
					RowCount: 500,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "circular")
	})

	t.Run("self-referencing table", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"employees": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
						{Name: "manager_id", Type: "integer", Nullable: true},
					},
					ForeignKeys: []*schema.ForeignKey{
						{
							Columns:           []string{"manager_id"},
							ReferencedTable:   "employees",
							ReferencedColumns: []string{"id"},
						},
					},
					RowCount: 100,
				},
			},
		}

		// Self-referencing is allowed (common pattern)
		errs := schema.Validate(s)
		assert.Empty(t, errs)
	})
}

func TestValidateConstraints(t *testing.T) {
	t.Run("unique constraint on non-existent column", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
						{Name: "email", Type: "varchar(255)"},
					},
					UniqueConstraints: []*schema.UniqueConstraint{
						{
							Columns: []string{"username"},
						},
					},
					RowCount: 100,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "username")
	})

	t.Run("index on non-existent column", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
					},
					Indexes: []*schema.Index{
						{
							Columns: []string{"email"},
						},
					},
					RowCount: 100,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "email")
	})
}

func TestValidateRowCount(t *testing.T) {
	t.Run("zero row count", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
					},
					RowCount: 0,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "row_count")
	})

	t.Run("negative row count", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
					},
					RowCount: -10,
				},
			},
		}

		errs := schema.Validate(s)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "row_count")
	})
}

func TestValidateMultipleErrors(t *testing.T) {
	t.Run("accumulate multiple errors", func(t *testing.T) {
		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{Name: "testdb"},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "", Type: "serial"},           // Empty name
						{Name: "email", Type: ""},             // Empty type
						{Name: "age", Type: "invalid_type"},   // Invalid type
					},
					RowCount: -5, // Invalid row count
				},
			},
		}

		errs := schema.Validate(s)
		assert.GreaterOrEqual(t, len(errs), 3, "should have at least 3 errors")
	})
}