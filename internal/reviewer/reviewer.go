package reviewer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/daydemir/milhouse/internal/config"
	"github.com/daydemir/milhouse/internal/display"
	"github.com/daydemir/milhouse/internal/llm"
	"github.com/daydemir/milhouse/internal/prd"
	"github.com/daydemir/milhouse/internal/prompts"
)

// ReviewerResult contains the result of a reviewer run
type ReviewerResult struct {
	Verified      []string // PRD IDs that were verified (promoted to true)
	Rejected      []string // PRD IDs that were rejected (reverted to false)
	LoopRisk      []string // PRD IDs at risk of looping
	PlanUpdated   []string // PRD IDs whose plans were updated (bailout handling)
	PromptUpdated []string // Phase names whose prompts were updated
	Error         error
}

// Run executes the reviewer agent
func Run(ctx context.Context, basePath string, prdFile *prd.PRDFileData, iteration int, cfg *config.Config) (*ReviewerResult, error) {
	// Nil guard - use default config if none provided
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	result := &ReviewerResult{}

	prompt := buildReviewerPrompt(basePath, prdFile, iteration, cfg)

	display.AgentHeader("reviewer", "review")

	execResult, err := runClaude(ctx, basePath, prompt, cfg)
	if err != nil {
		result.Error = err
		return result, err
	}

	// Process signals from the reviewer output
	for _, signal := range execResult.GetSignals() {
		switch signal.Type {
		case llm.SignalVerified:
			result.Verified = append(result.Verified, signal.PRDID)
		case llm.SignalRejected:
			result.Rejected = append(result.Rejected, signal.PRDID)
		case llm.SignalLoopRisk:
			result.LoopRisk = append(result.LoopRisk, signal.PRDID)
		case llm.SignalPlanUpdated:
			result.PlanUpdated = append(result.PlanUpdated, signal.PRDID)
		case llm.SignalPromptUpdated:
			result.PromptUpdated = append(result.PromptUpdated, signal.Details)
		}
	}

	return result, nil
}

// ShouldRunReviewer determines if the reviewer should run
// It should run if there are pending PRDs, active PRDs (for bailout handling), or open PRDs
func ShouldRunReviewer(prdFile *prd.PRDFileData) bool {
	// Always run if there are pending PRDs
	if len(prdFile.GetPendingPRDs()) > 0 {
		return true
	}

	// Also run if there are active PRDs (to handle bailouts)
	if len(prdFile.GetActivePRDs()) > 0 {
		return true
	}

	// Also run if there are open PRDs (to cross-pollinate observations)
	if len(prdFile.GetOpenPRDs()) > 0 {
		return true
	}

	return false
}

func runClaude(ctx context.Context, basePath, prompt string, cfg *config.Config) (*llm.ConsoleHandler, error) {
	phaseConfig := cfg.GetPhaseConfig("reviewer")

	claude := llm.NewClaude("")

	// Create a cancellable context for this execution
	execCtx, cancelExec := context.WithCancel(ctx)
	defer cancelExec()

	opts := llm.ExecuteOptions{
		Prompt:       prompt,
		Model:        phaseConfig.Model,
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
	handler := llm.NewConsoleHandlerWithTerminate(phaseConfig.MaxTokens, cancelExec)

	// Parse the stream
	llm.ParseStream(reader, handler, cancelExec)

	fmt.Println() // Ensure newline after output
	handler.DisplayFinalTokenUsage()

	return handler, nil
}

func buildReviewerPrompt(basePath string, prdFile *prd.PRDFileData, iteration int, cfg *config.Config) string {
	phaseConfig := cfg.GetPhaseConfig("reviewer")

	allPRDsJSON, _ := json.MarshalIndent(prdFile.PRDs, "", "  ")
	progressContent := readLastLines(prd.GetMillhousePath(basePath, prd.ProgressFile), phaseConfig.ProgressLines)

	// Collect active plans
	activePlans := make(map[string]string)
	for _, p := range prdFile.GetActivePRDs() {
		planPath := prd.GetPlanPath(basePath, p.ID)
		if content := readFileContent(planPath); content != "" {
			activePlans[p.ID] = content
		}
	}

	// Also include plans for pending PRDs (they still have plans until verified/rejected)
	for _, p := range prdFile.GetPendingPRDs() {
		planPath := prd.GetPlanPath(basePath, p.ID)
		if content := readFileContent(planPath); content != "" {
			activePlans[p.ID] = content
		}
	}

	reviewerAugmentation := prompts.LoadAugmentation(basePath, "reviewer")

	// Load prompt files for self-improvement capability
	plannerPrompt := prompts.LoadAugmentation(basePath, "planner")
	builderPrompt := prompts.LoadAugmentation(basePath, "builder")
	reviewerPrompt := prompts.LoadAugmentation(basePath, "reviewer")

	return prompts.BuildReviewerPrompt(prompts.ReviewerData{
		AllPRDsJSON:          string(allPRDsJSON),
		ActivePlans:          activePlans,
		ProgressContent:      progressContent,
		Iteration:            iteration,
		ReviewerAugmentation: reviewerAugmentation,
		ReviewerPromptMode:   phaseConfig.ReviewerPromptMode,
		PlannerPrompt:        plannerPrompt,
		BuilderPrompt:        builderPrompt,
		ReviewerPrompt:       reviewerPrompt,
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
