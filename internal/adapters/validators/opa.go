// Package validators provides adapters that bridge existing implementations
// with the new orchestrator interfaces.
package validators

import (
	"context"
	"fmt"

	"github.com/cultureamp/terraform-buildkite-plugin/internal/adapters/validators/opa"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/rs/zerolog/log"
)

// OpaValidatorAdapter adapts the existing OPA PolicyValidator to implement
// the new Validator interface required by the orchestrator.
type OpaValidatorAdapter struct {
	// policyValidator is the underlying OPA policy validator
	policyValidator opa.PolicyValidator

	// name provides a human-readable identifier for this validator
	name string

	// config stores the original validation configuration for reference
	config *OpaValidation
}

// NewOpaValidatorAdapter creates a new validator adapter for OPA validation.
//
// Parameters:
//   - validationConfig: The OPA validation configuration
//   - name: A human-readable name for this validator instance
//
// Returns:
//   - A validator that implements the orchestrator Validator interface
//
// Example:
//
//	validator := adapter.NewOpaValidatorAdapter(&config.OpaValidation{
//	    Bundle: "/path/to/policies.tar.gz",
//	    Query:  "data.terraform.violations",
//	}, "security-policies")
func NewOpaValidatorAdapter(validationConfig *OpaValidation, name string) Validator {
	if validationConfig == nil {
		log.Warn().Str("name", name).Msg("Creating OPA validator adapter with nil config")
		validationConfig = &OpaValidation{}
	}

	if name == "" {
		name = fmt.Sprintf("opa-%s", validationConfig.Query)
	}

	log.Info().
		Str("name", name).
		Str("bundle", validationConfig.Bundle).
		Str("query", validationConfig.Query).
		Str("condition", validationConfig.Condition).
		Msg("Creating OPA validator adapter")

	policyValidator := opa.NewRego(validationConfig.Bundle, validationConfig.Query, validationConfig.Condition)

	return &OpaValidatorAdapter{
		policyValidator: policyValidator,
		name:            name,
		config:          validationConfig,
	}
}

// Validate evaluates the OPA policy against the provided Terraform plan
// and converts the results to the orchestrator's ValidationResult format.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - plan: The Terraform plan to validate
//
// Returns:
//   - ValidationResult containing pass/fail status and detailed failures
//   - An error if the validation process itself fails
//
// The adapter converts OPA policy violations into structured ValidationFailure
// objects with appropriate context and details.
func (v *OpaValidatorAdapter) Validate(ctx context.Context, plan *tfjson.Plan) (ValidationResult, error) {
	log.Info().
		Str("validator", v.name).
		Msg("Starting OPA policy validation")

	// Evaluate the OPA policy against the plan
	violations, err := v.policyValidator.Eval(ctx, plan)
	if err != nil {
		log.Error().
			Err(err).
			Str("validator", v.name).
			Msg("OPA policy evaluation failed")
		return ValidationResult{}, fmt.Errorf("OPA policy evaluation failed for %s: %w", v.name, err)
	}

	// Convert violations to ValidationResult format
	result := v.convertViolationsToResult(violations)

	log.Info().
		Str("validator", v.name).
		Bool("passed", result.Passed).
		Int("violations", len(result.Failures)).
		Msg("OPA policy validation completed")

	return result, nil
}

// convertViolationsToResult converts OPA policy violations to ValidationResult format.
//
// This method handles the conversion from the generic []any violations returned
// by OPA to the structured ValidationFailure format expected by the orchestrator.
func (v *OpaValidatorAdapter) convertViolationsToResult(violations []any) ValidationResult {
	// If no violations, validation passed
	if len(violations) == 0 {
		log.Debug().
			Str("validator", v.name).
			Msg("No policy violations found - validation passed")

		return ValidationResult{
			Passed:   true,
			Failures: nil,
		}
	}

	// Convert violations to structured failures
	failures := make([]ValidationFailure, 0, len(violations))

	for i, violation := range violations {
		failure := v.convertViolationToFailure(violation, i)
		failures = append(failures, failure)
	}

	log.Debug().
		Str("validator", v.name).
		Int("violationCount", len(violations)).
		Msg("Policy violations found - validation failed")

	return ValidationResult{
		Passed:   false,
		Failures: failures,
	}
}

// convertViolationToFailure converts a single OPA violation to ValidationFailure format.
//
// This method attempts to extract structured information from the violation,
// handling both simple string violations and complex structured violations.
func (v *OpaValidatorAdapter) convertViolationToFailure(violation any, index int) ValidationFailure {
	failure := ValidationFailure{
		Type: v.config.Query,
	} // Handle different violation formats
	switch violationData := violation.(type) {
	case string:
		// Simple string violation
		failure.Message = violationData
		failure.Path = fmt.Sprintf("violation[%d]", index)

	case map[string]any:
		// Structured violation object
		failure.Message = v.extractMessage(violationData)
		failure.Path = v.extractPath(violationData)
		failure.Details = violationData

	default:
		// Unknown format - convert to string
		failure.Message = fmt.Sprintf("Policy violation: %v", violation)
		failure.Path = fmt.Sprintf("violation[%d]", index)
		failure.Details = map[string]any{"raw_violation": violation}
	}

	// Ensure we have a message
	if failure.Message == "" {
		failure.Message = fmt.Sprintf("Policy violation %d", index+1)
	}

	log.Debug().
		Str("validator", v.name).
		Str("message", failure.Message).
		Str("path", failure.Path).
		Str("type", failure.Type).
		Msg("Converted violation to failure")

	return failure
}

// extractMessage attempts to extract a human-readable message from a structured violation.
func (v *OpaValidatorAdapter) extractMessage(violation map[string]any) string {
	// Try common message fields
	for _, key := range []string{"message", "msg", "description", "error", "reason"} {
		if msg, ok := violation[key]; ok {
			if msgStr, msgOk := msg.(string); msgOk && msgStr != "" {
				return msgStr
			}
		}
	}

	// Fallback to string representation
	return fmt.Sprintf("Policy violation: %v", violation)
}

// extractPath attempts to extract a path or location from a structured violation.
func (v *OpaValidatorAdapter) extractPath(violation map[string]any) string {
	// Try common path fields
	for _, key := range []string{"path", "location", "resource", "field", "attribute"} {
		if path, ok := violation[key]; ok {
			if pathStr, pathOk := path.(string); pathOk && pathStr != "" {
				return pathStr
			}
		}
	}

	// Try to construct path from resource information
	if resource, ok := violation["resource"].(string); ok {
		if action, actionOk := violation["action"].(string); actionOk {
			return fmt.Sprintf("%s.%s", resource, action)
		}
		return resource
	}

	return ""
}
