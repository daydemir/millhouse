package display

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/suelio/millhouse/internal/prd"
)

const (
	defaultTermWidth = 80
	minTermWidth     = 40
	maxTermWidth     = 120
)

// Display handles styled terminal output
type Display struct {
	theme     *Theme
	termWidth int
	noColor   bool
}

// New creates a new Display with default settings
func New() *Display {
	return &Display{
		theme:     DefaultTheme(),
		termWidth: getTerminalWidth(),
		noColor:   false,
	}
}

// NewWithOptions creates a Display with custom options
func NewWithOptions(noColor bool) *Display {
	var theme *Theme
	if noColor {
		theme = NoColorTheme()
	} else {
		theme = DefaultTheme()
	}
	return &Display{
		theme:     theme,
		termWidth: getTerminalWidth(),
		noColor:   noColor,
	}
}

// getTerminalWidth returns the current terminal width
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width < minTermWidth {
		return defaultTermWidth
	}
	if width > maxTermWidth {
		return maxTermWidth
	}
	return width
}

// GetWidth returns the current terminal width
func (d *Display) GetWidth() int {
	return d.termWidth
}

// MillhouseBox prints a boxed message for Millhouse orchestration
func (d *Display) MillhouseBox(title string, lines ...string) {
	width := d.termWidth - 2 // Account for box borders

	// Top border with title
	titlePart := fmt.Sprintf("%s %s ", BoxTopLeft+BoxHorizontal, title)
	remaining := width - len(title) - 4
	if remaining < 0 {
		remaining = 0
	}
	topBorder := titlePart + strings.Repeat(BoxHorizontal, remaining) + BoxTopRight

	d.theme.MillhouseBox.Println(topBorder)

	// Content lines
	for _, line := range lines {
		wrapped := wrapText(line, width-4)
		for _, wl := range wrapped {
			d.theme.MillhouseBox.Print(BoxVertical + " ")
			d.theme.MillhouseText.Print(fmt.Sprintf("%-*s", width-4, wl))
			d.theme.MillhouseBox.Println(" " + BoxVertical)
		}
	}

	// Bottom border
	bottomBorder := BoxBottomLeft + strings.Repeat(BoxHorizontal, width) + BoxBottomRight
	d.theme.MillhouseBox.Println(bottomBorder)
}

// SectionBreak prints a heavy horizontal line for section separation
func (d *Display) SectionBreak() {
	d.theme.SectionBreak.Println(strings.Repeat(BoxHeavy, d.termWidth))
}

// IterationHeader prints the iteration header with section breaks
func (d *Display) IterationHeader(n, total int) {
	fmt.Println()
	d.SectionBreak()
	d.theme.MillhouseTitle.Printf("Iteration %d/%d\n", n, total)
	d.SectionBreak()
}

// Claude prints Claude output with gutter prefix
func (d *Display) Claude(text string, toolCount int) {
	d.ClaudeWithTokens(text, toolCount, 0, 0)
}

// ClaudeWithTokens prints Claude output with timestamp, gutter, tool count, and tokens
func (d *Display) ClaudeWithTokens(text string, toolCount int, usedTokens, maxTokens int) {
	timestamp := time.Now().Format("15:04:05")

	// Build the prefix: [timestamp] │
	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)
	d.theme.ClaudeGutter.Print(GutterClaude + " ")

	// Tool badge if any
	if toolCount > 0 {
		d.theme.ClaudeToolBadge.Printf("[%d] ", toolCount)
	}

	// Token display if provided
	if usedTokens > 0 && maxTokens > 0 {
		percentage := float64(usedTokens) / float64(maxTokens) * 100
		d.theme.ClaudeTokens.Printf("[%.1fK/%.0fK] ", float64(usedTokens)/1000, float64(maxTokens)/1000)
		_ = percentage // Could use for color selection
	}

	// Print the text
	d.theme.ClaudeText.Println(CleanText(text))
}

// ClaudeContinuation prints a continuation line with subdued gutter
func (d *Display) ClaudeContinuation(text string) {
	timestamp := time.Now().Format("15:04:05")
	d.theme.ClaudeGutter.Printf("  %s [%s] ", GutterCont, timestamp)
	d.theme.ClaudeText.Println(CleanText(text))
}

// ClaudeStreaming prints streaming Claude text (no newline)
func (d *Display) ClaudeStreaming(text string) {
	d.theme.ClaudeText.Print(text)
}

// AnalysisStart prints the analyzer start indicator
func (d *Display) AnalysisStart() {
	timestamp := time.Now().Format("15:04:05")
	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)
	d.theme.AnalyzerGutter.Printf("%s ", GutterAnalyzer)
	d.theme.AnalyzerText.Println("[analyzer] Starting analysis...")
}

// Analysis prints analyzer output with distinctive gutter
func (d *Display) Analysis(text string) {
	timestamp := time.Now().Format("15:04:05")
	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)
	d.theme.AnalyzerGutter.Printf("%s ", GutterAnalyzer)
	d.theme.AnalyzerText.Println(CleanText(text))
}

// Header prints a styled header (backwards compatible)
func (d *Display) Header(text string) {
	fmt.Println()
	d.MillhouseBox("MILLHOUSE", text)
}

// SubHeader prints a styled sub-header
func (d *Display) SubHeader(text string) {
	fmt.Println()
	d.theme.MillhouseTitle.Println(text)
}

// Success prints a success message with checkmark
func (d *Display) Success(text string) {
	timestamp := time.Now().Format("15:04:05")
	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)
	d.theme.Success.Printf("%s ", SymbolCheck)
	fmt.Println(text)
}

