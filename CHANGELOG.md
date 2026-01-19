# Changelog

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
