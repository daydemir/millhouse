package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/daydemir/milhouse/internal/display"
	"github.com/daydemir/milhouse/internal/prd"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Milhouse project",
	Long: `Create a .milhouse/ directory with starter files:
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

	milhousePath := filepath.Join(cwd, prd.MillhouseDir)

	// Check if .milhouse already exists
	if prd.MillhouseExists(cwd) {
		display.Error(".milhouse/ directory already exists")
		return fmt.Errorf("already initialized")
	}

	display.Header("Initializing Milhouse")

	// Create .milhouse directory
	if err := os.MkdirAll(milhousePath, 0755); err != nil {
		return fmt.Errorf("failed to create .milhouse directory: %w", err)
	}
	display.Success("Created .milhouse/")

	// Create evidence subdirectory
	evidencePath := filepath.Join(milhousePath, prd.EvidenceDir)
	if err := os.MkdirAll(evidencePath, 0755); err != nil {
		return fmt.Errorf("failed to create evidence directory: %w", err)
	}
	display.Success("Created .milhouse/evidence/")

	// Create prompts subdirectory
	promptsPath := filepath.Join(milhousePath, prd.PromptsDir)
	if err := os.MkdirAll(promptsPath, 0755); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}
	display.Success("Created .milhouse/prompts/")

	// Create empty prd.json
	prdContent := `{
  "prds": []
}
`
	if err := os.WriteFile(filepath.Join(milhousePath, prd.PRDFile), []byte(prdContent), 0644); err != nil {
		return fmt.Errorf("failed to create prd.json: %w", err)
	}
	display.Success("Created .milhouse/prd.json")

	// Create progress.md with header
	progressContent := fmt.Sprintf(`# Milhouse Progress Log

Initialized: %s

## Codebase Patterns
<!-- Add discovered patterns here for future agents -->

---

`, time.Now().Format("2006-01-02 15:04:05"))
	if err := os.WriteFile(filepath.Join(milhousePath, prd.ProgressFile), []byte(progressContent), 0644); err != nil {
		return fmt.Errorf("failed to create progress.md: %w", err)
	}
	display.Success("Created .milhouse/progress.md")

	// Create placeholder prompt.md
	promptContent := `# Codebase Context

This file provides context about the codebase for the autonomous agents.
Run 'mil chat' to have Claude help map your codebase.

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
	if err := os.WriteFile(filepath.Join(milhousePath, prd.PromptFile), []byte(promptContent), 0644); err != nil {
		return fmt.Errorf("failed to create prompt.md: %w", err)
	}
	display.Success("Created .milhouse/prompt.md")

	// Create empty augmentation files (users add content as needed)
	augmentationFiles := []string{"planner.md", "builder.md", "reviewer.md", "chat.md"}
	for _, filename := range augmentationFiles {
		emptyContent := ""
		if err := os.WriteFile(filepath.Join(promptsPath, filename), []byte(emptyContent), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}
	display.Success("Created .milhouse/prompts/ augmentation files")

	display.Info("Run 'mil chat' to add PRDs and map your codebase")
	display.Info("Run 'mil status' to see PRD status")
	display.Info("Customize agent behavior by editing .milhouse/prompts/*.md files")

	return nil
}
