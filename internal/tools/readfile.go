package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

// ReadFileTool creates the read_file tool.
func ReadFileTool() *Tool {
	return &Tool{
		Name:        "read_file",
		Description: "Read the contents of a file. Returns the file content with line numbers.",
		Parameters: map[string]llm.ParamDef{
			"path": {Type: "string", Description: "Absolute or relative path to the file", Required: true},
		},
		Execute: func(args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok || path == "" {
				return "", fmt.Errorf("path is required")
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("failed to read file: %w", err)
			}

			lines := strings.Split(string(data), "\n")
			var sb strings.Builder
			for i, line := range lines {
				fmt.Fprintf(&sb, "%4d | %s\n", i+1, line)
			}
			return sb.String(), nil
		},
	}
}
