package pipeline

import (
	"fmt"
	"io"

	"github.com/NhaLeTruc/datagen-cli/internal/generator"
	"github.com/NhaLeTruc/datagen-cli/internal/pgdump"
	"github.com/NhaLeTruc/datagen-cli/internal/schema"
)

// Coordinator orchestrates the data generation pipeline
type Coordinator struct {
	registry *generator.Registry
	detector *generator.SemanticDetector
}

// NewCoordinator creates a new pipeline coordinator
func NewCoordinator() *Coordinator {
	return &Coordinator{
		registry: generator.DefaultRegistry(),
		detector: generator.NewSemanticDetector(),
	}
}

// Execute runs the complete pipeline: parse → validate → generate → write
func (c *Coordinator) Execute(schemaJSON io.Reader, output io.Writer, seed int64) error {
	// Parse schema
	s, err := schema.Parse(schemaJSON)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Validate schema
	errors := schema.Validate(s)
	if len(errors) > 0 {
		return fmt.Errorf("schema validation failed: %v", errors[0])
	}

	// Create SQL writer
	writer := pgdump.NewSQLWriter(output)

	// Write schema structure
	if err := writer.WriteSchema(s); err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}

	// Generate and write data for each table
	for tableName, table := range s.Tables {
		if err := c.generateTableData(writer, tableName, table, seed); err != nil {
			return fmt.Errorf("failed to generate data for table %s: %w", tableName, err)
		}
	}

	return nil
}

// generateTableData generates data for a single table
func (c *Coordinator) generateTableData(writer *pgdump.SQLWriter, tableName string, table *schema.Table, seed int64) error {
	ctx := generator.NewContextWithSeed(seed)
	ctx.TableName = tableName

	// Generate rows
	for rowIdx := 0; rowIdx < table.RowCount; rowIdx++ {
		ctx.RowIndex = rowIdx
		row := make(map[string]interface{})

		// Generate value for each column
		for _, col := range table.Columns {
			ctx.ColumnName = col.Name
			val, err := c.generateColumnValue(ctx, col)
			if err != nil {
				return fmt.Errorf("failed to generate value for column %s: %w", col.Name, err)
			}
			row[col.Name] = val
		}

		// Write INSERT statement
		columnNames := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			columnNames[i] = col.Name
		}

		if err := writer.WriteInsert(tableName, columnNames, row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// generateColumnValue generates a value for a column
func (c *Coordinator) generateColumnValue(ctx *generator.Context, col *schema.Column) (interface{}, error) {
	var genType string

	// First, try semantic detection based on column name
	if c.detector != nil {
		semanticType := c.detector.GetSemanticType(col.Name)
		if semanticType != "" {
			genType = semanticType
		}
	}

	// If no semantic match, use PostgreSQL type
	if genType == "" {
		genType = c.mapTypeToGenerator(col.Type)
	}

	// Try to get generator from registry
	gen, err := c.registry.Get(genType)
	if err != nil {
		// Fallback to varchar for unknown types
		gen = generator.NewVarcharGenerator(255)
	}

	return gen.Generate(ctx)
}

// mapTypeToGenerator maps PostgreSQL types to generator types
func (c *Coordinator) mapTypeToGenerator(pgType string) string {
	switch pgType {
	case "serial", "bigserial", "smallserial":
		return "serial"
	case "integer", "int", "bigint", "smallint":
		return "integer"
	case "boolean", "bool":
		return "boolean"
	case "timestamp", "timestamptz", "timestamp with time zone",
		"timestamp without time zone", "date", "time":
		return "timestamp"
	case "text":
		return "text"
	default:
		// varchar, char, etc.
		return "varchar"
	}
}

// RegisterBasicGenerators registers all basic generators
func (c *Coordinator) RegisterBasicGenerators() {
	c.registry.Register("integer", generator.NewIntegerGenerator())
	c.registry.Register("varchar", generator.NewVarcharGenerator(255))
	c.registry.Register("text", generator.NewTextGenerator())
	c.registry.Register("timestamp", generator.NewTimestampGenerator())
	c.registry.Register("boolean", generator.NewBooleanGenerator())
	c.registry.Register("serial", generator.NewSerialGenerator())
}

// RegisterSemanticGenerators registers all semantic generators
func (c *Coordinator) RegisterSemanticGenerators() {
	c.registry.Register("email", generator.NewEmailGenerator())
	c.registry.Register("phone", generator.NewPhoneGenerator())
	c.registry.Register("first_name", generator.NewFirstNameGenerator())
	c.registry.Register("last_name", generator.NewLastNameGenerator())
	c.registry.Register("full_name", generator.NewFullNameGenerator())
	c.registry.Register("address", generator.NewAddressGenerator())
	c.registry.Register("city", generator.NewCityGenerator())
	c.registry.Register("country", generator.NewCountryGenerator())
	c.registry.Register("postal_code", generator.NewPostalCodeGenerator())
	c.registry.Register("created_at", generator.NewCreatedAtGenerator())
	c.registry.Register("updated_at", generator.NewUpdatedAtGenerator())
}

// GetRegistry returns the generator registry
func (c *Coordinator) GetRegistry() *generator.Registry {
	return c.registry
}