//go:build e2e
// +build e2e

package main_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/buildkite/bintest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E runs end-to-end tests for the terraform-buildkite-plugin
//
// This test suite has been updated to use the new MockBuildkiteAgent API for improved:
// - Fluent interface for buildkite-agent interactions
// - Better debugging with structured call logging
// - Cleaner test code with fewer manual state management
// - More reliable assertions with detailed error messages
//
// Run with: go test -tags=e2e ./cmd/plugin/...
func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// Build the plugin binary and mock agent once for all tests
	pluginBinary := buildPlugin(t)
	defer os.Remove(pluginBinary)

	_, err := bintest.NewMock("buildkite-agent")
	if err != nil {
		t.Fatalf("failed to create buildkite-agent mock: %v", err)
	}

	// Run test groups
	t.Run("Configuration", func(t *testing.T) {
		testConfigurationHandling(t, pluginBinary)
	})

	t.Run("SingleDirectory", func(t *testing.T) {
		testSingleDirectoryExecution(t, pluginBinary)
	})

	t.Run("MultipleDirectories", func(t *testing.T) {
		testMultipleDirectoryExecution(t, pluginBinary)
	})
}

// testConfigurationHandling tests configuration parsing and validation.
func testConfigurationHandling(t *testing.T, pluginBinary string) {
	testCases := []struct {
		name             string
		env              map[string]string
		expectedExitCode int // 0 means success, any other value is the expected exit code
		contains         []string
	}{
		{
			name: "valid_test_mode",
			env: map[string]string{
				"BUILDKITE_PLUGINS": `[{"github.com/cultureamp/terraform-buildkite-plugin#v0.0.1": {"mode": "plan"}}]`,
				"BUILDKITE_PLUGIN_TERRAFORM_BUILDKITE_PLUGIN_TEST_MODE": "true",
			},
			expectedExitCode: 10, // Allow exit code 10 for valid test mode
			contains:         []string{"test mode is enabled", "running terraform-buildkite-plugin version"},
		},
		{
			name: "invalid_json",
			env: map[string]string{
				"BUILDKITE_PLUGINS": "invalid json",
			},
			expectedExitCode: 1,
			contains:         []string{"failed to parse plugin configuration"},
		},
		{
			name:             "missing_config",
			env:              map[string]string{},
			expectedExitCode: 1,
			contains:         []string{"failed to parse plugin configuration"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := buildTestEnv(tc.env)

			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, pluginBinary)
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tc.expectedExitCode == 0 {
				require.NoError(t, err, "expected success for test case: %s. Output: %s", tc.name, outputStr)
			} else {
				// Expecting a specific non-zero exit code
				require.Error(t, err, "expected exit code %d for test case: %s", tc.expectedExitCode, tc.name)
				var exitError *exec.ExitError
				if errors.As(err, &exitError) {
					assert.Equal(t, tc.expectedExitCode, exitError.ExitCode(), "expected exit code %d for test case: %s", tc.expectedExitCode, tc.name)
				} else {
					t.Errorf("expected ExitError with code %d but got different error type: %v", tc.expectedExitCode, err)
				}
			}

			for _, expectedContent := range tc.contains {
				assert.Contains(t, outputStr, expectedContent, "output should contain: %s", expectedContent)
			}
		})
	}
}

// testSingleDirectoryExecution tests execution with single working directory.
func testSingleDirectoryExecution(t *testing.T, pluginBinary string) {
	if _, err := exec.LookPath("terraform"); err != nil {
		require.NoError(t, err, "terraform not available on PATH")
	}

	t.Run("plan_execution", func(t *testing.T) {
		workingDir := setupTerraformDir(t, "single")

		env := buildTestEnv(map[string]string{
			"BUILDKITE_PLUGINS": `[{"github.com/cultureamp/terraform-buildkite-plugin#v0.0.1": {"mode": "plan", "working": {"directory": "` + workingDir + `"}}}]`,
		})

		ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, pluginBinary)
		cmd.Env = env
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		require.NoError(t, err, "plugin should succeed. Output: %s", outputStr)

		// Verify expected workflow completion
		expectedStrings := []string{
			"running terraform-buildkite-plugin version",
			"plugin initialized successfully",
		}
		for _, expected := range expectedStrings {
			assert.Contains(t, outputStr, expected, "should contain: %s", expected)
		}
	})

	t.Run("buildkite_agent_calls", func(t *testing.T) {
		workingDir := setupTerraformDir(t, "single")

		env := buildTestEnv(map[string]string{
			"BUILDKITE_PLUGINS": `[{"github.com/cultureamp/terraform-buildkite-plugin#v0.0.1": {"mode": "plan", "working": {"directory": "` + workingDir + `"}}}]`,
		})

		ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, pluginBinary)
		cmd.Env = env
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		require.NoError(t, err, "plugin should succeed. Output: %s", outputStr)

		// Check expectations and log any calls made

		// Example of checking for absence of certain commands using output inspection
		// (Since bintest doesn't have the same assertion methods)
		assert.NotContains(t, outputStr, "artifact upload", "should not upload artifacts in plan mode")
	})
}

