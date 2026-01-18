package prompts

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed *.tmpl
var templates embed.FS

var (
	plannerTmpl  *template.Template
	builderTmpl  *template.Template
	reviewerTmpl *template.Template
	chatTmpl     *template.Template
)

func init() {
	plannerTmpl = template.Must(template.ParseFS(templates, "planner.tmpl"))
	builderTmpl = template.Must(template.ParseFS(templates, "builder.tmpl"))
	reviewerTmpl = template.Must(template.ParseFS(templates, "reviewer.tmpl"))
	chatTmpl = template.Must(template.ParseFS(templates, "chat.tmpl"))
}

// PlannerData contains data for the planner prompt template
type PlannerData struct {
	PromptMD        string // Codebase patterns from prompt.md
	OpenPRDsJSON    string // JSON of open PRDs (passes=false)
	ProgressContent string // Last lines of progress.md
	Timestamp       string // Current timestamp
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
	PromptMD        string // Codebase patterns from prompt.md
	ActivePRDJSON   string // JSON of the active PRD being worked on
	PlanContent     string // Content of the plan file
	ProgressContent string // Last lines of progress.md
	Timestamp       string // Current timestamp
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
	AllPRDsJSON     string            // JSON of ALL PRDs
	ActivePlans     map[string]string // Map of PRD ID to plan content
	ProgressContent string            // Last lines of progress.md
	Iteration       int               // Current iteration count
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
	PendingPRDs      int
	CompletePRDs     int
	ProgressLines    int
	HasPromptContent bool
}

// BuildChatPrompt renders the chat prompt template
func BuildChatPrompt(data ChatData) string {
	var buf bytes.Buffer
	if err := chatTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}
