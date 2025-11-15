package pgdump_test

import (
	"bytes"
	"fmt"
	"strings"
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

func TestSQLWriterBatchInsert(t *testing.T) {
	t.Run("write batched INSERT statements", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewSQLWriter(buf)

		rows := []map[string]interface{}{
			{"id": 1, "email": "user1@example.com"},
			{"id": 2, "email": "user2@example.com"},
			{"id": 3, "email": "user3@example.com"},
		}

		err := writer.WriteBatchInsert("users", []string{"id", "email"}, rows, 10)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "INSERT INTO users")
		assert.Contains(t, output, "(id, email) VALUES")
		assert.Contains(t, output, "user1@example.com")
		assert.Contains(t, output, "user2@example.com")
		assert.Contains(t, output, "user3@example.com")
		// All rows should be in one INSERT statement
		assert.Equal(t, 1, strings.Count(output, "INSERT INTO"))
	})

	t.Run("respect batch size", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewSQLWriter(buf)

		rows := make([]map[string]interface{}, 5)
		for i := 0; i < 5; i++ {
			rows[i] = map[string]interface{}{
				"id":    i + 1,
				"email": fmt.Sprintf("user%d@example.com", i+1),
			}
		}

		err := writer.WriteBatchInsert("users", []string{"id", "email"}, rows, 2)
		require.NoError(t, err)

		output := buf.String()
		// Should have 3 INSERT statements (2+2+1)
		assert.Equal(t, 3, strings.Count(output, "INSERT INTO"))
	})

	t.Run("handle empty rows", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewSQLWriter(buf)

		rows := []map[string]interface{}{}

		err := writer.WriteBatchInsert("users", []string{"id", "email"}, rows, 10)
		require.NoError(t, err)

		output := buf.String()
		assert.Empty(t, output)
	})

	t.Run("default batch size when zero", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewSQLWriter(buf)

		rows := make([]map[string]interface{}, 5)
		for i := 0; i < 5; i++ {
			rows[i] = map[string]interface{}{
				"id":    i + 1,
				"email": fmt.Sprintf("user%d@example.com", i+1),
			}
		}

		err := writer.WriteBatchInsert("users", []string{"id", "email"}, rows, 0)
		require.NoError(t, err)

		output := buf.String()
		// With default batch size of 100, all 5 rows should be in one INSERT
		assert.Equal(t, 1, strings.Count(output, "INSERT INTO"))
	})
}