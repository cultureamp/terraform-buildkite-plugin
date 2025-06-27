package testhelpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// LoadSampleFile loads and validates a JSON sample file for testing.
func LoadSampleFile(t *testing.T, samplesPath, filename string) interface{} {
	t.Helper()

	inputFile := filepath.Join(samplesPath, filename)

	// Check that the file exists
	require.FileExists(t, inputFile, "Sample file %s should exist", filename)

	// Read the file
	inputData, err := os.ReadFile(inputFile)
	require.NoError(t, err, "Failed to read sample file %s", filename)
	require.NotEmpty(t, inputData, "Sample file %s should not be empty", filename)

	// Parse as JSON
	var jsonData interface{}
	err = json.Unmarshal(inputData, &jsonData)
	require.NoError(t, err, "Sample file %s should contain valid JSON", filename)

	t.Logf("Loaded sample file: %s (%d bytes)", filename, len(inputData))

	return jsonData
}
