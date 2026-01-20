# Changelog

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
