package runner

import (
	"os"
	"path/filepath"
)

// DetectFramework returns the test framework for the given directory.
// It checks for go.mod, Cargo.toml, pytest.ini/setup.py/pyproject.toml,
// and package.json (distinguishing vitest from plain npm).
func DetectFramework(dir string) string {
	if fileExists(filepath.Join(dir, "go.mod")) {
		return "go"
	}
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		return "cargo"
	}
	if fileExists(filepath.Join(dir, "pytest.ini")) ||
		fileExists(filepath.Join(dir, "setup.py")) ||
		fileExists(filepath.Join(dir, "pyproject.toml")) {
		return "pytest"
	}
	if fileExists(filepath.Join(dir, "package.json")) {
		// Prefer vitest when vitest.config.* exists.
		matches, _ := filepath.Glob(filepath.Join(dir, "vitest.config.*"))
		if len(matches) > 0 {
			return "vitest"
		}
		return "npm"
	}
	return "unknown"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
