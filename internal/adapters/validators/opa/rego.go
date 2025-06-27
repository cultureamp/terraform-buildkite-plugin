// Package opa provides Open Policy Agent (OPA) validation capabilities for Terraform configurations.
//
// This package enables evaluation of OPA policies against Terraform plan data to enforce
// organizational compliance and security policies. It supports loading policy bundles from
// local files or remote URLs and executing queries against Terraform plan JSON data.
//
// Example usage:
//
//	validator := opa.NewRego(&config.OpaValidation{
//		Bundle: "/path/to/policies.tar.gz",
//		Query:  "data.terraform.deny",
//	})
//
//	violations, err := validator.Eval(ctx, terraformPlanData)
//	if err != nil {
//		log.Fatal().Err(err).Msg("Policy evaluation failed")
//	}
//
//	if len(violations) > 0 {
//		log.Error().Interface("violations", violations).Msg("Policy violations found")
//	}
package opa

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

// PolicyValidator defines the interface for evaluating OPA policies against input data.
//
// This interface abstracts the OPA policy evaluation process, allowing for different
// implementations and easier testing through mocking.
type PolicyValidator interface {
	// Eval evaluates the configured OPA policy against the provided input data.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - input: The data to evaluate against the policy (typically Terraform plan JSON)
	//
	// Returns:
	//   - A slice of violations or results from the policy evaluation
	//   - An error if the evaluation fails
	//
	// The returned slice contains policy violations or query results. An empty slice
	// indicates no violations were found (policy passed).
	Eval(ctx context.Context, input any) ([]any, error)
}

// regoEvaluator implements PolicyValidator using the Open Policy Agent.
type regoEvaluator struct {
	// rego is the prepared OPA query evaluator
	rego *rego.Rego
	// condition specifies an optional JSON path to filter results from policy evaluation
	condition string
	// bundle contains the path or URL to the OPA policy bundle
	bundle string
	// query contains the OPA query string to execute
	query string
}

// NewRego creates a new PolicyValidator configured with the provided OPA validation settings.
//
// Parameters:
//   - v: OPA validation configuration containing bundle path, query, and optional condition
//   - opts: Optional configuration functions for customizing the validator (currently unused)
//
// Returns:
//   - A configured PolicyValidator ready for policy evaluation
//
// The function loads the specified policy bundle and prepares the query for execution.
// If the bundle path is invalid or the query is malformed, subsequent Eval() calls will fail.
func NewRego(bundle, query, condition string, _ ...func(r *PolicyValidator)) PolicyValidator {
	log.Info().
		Str("bundle", bundle).
		Str("query", query).
		Str("condition", condition).
		Msg("Creating new OPA policy validator")

	cfg := &regoEvaluator{
		rego: rego.New(
			rego.Load([]string{bundle}, nil),
			rego.Query(query),
		),
		condition: condition,
		bundle:    bundle,
		query:     query,
	}
	return cfg
}

// Eval evaluates the configured OPA policy against the provided input data.
//
// This method prepares and executes the OPA query against the input data, then filters
// the results based on the configured condition (if any). The input is typically
// Terraform plan JSON data that will be evaluated against organizational policies.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The data to evaluate (typically unmarshaled Terraform plan JSON)
//
// Returns:
//   - A slice of policy violations or query results
//   - An error if preparation or evaluation fails
//
// The method logs the evaluation process and returns an empty slice if no violations
// are found, indicating the policy passed successfully.
func (r *regoEvaluator) Eval(ctx context.Context, input any) ([]any, error) {
	log.Info().
		Str("bundle", r.bundle).
		Str("query", r.query).
		Str("condition", r.condition).
		Msg("Starting OPA policy evaluation")

	// Prepare the query for evaluation
	query, err := r.rego.PrepareForEval(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Str("bundle", r.bundle).
			Str("query", r.query).
			Msg("Failed to prepare OPA query")
		return nil, fmt.Errorf("failed to prepare OPA query: %w", err)
	}

	log.Debug().Msg("OPA query prepared successfully")

	// Execute the query against the input data
	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		log.Error().
			Err(err).
			Str("query", r.query).
			Msg("Failed to evaluate OPA query")
		return nil, fmt.Errorf("failed to evaluate OPA query: %w", err)
	}

	log.Debug().
		Int("resultCount", len(results)).
		Msg("OPA query evaluation completed")

	violations := make([]any, 0)
	for _, result := range results {
		for _, expr := range result.Expressions {
			if expr.Value != nil {
				log.Debug().
					Interface("expressionValue", expr.Value).
					Msg("Processing expression result")
				var filteredValues []any
				filteredValues, err = r.filterResult(expr.Value, r.condition)
				if err != nil {
					log.Error().
						Err(err).
						Str("condition", r.condition).
						Msg("Failed to filter policy result")
					return nil, fmt.Errorf("failed to filter result with condition %s: %w", r.condition, err)
				}

				if len(filteredValues) > 0 {
					log.Debug().
						Interface("filteredValues", filteredValues).
						Int("count", len(filteredValues)).
						Msg("Found policy violations")
					violations = append(violations, filteredValues...)
				} else {
					log.Debug().Msg("No violations found in expression result")
				}
			}
		}
	}

	log.Info().
		Int("violationCount", len(violations)).
		Msg("OPA policy evaluation completed")

	return violations, nil
}

