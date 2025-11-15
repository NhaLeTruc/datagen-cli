package pgdump

import (
	"fmt"
	"io"

	"github.com/NhaLeTruc/datagen-cli/internal/schema"
)

// Writer interface for different output formats
type Writer interface {
	WriteSchema(s *schema.Schema) error
}

// RowWriter interface for writers that support row-by-row writing
type RowWriter interface {
	Writer
	WriteInsert(tableName string, columns []string, row map[string]interface{}) error
}

// COPYRowWriter interface for writers that support COPY format
type COPYRowWriter interface {
	Writer
	WriteCopyHeader(tableName string, columns []string) error
	WriteCopyRow(columns []string, row map[string]interface{}) error
	WriteCopyFooter() error
}

// NewWriter creates a writer based on the specified format
func NewWriter(output io.Writer, format string) (Writer, error) {
	switch format {
	case "sql":
		return NewSQLWriter(output), nil
	case "copy":
		return NewCOPYWriter(output), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// IsRowWriter checks if a writer supports row-by-row INSERT statements
func IsRowWriter(w Writer) (RowWriter, bool) {
	rw, ok := w.(RowWriter)
	return rw, ok
}

// IsCOPYRowWriter checks if a writer supports COPY format
func IsCOPYRowWriter(w Writer) (COPYRowWriter, bool) {
	cw, ok := w.(COPYRowWriter)
	return cw, ok
}
