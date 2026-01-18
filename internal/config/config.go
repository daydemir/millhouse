package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// Configuration directories
	MillhouseDir = ".milhouse"
	ConfigFile   = "config.yaml"

	// Model validation
	ModelHaiku  = "haiku"
	ModelSonnet = "sonnet"
	ModelOpus   = "opus"

	// Token limits
	MinTokens = 10000
	MaxTokens = 200000

	// Progress lines limits
	MinProgressLines = 10
	MaxProgressLines = 1000
)

// PhaseConfig represents configuration for a specific phase (planner, builder, reviewer)
type PhaseConfig struct {
	Model         string `yaml:"model,omitempty"`
	MaxTokens     int    `yaml:"maxTokens,omitempty"`
	ProgressLines int    `yaml:"progressLines,omitempty"`
}

// GlobalConfig represents global defaults applied to all phases
type GlobalConfig struct {
	Model     string `yaml:"model,omitempty"`
	MaxTokens int    `yaml:"maxTokens,omitempty"`
}

// Config represents the entire configuration structure
type Config struct {
	Phases struct {
		Planner  PhaseConfig `yaml:"planner,omitempty"`
		Builder  PhaseConfig `yaml:"builder,omitempty"`
		Reviewer PhaseConfig `yaml:"reviewer,omitempty"`
		Chat     PhaseConfig `yaml:"chat,omitempty"`
	} `yaml:"phases,omitempty"`
	Global       GlobalConfig `yaml:"global,omitempty"`
	ContextFiles []string     `yaml:"contextFiles,omitempty"`
}

// DefaultConfig returns the default configuration matching current hardcoded values
func DefaultConfig() *Config {
	cfg := &Config{}

	// Set phase-specific defaults
	cfg.Phases.Planner = PhaseConfig{
		Model:         ModelSonnet,
		MaxTokens:     80000,
		ProgressLines: 20,
	}
	cfg.Phases.Builder = PhaseConfig{
		Model:         ModelSonnet,
		MaxTokens:     100000,
		ProgressLines: 20,
	}
	cfg.Phases.Reviewer = PhaseConfig{
		Model:         ModelSonnet,
		MaxTokens:     80000,
		ProgressLines: 200,
	}
	cfg.Phases.Chat = PhaseConfig{
		Model: ModelSonnet,
		// No MaxTokens - chat runs in interactive mode without token limits
	}

	// Set global defaults
	cfg.Global = GlobalConfig{
		Model:     ModelSonnet,
		MaxTokens: 100000,
	}

	return cfg
}

// Load reads configuration from project-specific config file
// Precedence: project config (.milhouse/config.yaml) overrides defaults
func Load(basePath string) (*Config, error) {
	cfg := DefaultConfig()

	// Load project-specific config (overrides defaults)
	projectPath := filepath.Join(basePath, MillhouseDir, ConfigFile)
	if projectCfg, err := loadFromFile(projectPath); err == nil {
		// Merge project config over defaults
		cfg = mergeConfigs(cfg, projectCfg)
	}

	// Validate the final config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return cfg, nil
}

