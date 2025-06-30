package testhelpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/otiai10/copy"
	gitignore "github.com/sabhiram/go-gitignore"
)

// shouldIncludeFile determines whether a file or directory should be included when using an includeFiles filter.
// It performs multiple checks to match files and directories flexibly:
//   - Direct filename match (e.g., "main.tf" matches "main.tf")
//   - Direct relative path match (e.g., "output/file.txt" matches "output/file.txt")
//   - Directory name match for directories (e.g., "output" matches directory named "output")
//   - Parent directory inclusion (e.g., "output/" includes all files under "output/")
//
// Parameters:
//   - srcinfo: File or directory information from os.FileInfo
//   - relPath: Relative path of the file from the source directory root
//   - includeFiles: List of files/directories to include (can be filenames or paths)
//
// Returns true if the file should be included, false otherwise.
func shouldIncludeFile(srcinfo os.FileInfo, relPath string, includeFiles []string) bool {
	fileName := srcinfo.Name()

	for _, includeFile := range includeFiles {
		// Check both the filename and the relative path
		if fileName == includeFile || relPath == includeFile {
			return true
		}
		// Also check if we're inside a directory that should be included
		if srcinfo.IsDir() && includeFile == fileName {
			return true
		}
		// Check if the current path is under an included directory
		if strings.HasPrefix(relPath, includeFile+"/") {
			return true
		}
	}
	return false
}

// shouldSkipByGitignore determines whether a file or directory should be skipped based on gitignore patterns.
// This function handles gitignore-based filtering to exclude files that should not be copied.
//
// Special behavior:
//   - Always skips the .gitignore file itself to prevent it from being copied to the destination
//   - Uses the gitignore library to match paths against ignore patterns
//   - If no gitignore is provided (nil), only skips the .gitignore file itself
//
// Parameters:
//   - relPath: Relative path of the file from the source directory root
//   - ignore: Compiled gitignore patterns (can be nil if no gitignore filtering is needed)
//
// Returns true if the file should be skipped (not copied), false if it should be included.
func shouldSkipByGitignore(relPath string, ignore *gitignore.GitIgnore) bool {
	// Always skip the .gitignore itself
	if relPath == ".gitignore" {
		return true
	}

	if ignore != nil && ignore.MatchesPath(relPath) {
		return true
	}

	return false
}

// getCopyOptions creates a skip function for use with copy.Options that handles file filtering
// based on both explicit include lists and gitignore patterns.
//
// The filtering logic follows this priority:
//  1. If includeFiles is provided, only files matching the include list are copied
//  2. If includeFiles is nil, gitignore patterns are used for filtering
//  3. The .gitignore file itself is always excluded from copying
//
// This function is designed to work with the github.com/otiai10/copy library's Options.Skip field.
//
// Parameters:
//   - srcDir: Source directory path used to calculate relative paths
//   - includeFiles: Optional list of specific files/directories to include (nil means include all except gitignored)
//   - ignore: Compiled gitignore patterns for filtering (can be nil)
//
// Returns a skip function that takes (srcinfo, src, dest) and returns (shouldSkip bool, error).
func getCopyOptions(
	srcDir string,
	includeFiles []string,
	ignore *gitignore.GitIgnore,
) func(srcinfo os.FileInfo, src, dest string) (bool, error) {
	return func(srcinfo os.FileInfo, src, _ string) (bool, error) {
		// Get relative path from srcDir
		relPath, err := filepath.Rel(srcDir, src)
		if err != nil {
			return false, err
		}

		// If we have specific files to include, only include those
		if includeFiles != nil {
			return !shouldIncludeFile(srcinfo, relPath, includeFiles), nil
		}

		// Otherwise, skip files based on gitignore patterns
		return shouldSkipByGitignore(relPath, ignore), nil
	}
}

// getGitIgnore loads and compiles a .gitignore file from the specified source directory.
// This function is a test helper that ensures the .gitignore file can be loaded and parsed correctly.
//
// The function will fail the test if:
//   - The .gitignore file cannot be found at the expected path
//   - The .gitignore file cannot be compiled due to syntax errors
//
// Parameters:
//   - t: Testing instance for helper marking and error reporting
//   - srcDir: Directory containing the .gitignore file to load
//
// Returns a compiled GitIgnore instance that can be used for path matching.
func getGitIgnore(t *testing.T, srcDir string) (*gitignore.GitIgnore, error) {
	t.Helper()
	gitignorePath := filepath.Join(srcDir, ".gitignore")
	return gitignore.CompileIgnoreFile(gitignorePath)
}

// CopyDir copies a directory from a testdata folder to a destination directory with optional filtering.
// This is a test helper function designed for setting up test fixtures with flexible file inclusion/exclusion.
//
// The function supports two filtering modes:
//  1. Explicit inclusion: When includeFiles is provided, only specified files/directories are copied
//  2. Gitignore filtering: When useGitIgnore is true, files matching .gitignore patterns are excluded
//
// Source directory resolution:
//   - The source is resolved as "{testdataDir}/{name}"
//   - The name parameter specifies which subdirectory under testdataDir to copy from
//
// Filtering behavior:
//   - If includeFiles is non-nil, only those specific files/directories are copied
//   - If includeFiles is nil and useGitIgnore is true, gitignore patterns are applied
//   - If both includeFiles is nil and useGitIgnore is false, all files are copied
//   - The .gitignore file itself is never copied to the destination
//
// Parameters:
//   - t: Testing instance for helper marking and error reporting
//   - testdataDir: Base directory containing test data (e.g., "./testdata")
//   - name: Subdirectory name under testdataDir to copy from
//   - dstDir: Destination directory path where files will be copied
//   - includeFiles: Optional list of specific files/directories to include (nil means include all)
//   - useGitIgnore: Whether to apply gitignore filtering when includeFiles is nil
//
// Returns an error if the copy operation fails, nil on success.
//
// Example usage:
//
//	// Copy only main.tf from testdata/terraform_config to /tmp/test
//	err := CopyDir(t, "./testdata", "terraform_config", "/tmp/test", []string{"main.tf"}, false)
//
//	// Copy all files except those in .gitignore from testdata/full_project to /tmp/test
//	err := CopyDir(t, "./testdata", "full_project", "/tmp/test", nil, true)
func CopyDir(
	t *testing.T,
	testdataDir, name, dstDir string,
	includeFiles []string,
	useGitIgnore bool,
) error {
	t.Helper()
	srcDir := filepath.Join(testdataDir, name)
	var ignore *gitignore.GitIgnore
	if useGitIgnore {
		var err error
		ignore, err = getGitIgnore(t, testdataDir)
		if err != nil {
			return err
		}
	}
	opts := copy.Options{
		Skip: getCopyOptions(srcDir, includeFiles, ignore),
	}

	return copy.Copy(srcDir, dstDir, opts)
}
