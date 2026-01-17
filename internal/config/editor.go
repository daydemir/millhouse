package config

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const (
	editorHeight = 30
	editorWidth  = 80
)

// Editor represents the interactive configuration editor
type Editor struct {
	config       *Config
	inputs       map[string]textinput.Model
	fieldOrder   []string
	currentField int
	err          error
	saved        bool
	message      string
	basePath     string
}

// NewEditor creates a new interactive editor
func NewEditor(basePath string, cfg *Config) *Editor {
	inputs := make(map[string]textinput.Model)

	// Create text inputs for each editable field
	fields := []struct {
		name         string
		value        string
		placeholder  string
		width        int
	}{
		// Global settings
		{"globalModel", cfg.Global.Model, "sonnet", 10},
		{"globalMaxTokens", fmt.Sprintf("%d", cfg.Global.MaxTokens), "100000", 10},

		// Planner settings
		{"plannerModel", cfg.Phases.Planner.Model, "sonnet", 10},
		{"plannerMaxTokens", fmt.Sprintf("%d", cfg.Phases.Planner.MaxTokens), "80000", 10},
		{"plannerProgressLines", fmt.Sprintf("%d", cfg.Phases.Planner.ProgressLines), "20", 10},

		// Builder settings
		{"builderModel", cfg.Phases.Builder.Model, "sonnet", 10},
		{"builderMaxTokens", fmt.Sprintf("%d", cfg.Phases.Builder.MaxTokens), "100000", 10},
		{"builderProgressLines", fmt.Sprintf("%d", cfg.Phases.Builder.ProgressLines), "20", 10},

		// Reviewer settings
		{"reviewerModel", cfg.Phases.Reviewer.Model, "sonnet", 10},
		{"reviewerMaxTokens", fmt.Sprintf("%d", cfg.Phases.Reviewer.MaxTokens), "80000", 10},
		{"reviewerProgressLines", fmt.Sprintf("%d", cfg.Phases.Reviewer.ProgressLines), "200", 10},
	}

	for _, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.placeholder
		ti.SetValue(f.value)
		ti.Width = f.width
		inputs[f.name] = ti
		// Don't focus the first field until the editor starts
	}

	return &Editor{
		config:     cfg,
		inputs:     inputs,
		basePath:   basePath,
		fieldOrder: []string{
			"globalModel", "globalMaxTokens",
			"plannerModel", "plannerMaxTokens", "plannerProgressLines",
			"builderModel", "builderMaxTokens", "builderProgressLines",
			"reviewerModel", "reviewerMaxTokens", "reviewerProgressLines",
		},
		currentField: 0,
	}
}

// RunEditor starts the interactive editor
func RunEditor(basePath string) error {
	cfg, err := Load(basePath)
	if err != nil {
		cfg = DefaultConfig()
	}

	editor := NewEditor(basePath, cfg)

	// Focus the first field
	firstInput := editor.inputs[editor.fieldOrder[0]]
	firstInput.Focus()
	editor.inputs[editor.fieldOrder[0]] = firstInput

	p := tea.NewProgram(editor, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("editor error: %w", err)
	}

	return nil
}

