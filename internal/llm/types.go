package llm

// ToolDef defines a tool the LLM can call.
type ToolDef struct {
	Name        string
	Description string
	Parameters  map[string]ParamDef
}

// ParamDef defines a single parameter for a tool.
type ParamDef struct {
	Type        string
	Description string
	Required    bool
}

// Message represents a conversation message.
type Message struct {
	Role       string      // "user", "model", "tool"
	Text       string      // text content
	ToolCalls  []ToolCall  // tool calls from model
	ToolResult *ToolResult // tool result from execution
}

// ToolCall represents a function call request from the model.
type ToolCall struct {
	ID   string
	Name string
	Args map[string]any
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	CallID string
	Name   string
	Output string
}
