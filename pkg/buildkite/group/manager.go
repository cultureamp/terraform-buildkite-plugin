// Package group provides utilities for Buildkite's logging and annotation features.
//
// # Overview
//
// The main feature is [Manager], which creates organized log output using Buildkite's
// log grouping functionality to structure build output in the Buildkite UI.
//
// # Basic Usage
//
//	gm := buildkite.NewLogGroupManager(os.Stdout)
//	gm.Open("Build Process")
//	fmt.Println("Building application...")
//
//	gm.Closed("Dependencies")
//	fmt.Println("Installing packages...")
//
//	gm.Muted("Debug Information")
//	fmt.Println("Verbose debugging output...")
//
// # Convenience Functions
//
// For convenience functions that use a global instance, see the
// [github.com/cultureamp/terraform-buildkite-plugin/pkg/buildkite/group] subpackage.
//
// # External Links
//
// See the official Buildkite documentation for more details:
// https://buildkite.com/docs/pipelines/configure/managing-log-output#collapsing-output
package group

import (
	"fmt"
	"io"
)

// Manager creates and manages Buildkite log groups.
//
// Log groups organize build output into collapsible sections in the Buildkite UI,
// making it easier to navigate through large amounts of log output.
//
// # Log Group Types
//
// Three types are supported:
//
//   - Open groups (+++): Expanded by default, content immediately visible
//   - Closed groups (---): Collapsed by default, content hidden until clicked
//   - Muted groups (~~~): Visually de-emphasized, useful for verbose/debug output
//
// Each method has a formatted version (e.g., [Manager.OpenF]) that accepts printf-style
// formatting arguments for dynamic titles.
//
// # Example
//
//	gm := NewLogGroupManager(os.Stdout)
//	gm.Open("Test Results")
//	// ... test output appears here ...
//	gm.ClosedF("Performance Metrics (%d tests)", testCount)
//	// ... performance data appears here ...
type Manager interface {
	// Open creates an expanded log group that is visible by default.
	//
	// Equivalent to: +++ <title>
	Open(title string)

	// OpenF creates an expanded log group with printf-style formatting.
	//
	// Uses [fmt.Sprintf] internally for string formatting.
	OpenF(format string, a ...any)

	// Closed creates a collapsed log group that is hidden by default.
	//
	// Users must click to expand the group. Equivalent to: --- <title>
	Closed(title string)

	// ClosedF creates a collapsed log group with printf-style formatting.
	//
	// Uses [fmt.Sprintf] internally for string formatting.
	ClosedF(format string, a ...any)

	// Muted creates a visually de-emphasized log group.
	//
	// Useful for verbose output that should be available but not prominent.
	// Equivalent to: ~~~ <title>
	Muted(title string)

	// MutedF creates a muted log group with printf-style formatting.
	//
	// Uses [fmt.Sprintf] internally for string formatting.
	MutedF(format string, a ...any)

	// OpenCurrent forces the current log group to expand.
	//
	// Useful for automatically revealing important information like errors.
	// Equivalent to: ^^^ +++
	OpenCurrent()

	// SetOutput changes the output destination for log groups.
	//
	// Pass [io.Discard] to disable log group output entirely.
	// Returns the same [GroupManager] instance for method chaining.
	SetOutput(w io.Writer) Manager
}

// config implements the GroupManager interface and holds the output destination.
type config struct {
	writer io.Writer
}

// LogGroupType represents the different types of Buildkite log groups.
//
// Each type has a specific visual behavior in the Buildkite UI.
type LogGroupType string

// Buildkite log group type constants.
//
// These constants define the prefixes used to create different types of log groups:
//
//   - [buildkiteGroupTypeOpen]: Creates expanded groups (visible by default)
//   - [buildkiteGroupTypeClosed]: Creates collapsed groups (hidden by default)
//   - [buildkiteGroupTypeMuted]: Creates de-emphasized groups (reduced prominence)
const (
	// buildkiteGroupTypeClosed creates a collapsed group that users must click to expand.
	buildkiteGroupTypeClosed LogGroupType = "---"

	// buildkiteGroupTypeOpen creates an expanded group that is visible by default.
	buildkiteGroupTypeOpen LogGroupType = "+++"

	// buildkiteGroupTypeMuted creates a visually de-emphasized group.
	buildkiteGroupTypeMuted LogGroupType = "~~~"
)

// NewLogGroupManager creates a new [GroupManager] that outputs to the specified writer.
//
// If the provided writer is nil, the manager will use [io.Discard], effectively
// disabling log group output. This is useful for testing or when log groups
// are not desired.
//
// # Examples
//
//	// Output to stdout (typical usage)
//	gm := NewLogGroupManager(os.Stdout)
//
//	// Disable log groups
//	gm := NewLogGroupManager(nil)
//
//	// Output to a custom writer
//	var buf bytes.Buffer
//	gm := NewLogGroupManager(&buf)
func NewLogGroupManager(w io.Writer) Manager {
	if w == nil {
		w = io.Discard
	}
	return &config{writer: w}
}

