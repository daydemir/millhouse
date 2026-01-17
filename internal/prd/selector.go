package prd

import (
	"sort"
)

// SelectNext picks the best PRD to work on next
// Returns the PRD with the lowest priority where passes=false
// Returns nil if no open PRDs are available
func SelectNext(prdFile *PRDFileData) *PRD {
	open := prdFile.GetOpenPRDs()
	if len(open) == 0 {
		return nil
	}

	// Sort by priority (lowest first)
	sort.Slice(open, func(i, j int) bool {
		return open[i].Priority < open[j].Priority
	})

	// Return a pointer to the actual PRD in the slice, not the copy
	for i := range prdFile.PRDs {
		if prdFile.PRDs[i].ID == open[0].ID {
			return &prdFile.PRDs[i]
		}
	}

	return nil
}

// SelectNextPending picks a pending PRD for the analyzer to verify
// Returns nil if no pending PRDs are available
func SelectNextPending(prdFile *PRDFileData) *PRD {
	pending := prdFile.GetPendingPRDs()
	if len(pending) == 0 {
		return nil
	}

	// Return a pointer to the actual PRD in the slice
	for i := range prdFile.PRDs {
		if prdFile.PRDs[i].ID == pending[0].ID {
			return &prdFile.PRDs[i]
		}
	}

	return nil
}