// deduplicateStrings removes duplicate strings from a slice while preserving order
func deduplicateStrings(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// mergeConfigs merges override config into base config
// Non-empty values in override take precedence
// Creates a deep copy to avoid mutating the base config
func mergeConfigs(base, override *Config) *Config {
	// Create a new config, copying all values from base
	result := &Config{
		Global: GlobalConfig{
			Model:     base.Global.Model,
			MaxTokens: base.Global.MaxTokens,
		},
		ContextFiles: make([]string, len(base.ContextFiles)),
	}
	copy(result.ContextFiles, base.ContextFiles)

	// Copy phase configs
	result.Phases.Planner = base.Phases.Planner
	result.Phases.Builder = base.Phases.Builder
	result.Phases.Reviewer = base.Phases.Reviewer
	result.Phases.Chat = base.Phases.Chat

	// Merge global config
	if override.Global.Model != "" {
		result.Global.Model = override.Global.Model
	}
	if override.Global.MaxTokens != 0 {
		result.Global.MaxTokens = override.Global.MaxTokens
	}

	// Merge phase configs
	if override.Phases.Planner.Model != "" {
		result.Phases.Planner.Model = override.Phases.Planner.Model
	}
	if override.Phases.Planner.MaxTokens != 0 {
		result.Phases.Planner.MaxTokens = override.Phases.Planner.MaxTokens
	}
	if override.Phases.Planner.ProgressLines != 0 {
		result.Phases.Planner.ProgressLines = override.Phases.Planner.ProgressLines
	}

	if override.Phases.Builder.Model != "" {
		result.Phases.Builder.Model = override.Phases.Builder.Model
	}
	if override.Phases.Builder.MaxTokens != 0 {
		result.Phases.Builder.MaxTokens = override.Phases.Builder.MaxTokens
	}
	if override.Phases.Builder.ProgressLines != 0 {
		result.Phases.Builder.ProgressLines = override.Phases.Builder.ProgressLines
	}

	if override.Phases.Reviewer.Model != "" {
		result.Phases.Reviewer.Model = override.Phases.Reviewer.Model
	}
	if override.Phases.Reviewer.MaxTokens != 0 {
		result.Phases.Reviewer.MaxTokens = override.Phases.Reviewer.MaxTokens
	}
	if override.Phases.Reviewer.ProgressLines != 0 {
		result.Phases.Reviewer.ProgressLines = override.Phases.Reviewer.ProgressLines
	}

	if override.Phases.Chat.Model != "" {
		result.Phases.Chat.Model = override.Phases.Chat.Model
	}
	// No MaxTokens or ProgressLines for chat (interactive mode)

	// Merge context files with deduplication
	allFiles := append(base.ContextFiles, override.ContextFiles...)
	result.ContextFiles = deduplicateStrings(allFiles)

	return result
}

// Save writes configuration to .milhouse/config.yaml in the project directory
func Save(basePath string, cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	configDir := filepath.Join(basePath, MillhouseDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	configPath := filepath.Join(configDir, ConfigFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configPath, err)
	}

	return nil
}

// GetPhaseConfig returns the effective configuration for a specific phase
// It applies global fallbacks for any unset phase-specific values
func (c *Config) GetPhaseConfig(phase string) PhaseConfig {
	// Nil guard - return default if receiver is nil
	if c == nil {
		return DefaultConfig().GetPhaseConfig(phase)
	}

	var phaseConfig PhaseConfig

	switch phase {
	case "planner":
		phaseConfig = c.Phases.Planner
	case "builder":
		phaseConfig = c.Phases.Builder
	case "reviewer":
		phaseConfig = c.Phases.Reviewer
	case "chat":
		phaseConfig = c.Phases.Chat
	default:
		log.Printf("Warning: unknown phase '%s', using planner config as fallback", phase)
		phaseConfig = c.Phases.Planner // default fallback
	}

	// Apply global fallbacks
	if phaseConfig.Model == "" {
		phaseConfig.Model = c.Global.Model
	}
	if phaseConfig.MaxTokens == 0 {
		phaseConfig.MaxTokens = c.Global.MaxTokens
	}

	// For progress lines, we don't have a global default, so use phase defaults
	// This is because different phases may need different amounts of history
	if phaseConfig.ProgressLines == 0 {
		// Return the original phase's default (not modified by global)
		switch phase {
		case "planner":
			phaseConfig.ProgressLines = 20
		case "builder":
			phaseConfig.ProgressLines = 20
		case "reviewer":
			phaseConfig.ProgressLines = 200
		case "chat":
			phaseConfig.ProgressLines = 0 // Chat doesn't use progress lines
		}
	}

	return phaseConfig
}

// Validate checks that configuration values are within acceptable ranges
func (c *Config) Validate() error {
	validModels := map[string]bool{
		ModelHaiku:  true,
		ModelSonnet: true,
		ModelOpus:   true,
	}

	// Validate global config
	if c.Global.Model != "" && !validModels[c.Global.Model] {
		return fmt.Errorf("invalid global model '%s': must be 'haiku', 'sonnet', or 'opus'", c.Global.Model)
	}
	if c.Global.MaxTokens != 0 && (c.Global.MaxTokens < MinTokens || c.Global.MaxTokens > MaxTokens) {
		return fmt.Errorf("invalid global maxTokens %d: must be between %d and %d", c.Global.MaxTokens, MinTokens, MaxTokens)
	}

	// Validate phase configs
	phases := []struct {
		name   string
		config PhaseConfig
	}{
		{"planner", c.Phases.Planner},
		{"builder", c.Phases.Builder},
		{"reviewer", c.Phases.Reviewer},
		{"chat", c.Phases.Chat},
	}

	for _, p := range phases {
		if p.config.Model != "" && !validModels[p.config.Model] {
			return fmt.Errorf("invalid %s model '%s': must be 'haiku', 'sonnet', or 'opus'", p.name, p.config.Model)
		}
		if p.config.MaxTokens != 0 && (p.config.MaxTokens < MinTokens || p.config.MaxTokens > MaxTokens) {
			return fmt.Errorf("invalid %s maxTokens %d: must be between %d and %d", p.name, p.config.MaxTokens, MinTokens, MaxTokens)
		}
		if p.config.ProgressLines != 0 && (p.config.ProgressLines < MinProgressLines || p.config.ProgressLines > MaxProgressLines) {
			return fmt.Errorf("invalid %s progressLines %d: must be between %d and %d", p.name, p.config.ProgressLines, MinProgressLines, MaxProgressLines)
		}
	}

	return nil
}

// ApplyOverrides applies CLI flag overrides to the configuration
func (c *Config) ApplyOverrides(plannerModel, builderModel, reviewerModel, chatModel string,
	plannerTokens, builderTokens, reviewerTokens int) {
	if plannerModel != "" {
		c.Phases.Planner.Model = plannerModel
	}
	if builderModel != "" {
		c.Phases.Builder.Model = builderModel
	}
	if reviewerModel != "" {
		c.Phases.Reviewer.Model = reviewerModel
	}
	if chatModel != "" {
		c.Phases.Chat.Model = chatModel
	}

	if plannerTokens > 0 {
		c.Phases.Planner.MaxTokens = plannerTokens
	}
	if builderTokens > 0 {
		c.Phases.Builder.MaxTokens = builderTokens
	}
	if reviewerTokens > 0 {
		c.Phases.Reviewer.MaxTokens = reviewerTokens
	}
	// No chatTokens parameter - chat doesn't use token limits
}
