//go:build integration
// +build integration

package opa_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cultureamp/terraform-buildkite-plugin/internal/adapters/validators/opa"
	"github.com/cultureamp/terraform-buildkite-plugin/internal/testhelpers"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEval_PolicyValidation tests the complete OPA policy evaluation workflow using real policy bundles.
//
// This test validates the integration between the OPA evaluator and actual policy files,
// ensuring that both allow and deny scenarios work correctly with sample Terraform plan data.
func TestPolicyValidation(t *testing.T) {
	// Get test data paths and validate they exist

	bundlePath := testhelpers.GetWorkingDir(t, "/test/data/opa/terraform")
	samplesPath := testhelpers.GetWorkingDir(t, "/test/data/opa/samples")

	t.Run("policy validation with violations query", func(t *testing.T) {
		// Define test scenarios based on directory structure
		testScenarios := []struct {
			folderPath      string
			shouldBeAllowed bool
			description     string
		}{
			{
				folderPath:      "plans/allow",
				shouldBeAllowed: true,
				description:     "Valid resource changes that meet policy requirements",
			},
			{
				folderPath:      "plans/violation",
				shouldBeAllowed: false,
				description:     "Resource changes that violate policy requirements",
			},
		}

		for _, scenario := range testScenarios {
			t.Run(scenario.folderPath, func(t *testing.T) {
				// Get all JSON files in the scenario folder
				scenarioPath := filepath.Join(samplesPath, scenario.folderPath)
				files, err := filepath.Glob(filepath.Join(scenarioPath, "*.json"))
				require.NoError(t, err, "Failed to list files in %s", scenarioPath)
				require.NotEmpty(t, files, "No test files found in %s", scenarioPath)

				for _, file := range files {
					filename := filepath.Base(file)
					relativeFilePath := filepath.Join(scenario.folderPath, filename)

					t.Run(filename, func(t *testing.T) {
						// Load and validate the sample input file using helper
						jsonData := testhelpers.LoadSampleFile(t, samplesPath, relativeFilePath)

						// Create the OPA validator with violations query
						validator := opa.NewRego(bundlePath, "data.terraform.violations", "")

						require.NotNil(t, validator, "Failed to create OPA validator")

						// Evaluate the policy
						ctx := t.Context()
						results, evalErr := validator.Eval(ctx, jsonData)

						// The evaluation should succeed (no error) regardless of policy outcome
						require.NoError(
							t,
							evalErr,
							"OPA evaluation should not error for %s: %s",
							relativeFilePath,
							scenario.description,
						)
						assert.NotNil(t, results, "Results should not be nil for %s", relativeFilePath)

						// Verify expected policy outcome
						if scenario.shouldBeAllowed {
							assert.Empty(
								t,
								results,
								"Expected no policy violations for %s: %s",
								relativeFilePath,
								scenario.description,
							)
							log.Debug().
								Str("filename", relativeFilePath).
								Msg("Policy validation passed as expected")
						} else {
							assert.NotEmpty(t, results, "Expected policy violations for %s: %s", relativeFilePath, scenario.description)
							log.Debug().
								Str("filename", relativeFilePath).
								Interface("violations", results).
								Msg("Policy violations found as expected")
						}
					})
				}
			})
		}
	})
}

// TestIntegration_RealPolicyBundle tests integration with actual policy files.
func TestRealPolicyBundle(t *testing.T) {
	bundlePath := testhelpers.GetWorkingDir(t, "/test/data/opa/terraform")

	// Verify the policy bundle exists
	require.DirExists(t, bundlePath, "Policy bundle directory should exist")

	t.Run("policy bundle loading", func(t *testing.T) {
		validator := opa.NewRego(bundlePath, "data", "")

		// Test with minimal valid input
		input := map[string]interface{}{
			"resource_changes": []interface{}{},
		}

		results, err := validator.Eval(t.Context(), input)
		require.NoError(t, err, "Should successfully load and evaluate policy bundle")
		assert.NotNil(t, results, "Results should not be nil")
		t.Logf("Policy bundle evaluation returned %d results", len(results))
	})

	t.Run("violations query", func(t *testing.T) {
		validator := opa.NewRego(bundlePath, "data.terraform.violations", "")

		// Test with empty resource changes (should pass)
		input := map[string]interface{}{
			"resource_changes": []interface{}{},
		}

		results, err := validator.Eval(t.Context(), input)
		require.NoError(t, err, "Should successfully evaluate structured violations")
		assert.NotNil(t, results, "Results should not be nil")
		assert.Empty(t, results, "Empty resource changes should have no violations")
	})
}

// TestEval_ErrorHandling tests error scenarios and edge cases in policy evaluation.
func TestEvalErrorHandling(t *testing.T) {
	bundlePath := testhelpers.GetWorkingDir(t, "/test/data/opa/terraform")

	t.Run("invalid bundle path", func(t *testing.T) {
		validator := opa.NewRego("/nonexistent/path", "data.terraform.violations", "")

		results, err := validator.Eval(t.Context(), map[string]interface{}{})
		require.Error(t, err, "Should error with invalid bundle path")
		assert.Contains(t, err.Error(), "failed to prepare OPA query")
		assert.Nil(t, results)
	})

	t.Run("invalid query", func(t *testing.T) {
		validator := opa.NewRego(bundlePath, "invalid.query.syntax[", "")

		results, err := validator.Eval(t.Context(), map[string]interface{}{})
		require.Error(t, err, "Should error with invalid query syntax")
		assert.Contains(t, err.Error(), "failed to prepare OPA query")
		assert.Nil(t, results)
	})

	t.Run("nil input", func(t *testing.T) {
		validator := opa.NewRego(bundlePath, "data.terraform.violations", "")

		results, err := validator.Eval(t.Context(), nil)
		// Should not error with nil input - OPA can handle this
		require.NoError(t, err, "Should handle nil input gracefully")
		assert.NotNil(t, results, "Results should not be nil")
	})

	t.Run("context cancellation", func(t *testing.T) {
		validator := opa.NewRego(bundlePath, "data.terraform.violations", "")

		// Create a cancelled context
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		results, err := validator.Eval(ctx, map[string]interface{}{})
		// Depending on timing, this might succeed or fail with context cancelled
		if err != nil {
			assert.Contains(t, err.Error(), "context canceled")
		}
		// Results should be handled gracefully regardless
		t.Logf("Results with cancelled context: %v, Error: %v", results, err)
	})
}
