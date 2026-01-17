package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// Configuration directories
	MillhouseDir = ".millhouse"
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

	// Set global defaults
	cfg.Global = GlobalConfig{
		Model:     ModelSonnet,
		MaxTokens: 100000,
	}

	return cfg
}

// Load reads configuration from both user global and project-specific config files
// Precedence: project-specific (.millhouse/config.yaml) overrides user global (~/.millhouse/config.yaml)
func Load(basePath string) (*Config, error) {
	cfg := DefaultConfig()

	// Try loading user global config first
	userGlobalPath, err := getUserConfigPath()
	if err == nil {
		if userCfg, err := loadFromFile(userGlobalPath); err == nil {
			cfg = userCfg
		}
	}

	// Load project-specific config (overrides user global)
	projectPath := filepath.Join(basePath, MillhouseDir, ConfigFile)
	if projectCfg, err := loadFromFile(projectPath); err == nil {
		// Merge project config over user config
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
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// mergeConfigs merges override config into base config
// Non-empty values in override take precedence
func mergeConfigs(base, override *Config) *Config {
	result := base

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

	// Merge context files (append instead of replace)
	result.ContextFiles = append(base.ContextFiles, override.ContextFiles...)

	return result
}

// Save writes configuration to .millhouse/config.yaml in the project directory
func Save(basePath string, cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	configDir := filepath.Join(basePath, MillhouseDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, ConfigFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetPhaseConfig returns the effective configuration for a specific phase
// It applies global fallbacks for any unset phase-specific values
func (c *Config) GetPhaseConfig(phase string) PhaseConfig {
	var phaseConfig PhaseConfig

	switch phase {
	case "planner":
		phaseConfig = c.Phases.Planner
	case "builder":
		phaseConfig = c.Phases.Builder
	case "reviewer":
		phaseConfig = c.Phases.Reviewer
	default:
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

// getUserConfigPath returns the path to the user's global config file
func getUserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, MillhouseDir, ConfigFile), nil
}

// ApplyOverrides applies CLI flag overrides to the configuration
func (c *Config) ApplyOverrides(plannerModel, builderModel, reviewerModel string,
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

	if plannerTokens > 0 {
		c.Phases.Planner.MaxTokens = plannerTokens
	}
	if builderTokens > 0 {
		c.Phases.Builder.MaxTokens = builderTokens
	}
	if reviewerTokens > 0 {
		c.Phases.Reviewer.MaxTokens = reviewerTokens
	}
}
