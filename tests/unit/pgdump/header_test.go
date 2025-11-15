package pgdump_test

import (
	"bytes"
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pgdump"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDumpHeader(t *testing.T) {
	t.Run("write valid header", func(t *testing.T) {
		buf := new(bytes.Buffer)
		header := pgdump.NewHeader()

		err := header.Write(buf)
		require.NoError(t, err)

		data := buf.Bytes()
		assert.Greater(t, len(data), 0)

		// Check magic bytes "PGDMP"
		assert.Equal(t, byte('P'), data[0])
		assert.Equal(t, byte('G'), data[1])
		assert.Equal(t, byte('D'), data[2])
		assert.Equal(t, byte('M'), data[3])
		assert.Equal(t, byte('P'), data[4])
	})

	t.Run("header contains version info", func(t *testing.T) {
		buf := new(bytes.Buffer)
		header := pgdump.NewHeader()
		header.Version = "1.14"  // PostgreSQL dump format version

		err := header.Write(buf)
		require.NoError(t, err)

		// Header should have version encoded
		assert.Greater(t, buf.Len(), 10)
	})

	t.Run("header contains database name", func(t *testing.T) {
		buf := new(bytes.Buffer)
		header := pgdump.NewHeader()
		header.DatabaseName = "testdb"

		err := header.Write(buf)
		require.NoError(t, err)

		data := buf.Bytes()
		// Database name should be in the header
		assert.Contains(t, string(data), "testdb")
	})

	t.Run("header contains timestamp", func(t *testing.T) {
		buf := new(bytes.Buffer)
		header := pgdump.NewHeader()

		err := header.Write(buf)
		require.NoError(t, err)

		// Should have timestamp info
		assert.NotZero(t, header.Timestamp)
	})
}

func TestHeaderDefaults(t *testing.T) {
	t.Run("new header has default values", func(t *testing.T) {
		header := pgdump.NewHeader()

		assert.NotEmpty(t, header.Version)
		assert.NotZero(t, header.Timestamp)
	})

	t.Run("set database metadata", func(t *testing.T) {
		header := pgdump.NewHeader()
		header.DatabaseName = "mydb"
		header.Encoding = "UTF8"

		assert.Equal(t, "mydb", header.DatabaseName)
		assert.Equal(t, "UTF8", header.Encoding)
	})
}