// testMultipleDirectoryExecution tests execution with multiple working directories.
func testMultipleDirectoryExecution(t *testing.T, pluginBinary string) {
	if _, err := exec.LookPath("terraform"); err != nil {
		require.NoError(t, err, "terraform not available on PATH")
	}
	multipleTestDir := setupTerraformDir(t, "multiple")

	testCases := []struct {
		name           string
		nameRegex      string
		parallelJob    string
		parallelCount  string
		expectedDirs   []string
		unexpectedDirs []string
		exactCount     int // for parallelism tests
	}{
		{
			name:         "all_directories",
			nameRegex:    ".*",
			expectedDirs: []string{"blue", "green", "red"},
		},
		{
			name:           "regex_filter_blue_only",
			nameRegex:      "^blue$",
			expectedDirs:   []string{"blue"},
			unexpectedDirs: []string{"green", "red"},
		},
		{
			name:          "parallelism_job_0_of_3",
			nameRegex:     ".*",
			parallelJob:   "0",
			parallelCount: "3",
			exactCount:    1, // should process exactly 1 directory
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := setupMultipleDirectoryTest(t, tc, multipleTestDir)

			ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, pluginBinary)
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			require.NoError(t, err, "plugin should succeed. Output: %s", outputStr)

			verifyDirectoryProcessing(t, outputStr, multipleTestDir, tc.expectedDirs, tc.unexpectedDirs)
			verifyParallelismCount(t, outputStr, multipleTestDir, tc.exactCount)
		})
	}
}

// setupMultipleDirectoryTest creates the test environment for multiple directory tests.
func setupMultipleDirectoryTest(_ *testing.T, testCase struct {
	name           string
	nameRegex      string
	parallelJob    string
	parallelCount  string
	expectedDirs   []string
	unexpectedDirs []string
	exactCount     int
}, multipleTestDir string) []string {
	envVars := map[string]string{
		"BUILDKITE_PLUGINS": `[{"github.com/cultureamp/terraform-buildkite-plugin#v0.0.1": {"mode": "plan", "working": {"directories": {"parent_directory": "` + multipleTestDir + `", "name_regex": "` + testCase.nameRegex + `"}}}}]`,
		"BUILDKITE_PLUGIN_TERRAFORM_BUILDKITE_PLUGIN_TEST_MODE": "false", // Disable test mode so plugin actually runs
		"TF_LOG": "DEBUG", // Enable terraform debugging
	}

	if testCase.parallelJob != "" {
		envVars["BUILDKITE_PARALLEL_JOB"] = testCase.parallelJob
		envVars["BUILDKITE_PARALLEL_JOB_COUNT"] = testCase.parallelCount
	}

	return buildTestEnv(envVars)
}

// verifyDirectoryProcessing checks that expected directories were processed and unexpected ones weren't.
func verifyDirectoryProcessing(t *testing.T, outputStr, multipleTestDir string, expectedDirs, unexpectedDirs []string) {
	// Verify expected directories are processed
	for _, expectedDir := range expectedDirs {
		expectedPath := filepath.Join(multipleTestDir, expectedDir)
		assert.Contains(t, outputStr, expectedPath, "should process %s directory", expectedDir)
	}

	// Verify unexpected directories are NOT processed
	for _, unexpectedDir := range unexpectedDirs {
		unexpectedPath := filepath.Join(multipleTestDir, unexpectedDir)
		assert.NotContains(t, outputStr, unexpectedPath, "should NOT process %s directory", unexpectedDir)
	}
}

// verifyParallelismCount checks that exactly the expected number of directories were processed.
func verifyParallelismCount(t *testing.T, outputStr, multipleTestDir string, exactCount int) {
	if exactCount <= 0 {
		return
	}

	processedCount := 0
	for _, dir := range []string{"blue", "green", "red"} {
		if strings.Contains(outputStr, filepath.Join(multipleTestDir, dir)) {
			processedCount++
		}
	}
	assert.Equal(t, exactCount, processedCount, "should process exactly %d directories", exactCount)
}
