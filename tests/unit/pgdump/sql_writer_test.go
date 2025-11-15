package pgdump_test

import (
	"bytes"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pgdump"
	"github.com/NhaLeTruc/datagen-cli/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLWriter(t *testing.T) {
	t.Run("write CREATE TABLE statement", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewSQLWriter(buf)

		table := &schema.Table{
			Columns: []*schema.Column{
				{Name: "id", Type: "serial"},
				{Name: "email", Type: "varchar(255)"},
			},
			PrimaryKey: []string{"id"},
			RowCount:   10,
		}

		err := writer.WriteCreateTable("users", table)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "CREATE TABLE users")
		assert.Contains(t, output, "id serial")
		assert.Contains(t, output, "email varchar(255)")
		assert.Contains(t, output, "PRIMARY KEY (id)")
	})

	t.Run("write INSERT statements", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewSQLWriter(buf)

		row := map[string]interface{}{
			"id":    1,
			"email": "test@example.com",
		}

		err := writer.WriteInsert("users", []string{"id", "email"}, row)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "INSERT INTO users")
		assert.Contains(t, output, "(id, email)")
		assert.Contains(t, output, "VALUES")
		assert.Contains(t, output, "test@example.com")
	})

	t.Run("escape SQL strings", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewSQLWriter(buf)

		row := map[string]interface{}{
			"id":   1,
			"text": "O'Reilly's \"book\"",
		}

		err := writer.WriteInsert("items", []string{"id", "text"}, row)
		require.NoError(t, err)

		output := buf.String()
		// Should escape quotes
		assert.Contains(t, output, "O''Reilly''s")
	})

	t.Run("handle NULL values", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewSQLWriter(buf)

		row := map[string]interface{}{
			"id":          1,
			"description": nil,
		}

		err := writer.WriteInsert("items", []string{"id", "description"}, row)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "NULL")
	})
}

func TestSQLWriterComplete(t *testing.T) {
	t.Run("write complete schema", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewSQLWriter(buf)

		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{
				Name:     "testdb",
				Encoding: "UTF8",
			},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial"},
						{Name: "email", Type: "varchar(255)"},
					},
					PrimaryKey: []string{"id"},
					RowCount:   5,
				},
			},
		}

		err := writer.WriteSchema(s)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "CREATE DATABASE")
		assert.Contains(t, output, "CREATE TABLE users")
	})
}