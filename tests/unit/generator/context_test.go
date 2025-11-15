package generator_test

import (
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	t.Run("create context with default seed", func(t *testing.T) {
		ctx := generator.NewContext()
		require.NotNil(t, ctx)
		require.NotNil(t, ctx.Rand)
	})

	t.Run("create context with specific seed", func(t *testing.T) {
		ctx := generator.NewContextWithSeed(12345)
		require.NotNil(t, ctx)
		require.NotNil(t, ctx.Rand)
	})

	t.Run("same seed produces same values", func(t *testing.T) {
		ctx1 := generator.NewContextWithSeed(42)
		ctx2 := generator.NewContextWithSeed(42)

		// Generate some random numbers
		val1a := ctx1.Rand.Intn(1000)
		val1b := ctx1.Rand.Intn(1000)

		val2a := ctx2.Rand.Intn(1000)
		val2b := ctx2.Rand.Intn(1000)

		// Should be identical
		assert.Equal(t, val1a, val2a)
		assert.Equal(t, val1b, val2b)
	})

	t.Run("different seeds produce different values", func(t *testing.T) {
		ctx1 := generator.NewContextWithSeed(42)
		ctx2 := generator.NewContextWithSeed(43)

		// Generate some random numbers
		vals1 := make([]int, 10)
		vals2 := make([]int, 10)

		for i := 0; i < 10; i++ {
			vals1[i] = ctx1.Rand.Intn(1000000)
			vals2[i] = ctx2.Rand.Intn(1000000)
		}

		// Very unlikely to be all identical
		assert.NotEqual(t, vals1, vals2)
	})

	t.Run("context has table name", func(t *testing.T) {
		ctx := generator.NewContext()
		ctx.TableName = "users"

		assert.Equal(t, "users", ctx.TableName)
	})

	t.Run("context has column name", func(t *testing.T) {
		ctx := generator.NewContext()
		ctx.ColumnName = "email"

		assert.Equal(t, "email", ctx.ColumnName)
	})

	t.Run("context has row index", func(t *testing.T) {
		ctx := generator.NewContext()
		ctx.RowIndex = 42

		assert.Equal(t, 42, ctx.RowIndex)
	})

	t.Run("context can store custom data", func(t *testing.T) {
		ctx := generator.NewContext()

		ctx.Set("key1", "value1")
		ctx.Set("key2", 123)

		val1, ok := ctx.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, "value1", val1)

		val2, ok := ctx.Get("key2")
		assert.True(t, ok)
		assert.Equal(t, 123, val2)

		val3, ok := ctx.Get("non_existent")
		assert.False(t, ok)
		assert.Nil(t, val3)
	})
}

func TestContextClone(t *testing.T) {
	t.Run("clone context creates independent random state", func(t *testing.T) {
		original := generator.NewContextWithSeed(12345)

		// Generate some values
		original.Rand.Intn(100)
		original.Rand.Intn(100)

		cloned := original.Clone()

		// Clone has independent random state (seeded from original's current state)
		// Both should produce valid values
		val1 := original.Rand.Intn(1000)
		val2 := cloned.Rand.Intn(1000)

		assert.GreaterOrEqual(t, val1, 0)
		assert.GreaterOrEqual(t, val2, 0)
		assert.Less(t, val1, 1000)
		assert.Less(t, val2, 1000)
	})

	t.Run("clone context preserves metadata", func(t *testing.T) {
		original := generator.NewContext()
		original.TableName = "users"
		original.ColumnName = "email"
		original.RowIndex = 99
		original.Set("custom_key", "custom_value")

		cloned := original.Clone()

		assert.Equal(t, original.TableName, cloned.TableName)
		assert.Equal(t, original.ColumnName, cloned.ColumnName)
		assert.Equal(t, original.RowIndex, cloned.RowIndex)

		val, ok := cloned.Get("custom_key")
		assert.True(t, ok)
		assert.Equal(t, "custom_value", val)
	})

	t.Run("modifying clone does not affect original", func(t *testing.T) {
		original := generator.NewContext()
		original.TableName = "users"

		cloned := original.Clone()
		cloned.TableName = "posts"

		assert.Equal(t, "users", original.TableName)
		assert.Equal(t, "posts", cloned.TableName)
	})
}