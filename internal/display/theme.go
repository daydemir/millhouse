package display

import (
	"github.com/fatih/color"
)

// Box drawing characters
const (
	BoxTopLeft     = "┌"
	BoxTopRight    = "┐"
	BoxBottomLeft  = "└"
	BoxBottomRight = "┘"
	BoxHorizontal  = "─"
	BoxVertical    = "│"
	BoxHeavy       = "━"
)

// Status symbols
const (
	SymbolCheck    = "✓"
	SymbolCross    = "✗"
	SymbolWarning  = "⚠"
	SymbolRefresh  = "↻"
	SymbolCircle   = "○"
	SymbolHalf     = "◐"
	SymbolDot      = "·"
	SymbolArrow    = "→"
)

// Gutter constants
const (
	GutterClaude   = "│"   // For Claude output
	GutterAnalyzer = "○"   // For analyzer output
	GutterCont     = "·"   // For continuation lines
)

// Theme holds color functions for styled output
type Theme struct {
	// Millhouse orchestration colors (prominent)
	MillhouseTitle  *color.Color
	MillhouseBox    *color.Color
	MillhouseText   *color.Color
	SectionBreak    *color.Color

	// Claude output colors (subdued)
	ClaudeGutter    *color.Color
	ClaudeText      *color.Color
	ClaudeTimestamp *color.Color
	ClaudeToolBadge *color.Color
	ClaudeTokens    *color.Color

	// Analyzer colors
	AnalyzerGutter  *color.Color
	AnalyzerText    *color.Color

	// Status colors
	Success         *color.Color
	Error           *color.Color
	Warning         *color.Color
	Info            *color.Color

	// Dimmed/secondary
	Dim             *color.Color
	Bold            *color.Color

	// Active PRD highlighting
	ActivePRD       *color.Color
}

// DefaultTheme returns the default color theme
func DefaultTheme() *Theme {
	return &Theme{
		// Millhouse orchestration - cyan and bold for prominence
		MillhouseTitle:  color.New(color.FgCyan, color.Bold),
		MillhouseBox:    color.New(color.FgCyan),
		MillhouseText:   color.New(color.FgWhite),
		SectionBreak:    color.New(color.FgCyan, color.Bold),

		// Claude output - dim gray for subdued appearance
		ClaudeGutter:    color.New(color.FgHiBlack),
		ClaudeText:      color.New(color.FgHiBlack),
		ClaudeTimestamp: color.New(color.FgHiBlack),
		ClaudeToolBadge: color.New(color.FgBlue),
		ClaudeTokens:    color.New(color.FgHiBlack),

		// Analyzer - yellow tint
		AnalyzerGutter:  color.New(color.FgYellow),
		AnalyzerText:    color.New(color.FgYellow),

		// Status indicators
		Success:         color.New(color.FgGreen),
		Error:           color.New(color.FgRed),
		Warning:         color.New(color.FgYellow),
		Info:            color.New(color.FgBlue),

		// Text styles
		Dim:             color.New(color.FgHiBlack),
		Bold:            color.New(color.Bold),

		// Active PRD
		ActivePRD:       color.New(color.FgHiGreen, color.Bold),
	}
}

// NoColorTheme returns a theme with colors disabled
func NoColorTheme() *Theme {
	noColor := color.New()
	return &Theme{
		MillhouseTitle:  noColor,
		MillhouseBox:    noColor,
		MillhouseText:   noColor,
		SectionBreak:    noColor,
		ClaudeGutter:    noColor,
		ClaudeText:      noColor,
		ClaudeTimestamp: noColor,
		ClaudeToolBadge: noColor,
		ClaudeTokens:    noColor,
		AnalyzerGutter:  noColor,
		AnalyzerText:    noColor,
		Success:         noColor,
		Error:           noColor,
		Warning:         noColor,
		Info:            noColor,
		Dim:             noColor,
		Bold:            noColor,
		ActivePRD:       noColor,
	}
}
