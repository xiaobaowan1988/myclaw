package context

import (
	"testing"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

func TestTrimHistory(t *testing.T) {
	m := NewManager(5)

	// Under limit - should not trim
	msgs := []llm.Message{
		{Role: "user", Text: "hello"},
		{Role: "model", Text: "hi"},
	}
	result := m.TrimHistory(msgs)
	if len(result) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result))
	}

	// Over limit - should trim
	msgs = []llm.Message{
		{Role: "user", Text: "first"},      // 0 - keep (first)
		{Role: "model", Text: "response1"},  // 1
		{Role: "user", Text: "second"},      // 2
		{Role: "model", Text: "response2"},  // 3
		{Role: "user", Text: "third"},       // 4
		{Role: "model", Text: "response3"},  // 5
		{Role: "user", Text: "fourth"},      // 6
		{Role: "model", Text: "response4"},  // 7
	}
	result = m.TrimHistory(msgs)
	if len(result) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(result))
	}
	// First should be preserved
	if result[0].Text != "first" {
		t.Errorf("expected first message preserved, got '%s'", result[0].Text)
	}
	// Last should be the most recent
	if result[len(result)-1].Text != "response4" {
		t.Errorf("expected last message to be 'response4', got '%s'", result[len(result)-1].Text)
	}
}

func TestRepoRoot(t *testing.T) {
	m := NewManager(50)
	// We're in a git repo, so this should be non-empty
	if m.RepoRoot() == "" {
		t.Skip("not in a git repository")
	}
}

func TestRepoSummary(t *testing.T) {
	m := NewManager(50)
	if m.RepoRoot() == "" {
		t.Skip("not in a git repository")
	}
	summary := m.RepoSummary()
	if summary == "" {
		t.Error("expected non-empty repo summary")
	}
}
