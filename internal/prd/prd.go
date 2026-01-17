package prd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	MillhouseDir = ".millhouse"
	PRDFile      = "prd.json"
	ProgressFile = "progress.md"
	PromptFile   = "prompt.md"
	EvidenceDir  = "evidence"
)

// PassesStatus represents the tri-state passes field
// false = not attempted or needs work
// "pending" = agent claims complete, awaiting analyzer
// true = analyzer confirmed complete
type PassesStatus struct {
	Value interface{} // bool or string "pending"
}

func (p *PassesStatus) IsFalse() bool {
	if b, ok := p.Value.(bool); ok {
		return !b
	}
	return false
}

func (p *PassesStatus) IsPending() bool {
	if s, ok := p.Value.(string); ok {
		return s == "pending"
	}
	return false
}

func (p *PassesStatus) IsTrue() bool {
	if b, ok := p.Value.(bool); ok {
		return b
	}
	return false
}

func (p *PassesStatus) SetFalse() {
	p.Value = false
}

func (p *PassesStatus) SetPending() {
	p.Value = "pending"
}

func (p *PassesStatus) SetTrue() {
	p.Value = true
}

func (p PassesStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Value)
}

func (p *PassesStatus) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		p.Value = b
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "pending" {
			p.Value = s
			return nil
		}
		return fmt.Errorf("invalid passes string value: %s", s)
	}

	return fmt.Errorf("passes must be bool or 'pending'")
}

// PRD represents a single Product Requirements Document
type PRD struct {
	ID                 string       `json:"id"`
	Description        string       `json:"description"`
	AcceptanceCriteria []string     `json:"acceptanceCriteria"`
	Priority           int          `json:"priority"`
	Passes             PassesStatus `json:"passes"`
	Notes              string       `json:"notes"`
}

// PRDFile represents the prd.json file structure
type PRDFileData struct {
	PRDs []PRD `json:"prds"`
}

// Load reads and parses the prd.json file
func Load(basePath string) (*PRDFileData, error) {
	path := filepath.Join(basePath, MillhouseDir, PRDFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read prd.json: %w", err)
	}

	var prdFile PRDFileData
	if err := json.Unmarshal(data, &prdFile); err != nil {
		return nil, fmt.Errorf("failed to parse prd.json: %w", err)
	}

	return &prdFile, nil
}

// Save writes the prd.json file
func Save(basePath string, prdFile *PRDFileData) error {
	path := filepath.Join(basePath, MillhouseDir, PRDFile)
	data, err := json.MarshalIndent(prdFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal prd.json: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write prd.json: %w", err)
	}

	return nil
}

// GetOpenPRDs returns PRDs where passes=false
func (p *PRDFileData) GetOpenPRDs() []PRD {
	var open []PRD
	for _, prd := range p.PRDs {
		if prd.Passes.IsFalse() {
			open = append(open, prd)
		}
	}
	return open
}

// GetPendingPRDs returns PRDs where passes="pending"
func (p *PRDFileData) GetPendingPRDs() []PRD {
	var pending []PRD
	for _, prd := range p.PRDs {
		if prd.Passes.IsPending() {
			pending = append(pending, prd)
		}
	}
	return pending
}

// GetCompletePRDs returns PRDs where passes=true
func (p *PRDFileData) GetCompletePRDs() []PRD {
	var complete []PRD
	for _, prd := range p.PRDs {
		if prd.Passes.IsTrue() {
			complete = append(complete, prd)
		}
	}
	return complete
}

// FindByID finds a PRD by its ID
func (p *PRDFileData) FindByID(id string) *PRD {
	for i := range p.PRDs {
		if p.PRDs[i].ID == id {
			return &p.PRDs[i]
		}
	}
	return nil
}

// MillhouseExists checks if .millhouse directory exists
func MillhouseExists(basePath string) bool {
	path := filepath.Join(basePath, MillhouseDir)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// GetMillhousePath returns the full path to a millhouse file
func GetMillhousePath(basePath, filename string) string {
	return filepath.Join(basePath, MillhouseDir, filename)
}

// GetEvidencePath returns the path to an evidence file for a PRD
func GetEvidencePath(basePath, prdID string) string {
	return filepath.Join(basePath, MillhouseDir, EvidenceDir, prdID+"-evidence.md")
}