// filterResult extracts and filters values from OPA policy evaluation results.
//
// This method processes the raw policy evaluation result and optionally applies
// a JSON path condition to extract specific values. It handles both array and
// scalar results from policy evaluation.
//
// Parameters:
//   - resultValue: The raw result from OPA policy evaluation
//   - condition: Optional JSON path to extract specific values (empty string processes all)
//
// Returns:
//   - A slice of extracted values matching the condition
//   - An error if JSON marshaling or path extraction fails
//
// When condition is empty, all values from the result are returned. When condition
// is specified, only values at that JSON path are returned. Arrays are flattened
// into individual elements.
func (r *regoEvaluator) filterResult(resultValue any, condition string) ([]any, error) {
	log.Debug().
		Str("condition", condition).
		Interface("resultValue", resultValue).
		Msg("Filtering OPA evaluation result")

	values := make([]any, 0)

	// Validate that resultValue is not nil
	if resultValue == nil {
		log.Debug().Msg("Result value is nil, returning empty slice")
		return values, nil
	}

	// Marshal the result to JSON for path-based filtering
	jsonValue, err := json.Marshal(resultValue)
	if err != nil {
		log.Error().
			Err(err).
			Interface("resultValue", resultValue).
			Msg("Failed to marshal result value to JSON")
		return nil, fmt.Errorf("failed to marshal expression value to JSON: %w", err)
	}

	log.Debug().
		Str("jsonValue", string(jsonValue)).
		Msg("marshal result value to JSON")

	parse := gjson.Parse(string(jsonValue))

	// If no condition is specified, return all values
	if condition == "" {
		log.Debug().Msg("No condition specified, extracting all values")
		return r.extractAllValues(parse), nil
	}

	// Apply the condition to filter specific values
	log.Debug().
		Str("condition", condition).
		Msg("Applying condition filter")

	value := parse.Get(condition)
	if !value.Exists() {
		log.Debug().
			Str("condition", condition).
			Str("jsonValue", string(jsonValue)).
			Msg("Condition path not found in result, returning empty slice")
		return values, nil // Return empty slice instead of error for missing paths
	}

	return r.extractValues(value), nil
}

// extractAllValues extracts all values from a gjson.Result, handling arrays appropriately.
func (r *regoEvaluator) extractAllValues(parse gjson.Result) []any {
	values := make([]any, 0)

	if parse.IsArray() {
		log.Debug().Msg("Result is array, extracting individual elements")
		parse.ForEach(func(_, el gjson.Result) bool {
			values = append(values, el.Value())
			return true
		})
	} else {
		log.Debug().Msg("Result is scalar value")
		values = append(values, parse.Value())
	}

	log.Debug().
		Int("extractedCount", len(values)).
		Msg("Extracted values from result")

	return values
}

// extractValues extracts values from a gjson.Result at a specific path, handling arrays appropriately.
func (r *regoEvaluator) extractValues(value gjson.Result) []any {
	values := make([]any, 0)

	if value.IsArray() {
		log.Debug().Msg("Filtered result is array, extracting individual elements")
		value.ForEach(func(_, el gjson.Result) bool {
			values = append(values, el.Value())
			return true
		})
	} else {
		log.Debug().Msg("Filtered result is scalar value")
		values = append(values, value.Value())
	}

	log.Debug().
		Int("extractedCount", len(values)).
		Msg("Extracted values from filtered result")

	return values
}
