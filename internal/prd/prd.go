package prd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	MillhouseDir = ".milhouse"
	PRDFile      = "prd.json"
	ProgressFile = "progress.md"
	PromptFile   = "prompt.md"
	EvidenceDir  = "evidence"
	PlansDir     = "plans"
)

// PassesStatus represents the quad-state passes field
// false = open, not attempted or needs work
// "active" = planner selected, has plan, builder working on it
// "pending" = builder claims complete, awaiting reviewer
// true = reviewer confirmed complete
type PassesStatus struct {
	Value interface{} // bool or string "active"/"pending"
}

func (p *PassesStatus) IsFalse() bool {
	if b, ok := p.Value.(bool); ok {
		return !b
	}
	return false
}

func (p *PassesStatus) IsActive() bool {
	if s, ok := p.Value.(string); ok {
		return s == "active"
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

func (p *PassesStatus) SetActive() {
	p.Value = "active"
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
		if s == "active" || s == "pending" {
			p.Value = s
			return nil
		}
		return fmt.Errorf("invalid passes string value: %s", s)
	}

	return fmt.Errorf("passes must be bool or 'active'/'pending'")
}

// PRD represents a single Product Requirements Document
type PRD struct {
	ID                 string       `json:"id"`
	Description        string       `json:"description"`
	AcceptanceCriteria []string     `json:"acceptanceCriteria"`
	Priority           int          `json:"priority"`
	Passes             PassesStatus `json:"passes"`
	Notes              string       `json:"notes"`
	ActivePlan         string       `json:"activePlan,omitempty"` // Path to plan file when active
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

// MillhouseExists checks if .milhouse directory exists
func MillhouseExists(basePath string) bool {
	path := filepath.Join(basePath, MillhouseDir)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// GetMillhousePath returns the full path to a.milhouse file
func GetMillhousePath(basePath, filename string) string {
	return filepath.Join(basePath, MillhouseDir, filename)
}

// GetEvidencePath returns the path to an evidence file for a PRD
func GetEvidencePath(basePath, prdID string) string {
	return filepath.Join(basePath, MillhouseDir, EvidenceDir, prdID+"-evidence.md")
}

// GetActivePRDs returns PRDs where passes="active"
func (p *PRDFileData) GetActivePRDs() []PRD {
	var active []PRD
	for _, prd := range p.PRDs {
		if prd.Passes.IsActive() {
			active = append(active, prd)
		}
	}
	return active
}

// GetPlanPath returns the path to a plan file for a PRD
func GetPlanPath(basePath, prdID string) string {
	return filepath.Join(basePath, MillhouseDir, PlansDir, prdID+"-plan.md")
}

// EnsurePlansDir creates the plans directory if it doesn't exist
func EnsurePlansDir(basePath string) error {
	plansPath := filepath.Join(basePath, MillhouseDir, PlansDir)
	return os.MkdirAll(plansPath, 0755)
}

// DeletePlan removes a plan file for a PRD
func DeletePlan(basePath, prdID string) error {
	planPath := GetPlanPath(basePath, prdID)
	if err := os.Remove(planPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete plan file: %w", err)
	}
	return nil
}

// PlanExists checks if a plan file exists for a PRD
func PlanExists(basePath, prdID string) bool {
	planPath := GetPlanPath(basePath, prdID)
	_, err := os.Stat(planPath)
	return err == nil
}
