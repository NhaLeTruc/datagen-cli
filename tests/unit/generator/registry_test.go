package generator_test

import (
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockGenerator implements Generator interface for testing
type MockGenerator struct {
	name string
}

func (m *MockGenerator) Generate(ctx *generator.Context) (interface{}, error) {
	return "mock_value", nil
}

func (m *MockGenerator) Name() string {
	return m.name
}

func TestRegistry(t *testing.T) {
	t.Run("create new registry", func(t *testing.T) {
		reg := generator.NewRegistry()
		require.NotNil(t, reg)
	})

	t.Run("register and get generator", func(t *testing.T) {
		reg := generator.NewRegistry()
		mock := &MockGenerator{name: "test_gen"}

		err := reg.Register("test_gen", mock)
		require.NoError(t, err)

		gen, err := reg.Get("test_gen")
		require.NoError(t, err)
		assert.Equal(t, mock, gen)
	})

	t.Run("get non-existent generator returns error", func(t *testing.T) {
		reg := generator.NewRegistry()

		gen, err := reg.Get("non_existent")
		assert.Error(t, err)
		assert.Nil(t, gen)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("register duplicate generator returns error", func(t *testing.T) {
		reg := generator.NewRegistry()
		mock1 := &MockGenerator{name: "duplicate"}
		mock2 := &MockGenerator{name: "duplicate"}

		err := reg.Register("duplicate", mock1)
		require.NoError(t, err)

		err = reg.Register("duplicate", mock2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("list all generators", func(t *testing.T) {
		reg := generator.NewRegistry()

		reg.Register("gen1", &MockGenerator{name: "gen1"})
		reg.Register("gen2", &MockGenerator{name: "gen2"})
		reg.Register("gen3", &MockGenerator{name: "gen3"})

		names := reg.List()
		assert.Len(t, names, 3)
		assert.Contains(t, names, "gen1")
		assert.Contains(t, names, "gen2")
		assert.Contains(t, names, "gen3")
	})

	t.Run("has checks if generator exists", func(t *testing.T) {
		reg := generator.NewRegistry()
		reg.Register("exists", &MockGenerator{name: "exists"})

		assert.True(t, reg.Has("exists"))
		assert.False(t, reg.Has("does_not_exist"))
	})
}

func TestRegistryConcurrency(t *testing.T) {
	t.Run("concurrent register and get", func(t *testing.T) {
		reg := generator.NewRegistry()

		// Register some generators concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				name := string(rune('a' + id))
				reg.Register(name, &MockGenerator{name: name})
				done <- true
			}(i)
		}

		// Wait for all registrations
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all generators can be retrieved
		for i := 0; i < 10; i++ {
			name := string(rune('a' + i))
			gen, err := reg.Get(name)
			assert.NoError(t, err)
			assert.NotNil(t, gen)
		}
	})
}

func TestGlobalRegistry(t *testing.T) {
	t.Run("global registry is accessible", func(t *testing.T) {
		global := generator.DefaultRegistry()
		require.NotNil(t, global)
	})

	t.Run("can register to global registry", func(t *testing.T) {
		global := generator.DefaultRegistry()
		mock := &MockGenerator{name: "global_test"}

		err := global.Register("global_test", mock)
		require.NoError(t, err)

		gen, err := global.Get("global_test")
		require.NoError(t, err)
		assert.Equal(t, mock, gen)
	})
}