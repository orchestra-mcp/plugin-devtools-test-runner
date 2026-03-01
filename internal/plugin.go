package internal

import (
	"github.com/orchestra-mcp/sdk-go/plugin"
	"github.com/orchestra-mcp/plugin-devtools-test-runner/internal/tools"
)

// ToolsPlugin registers all test runner tools.
type ToolsPlugin struct{}

// RegisterTools registers all 8 test runner tools with the plugin builder.
func (tp *ToolsPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	builder.RegisterTool("test_run",
		"Run tests in a project directory, auto-detecting or using the specified framework",
		tools.TestRunSchema(), tools.TestRun())

	builder.RegisterTool("test_run_suite",
		"Run a named test suite or pattern within a project directory",
		tools.TestRunSuiteSchema(), tools.TestRunSuite())

	builder.RegisterTool("test_coverage",
		"Run tests with coverage reporting for the specified project directory",
		tools.TestCoverageSchema(), tools.TestCoverage())

	builder.RegisterTool("test_list",
		"List all available tests in a project directory",
		tools.TestListSchema(), tools.TestList())

	builder.RegisterTool("test_discover",
		"Discover test files in a project directory without running them",
		tools.TestDiscoverSchema(), tools.TestDiscover())

	builder.RegisterTool("test_results",
		"Scan a project directory for test result artifacts (JUnit XML, coverage files)",
		tools.TestResultsSchema(), tools.TestResults())

	builder.RegisterTool("test_run_file",
		"Run tests in a specific file within a project directory",
		tools.TestRunFileSchema(), tools.TestRunFile())

	builder.RegisterTool("test_status",
		"Check test infrastructure: detect framework and count test files",
		tools.TestStatusSchema(), tools.TestStatus())
}
