package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal/runner"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestCoverageSchema returns the JSON Schema for the test_coverage tool.
func TestCoverageSchema() *structpb.Struct {
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

// TestCoverage returns a tool handler that runs tests with coverage reporting.
func TestCoverage() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		dir := helpers.GetString(req.Arguments, "directory")
		framework := helpers.GetStringOr(req.Arguments, "framework", "")

		if framework == "" {
			framework = runner.DetectFramework(dir)
		}

		switch framework {
		case "go":
			// Run with coverage profile, then display function-level summary.
			coverFile := filepath.Join(os.TempDir(), "coverage.out")
			out, err := runner.Run(ctx, dir,
				"go", "test", "-coverprofile="+coverFile, "./...")
			if err != nil {
				return helpers.ErrorResult("test_failed", err.Error()), nil
			}
			summary, err := runner.Run(ctx, dir,
				"go", "tool", "cover", "-func="+coverFile)
			if err != nil {
				// Return partial output even if cover tool fails.
				return helpers.TextResult(out + "\n\n(coverage summary unavailable: " + err.Error() + ")"), nil
			}
			return helpers.TextResult(out + "\n\n" + summary), nil

		case "cargo":
			out, err := runner.Run(ctx, dir, "cargo", "test")
			if err != nil {
				return helpers.ErrorResult("test_failed", err.Error()), nil
			}
			note := "\n\nNote: install cargo-tarpaulin for detailed coverage: cargo install cargo-tarpaulin"
			return helpers.TextResult(out + note), nil

		case "pytest":
			out, err := runner.Run(ctx, dir, "python", "-m", "pytest", "--cov")
			if err != nil {
				return helpers.ErrorResult("test_failed", err.Error()), nil
			}
			return helpers.TextResult(out), nil

		case "vitest":
			out, err := runner.Run(ctx, dir, "npx", "vitest", "run", "--coverage")
			if err != nil {
				return helpers.ErrorResult("test_failed", err.Error()), nil
			}
			return helpers.TextResult(out), nil

		case "npm":
			out, err := runner.Run(ctx, dir, "npm", "test")
			if err != nil {
				return helpers.ErrorResult("test_failed", err.Error()), nil
			}
			return helpers.TextResult(out), nil

		default:
			return helpers.ErrorResult("unknown_framework",
				fmt.Sprintf("cannot detect test framework in %s; specify framework explicitly", dir)), nil
		}
	}
}
