package llm

import "testing"

func TestOnTokenUsage_InputTokensAccumulated(t *testing.T) {
	handler := NewConsoleHandler()

	// Simulate first turn with 35000 input tokens
	handler.OnTokenUsage(TokenStats{InputTokens: 35000, OutputTokens: 100})

	// Simulate second turn with 40000 input tokens (should accumulate)
	handler.OnTokenUsage(TokenStats{InputTokens: 40000, OutputTokens: 200})

	stats := handler.GetTokenStats()

	if stats.InputTokens != 75000 {
		t.Errorf("InputTokens should be 75000 (accumulated), got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 300 {
		t.Errorf("OutputTokens should be 300 (accumulated), got %d", stats.OutputTokens)
	}
}

func TestOnTokenUsage_AccumulatesAllTokens(t *testing.T) {
	handler := NewConsoleHandler()

	handler.OnTokenUsage(TokenStats{InputTokens: 30000, OutputTokens: 0})
	handler.OnTokenUsage(TokenStats{InputTokens: 35000, OutputTokens: 100})
	handler.OnTokenUsage(TokenStats{InputTokens: 33000, OutputTokens: 50})

	stats := handler.GetTokenStats()

	if stats.InputTokens != 98000 {
		t.Errorf("InputTokens should be 98000 (accumulated), got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 150 {
		t.Errorf("OutputTokens should be 150 (accumulated), got %d", stats.OutputTokens)
	}
}

func TestThresholdTriggersTermination(t *testing.T) {
	terminated := false
	handler := NewConsoleHandlerWithTerminate(1000, func() {
		terminated = true
	})

	handler.OnTokenUsage(TokenStats{InputTokens: 800, OutputTokens: 300})

	if !terminated {
		t.Error("Expected termination callback to be called")
	}
	if !handler.ShouldTerminate() {
		t.Error("Expected ShouldTerminate() to return true")
	}
}

func TestOnTokenUsage_CacheReadTokensTracked(t *testing.T) {
	handler := NewConsoleHandler()

	handler.OnTokenUsage(TokenStats{
		InputTokens:     30000,
		OutputTokens:    100,
		CacheReadTokens: 5000,
	})
	handler.OnTokenUsage(TokenStats{
		InputTokens:     35000,
		OutputTokens:    200,
		CacheReadTokens: 7000,
	})

	stats := handler.GetTokenStats()

	if stats.InputTokens != 65000 {
		t.Errorf("InputTokens should be 65000 (accumulated), got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 300 {
		t.Errorf("OutputTokens should be 300 (accumulated), got %d", stats.OutputTokens)
	}
	if stats.CacheReadTokens != 12000 {
		t.Errorf("CacheReadTokens should be 12000 (accumulated), got %d", stats.CacheReadTokens)
	}
	// Total should not include CacheReadTokens
	if stats.TotalTokens != 65300 {
		t.Errorf("TotalTokens should be 65300 (Input + Output), got %d", stats.TotalTokens)
	}
}

func TestThresholdNotTriggeredBelowLimit(t *testing.T) {
	terminated := false
	handler := NewConsoleHandlerWithTerminate(100000, func() {
		terminated = true
	})

	handler.OnTokenUsage(TokenStats{InputTokens: 50000, OutputTokens: 1000})

	if terminated {
		t.Error("Expected termination callback NOT to be called")
	}
	if handler.ShouldTerminate() {
		t.Error("Expected ShouldTerminate() to return false")
	}
}
