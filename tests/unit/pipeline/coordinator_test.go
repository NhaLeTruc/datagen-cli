package pipeline_test

import (
	"testing"

	"github.com/NhaLeTruc/datagen-cli/internal/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemanticRegistration(t *testing.T) {
	t.Run("register semantic generators", func(t *testing.T) {
		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterSemanticGenerators()

		// Verify semantic generators are registered
		registry := coordinator.GetRegistry()

		_, err := registry.Get("email")
		require.NoError(t, err)

		_, err = registry.Get("phone")
		require.NoError(t, err)

		_, err = registry.Get("first_name")
		require.NoError(t, err)

		_, err = registry.Get("last_name")
		require.NoError(t, err)

		_, err = registry.Get("full_name")
		require.NoError(t, err)

		_, err = registry.Get("address")
		require.NoError(t, err)

		_, err = registry.Get("city")
		require.NoError(t, err)

		_, err = registry.Get("country")
		require.NoError(t, err)

		_, err = registry.Get("postal_code")
		require.NoError(t, err)

		_, err = registry.Get("created_at")
		require.NoError(t, err)

		_, err = registry.Get("updated_at")
		require.NoError(t, err)
	})

	t.Run("semantic and basic generators coexist", func(t *testing.T) {
		coordinator := pipeline.NewCoordinator()
		coordinator.RegisterBasicGenerators()
		coordinator.RegisterSemanticGenerators()

		registry := coordinator.GetRegistry()

		// Both types available
		_, err := registry.Get("integer")
		assert.NoError(t, err)

		_, err = registry.Get("email")
		assert.NoError(t, err)
	})
}