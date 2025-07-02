# flags

set unstable := true

# Configurable Variables

output_dir := "bin"
go_test_tags := "integration,e2e"

# Computed Variables (based on the current Go environment)

go_arch := shell("go env GOARCH")
go_os := shell("go env GOOS")
docker-check := 'command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1'

# Static Variables (we dont intend people to change these)

[private]
valid_exit_code := "0"
[private]
plugin_name := "terraform-buildkite-plugin"
[private]
sample_plugin_vars := shell("jq -c . test/plugin/buildkite-plugins-var.json | jq -R")
[private]
output_bin_base := output_dir + "/" + plugin_name
[private]
coverage_dir := "coverage"

# Helper Functions

# Default and Help Commands
[group('helper')]
default: help

# Show current variable values (for debugging)
[group('helper')]
show-vars:
    @echo "Current Variables:"
    @echo "    plugin_name: {{ BLUE }}{{ plugin_name }}{{ NORMAL }}"
    @echo "    go_arch: {{ BLUE }}{{ go_arch }}{{ NORMAL }}"
    @echo "    go_os: {{ BLUE }}{{ go_os }}{{ NORMAL }}"
    @echo "    go_test_tags: {{ BLUE }}{{ go_test_tags }}{{ NORMAL }}"

# A helper function to print a warning when Docker is not available
[private]
docker-warning command:
    @echo "{{ style("error") }}⚠️  Warning: Docker is not available or not running, so '{{ command }}' was skipped.{{ NORMAL }}"

# Show all available recipes
[group('helper')]
help:
    @just --list
    @echo ""
    @just show-vars

# Create the output directory for build artifacts
[private]
create-bin-dir:
    @mkdir -p {{ output_dir }}

# Create the artifacts directory for test outputs
[private]
create-coverage-dir:
    @mkdir -p {{ coverage_dir }}

# Dependency Management

# Download Go module dependencies
[group('golang')]
download:
    go mod download

# Verify downloaded dependencies match checksums
[group('golang')]
verify: download
    go mod verify

# Tidy Go module dependencies
[group('golang')]
tidy: download
    go mod tidy

# Ensure dependencies are tidy and there are no uncommitted changes
[group('golang')]
ensure-deps: download tidy
    @git diff --exit-code

# Building

# A private helper to build for a specific OS/arch (internal helper)
[group('golang')]
_build target_os=go_os target_arch=go_arch: create-bin-dir tidy
    GOARCH={{ target_arch }} \
    GOOS={{ target_os }} \
    goreleaser build \
    --snapshot \
    --clean \
    --single-target \
    --output {{ output_bin_base }}

# Build for the plugin binary
[group('golang')]
build: _build

# Build and release the plugin using goreleaser
[group('golang')]
release: tidy
    goreleaser release --snapshot --clean --draft

# Running

# A private helper to build and run for a specific OS/arch (internal helper)
[group('golang')]
_run target_os=go_os target_arch=go_arch valid_exit_code=valid_exit_code: (_build target_os target_arch)
    ./{{ output_bin_base }} || [ "$?" -eq {{ valid_exit_code }} ]

# Build and run for the plugin
[group('golang')]
run: _run

# Run the plugin in test mode with sample configuration
[group('golang')]
run-test-mode:
    BUILDKITE_PLUGINS={{ sample_plugin_vars }} \
    BUILDKITE_PLUGIN_{{ shoutysnakecase(plugin_name) }}_TEST_MODE=true \
    just valid_exit_code=10 run

# Code Quality
# Linting

# Run golangci-lint on the codebase
[group('golang')]
[group('lint')]
lint-go: download
    golangci-lint run ./...

# Run cspell for spell checking
[group('lint')]
[group('tools')]
lint-cspell:
    cspell **

# Run markdownlint for linting markdown files
[group('lint')]
[group('tools')]
lint-markdown:
    markdownlint-cli2

# Run shellcheck for linting bash/sh files
[group('lint')]
[group('tools')]
lint-shellcheck:
    bash -O globstar -c 'shellcheck **/*.bash ./hooks/*'

# Run actionlint for linting github action files
[group('lint')]
[group('tools')]
lint-actions:
    actionlint

# Run Buildkite plugin linting (requires Docker)
[group('lint')]
[group('tools')]
lint-buildkite-plugin:
    @{{ docker-check }} && docker compose run --rm buildkite-plugin-lint || just docker-warning buildkite-plugin-lint

# Run all linting commands (Go, spellcheck, markdown)
[group('lint')]
lint: lint-go lint-cspell lint-markdown lint-buildkite-plugin lint-shellcheck lint-actions

# Formatting and Static Analysis

# Format code using go fmt
[group('golang')]
[group('lint')]
fmt:
    go fmt ./...

# Run go vet for static analysis
[group('golang')]
[group('lint')]
vet: download
    go vet ./...

# Testing

# Run unit tests
[group('golang')]
[group('test')]
test-unit: download
    gotestsum --format pkgname --format-icons default -- ./...

# Run unit tests, integration tests and e2e tests
[group('golang')]
[group('test')]
test-full: download
    gotestsum --format pkgname --format-icons default -- -tags={{ go_test_tags }} ./...

# Run tests with atomic coverage and generate coverage reports using gotestsum and then open the report in a browser
[group('golang')]
[group('test')]
test-coverage: create-coverage-dir download
    gotestsum --format pkgname --format-icons default -- -tags=integration -race -covermode=atomic -coverprofile={{ coverage_dir }}/c.out ./...
    go tool cover -html={{ coverage_dir }}/c.out -o {{ coverage_dir }}/report.html
    open {{ coverage_dir }}/report.html

# Run BATS script tests using CLI (primary method)
[group('bash')]
[group('test')]
test-scripts:
    bats test/scripts

# Run all tests (Go tests, script tests, and plugin lint)
[group('test')]
test: test-full test-scripts

# Utilities

# Generate JSON schema (for config validation, etc.)
[group('golang')]
[group('tools')]
generate-schema:
    go run ./tools/schema/

# Remove build artifacts and coverage files
[group('golang')]
clean:
    rm -rf bin artifacts

# Run all quality checks: format, vet, lint, test, and script tests (used in CI)
[group('lint')]
[group('test')]
ci: fmt vet lint test generate-schema
