package prompts

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/daydemir/milhouse/internal/prd"
)

//go:embed *.tmpl
var templates embed.FS

var (
	sharedTmpl   *template.Template
	plannerTmpl  *template.Template
	builderTmpl  *template.Template
	reviewerTmpl *template.Template
	chatTmpl     *template.Template
)

func init() {
	// Parse shared components first
	sharedTmpl = template.Must(template.ParseFS(templates, "shared.tmpl"))

	// Parse each agent template with shared components
	plannerTmpl = template.Must(template.Must(sharedTmpl.Clone()).ParseFS(templates, "planner.tmpl"))
	builderTmpl = template.Must(template.Must(sharedTmpl.Clone()).ParseFS(templates, "builder.tmpl"))
	reviewerTmpl = template.Must(template.Must(sharedTmpl.Clone()).ParseFS(templates, "reviewer.tmpl"))
	chatTmpl = template.Must(template.ParseFS(templates, "chat.tmpl"))
}

// PlannerData contains data for the planner prompt template
type PlannerData struct {
	PromptMD            string // Codebase patterns from prompt.md
	OpenPRDsJSON        string // JSON of open PRDs (passes=false)
	ProgressContent     string // Last lines of progress.md
	Timestamp           string // Current timestamp
	PlannerAugmentation string // Optional project-specific planner guidance
}

// BuildPlannerPrompt renders the planner prompt template
func BuildPlannerPrompt(data PlannerData) string {
	var buf bytes.Buffer
	if err := plannerTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// BuilderData contains data for the builder prompt template
type BuilderData struct {
	PromptMD            string // Codebase patterns from prompt.md
	ActivePRDJSON       string // JSON of the active PRD being worked on
	PlanContent         string // Content of the plan file
	ProgressContent     string // Last lines of progress.md
	Timestamp           string // Current timestamp
	BuilderAugmentation string // Optional project-specific builder guidance
}

// BuildBuilderPrompt renders the builder prompt template
func BuildBuilderPrompt(data BuilderData) string {
	var buf bytes.Buffer
	if err := builderTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// ReviewerData contains data for the reviewer prompt template
type ReviewerData struct {
	AllPRDsJSON          string            // JSON of ALL PRDs
	ActivePlans          map[string]string // Map of PRD ID to plan content
	ProgressContent      string            // Last lines of progress.md
	Iteration            int               // Current iteration count
	ReviewerAugmentation string            // Optional project-specific reviewer guidance
	// Prompt improvement fields
	ReviewerPromptMode   string            // "standard", "enhanced", "aggressive"
	PlannerPrompt        string            // Content of .milhouse/prompts/planner.md
	BuilderPrompt        string            // Content of .milhouse/prompts/builder.md
	ReviewerPrompt       string            // Content of .milhouse/prompts/reviewer.md
}

// BuildReviewerPrompt renders the reviewer prompt template
func BuildReviewerPrompt(data ReviewerData) string {
	var buf bytes.Buffer
	if err := reviewerTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// ChatData contains data for the chat prompt template
type ChatData struct {
	TotalPRDs        int
	OpenPRDs         int
	ActivePRDs       int
	PendingPRDs      int
	CompletePRDs     int
	ProgressLines    int
	HasPromptContent bool
	ChatAugmentation string // Optional project-specific chat guidance
}

// BuildChatPrompt renders the chat prompt template
func BuildChatPrompt(data ChatData) string {
	var buf bytes.Buffer
	if err := chatTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// LoadAugmentation reads a phase-specific augmentation file
// Returns empty string if file doesn't exist (augmentations are optional)
func LoadAugmentation(basePath, phase string) string {
	augPath := GetAugmentationPath(basePath, phase)
	content, err := os.ReadFile(augPath)
	if err != nil {
		// File doesn't exist or can't be read - this is OK, augmentations are optional
		return ""
	}

	// Trim whitespace to detect empty files
	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return ""
	}

	return trimmed
}

// GetAugmentationPath returns the path to a phase augmentation file
func GetAugmentationPath(basePath, phase string) string {
	return filepath.Join(basePath, prd.MillhouseDir, prd.PromptsDir, phase+".md")
}

// EnsurePromptsDir creates the prompts directory if it doesn't exist
func EnsurePromptsDir(basePath string) error {
	promptsPath := filepath.Join(basePath, prd.MillhouseDir, prd.PromptsDir)
	return os.MkdirAll(promptsPath, 0755)
}
