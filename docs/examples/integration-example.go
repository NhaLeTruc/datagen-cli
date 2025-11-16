package main

// This is a conceptual example showing how the pipeline integrates
// distribution, pattern, and rules generators

import (
	"github.com/NhaLeTruc/datagen-cli/internal/generator"
	"github.com/NhaLeTruc/datagen-cli/internal/schema"
)

func generateRowData(table *schema.Table, ctx *generator.Context) (map[string]interface{}, error) {
	row := make(map[string]interface{})

	// Process columns in order
	for _, column := range table.Columns {
		var gen generator.Generator
		var err error

		// Determine which generator to use based on column configuration
		gen, err = selectGenerator(column, ctx)
		if err != nil {
			return nil, err
		}

		// Update context for this column
		ctx.ColumnName = column.Name

		// Generate the value
		value, err := gen.Generate(ctx)
		if err != nil {
			return nil, err
		}

		// Store in row (so subsequent columns can reference it via rules)
		row[column.Name] = value
		ctx.RowData[column.Name] = value
	}

	return row, nil
}

func selectGenerator(column *schema.Column, ctx *generator.Context) (generator.Generator, error) {
	registry := generator.DefaultRegistry()

	// Priority 1: Business Rules (highest priority - can override everything)
	if len(column.Rules) > 0 {
		// Get base generator for fallback
		baseGen, _ := getBaseGenerator(column, ctx)
		return generator.NewRulesGenerator(column.Rules, baseGen), nil
	}

	// Priority 2: Distribution (for weighted/statistical generation)
	if column.Distribution != nil {
		return generator.NewDistributionGenerator(column.Distribution), nil
	}

	// Priority 3: Pattern (for template-based generation)
	if column.Pattern != nil {
		return generator.NewPatternGenerator(column.Pattern), nil
	}

	// Priority 4: Custom Generator Type
	if column.GeneratorType != "" {
		return registry.Get(column.GeneratorType)
	}

	// Priority 5: Semantic Detection (auto-detect from column name)
	if gen := detectSemanticGenerator(column.Name); gen != nil {
		return gen, nil
	}

	// Priority 6: Basic Type Generator (fallback)
	return getBasicTypeGenerator(column.Type)
}

func getBaseGenerator(column *schema.Column, ctx *generator.Context) (generator.Generator, error) {
	// Return the generator that would be used without rules
	if column.Distribution != nil {
		return generator.NewDistributionGenerator(column.Distribution), nil
	}
	if column.Pattern != nil {
		return generator.NewPatternGenerator(column.Pattern), nil
	}
	// ... etc
	return nil, nil
}

// Stub functions (would be implemented elsewhere)
func detectSemanticGenerator(columnName string) generator.Generator { return nil }
func getBasicTypeGenerator(typeName string) (generator.Generator, error) { return nil, nil }
