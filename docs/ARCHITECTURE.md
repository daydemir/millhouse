# Millhouse Architecture

This document describes the architecture of Millhouse, an autonomous PRD (Product Requirements Document) execution system.

## Overview

Millhouse uses a three-phase iteration cycle to implement PRDs autonomously:

```
┌─────────┐     ┌─────────┐     ┌──────────┐
│ Planner │ ──► │ Builder │ ──► │ Reviewer │
└─────────┘     └─────────┘     └──────────┘
     │               │               │
     │               │               │
     └───────────────┴───────────────┘
              (iteration loop)
```

Each iteration:
1. **Planner** selects an open PRD and creates an implementation plan
2. **Builder** executes the plan to implement the PRD
3. **Reviewer** verifies completion or updates plans for bailouts

## PRD State Machine

PRDs transition through four states during their lifecycle:

```
false (open) → "active" (has plan) → "pending" (claimed done) → true (verified)
         ↑                   ↓                    ↓
         └───── REJECTED ────┴──── BAILOUT ───────┘
                (reviewer)        (stays active)
```

### State Definitions

| State | Field Value | Description |
|-------|-------------|-------------|
| **Open** | `passes: false` | Not attempted or rejected, needs work |
| **Active** | `passes: "active"` | Planner selected, has plan, builder working |
| **Pending** | `passes: "pending"` | Builder claims complete, awaiting reviewer |
| **Complete** | `passes: true` | Reviewer verified, work is done |

### Transitions

| From | To | Triggered By | Action |
|------|-----|--------------|--------|
| Open | Active | Planner | Planner creates plan, sets `activePlan` path |
| Active | Pending | Builder | Builder completes all criteria, signals `PRD_COMPLETE` |
| Active | Active | Builder bailout | Reviewer updates plan with progress |
| Pending | Complete | Reviewer | Reviewer verifies, deletes plan |
| Pending | Open | Reviewer | Reviewer rejects, deletes plan, adds notes |

## Agent Responsibilities

### Planner (`internal/planner/`)

The Planner agent runs at the start of each iteration when there are open PRDs and no active PRDs.

**Responsibilities:**
- Analyze all open PRDs for dependencies and priorities
- Select the best candidate PRD to work on
- Explore the codebase to understand implementation context
- Create a detailed implementation plan
- Save plan to `.millhouse/plans/{prd-id}-plan.md`
- Set PRD state to "active" with `activePlan` path

**Signals:**
- `###PLAN_COMPLETE:{prd-id}###` - Plan created successfully
- `###PLAN_SKIPPED:{reason}###` - No planning needed
- `###BLOCKED:{reason}###` - Cannot create plan

### Builder (`internal/builder/`)

The Builder agent executes plans for active PRDs.

**Responsibilities:**
- Read the implementation plan for the active PRD
- Execute each step in sequence
- Verify each step as specified in the plan
- Create evidence file with verification details
- Commit changes with descriptive messages
- Signal completion when all criteria are met

**Signals:**
- `###PRD_COMPLETE###` - All acceptance criteria met
- `###BAILOUT:{reason}###` - Context limit reached, partial work done
- `###BLOCKED:{reason}###` - Human intervention needed

### Reviewer (`internal/reviewer/`)

The Reviewer agent runs after the Builder phase to verify work and manage plans.

**Responsibilities:**
- Verify pending PRDs against acceptance criteria and evidence
- Handle bailouts by updating plans with progress
- Clean up plans when PRDs are verified or rejected
- Cross-pollinate learnings across PRDs
- Prevent stuck loops by detecting repeated failures

**Signals:**
- `###VERIFIED:{prd-id}###` - PRD confirmed complete
- `###REJECTED:{prd-id}:{reason}###` - PRD needs more work
- `###PLAN_UPDATED:{prd-id}###` - Plan updated after bailout
- `###LOOP_RISK:{prd-id}###` - PRD stuck in loop
- `###ANALYSIS_COMPLETE###` - Review phase done

## File Organization

```
.millhouse/
├── prd.json           # PRD definitions and state
├── progress.md        # Iteration logs and learnings
├── prompt.md          # Codebase patterns and context
├── evidence/          # Verification evidence files
│   └── {prd-id}-evidence.md
└── plans/             # Implementation plans (ephemeral)
    └── {prd-id}-plan.md
```

### Plan Files

Plans are ephemeral - created by Planner, executed by Builder, cleaned by Reviewer.

```markdown
# Plan: {prd-id}

## Overview
<summary>
Brief description of approach
</summary>

## Implementation Steps
<steps>
### Step 1: {title}
<files>path/to/file.go - Create/Modify: description</files>
<actions>
1. Specific action
2. Another action
</actions>
<verification>How to verify step complete</verification>
</steps>

## Acceptance Criteria Mapping
<criteria_mapping>
| Criterion | Step(s) |
|-----------|---------|
| First criterion | 1, 3 |
</criteria_mapping>
```

## Iteration Flow

```go
for i := 1; i <= iterations; i++ {
    // Load fresh PRD state
    prdFile := prd.Load(basePath)

    // Phase 1: Planner (if no active PRD)
    if planner.ShouldRunPlanner(prdFile) {
        planner.Run(ctx, basePath, prdFile)
        prdFile = prd.Load(basePath) // Reload after changes
    }

    // Phase 2: Builder (if active PRD exists)
    if builder.ShouldRunBuilder(prdFile) {
        builder.Run(ctx, basePath, prdFile)
        prdFile = prd.Load(basePath) // Reload after changes
    }

    // Phase 3: Reviewer (if pending or active PRDs)
    if reviewer.ShouldRunReviewer(prdFile) {
        reviewer.Run(ctx, basePath, prdFile, i)
    }
}
```

## Token Management

Each agent has a token threshold to prevent context exhaustion:

| Agent | Threshold | Rationale |
|-------|-----------|-----------|
| Planner | 80K | Planning is exploratory, smaller scope |
| Builder | 100K | Implementation needs most context |
| Reviewer | 80K | Verification is targeted, smaller scope |

When approaching the threshold, agents should:
1. Save progress to `progress.md`
2. Update PRD notes with current state
3. Signal bailout with reason

## Signal Protocol

Signals use a triple-hash format for reliable detection:

```
###SIGNAL_TYPE###             # Simple signal
###SIGNAL_TYPE:details###     # Signal with details
###SIGNAL_TYPE:id:reason###   # Signal with multiple parts
```

## Design Principles

1. **Plans are ephemeral, PRDs are persistent** - Plans are created/destroyed each cycle; PRD state is the source of truth

2. **One active PRD at a time** - Ensures focused execution and clear state

3. **Reviewer owns lifecycle** - Plan updates, cleanup, and state transitions are reviewer's responsibility

4. **Fail forward** - Bailouts preserve progress; rejections add guidance for next attempt

5. **Evidence-based verification** - Reviewer checks evidence files, not just builder claims
