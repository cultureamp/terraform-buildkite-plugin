// Package opa_test provides test utilities and setup for OPA policy validation testing.
//
// This package contains helper functions and test configuration to support comprehensive
// testing of OPA policy evaluation functionality.
package opa_test

import (
	"os"
	"testing"

	"github.com/cultureamp/terraform-buildkite-plugin/pkg/buildkite/group"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TestMain configures the test environment before running tests.
//
// This function sets up logging configuration to reduce noise during testing
// and ensures consistent test environment across all test runs.
func TestMain(m *testing.M) {
	// Configure logging to reduce noise during tests
	//nolint:reassign // sinencing the global logger to avoid output during tests
	log.Logger = zerolog.New(nil).Level(zerolog.WarnLevel)

	// Disable buildkite group output during tests
	group.SetOutput(nil)

	// Run the tests
	exitCode := m.Run()

	// Exit with the test result code
	os.Exit(exitCode)
}
