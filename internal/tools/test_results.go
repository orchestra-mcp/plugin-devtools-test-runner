package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestResultsSchema returns the JSON Schema for the test_results tool.
func TestResultsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Project directory to scan for test result files",
			},
			"format": map[string]any{
				"type":        "string",
				"description": "Result file format to look for: junit, coverage, all (default: all)",
				"enum":        []any{"junit", "coverage", "all"},
			},
		},
		"required": []any{"directory"},
	})
	return s
}

// TestResults returns a handler that scans a project directory for test result
// artifacts (JUnit XML, Go coverage profiles, lcov files) and summarises them.
func TestResults() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		dir := helpers.GetString(req.Arguments, "directory")
		format := helpers.GetStringOr(req.Arguments, "format", "all")

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return helpers.ErrorResult("not_found", fmt.Sprintf("directory not found: %s", dir)), nil
		}

		// Patterns by format.
		var patterns []string
		switch format {
		case "junit":
			patterns = []string{"*.xml", "junit*.xml", "test-results*.xml"}
		case "coverage":
			patterns = []string{"coverage.out", "coverage.xml", "lcov.info", "*.lcov"}
		default: // "all"
			patterns = []string{
				"*.xml", "junit*.xml",
				"coverage.out", "coverage.xml", "lcov.info", "*.lcov",
				"test-results.json", "playwright-report/*.json",
			}
		}

		type resultFile struct {
			Path     string
			Size     int64
			Modified time.Time
		}
		var found []resultFile

		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				if info != nil && info.IsDir() {
					name := info.Name()
					if name == "node_modules" || name == "vendor" || name == "target" ||
						(len(name) > 0 && name[0] == '.') {
						return filepath.SkipDir
					}
				}
				return nil
			}
			for _, pat := range patterns {
				matched, matchErr := filepath.Match(pat, info.Name())
				if matchErr == nil && matched {
					found = append(found, resultFile{
						Path:     path,
						Size:     info.Size(),
						Modified: info.ModTime(),
					})
					break
				}
			}
			return nil
		})

		// Sort by modification time descending.
		sort.Slice(found, func(i, j int) bool {
			return found[i].Modified.After(found[j].Modified)
		})

		var sb strings.Builder
		fmt.Fprintf(&sb, "## Test Results: %s\n\n", dir)
		fmt.Fprintf(&sb, "- **Format filter**: %s\n", format)
		fmt.Fprintf(&sb, "- **Result files found**: %d\n\n", len(found))

		if len(found) == 0 {
			sb.WriteString("No test result files found. Run tests first to generate results.\n")
		} else {
			sb.WriteString("### Result files (newest first)\n")
			for _, f := range found {
				rel, err := filepath.Rel(dir, f.Path)
				if err != nil {
					rel = f.Path
				}
				fmt.Fprintf(&sb, "- **%s** (%d bytes, %s)\n",
					rel, f.Size, f.Modified.Format("2006-01-02 15:04:05"))
			}
		}

		return helpers.TextResult(sb.String()), nil
	}
}
