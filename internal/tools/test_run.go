package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal/runner"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestRunSchema returns the JSON Schema for the test_run tool.
func TestRunSchema() *structpb.Struct {
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
			"file": map[string]any{
				"type":        "string",
				"description": "Optional: specific test file to run",
			},
			"test_name": map[string]any{
				"type":        "string",
				"description": "Optional: specific test name or pattern to run",
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// TestRun returns a tool handler that runs tests in the given directory.
func TestRun() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		dir := helpers.GetString(req.Arguments, "directory")
		framework := helpers.GetStringOr(req.Arguments, "framework", "")
		file := helpers.GetString(req.Arguments, "file")
		testName := helpers.GetString(req.Arguments, "test_name")

		if framework == "" {
			framework = runner.DetectFramework(dir)
		}

		var args []string
		switch framework {
		case "go":
			args = []string{"go", "test"}
			if testName != "" {
				args = append(args, "-run", testName)
			}
			if file != "" {
				args = append(args, file)
			} else {
				args = append(args, "./...")
			}
		case "cargo":
			args = []string{"cargo", "test"}
			if testName != "" {
				args = append(args, testName)
			}
		case "vitest":
			args = []string{"npx", "vitest", "run"}
			if file != "" {
				args = append(args, file)
			}
		case "npm":
			args = []string{"npm", "test"}
		case "pytest":
			args = []string{"python", "-m", "pytest"}
			if file != "" {
				args = append(args, file)
			}
			if testName != "" {
				args = append(args, "-k", testName)
			}
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
