package pipeline

import (
	"fmt"
	"io"
	"time"

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
// Uses SQL format by default
func (c *Coordinator) Execute(schemaJSON io.Reader, output io.Writer, seed int64) error {
	return c.ExecuteWithFormat(schemaJSON, output, seed, "sql")
}

// ExecuteWithFormat runs the complete pipeline with specified output format
func (c *Coordinator) ExecuteWithFormat(schemaJSON io.Reader, output io.Writer, seed int64, format string) error {
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

	// Create writer based on format
	writer, err := pgdump.NewWriter(output, format)
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}

	// Write schema structure
	if err := writer.WriteSchema(s); err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}

	// Generate and write data for each table
	for tableName, table := range s.Tables {
		if err := c.generateTableDataWithWriter(writer, tableName, table, seed); err != nil {
			return fmt.Errorf("failed to generate data for table %s: %w", tableName, err)
		}
	}

	return nil
}

// generateTableDataWithWriter generates data for a single table using any Writer
func (c *Coordinator) generateTableDataWithWriter(writer pgdump.Writer, tableName string, table *schema.Table, seed int64) error {
	ctx := generator.NewContextWithSeed(seed)
	ctx.TableName = tableName

	columnNames := make([]string, len(table.Columns))
	for i, col := range table.Columns {
		columnNames[i] = col.Name
	}

	// Check if writer supports row-by-row INSERTs (SQL format)
	if rowWriter, ok := pgdump.IsRowWriter(writer); ok {
		// Generate rows and write INSERT statements
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

			if err := rowWriter.WriteInsert(tableName, columnNames, row); err != nil {
				return fmt.Errorf("failed to write row: %w", err)
			}
		}
		return nil
	}

	// Check if writer supports COPY format
	if copyWriter, ok := pgdump.IsCOPYRowWriter(writer); ok {
		// Write COPY header
		if err := copyWriter.WriteCopyHeader(tableName, columnNames); err != nil {
			return fmt.Errorf("failed to write COPY header: %w", err)
		}

		// Generate rows and write COPY data
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

			if err := copyWriter.WriteCopyRow(columnNames, row); err != nil {
				return fmt.Errorf("failed to write COPY row: %w", err)
			}
		}

		// Write COPY footer
		if err := copyWriter.WriteCopyFooter(); err != nil {
			return fmt.Errorf("failed to write COPY footer: %w", err)
		}
		return nil
	}

	return fmt.Errorf("writer does not support row-by-row output")
}

// generateTableData generates data for a single table (deprecated - use generateTableDataWithWriter)
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
	// First, check if there's a custom generator_config
	if col.GeneratorConfig != nil && len(col.GeneratorConfig) > 0 {
		return c.generateWithConfig(ctx, col)
	}

	var genType string

	// Try semantic detection based on column name
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

// generateWithConfig creates a custom generator based on generator_config
func (c *Coordinator) generateWithConfig(ctx *generator.Context, col *schema.Column) (interface{}, error) {
	// Use GeneratorType field if set, otherwise look for "type" in config
	genType := col.GeneratorType
	if genType == "" {
		var ok bool
		genType, ok = col.GeneratorConfig["type"].(string)
		if !ok {
			return nil, fmt.Errorf("generator_config missing 'type' field")
		}
	}

	var gen generator.Generator

	switch genType {
	case "weighted_enum":
		weights := make(map[string]float64)

		// Support two formats:
		// 1. Map format: {"weights": {"completed": 0.70, "pending": 0.20}}
		// 2. Array format: {"values": ["active", "inactive"], "weights": [80, 15]}
		if weightsMap, ok := col.GeneratorConfig["weights"].(map[string]interface{}); ok {
			// Format 1: Map format
			for k, v := range weightsMap {
				if f, ok := v.(float64); ok {
					weights[k] = f
				}
			}
		} else if values, ok := col.GeneratorConfig["values"].([]interface{}); ok {
			// Format 2: Array format
			weightsArray, _ := col.GeneratorConfig["weights"].([]interface{})
			for i, val := range values {
				key := fmt.Sprintf("%v", val)
				weight := 1.0 // Default equal weight
				if i < len(weightsArray) {
					if f, ok := weightsArray[i].(float64); ok {
						weight = f / 100.0 // Convert percentage to decimal
					}
				}
				weights[key] = weight
			}
		}
		gen = generator.NewWeightedEnumGenerator(weights)

	case "pattern":
		pattern, _ := col.GeneratorConfig["pattern"].(string)
		gen = generator.NewPatternGenerator(pattern)

	case "template":
		template, _ := col.GeneratorConfig["template"].(string)
		gen = generator.NewTemplateGenerator(template)

	case "integer_range":
		min := int64(0)
		max := int64(100)
		if minVal, ok := col.GeneratorConfig["min"].(float64); ok {
			min = int64(minVal)
		}
		if maxVal, ok := col.GeneratorConfig["max"].(float64); ok {
			max = int64(maxVal)
		}
		gen = generator.NewIntegerRangeGenerator(min, max)

	case "timeseries":
		// Parse timeseries config
		startStr, _ := col.GeneratorConfig["start"].(string)
		endStr, _ := col.GeneratorConfig["end"].(string)
		intervalStr, _ := col.GeneratorConfig["interval"].(string)
		pattern, _ := col.GeneratorConfig["pattern"].(string)

		// Parse times (simplified - assumes RFC3339 format)
		startTime, _ := time.Parse(time.RFC3339, startStr)
		endTime, _ := time.Parse(time.RFC3339, endStr)

		// Parse interval (simplified - assumes format like "1h")
		interval := time.Hour
		if len(intervalStr) > 0 {
			if parsed, err := time.ParseDuration(intervalStr); err == nil {
				interval = parsed
			}
		}

		gen = generator.NewTimeSeriesGenerator(startTime, endTime, interval, pattern)

	default:
		return nil, fmt.Errorf("unknown generator type: %s", genType)
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

// RegisterCustomGenerators registers all custom generators
// Note: These are registered as templates, actual instances created based on config
func (c *Coordinator) RegisterCustomGenerators() {
	// Custom generators are instantiated dynamically based on generator_config
	// We don't pre-register them since they require parameters
}

// GetRegistry returns the generator registry
func (c *Coordinator) GetRegistry() *generator.Registry {
	return c.registry
}