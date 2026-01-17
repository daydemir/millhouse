package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test that all default values are set
	if cfg.Global.Model != ModelSonnet {
		t.Errorf("Expected global model %s, got %s", ModelSonnet, cfg.Global.Model)
	}

	if cfg.Global.MaxTokens != 100000 {
		t.Errorf("Expected global maxTokens 100000, got %d", cfg.Global.MaxTokens)
	}

	// Test planner defaults
	if cfg.Phases.Planner.Model != ModelSonnet {
		t.Errorf("Expected planner model %s, got %s", ModelSonnet, cfg.Phases.Planner.Model)
	}

	if cfg.Phases.Planner.MaxTokens != 80000 {
		t.Errorf("Expected planner maxTokens 80000, got %d", cfg.Phases.Planner.MaxTokens)
	}

	if cfg.Phases.Planner.ProgressLines != 20 {
		t.Errorf("Expected planner progressLines 20, got %d", cfg.Phases.Planner.ProgressLines)
	}

	// Test builder defaults
	if cfg.Phases.Builder.Model != ModelSonnet {
		t.Errorf("Expected builder model %s, got %s", ModelSonnet, cfg.Phases.Builder.Model)
	}

	if cfg.Phases.Builder.MaxTokens != 100000 {
		t.Errorf("Expected builder maxTokens 100000, got %d", cfg.Phases.Builder.MaxTokens)
	}

	if cfg.Phases.Builder.ProgressLines != 20 {
		t.Errorf("Expected builder progressLines 20, got %d", cfg.Phases.Builder.ProgressLines)
	}

	// Test reviewer defaults
	if cfg.Phases.Reviewer.Model != ModelSonnet {
		t.Errorf("Expected reviewer model %s, got %s", ModelSonnet, cfg.Phases.Reviewer.Model)
	}

	if cfg.Phases.Reviewer.MaxTokens != 80000 {
		t.Errorf("Expected reviewer maxTokens 80000, got %d", cfg.Phases.Reviewer.MaxTokens)
	}

	if cfg.Phases.Reviewer.ProgressLines != 200 {
		t.Errorf("Expected reviewer progressLines 200, got %d", cfg.Phases.Reviewer.ProgressLines)
	}
}

func TestGetPhaseConfig(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		phase    string
		expected PhaseConfig
	}{
		{
			"planner",
			PhaseConfig{Model: ModelSonnet, MaxTokens: 80000, ProgressLines: 20},
		},
		{
			"builder",
			PhaseConfig{Model: ModelSonnet, MaxTokens: 100000, ProgressLines: 20},
		},
		{
			"reviewer",
			PhaseConfig{Model: ModelSonnet, MaxTokens: 80000, ProgressLines: 200},
		},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			phaseConfig := cfg.GetPhaseConfig(tt.phase)

			if phaseConfig.Model != tt.expected.Model {
				t.Errorf("Expected model %s, got %s", tt.expected.Model, phaseConfig.Model)
			}

			if phaseConfig.MaxTokens != tt.expected.MaxTokens {
				t.Errorf("Expected maxTokens %d, got %d", tt.expected.MaxTokens, phaseConfig.MaxTokens)
			}

			if phaseConfig.ProgressLines != tt.expected.ProgressLines {
				t.Errorf("Expected progressLines %d, got %d", tt.expected.ProgressLines, phaseConfig.ProgressLines)
			}
		})
	}
}

