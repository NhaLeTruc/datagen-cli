package generator_test

import (
	"testing"
	"time"

	"github.com/NhaLeTruc/datagen-cli/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegerGenerator(t *testing.T) {
	t.Run("generate integer values", func(t *testing.T) {
		gen := generator.NewIntegerGenerator()
		ctx := generator.NewContextWithSeed(42)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)
		require.NotNil(t, val)

		intVal, ok := val.(int64)
		assert.True(t, ok, "expected int64 type")
		assert.NotZero(t, intVal)
	})

	t.Run("deterministic with same seed", func(t *testing.T) {
		gen := generator.NewIntegerGenerator()

		ctx1 := generator.NewContextWithSeed(12345)
		val1, _ := gen.Generate(ctx1)

		ctx2 := generator.NewContextWithSeed(12345)
		val2, _ := gen.Generate(ctx2)

		assert.Equal(t, val1, val2)
	})

	t.Run("name is integer", func(t *testing.T) {
		gen := generator.NewIntegerGenerator()
		assert.Equal(t, "integer", gen.Name())
	})
}

func TestVarcharGenerator(t *testing.T) {
	t.Run("generate varchar values", func(t *testing.T) {
		gen := generator.NewVarcharGenerator(255)
		ctx := generator.NewContextWithSeed(42)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)
		require.NotNil(t, val)

		strVal, ok := val.(string)
		assert.True(t, ok, "expected string type")
		assert.NotEmpty(t, strVal)
		assert.LessOrEqual(t, len(strVal), 255)
	})

	t.Run("respects max length", func(t *testing.T) {
		gen := generator.NewVarcharGenerator(50)
		ctx := generator.NewContextWithSeed(42)

		for i := 0; i < 100; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)

			strVal := val.(string)
			assert.LessOrEqual(t, len(strVal), 50)
		}
	})

	t.Run("name is varchar", func(t *testing.T) {
		gen := generator.NewVarcharGenerator(100)
		assert.Equal(t, "varchar", gen.Name())
	})
}

func TestTextGenerator(t *testing.T) {
	t.Run("generate text values", func(t *testing.T) {
		gen := generator.NewTextGenerator()
		ctx := generator.NewContextWithSeed(42)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)
		require.NotNil(t, val)

		strVal, ok := val.(string)
		assert.True(t, ok, "expected string type")
		assert.NotEmpty(t, strVal)
	})

	t.Run("generates variable length text", func(t *testing.T) {
		gen := generator.NewTextGenerator()
		ctx := generator.NewContextWithSeed(42)

		lengths := make(map[int]bool)
		for i := 0; i < 50; i++ {
			val, _ := gen.Generate(ctx)
			strVal := val.(string)
			lengths[len(strVal)] = true
		}

		// Should have variety in lengths
		assert.Greater(t, len(lengths), 1)
	})

	t.Run("name is text", func(t *testing.T) {
		gen := generator.NewTextGenerator()
		assert.Equal(t, "text", gen.Name())
	})
}

func TestTimestampGenerator(t *testing.T) {
	t.Run("generate timestamp values", func(t *testing.T) {
		gen := generator.NewTimestampGenerator()
		ctx := generator.NewContextWithSeed(42)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)
		require.NotNil(t, val)

		timeVal, ok := val.(time.Time)
		assert.True(t, ok, "expected time.Time type")
		assert.False(t, timeVal.IsZero())
	})

	t.Run("generates timestamps within reasonable range", func(t *testing.T) {
		gen := generator.NewTimestampGenerator()
		ctx := generator.NewContextWithSeed(42)

		now := time.Now()
		pastYear := now.AddDate(-1, 0, 0)

		for i := 0; i < 50; i++ {
			val, err := gen.Generate(ctx)
			require.NoError(t, err)

			timeVal := val.(time.Time)
			// Should be within the past year and not in the future
			assert.True(t, timeVal.After(pastYear))
			assert.True(t, timeVal.Before(now.Add(time.Hour)))
		}
	})

	t.Run("name is timestamp", func(t *testing.T) {
		gen := generator.NewTimestampGenerator()
		assert.Equal(t, "timestamp", gen.Name())
	})
}

func TestBooleanGenerator(t *testing.T) {
	t.Run("generate boolean values", func(t *testing.T) {
		gen := generator.NewBooleanGenerator()
		ctx := generator.NewContextWithSeed(42)

		val, err := gen.Generate(ctx)
		require.NoError(t, err)
		require.NotNil(t, val)

		_, ok := val.(bool)
		assert.True(t, ok, "expected bool type")
	})

	t.Run("generates both true and false", func(t *testing.T) {
		gen := generator.NewBooleanGenerator()
		ctx := generator.NewContextWithSeed(42)

		trueCount := 0
		falseCount := 0

		for i := 0; i < 100; i++ {
			val, _ := gen.Generate(ctx)
			boolVal := val.(bool)
			if boolVal {
				trueCount++
			} else {
				falseCount++
			}
		}

		// Should have both values
		assert.Greater(t, trueCount, 0)
		assert.Greater(t, falseCount, 0)
	})

	t.Run("name is boolean", func(t *testing.T) {
		gen := generator.NewBooleanGenerator()
		assert.Equal(t, "boolean", gen.Name())
	})
}

func TestSerialGenerator(t *testing.T) {
	t.Run("generate serial sequence", func(t *testing.T) {
		gen := generator.NewSerialGenerator()
		ctx := generator.NewContextWithSeed(42)

		val1, err := gen.Generate(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), val1)

		val2, err := gen.Generate(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), val2)

		val3, err := gen.Generate(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(3), val3)
	})

	t.Run("each context has independent sequence", func(t *testing.T) {
		gen := generator.NewSerialGenerator()

		ctx1 := generator.NewContextWithSeed(42)
		val1, _ := gen.Generate(ctx1)
		assert.Equal(t, int64(1), val1)

		ctx2 := generator.NewContextWithSeed(43)
		val2, _ := gen.Generate(ctx2)
		assert.Equal(t, int64(1), val2)
	})

	t.Run("name is serial", func(t *testing.T) {
		gen := generator.NewSerialGenerator()
		assert.Equal(t, "serial", gen.Name())
	})
}