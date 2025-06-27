# Configurable Variables

architectures := "amd64 arm64"
unsupported_builds := ""
platforms := "linux darwin"
output_dir := "bin"
go_test_tags := "integration,e2e"

# Computed Variables (based on the current Go environment)

go_arch := shell("go env GOARCH")
go_os := shell("go env GOOS")

# Static Variables (we dont intend people to change these)

[private]
valid_exit_code := "0"
[private]
plugin_name := "terraform-buildkite-plugin"
[private]
entrypoint_dir := "./cmd/plugin"
[private]
sample_plugin_vars := shell("jq -c . test/plugin/buildkite-plugins-var.json | jq -R")

# Default and Help Commands
[group('helper')]
default: help

# Show current variable values (for debugging)
[group('helper')]
show-vars:
    @echo "Current Variables:"
    @echo "    plugin_name: {{ BLUE }}{{ plugin_name }}{{ NORMAL }}"
    @echo "    architectures: {{ BLUE }}{{ architectures }}{{ NORMAL }}"
    @echo "    platforms: {{ BLUE }}{{ platforms }}{{ NORMAL }}"
    @echo "    output_dir: {{ BLUE }}{{ output_dir }}{{ NORMAL }}"
    @echo "    go_arch: {{ BLUE }}{{ go_arch }}{{ NORMAL }}"
    @echo "    go_os: {{ BLUE }}{{ go_os }}{{ NORMAL }}"
    @echo "    go_test_tags: {{ BLUE }}{{ go_test_tags }}{{ NORMAL }}"

# Show all available recipes
[group('helper')]
help:
    @just --list
    @echo ""
    @just show-vars

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

# Build for a specific OS/arch (internal helper)
[group('golang')]
_build target_os=go_os target_arch=go_arch:
    @CGO_ENABLED=0 \
    GOARCH={{ target_arch }} \
    GOOS={{ target_os }} \
    go build \
    -ldflags="-w -s -extldflags \"-static\"" \
    -o {{ output_dir }}/{{ plugin_name }}_{{ target_os }}_{{ target_arch }} \
    {{ entrypoint_dir }}

# Build for the current architecture
[group('golang')]
build: download _build

# Build binaries for all supported OS/arch combinations
[group('golang')]
build-all: download
    for platform in {{ platforms }}; do \
        for arch in {{ architectures }}; do \
            combo="$$platform/$$arch"; \
            if ! printf "%s\n" {{ unsupported_builds }} | grep -q "^$$combo$$"; then \
                echo "Building for $$combo"; \
                just _build "$$platform" "$$arch"; \
            else \
                echo "Skipping unsupported combo: $$combo"; \
            fi; \
        done; \
    done

# Running

# Build and run for a specific OS/arch (internal helper)
[group('golang')]
_run target_os=go_os target_arch=go_arch valid_exit_code=valid_exit_code: (_build target_os target_arch)
    ./{{ output_dir }}/{{ plugin_name }}_{{ target_os }}_{{ target_arch }} || [ "$?" -eq {{ valid_exit_code }} ]

# Run the plugin binary
[group('golang')]
run: _run

# Run the plugin in test mode with sample configuration
[group('golang')]
run-test-entry:
    @BUILDKITE_PLUGINS={{ sample_plugin_vars }} \
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
    markdownlint .

# Run shellcheck for linting bash/sh files
[group('lint')]
[group('tools')]
lint-shellcheck:
    bash -O globstar -c 'shellcheck **/*.bash ./hooks/command'

# Run actionlint for linting github action files
[group('lint')]
[group('tools')]
lint-github-actions:
    actionlint

# Run Buildkite plugin linting (requires Docker)
[group('lint')]
[group('tools')]
lint-buildkite-plugin:
    docker compose run --rm buildkite-plugin-lint

# Run all linting commands (Go, spellcheck, markdown)
[group('lint')]
lint: lint-go lint-cspell lint-markdown lint-buildkite-plugin lint-shellcheck lint-github-actions

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

# Run tests with race detector and coverage
[group('golang')]
[group('test')]
test: download
    go test -tags={{ go_test_tags }} -race -cover ./...

# Run tests and output coverage in JSON format (requires gotestfmt)
[group('golang')]
[group('test')]
test-coverage: download
    go  test -tags={{ go_test_tags }} ./... -json | gotestfmt

# Run tests for CI with atomic coverage and generate coverage reports
[group('golang')]
[group('test')]
test-ci: download
    mkdir artifacts
    go test -tags={{ go_test_tags }} ./... -covermode=atomic -coverprofile=artifacts/count.out
    go tool cover -func=artifacts/count.out | tee artifacts/coverage.out

# Run BATS script tests using CLI (primary method)
[group('bash')]
[group('test')]
test-scripts:
    bats test/scripts

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

# Run all tests (Go tests, script tests, and plugin lint)
[group('test')]
test-all: test test-scripts

# Run all quality checks: format, vet, lint, test, and script tests (used in CI)
[group('lint')]
[group('test')]
ci: fmt vet lint test-all
