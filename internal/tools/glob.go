package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

// GlobTool creates the glob tool for finding files by pattern.
func GlobTool() *Tool {
	return &Tool{
		Name:        "glob",
		Description: "Find files matching a glob pattern. Returns matching file paths.",
		Parameters: map[string]llm.ParamDef{
			"pattern": {Type: "string", Description: "Glob pattern (e.g., '**/*.go', 'src/*.ts')", Required: true},
			"path":    {Type: "string", Description: "Base directory to search from (default: current directory)", Required: false},
		},
		Execute: func(args map[string]any) (string, error) {
			pattern, _ := args["pattern"].(string)
			if pattern == "" {
				return "", fmt.Errorf("pattern is required")
			}

			dir, _ := args["path"].(string)
			if dir == "" {
				dir = "."
			}

			var matches []string
			maxResults := 200

			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if len(matches) >= maxResults {
					return filepath.SkipAll
				}

				// Skip hidden dirs
				if info.IsDir() {
					base := filepath.Base(path)
					if strings.HasPrefix(base, ".") && path != dir {
						return filepath.SkipDir
					}
					return nil
				}

				matched, _ := filepath.Match(pattern, filepath.Base(path))
				if matched {
					matches = append(matches, path)
				}
				return nil
			})
			if err != nil {
				return "", err
			}

			if len(matches) == 0 {
				return "No files found matching pattern.", nil
			}

			return strings.Join(matches, "\n"), nil
		},
	}
}
