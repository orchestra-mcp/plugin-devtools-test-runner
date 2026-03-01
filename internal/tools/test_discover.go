package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal/runner"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestDiscoverSchema returns the JSON Schema for the test_discover tool.
func TestDiscoverSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Absolute path to the project directory to scan for tests",
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

// TestDiscover returns a handler that discovers test files and reports a
// summary of detected tests without executing them.
func TestDiscover() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
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
		case "pytest":
			patterns = []string{"test_*.py", "*_test.py"}
		case "vitest", "npm":
			patterns = []string{
				"*.test.ts", "*.test.tsx", "*.spec.ts", "*.spec.tsx",
				"*.test.js", "*.spec.js",
			}
		default:
			return helpers.ErrorResult("unknown_framework",
				fmt.Sprintf("cannot detect test framework in %s; specify framework explicitly", dir)), nil
		}

		files, err := findTestFiles(dir, patterns)
		if err != nil {
			return helpers.ErrorResult("scan_error", fmt.Sprintf("error scanning %s: %v", dir, err)), nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "## Test Discovery: %s\n\n", dir)
		fmt.Fprintf(&sb, "- **Framework**: %s\n", framework)
		fmt.Fprintf(&sb, "- **Files found**: %d\n\n", len(files))

		if len(files) == 0 {
			sb.WriteString("No test files found.\n")
		} else {
			sb.WriteString("### Test files\n")
			for _, f := range files {
				rel, err := filepath.Rel(dir, f)
				if err != nil {
					rel = f
				}
				fmt.Fprintf(&sb, "- %s\n", rel)
			}
		}

		return helpers.TextResult(sb.String()), nil
	}
}
