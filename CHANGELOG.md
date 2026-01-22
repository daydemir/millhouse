# Changelog

## [0.4.13] - 2026-01-22

### Critical Fixes
- **Fixed prd.json parsing failure when LLM corrupts JSON structure**
  - `Load()` now detects bare arrays `[...]` and auto-wraps in `{"prds": ...}`
  - Auto-saves the fixed version after recovery
  - Handles empty files gracefully
  - Prints warning so user knows recovery happened
  - Prevents `mil run` and `mil chat` from crashing on malformed prd.json

- **Fixed token counts not displaying after phase completion**
  - Matched Ralph's proven token counting approach
  - Extract tokens from `assistant` event (where Ralph gets them)
  - Removed double-counting from `result` event
  - Simplified TotalTokens = Input + Output (exclude cache tokens)
  - Added fallback warning when no token data captured

### Root Cause
- Token extraction was happening in wrong event (`result`) instead of `assistant`
- `result` event was double-counting tokens already captured from `message_start`/`message_delta`
- In some cases, `event.Usage` in result was nil, causing zero tokens

### Changed
- `internal/prd/prd.go`:
  - Added `bytes` import for JSON detection
  - Rewrote `Load()` function with resilient parsing (detects bare arrays, auto-fixes, handles empty files)
- `internal/llm/output.go`:
  - Added token extraction from `assistant` event (lines 336-346)
  - Removed token extraction from `result` event (was causing double-counting)
  - Changed `recalculateTotalAndCheckThreshold()` to exclude cache tokens from total
  - Added warning display when no tokens captured in `DisplayFinalTokenUsage()`
- `internal/llm/output_test.go`:
  - Updated `TestOnTokenUsage_CacheReadTokensTracked()` - TotalTokens now 65300 (not 77300)
  - Updated `TestOnTokenUsageCumulative_TakesMaxOutputTokens()` - TotalTokens now 30300 (not 32100)

### Technical
- Ralph (working reference) only processes `assistant` event for token data
- Ralph's TotalTokens = InputTokens + OutputTokens (excludes cache)
- Cache tokens still tracked separately for debugging but not in total

### Breaking Changes
None - token counting changes are internal implementation details.

### Migration Notes
No action required. Token displays will now show accurate values after each phase.

## [0.4.11] - 2026-01-22

### Critical Fixes
- **Fixed token count display showing ~0.1K when actual usage is ~40K+**
  - TotalTokens now includes all 4 token types: input, output, cache_read, cache_creation
  - Added `CacheCreationTokens` field to TokenStats struct
  - Updated `recalculateTotalAndCheckThreshold()` to include cache tokens in total
  - Updated `ParseStream()` to capture `cache_creation_input_tokens` from Claude CLI
  - Updated `OnTokenUsage()` to accumulate CacheCreationTokens

### Root Cause
- With prompt caching enabled, most tokens come from `cache_read_input_tokens` (~30-50K)
- Previous calculation only summed `input_tokens` + `output_tokens` (~100-300 total)
- Result: displayed "0.1K" instead of actual "40K" context usage

### Changed
- `internal/llm/output.go`:
  - TokenStats struct: Added `CacheCreationTokens` field
  - `recalculateTotalAndCheckThreshold()`: Now sums all 4 token types
  - `OnTokenUsage()`: Now accumulates CacheCreationTokens
  - `ParseStream()`: Now captures CacheCreationTokens from result event
- `internal/llm/output_test.go`:
  - Updated `TestOnTokenUsage_CacheReadTokensTracked()` expectations
  - Updated `TestOnTokenUsageCumulative_TakesMaxOutputTokens()` expectations

### Technical
- Token types from Claude CLI with prompt caching:
  - `input_tokens`: New incremental prompt tokens (~3-10)
  - `output_tokens`: Generated response tokens (~100-300)
  - `cache_read_input_tokens`: Tokens read from cache (~30K-50K) - bulk of context
  - `cache_creation_input_tokens`: Tokens used to create new cache
