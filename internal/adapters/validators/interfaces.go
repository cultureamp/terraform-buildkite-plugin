package validators

import (
	"context"

	tfjson "github.com/hashicorp/terraform-json"
)

// Validator defines the interface for validating Terraform plans.
//
// Implementations should evaluate the provided plan against organizational
// policies, security requirements, or compliance rules.
type Validator interface {
	// Validate evaluates a Terraform plan and returns validation results.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - plan: The Terraform plan to validate
	//
	// Returns:
	//   - ValidationResult containing the outcome and any failures
	//   - An error if validation cannot be performed
	Validate(ctx context.Context, plan *tfjson.Plan) (ValidationResult, error)
}

// ValidationFailure represents a single validation failure with detailed information.
type ValidationFailure struct {
	// Type categorizes the failure (e.g., "policy", "security", "compliance")
	Type string `json:"type"`

	// Message provides a human-readable description of the failure
	Message string `json:"message"`

	// Path specifies the location in the configuration where the failure occurred
	Path string `json:"path,omitempty"`

	// Details contains additional structured information about the failure
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationResult aggregates the outcome of validation operations.
type ValidationResult struct {
	// Passed indicates whether validation was successful
	Passed bool `json:"passed"`

	// Failures contains detailed information about any validation failures
	Failures []ValidationFailure `json:"failures"`
}
