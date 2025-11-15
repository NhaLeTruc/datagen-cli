package pgdump_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pgdump"
	"github.com/NhaLeTruc/datagen-cli/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCOPYWriter(t *testing.T) {
	t.Run("write CREATE TABLE statement", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewCOPYWriter(buf)

		table := &schema.Table{
			Columns: []*schema.Column{
				{Name: "id", Type: "serial", Nullable: false},
				{Name: "email", Type: "varchar(255)", Nullable: false},
				{Name: "name", Type: "text", Nullable: true},
			},
			PrimaryKey: []string{"id"},
			RowCount:   10,
		}

		err := writer.WriteCreateTable("users", table)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "CREATE TABLE users")
		assert.Contains(t, output, "id serial NOT NULL")
		assert.Contains(t, output, "email varchar(255) NOT NULL")
		assert.Contains(t, output, "name text")
		assert.Contains(t, output, "PRIMARY KEY (id)")
	})

	t.Run("write COPY FROM stdin statement", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewCOPYWriter(buf)

		err := writer.WriteCopyHeader("users", []string{"id", "email", "name"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "COPY users (id, email, name) FROM stdin")
	})

	t.Run("write data rows in TSV format", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewCOPYWriter(buf)

		row := map[string]interface{}{
			"id":    1,
			"email": "test@example.com",
			"name":  "John Doe",
		}

		err := writer.WriteCopyRow([]string{"id", "email", "name"}, row)
		require.NoError(t, err)

		output := buf.String()
		// Should be tab-separated
		assert.Contains(t, output, "1\ttest@example.com\tJohn Doe")
	})

	t.Run("handle NULL values in COPY", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewCOPYWriter(buf)

		row := map[string]interface{}{
			"id":    1,
			"email": "test@example.com",
			"name":  nil,
		}

		err := writer.WriteCopyRow([]string{"id", "email", "name"}, row)
		require.NoError(t, err)

		output := buf.String()
		// NULL should be represented as \N
		assert.Contains(t, output, "\\N")
	})

	t.Run("escape special characters in COPY", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewCOPYWriter(buf)

		row := map[string]interface{}{
			"id":   1,
			"text": "Line1\nLine2\tTabbed",
		}

		err := writer.WriteCopyRow([]string{"id", "text"}, row)
		require.NoError(t, err)

		output := buf.String()
		// Should escape newlines and tabs
		assert.Contains(t, output, "\\n")
		assert.Contains(t, output, "\\t")
	})

	t.Run("write COPY footer", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewCOPYWriter(buf)

		err := writer.WriteCopyFooter()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "\\.")
	})
}

func TestCOPYWriterComplete(t *testing.T) {
	t.Run("write complete schema with COPY format", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewCOPYWriter(buf)

		s := &schema.Schema{
			Version: "1.0",
			Database: schema.DatabaseConfig{
				Name:     "testdb",
				Encoding: "UTF8",
			},
			Tables: map[string]*schema.Table{
				"users": {
					Columns: []*schema.Column{
						{Name: "id", Type: "serial", Nullable: false},
						{Name: "email", Type: "varchar(255)", Nullable: false},
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
		// WriteSchema only writes table structure, not COPY commands
		assert.NotContains(t, output, "COPY users")
	})

	t.Run("write full COPY workflow", func(t *testing.T) {
		buf := new(bytes.Buffer)
		writer := pgdump.NewCOPYWriter(buf)

		// Write header
		err := writer.WriteCopyHeader("users", []string{"id", "email"})
		require.NoError(t, err)

		// Write rows
		rows := []map[string]interface{}{
			{"id": 1, "email": "user1@example.com"},
			{"id": 2, "email": "user2@example.com"},
			{"id": 3, "email": "user3@example.com"},
		}

		for _, row := range rows {
			err := writer.WriteCopyRow([]string{"id", "email"}, row)
			require.NoError(t, err)
		}

		// Write footer
		err = writer.WriteCopyFooter()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "COPY users (id, email) FROM stdin")
		// Should have: 1 header line + 3 data rows + 1 footer line = 5 newlines
		assert.Equal(t, 5, strings.Count(output, "\n"))
		assert.Contains(t, output, "\\.")
	})
}
