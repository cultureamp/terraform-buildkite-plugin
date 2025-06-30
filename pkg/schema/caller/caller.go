package caller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Caller provides a method to determine the relative path to the directory of the main.go file in the call stack.
type Caller interface {
	// CallPath returns the relative path to the directory containing main.go, relative to the current working directory.
	CallPath() (string, error)
}

// WorkingDirFunc defines a function that returns the current working directory.
type WorkingDirFunc func() (string, error)

// MatcherFn defines a function that determines if a runtime.Frame matches a desired file.
type MatcherFn func(frame runtime.Frame) bool

// FindCallerFunc defines a function that finds a runtime.Frame matching a given matcher.
type FindCallerFunc func(matcher func(frame runtime.Frame) bool) (runtime.Frame, error)

// caller implements the Caller interface and holds dependencies for path resolution.
type caller struct {
	workingDirFn WorkingDirFunc
	findCallerFn FindCallerFunc
	matcherFn    MatcherFn
}

// ConfigOption configures a caller instance.
type ConfigOption func(*caller)

// WithWorkingDirFn sets a custom function for retrieving the working directory.
func WithWorkingDirFn(fn WorkingDirFunc) ConfigOption {
	return func(g *caller) {
		g.workingDirFn = fn
	}
}

// WithFindCallerFn sets a custom function for finding the caller frame.
func WithFindCallerFn(fn FindCallerFunc) ConfigOption {
	return func(g *caller) {
		g.findCallerFn = fn
	}
}

// WithMatcherFn sets a custom matcher function for identifying the desired frame.
func WithMatcherFn(fn MatcherFn) ConfigOption {
	return func(g *caller) {
		g.matcherFn = fn
	}
}

// New creates a new Caller with optional configuration overrides.
// By default, it searches for the frame where the file is named "main.go".
func New(opts ...ConfigOption) Caller {
	g := &caller{
		findCallerFn: findCaller,
		workingDirFn: os.Getwd,
		matcherFn: func(f runtime.Frame) bool {
			return filepath.Base(f.File) == "main.go"
		},
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// PathResolver is a legacy struct for compatibility; prefer using Caller and its options.
type PathResolver struct {
	WorkingDir func() (string, error)
	FindCaller func(matcher func(frame runtime.Frame) bool) (runtime.Frame, error)
}

// CallPath returns the relative path to the directory of the main.go file, relative to the current working directory.
// It uses the configured findCallerFn, matcherFn, and workingDirFn.
func (c *caller) CallPath() (string, error) {
	// Find the frame for the main.go file
	frame, err := c.findCallerFn(c.matcherFn)
	if err != nil {
		return "", fmt.Errorf("failed to find entrypoint caller: %w", err)
	}

	// Get the current working directory
	cwd, err := c.workingDirFn()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Calculate and normalize the relative path
	relPath, err := filepath.Rel(cwd, filepath.Dir(frame.File))
	if err != nil {
		return "", fmt.Errorf("failed to make path relative: %w", err)
	}
	relPath = filepath.ToSlash(relPath)

	// Ensure the path starts with "./" or "../"
	if !strings.HasPrefix(relPath, "./") && !strings.HasPrefix(relPath, "../") {
		relPath = "./" + relPath
	}

	return relPath, nil
}

// findCaller is a helper function that finds the first frame in the call stack matching the provided matcher.
// Returns an error if no matching frame is found.
func findCaller(matcher func(frame runtime.Frame) bool) (runtime.Frame, error) {
	const maxCallerDepth = 32
	pc := make([]uintptr, maxCallerDepth)
	n := runtime.Callers(0, pc)
	pc = pc[:n]

	frames := runtime.CallersFrames(pc)

	for {
		frame, more := frames.Next()
		if matcher(frame) {
			return frame, nil
		}
		if !more {
			break
		}
	}

	return runtime.Frame{}, errors.New("no matching frame found")
}
