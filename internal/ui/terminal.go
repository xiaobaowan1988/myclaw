package ui

import (
	"fmt"
	"strings"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	BoldGreen = "\033[1;32m"
	BoldCyan  = "\033[1;36m"
	BoldBlue  = "\033[1;34m"
)

// Printer handles terminal output with colors and formatting.
type Printer struct {
	NoColor bool
}

// NewPrinter creates a new terminal printer.
func NewPrinter(noColor bool) *Printer {
	return &Printer{NoColor: noColor}
}

// color wraps text in ANSI color if color is enabled.
func (p *Printer) color(code, text string) string {
	if p.NoColor {
		return text
	}
	return code + text + Reset
}

// Banner prints the startup banner.
func (p *Printer) Banner() {
	fmt.Println(p.color(BoldCyan, `
   ___                    ____ _
  / _ \ _ __   ___ _ __  / ___| | __ ___      __
 | | | | '_ \ / _ \ '_ \| |   | |/ _' \ \ /\ / /
 | |_| | |_) |  __/ | | | |___| | (_| |\ V  V /
  \___/| .__/ \___|_| |_|\____|_|\__,_| \_/\_/
       |_|`))
	fmt.Println(p.color(Dim, "  AI Coding Agent powered by Gemini Flash"))
	fmt.Println()
}

// Prompt prints the input prompt.
func (p *Printer) Prompt() {
	fmt.Print(p.color(BoldGreen, "❯ "))
}

// ToolCall prints a tool invocation indicator.
func (p *Printer) ToolCall(name string, args map[string]any) {
	var detail string
	switch name {
	case "read_file":
		detail = fmt.Sprintf("%v", args["path"])
	case "write_file":
		detail = fmt.Sprintf("%v", args["path"])
	case "edit_file":
		detail = fmt.Sprintf("%v", args["path"])
	case "search":
		detail = fmt.Sprintf("pattern=%v", args["pattern"])
	case "glob":
		detail = fmt.Sprintf("%v", args["pattern"])
	case "list_dir":
		detail = fmt.Sprintf("%v", args["path"])
	case "run_cmd":
		cmd := fmt.Sprintf("%v", args["command"])
		if len(cmd) > 60 {
			cmd = cmd[:60] + "..."
		}
		detail = cmd
	default:
		detail = name
	}
	fmt.Println(p.color(Yellow, fmt.Sprintf("  [%s: %s]", name, detail)))
}

// Error prints an error message.
func (p *Printer) Error(msg string) {
	fmt.Println(p.color(Red, "Error: "+msg))
}

// Info prints an info message.
func (p *Printer) Info(msg string) {
	fmt.Println(p.color(Dim, msg))
}

// Separator prints a horizontal line.
func (p *Printer) Separator() {
	fmt.Println(p.color(Dim, strings.Repeat("─", 60)))
}