- All 4 types now contribute to TotalTokens for accurate threshold checking

### Breaking Changes
None - token counting changes are internal implementation details.

### Migration Notes
No action required. Token displays will now show accurate values reflecting actual context usage.

## [0.4.10] - 2026-01-20

### Critical Fixes
- **Restored Claude CLI stream event handling** - Fixed regression from v0.4.7 where `mil run` stopped working
  - Restored `message_start` event handler for initial input token counting
  - Restored `message_delta` event handler for incremental output token updates
  - Re-added `OnTokenUsageCumulative()` method to OutputHandler interface
  - Token counting now works correctly with Claude CLI's actual stream-json event format
  - Planner now successfully creates plans and selects PRDs
  - Builder now executes plans for active PRDs
  - Complete planner → builder → reviewer workflow restored

### Root Cause
- v0.4.7 removed `message_start` and `message_delta` handlers in favor of `assistant` event
- Claude CLI's `--output-format stream-json` does NOT emit `assistant` events
- Actual events: `message_start`, `content_block_delta`, `message_delta`, `message_stop`
- Without proper handlers, token counts stayed at 0 and threshold checking failed

### Changed
- `internal/llm/output.go`:
  - Added `message_start` case in ParseStream() - handles initial input tokens
  - Added `message_delta` case in ParseStream() - handles cumulative output tokens
  - Restored `OnTokenUsageCumulative()` method - takes max for output tokens, accumulates cache reads
  - Updated OutputHandler interface with `OnTokenUsageCumulative()` method
- `internal/llm/output_test.go`:
  - Added `TestOnTokenUsageCumulative_TakesMaxOutputTokens()` - verifies max() behavior for output tokens

### Technical
- `message_start` provides snapshot of input tokens (counted once via OnTokenUsage)
- `message_delta` provides cumulative output tokens (take max via OnTokenUsageCumulative)
- Cache read tokens accumulate in both methods
- Maintains improved token counting logic from v0.4.7+ while supporting actual Claude CLI events

### Breaking Changes
None - internal implementation fix only.

### Migration Notes
No action required. `mil run` will now work correctly again:
- Token counts update in real-time during execution
- Planner creates plans and selects PRDs as expected
- Builder executes plans for active PRDs
- Full workflow iterations complete successfully

## [0.4.8] - 2026-01-20

### Fixed
- **#66**: Display label now shows "[reviewer]" instead of "[analyzer]" during Phase 3
  - Updated display methods to use "reviewer" terminology consistently
  - Renamed theme colors from Analyzer* to Reviewer* for clarity
  - Removed unused `Analysis()` display method
  - Updated help text and comments to reference "reviewer" agent
  - Files: internal/display/display.go, internal/display/theme.go, internal/cli/root.go, internal/prd/selector.go
- **#65**: MillChat now uses PRD helper functions to avoid reading large prd.json files
  - Added ActivePRDs count to chat context (shows in-progress PRDs)
  - Updated chat prompt to document available PRD helper functions
  - Added guidance to avoid reading entire prd.json when unnecessary
  - Chat now displays: "X total (Y open, Z active, A pending, B complete)"
  - Files: internal/prompts/prompts.go, internal/builder/builder.go, internal/prompts/chat.tmpl

### Changed
- Phase 3 Reviewer now displays "[reviewer] Starting review..." instead of "[analyzer] Starting analysis..."
- Theme field names updated: AnalyzerGutter → ReviewerGutter, AnalyzerText → ReviewerText
- Gutter constant renamed: GutterAnalyzer → GutterReviewer

### Improved
- Chat experience by providing better context about PRD states (now shows active PRDs)
- Chat performance by guiding Claude to use efficient helper functions instead of reading full prd.json
- Code clarity by using consistent "reviewer" terminology throughout the codebase

