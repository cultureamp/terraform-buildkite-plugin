package testhelpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// FindModuleRoot locates the root of the Go module by finding the directory containing the go.mod file.
func FindModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err, "get working directory")
	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // hit root `/`
		}
		dir = parent
	}
	require.FailNow(t, "could not find go.mod from current directory upwards")
	return ""
}

// GetWorkingDir returns the absolute path to the test working directory.
func GetWorkingDir(t *testing.T, subPath string) string {
	t.Helper()
	moduleRoot := FindModuleRoot(t)
	return filepath.Join(moduleRoot, subPath)
}
