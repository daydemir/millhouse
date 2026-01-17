package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/suelio/millhouse/internal/display"
	"github.com/suelio/millhouse/internal/prd"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Millhouse project",
	Long: `Create a .millhouse/ directory with starter files:
  - prd.json     PRD list with statuses
  - progress.md  Append-only observations
  - prompt.md    Codebase context for agents
  - evidence/    Evidence files for pending PRDs`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	millhousePath := filepath.Join(cwd, prd.MillhouseDir)

	// Check if .millhouse already exists
	if prd.MillhouseExists(cwd) {
		display.Error(".millhouse/ directory already exists")
		return fmt.Errorf("already initialized")
	}

	display.Header("Initializing Millhouse")

	// Create .millhouse directory
	if err := os.MkdirAll(millhousePath, 0755); err != nil {
		return fmt.Errorf("failed to create .millhouse directory: %w", err)
	}
	display.Success("Created .millhouse/")

	// Create evidence subdirectory
	evidencePath := filepath.Join(millhousePath, prd.EvidenceDir)
	if err := os.MkdirAll(evidencePath, 0755); err != nil {
		return fmt.Errorf("failed to create evidence directory: %w", err)
	}
	display.Success("Created .millhouse/evidence/")

	// Create empty prd.json
	prdContent := `{
  "prds": []
}
`
	if err := os.WriteFile(filepath.Join(millhousePath, prd.PRDFile), []byte(prdContent), 0644); err != nil {
		return fmt.Errorf("failed to create prd.json: %w", err)
	}
	display.Success("Created .millhouse/prd.json")

	// Create progress.md with header
	progressContent := fmt.Sprintf(`# Millhouse Progress Log

Initialized: %s

## Codebase Patterns
<!-- Add discovered patterns here for future agents -->

---

`, time.Now().Format("2006-01-02 15:04:05"))
	if err := os.WriteFile(filepath.Join(millhousePath, prd.ProgressFile), []byte(progressContent), 0644); err != nil {
		return fmt.Errorf("failed to create progress.md: %w", err)
	}
	display.Success("Created .millhouse/progress.md")

	// Create placeholder prompt.md
	promptContent := `# Codebase Context

This file provides context about the codebase for the autonomous agents.
Run 'mill discuss' to have Claude help map your codebase.

## Project Overview
<!-- Describe what this project does -->

## Directory Structure
<!-- Key directories and their purposes -->

## Technology Stack
<!-- Languages, frameworks, tools -->

## Build & Test Commands
<!-- How to build, test, lint, etc. -->

## Code Patterns
<!-- Important patterns and conventions -->

## Key Files
<!-- Critical files that agents should know about -->
`
	if err := os.WriteFile(filepath.Join(millhousePath, prd.PromptFile), []byte(promptContent), 0644); err != nil {
		return fmt.Errorf("failed to create prompt.md: %w", err)
	}
	display.Success("Created .millhouse/prompt.md")

	display.Info("Run 'mill discuss' to add PRDs and map your codebase")
	display.Info("Run 'mill status' to see PRD status")

	return nil
}