## [0.4.7] - 2026-01-20

### Critical Fixes
- **#58**: Adopted Ralph's token counting logic for accuracy and simplicity
  - Changed InputTokens from max() to simple accumulation (matches Ralph's proven approach)
  - Added CacheReadTokens tracking (tracked separately, not included in total)
  - Removed complex dual-method system (OnTokenUsage + OnTokenUsageCumulative)
  - Simplified stream parsing to only process tokens from `assistant` event (final authoritative counts)
  - Token counts now accurately accumulate across multi-turn conversations
  - Fixed discrepancies between displayed and actual token usage

### Changed
- `internal/llm/output.go` - Complete token counting refactor:
  - TokenStats: Added `CacheReadTokens` field
  - UsageBlock: Added `CacheCreationTokens` and `CacheReadTokens` fields
  - OnTokenUsage(): Changed to simple accumulation (`InputTokens += usage.InputTokens`)
  - Removed `updateInputTokens()` method (no longer needed)
  - Removed `OnTokenUsageCumulative()` method (simplified to single method)
  - OutputHandler interface: Removed `OnTokenUsageCumulative()` method
  - ParseStream(): Removed token handling from `message_start` and `message_delta` events
  - ParseStream(): Now only processes tokens from `assistant` event
- `internal/llm/output_test.go` - Updated all tests for accumulation behavior:
  - Renamed `TestOnTokenUsage_InputTokensNotAccumulated` → `TestOnTokenUsage_InputTokensAccumulated`
  - Renamed `TestOnTokenUsage_TakesMaxInput` → `TestOnTokenUsage_AccumulatesAllTokens`
  - Removed `TestOnTokenUsageCumulative_TakesMaxInput` (method removed)
  - Removed `TestOnTokenUsageCumulative_ReplacesOutputTokens` (method removed)
  - Added `TestOnTokenUsage_CacheReadTokensTracked` (new functionality)

### Improved
- Token counting accuracy through simple accumulation logic
- Code maintainability by removing complex dual-method system
- Reliability by using Claude API's final authoritative token counts

### Technical
- Token accumulation now matches daydemir-ralph reference implementation
- CacheReadTokens tracked separately (not included in TotalTokens calculation)
- Single event source (`assistant`) prevents double-counting and synchronization issues
- All tests passing with new accumulation behavior

### Breaking Changes
None - token counting changes are internal implementation details.

### Migration Notes
No action required. Token counting improvements are transparent to users:
- Token displays will show accurate accumulated counts
- CacheReadTokens tracked internally (may be displayed in future release)

## [0.4.6] - 2026-01-19

### Added
- **#55**: XML structure support in PRD notes and descriptions
  - Chat agent now generates structured XML for complex notes (hints, blockers, gotchas, references)
  - Planner parses XML `<blockers>` tags for automatic dependency detection
  - Builder highlights XML hints and gotchas for implementation guidance
  - Reviewer cross-checks XML gotchas and references during verification
  - Fully backward compatible with plain text PRDs
  - Note: XML is for agent-to-agent communication only, `mil status` output unchanged

- **#38**: Thorough PRD validation in Planner
  - Added validation step 1.5 between "Analyze PRDs" and "Select PRD"
  - Validates logical coherence, architectural fit, assumptions, and feasibility
  - Flags PRDs with vague criteria, contradictions, or missing details
  - Blocks PRDs with severe issues (contradictory requirements, requires manual intervention)
  - Includes comprehensive validation examples in planner.tmpl
  - Documents validation findings in PRD notes using XML format

- **#19**: Prompt template optimizations
  - Created `internal/prompts/shared.tmpl` for common template partials
  - Reduced template verbosity by ~50 lines while maintaining all functionality
  - Standardized XML structure across all templates (element-based, consistent nesting)
  - Added XML usage guidance to all agent templates

### Changed
- `internal/prompts/chat.tmpl` - Added XML format guidance with examples (45 lines added)
- `internal/prompts/planner.tmpl` - Added validation step 1.5, XML parsing guidance, and validation examples (70 lines added)
- `internal/prompts/builder.tmpl` - Added PRD notes parsing section, trimmed context_management (16 lines saved)
- `internal/prompts/reviewer.tmpl` - Added XML awareness to verification and blocker evaluation (31 lines saved)
- `internal/prompts/prompts.go` - Updated to load shared.tmpl with template cloning
- `internal/prompts/planner.tmpl` - Changed `<prds status="open">` to `<prds><status>open</status>`
- `internal/prompts/reviewer.tmpl` - Changed `<plan prd="...">` to `<plan><prd_id>...</prd_id>`

### Improved
- PRD quality through Chat agent's XML schema guidance
- Agent communication through structured XML parsing
- Planner thoroughness through assumption validation
- Template maintainability through shared partials
- Prompt clarity through reduced verbosity

### Technical
- Template line count: ~807 → ~760 (6% reduction) while adding significant functionality
- Agents parse XML as text using natural language understanding (no Go parsing utilities needed)
- All template changes are backward compatible with existing plain text PRDs

### Breaking Changes
None - all changes are backward compatible with existing PRDs.

### Migration Notes
No action required. XML format is optional and only used when Chat agent determines it's beneficial:
- Existing plain text PRDs continue to work unchanged
- New PRDs may use XML for complex notes via Chat agent
- Both formats work simultaneously
- Planner validation works with both XML and plain text formats

## [0.4.4] - 2026-01-19

### Critical Fixes
- **#58**: Fixed token count display showing incorrect values during streaming
  - Removed inline token display (was showing stale counts)
  - Enhanced final token display with detailed Input/Output/Total breakdown
  - Token counts now accurately match agent's reported usage

- **#56**: Added strict git verification (Hawk Mode) to catch phantom commits
  - Created `internal/git/verify.go` with comprehensive verification utilities
  - Updated Reviewer prompt with automated git verification protocol
  - Reviewer now verifies commit existence, file presence, working tree cleanliness
  - Enhanced Builder evidence structure to include verification output

- **#59**: Made Reviewer skeptical of blocker claims
  - Added automation-first investigation protocol
  - Requires automation checklist before accepting blockers
  - Spawns Opus subagent for deep research when needed
  - Documents investigation evidence in PRD notes

### UX Improvements
- **#63**: Pending PRDs now visible in `mil status` compact mode
  - Shows up to 10 pending PRDs with pause symbol (⏸)
  - Displays "X more..." for overflow
  - Improves visibility into verification queue

- **#64**: Early exit when no productive work detected
  - Created iteration state tracking system
  - Exits after 2 consecutive idle iterations (configurable)
  - Saves tokens by preventing redundant work
  - Added `earlyExit` configuration option

### Added
- `internal/git/verify.go` - Git verification utilities
- `internal/git/verify_test.go` - Comprehensive test suite
- `internal/cli/iteration_state.go` - Iteration state tracking
- `internal/display/display.go::TokenUsageDetailed()` - Enhanced token display

### Changed
- `internal/llm/output.go` - Removed inline token display during streaming
- `internal/prompts/reviewer.tmpl` - Added strict verification and blocker evaluation sections
- `internal/prompts/builder.tmpl` - Enhanced evidence structure requirements
- `internal/display/display.go::PRDStatusCompact()` - Added pause symbol for pending PRDs
- `internal/cli/status.go` - Added pending PRDs section in compact mode
- `internal/config/config.go` - Added EarlyExitConfig
- `internal/cli/run.go` - Added early exit detection logic

### Breaking Changes
None - all changes are backward compatible.

### Migration Notes
No action required. New features are enabled by default with sensible defaults:
- Early exit enabled with 2-iteration threshold
- Can be disabled in `.milhouse/config.yaml` if needed