// Init implements the bubbletea Model interface
func (e *Editor) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements the bubbletea Model interface
func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return e, tea.Quit

		case tea.KeyCtrlS:
			if err := e.saveConfig(); err != nil {
				e.err = err
				e.message = fmt.Sprintf("Error: %v", err)
			} else {
				e.saved = true
				e.message = "✓ Configuration saved successfully!"
			}
			return e, nil

		case tea.KeyEscape:
			return e, tea.Quit

		case tea.KeyUp:
			// Move to previous field
			if e.currentField > 0 {
				currentInput := e.inputs[e.fieldOrder[e.currentField]]
				currentInput.Blur()
				e.inputs[e.fieldOrder[e.currentField]] = currentInput

				e.currentField--

				nextInput := e.inputs[e.fieldOrder[e.currentField]]
				nextInput.Focus()
				e.inputs[e.fieldOrder[e.currentField]] = nextInput

				e.err = nil
			}
			return e, nil

		case tea.KeyDown, tea.KeyTab:
			// Move to next field
			if e.currentField < len(e.fieldOrder)-1 {
				currentInput := e.inputs[e.fieldOrder[e.currentField]]
				currentInput.Blur()
				e.inputs[e.fieldOrder[e.currentField]] = currentInput

				e.currentField++

				nextInput := e.inputs[e.fieldOrder[e.currentField]]
				nextInput.Focus()
				e.inputs[e.fieldOrder[e.currentField]] = nextInput

				e.err = nil
			}
			return e, nil

		case tea.KeyEnter:
			// Move to next field on Enter
			if e.currentField < len(e.fieldOrder)-1 {
				currentInput := e.inputs[e.fieldOrder[e.currentField]]
				currentInput.Blur()
				e.inputs[e.fieldOrder[e.currentField]] = currentInput

				e.currentField++

				nextInput := e.inputs[e.fieldOrder[e.currentField]]
				nextInput.Focus()
				e.inputs[e.fieldOrder[e.currentField]] = nextInput
			}
			return e, nil
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	currentFieldName := e.fieldOrder[e.currentField]
	input := e.inputs[currentFieldName]
	input, cmd = input.Update(msg)
	e.inputs[currentFieldName] = input
	return e, cmd
}

// View implements the bubbletea Model interface
func (e *Editor) View() string {
	var s string

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		MarginTop(1).
		MarginBottom(1)

	// Header
	s += headerStyle.Render("⚙️  Millhouse Configuration Editor") + "\n"
	s += "Use ↑/↓ to navigate • Tab/Enter to move to next • Ctrl+S to save • ESC to cancel\n\n"

	// Global Settings
	s += sectionStyle.Render("Global Settings") + "\n"
	s += e.renderField("globalModel", "Model", e.currentField == 0) + "\n"
	s += e.renderField("globalMaxTokens", "Max Tokens", e.currentField == 1) + "\n"

	// Planner Settings
	s += sectionStyle.Render("Planner Phase") + "\n"
	s += e.renderField("plannerModel", "Model", e.currentField == 2) + "\n"
	s += e.renderField("plannerMaxTokens", "Max Tokens", e.currentField == 3) + "\n"
	s += e.renderField("plannerProgressLines", "Progress Lines", e.currentField == 4) + "\n"

	// Builder Settings
	s += sectionStyle.Render("Builder Phase") + "\n"
	s += e.renderField("builderModel", "Model", e.currentField == 5) + "\n"
	s += e.renderField("builderMaxTokens", "Max Tokens", e.currentField == 6) + "\n"
	s += e.renderField("builderProgressLines", "Progress Lines", e.currentField == 7) + "\n"

	// Reviewer Settings
	s += sectionStyle.Render("Reviewer Phase") + "\n"
	s += e.renderField("reviewerModel", "Model", e.currentField == 8) + "\n"
	s += e.renderField("reviewerMaxTokens", "Max Tokens", e.currentField == 9) + "\n"
	s += e.renderField("reviewerProgressLines", "Progress Lines", e.currentField == 10) + "\n"

	// Status message
	s += "\n"
	if e.saved {
		s += lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Render("✓ " + e.message) + "\n"
		e.saved = false
	} else if e.err != nil {
		s += lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render("✗ " + e.message) + "\n"
		e.err = nil
	}

	// Footer
	s += lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginTop(1).
		Render("[Ctrl+S] Save  [ESC] Cancel") + "\n"

	return s
}

