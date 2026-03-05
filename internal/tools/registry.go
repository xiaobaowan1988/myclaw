package tools

import (
	"fmt"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

// ExecuteFunc is the function signature for tool execution.
type ExecuteFunc func(args map[string]any) (string, error)

// Tool represents a tool that the agent can use.
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]llm.ParamDef
	Execute     ExecuteFunc
}

// Registry holds all available tools.
type Registry struct {
	tools map[string]*Tool
}

// NewRegistry creates a new tool registry with all built-in tools.
func NewRegistry() *Registry {
	r := &Registry{tools: make(map[string]*Tool)}
	r.Register(ReadFileTool())
	r.Register(WriteFileTool())
	r.Register(EditFileTool())
	r.Register(SearchTool())
	r.Register(GlobTool())
	r.Register(ListDirTool())
	r.Register(RunCmdTool())
	return r
}

// Register adds a tool to the registry.
func (r *Registry) Register(t *Tool) {
	r.tools[t.Name] = t
}

// Execute runs a tool by name with the given arguments.
func (r *Registry) Execute(name string, args map[string]any) (string, error) {
	tool, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return tool.Execute(args)
}

// ToolDefs returns all tools as LLM tool definitions.
func (r *Registry) ToolDefs() []llm.ToolDef {
	var defs []llm.ToolDef
	for _, t := range r.tools {
		defs = append(defs, llm.ToolDef{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Parameters,
		})
	}
	return defs
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (*Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}
