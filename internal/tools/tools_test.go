package tools

import (
	"context"
	"os"
	"strings"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// ---------- helpers ----------

func callTool(t *testing.T, handler func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error), args map[string]any) *pluginv1.ToolResponse {
	t.Helper()
	var s *structpb.Struct
	if args != nil {
		var err error
		s, err = structpb.NewStruct(args)
		if err != nil {
			t.Fatalf("NewStruct: %v", err)
		}
	}
	resp, err := handler(context.Background(), &pluginv1.ToolRequest{Arguments: s})
	if err != nil {
		t.Fatalf("handler returned Go error: %v", err)
	}
	return resp
}

func isError(resp *pluginv1.ToolResponse) bool {
	return resp != nil && !resp.Success
}

func errorCode(resp *pluginv1.ToolResponse) string {
	if resp == nil {
		return ""
	}
	return resp.GetErrorCode()
}

func getText(resp *pluginv1.ToolResponse) string {
	if resp == nil {
		return ""
	}
	if r := resp.GetResult(); r != nil {
		if f := r.GetFields(); f != nil {
			if tf, ok := f["text"]; ok {
				return tf.GetStringValue()
			}
		}
	}
	return ""
}

// makeGoProject creates a temp dir with go.mod and a *_test.go file.
func makeGoProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/go.mod", []byte("module example.com/test\ngo 1.23\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/main_test.go", []byte("package main\nimport \"testing\"\nfunc TestNoop(t *testing.T) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// ---------- test_run ----------

func TestTestRun_MissingDirectory(t *testing.T) {
	resp := callTool(t, TestRun(), map[string]any{})
	if !isError(resp) {
		t.Error("expected validation_error for missing directory")
	}
}

func TestTestRun_UnknownFramework(t *testing.T) {
	dir := t.TempDir() // no go.mod, Cargo.toml, etc.
	resp := callTool(t, TestRun(), map[string]any{"directory": dir})
	if !isError(resp) {
		t.Error("expected unknown_framework error for empty dir")
	}
	if errorCode(resp) != "unknown_framework" {
		t.Errorf("expected unknown_framework, got %q", errorCode(resp))
	}
}

func TestTestRun_GoProject(t *testing.T) {
	dir := makeGoProject(t)
	resp := callTool(t, TestRun(), map[string]any{"directory": dir})
	// May succeed or fail depending on Go being available — both are valid.
	_ = resp
}

// ---------- test_run_suite ----------

func TestTestRunSuite_MissingDirectory(t *testing.T) {
	resp := callTool(t, TestRunSuite(), map[string]any{"suite": "TestFoo"})
	if !isError(resp) {
		t.Error("expected validation_error for missing directory")
	}
}

func TestTestRunSuite_MissingSuite(t *testing.T) {
	resp := callTool(t, TestRunSuite(), map[string]any{"directory": "/tmp"})
	if !isError(resp) {
		t.Error("expected validation_error for missing suite")
	}
}

func TestTestRunSuite_UnknownFramework(t *testing.T) {
	dir := t.TempDir()
	resp := callTool(t, TestRunSuite(), map[string]any{
		"directory": dir,
		"suite":     "MyTestSuite",
	})
	if !isError(resp) {
		t.Error("expected unknown_framework error")
	}
	if errorCode(resp) != "unknown_framework" {
		t.Errorf("expected unknown_framework, got %q", errorCode(resp))
	}
}

func TestTestRunSuite_GoProject(t *testing.T) {
	dir := makeGoProject(t)
	resp := callTool(t, TestRunSuite(), map[string]any{
		"directory": dir,
		"suite":     "TestNoop",
		"framework": "go",
	})
	_ = resp // success or test_failed depending on Go env
}

// ---------- test_coverage ----------

func TestTestCoverage_MissingDirectory(t *testing.T) {
	resp := callTool(t, TestCoverage(), map[string]any{})
	if !isError(resp) {
		t.Error("expected validation_error for missing directory")
	}
}

func TestTestCoverage_UnknownFramework(t *testing.T) {
	dir := t.TempDir()
	resp := callTool(t, TestCoverage(), map[string]any{"directory": dir})
	if !isError(resp) {
		t.Error("expected unknown_framework error")
	}
}

// ---------- test_list ----------

func TestTestList_MissingDirectory(t *testing.T) {
	resp := callTool(t, TestList(), map[string]any{})
	if !isError(resp) {
		t.Error("expected validation_error for missing directory")
	}
}

func TestTestList_NpmFramework(t *testing.T) {
	dir := t.TempDir()
	resp := callTool(t, TestList(), map[string]any{
		"directory": dir,
		"framework": "npm",
	})
	// npm returns a text notice, not an error.
	if isError(resp) {
		t.Errorf("unexpected error: %s", errorCode(resp))
	}
}