// renderField renders a single configuration field
func (e *Editor) renderField(fieldName, label string, focused bool) string {
	input := e.inputs[fieldName]

	labelStyle := lipgloss.NewStyle().
		Width(15).
		Align(lipgloss.Right).
		Foreground(lipgloss.Color("8"))

	if focused {
		return lipgloss.JoinHorizontal(lipgloss.Center,
			labelStyle.Render(label+":"),
			"  ",
			lipgloss.NewStyle().
				Background(lipgloss.Color("8")).
				Padding(0, 1).
				Render(input.View()),
		)
	}

	return lipgloss.JoinHorizontal(lipgloss.Center,
		labelStyle.Render(label+":"),
		"  ",
		input.View(),
	)
}

// saveConfig saves the edited configuration
func (e *Editor) saveConfig() error {
	// Parse and validate all fields
	newConfig := DefaultConfig()

	// Global settings
	globalModel := e.inputs["globalModel"].Value()
	if globalModel != "" {
		newConfig.Global.Model = globalModel
	}

	if globalTokensStr := e.inputs["globalMaxTokens"].Value(); globalTokensStr != "" {
		if tokens, err := strconv.Atoi(globalTokensStr); err == nil {
			newConfig.Global.MaxTokens = tokens
		} else {
			return fmt.Errorf("invalid global maxTokens: %w", err)
		}
	}

	// Planner settings
	if plannerModel := e.inputs["plannerModel"].Value(); plannerModel != "" {
		newConfig.Phases.Planner.Model = plannerModel
	}

	if plannerTokensStr := e.inputs["plannerMaxTokens"].Value(); plannerTokensStr != "" {
		if tokens, err := strconv.Atoi(plannerTokensStr); err == nil {
			newConfig.Phases.Planner.MaxTokens = tokens
		} else {
			return fmt.Errorf("invalid planner maxTokens: %w", err)
		}
	}

	if plannerLinesStr := e.inputs["plannerProgressLines"].Value(); plannerLinesStr != "" {
		if lines, err := strconv.Atoi(plannerLinesStr); err == nil {
			newConfig.Phases.Planner.ProgressLines = lines
		} else {
			return fmt.Errorf("invalid planner progressLines: %w", err)
		}
	}

	// Builder settings
	if builderModel := e.inputs["builderModel"].Value(); builderModel != "" {
		newConfig.Phases.Builder.Model = builderModel
	}

	if builderTokensStr := e.inputs["builderMaxTokens"].Value(); builderTokensStr != "" {
		if tokens, err := strconv.Atoi(builderTokensStr); err == nil {
			newConfig.Phases.Builder.MaxTokens = tokens
		} else {
			return fmt.Errorf("invalid builder maxTokens: %w", err)
		}
	}

	if builderLinesStr := e.inputs["builderProgressLines"].Value(); builderLinesStr != "" {
		if lines, err := strconv.Atoi(builderLinesStr); err == nil {
			newConfig.Phases.Builder.ProgressLines = lines
		} else {
			return fmt.Errorf("invalid builder progressLines: %w", err)
		}
	}

	// Reviewer settings
	if reviewerModel := e.inputs["reviewerModel"].Value(); reviewerModel != "" {
		newConfig.Phases.Reviewer.Model = reviewerModel
	}

	if reviewerTokensStr := e.inputs["reviewerMaxTokens"].Value(); reviewerTokensStr != "" {
		if tokens, err := strconv.Atoi(reviewerTokensStr); err == nil {
			newConfig.Phases.Reviewer.MaxTokens = tokens
		} else {
			return fmt.Errorf("invalid reviewer maxTokens: %w", err)
		}
	}

	if reviewerLinesStr := e.inputs["reviewerProgressLines"].Value(); reviewerLinesStr != "" {
		if lines, err := strconv.Atoi(reviewerLinesStr); err == nil {
			newConfig.Phases.Reviewer.ProgressLines = lines
		} else {
			return fmt.Errorf("invalid reviewer progressLines: %w", err)
		}
	}

	// Validate the new config
	if err := newConfig.Validate(); err != nil {
		return err
	}

	// Save to file
	if err := Save(e.basePath, newConfig); err != nil {
		return err
	}

	return nil
}
