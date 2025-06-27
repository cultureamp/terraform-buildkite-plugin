// Package group provides convenient global functions for Buildkite log grouping.
//
// This package wraps [buildkite.GroupManager] with a global instance that outputs
// to [os.Stderr], eliminating the need to manage a [buildkite.GroupManager] instance for simple use cases.
//
// # Basic Usage
//
//	group.Open("Build Process")
//	fmt.Println("Building application...")
//
//	group.Closed("Dependencies")
//	fmt.Println("Installing packages...")
//
//	group.MutedF("Debug info (level %d)", debugLevel)
//	fmt.Println("Verbose debugging output...")
//
// # Advanced Usage
//
// For advanced usage or multiple group managers, use [buildkite.NewLogGroupManager] directly.
package group

import (
	"io"
	"os"
)

// std is the global GroupManager instance used by all package-level functions.
// It outputs to os.Stderr by default. Use [SetOutput] to change the destination.
//
//nolint:gochecknoglobals // we use a global instance for convenience
var std = NewLogGroupManager(os.Stderr)

// Open creates an expanded Buildkite log group.
// See [buildkite.GroupManager.Open] for detailed documentation.
func Open(title string) {
	std.Open(title)
}

// OpenF creates an expanded Buildkite log group with printf-style formatting.
// See [buildkite.GroupManager.OpenF] for detailed documentation.
func OpenF(format string, a ...any) {
	std.OpenF(format, a...)
}

// Closed creates a collapsed Buildkite log group.
// See [buildkite.GroupManager.Closed] for detailed documentation.
func Closed(title string) {
	std.Closed(title)
}

// ClosedF creates a collapsed Buildkite log group with printf-style formatting.
// See [buildkite.GroupManager.ClosedF] for detailed documentation.
func ClosedF(format string, a ...any) {
	std.ClosedF(format, a...)
}

// Muted creates a visually de-emphasized Buildkite log group.
// See [buildkite.GroupManager.Muted] for detailed documentation.
func Muted(title string) {
	std.Muted(title)
}

// MutedF creates a visually de-emphasized Buildkite log group with printf-style formatting.
// See [buildkite.GroupManager.MutedF] for detailed documentation.
func MutedF(format string, a ...any) {
	std.MutedF(format, a...)
}

// OpenCurrent forces the current log group to expand.
// See [buildkite.GroupManager.OpenCurrent] for detailed documentation.
func OpenCurrent() {
	std.OpenCurrent()
}

// SetOutput changes the output destination for log groups.
// See [buildkite.GroupManager.SetOutput] for detailed documentation.
func SetOutput(w io.Writer) {
	std.SetOutput(w)
}
