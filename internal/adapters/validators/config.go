package validators

// OpaValidation configures Open Policy Agent (OPA) policy validation.
//
// OPA validation allows enforcement of organizational policies and compliance
// rules against Terraform configurations before they are applied.
type OpaValidation struct {
	// Bundle specifies the OPA policy bundle location.
	// This can be a local file path or a URL to a remote bundle containing
	// the policies to evaluate against Terraform configurations.
	Bundle string `json:"bundle" validate:"required" jsonschema:"title=bundle,description=OPA bundle path or URL for policy validation"`

	// Query is the OPA query to evaluate.
	// This should be the fully qualified path to the policy query
	// (e.g., "data.terraform.kafka.deny") that will be evaluated.
	Query string `json:"query" validate:"required" jsonschema:"title=query,description=OPA query to evaluate"`

	Condition string `json:"condition,omitempty" jsonschema:"title=condition,description=The condition we evaluate to determine if the policy results pass or fail"`
}

// Validation contains configuration for various validation mechanisms.
//
// This struct aggregates different types of validation that can be
// performed on Terraform configurations, currently supporting OPA
// policy validation with extensibility for additional validation types.
type Validation struct {
	// Opa configures Open Policy Agent validation for Terraform configurations.
	// When configured, OPA policies will be evaluated against the Terraform
	// plan or configuration before execution.
	Opa *OpaValidation `json:"opa,omitempty" jsonschema:"title=opa,description=OPA (Open Policy Agent) validation configuration"`
}

type Config struct {
	// Validations contains a list of validation configurations.
	// Each validation will be executed against the Terraform configuration
	// before the main operation is performed.
	Validations []Validation `json:"validations,omitempty" jsonschema:"title=validations,description=A list of validation adapters"`
}

type Validations struct {
	// Validations contains a list of validation configurations.
	// Each validation will be executed against the Terraform configuration
	// before the main operation is performed.
	Validations []Validation `json:"validations,omitempty" jsonschema:"title=validations,description=A list of validation adapters"`
}
