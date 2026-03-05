package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xiaobaowan1988/openclaw/config"
	"github.com/xiaobaowan1988/openclaw/internal/agent"
	"github.com/xiaobaowan1988/openclaw/internal/llm"
	"github.com/xiaobaowan1988/openclaw/internal/tools"
	"github.com/xiaobaowan1988/openclaw/internal/ui"
)

var (
	flagModel   string
	flagNoColor bool
	flagVerbose bool
)

var rootCmd = &cobra.Command{
	Use:   "openclaw [message]",
	Short: "OpenClaw - AI Coding Agent powered by Gemini Flash",
	Long:  "OpenClaw is an open-source AI coding agent that helps developers with code generation, review, refactoring, and debugging using Google's Gemini Flash model.",
	RunE:  runAgent,
}

func init() {
	rootCmd.Flags().StringVar(&flagModel, "model", "", "Gemini model to use (default: gemini-2.0-flash)")
	rootCmd.Flags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	rootCmd.Flags().BoolVar(&flagVerbose, "verbose", false, "Enable verbose output")
}

func Execute() error {
	return rootCmd.Execute()
}

func runAgent(cmd *cobra.Command, args []string) error {
	printer := ui.NewPrinter(flagNoColor)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		printer.Error(err.Error())
		return err
	}

	if flagModel != "" {
		cfg.Model = flagModel
	}

	// Create LLM client
	client, err := llm.NewClient(cfg.APIKey, cfg.Model, llm.ClientConfig{
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
	})
	if err != nil {
		printer.Error(err.Error())
		return err
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Create agent
	ag := agent.New(agent.Config{
		Client:   client,
		Registry: registry,
		OnText: func(chunk string) {
			fmt.Print(chunk)
		},
		OnTool: func(name string, args map[string]any) {
			printer.ToolCall(name, args)
		},
	})

	// One-shot mode: message passed as argument
	if len(args) > 0 {
		message := strings.Join(args, " ")
		_, err := ag.Run(context.Background(), message)
		if err != nil {
			printer.Error(err.Error())
			return err
		}
		fmt.Println()
		return nil
	}

	// Interactive mode
	printer.Banner()
	printer.Info("Type your message and press Enter. Type 'exit' or 'quit' to leave.")
	printer.Info(fmt.Sprintf("Model: %s", cfg.Model))
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for {
		printer.Prompt()
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			printer.Info("Goodbye!")
			break
		}
		if input == "/reset" {
			ag.Reset()
			printer.Info("Conversation reset.")
			continue
		}

		fmt.Println()
		_, err := ag.Run(context.Background(), input)
		if err != nil {
			printer.Error(err.Error())
		}
		fmt.Println()
		fmt.Println()
	}

	return nil
}
