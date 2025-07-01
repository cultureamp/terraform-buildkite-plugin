# Terraform Buildkite Plugin

> [!CAUTION]
> This plugin is still in development and is not yet suitable for production usage.

A Buildkite plugin for processing Terraform working directories, enabling you to perform operations such as plan and
apply across your infrastructure. Features include support for looping over multiple working directories, Open Policy
Agent validation checks against Terraform plans, and rich Buildkite annotations that detail the success or failure of
operations.

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

## Contributing

Interested in contributing? Please see our [Contributing Guide](./.github/CONTRIBUTING.md) for development setup,
coding standards, and workflow information.

## Example

Add the following lines to your `pipeline.yml`:

```yml
steps:
  - label: ":terraform: Plan infrastructure"
    plugins:
      - cultureamp/terraform#v0.1.0:
          mode: plan
          working:
            directories:
              parent_directory: ./terraform
              name_regex: ".*"
          validations:
            - opa:
                bundle: ./policies
                query: "data.terraform.allow"
          outputs:
            - buildkite_annotation:
                template: "Plan completed for {{.WorkingDirectory}}"
                context: terraform-plan
```

## Configuration

### `mode` (Required, string)

Operation mode for the plugin. Supported values:

- `plan` - Run terraform plan
- `apply` - Run terraform apply

### `working` (Required, object)

Configuration for the working directories containing Terraform configurations.

#### `working.directories` (Required, object)

Configuration for multiple working directories:

- `parent_directory` (string) - Parent directory containing Terraform configurations
- `name_regex` (string) - Regular expression to filter directory names
- `artifact` (string) - Artifact path containing Terraform configurations

#### `working.directory` (string)

Single working directory path (alternative to `directories`).

### `validations` (Optional, array)

List of validation adapters:

#### `validations[].opa` (object)

OPA (Open Policy Agent) validation configuration:

- `bundle` (Required, string) - OPA bundle path or URL for policy validation
- `query` (Required, string) - OPA query to evaluate
- `condition` (string) - Condition to determine if policy results pass or fail

### `outputs` (Optional, array)

List of output adaptors:

#### `outputs[].buildkite_annotation` (object)

Buildkite pipeline annotation configuration:

- `template` (string) - Template for formatting the output
- `context` (string) - Context for the output formatting
- `vars` (array) - Variables to be used in output formatting
- `computed_vars` (array) - Variables computed from Terraform output

### `terraform` (Optional, object)

Terraform execution options:

- `exec_path` (string) - Path to the Terraform executable
- `init_options` (object) - Options for terraform init command
  - `plugin_dir` (Required, string) - Directory containing Terraform plugins
  - `get_plugins` (Required, boolean) - Whether to automatically download plugins

## Placeholder

Some change needs more change