func TestGetPhaseConfigWithGlobalFallback(t *testing.T) {
	cfg := DefaultConfig()

	// Set global model
	cfg.Global.Model = ModelHaiku
	cfg.Global.MaxTokens = 50000

	// Clear phase-specific settings to test fallback
	cfg.Phases.Planner.Model = ""
	cfg.Phases.Planner.MaxTokens = 0

	phaseConfig := cfg.GetPhaseConfig("planner")

	if phaseConfig.Model != ModelHaiku {
		t.Errorf("Expected global fallback model %s, got %s", ModelHaiku, phaseConfig.Model)
	}

	if phaseConfig.MaxTokens != 50000 {
		t.Errorf("Expected global fallback maxTokens 50000, got %d", phaseConfig.MaxTokens)
	}

	// ProgressLines should use default, not global fallback
	if phaseConfig.ProgressLines != 20 {
		t.Errorf("Expected default progressLines 20, got %d", phaseConfig.ProgressLines)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			"valid default config",
			DefaultConfig(),
			false,
		},
		{
			"invalid model",
			&Config{
				Global: GlobalConfig{Model: "invalid-model"},
			},
			true,
		},
		{
			"invalid token limit too low",
			&Config{
				Global: GlobalConfig{Model: ModelSonnet, MaxTokens: 1000},
			},
			true,
		},
		{
			"invalid token limit too high",
			&Config{
				Global: GlobalConfig{Model: ModelSonnet, MaxTokens: 300000},
			},
			true,
		},
		{
			"valid token limit at min boundary",
			&Config{
				Global: GlobalConfig{Model: ModelSonnet, MaxTokens: MinTokens},
			},
			false,
		},
		{
			"valid token limit at max boundary",
			&Config{
				Global: GlobalConfig{Model: ModelSonnet, MaxTokens: MaxTokens},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "millhouse-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .millhouse directory
	millhouseDir := filepath.Join(tmpDir, MillhouseDir)
	if err := os.MkdirAll(millhouseDir, 0755); err != nil {
		t.Fatalf("Failed to create .millhouse directory: %v", err)
	}

	// Create a test config
	originalConfig := DefaultConfig()
	originalConfig.Phases.Planner.Model = ModelHaiku
	originalConfig.Phases.Builder.MaxTokens = 150000
	originalConfig.Phases.Reviewer.ProgressLines = 100

	// Save the config
	if err := Save(tmpDir, originalConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the config
	loadedConfig, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if loadedConfig.Phases.Planner.Model != ModelHaiku {
		t.Errorf("Expected planner model %s, got %s", ModelHaiku, loadedConfig.Phases.Planner.Model)
	}

	if loadedConfig.Phases.Builder.MaxTokens != 150000 {
		t.Errorf("Expected builder maxTokens 150000, got %d", loadedConfig.Phases.Builder.MaxTokens)
	}

	if loadedConfig.Phases.Reviewer.ProgressLines != 100 {
		t.Errorf("Expected reviewer progressLines 100, got %d", loadedConfig.Phases.Reviewer.ProgressLines)
	}
}

func TestMergeConfigs(t *testing.T) {
	base := DefaultConfig()
	base.Global.Model = ModelSonnet
	base.Global.MaxTokens = 100000
	base.Phases.Planner.Model = ModelSonnet

	override := &Config{}
	override.Global.Model = ModelHaiku
	override.Phases.Builder.Model = ModelOpus
	override.Phases.Builder.MaxTokens = 120000

	merged := mergeConfigs(base, override)

	// Check that override values are applied
	if merged.Global.Model != ModelHaiku {
		t.Errorf("Expected global model %s, got %s", ModelHaiku, merged.Global.Model)
	}

	// Check that base values are preserved when not overridden
	if merged.Global.MaxTokens != 100000 {
		t.Errorf("Expected global maxTokens 100000, got %d", merged.Global.MaxTokens)
	}

	// Check that builder was overridden
	if merged.Phases.Builder.Model != ModelOpus {
		t.Errorf("Expected builder model %s, got %s", ModelOpus, merged.Phases.Builder.Model)
	}

	if merged.Phases.Builder.MaxTokens != 120000 {
		t.Errorf("Expected builder maxTokens 120000, got %d", merged.Phases.Builder.MaxTokens)
	}

	// Check that planner was preserved
	if merged.Phases.Planner.Model != ModelSonnet {
		t.Errorf("Expected planner model %s, got %s", ModelSonnet, merged.Phases.Planner.Model)
	}
}

func TestApplyOverrides(t *testing.T) {
	cfg := DefaultConfig()

	// Apply overrides
	cfg.ApplyOverrides(ModelHaiku, ModelOpus, "", 50000, 150000, 0)

	// Check overrides were applied
	if cfg.Phases.Planner.Model != ModelHaiku {
		t.Errorf("Expected planner model %s, got %s", ModelHaiku, cfg.Phases.Planner.Model)
	}

	if cfg.Phases.Builder.Model != ModelOpus {
		t.Errorf("Expected builder model %s, got %s", ModelOpus, cfg.Phases.Builder.Model)
	}

	if cfg.Phases.Planner.MaxTokens != 50000 {
		t.Errorf("Expected planner maxTokens 50000, got %d", cfg.Phases.Planner.MaxTokens)
	}

	if cfg.Phases.Builder.MaxTokens != 150000 {
		t.Errorf("Expected builder maxTokens 150000, got %d", cfg.Phases.Builder.MaxTokens)
	}

	// Check that empty values don't override
	if cfg.Phases.Reviewer.Model != ModelSonnet {
		t.Errorf("Expected reviewer model %s, got %s", ModelSonnet, cfg.Phases.Reviewer.Model)
	}

	if cfg.Phases.Reviewer.MaxTokens != 80000 {
		t.Errorf("Expected reviewer maxTokens 80000, got %d", cfg.Phases.Reviewer.MaxTokens)
	}
}

func TestValidProgressLinesRange(t *testing.T) {
	tests := []struct {
		name          string
		progressLines int
		wantErr       bool
	}{
		{"valid at min", MinProgressLines, false},
		{"valid at max", MaxProgressLines, false},
		{"valid in range", 50, false},
		{"invalid below min", MinProgressLines - 1, true},
		{"invalid above max", MaxProgressLines + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Phases.Planner.ProgressLines = tt.progressLines

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
