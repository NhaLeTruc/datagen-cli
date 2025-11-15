package generator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/lucasjones/reggen"
)

// WeightedEnumGenerator generates enum values based on weighted distribution
type WeightedEnumGenerator struct {
	values       []string
	cumulWeights []float64
}

// NewWeightedEnumGenerator creates a weighted enum generator
func NewWeightedEnumGenerator(weights map[string]float64) *WeightedEnumGenerator {
	// Sort keys for deterministic behavior
	keys := make([]string, 0, len(weights))
	for k := range weights {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build cumulative distribution
	cumul := make([]float64, len(keys))
	sum := 0.0
	for i, k := range keys {
		sum += weights[k]
		cumul[i] = sum
	}

	// Normalize if sum != 1.0
	if sum != 1.0 && sum > 0 {
		for i := range cumul {
			cumul[i] /= sum
		}
	}

	return &WeightedEnumGenerator{
		values:       keys,
		cumulWeights: cumul,
	}
}

func (g *WeightedEnumGenerator) Generate(ctx *Context) (interface{}, error) {
	r := ctx.Rand.Float64()
	for i, w := range g.cumulWeights {
		if r <= w {
			return g.values[i], nil
		}
	}
	// Fallback to last value
	return g.values[len(g.values)-1], nil
}

func (g *WeightedEnumGenerator) Name() string {
	return "weighted_enum"
}

// PatternGenerator generates strings matching a regex pattern
type PatternGenerator struct {
	pattern string
	gen     *reggen.Generator
}

// NewPatternGenerator creates a pattern-based generator
func NewPatternGenerator(pattern string) *PatternGenerator {
	gen, err := reggen.NewGenerator(pattern)
	if err != nil {
		// Fallback to simple pattern if regex is invalid
		gen, _ = reggen.NewGenerator(`[A-Za-z0-9]+`)
	}

	return &PatternGenerator{
		pattern: pattern,
		gen:     gen,
	}
}

func (g *PatternGenerator) Generate(ctx *Context) (interface{}, error) {
	// Use context rand source for determinism
	g.gen.SetSeed(ctx.Rand.Int63())
	return g.gen.Generate(20), nil // max length 20
}

func (g *PatternGenerator) Name() string {
	return "pattern"
}

// TemplateGenerator generates strings from templates with placeholders
type TemplateGenerator struct {
	template string
}

// NewTemplateGenerator creates a template-based generator
func NewTemplateGenerator(template string) *TemplateGenerator {
	return &TemplateGenerator{template: template}
}

func (g *TemplateGenerator) Generate(ctx *Context) (interface{}, error) {
	result := g.template

	// Replace {{year}} with current year
	if strings.Contains(result, "{{year}}") {
		year := fmt.Sprintf("%d", time.Now().Year())
		result = strings.ReplaceAll(result, "{{year}}", year)
	}

	// Replace {{seq:N}} with zero-padded sequence
	seqPattern := regexp.MustCompile(`\{\{seq(?::(\d+))?\}\}`)
	matches := seqPattern.FindAllStringSubmatch(result, -1)
	for _, match := range matches {
		width := 1
		if len(match) > 1 && match[1] != "" {
			fmt.Sscanf(match[1], "%d", &width)
		}

		// Get sequence from context
		key := fmt.Sprintf("template_seq_%s", g.template)
		val, exists := ctx.Get(key)
		var seq int64
		if exists {
			seq = val.(int64)
		}
		seq++
		ctx.Set(key, seq)

		formatted := fmt.Sprintf("%0*d", width, seq)
		result = strings.Replace(result, match[0], formatted, 1)
	}

	// Replace {{rand:N}} with random alphanumeric string
	randPattern := regexp.MustCompile(`\{\{rand:(\d+)\}\}`)
	matches = randPattern.FindAllStringSubmatch(result, -1)
	for _, match := range matches {
		length := 8
		if len(match) > 1 {
			fmt.Sscanf(match[1], "%d", &length)
		}

		chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		randStr := make([]byte, length)
		for i := range randStr {
			randStr[i] = chars[ctx.Rand.Intn(len(chars))]
		}

		result = strings.Replace(result, match[0], string(randStr), 1)
	}

	return result, nil
}

func (g *TemplateGenerator) Name() string {
	return "template"
}

// IntegerRangeGenerator generates integers within a specified range
type IntegerRangeGenerator struct {
	min int64
	max int64
}

// NewIntegerRangeGenerator creates an integer range generator
func NewIntegerRangeGenerator(min, max int64) *IntegerRangeGenerator {
	return &IntegerRangeGenerator{min: min, max: max}
}

func (g *IntegerRangeGenerator) Generate(ctx *Context) (interface{}, error) {
	if g.min == g.max {
		return g.min, nil
	}

	// Generate random value in range [min, max]
	rangeSize := g.max - g.min + 1
	return g.min + ctx.Rand.Int63n(rangeSize), nil
}

func (g *IntegerRangeGenerator) Name() string {
	return "integer_range"
}