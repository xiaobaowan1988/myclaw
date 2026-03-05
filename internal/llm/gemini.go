package llm

import (
	"context"
	"fmt"
	"google.golang.org/genai"
)

// Client wraps the Gemini API for the agent.
type Client struct {
	client *genai.Client
	model  string
	config ClientConfig
}

// ClientConfig holds configuration for the LLM client.
type ClientConfig struct {
	MaxTokens   int32
	Temperature float32
}

// NewClient creates a new Gemini client.
func NewClient(apiKey, model string, cfg ClientConfig) (*Client, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client: client,
		model:  model,
		config: cfg,
	}, nil
}

// GenerateStream sends messages to Gemini and streams the response.
// It calls onText for text chunks and returns any tool calls.
func (c *Client) GenerateStream(
	ctx context.Context,
	systemPrompt string,
	messages []Message,
	tools []ToolDef,
	onText func(chunk string),
) ([]ToolCall, error) {
	geminiConfig := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, "user"),
		Temperature:       genai.Ptr(c.config.Temperature),
		MaxOutputTokens:   c.config.MaxTokens,
	}

	// Convert tools to Gemini format
	if len(tools) > 0 {
		geminiConfig.Tools = convertTools(tools)
	}

	// Build contents from messages
	contents := convertMessages(messages)

	var allToolCalls []ToolCall

	for resp, err := range c.client.Models.GenerateContentStream(ctx, c.model, contents, geminiConfig) {
		if err != nil {
			return nil, fmt.Errorf("stream error: %w", err)
		}
		for _, candidate := range resp.Candidates {
			if candidate.Content == nil {
				continue
			}
			for _, part := range candidate.Content.Parts {
				if part.Text != "" && onText != nil {
					onText(part.Text)
				}
				if part.FunctionCall != nil {
					allToolCalls = append(allToolCalls, ToolCall{
						ID:   part.FunctionCall.Name,
						Name: part.FunctionCall.Name,
						Args: part.FunctionCall.Args,
					})
				}
			}
		}
	}

	return allToolCalls, nil
}

// Generate sends messages and returns the full response (non-streaming).
func (c *Client) Generate(
	ctx context.Context,
	systemPrompt string,
	messages []Message,
	tools []ToolDef,
) (string, []ToolCall, error) {
	var text string
	toolCalls, err := c.GenerateStream(ctx, systemPrompt, messages, tools, func(chunk string) {
		text += chunk
	})
	return text, toolCalls, err
}

func convertTools(tools []ToolDef) []*genai.Tool {
	var declarations []*genai.FunctionDeclaration
	for _, t := range tools {
		props := make(map[string]*genai.Schema)
		var required []string
		for name, p := range t.Parameters {
			props[name] = &genai.Schema{
				Type:        genai.Type(p.Type),
				Description: p.Description,
			}
			if p.Required {
				required = append(required, name)
			}
		}
		declarations = append(declarations, &genai.FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,
			Parameters: &genai.Schema{
				Type:       genai.TypeObject,
				Properties: props,
				Required:   required,
			},
		})
	}
	return []*genai.Tool{{FunctionDeclarations: declarations}}
}

func convertMessages(messages []Message) []*genai.Content {
	var contents []*genai.Content
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			contents = append(contents, genai.NewContentFromText(msg.Text, "user"))
		case "model":
			if msg.Text != "" {
				contents = append(contents, genai.NewContentFromText(msg.Text, "model"))
			}
			if len(msg.ToolCalls) > 0 {
				var parts []*genai.Part
				for _, tc := range msg.ToolCalls {
					parts = append(parts, &genai.Part{
						FunctionCall: &genai.FunctionCall{
							Name: tc.Name,
							Args: tc.Args,
						},
					})
				}
				contents = append(contents, &genai.Content{
					Role:  "model",
					Parts: parts,
				})
			}
		case "tool":
			if msg.ToolResult != nil {
				contents = append(contents, &genai.Content{
					Role: "user",
					Parts: []*genai.Part{
						{
							FunctionResponse: &genai.FunctionResponse{
								Name: msg.ToolResult.Name,
								Response: map[string]any{
									"output": msg.ToolResult.Output,
								},
							},
						},
					},
				})
			}
		}
	}
	return contents
}
