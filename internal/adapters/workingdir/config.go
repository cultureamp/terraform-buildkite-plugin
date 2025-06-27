package workingdir

import (
	"github.com/invopop/jsonschema"
)

// Parallelism contains Buildkite parallel job information.
//
// This struct captures the parallel execution context provided by Buildkite
// when running jobs in parallel mode, allowing the plugin to coordinate
// work across multiple parallel job instances.
type Parallelism struct {
	// ParallelJob is the zero-based index of the current parallel job.
	// For example, in a 3-job parallel build, this would be 0, 1, or 2.
	ParallelJob *int `json:"parallel_job" env:"BUILDKITE_PARALLEL_JOB" validate:"omitempty,required_with=ParallelJobCount,ltefield=ParallelJobCount"`

	// ParallelJobCount is the total number of parallel jobs in the build.
	// This allows the plugin to understand the total parallelism context.
	ParallelJobCount *int `json:"parallel_job_count" env:"BUILDKITE_PARALLEL_JOB_COUNT" validate:"omitempty,required_with=ParallelJob"`
}

// Directories configures multiple Terraform working directory discovery.
//
// This struct supports two mutually exclusive modes for locating Terraform
// configurations: directory-based discovery and artifact-based extraction.
// The validation tags ensure only one mode is configured at a time.
type Directories struct {
	// ParentDirectory is the root directory containing Terraform configurations.
	// The plugin will search this directory (and subdirectories) for Terraform files.
	// Cannot be used together with Artifact.
	ParentDirectory string `json:"parent_directory,omitempty" validate:"dir,excluded_with=Artifact" jsonschema:"title=parent_directory,description=Parent directory containing Terraform configurations"`

	// Artifact specifies a path to an artifact containing Terraform configurations.
	// The plugin will extract and process Terraform files from this artifact.
	// Cannot be used together with ParentDirectory.
	Artifact string `json:"artifact,omitempty" validate:"omitempty,file,excluded_with=ParentDirectory" jsonschema:"title=artifact,description=Artifact path containing Terraform configurations"`

	// NameRegex is an optional regular expression to filter directory names.
	// When specified, only directories matching this pattern will be processed.
	NameRegex string `json:"name_regex,omitempty" jsonschema:"title=name_regex,description=Regular expression to filter directory names"`
}

type Working struct {
	// WorkingDirectory specifies a single Terraform working directory.
	// This is mutually exclusive with WorkingDirectories for multiple directory support.
	Directory *string `json:"directory,omitempty" validate:"omitempty,dir,excluded_with=Directories" jsonschema:"title=directory,description=Single working directory path"`

	// WorkingDirectories configures multiple working directory discovery.
	// This is mutually exclusive with WorkingDirectory for single directory mode.
	Directories *Directories `json:"directories" validate:"omitempty,excluded_with=Directory" jsonschema:"title=directories,description=Configuration for multiple working directories"`

	// Parallelism contains Buildkite parallel job context information.
	// This is automatically populated from Buildkite environment variables
	// and is not typically set via JSON configuration.
	Parallelism *Parallelism `json:"parallelism" jsonschema:"-"`
}

// JSONSchemaExtend adds oneOf constraint to ensure exactly one of Directory or Directories is required.
func (w *Working) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.OneOf = []*jsonschema.Schema{
		{
			Required: []string{"directory"},
			Not: &jsonschema.Schema{
				Required: []string{"directories"},
			},
		},
		{
			Required: []string{"directories"},
			Not: &jsonschema.Schema{
				Required: []string{"directory"},
			},
		},
	}
}
