package testutil

import (
	"path/filepath"
	"runtime"
)

// TestdataPath returns the absolute path to a file in the testdata directory.
// It uses runtime information to locate the testdata directory relative to the
// source file location, making it work consistently across different operating
// systems and test execution contexts.
func TestdataPath(filename string) string {
	// Get the current file's directory
	_, currentFile, _, _ := runtime.Caller(0)
	// Go up to the project root (from internal/testutil/ to root)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	// Return the path to the testdata file
	return filepath.Join(projectRoot, "testdata", filename)
}