// group creates a Buildkite log group with the specified title and type.
//
// Outputs one of: --- <title>, +++ <title>, or ~~~ <title>
//
// See the official documentation:
// https://buildkite.com/docs/pipelines/configure/managing-log-output#collapsing-output
func group(w io.Writer, title string, groupType LogGroupType) {
	if w != nil {
		fmt.Fprintf(w, "%s %s\n", groupType, title)
	}
}

// Open creates an expanded Buildkite log group with the specified title.
//
// Open groups are expanded by default in the Buildkite UI, making their content
// immediately visible to users. This is ideal for important information that
// should be prominently displayed, such as build results or critical errors.
//
// Output format: +++ <title>
//
//	gm.Open("Build Results")
//	fmt.Println("✓ All tests passed")
//	fmt.Println("✓ Code coverage: 95%")
//
// See also [GroupManager.Closed] for collapsed groups and [GroupManager.Muted] for de-emphasized groups.
func (g *config) Open(title string) {
	group(g.writer, title, buildkiteGroupTypeOpen)
}

// OpenF creates an expanded Buildkite log group with printf-style formatting.
//
// This is a convenience method that combines [fmt.Sprintf] formatting with [GroupManager.Open].
// See [GroupManager.Open] for detailed behavior and examples.
//
//	gm.OpenF("Test Results (%d/%d passed)", passed, total)
//	gm.OpenF("Deployment to %s environment", env)
func (g *config) OpenF(format string, a ...any) {
	g.Open(fmt.Sprintf(format, a...))
}

// Closed creates a collapsed Buildkite log group with the specified title.
//
// Closed groups are collapsed by default in the Buildkite UI, hiding their content
// until a user clicks to expand them. This is useful for grouping detailed output
// that might clutter the main view but should still be accessible when needed.
//
// Output format: --- <title>
//
//	gm.Closed("Detailed Logs")
//	fmt.Println("Processing file 1 of 100...")
//	// ... more detailed output ...
//
// See also [GroupManager.Open] for expanded groups and [GroupManager.Muted] for de-emphasized groups.
func (g *config) Closed(title string) {
	group(g.writer, title, buildkiteGroupTypeClosed)
}

// ClosedF creates a collapsed Buildkite log group with printf-style formatting.
//
// This is a convenience method that combines [fmt.Sprintf] formatting with [GroupManager.Closed].
// See [GroupManager.Closed] for detailed behavior and usage patterns.
//
//	gm.ClosedF("Processing %d files", fileCount)
//	gm.ClosedF("Debug info for %s module", moduleName)
func (g *config) ClosedF(format string, a ...any) {
	g.Closed(fmt.Sprintf(format, a...))
}

// Muted creates a visually de-emphasized Buildkite log group with the specified title.
//
// Muted groups are visually de-emphasized in the Buildkite UI, appearing with
// reduced visual prominence. This is perfect for verbose output, debugging
// information, or supplementary details that should be available but not
// distract from the main build output.
//
// Output format: ~~~ <title>
//
//	gm.Muted("Verbose Debugging")
//	fmt.Println("Debug: Variable x = 42")
//	fmt.Println("Debug: Function foo() called")
//	// ... more debug output ...
//
// See also [GroupManager.Open] for expanded groups and [GroupManager.Closed] for collapsed groups.
func (g *config) Muted(title string) {
	group(g.writer, title, buildkiteGroupTypeMuted)
}

// MutedF creates a muted Buildkite log group with printf-style formatting.
//
// This is a convenience method that combines [fmt.Sprintf] formatting with [GroupManager.Muted].
// See [GroupManager.Muted] for detailed behavior and usage patterns.
//
//	gm.MutedF("Debug info (level %d)", debugLevel)
//	gm.MutedF("Performance metrics for %s", componentName)
func (g *config) MutedF(format string, a ...any) {
	g.Muted(fmt.Sprintf(format, a...))
}

// OpenCurrent forces the current Buildkite log group to expand.
//
// This method emits a special control sequence that instructs Buildkite to
// automatically expand the currently active log group. This is particularly
// useful for revealing important information dynamically, such as automatically
// expanding a group when an error occurs.
//
// Output: ^^^ +++
//
//	gm.Closed("Build Process")
//	// ... normal build output ...
//	if err != nil {
//	    gm.OpenCurrent() // Expand to show the error
//	    fmt.Printf("Error: %v\n", err)
//	}
//
// Note: This only affects the most recently created log group.
func (g *config) OpenCurrent() {
	if g.writer != nil {
		fmt.Fprintln(g.writer, "^^^ +++")
	}
}

// SetOutput changes the output destination for all subsequent log group operations.
//
// This method allows you to redirect log group output to a different writer,
// which is useful for testing, logging to files, or temporarily disabling
// log group output by setting the writer to [io.Discard].
//
// The method returns the same [GroupManager] instance to enable method chaining.
//
//	// Redirect to a file
//	file, _ := os.Create("build.log")
//	gm.SetOutput(file).Open("Build Started")
//
//	// Disable log groups temporarily
//	gm.SetOutput(io.Discard).Closed("Hidden Section")
//
//	// Re-enable to stdout
//	gm.SetOutput(os.Stdout).Open("Visible Again")
func (g *config) SetOutput(w io.Writer) Manager {
	if w == nil {
		g.writer = io.Discard
	} else {
		g.writer = w
	}
	return g
}
