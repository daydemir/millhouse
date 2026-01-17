package analyzer

import (
	"context"

	"github.com/suelio/millhouse/internal/executor"
	"github.com/suelio/millhouse/internal/prd"
)

// AnalyzerResult contains the result of an analyzer run
type AnalyzerResult struct {
	Verified []string // PRD IDs that were verified (promoted to true)
	Rejected []string // PRD IDs that were rejected (reverted to false)
	LoopRisk []string // PRD IDs at risk of looping
	Error    error
}

// Run executes the analyzer agent
func Run(ctx context.Context, basePath string, prdFile *prd.PRDFileData, iteration int) (*AnalyzerResult, error) {
	result := &AnalyzerResult{}

	// Run the analyzer via Claude
	execResult, err := executor.RunAnalyzer(ctx, basePath, prdFile, iteration)
	if err != nil {
		result.Error = err
		return result, err
	}

	// Process signals from the analyzer output
	for _, signal := range execResult.Signals {
		switch signal.Type {
		case executor.SignalVerified:
			result.Verified = append(result.Verified, signal.PRDID)
		case executor.SignalRejected:
			result.Rejected = append(result.Rejected, signal.PRDID)
		case executor.SignalLoopRisk:
			result.LoopRisk = append(result.LoopRisk, signal.PRDID)
		}
	}

	return result, nil
}

// ShouldRunAnalyzer determines if the analyzer should run
// It should run if there are pending PRDs or if progress was made
func ShouldRunAnalyzer(prdFile *prd.PRDFileData) bool {
	// Always run if there are pending PRDs
	if len(prdFile.GetPendingPRDs()) > 0 {
		return true
	}

	// Also run if there are open PRDs (to cross-pollinate observations)
	if len(prdFile.GetOpenPRDs()) > 0 {
		return true
	}

	return false
}
