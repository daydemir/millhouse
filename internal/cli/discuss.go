package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/suelio/millhouse/internal/builder"
	"github.com/suelio/millhouse/internal/config"
	"github.com/suelio/millhouse/internal/display"
	"github.com/suelio/millhouse/internal/prd"
)

var discussCmd = &cobra.Command{
	Use:   "discuss",
	Short: "Interactive Claude session for PRD management",
	Long: `Start an interactive Claude session to:
  - Add new PRDs to prd.json
  - Update existing PRDs (description, criteria, priority)
  - Remove PRDs from prd.json
  - Explore codebase and update prompt.md
  - Answer questions about project state`,
	RunE: runDiscuss,
}

func init() {
	rootCmd.AddCommand(discussCmd)
}

func runDiscuss(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if !prd.MillhouseExists(cwd) {
		display.Error(".millhouse/ directory not found")
		display.Info("Run 'mill init' to initialize")
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

	// Create context for the session
	ctx := context.Background()

	display.Header("Millhouse Discuss")
	display.Info("Starting interactive session...")
	display.Divider()

	// Run interactive Claude session
	if err := builder.RunDiscuss(ctx, cwd, prdFile, cfg); err != nil {
		return fmt.Errorf("discuss session error: %w", err)
	}

	return nil
}
