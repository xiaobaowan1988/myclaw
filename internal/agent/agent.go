package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	ctxmgr "github.com/xiaobaowan1988/openclaw/internal/context"
	"github.com/xiaobaowan1988/openclaw/internal/llm"
	"github.com/xiaobaowan1988/openclaw/internal/tools"
)

const maxToolLoops = 25

// Agent is the core agent that orchestrates LLM and tools.
type Agent struct {
	client   *llm.Client
	registry *tools.Registry
	ctxMgr   *ctxmgr.Manager
	history  []llm.Message
	system   string
	onText   func(string)
	onTool   func(name string, args map[string]any)
}

// Config holds agent configuration.
type Config struct {
	Client   *llm.Client
	Registry *tools.Registry
	OnText   func(string)         // called when model streams text
	OnTool   func(string, map[string]any) // called when a tool is invoked
}

// New creates a new agent.
func New(cfg Config) *Agent {
	cm := ctxmgr.NewManager(50) // keep up to 50 messages
	return &Agent{
		client:   cfg.Client,
		registry: cfg.Registry,
		ctxMgr:   cm,
		history:  nil,
		system:   buildSystemPrompt(cm),
		onText:   cfg.OnText,
		onTool:   cfg.OnTool,
	}
}

// Run processes a user message and runs the agent loop until the model
// produces a text response (no more tool calls).
func (a *Agent) Run(ctx context.Context, userMessage string) (string, error) {
	a.history = append(a.history, llm.Message{
		Role: "user",
		Text: userMessage,
	})

	// Trim history if too long
	a.history = a.ctxMgr.TrimHistory(a.history)

	toolDefs := a.registry.ToolDefs()

	for i := 0; i < maxToolLoops; i++ {
		var textResponse strings.Builder

		toolCalls, err := a.client.GenerateStream(
			ctx,
			a.system,
			a.history,
			toolDefs,
			func(chunk string) {
				textResponse.WriteString(chunk)
				if a.onText != nil {
					a.onText(chunk)
				}
			},
		)
		if err != nil {
			return "", fmt.Errorf("LLM error: %w", err)
		}

		// If we got text and no tool calls, we're done
		if len(toolCalls) == 0 {
			text := textResponse.String()
			a.history = append(a.history, llm.Message{
				Role: "model",
				Text: text,
			})
			return text, nil
		}

		// Record the model's tool call message
		a.history = append(a.history, llm.Message{
			Role:      "model",
			Text:      textResponse.String(),
			ToolCalls: toolCalls,
		})

		// Execute each tool call and record results
		for _, tc := range toolCalls {
			if a.onTool != nil {
				a.onTool(tc.Name, tc.Args)
			}

			output, err := a.registry.Execute(tc.Name, tc.Args)
			if err != nil {
				output = fmt.Sprintf("Error: %v", err)
			}

			a.history = append(a.history, llm.Message{
				Role: "tool",
				ToolResult: &llm.ToolResult{
					CallID: tc.ID,
					Name:   tc.Name,
					Output: output,
				},
			})
		}
	}

	return "", fmt.Errorf("agent exceeded maximum tool loop iterations (%d)", maxToolLoops)
}

// Reset clears the conversation history.
func (a *Agent) Reset() {
	a.history = nil
}

// History returns the current conversation history.
func (a *Agent) History() []llm.Message {
	return a.history
}

func buildSystemPrompt(cm *ctxmgr.Manager) string {
	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(cwd)

	prompt := fmt.Sprintf(`You are OpenClaw, an AI coding agent. You help developers by reading, writing, and modifying code.

## Environment
- OS: %s/%s
- Working directory: %s

## Instructions
- Use the provided tools to interact with the filesystem and run commands.
- Read files before modifying them to understand existing code.
- Use edit_file for targeted changes, write_file only for new files or full rewrites.
- When searching for code, use the search or glob tools.
- Be concise and direct in your responses.
- When you're done with a task, summarize what you did.
`, runtime.GOOS, runtime.GOARCH, absPath)

	if summary := cm.RepoSummary(); summary != "" {
		prompt += "\n## Repository\n" + summary
	}

	return prompt
}
