package generator

import (
	"fmt"

	"github.com/NhaLeTruc/datagen-cli/internal/schema"
)

// RulesGenerator generates values based on conditional business rules
type RulesGenerator struct {
	rules        []*schema.BusinessRule
	baseGenerator Generator // Fallback generator if no rules match
}

// NewRulesGenerator creates a rules-based generator
func NewRulesGenerator(rules []*schema.BusinessRule, baseGen Generator) *RulesGenerator {
	return &RulesGenerator{
		rules:        rules,
		baseGenerator: baseGen,
	}
}

func (g *RulesGenerator) Name() string {
	return "rules"
}

func (g *RulesGenerator) Generate(ctx *Context) (interface{}, error) {
	// Evaluate rules in order until one matches
	for _, rule := range g.rules {
		if g.evaluateCondition(ctx, rule.Condition) {
			// Apply the "then" action
			return g.applyAction(ctx, rule.Then)
		}
	}

	// If no rule matched and there's an else, use it
	// Or fall back to base generator
	if g.baseGenerator != nil {
		return g.baseGenerator.Generate(ctx)
	}

	return nil, fmt.Errorf("no matching rule and no base generator")
}

// evaluateCondition checks if a condition matches the current row context
func (g *RulesGenerator) evaluateCondition(ctx *Context, condition map[string]interface{}) bool {
	if len(condition) == 0 {
		return true // Empty condition always matches
	}

	// Check each condition field
	for field, expectedValue := range condition {
		// Get the actual value from the current row context
		actualValue, exists := ctx.RowData[field]
		if !exists {
			return false // Field doesn't exist in row
		}

		// Compare values (simplified comparison, could be extended)
		if !valuesMatch(actualValue, expectedValue) {
			return false
		}
	}

	return true // All conditions matched
}

// applyAction applies the action specified in a rule
func (g *RulesGenerator) applyAction(ctx *Context, action map[string]interface{}) (interface{}, error) {
	// Handle different action types
	if value, ok := action["value"]; ok {
		// Direct value assignment
		return value, nil
	}

	if min, hasMin := action["min"]; hasMin {
		max, hasMax := action["max"]
		if !hasMax {
			return nil, fmt.Errorf("max required when min is specified")
		}

		// Generate random value in range
		return generateInRange(ctx, min, max)
	}

	if generator, ok := action["generator"].(string); ok {
		// Use a specific generator
		config := action["config"].(map[string]interface{})
		return generateWithConfig(ctx, generator, config)
	}

	return nil, fmt.Errorf("invalid action format")
}

// valuesMatch compares two values for equality
// Supports string, numeric, and boolean comparisons
func valuesMatch(actual, expected interface{}) bool {
	// Simple type-agnostic comparison
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

// generateInRange generates a value within a specified range
func generateInRange(ctx *Context, min, max interface{}) (interface{}, error) {
	// Convert to float64 for range calculation
	minVal := toFloat64(min)
	maxVal := toFloat64(max)

	if minVal > maxVal {
		return nil, fmt.Errorf("min (%v) cannot be greater than max (%v)", min, max)
	}

	// Generate random value in range
	value := ctx.Rand.Float64()*(maxVal-minVal) + minVal

	// Return as integer if both bounds are integers
	if isInteger(min) && isInteger(max) {
		return int(value), nil
	}

	return value, nil
}

// generateWithConfig generates using a named generator with config
func generateWithConfig(ctx *Context, generatorName string, config map[string]interface{}) (interface{}, error) {
	// Get generator from registry
	registry := DefaultRegistry()
	gen, err := registry.Get(generatorName)
	if err != nil {
		return nil, fmt.Errorf("generator %s not found: %w", generatorName, err)
	}

	// Generate value
	return gen.Generate(ctx)
}

// isInteger checks if a value is an integer type
func isInteger(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}