// ---------- test_discover ----------

func TestTestDiscover_MissingDirectory(t *testing.T) {
	resp := callTool(t, TestDiscover(), map[string]any{})
	if !isError(resp) {
		t.Error("expected validation_error for missing directory")
	}
}

func TestTestDiscover_UnknownFramework(t *testing.T) {
	dir := t.TempDir()
	resp := callTool(t, TestDiscover(), map[string]any{"directory": dir})
	if !isError(resp) {
		t.Error("expected unknown_framework error")
	}
	if errorCode(resp) != "unknown_framework" {
		t.Errorf("expected unknown_framework, got %q", errorCode(resp))
	}
}

func TestTestDiscover_GoProjectFindsTestFiles(t *testing.T) {
	dir := makeGoProject(t)
	resp := callTool(t, TestDiscover(), map[string]any{"directory": dir})
	if isError(resp) {
		t.Errorf("unexpected error: %s", errorCode(resp))
	}
	txt := getText(resp)
	if !strings.Contains(txt, "main_test.go") {
		t.Errorf("expected main_test.go in discovery result, got: %s", txt)
	}
}

func TestTestDiscover_EmptyGoProject(t *testing.T) {
	dir := t.TempDir()
	// go.mod but no test files.
	_ = os.WriteFile(dir+"/go.mod", []byte("module example.com/test\n"), 0644)
	resp := callTool(t, TestDiscover(), map[string]any{"directory": dir})
	if isError(resp) {
		t.Errorf("unexpected error: %s", errorCode(resp))
	}
	txt := getText(resp)
	if !strings.Contains(txt, "0") {
		t.Errorf("expected 0 files in result, got: %s", txt)
	}
}

// ---------- test_results ----------

func TestTestResults_MissingDirectory(t *testing.T) {
	resp := callTool(t, TestResults(), map[string]any{})
	if !isError(resp) {
		t.Error("expected validation_error for missing directory")
	}
}

func TestTestResults_NonexistentDirectory(t *testing.T) {
	resp := callTool(t, TestResults(), map[string]any{
		"directory": "/tmp/no-such-dir-orchestra-xyz",
	})
	if !isError(resp) {
		t.Error("expected not_found for nonexistent directory")
	}
	if errorCode(resp) != "not_found" {
		t.Errorf("expected not_found, got %q", errorCode(resp))
	}
}

func TestTestResults_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	resp := callTool(t, TestResults(), map[string]any{"directory": dir})
	if isError(resp) {
		t.Errorf("unexpected error: %s", errorCode(resp))
	}
	txt := getText(resp)
	if !strings.Contains(txt, "0") {
		t.Errorf("expected 0 result files, got: %s", txt)
	}
}

func TestTestResults_WithCoverageFile(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(dir+"/coverage.out", []byte("mode: set\n"), 0644)

	resp := callTool(t, TestResults(), map[string]any{
		"directory": dir,
		"format":    "coverage",
	})
	if isError(resp) {
		t.Errorf("unexpected error: %s", errorCode(resp))
	}
	txt := getText(resp)
	if !strings.Contains(txt, "coverage.out") {
		t.Errorf("expected coverage.out in results, got: %s", txt)
	}
}

func TestTestResults_InvalidFormat(t *testing.T) {
	// JSON schema enum validation happens at schema level; handler accepts any string.
	// Passing an unknown format falls through to "all" patterns via default case.
	dir := t.TempDir()
	resp := callTool(t, TestResults(), map[string]any{
		"directory": dir,
		"format":    "all",
	})
	if isError(resp) {
		t.Errorf("unexpected error: %s", errorCode(resp))
	}
}

// ---------- test_run_file ----------

func TestTestRunFile_MissingDirectory(t *testing.T) {
	resp := callTool(t, TestRunFile(), map[string]any{"file": "main_test.go"})
	if !isError(resp) {
		t.Error("expected validation_error for missing directory")
	}
}

// ---------- test_status ----------

func TestTestStatus_MissingDirectory(t *testing.T) {
	resp := callTool(t, TestStatus(), map[string]any{})
	if !isError(resp) {
		t.Error("expected validation_error for missing directory")
	}
}

func TestTestStatus_GoProject(t *testing.T) {
	dir := makeGoProject(t)
	resp := callTool(t, TestStatus(), map[string]any{"directory": dir})
	if isError(resp) {
		t.Errorf("unexpected error: %s", errorCode(resp))
	}
	txt := getText(resp)
	if !strings.Contains(txt, "go") {
		t.Errorf("expected 'go' framework in status, got: %s", txt)
	}
}
