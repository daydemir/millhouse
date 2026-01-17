package executor

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
const TokenThreshold = 100000

// ExecutorResult contains the result of an executor run
type ExecutorResult struct {
	Signals     []Signal
	TotalTokens int
	Output      string
	Error       error
}

// RunExecutor runs the executor agent to work on PRDs
func RunExecutor(ctx context.Context, basePath string, prdFile *prd.PRDFileData) (*ExecutorResult, error) {
	prompt := buildExecutorPrompt(basePath, prdFile)

	display.AgentHeader("executor", "selecting PRD")

	return runClaude(ctx, basePath, prompt)
}

// RunAnalyzer runs the analyzer agent
func RunAnalyzer(ctx context.Context, basePath string, prdFile *prd.PRDFileData, iteration int) (*ExecutorResult, error) {
	prompt := buildAnalyzerPrompt(basePath, prdFile, iteration)

	display.AgentHeader("analyzer", "review")

	return runClaude(ctx, basePath, prompt)
}

// RunDiscuss runs an interactive Claude session
func RunDiscuss(ctx context.Context, basePath string, prdFile *prd.PRDFileData) error {
	prompt := buildDiscussPrompt(basePath, prdFile)

	// For discuss, we run interactively (not stream-json)
	return runClaudeInteractive(ctx, basePath, prompt)
}

func runClaude(ctx context.Context, basePath, prompt string) (*ExecutorResult, error) {
	result := &ExecutorResult{}

	claude := llm.NewClaude("")

	// Create a cancellable context for this execution
	execCtx, cancelExec := context.WithCancel(ctx)
	defer cancelExec()

	opts := llm.ExecuteOptions{
		Prompt:       prompt,
		Model:        "sonnet",
		AllowedTools: []string{"Read", "Write", "Edit", "Bash", "Glob", "Grep"},
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

	// Convert handler results to ExecutorResult
	result.Output = handler.GetOutput()
	result.TotalTokens = handler.GetTokenStats().TotalTokens

	// Convert llm.Signal to executor.Signal
	for _, sig := range handler.GetSignals() {
		result.Signals = append(result.Signals, Signal{
			Type:    sig.Type,
			Details: sig.Details,
			PRDID:   sig.PRDID,
		})
	}

	fmt.Println() // Ensure newline after output
	if result.TotalTokens > 0 {
		display.TokenUsage(0, 0, result.TotalTokens)
	}

	return result, nil
}

func runClaudeInteractive(ctx context.Context, basePath, prompt string) error {
	claude := llm.NewClaude("")

	opts := llm.ExecuteOptions{
		SystemPrompt: prompt,
		ContextFiles: []string{
			prd.GetMillhousePath(basePath, prd.PRDFile),
			prd.GetMillhousePath(basePath, prd.ProgressFile),
			prd.GetMillhousePath(basePath, prd.PromptFile),
		},
		WorkDir: basePath,
	}

	return claude.ExecuteInteractive(ctx, opts)
}

func buildExecutorPrompt(basePath string, prdFile *prd.PRDFileData) string {
	promptMD := readFileContent(prd.GetMillhousePath(basePath, prd.PromptFile))
	openPRDs := prdFile.GetOpenPRDs()
	openPRDsJSON, _ := json.MarshalIndent(openPRDs, "", "  ")
	progressContent := readLastLines(prd.GetMillhousePath(basePath, prd.ProgressFile), 20)

	return prompts.BuildExecutorPrompt(prompts.ExecutorData{
		PromptMD:        promptMD,
		OpenPRDsJSON:    string(openPRDsJSON),
		ProgressContent: progressContent,
		Timestamp:       time.Now().Format("2006-01-02 15:04"),
	})
}

func buildAnalyzerPrompt(basePath string, prdFile *prd.PRDFileData, iteration int) string {
	allPRDsJSON, _ := json.MarshalIndent(prdFile.PRDs, "", "  ")
	progressContent := readLastLines(prd.GetMillhousePath(basePath, prd.ProgressFile), 50)

	return prompts.BuildAnalyzerPrompt(prompts.AnalyzerData{
		AllPRDsJSON:     string(allPRDsJSON),
		ProgressContent: progressContent,
		Iteration:       iteration,
	})
}

func buildDiscussPrompt(basePath string, prdFile *prd.PRDFileData) string {
	open := prdFile.GetOpenPRDs()
	pending := prdFile.GetPendingPRDs()
	complete := prdFile.GetCompletePRDs()

	promptContent := readFileContent(prd.GetMillhousePath(basePath, prd.PromptFile))
	hasPromptContent := len(promptContent) > 200

	progressContent := readFileContent(prd.GetMillhousePath(basePath, prd.ProgressFile))
	progressLines := len(strings.Split(progressContent, "\n"))

	return prompts.BuildDiscussPrompt(prompts.DiscussData{
		TotalPRDs:        len(prdFile.PRDs),
		OpenPRDs:         len(open),
		PendingPRDs:      len(pending),
		CompletePRDs:     len(complete),
		ProgressLines:    progressLines,
		HasPromptContent: hasPromptContent,
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

func formatCriteria(criteria []string) string {
	var result strings.Builder
	for _, c := range criteria {
		result.WriteString("- ")
		result.WriteString(c)
		result.WriteString("\n")
	}
	return result.String()
}
