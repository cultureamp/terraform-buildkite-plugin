// Package outputs provides adapters for integrating existing components
// with the orchestrator interfaces.
package outputs

// ComputedVar defines a variable computed from Terraform output.
//
// Computed variables allow extraction of specific values from Terraform
// output using regular expressions, making them available for use in
// output templates and other plugin operations.
type ComputedVar struct {
	// Name is the identifier for the computed variable.
	// This name will be used to reference the variable in templates.
	Name string `json:"name" jsonschema:"title=name,description=Name of the computed variable"`

	// From specifies the source field to extract the variable from.
	// This should reference a field in the Terraform output or state.
	From string `json:"from" jsonschema:"title=from,description=Source field to extract the variable from"`

	// Regex is the regular expression used to extract the variable value.
	// The first capture group from this regex will be used as the variable value.
	Regex string `json:"regex" jsonschema:"title=regex,description=Regular expression to extract the variable value"`
}

type BuildkiteAnnotation struct {
	// Template is the template string used for formatting output.
	// This can include variable references that will be replaced with
	// actual values during output generation.
	Template string `json:"template,omitempty" jsonschema:"title=template,description=Template for formatting the output"`

	// Context provides additional context for output formatting.
	// This can be used to specify the context or environment where
	// the output will be displayed.
	Context string `json:"context,omitempty" jsonschema:"title=context,description=Context for the output formatting"`

	// Vars contains static variables for use in output templates.
	// These key-value pairs will be available for substitution in the template.
	Vars []map[string]string `json:"vars,omitempty" jsonschema:"title=vars,description=Variables to be used in output formatting"`

	// ComputedVars contains variables computed from Terraform output.
	// These variables are dynamically extracted from Terraform execution
	// results and made available for template substitution.
	ComputedVars []ComputedVar `json:"computed_vars,omitempty" jsonschema:"title=computed_vars,description=Variables computed from Terraform output"`
}

// Output configures how plugin results are formatted and presented.
//
// This struct controls the output formatting for Terraform operations,
// supporting templates, context variables, and computed values for
// flexible result presentation.
type Output struct {
	// Annotation configures OBuildkite pipeline annotation output
	BuildkiteAnnotation *BuildkiteAnnotation `json:"buildkite_annotation,omitempty" jsonschema:"title=annotation,description=Buildkite pipeline annotation configuration"`
}

type Outputs struct {
	// Output configures how plugin results are formatted and presented.
	// This controls the output format, templates, and variables used
	// for presenting Terraform operation results.
	Outputs []Output `json:"outputs,omitempty" jsonschema:"title=outputs,description=A list of output adaptors"`
}
