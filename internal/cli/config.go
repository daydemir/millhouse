package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/suelio/millhouse/internal/config"
	"github.com/suelio/millhouse/internal/display"
	"github.com/suelio/millhouse/internal/prd"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage.milhouse configuration",
	Long: `Manage configuration for model selection, token limits, and other settings.

Configuration is loaded from:
1. CLI flags (highest priority)
2. .milhouse/config.yaml (project-specific)
3. Built-in defaults (lowest priority)`,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open interactive configuration editor",
	Long:  "Open an interactive TUI to edit configuration settings for planner, builder, and reviewer phases.",
	RunE:  runConfigEdit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current effective configuration",
	Long:  "Show the effective configuration (merged from all sources) in YAML format.",
	RunE:  runConfigShow,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration with defaults",
	Long:  "Create a .milhouse/config.yaml file with default settings in the current project.",
	RunE:  runConfigInit,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if !prd.MillhouseExists(cwd) {
		display.Error(".milhouse/ directory not found")
		display.Info("Run 'mil init' to initialize")
		return fmt.Errorf("not initialized")
	}

	display.Header("Millhouse Configuration Editor")
	display.Info("Opening interactive editor...")
	display.Divider()

	if err := config.RunEditor(cwd); err != nil {
		return fmt.Errorf("editor error: %w", err)
	}

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if !prd.MillhouseExists(cwd) {
		display.Error(".milhouse/ directory not found")
		display.Info("Run 'mil init' to initialize")
		return fmt.Errorf("not initialized")
	}

	// Load the effective configuration
	cfg, err := config.Load(cwd)
	if err != nil {
		display.Warning(fmt.Sprintf("Warning: %v", err))
		cfg = config.DefaultConfig()
	}

	// Marshal to YAML for display
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to format config: %w", err)
	}

	display.Header("Current Configuration")
	display.Info("Effective configuration (merged from all sources):")
	display.Divider()
	fmt.Print(string(data))

	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if !prd.MillhouseExists(cwd) {
		display.Error(".milhouse/ directory not found")
		display.Info("Run 'mil init' to initialize")
		return fmt.Errorf("not initialized")
	}

	// Check if config already exists
	configPath := config.MillhouseDir + "/" + config.ConfigFile
	if _, err := os.Stat(configPath); err == nil {
		display.Warning(fmt.Sprintf("Configuration file already exists: %s", configPath))
		return nil
	}

	// Create default config and save
	cfg := config.DefaultConfig()
	if err := config.Save(cwd, cfg); err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
	}

	display.Success(fmt.Sprintf("Created configuration file: %s", configPath))
	display.Info("You can now edit it with: mil config edit")
	display.Info("Or view it with: mil config show")

	return nil
}
