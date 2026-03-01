package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal/runner"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestRunFileSchema returns the JSON Schema for the test_run_file tool.
func TestRunFileSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Absolute path to the project directory",
			},
			"file": map[string]any{
				"type":        "string",
				"description": "Path to the test file (relative to directory or absolute)",
			},
			"framework": map[string]any{
				"type":        "string",
				"description": "Test framework: go, cargo, npm, pytest, or vitest. Auto-detected if omitted.",
				"enum":        []any{"go", "cargo", "npm", "pytest", "vitest"},
			},
		},
		"required": []any{"directory", "file"},
	})
	return s
}

// TestRunFile returns a tool handler that runs tests in a specific file.
func TestRunFile() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory", "file"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		dir := helpers.GetString(req.Arguments, "directory")
		file := helpers.GetString(req.Arguments, "file")
		framework := helpers.GetStringOr(req.Arguments, "framework", "")

		if framework == "" {
			framework = runner.DetectFramework(dir)
		}

		var args []string
		switch framework {
		case "go":
			// For Go, pass the package path containing the file.
			// The user should pass a package path like ./pkg/foo or a specific file.
			args = []string{"go", "test", file}
		case "cargo":
			// Cargo does not support running a single test file directly;
			// pass the file name as a filter pattern.
			args = []string{"cargo", "test", file}
		case "pytest":
			args = []string{"python", "-m", "pytest", file}
		case "vitest":
			args = []string{"npx", "vitest", "run", file}
		case "npm":
			args = []string{"npm", "test", "--", file}
		default:
			return helpers.ErrorResult("unknown_framework",
				fmt.Sprintf("cannot detect test framework in %s; specify framework explicitly", dir)), nil
		}

		out, err := runner.Run(ctx, dir, args...)
		if err != nil {
			return helpers.ErrorResult("test_failed", err.Error()), nil
		}
		return helpers.TextResult(out), nil
	}
}
