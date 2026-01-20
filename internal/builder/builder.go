package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/daydemir/milhouse/internal/config"
	"github.com/daydemir/milhouse/internal/display"
	"github.com/daydemir/milhouse/internal/llm"
	"github.com/daydemir/milhouse/internal/prd"
	"github.com/daydemir/milhouse/internal/prompts"
)

// BuilderResult contains the result of a builder run
type BuilderResult struct {
	Signals     []llm.Signal
	TotalTokens int
	Output      string
	Error       error
}

// Run executes the builder agent to implement the active PRD's plan
func Run(ctx context.Context, basePath string, prdFile *prd.PRDFileData, cfg *config.Config) (*BuilderResult, error) {
	// Nil guard - use default config if none provided
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Get the active PRD
	activePRDs := prdFile.GetActivePRDs()
	if len(activePRDs) == 0 {
		return &BuilderResult{}, fmt.Errorf("no active PRD found")
	}

	activePRD := activePRDs[0]
	prompt := buildBuilderPrompt(basePath, &activePRD, cfg)

	display.AgentHeader("builder", "executing plan for "+activePRD.ID)

	return runClaude(ctx, basePath, prompt, cfg)
}

// RunChat runs an interactive Claude session
func RunChat(ctx context.Context, basePath string, prdFile *prd.PRDFileData, cfg *config.Config) error {
	prompt := buildChatPrompt(basePath, prdFile)

	// For chat, we run interactively (not stream-json)
	return runClaudeInteractive(ctx, basePath, prompt, cfg)
}

// ShouldRunBuilder determines if the builder should run
// It should run if there's an active PRD with a plan
func ShouldRunBuilder(prdFile *prd.PRDFileData) bool {
	return len(prdFile.GetActivePRDs()) > 0
}

func runClaude(ctx context.Context, basePath, prompt string, cfg *config.Config) (*BuilderResult, error) {
	result := &BuilderResult{}

	phaseConfig := cfg.GetPhaseConfig("builder")

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

	// Convert handler results to BuilderResult
	result.Output = handler.GetOutput()
	result.TotalTokens = handler.GetTokenStats().TotalTokens
	result.Signals = handler.GetSignals()

	fmt.Println() // Ensure newline after output
	handler.DisplayFinalTokenUsage()

	return result, nil
}

func runClaudeInteractive(ctx context.Context, basePath, prompt string, cfg *config.Config) error {
	phaseConfig := cfg.GetPhaseConfig("chat")

	claude := llm.NewClaude("")

	opts := llm.ExecuteOptions{
		SystemPrompt: prompt,
		Model:        phaseConfig.Model,
		ContextFiles: []string{
			prd.GetMillhousePath(basePath, prd.PRDFile),
			prd.GetMillhousePath(basePath, prd.ProgressFile),
			prd.GetMillhousePath(basePath, prd.PromptFile),
		},
		WorkDir: basePath,
	}

	return claude.ExecuteInteractive(ctx, opts)
}

func buildBuilderPrompt(basePath string, activePRD *prd.PRD, cfg *config.Config) string {
	phaseConfig := cfg.GetPhaseConfig("builder")

	promptMD := readFileContent(prd.GetMillhousePath(basePath, prd.PromptFile))
	activePRDJSON, _ := json.MarshalIndent(activePRD, "", "  ")
	progressContent := readLastLines(prd.GetMillhousePath(basePath, prd.ProgressFile), phaseConfig.ProgressLines)
	planContent := readFileContent(prd.GetPlanPath(basePath, activePRD.ID))
	builderAugmentation := prompts.LoadAugmentation(basePath, "builder")

	return prompts.BuildBuilderPrompt(prompts.BuilderData{
		PromptMD:            promptMD,
		ActivePRDJSON:       string(activePRDJSON),
		PlanContent:         planContent,
		ProgressContent:     progressContent,
		Timestamp:           time.Now().Format("2006-01-02 15:04"),
		BuilderAugmentation: builderAugmentation,
	})
}

func buildChatPrompt(basePath string, prdFile *prd.PRDFileData) string {
	open := prdFile.GetOpenPRDs()
	active := prdFile.GetActivePRDs()
	pending := prdFile.GetPendingPRDs()
	complete := prdFile.GetCompletePRDs()

	promptContent := readFileContent(prd.GetMillhousePath(basePath, prd.PromptFile))
	hasPromptContent := len(promptContent) > 200

	progressContent := readFileContent(prd.GetMillhousePath(basePath, prd.ProgressFile))
	progressLines := len(strings.Split(progressContent, "\n"))

	chatAugmentation := prompts.LoadAugmentation(basePath, "chat")

	return prompts.BuildChatPrompt(prompts.ChatData{
		TotalPRDs:        len(prdFile.PRDs),
		OpenPRDs:         len(open),
		ActivePRDs:       len(active),
		PendingPRDs:      len(pending),
		CompletePRDs:     len(complete),
		ProgressLines:    progressLines,
		HasPromptContent: hasPromptContent,
		ChatAugmentation: chatAugmentation,
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
