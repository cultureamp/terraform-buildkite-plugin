// Package terraform provides adapters for integrating existing components
// with the orchestrator interfaces.
package terraform

type InitOptions struct {
	// PluginDir specifies the directory where Terraform plugins are stored.
	PluginDir *string `json:"plugin_dir"  validate:"omitempty,dir"     jsonschema:"title=plugin_dir,description=Directory containing Terraform plugins"`
	// GetPlugins indicates whether to automatically download Terraform plugins.
	Get *bool `json:"get_plugins" validate:"omitempty,boolean" jsonschema:"title=get_plugins,description=Whether to automatically download Terraform plugins"`
}
type Options struct {
	// ExecPath specifies the path to the Terraform executable.
	ExecPath *string `json:"exec_path,omitempty"    validate:"omitempty,file" jsonschema:"title=exec_path,description=Path to the Terraform executable, defaults to a lookup in the PATH environment variable"`
	// InitOptions contains options for running `terraform init`.
	InitOptions *InitOptions `json:"init_options,omitempty"                           jsonschema:"title=init,description=Options for the terraform init command"`
}
