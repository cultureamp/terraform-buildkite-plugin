# Contributing

Thank you for your interest in contributing to the Terraform Buildkite Plugin!

Before opening a pull request, please review and follow the guidelines outlined in this document.

If you plan to submit a pull request, we ask that you first create an
[issue](https://github.com/cultureamp/terraform-buildkite-plugin/issues). For new features or
modifications to existing functionality, please start a discussion with the maintainers. For
straightforward bug fixes, an issue is enough without a preliminary discussion.

## Development Setup

This project uses [devbox](https://github.com/jetpack-io/devbox) for managing Go and CLI tools,
and [just](https://github.com/casey/just) for task automation.

### Prerequisites

- [devbox](https://github.com/jetpack-io/devbox) - for managing Go and all CLI tools
- [direnv](https://direnv.net/) - for loading devbox and managing secrets
- [docker](https://www.docker.com/) - for the buildkite plugin linter

### Getting Started

1. Fork this repository and clone your fork locally
2. Install prerequisites (follow the [devbox quickstart guide](https://www.jetify.com/docs/devbox/quickstart/))
3. Set up the development environment:

   ```bash
   direnv allow .
   ```

4. Install dependencies:

   ```bash
   just download
   ```

5. View available commands:

   ```bash
   just
   ```

## Development Workflow

1. Create a new branch for your changes
2. Make your changes following our [coding standards](#coding-standards)
3. Add tests for your changes (see [Testing](#testing))
4. Ensure all checks pass (see [Quality Checks](#quality-checks))
5. Update documentation if necessary
6. Commit using [Conventional Commits](https://www.conventionalcommits.org/) style
7. Open a pull request

## Testing

All changes must include appropriate tests. This project has multiple test types:

### Running Tests

```bash
# Run unit tests
just test

# Run script tests
just test-scripts

# Run all tests (unit + scripts)
just test-all
```

### Test Requirements

- Unit tests for new functionality
- Integration tests for plugin behavior
- End-to-end tests for complete workflows
- All tests must pass before merging

## Quality Checks

Before submitting a pull request, ensure all quality checks pass:

### Complete Validation

Run the full suite of checks (recommended):

```bash
# Run all quality checks: format, vet, lint, and all tests
just ci
```

### Individual Checks

You can also run individual checks:

### Code Formatting

```bash
# Format Go code
just fmt

# Lint code
just lint
```

### Plugin Validation

```bash
# Lint the plugin configuration
just lint-buildkite-plugin
```

## Coding Standards

- Follow Go conventions and the [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- Use `just fmt` for formatting
- Ensure code passes `just vet`
- Add appropriate error handling
- Write clear, descriptive commit messages
- Update documentation for user-facing changes
- Program to interfaces, not implementations (accept interfaces, return structs)
- Use dependency injection to make components testable and loosely coupled
- Structure code so dependencies are injectable rather than hardcoded

## Plugin Development

When working on plugin functionality:

- Test with both `plan` and `apply` modes
- Verify compatibility with multiple Terraform versions
- Test with various working directory configurations
- Validate OPA policy integration if applicable
- Ensure Buildkite annotations render correctly

## Documentation

Update documentation when making changes:

- Update `README.md` for user-facing features
- Update `plugin.yml` schema for configuration changes
- Add inline code documentation for complex logic
- Update examples if configuration changes

### Schema Updates

If you modify the plugin configuration structure, regenerate the schema:

```bash
just generate-schema
```

This ensures the `plugin.yml` schema stays in sync with the Go configuration structs.

## Releasing

Releases are handled by maintainers:

1. Merge approved pull requests
2. Update version references in documentation
3. Create and push a version tag (triggers automated release)

```bash
git tag v0.1.0
git push --tags
```

## Getting Help

- Check existing [issues](https://github.com/cultureamp/terraform-buildkite-plugin/issues)
- Review the [README](https://github.com/cultureamp/terraform-buildkite-plugin/blob/main/README.md)
- Ask questions using the question issue template
