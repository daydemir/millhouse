package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/daydemir/milhouse/internal/builder"
	"github.com/daydemir/milhouse/internal/config"
	"github.com/daydemir/milhouse/internal/display"
	"github.com/daydemir/milhouse/internal/prd"
)

var chatModelFlag string

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Interactive Claude session for PRD management",
	Long: `Start an interactive Claude session to:
  - Add new PRDs to prd.json
  - Update existing PRDs (description, criteria, priority)
  - Remove PRDs from prd.json
  - Explore codebase and update prompt.md
  - Answer questions about project state

This command launches an interactive session - no arguments needed.`,
	RunE: runChat,
}

func init() {
	chatCmd.Flags().StringVar(&chatModelFlag, "model", "", "Override chat model (haiku, sonnet, opus)")
	rootCmd.AddCommand(chatCmd)
}

func runChat(cmd *cobra.Command, args []string) error {
	// Reject any arguments since chat is interactive-only
	if len(args) > 0 {
		display.Warning("The 'chat' command is interactive and does not accept arguments")
		display.Info("Simply run 'mil chat' to start an interactive session")
		return fmt.Errorf("unexpected arguments")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if !prd.MillhouseExists(cwd) {
		display.Error(".milhouse/ directory not found")
		display.Info("Run 'mil init' to initialize")
		return fmt.Errorf("not initialized")
	}

	prdFile, err := prd.Load(cwd)
	if err != nil {
		return fmt.Errorf("failed to load PRDs: %w", err)
	}

	// Load configuration
	cfg, err := config.Load(cwd)
	if err != nil {
		display.Warning(fmt.Sprintf("Failed to load config: %v, using defaults", err))
		cfg = config.DefaultConfig()
	}

	// Apply CLI flag override
	if chatModelFlag != "" {
		cfg.Phases.Chat.Model = chatModelFlag
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		display.Error(fmt.Sprintf("Invalid configuration: %v", err))
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create context for the session
	ctx := context.Background()

	display.Header("Milhouse Chat")
	display.Info("Starting interactive session...")
	display.Divider()

	// Run interactive Claude session
	if err := builder.RunChat(ctx, cwd, prdFile, cfg); err != nil {
		return fmt.Errorf("chat session error: %w", err)
	}

	return nil
}
