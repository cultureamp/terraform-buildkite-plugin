package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xphir/terraform-buildkite-plugin/internal/testhelpers"
)

// buildPlugin compiles the main.go file and returns the path to the binary.
func buildPlugin(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	pluginBinary := filepath.Join(tempDir, "terraform-buildkite-plugin")

	cmd := exec.Command("go", "build", "-o", pluginBinary, "./main.go")
	cmd.Dir = filepath.Dir(".")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to build plugin: %s", string(output))

	return pluginBinary
}

// setupTerraformDir sets up terraform test directories and returns the path.
func setupTerraformDir(t *testing.T, testType string) string {
	t.Helper()

	tempDir := t.TempDir()
	var targetDir string
	var files []string

	switch testType {
	case "single":
		targetDir = tempDir
		files = []string{"main.tf"}
	case "multiple":
		targetDir = filepath.Join(tempDir, "multiple")
		files = nil // copy entire directory structure
	default:
		t.Fatalf("unknown test type: %s", testType)
	}

	err := testhelpers.CopyDir(t, "./testdata", testType, targetDir, files, testType == "multiple")
	require.NoError(t, err, "should copy test files for %s", testType)

	if testType == "multiple" {
		// Initialize terraform in each subdirectory
		subdirs := []string{"blue", "green", "red"}
		for _, subdir := range subdirs {
			subdirPath := filepath.Join(targetDir, subdir)
			initCmd := exec.Command("terraform", "init")
			initCmd.Dir = subdirPath
			initOutput, initErr := initCmd.CombinedOutput()
			require.NoError(t, initErr, "terraform init should succeed in %s: %s", subdir, string(initOutput))
		}
	} else {
		// Initialize terraform in the single directory
		initCmd := exec.Command("terraform", "init")
		initCmd.Dir = targetDir
		initOutput, initErr := initCmd.CombinedOutput()
		require.NoError(t, initErr, "terraform init should succeed: %s", string(initOutput))
	}

	return targetDir
}

// buildTestEnv builds environment variables for tests.
func buildTestEnv(customVars map[string]string) []string {
	env := []string{
		"HOME=" + os.Getenv("HOME"),
		"LOG_LEVEL=debug",
	}

	// Add mock agent to PATH if provided
	env = append(env, "PATH="+os.Getenv("PATH"))

	// Add custom environment variables
	for key, value := range customVars {
		env = append(env, key+"="+value)
	}

	return env
}
