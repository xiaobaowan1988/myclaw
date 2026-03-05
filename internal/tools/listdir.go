package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

// ListDirTool creates the list_dir tool.
func ListDirTool() *Tool {
	return &Tool{
		Name:        "list_dir",
		Description: "List the contents of a directory. Shows files and subdirectories with sizes.",
		Parameters: map[string]llm.ParamDef{
			"path": {Type: "string", Description: "Path to the directory (default: current directory)", Required: false},
		},
		Execute: func(args map[string]any) (string, error) {
			path, _ := args["path"].(string)
			if path == "" {
				path = "."
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				return "", fmt.Errorf("failed to read directory: %w", err)
			}

			var sb strings.Builder
			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				if entry.IsDir() {
					fmt.Fprintf(&sb, "  [dir]  %s/\n", entry.Name())
				} else {
					fmt.Fprintf(&sb, "  %6d  %s\n", info.Size(), entry.Name())
				}
			}

			if sb.Len() == 0 {
				return "Directory is empty.", nil
			}
			return sb.String(), nil
		},
	}
}
