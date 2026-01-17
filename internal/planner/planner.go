package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/suelio/millhouse/internal/display"
	"github.com/suelio/millhouse/internal/llm"
	"github.com/suelio/millhouse/internal/prd"
	"github.com/suelio/millhouse/internal/prompts"
)

// TokenThreshold is the max tokens before forced bailout
const TokenThreshold = 80000

// PlannerResult contains the result of a planner run
type PlannerResult struct {
	PRDID       string       // PRD ID that was selected and planned
	PlanPath    string       // Path to the created plan file
	Signals     []llm.Signal // All signals from the planner
	TotalTokens int
	Output      string
	Skipped     bool   // True if planner skipped (no open PRDs or active exists)
	SkipReason  string // Reason for skipping
	Error       error
}

// Run executes the planner agent to select a PRD and create a plan
func Run(ctx context.Context, basePath string, prdFile *prd.PRDFileData) (*PlannerResult, error) {
	result := &PlannerResult{}

	// Check if we should run
	if !ShouldRunPlanner(prdFile) {
		result.Skipped = true
		if len(prdFile.GetActivePRDs()) > 0 {
			result.SkipReason = "active PRD exists"
		} else {
			result.SkipReason = "no open PRDs"
		}
		return result, nil
	}

	// Ensure plans directory exists
	if err := prd.EnsurePlansDir(basePath); err != nil {
		return nil, fmt.Errorf("failed to create plans directory: %w", err)
	}

	prompt := buildPlannerPrompt(basePath, prdFile)

	display.AgentHeader("planner", "selecting PRD and creating plan")

	execResult, err := runClaude(ctx, basePath, prompt)
	if err != nil {
		result.Error = err
		return result, err
	}

	result.Output = execResult.Output
	result.TotalTokens = execResult.TotalTokens
	result.Signals = execResult.Signals

	// Process signals to extract PRD ID
	for _, signal := range execResult.Signals {
		switch signal.Type {
		case llm.SignalPlanComplete:
			result.PRDID = signal.PRDID
			result.PlanPath = prd.GetPlanPath(basePath, signal.PRDID)
		case llm.SignalPlanSkipped:
			result.Skipped = true
			result.SkipReason = signal.Details
		}
	}

	return result, nil
}

// ShouldRunPlanner determines if the planner should run
// Planner should run only if there are open PRDs AND no active PRDs
func ShouldRunPlanner(prdFile *prd.PRDFileData) bool {
	// Skip if there's already an active PRD
	if len(prdFile.GetActivePRDs()) > 0 {
		return false
	}

	// Skip if there are no open PRDs to plan
	if len(prdFile.GetOpenPRDs()) == 0 {
		return false
	}

	return true
}

func runClaude(ctx context.Context, basePath, prompt string) (*PlannerResult, error) {
	result := &PlannerResult{}

	claude := llm.NewClaude("")

	// Create a cancellable context for this execution
	execCtx, cancelExec := context.WithCancel(ctx)
	defer cancelExec()

	opts := llm.ExecuteOptions{
		Prompt:       prompt,
		Model:        "sonnet",
		AllowedTools: []string{
			"Read", "Write", "Edit", "Bash", "Glob", "Grep",
			"Task", "TodoWrite", "WebSearch", "WebFetch",
		},
		ContextFiles: []string{
			prd.GetMillhousePath(basePath, prd.PRDFile),
			prd.GetMillhousePath(basePath, prd.ProgressFile),
			prd.GetMillhousePath(basePath, prd.PromptFile),
		},
		WorkDir: basePath,
	}

	reader, err := claude.Execute(execCtx, opts)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Create handler with termination support
	handler := llm.NewConsoleHandlerWithTerminate(TokenThreshold, cancelExec)

	// Parse the stream
	llm.ParseStream(reader, handler, cancelExec)

	// Convert handler results to PlannerResult
	result.Output = handler.GetOutput()
	result.TotalTokens = handler.GetTokenStats().TotalTokens
	result.Signals = handler.GetSignals()

	fmt.Println() // Ensure newline after output
	if result.TotalTokens > 0 {
		display.TokenUsage(0, 0, result.TotalTokens)
	}

	return result, nil
}

func buildPlannerPrompt(basePath string, prdFile *prd.PRDFileData) string {
	promptMD := readFileContent(prd.GetMillhousePath(basePath, prd.PromptFile))
	openPRDs := prdFile.GetOpenPRDs()
	openPRDsJSON, _ := json.MarshalIndent(openPRDs, "", "  ")
	progressContent := readLastLines(prd.GetMillhousePath(basePath, prd.ProgressFile), 20)

	return prompts.BuildPlannerPrompt(prompts.PlannerData{
		PromptMD:        promptMD,
		OpenPRDsJSON:    string(openPRDsJSON),
		ProgressContent: progressContent,
		Timestamp:       time.Now().Format("2006-01-02 15:04"),
	})
}

func readFileContent(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func readLastLines(path string, n int) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) <= n {
		return string(content)
	}

	return strings.Join(lines[len(lines)-n:], "\n")
}
