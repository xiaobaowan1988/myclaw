package context

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xiaobaowan1988/openclaw/internal/llm"
)

// Manager handles context window management for the agent.
type Manager struct {
	maxMessages int
	repoRoot    string
}

// NewManager creates a new context manager.
func NewManager(maxMessages int) *Manager {
	root := detectGitRoot()
	return &Manager{
		maxMessages: maxMessages,
		repoRoot:    root,
	}
}

// RepoRoot returns the detected git repository root, or empty string.
func (m *Manager) RepoRoot() string {
	return m.repoRoot
}

// TrimHistory trims conversation history to stay within limits.
// It preserves the first message (initial user prompt) and the most recent messages.
func (m *Manager) TrimHistory(messages []llm.Message) []llm.Message {
	if len(messages) <= m.maxMessages {
		return messages
	}

	// Keep first message + last (maxMessages-1) messages
	keep := m.maxMessages - 1
	result := make([]llm.Message, 0, m.maxMessages)
	result = append(result, messages[0])
	result = append(result, messages[len(messages)-keep:]...)
	return result
}

// RepoSummary returns a brief summary of the repository structure.
func (m *Manager) RepoSummary() string {
	if m.repoRoot == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Git repository: %s\n", m.repoRoot))

	// List top-level files and dirs
	entries, err := os.ReadDir(m.repoRoot)
	if err != nil {
		return sb.String()
	}

	sb.WriteString("Top-level structure:\n")
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if entry.IsDir() {
			sb.WriteString(fmt.Sprintf("  %s/\n", entry.Name()))
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", entry.Name()))
		}
	}

	return sb.String()
}

func detectGitRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: walk up from cwd
		cwd, err := os.Getwd()
		if err != nil {
			return ""
		}
		dir := cwd
		for {
			if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		return ""
	}
	return strings.TrimSpace(string(output))
}