// Error prints an error message with X
func (d *Display) Error(text string) {
	timestamp := time.Now().Format("15:04:05")
	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)
	d.theme.Error.Printf("%s ", SymbolCross)
	fmt.Println(text)
}

// Warning prints a warning message
func (d *Display) Warning(text string) {
	timestamp := time.Now().Format("15:04:05")
	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)
	d.theme.Warning.Printf("%s ", SymbolWarning)
	fmt.Println(text)
}

// Info prints an info message
func (d *Display) Info(text string) {
	timestamp := time.Now().Format("15:04:05")
	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)
	d.theme.Info.Printf("%s ", SymbolArrow)
	fmt.Println(text)
}

// Signal prints a detected signal with warning style
func (d *Display) Signal(signal, details string) {
	timestamp := time.Now().Format("15:04:05")
	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)
	d.theme.Warning.Printf("%s >>> %s", SymbolWarning, signal)
	if details != "" {
		fmt.Printf(": %s", details)
	}
	fmt.Println()
}

// TokenUsage prints token usage information
func (d *Display) TokenUsage(input, output, total int) {
	timestamp := time.Now().Format("15:04:05")
	percentage := float64(total) / 100000 * 100

	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)

	var statusColor = d.theme.Success
	if percentage > 90 {
		statusColor = d.theme.Error
	} else if percentage > 70 {
		statusColor = d.theme.Warning
	}

	statusColor.Printf("%s Tokens: %d (%.1f%% of 100K)\n", SymbolCheck, total, percentage)
}

// PRDStatus prints PRD status with color coding
func (d *Display) PRDStatus(p prd.PRD) {
	var status string
	var statusColor = d.theme.Error

	if p.Passes.IsTrue() {
		status = "complete"
		statusColor = d.theme.Success
	} else if p.Passes.IsPending() {
		status = "pending"
		statusColor = d.theme.Warning
	} else {
		status = "open"
		statusColor = d.theme.Error
	}

	statusColor.Printf("  [%s]", status)
	fmt.Printf(" P%d ", p.Priority)
	d.theme.Bold.Print(p.ID)
	fmt.Printf(": %s\n", p.Description)

	if p.Notes != "" {
		notes := Truncate(p.Notes, 60)
		d.theme.Dim.Printf("       %s\n", notes)
	}
}

// Summary prints a summary line
func (d *Display) Summary(open, pending, complete int) {
	total := open + pending + complete
	d.theme.Bold.Printf("\nTotal: %d ", total)
	fmt.Print("(")
	d.theme.Error.Printf("%d open", open)
	fmt.Print(", ")
	d.theme.Warning.Printf("%d pending", pending)
	fmt.Print(", ")
	d.theme.Success.Printf("%d complete", complete)
	fmt.Println(")")
}

// Divider prints a horizontal divider
func (d *Display) Divider() {
	d.theme.Dim.Println(strings.Repeat(BoxHorizontal, 50))
}

// AgentHeader prints a header for agent execution
func (d *Display) AgentHeader(agentType, prdID string) {
	timestamp := time.Now().Format("15:04:05")
	d.theme.ClaudeTimestamp.Printf("[%s] ", timestamp)
	d.theme.ClaudeGutter.Printf("%s ", GutterClaude)

	switch agentType {
	case "executor":
		d.theme.Info.Printf("[%s]", agentType)
	case "analyzer":
		d.theme.AnalyzerText.Printf("[%s]", agentType)
	default:
		fmt.Printf("[%s]", agentType)
	}

	fmt.Printf(" Working on: ")
	d.theme.Bold.Println(prdID)
}

// --- Text Utilities ---

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
}

// Truncate truncates text to max length with ellipsis
func Truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

// CleanText removes control characters and normalizes whitespace
func CleanText(text string) string {
	// Replace multiple whitespace with single space
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

// --- Backwards Compatibility Layer ---

var defaultDisplay *Display

func init() {
	defaultDisplay = New()
}

// Package-level functions that delegate to the default Display instance

// Header prints a styled header
func Header(text string) {
	defaultDisplay.Header(text)
}

// SubHeader prints a styled sub-header
func SubHeader(text string) {
	defaultDisplay.SubHeader(text)
}

// Success prints a success message
func Success(text string) {
	defaultDisplay.Success(text)
}

// Error prints an error message
func Error(text string) {
	defaultDisplay.Error(text)
}

// Warning prints a warning message
func Warning(text string) {
	defaultDisplay.Warning(text)
}

// Info prints an info message
func Info(text string) {
	defaultDisplay.Info(text)
}

// PRDStatus prints PRD status with color coding
func PRDStatus(p prd.PRD) {
	defaultDisplay.PRDStatus(p)
}

// Summary prints a summary line
func Summary(open, pending, complete int) {
	defaultDisplay.Summary(open, pending, complete)
}

// Divider prints a horizontal divider
func Divider() {
	defaultDisplay.Divider()
}

// IterationHeader prints the header for an iteration
func IterationHeader(n, total int) {
	defaultDisplay.IterationHeader(n, total)
}

// AgentHeader prints a header for agent execution
func AgentHeader(agentType, prdID string) {
	defaultDisplay.AgentHeader(agentType, prdID)
}

// Signal prints a detected signal
func Signal(signal, details string) {
	defaultDisplay.Signal(signal, details)
}

// TokenUsage prints token usage information
func TokenUsage(input, output, total int) {
	defaultDisplay.TokenUsage(input, output, total)
}
