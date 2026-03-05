package llm

import (
	"testing"
)

func TestConvertTools(t *testing.T) {
	tools := []ToolDef{
		{
			Name:        "read_file",
			Description: "Read a file",
			Parameters: map[string]ParamDef{
				"path": {Type: "string", Description: "File path", Required: true},
			},
		},
	}

	result := convertTools(tools)
	if len(result) != 1 {
		t.Fatalf("expected 1 tool group, got %d", len(result))
	}
	if len(result[0].FunctionDeclarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(result[0].FunctionDeclarations))
	}
	decl := result[0].FunctionDeclarations[0]
	if decl.Name != "read_file" {
		t.Errorf("expected name 'read_file', got '%s'", decl.Name)
	}
	if decl.Parameters == nil {
		t.Fatal("expected parameters, got nil")
	}
	if _, ok := decl.Parameters.Properties["path"]; !ok {
		t.Error("expected 'path' property")
	}
}

func TestConvertMessages(t *testing.T) {
	messages := []Message{
		{Role: "user", Text: "Hello"},
		{Role: "model", Text: "Hi there"},
		{Role: "user", Text: "Read a file"},
		{Role: "model", ToolCalls: []ToolCall{
			{Name: "read_file", Args: map[string]any{"path": "main.go"}},
		}},
		{Role: "tool", ToolResult: &ToolResult{Name: "read_file", Output: "package main"}},
	}

	contents := convertMessages(messages)
	if len(contents) != 5 {
		t.Fatalf("expected 5 contents, got %d", len(contents))
	}
	if contents[0].Role != "user" {
		t.Errorf("expected role 'user', got '%s'", contents[0].Role)
	}
	if contents[1].Role != "model" {
		t.Errorf("expected role 'model', got '%s'", contents[1].Role)
	}
}
