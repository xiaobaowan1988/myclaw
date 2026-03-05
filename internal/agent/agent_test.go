package agent

import (
	"runtime"
	"strings"
	"testing"

	ctxmgr "github.com/xiaobaowan1988/openclaw/internal/context"
	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

func TestBuildSystemPrompt(t *testing.T) {
	cm := ctxmgr.NewManager(50)
	prompt := buildSystemPrompt(cm)

	if !strings.Contains(prompt, "OpenClaw") {
		t.Error("system prompt should contain 'OpenClaw'")
	}
	if !strings.Contains(prompt, runtime.GOOS) {
		t.Error("system prompt should contain OS info")
	}
}

func TestAgentNew(t *testing.T) {
	a := New(Config{
		Client:   nil,
		Registry: nil,
	})
	if a == nil {
		t.Fatal("expected non-nil agent")
	}
	if len(a.History()) != 0 {
		t.Error("expected empty history")
	}
}

func TestAgentReset(t *testing.T) {
	a := New(Config{})
	a.history = append(a.history, llm.Message{Role: "user", Text: "hello"})
	if len(a.History()) != 1 {
		t.Fatal("expected 1 message")
	}
	a.Reset()
	if len(a.History()) != 0 {
		t.Error("expected empty history after reset")
	}
}
