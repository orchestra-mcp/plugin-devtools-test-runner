package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal/runner"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestStatusSchema returns the JSON Schema for the test_status tool.
func TestStatusSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Absolute path to the project directory",
			},
			"framework": map[string]any{
				"type":        "string",
				"description": "Test framework: go, cargo, npm, pytest, or vitest. Auto-detected if omitted.",
				"enum":        []any{"go", "cargo", "npm", "pytest", "vitest"},
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// TestStatus returns a tool handler that checks the test infrastructure status.
func TestStatus() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		dir := helpers.GetString(req.Arguments, "directory")
		framework := helpers.GetStringOr(req.Arguments, "framework", "")

		detected := runner.DetectFramework(dir)
		if framework == "" {
			framework = detected
		}

		var patterns []string
		switch framework {
		case "go":
			patterns = []string{"*_test.go"}
		case "cargo":
			patterns = []string{"*.rs"}
		case "pytest", "vitest", "npm":
			patterns = []string{"*.test.ts", "*.test.tsx", "*.spec.ts", "*.spec.tsx",
				"*.test.js", "*.spec.js", "test_*.py", "*_test.py"}
		default:
			patterns = []string{"*_test.go", "*.test.ts", "*.spec.ts", "test_*.py"}
		}

		testFiles, err := findTestFiles(dir, patterns)
		if err != nil {
			return helpers.ErrorResult("scan_error", fmt.Sprintf("error scanning %s: %v", dir, err)), nil
		}

		var out strings.Builder
		fmt.Fprintf(&out, "Test infrastructure status for: %s\n\n", dir)
		fmt.Fprintf(&out, "Detected framework : %s\n", detected)
		fmt.Fprintf(&out, "Selected framework : %s\n", framework)
		fmt.Fprintf(&out, "Test files found   : %d\n", len(testFiles))

		if len(testFiles) > 0 {
			out.WriteString("\nTest files:\n")
			for _, f := range testFiles {
				rel, err := filepath.Rel(dir, f)
				if err != nil {
					rel = f
				}
				fmt.Fprintf(&out, "  %s\n", rel)
			}
		}

		return helpers.TextResult(out.String()), nil
	}
}

// findTestFiles walks the directory tree and collects files matching any of the patterns.
func findTestFiles(root string, patterns []string) ([]string, error) {
	var results []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() {
			name := info.Name()
			// Skip hidden dirs, vendor, node_modules, and target.
			if name == "vendor" || name == "node_modules" || name == "target" ||
				(len(name) > 0 && name[0] == '.') {
				return filepath.SkipDir
			}
			return nil
		}
		for _, pat := range patterns {
			matched, err := filepath.Match(pat, info.Name())
			if err == nil && matched {
				results = append(results, path)
				break
			}
		}
		return nil
	})
	return results, err
}
