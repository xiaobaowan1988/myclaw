package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

// SearchTool creates the search tool (grep-like).
func SearchTool() *Tool {
	return &Tool{
		Name:        "search",
		Description: "Search for a regex pattern across files in a directory. Returns matching lines with file paths and line numbers.",
		Parameters: map[string]llm.ParamDef{
			"pattern": {Type: "string", Description: "Regex pattern to search for", Required: true},
			"path":    {Type: "string", Description: "Directory to search in (default: current directory)", Required: false},
			"glob":    {Type: "string", Description: "File glob pattern to filter files (e.g., '*.go')", Required: false},
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

			globPattern, _ := args["glob"].(string)

			re, err := regexp.Compile(pattern)
			if err != nil {
				return "", fmt.Errorf("invalid regex: %w", err)
			}

			var results strings.Builder
			matchCount := 0
			maxMatches := 100

			err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if matchCount >= maxMatches {
					return filepath.SkipAll
				}

				// Skip hidden dirs and common non-text dirs
				for _, part := range strings.Split(path, string(filepath.Separator)) {
					if strings.HasPrefix(part, ".") || part == "node_modules" || part == "vendor" {
						return nil
					}
				}

				// Apply glob filter
				if globPattern != "" {
					matched, _ := filepath.Match(globPattern, filepath.Base(path))
					if !matched {
						return nil
					}
				}

				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}

				lines := strings.Split(string(data), "\n")
				for i, line := range lines {
					if re.MatchString(line) {
						fmt.Fprintf(&results, "%s:%d: %s\n", path, i+1, line)
						matchCount++
						if matchCount >= maxMatches {
							break
						}
					}
				}
				return nil
			})
			if err != nil {
				return "", err
			}

			if matchCount == 0 {
				return "No matches found.", nil
			}

			result := results.String()
			if matchCount >= maxMatches {
				result += fmt.Sprintf("\n... (truncated, showing first %d matches)", maxMatches)
			}
			return result, nil
		},
	}
}
