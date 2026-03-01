package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal/runner"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestListSchema returns the JSON Schema for the test_list tool.
func TestListSchema() *structpb.Struct {
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

// TestList returns a tool handler that lists all available tests.
func TestList() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		dir := helpers.GetString(req.Arguments, "directory")
		framework := helpers.GetStringOr(req.Arguments, "framework", "")

		if framework == "" {
			framework = runner.DetectFramework(dir)
		}

		var args []string
		switch framework {
		case "go":
			args = []string{"go", "test", "-list", ".*", "./..."}
		case "cargo":
			args = []string{"cargo", "test", "--", "--list"}
		case "pytest":
			args = []string{"python", "-m", "pytest", "--collect-only", "-q"}
		case "vitest":
			args = []string{"npx", "vitest", "list"}
		case "npm":
			// npm test --list-tests is not standard; fall back to a notice.
			return helpers.TextResult("Listing individual tests is not supported for plain npm. Use vitest or pytest frameworks."), nil
		default:
			return helpers.ErrorResult("unknown_framework",
				fmt.Sprintf("cannot detect test framework in %s; specify framework explicitly", dir)), nil
		}

		out, err := runner.Run(ctx, dir, args...)
		if err != nil {
			return helpers.ErrorResult("list_failed", err.Error()), nil
		}
		return helpers.TextResult(out), nil
	}
}
