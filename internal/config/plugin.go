// Package config provides configuration management for the Terraform Buildkite plugin.
//
// This package handles loading, parsing, and validating plugin configurations
// from environment variables and JSON data sources. It supports both single
// and multiple working directory configurations for Terraform operations.
package config

import (
	"github.com/cultureamp/terraform-buildkite-plugin/internal/adapters/outputs"
	"github.com/cultureamp/terraform-buildkite-plugin/internal/adapters/terraform"
	"github.com/cultureamp/terraform-buildkite-plugin/internal/adapters/validators"
	"github.com/cultureamp/terraform-buildkite-plugin/internal/adapters/workingdir"
)

type Mode string

const (
	Plan  Mode = "plan"
	Apply Mode = "apply"
)

// Plugin represents the complete configuration for a Terraform Buildkite plugin instance.
//
// This struct defines all configuration options available for the plugin,
// supporting both environment variable and JSON configuration sources.
// The validation tags ensure configuration consistency and completeness.
type Plugin struct {
	// Mode specifies the Terraform operation to perform.
	// Valid values: "plan" for planning operations, "apply" for apply operations
	Mode Mode `json:"mode" validate:"required,oneof=plan apply" jsonschema:"title=mode,description=Operation mode for the plugin (plan or apply)"`

	// Working contains configuration for the working directories
	Working *workingdir.Working `json:"working" jsonschema:"title=working,description=Configuration for the working directories containing Terraform configurations"`

	// Terraform contains options for executing Terraform commands.
	Terraform *terraform.Options `json:"terraform,omitempty" jsonschema:"title=terraform,description=Terraform execution options including plugin directory, executable path, and plugin management"`

	// Outputs defines how plugin results are formatted and presented.
	outputs.Outputs

	// Validations contains configurations for various validation mechanisms.
	validators.Validations
}
