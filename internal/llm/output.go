package llm

import (
	"bufio"
	"encoding/json"
	"io"
	"regexp"
	"strings"

	"github.com/suelio/millhouse/internal/display"
)

// Signal types (Millhouse-specific)
const (
	SignalPRDComplete      = "PRD_COMPLETE"
	SignalBailout          = "BAILOUT"
	SignalBlocked          = "BLOCKED"
	SignalAnalysisComplete = "ANALYSIS_COMPLETE"
	SignalVerified         = "VERIFIED"
	SignalRejected         = "REJECTED"
	SignalLoopRisk         = "LOOP_RISK"
)

// Signal represents a detected signal from agent output
type Signal struct {
	Type    string
	Details string
	PRDID   string // For VERIFIED, REJECTED, LOOP_RISK
}

// TokenStats tracks token usage during execution
type TokenStats struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// OutputHandler handles parsed stream events
type OutputHandler interface {
	OnToolUse(name string)
	OnText(text string)
	OnDone(result string)
	OnSignal(signal Signal)
	OnTokenUsage(usage TokenStats)
	OnTokenUsageCumulative(usage TokenStats) // For message_delta cumulative counts
	GetSignals() []Signal
	GetTokenStats() TokenStats
	GetOutput() string
	ShouldTerminate() bool
}

// StreamEvent represents a single event from Claude's stream-json output
type StreamEvent struct {
	Type    string          `json:"type"`
	Message *MessageContent `json:"message,omitempty"`
	Result  string          `json:"result,omitempty"`
	Delta   *DeltaContent   `json:"delta,omitempty"`
	Usage   *UsageBlock     `json:"usage,omitempty"`
}

// MessageContent represents the message field in stream events
type MessageContent struct {
	Content []ContentBlock `json:"content,omitempty"`
	Usage   *UsageBlock    `json:"usage,omitempty"`
}

// ContentBlock represents a content block (text or tool_use)
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"` // for tool_use
}

// DeltaContent represents incremental content updates
type DeltaContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// UsageBlock represents token usage data from Claude's output
type UsageBlock struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ConsoleHandler implements OutputHandler for terminal output
type ConsoleHandler struct {
	signals        []Signal
	tokenStats     TokenStats
	tokenThreshold int
	output         strings.Builder
	onTerminate    func()
	shouldStop     bool
	display        *display.Display
	toolCount      int
	textBuffer     strings.Builder
}

// NewConsoleHandler creates a basic console handler
func NewConsoleHandler() *ConsoleHandler {
	return &ConsoleHandler{
		tokenThreshold: 100000, // 100K for Millhouse
		display:        display.New(),
	}
}

// NewConsoleHandlerWithThreshold creates a handler with custom token threshold
func NewConsoleHandlerWithThreshold(threshold int) *ConsoleHandler {
	return &ConsoleHandler{
		tokenThreshold: threshold,
		display:        display.New(),
	}
}

// NewConsoleHandlerWithTerminate creates a handler with token limit termination support
func NewConsoleHandlerWithTerminate(threshold int, onTerminate func()) *ConsoleHandler {
	return &ConsoleHandler{
		tokenThreshold: threshold,
		onTerminate:    onTerminate,
		display:        display.New(),
	}
}

// NewConsoleHandlerWithDisplay creates a handler with a custom display instance
func NewConsoleHandlerWithDisplay(d *display.Display, threshold int, onTerminate func()) *ConsoleHandler {
	return &ConsoleHandler{
		tokenThreshold: threshold,
		onTerminate:    onTerminate,
		display:        d,
	}
}

func (h *ConsoleHandler) OnToolUse(name string) {
	// Increment tool count for display
	h.toolCount++
}

func (h *ConsoleHandler) OnText(text string) {
	// Buffer text and print with styled output
	h.textBuffer.WriteString(text)
	h.output.WriteString(text)

	// Check for WORKING ON pattern and highlight
	if matches := workingOnPattern.FindStringSubmatch(text); matches != nil {
		h.display.ActivePRD(matches[1])
		return // Don't double-print
	}

	// Stream text with subdued styling
	h.display.ClaudeStreaming(text)
}

