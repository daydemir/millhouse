package prompts

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed *.tmpl
var templates embed.FS

var (
	executorTmpl *template.Template
	analyzerTmpl *template.Template
	discussTmpl  *template.Template
)

func init() {
	executorTmpl = template.Must(template.ParseFS(templates, "executor.tmpl"))
	analyzerTmpl = template.Must(template.ParseFS(templates, "analyzer.tmpl"))
	discussTmpl = template.Must(template.ParseFS(templates, "discuss.tmpl"))
}

// ExecutorData contains data for the executor prompt template
type ExecutorData struct {
	PromptMD        string
	OpenPRDsJSON    string
	ProgressContent string
	Timestamp       string
}

// BuildExecutorPrompt renders the executor prompt template
func BuildExecutorPrompt(data ExecutorData) string {
	var buf bytes.Buffer
	if err := executorTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// AnalyzerData contains data for the analyzer prompt template
type AnalyzerData struct {
	AllPRDsJSON     string
	ProgressContent string
	Iteration       int
}

// BuildAnalyzerPrompt renders the analyzer prompt template
func BuildAnalyzerPrompt(data AnalyzerData) string {
	var buf bytes.Buffer
	if err := analyzerTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// DiscussData contains data for the discuss prompt template
type DiscussData struct {
	TotalPRDs        int
	OpenPRDs         int
	PendingPRDs      int
	CompletePRDs     int
	ProgressLines    int
	HasPromptContent bool
}

// BuildDiscussPrompt renders the discuss prompt template
func BuildDiscussPrompt(data DiscussData) string {
	var buf bytes.Buffer
	if err := discussTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}
