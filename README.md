# OpenClaw

An open-source AI coding agent built with Go and Google's Gemini Flash model.

OpenClaw helps developers with code generation, review, refactoring, debugging, and explanation through a terminal-based interactive interface.

## Features

- **Agent Loop**: Autonomous think → act → observe cycle with tool calling
- **7 Built-in Tools**: read_file, write_file, edit_file, search, glob, list_dir, run_cmd
- **Streaming Output**: Real-time response streaming from Gemini Flash
- **Context Management**: Automatic conversation history trimming and repo-aware context
- **Safety Guardrails**: Blocks destructive commands (rm -rf /, force push to main, etc.)
- **Interactive & One-shot Modes**: Chat interactively or pass a single message

## Quick Start

```bash
# Set your Gemini API key
export GEMINI_API_KEY="your-api-key"

# Build
go build -o openclaw .

# Interactive mode
./openclaw

# One-shot mode
./openclaw "explain the main function in this project"
```

## Configuration

Set via environment variable or `~/.openclaw.yaml`:

```yaml
api_key: "your-gemini-api-key"
model: "gemini-2.0-flash"
max_tokens: 8192
temperature: 0.2
```

## Project Structure

```
openclaw/
├── main.go                     # Entrypoint
├── cmd/root.go                 # CLI (Cobra)
├── config/config.go            # Configuration
├── internal/
│   ├── agent/agent.go          # Agent loop
│   ├── llm/gemini.go           # Gemini Flash API client
│   ├── tools/                  # 7 built-in tools
│   ├── context/manager.go      # Context window management
│   └── ui/                     # Terminal rendering
```

## Running Tests

```bash
go test ./... -v
```

## License

MIT
