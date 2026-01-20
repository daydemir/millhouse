package cli

import (
	"github.com/daydemir/milhouse/internal/llm"
	"github.com/daydemir/milhouse/internal/prd"
)

// IterationState captures the state of the system at a point in time
type IterationState struct {
	OpenCount     int
	ActiveCount   int
	PendingCount  int
	CompleteCount int
	SignalTypes   []string // e.g., ["VERIFIED", "PLAN_UPDATED"]
}

// Equals compares two iteration states for equality
func (s *IterationState) Equals(other *IterationState) bool {
	if s == nil || other == nil {
		return false
	}

	if s.OpenCount != other.OpenCount ||
		s.ActiveCount != other.ActiveCount ||
		s.PendingCount != other.PendingCount ||
		s.CompleteCount != other.CompleteCount {
		return false
	}

	// Check if signal types are the same
	if len(s.SignalTypes) != len(other.SignalTypes) {
		return false
	}

	// Compare signal types (order-independent)
	signalMap := make(map[string]int)
	for _, sig := range s.SignalTypes {
		signalMap[sig]++
	}
	for _, sig := range other.SignalTypes {
		signalMap[sig]--
		if signalMap[sig] < 0 {
			return false
		}
	}

	return true
}

// IsIdle returns true if the iteration produced no productive signals
func (s *IterationState) IsIdle() bool {
	// Productive signals that indicate actual work being done
	productiveSignals := map[string]bool{
		llm.SignalVerified:     true,
		llm.SignalRejected:     true,
		llm.SignalPlanComplete: true,
		llm.SignalPlanUpdated:  true,
		llm.SignalPRDComplete:  true,
	}

	for _, sig := range s.SignalTypes {
		if productiveSignals[sig] {
			return false
		}
	}

	return len(s.SignalTypes) == 0 || allNonProductive(s.SignalTypes)
}

// allNonProductive checks if all signals are non-productive
func allNonProductive(signals []string) bool {
	nonProductiveSignals := map[string]bool{
		llm.SignalLoopRisk:         true,
		llm.SignalAnalysisComplete: true,
		llm.SignalBlocked:          true,
	}

	for _, sig := range signals {
		if !nonProductiveSignals[sig] {
			return false
		}
	}

	return true
}

// CaptureIterationState captures the current state of PRDs and signals
func CaptureIterationState(prdFile *prd.PRDFileData, signals []llm.Signal) *IterationState {
	state := &IterationState{
		SignalTypes: make([]string, 0),
	}

	// Count PRDs by state
	for _, p := range prdFile.PRDs {
		if p.Passes.IsTrue() {
			state.CompleteCount++
		} else if p.Passes.IsPending() {
			state.PendingCount++
		} else if p.Passes.IsActive() {
			state.ActiveCount++
		} else {
			state.OpenCount++
		}
	}

	// Collect signal types (deduplicated by type, not by PRD)
	signalSet := make(map[string]bool)
	for _, sig := range signals {
		signalSet[sig.Type] = true
	}
	for sigType := range signalSet {
		state.SignalTypes = append(state.SignalTypes, sigType)
	}

	return state
}
