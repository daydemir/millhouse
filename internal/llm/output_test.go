package llm

import "testing"

func TestOnTokenUsage_InputTokensNotAccumulated(t *testing.T) {
	handler := NewConsoleHandler()

	// Simulate message_start with 35000 input tokens
	handler.OnTokenUsage(TokenStats{InputTokens: 35000, OutputTokens: 100})

	// Simulate assistant event with same context (should NOT accumulate)
	handler.OnTokenUsage(TokenStats{InputTokens: 35000, OutputTokens: 200})

	stats := handler.GetTokenStats()

	if stats.InputTokens != 35000 {
		t.Errorf("InputTokens should be 35000 (max), got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 300 {
		t.Errorf("OutputTokens should be 300 (accumulated), got %d", stats.OutputTokens)
	}
}

func TestOnTokenUsage_TakesMaxInput(t *testing.T) {
	handler := NewConsoleHandler()

	handler.OnTokenUsage(TokenStats{InputTokens: 30000, OutputTokens: 0})
	handler.OnTokenUsage(TokenStats{InputTokens: 35000, OutputTokens: 100})
	handler.OnTokenUsage(TokenStats{InputTokens: 33000, OutputTokens: 50})

	stats := handler.GetTokenStats()

	if stats.InputTokens != 35000 {
		t.Errorf("InputTokens should be 35000 (max), got %d", stats.InputTokens)
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

func TestOnTokenUsageCumulative_TakesMaxInput(t *testing.T) {
	handler := NewConsoleHandler()

	// Set initial input tokens
	handler.OnTokenUsage(TokenStats{InputTokens: 30000, OutputTokens: 0})

	// Cumulative call with higher input should update
	handler.OnTokenUsageCumulative(TokenStats{InputTokens: 35000, OutputTokens: 500})

	stats := handler.GetTokenStats()

	if stats.InputTokens != 35000 {
		t.Errorf("InputTokens should be 35000 (max), got %d", stats.InputTokens)
	}
	if stats.OutputTokens != 500 {
		t.Errorf("OutputTokens should be 500 (cumulative), got %d", stats.OutputTokens)
	}
}

func TestOnTokenUsageCumulative_ReplacesOutputTokens(t *testing.T) {
	handler := NewConsoleHandler()

	// Initial tokens
	handler.OnTokenUsage(TokenStats{InputTokens: 30000, OutputTokens: 100})

	// Cumulative output should replace, not add
	handler.OnTokenUsageCumulative(TokenStats{OutputTokens: 500})
	handler.OnTokenUsageCumulative(TokenStats{OutputTokens: 800})

	stats := handler.GetTokenStats()

	if stats.OutputTokens != 800 {
		t.Errorf("OutputTokens should be 800 (replaced), got %d", stats.OutputTokens)
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