func (h *ConsoleHandler) OnDone(result string) {
	// Capture result text
	h.output.WriteString(result)
}

func (h *ConsoleHandler) OnSignal(signal Signal) {
	h.signals = append(h.signals, signal)

	// Terminal signals should stop execution
	if isTerminalSignal(signal) {
		h.shouldStop = true
	}
}

func (h *ConsoleHandler) OnTokenUsage(usage TokenStats) {
	h.tokenStats.InputTokens += usage.InputTokens
	h.tokenStats.OutputTokens += usage.OutputTokens
	h.tokenStats.TotalTokens = h.tokenStats.InputTokens + h.tokenStats.OutputTokens

	// Display token usage with styled output
	h.display.TokenUsage(h.tokenStats.InputTokens, h.tokenStats.OutputTokens, h.tokenStats.TotalTokens)

	// Reset tool count after displaying
	h.toolCount = 0

	// Check threshold and trigger termination if exceeded
	if h.tokenStats.TotalTokens >= h.tokenThreshold {
		h.shouldStop = true
		h.signals = append(h.signals, Signal{
			Type:    SignalBailout,
			Details: "token limit exceeded",
		})
		if h.onTerminate != nil {
			h.onTerminate()
		}
	}
}

func (h *ConsoleHandler) OnTokenUsageCumulative(usage TokenStats) {
	// Replace with cumulative value (message_delta provides cumulative counts)
	if usage.OutputTokens > 0 {
		h.tokenStats.OutputTokens = usage.OutputTokens
	}
	h.tokenStats.TotalTokens = h.tokenStats.InputTokens + h.tokenStats.OutputTokens

	// Display token usage with styled output
	h.display.TokenUsage(h.tokenStats.InputTokens, h.tokenStats.OutputTokens, h.tokenStats.TotalTokens)

	// Reset tool count after displaying
	h.toolCount = 0

	// Check threshold and trigger termination if exceeded
	if h.tokenStats.TotalTokens >= h.tokenThreshold {
		h.shouldStop = true
		h.signals = append(h.signals, Signal{
			Type:    SignalBailout,
			Details: "token limit exceeded",
		})
		if h.onTerminate != nil {
			h.onTerminate()
		}
	}
}

func (h *ConsoleHandler) GetSignals() []Signal {
	return h.signals
}

func (h *ConsoleHandler) GetTokenStats() TokenStats {
	return h.tokenStats
}

func (h *ConsoleHandler) GetOutput() string {
	return h.output.String()
}

func (h *ConsoleHandler) ShouldTerminate() bool {
	return h.shouldStop
}

// GetToolCount returns the current tool use count
func (h *ConsoleHandler) GetToolCount() int {
	return h.toolCount
}

// SetDisplay sets the display instance for styled output
func (h *ConsoleHandler) SetDisplay(d *display.Display) {
	h.display = d
}

// isTerminalSignal returns true if the signal indicates the agent should stop
func isTerminalSignal(s Signal) bool {
	return s.Type == SignalPRDComplete ||
		s.Type == SignalBailout ||
		s.Type == SignalBlocked ||
		s.Type == SignalAnalysisComplete
}

// Signal patterns (Millhouse-specific)
var (
	prdCompletePattern      = regexp.MustCompile(`###PRD_COMPLETE###`)
	bailoutPattern          = regexp.MustCompile(`###BAILOUT:(.+?)###`)
	blockedPattern          = regexp.MustCompile(`###BLOCKED:(.+?)###`)
	analysisCompletePattern = regexp.MustCompile(`###ANALYSIS_COMPLETE###`)
	verifiedPattern         = regexp.MustCompile(`###VERIFIED:(.+?)###`)
	rejectedPattern         = regexp.MustCompile(`###REJECTED:(.+?):(.+?)###`)
	loopRiskPattern         = regexp.MustCompile(`###LOOP_RISK:(.+?)###`)
	workingOnPattern        = regexp.MustCompile(`(?:\*\*)?WORKING ON:\s*([a-z0-9-]+)(?:\*\*)?`)
)

