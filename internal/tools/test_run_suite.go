package tools

import (
	"context"
	"fmt"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal/runner"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestRunSuiteSchema returns the JSON Schema for the test_run_suite tool.
func TestRunSuiteSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Absolute path to the project directory",
			},
			"suite": map[string]any{
				"type":        "string",
				"description": "Test suite name or pattern to run (e.g. 'TestAuth', 'auth/**')",
			},
			"framework": map[string]any{
				"type":        "string",
				"description": "Test framework: go, cargo, npm, pytest, or vitest. Auto-detected if omitted.",
				"enum":        []any{"go", "cargo", "npm", "pytest", "vitest"},
			},
			"timeout": map[string]any{
				"type":        "integer",
				"description": "Timeout in seconds for the test run (default: 120)",
			},
		},
		"required": []any{"directory", "suite"},
	})
	return s
}

// TestRunSuite returns a handler that runs a named test suite or pattern within
// a project directory.
func TestRunSuite() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory", "suite"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		dir := helpers.GetString(req.Arguments, "directory")
		suite := helpers.GetString(req.Arguments, "suite")
		framework := helpers.GetStringOr(req.Arguments, "framework", "")

		if framework == "" {
			framework = runner.DetectFramework(dir)
		}

		var args []string
		switch framework {
		case "go":
			args = []string{"go", "test", "-v", "-run", suite, "./..."}
		case "cargo":
			args = []string{"cargo", "test", suite}
		case "pytest":
			args = []string{"python", "-m", "pytest", "-k", suite, "-v"}
		case "vitest":
			args = []string{"npx", "vitest", "run", "--reporter=verbose", suite}
		case "npm":
			// npm doesn't natively support suite filtering; pass suite as env
			args = []string{"npm", "test", "--", suite}
		default:
			return helpers.ErrorResult("unknown_framework",
				fmt.Sprintf("cannot detect test framework in %s; specify framework explicitly", dir)), nil
		}

		out, err := runner.Run(ctx, dir, args...)
		if err != nil {
			// Include partial output in the error for debugging.
			msg := strings.TrimSpace(err.Error())
			if out != "" {
				msg = out + "\n\n" + msg
			}
			return helpers.ErrorResult("test_failed", msg), nil
		}
		return helpers.TextResult(out), nil
	}
}
