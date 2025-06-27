package opa_test

import (
	"testing"

	"github.com/cultureamp/terraform-buildkite-plugin/internal/adapters/validators/opa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRego_Configuration tests the validator creation and configuration.
func TestNewRegoConfiguration(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		validator := opa.NewRego("/path/to/bundle", "data.example.policy", "violations")
		assert.NotNil(t, validator, "Should create validator with valid config")

		// Verify the validator implements the expected interface
		assert.Implements(t, (*opa.PolicyValidator)(nil), validator, "Should implement PolicyValidator interface")
	})

	t.Run("empty configuration", func(t *testing.T) {
		validator := opa.NewRego("", "", "")
		assert.NotNil(t, validator, "Should create validator even with empty config")

		// The validator should be created but fail on evaluation
		results, err := validator.Eval(t.Context(), map[string]interface{}{})
		require.Error(t, err, "Should error with empty bundle and query")
		assert.Nil(t, results)
	})
}
