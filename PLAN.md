# OpenClaw — AI Coding Agent (Go + Gemini Flash)

## Overview

OpenClaw is an open-source, terminal-based AI coding agent built in Go that uses
Google's Gemini Flash model to assist developers with code generation, review,
refactoring, debugging, and explanation. It operates as a CLI tool that
understands your codebase context and can read/write files autonomously.

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   CLI (Cobra)                    │
│         User input / output / rendering          │
├─────────────────────────────────────────────────┤
│                 Agent Loop                       │
│   prompt → LLM → tool calls → execute → repeat  │
├────────────┬────────────┬───────────────────────┤
│  Tool:     │  Tool:     │  Tool:                │
│  ReadFile  │  WriteFile │  Search (grep/glob)   │
├────────────┼────────────┼───────────────────────┤
│  Tool:     │  Tool:     │  Tool:                │
│  RunCmd    │  ListDir   │  EditFile             │
├────────────┴────────────┴───────────────────────┤
│             Gemini Flash API Client              │
│     (function calling / streaming / context)     │
├─────────────────────────────────────────────────┤
│             Context Manager                      │
│   (file indexing, token budgeting, history)       │
└─────────────────────────────────────────────────┘
```

---

## Implementation Plan

### Phase 1: Project Scaffold & Gemini Client

**Step 1 — Initialize Go project**
- `go mod init github.com/xiaobaowan1988/openclaw`
- Set up directory structure:
  ```
  openclaw/
  ├── main.go              # Entrypoint
  ├── go.mod
  ├── go.sum
  ├── cmd/
  │   └── root.go          # Cobra CLI root command
  ├── internal/
  │   ├── agent/
  │   │   ├── agent.go     # Agent loop (think → act → observe)
  │   │   └── agent_test.go
  │   ├── llm/
  │   │   ├── gemini.go    # Gemini Flash API client
  │   │   ├── types.go     # Request/response types
  │   │   └── gemini_test.go
  │   ├── tools/
  │   │   ├── registry.go  # Tool registration & dispatch
  │   │   ├── readfile.go  # Read file tool
  │   │   ├── writefile.go # Write/create file tool
  │   │   ├── editfile.go  # Edit file (search & replace) tool
  │   │   ├── search.go    # Grep/glob search tool
  │   │   ├── listdir.go   # List directory tool
  │   │   ├── runcmd.go    # Execute shell command tool
  │   │   └── tools_test.go
  │   ├── context/
  │   │   ├── manager.go   # Context window management
  │   │   └── manager_test.go
  │   └── ui/
  │       ├── terminal.go  # Terminal rendering (markdown, colors)
  │       └── spinner.go   # Loading indicators
  └── config/
      └── config.go        # Configuration (API key, model, etc.)
  ```

**Step 2 — Gemini Flash API client** (`internal/llm/`)
- Use `google.golang.org/genai` (Google's official Go SDK for Gemini)
- Support streaming responses for real-time output
- Implement function calling (tool_use) support
- Handle API key via `GEMINI_API_KEY` env var
- Model: `gemini-2.0-flash` (fast, cheap, good at code)

**Step 3 — Configuration** (`config/`)
- Load from env vars + optional `~/.openclaw.yaml`
- Settings: API key, model name, max tokens, temperature

### Phase 2: Tool System

**Step 4 — Tool registry** (`internal/tools/`)
- Define `Tool` interface:
  ```go
  type Tool struct {
      Name        string
      Description string
      Parameters  map[string]Parameter
      Execute     func(args map[string]any) (string, error)
  }
  ```
- Registry maps tool names → tool definitions
- Convert tools to Gemini function declarations automatically

**Step 5 — Implement core tools**
| Tool       | Purpose                                      |
|------------|----------------------------------------------|
| `read_file`  | Read file contents (with line numbers)     |
| `write_file` | Create or overwrite a file                 |
| `edit_file`  | Search-and-replace within a file           |
| `search`     | Grep for pattern across files              |
| `glob`       | Find files by pattern                      |
| `list_dir`   | List directory contents                    |
| `run_cmd`    | Execute a shell command (with timeout)     |

### Phase 3: Agent Loop

**Step 6 — Agent loop** (`internal/agent/`)
- Core loop:
  1. Build prompt with system message + conversation history + user input
  2. Call Gemini Flash with tools
  3. If response contains tool calls → execute them, append results, go to 2
  4. If response is text → display to user, wait for next input
- Implement token budgeting to stay within context limits
- Support multi-turn conversation with history

**Step 7 — System prompt**
- Define agent persona and capabilities
- Include current working directory, OS info
- Instruct the model on tool usage patterns

### Phase 4: CLI & User Experience

**Step 8 — CLI with Cobra** (`cmd/`)
- `openclaw` — start interactive chat session
- `openclaw "fix the bug in main.go"` — one-shot mode
- Flags: `--model`, `--verbose`, `--no-stream`

**Step 9 — Terminal UI** (`internal/ui/`)
- Render markdown in terminal (bold, code blocks, lists)
- Streaming token output (print as tokens arrive)
- Colored tool call indicators (e.g., `[reading file.go...]`)
- Spinner for waiting states

### Phase 5: Context & Polish

**Step 10 — Context management** (`internal/context/`)
- Auto-detect git repo root
- Track which files have been read in the session
- Truncate old messages when approaching token limit
- Optionally include repo structure summary in initial context

**Step 11 — Safety & guardrails**
- Confirm before running destructive commands (`rm`, `git push --force`, etc.)
- Sandbox shell commands with timeout
- Limit file write to project directory

**Step 12 — Testing & documentation**
- Unit tests for tools, agent loop, API client
- Integration test with mock Gemini responses
- README with setup instructions and usage examples

---

## Key Dependencies

| Package                          | Purpose                    |
|----------------------------------|----------------------------|
| `github.com/spf13/cobra`        | CLI framework              |
| `google.golang.org/genai`       | Gemini API (official SDK)  |
| `github.com/charmbracelet/glamour` | Terminal markdown rendering |
| `github.com/charmbracelet/lipgloss` | Terminal styling          |
| `gopkg.in/yaml.v3`              | Config file parsing        |

---

## Execution Order

```
Phase 1 (Steps 1-3)  →  Scaffold + Gemini client working
Phase 2 (Steps 4-5)  →  Tools execute and return results
Phase 3 (Steps 6-7)  →  Agent loop: LLM + tools in a loop ← MVP here
Phase 4 (Steps 8-9)  →  Polished CLI experience
Phase 5 (Steps 10-12) → Context mgmt, safety, tests
```

**MVP milestone**: After Phase 3, you have a working coding agent that can
read/write files and answer questions about your codebase.
