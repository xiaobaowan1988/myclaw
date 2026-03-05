package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	// Should have 7 tools
	defs := r.ToolDefs()
	if len(defs) != 7 {
		t.Fatalf("expected 7 tool defs, got %d", len(defs))
	}

	// Should find known tools
	for _, name := range []string{"read_file", "write_file", "edit_file", "search", "glob", "list_dir", "run_cmd"} {
		if _, ok := r.Get(name); !ok {
			t.Errorf("tool %s not found in registry", name)
		}
	}

	// Unknown tool should error
	_, err := r.Execute("nonexistent", nil)
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestReadFileTool(t *testing.T) {
	// Create temp file
	tmp := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(tmp, []byte("line1\nline2\nline3"), 0644)

	r := NewRegistry()
	result, err := r.Execute("read_file", map[string]any{"path": tmp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "line1") || !strings.Contains(result, "line2") {
		t.Errorf("expected file content, got: %s", result)
	}
	// Should have line numbers
	if !strings.Contains(result, "1 |") {
		t.Errorf("expected line numbers, got: %s", result)
	}
}

func TestWriteFileTool(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "subdir", "test.txt")

	r := NewRegistry()
	result, err := r.Execute("write_file", map[string]any{"path": tmp, "content": "hello world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Successfully wrote") {
		t.Errorf("unexpected result: %s", result)
	}

	// Verify file was written
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(data))
	}
}

func TestEditFileTool(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "edit.txt")
	os.WriteFile(tmp, []byte("func old() {}"), 0644)

	r := NewRegistry()
	result, err := r.Execute("edit_file", map[string]any{
		"path":       tmp,
		"old_string": "func old() {}",
		"new_string": "func newFunc() {}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Successfully edited") {
		t.Errorf("unexpected result: %s", result)
	}

	data, _ := os.ReadFile(tmp)
	if string(data) != "func newFunc() {}" {
		t.Errorf("expected edited content, got '%s'", string(data))
	}

	// Test not found
	_, err = r.Execute("edit_file", map[string]any{
		"path":       tmp,
		"old_string": "nonexistent",
		"new_string": "replacement",
	})
	if err == nil {
		t.Error("expected error for not found string")
	}
}

func TestSearchTool(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc hello() {}"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("nothing here"), 0644)

	r := NewRegistry()
	result, err := r.Execute("search", map[string]any{"pattern": "func.*hello", "path": dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "hello") {
		t.Errorf("expected match, got: %s", result)
	}

	// With glob filter
	result, err = r.Execute("search", map[string]any{"pattern": "func", "path": dir, "glob": "*.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "No matches") {
		t.Errorf("expected no matches with glob filter, got: %s", result)
	}
}

func TestGlobTool(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte(""), 0644)

	r := NewRegistry()
	result, err := r.Execute("glob", map[string]any{"pattern": "*.go", "path": dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "a.go") || !strings.Contains(result, "b.go") {
		t.Errorf("expected go files, got: %s", result)
	}
	if strings.Contains(result, "c.txt") {
		t.Errorf("should not contain txt file, got: %s", result)
	}
}

func TestListDirTool(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	r := NewRegistry()
	result, err := r.Execute("list_dir", map[string]any{"path": dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "file.txt") {
		t.Errorf("expected file.txt, got: %s", result)
	}
	if !strings.Contains(result, "subdir/") {
		t.Errorf("expected subdir/, got: %s", result)
	}
}

func TestRunCmdTool(t *testing.T) {
	r := NewRegistry()
	result, err := r.Execute("run_cmd", map[string]any{"command": "echo hello world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "hello world") {
		t.Errorf("expected 'hello world', got: %s", result)
	}
}

func TestRunCmdSafety(t *testing.T) {
	r := NewRegistry()

	destructive := []string{
		"rm -rf /",
		"rm -rf ~",
		"git push --force origin main",
		"git reset --hard",
	}
	for _, cmd := range destructive {
		_, err := r.Execute("run_cmd", map[string]any{"command": cmd})
		if err == nil {
			t.Errorf("expected error for destructive command: %s", cmd)
		}
		if !strings.Contains(err.Error(), "blocked") {
			t.Errorf("expected 'blocked' in error for '%s', got: %v", cmd, err)
		}
	}

	// Safe commands should work
	_, err := r.Execute("run_cmd", map[string]any{"command": "echo safe"})
	if err != nil {
		t.Errorf("safe command should not be blocked: %v", err)
	}
}
