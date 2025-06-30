// Package common provides shared utility functions used across the Terraform Buildkite plugin.
//
// This package contains helper functions and utilities that are used by multiple
// components of the plugin, focusing on environment variable handling, file
// operations, and other common tasks.
package common

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// FetchEnv retrieves an environment variable value with a fallback default.
//
// This function provides a safe way to access environment variables with
// a default value when the variable is not set or is empty. It uses
// [os.LookupEnv] to distinguish between unset variables and variables
// set to empty strings.
//
// # Parameters
//
//   - key: The name of the environment variable to retrieve
//   - fallback: The default value to return if the variable is not set
//
// # Returns
//
// # The environment
//
// # Example
//
//	// Get log level with default
//	logLevel := FetchEnv("LOG_LEVEL", "info")
//
//	// Get required config with empty fallback
//	config := FetchEnv("BUILDKITE_PLUGINS", "")
func FetchEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Debug().Str("env", key).Msg("using fallback for unset environment variable")
	return fallback
}

// WritePrettyJSON prints structured data as pretty JSON with syntax highlighting.
//
// This function is useful for displaying plugin configuration or other data in a human-readable format.
// It marshals the data to JSON and applies syntax highlighting for improved readability in terminal environments.
//
// # Parameters
//
//   - data: The data structure to format and display
//   - w: The io.Writer to write the output to (e.g., os.Stdout)
//
// # Returns
//
// An error if JSON marshaling or syntax highlighting fails.
//
// # Example
//
//	err := WritePrettyJSON(myStruct, os.Stdout)
//	if err != nil {
//	    log.Error().Err(err).Msg("pretty print failed")
//	}
func WritePrettyJSON(data any, w io.Writer) error {
	log.Debug().Msg("pretty printing interface as JSON")
	d, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal JSON for pretty print")
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	err = quick.Highlight(w, string(d)+"\n", "json", "terminal", "github-dark")
	if err != nil {
		log.Error().Err(err).Msg("failed to highlight JSON for pretty print")
		return fmt.Errorf("failed to highlight JSON: %w", err)
	}
	return nil
}

// SetLogLevel sets the global zerolog log level from a string value.
//
// This function parses a log level string (e.g., "info", "debug", "warn") and applies it to the global zerolog logger.
// If the string is invalid, it falls back to info level and logs a warning.
//
// Supported log levels (case-insensitive):
//   - trace: Very detailed debugging information
//   - debug: Debugging information for development
//   - info: General information messages (default)
//   - warn: Warning messages for potential issues
//   - error: Error messages for failures
//   - fatal: Fatal errors that cause program termination
//   - panic: Panic-level errors with stack traces
//
// Parameters:
//   - levelStr: String representation of the desired log level
//
// Example:
//
//	common.SetLogLevel("debug")
func SetLogLevel(levelStr string) {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		log.Warn().Err(err).Msgf("Invalid log level '%s', defaulting to info", levelStr)
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
}

// IsTestMode determines if the plugin is running in test mode.
//
// This function checks for a plugin-specific environment variable to determine
// if test mode is enabled. The environment variable follows the Buildkite plugin
// convention: BUILDKITE_PLUGIN_{PLUGIN_NAME}_TEST_MODE, where the plugin name is
// converted to SCREAMING_SNAKE_CASE.
//
// Test mode is enabled if the environment variable is set to a truthy value:
//   - "true", "1", or "yes" (case-insensitive)
//
// Parameters:
//   - plugin: The plugin name (string) used to construct the environment variable name.
//
// Returns:
//   - true if test mode is enabled, false otherwise.
//
// Example:
//
//	if common.IsTestMode("terraform-buildkite-plugin") {
//	    // Test mode logic
//	}
func IsTestMode(plugin string) bool {
	name := fmt.Sprintf("BUILDKITE_PLUGIN_%s_TEST_MODE", strcase.ToScreamingSnake(plugin))
	val := FetchEnv(name, "false")
	switch strings.ToLower(val) {
	case "true", "1", "yes":
		log.Debug().Str("name", name).Str("value", val).Msgf("test mode is enabled")
		return true
	}
	return false
}

// ParseLogLevel retrieves and parses a log level from an environment variable, with a fallback default.
//
// This function looks up the specified environment variable for a log level string (e.g., "info", "debug").
// If the variable is not set or contains an invalid value, it falls back to the provided defaultLevel.
// A warning is logged if the value is invalid.
//
// Parameters:
//   - envVar: The name of the environment variable to check for the log level.
//   - defaultLevel: The zerolog.Level to use if the environment variable is unset or invalid.
//
// Returns:
//   - The resolved zerolog.Level value.
//
// Example:
//
//	level := common.ParseLogLevel("LOG_LEVEL", zerolog.InfoLevel)
func ParseLogLevel(envVar string, defaultLevel zerolog.Level) zerolog.Level {
	levelStr := FetchEnv(envVar, defaultLevel.String())
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		log.Warn().Err(err).Msgf("Invalid log level '%s', defaulting to %s", levelStr, defaultLevel.String())
		return defaultLevel
	}
	return level
}
