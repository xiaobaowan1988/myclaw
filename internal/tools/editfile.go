package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

// EditFileTool creates the edit_file tool.
func EditFileTool() *Tool {
	return &Tool{
		Name:        "edit_file",
		Description: "Edit a file by replacing an exact string match with new content. The old_string must match exactly (including whitespace).",
		Parameters: map[string]llm.ParamDef{
			"path":       {Type: "string", Description: "Path to the file to edit", Required: true},
			"old_string": {Type: "string", Description: "The exact string to find and replace", Required: true},
			"new_string": {Type: "string", Description: "The replacement string", Required: true},
		},
		Execute: func(args map[string]any) (string, error) {
			path, _ := args["path"].(string)
			oldStr, _ := args["old_string"].(string)
			newStr, _ := args["new_string"].(string)

			if path == "" {
				return "", fmt.Errorf("path is required")
			}
			if oldStr == "" {
				return "", fmt.Errorf("old_string is required")
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("failed to read file: %w", err)
			}

			content := string(data)
			count := strings.Count(content, oldStr)
			if count == 0 {
				return "", fmt.Errorf("old_string not found in file")
			}
			if count > 1 {
				return "", fmt.Errorf("old_string found %d times, must be unique (found %d matches)", count, count)
			}

			newContent := strings.Replace(content, oldStr, newStr, 1)
			if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
				return "", fmt.Errorf("failed to write file: %w", err)
			}

			return fmt.Sprintf("Successfully edited %s (replaced 1 occurrence)", path), nil
		},
	}
}
