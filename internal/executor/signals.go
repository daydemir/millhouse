package executor

// Signal types
const (
	SignalPRDComplete      = "PRD_COMPLETE"
	SignalBailout          = "BAILOUT"
	SignalBlocked          = "BLOCKED"
	SignalAnalysisComplete = "ANALYSIS_COMPLETE"
	SignalVerified         = "VERIFIED"
	SignalRejected         = "REJECTED"
	SignalLoopRisk         = "LOOP_RISK"
)

// Signal represents a detected signal from agent output
type Signal struct {
	Type    string
	Details string
	PRDID   string // For VERIFIED, REJECTED, LOOP_RISK
}

// HasSignal checks if a specific signal type exists in the list
func HasSignal(signals []Signal, signalType string) bool {
	for _, s := range signals {
		if s.Type == signalType {
			return true
		}
	}
	return false
}

// GetSignal returns the first signal of the given type
func GetSignal(signals []Signal, signalType string) *Signal {
	for _, s := range signals {
		if s.Type == signalType {
			return &s
		}
	}
	return nil
}

// IsTerminalSignal returns true if the signal indicates the agent should stop
func IsTerminalSignal(s Signal) bool {
	return s.Type == SignalPRDComplete ||
		s.Type == SignalBailout ||
		s.Type == SignalBlocked ||
		s.Type == SignalAnalysisComplete
}
