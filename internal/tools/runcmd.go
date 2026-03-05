package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

// RunCmdTool creates the run_cmd tool.
func RunCmdTool() *Tool {
	return &Tool{
		Name:        "run_cmd",
		Description: "Execute a shell command and return its output. Commands have a 60-second timeout.",
		Parameters: map[string]llm.ParamDef{
			"command": {Type: "string", Description: "The shell command to execute", Required: true},
		},
		Execute: func(args map[string]any) (string, error) {
			command, _ := args["command"].(string)
			if command == "" {
				return "", fmt.Errorf("command is required")
			}

			// Safety check for destructive commands
			if isDestructive(command) {
				return "", fmt.Errorf("blocked: command appears destructive (%s). Use with caution", command)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "sh", "-c", command)
			output, err := cmd.CombinedOutput()

			var result strings.Builder
			if len(output) > 0 {
				result.Write(output)
			}

			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					result.WriteString("\n[command timed out after 60 seconds]")
				} else {
					fmt.Fprintf(&result, "\n[exit code: %s]", err.Error())
				}
			}

			// Truncate very long output
			s := result.String()
			if len(s) > 50000 {
				s = s[:25000] + "\n\n... [output truncated] ...\n\n" + s[len(s)-25000:]
			}

			return s, nil
		},
	}
}

// Destructive command patterns to block.
var destructivePatterns = []string{
	"rm -rf /",
	"rm -rf ~",
	"mkfs.",
	"dd if=",
	":(){:|:&};:",
	"chmod -R 777 /",
	"git push --force origin main",
	"git push --force origin master",
	"git reset --hard",
	"> /dev/sda",
}

func isDestructive(command string) bool {
	lower := strings.ToLower(strings.TrimSpace(command))
	for _, pattern := range destructivePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
