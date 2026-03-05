package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

// WriteFileTool creates the write_file tool.
func WriteFileTool() *Tool {
	return &Tool{
		Name:        "write_file",
		Description: "Create or overwrite a file with the given content.",
		Parameters: map[string]llm.ParamDef{
			"path":    {Type: "string", Description: "Path to the file to write", Required: true},
			"content": {Type: "string", Description: "Content to write to the file", Required: true},
		},
		Execute: func(args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok || path == "" {
				return "", fmt.Errorf("path is required")
			}
			content, ok := args["content"].(string)
			if !ok {
				return "", fmt.Errorf("content is required")
			}

			// Ensure parent directory exists
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}

			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return "", fmt.Errorf("failed to write file: %w", err)
			}

			return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
		},
	}
}