// ParseStream reads the Claude stream-json output and calls the handler
// onTerminate is called when a termination signal is detected
func ParseStream(reader io.Reader, handler OutputHandler, onTerminate func()) error {
	scanner := bufio.NewScanner(reader)
	// Increase buffer size for large JSON lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 16*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event StreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Skip malformed JSON lines
			continue
		}

		switch event.Type {
		case "message_start":
			// Handle initial input tokens from message_start
			if event.Message != nil && event.Message.Usage != nil {
				handler.OnTokenUsage(TokenStats{
					InputTokens:  event.Message.Usage.InputTokens,
					OutputTokens: event.Message.Usage.OutputTokens,
				})
			}

		case "content_block_delta":
			if event.Delta != nil && event.Delta.Type == "text_delta" {
				handler.OnText(event.Delta.Text)
				checkSignals(event.Delta.Text, handler)
			}

		case "message_delta":
			// message_delta provides cumulative output_tokens, not incremental
			if event.Usage != nil {
				handler.OnTokenUsageCumulative(TokenStats{
					OutputTokens: event.Usage.OutputTokens,
				})
			}

		case "assistant":
			if event.Message != nil {
				// Parse token usage
				if event.Message.Usage != nil {
					handler.OnTokenUsage(TokenStats{
						InputTokens:  event.Message.Usage.InputTokens,
						OutputTokens: event.Message.Usage.OutputTokens,
					})
				}

				for _, content := range event.Message.Content {
					switch content.Type {
					case "tool_use":
						handler.OnToolUse(content.Name)
					case "text":
						handler.OnText(content.Text)
						checkSignals(content.Text, handler)
					}
				}
			}

		case "result":
			// Check for signals in result text
			checkSignals(event.Result, handler)
			handler.OnDone(event.Result)
		}

		// Check if we should terminate
		if handler.ShouldTerminate() {
			if onTerminate != nil {
				onTerminate()
			}
			return nil
		}
	}

	return scanner.Err()
}

// checkSignals looks for Millhouse signal patterns in text
func checkSignals(text string, handler OutputHandler) {
	// Check for PRD_COMPLETE
	if prdCompletePattern.MatchString(text) {
		handler.OnSignal(Signal{Type: SignalPRDComplete})
	}

	// Check for BAILOUT
	if matches := bailoutPattern.FindStringSubmatch(text); matches != nil {
		handler.OnSignal(Signal{
			Type:    SignalBailout,
			Details: strings.TrimSpace(matches[1]),
		})
	}

	// Check for BLOCKED
	if matches := blockedPattern.FindStringSubmatch(text); matches != nil {
		handler.OnSignal(Signal{
			Type:    SignalBlocked,
			Details: strings.TrimSpace(matches[1]),
		})
	}

	// Check for ANALYSIS_COMPLETE
	if analysisCompletePattern.MatchString(text) {
		handler.OnSignal(Signal{Type: SignalAnalysisComplete})
	}

	// Check for VERIFIED
	if matches := verifiedPattern.FindAllStringSubmatch(text, -1); matches != nil {
		for _, match := range matches {
			handler.OnSignal(Signal{
				Type:  SignalVerified,
				PRDID: strings.TrimSpace(match[1]),
			})
		}
	}

	// Check for REJECTED
	if matches := rejectedPattern.FindAllStringSubmatch(text, -1); matches != nil {
		for _, match := range matches {
			handler.OnSignal(Signal{
				Type:    SignalRejected,
				PRDID:   strings.TrimSpace(match[1]),
				Details: strings.TrimSpace(match[2]),
			})
		}
	}

	// Check for LOOP_RISK
	if matches := loopRiskPattern.FindAllStringSubmatch(text, -1); matches != nil {
		for _, match := range matches {
			handler.OnSignal(Signal{
				Type:  SignalLoopRisk,
				PRDID: strings.TrimSpace(match[1]),
			})
		}
	}
}
