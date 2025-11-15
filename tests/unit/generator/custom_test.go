package generator_test

import (
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeightedEnumGenerator(t *testing.T) {
	t.Run("generate values according to weights", func(t *testing.T) {
		// Test distribution: active (80%), inactive (15%), pending (5%)
		weights := map[string]float64{
			"active":   0.80,
			"inactive": 0.15,
			"pending":  0.05,
		}

		gen := generator.NewWeightedEnumGenerator(weights)
		ctx := generator.NewContextWithSeed(42)

		// Generate 1000 samples
		counts := make(map[string]int)
		sampleSize := 1000
		for i := 0; i < sampleSize; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)

			str, ok := val.(string)
			require.True(t, ok, "value should be string")
			counts[str]++
		}

		// Verify distribution within ±3%
		assert.InDelta(t, 800, counts["active"], 30, "active should be ~80% ±3%")
		assert.InDelta(t, 150, counts["inactive"], 30, "inactive should be ~15% ±3%")
		assert.InDelta(t, 50, counts["pending"], 30, "pending should be ~5% ±3%")
	})

	t.Run("handle equal weights", func(t *testing.T) {
		weights := map[string]float64{
			"red":   0.33,
			"green": 0.33,
			"blue":  0.34,
		}

		gen := generator.NewWeightedEnumGenerator(weights)
		ctx := generator.NewContextWithSeed(100)

		counts := make(map[string]int)
		for i := 0; i < 900; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)
			counts[val.(string)]++
		}

		// Each should be roughly equal
		assert.InDelta(t, 300, counts["red"], 50)
		assert.InDelta(t, 300, counts["green"], 50)
		assert.InDelta(t, 300, counts["blue"], 50)
	})

	t.Run("name is weighted_enum", func(t *testing.T) {
		weights := map[string]float64{"a": 1.0}
		gen := generator.NewWeightedEnumGenerator(weights)
		assert.Equal(t, "weighted_enum", gen.Name())
	})
}

func TestPatternGenerator(t *testing.T) {
	t.Run("generate values matching regex pattern", func(t *testing.T) {
		// Pattern for US phone: (XXX) XXX-XXXX
		pattern := `\(\d{3}\) \d{3}-\d{4}`
		gen := generator.NewPatternGenerator(pattern)
		ctx := generator.NewContextWithSeed(42)

		for i := 0; i < 10; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)

			str, ok := val.(string)
			require.True(t, ok)
			assert.Regexp(t, pattern, str)
		}
	})

	t.Run("generate simple alphanumeric pattern", func(t *testing.T) {
		pattern := `[A-Z]{3}-\d{4}`
		gen := generator.NewPatternGenerator(pattern)
		ctx := generator.NewContextWithSeed(100)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)

		str := val.(string)
		assert.Regexp(t, `^[A-Z]{3}-\d{4}$`, str)
		assert.Len(t, str, 8) // ABC-1234
	})

	t.Run("name is pattern", func(t *testing.T) {
		gen := generator.NewPatternGenerator(`\d+`)
		assert.Equal(t, "pattern", gen.Name())
	})
}

func TestTemplateGenerator(t *testing.T) {
	t.Run("replace year placeholder", func(t *testing.T) {
		template := "INV-{{year}}-{{seq}}"
		gen := generator.NewTemplateGenerator(template)
		ctx := generator.NewContextWithSeed(42)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)

		str := val.(string)
		assert.Contains(t, str, "INV-")
		assert.Regexp(t, `INV-\d{4}-\d+`, str)
	})

	t.Run("replace sequence placeholder", func(t *testing.T) {
		template := "ORDER-{{seq:5}}"
		gen := generator.NewTemplateGenerator(template)
		ctx := generator.NewContextWithSeed(42)

		// Generate multiple values to verify sequence
		vals := make([]string, 3)
		for i := 0; i < 3; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)
			vals[i] = val.(string)
		}

		assert.Equal(t, "ORDER-00001", vals[0])
		assert.Equal(t, "ORDER-00002", vals[1])
		assert.Equal(t, "ORDER-00003", vals[2])
	})

	t.Run("replace random placeholder", func(t *testing.T) {
		template := "USER-{{rand:8}}"
		gen := generator.NewTemplateGenerator(template)
		ctx := generator.NewContextWithSeed(42)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)

		str := val.(string)
		assert.Regexp(t, `^USER-[A-Z0-9]{8}$`, str)
	})

	t.Run("combine multiple placeholders", func(t *testing.T) {
		template := "{{year}}-{{seq:3}}-{{rand:4}}"
		gen := generator.NewTemplateGenerator(template)
		ctx := generator.NewContextWithSeed(42)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)

		str := val.(string)
		assert.Regexp(t, `^\d{4}-\d{3}-[A-Z0-9]{4}$`, str)
	})

	t.Run("name is template", func(t *testing.T) {
		gen := generator.NewTemplateGenerator("test")
		assert.Equal(t, "template", gen.Name())
	})
}

func TestIntegerRangeGenerator(t *testing.T) {
	t.Run("generate within range", func(t *testing.T) {
		gen := generator.NewIntegerRangeGenerator(10, 20)
		ctx := generator.NewContextWithSeed(42)

		for i := 0; i < 100; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)

			num, ok := val.(int64)
			require.True(t, ok)
			assert.GreaterOrEqual(t, num, int64(10))
			assert.LessOrEqual(t, num, int64(20))
		}
	})

	t.Run("handle single value range", func(t *testing.T) {
		gen := generator.NewIntegerRangeGenerator(42, 42)
		ctx := generator.NewContextWithSeed(100)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(42), val)
	})

	t.Run("name is integer_range", func(t *testing.T) {
		gen := generator.NewIntegerRangeGenerator(1, 10)
		assert.Equal(t, "integer_range", gen.Name())
	})
}