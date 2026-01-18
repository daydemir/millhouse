# Milhouse

> autonomous claude code ralph loops, but with glasses

<p align="center">
  <img src="assets/milhouse.webp" alt="milhouse" height="300">
</p>

> [!CAUTION]
> **Use at your own risk.** Milhouse runs with `--dangerously-skip-permissions` under the hood.
>
> **Recommended:**
> - Use [Claude Code Safety Net](https://github.com/kenryu42/claude-code-safety-net) for protection
> - Claude Code Max plan recommended (consumes significant tokens)
> - Start with small iterations to understand behavior

Milhouse is an autonomous Claude Code runner built on the concept of Ralph Loops. You create Product Requirement Documents (PRDs) by calling `mil chat`, then run `mil run` and Milhouse spawns agents to plan, build, and review work to complete your PRDs.

## Inspiration

Milhouse builds on the Ralph Loop pattern for autonomous coding agents.

**Original Article:**
- [Ralph: The AI Coding Agent - Jeff Huntley](https://huntleyjoseph.com/blog/ralph)

**Related Projects:**
- [glittercowboy/get-shit-done](https://github.com/glittercowboy/get-shit-done) - GSD framework
- [michaelshimeles/ralphy](https://github.com/michaelshimeles/ralphy) - Ralph implementation
- [snarktank/ralph](https://github.com/snarktank/ralph) - Ralph variant
- [davis7dotsh/r8y-elixir-version](https://github.com/davis7dotsh/r8y-elixir-version) - see RALPH_LAND folder
- [AndyMik90/Auto-Claude](https://github.com/AndyMik90/Auto-Claude) - Auto-Claude

**Videos:**
- [We need to talk about Ralph by Theo - t3․gg](https://youtu.be/Yr9O6KFwbW4?si=xM6EOL8Pdvn83vAx)
- [The Ralph Wiggum Loop from 1st principles by Geoffrey Huntley](https://youtu.be/4Nna09dG_c0?si=4zFYeH1qja1piOEh)
- [An early preview of loom by Geoffrey Huntley](https://youtu.be/zX_Wq9wAyxI?si=Tirq4A-N3lJkxson)
- [Stop Using The Ralph Loop Plugin by Chase AI](https://youtu.be/yAE3ONleUas?si=XTzFJN3We-TNVUUy)
- [The New Claude Code Meta by Chase AI](https://youtu.be/SqmXS8q_2BM?si=tbcwW5WM4n34aTgq)

## Table of Contents

- [Quick Overview](#quick-overview)
- [How mil run Works](#how-mil-run-works)
- [Prerequisites & Installation](#prerequisites--installation)
- [Quick Start](#quick-start)
- [Basic Commands](#basic-commands)
- [Next Steps](#next-steps)
- [License](#license)

## Quick Overview

Milhouse automates development through three phases:

1. **Planner** — Analyzes requirements and creates detailed plans
2. **Builder** — Implements changes based on the plan
3. **Reviewer** — Tests and validates the implementation

**Workflow:** You provide a PRD → Planner creates a plan → Builder implements → Reviewer validates → Feedback loop continues

**PRD States:** Open → Active → Pending → Complete

Each iteration cycles through all phases. Run `mil run 1` for one iteration, `mil run 5` for five, etc.

## How `mil run` Works

Milhouse orchestrates three autonomous agents in a resilient feedback loop. Each iteration completes a full cycle, with work flowing through PRD states: **Open → Active → Pending → Complete**.

### The Planner Agent

**Goals:**
- Select one open PRD based on dependencies, priority, and readiness
- Explore the codebase to understand implementation context
- Create a detailed plan with specific files, functions, and verification steps
- Map acceptance criteria to implementation steps

The Planner only runs when no active PRD exists. It uses sub-agents to discover patterns and writes plans to `.milhouse/plans/{prd-id}-plan.md`.

### The Builder Agent

**Goals:**
- Execute the Planner's steps sequentially
- Verify each step (typecheck, lint, test as specified)
- Commit changes incrementally with descriptive messages
- Document discoveries and patterns in `progress.md`
- Create evidence files showing what was completed and how it was verified

The Builder monitors token usage and gracefully bails out at ~100K tokens to preserve context. When bailing, it documents progress so work can resume in the next iteration. Upon completion, it creates `.milhouse/evidence/{prd-id}-evidence.md` with verification flags (PARTIAL, INDIRECT, ASSUMPTION) to signal any limitations.

### The Reviewer Agent

**Core mandate:** Never leave state unchanged—always modify something to improve the situation.

**Goals:**
- Verify pending PRDs meet all acceptance criteria (not just "build succeeded")
- Assess verification flags and reject insufficient work
- Update plans when the Builder bails out (mark completed steps, clarify remaining work)
- Detect infinite loops (PRDs stuck in the same state for 2+ iterations)
- Cross-pollinate learnings across PRDs to prevent repeated failures

The Reviewer catches incomplete verifications, preserves progress from bailouts, and actively breaks stuck cycles by suggesting alternative approaches or reprioritizing work.

### Resilience & Iteration

The system handles interruptions and improves over time through:

- **Token-based bailout**: Agents stop before hitting limits, preserving context for resumption
- **Plan lifecycle**: Plans are updated with actual progress, not discarded on failure
- **Evidence flags**: Builders honestly flag verification limitations; Reviewers assess sufficiency
- **Loop detection**: Reviewer identifies stuck cycles and forces state changes
- **Cross-pollination**: Discoveries from one PRD improve all future PRDs via `progress.md`
- **State machine**: Enforces workflow (only one active PRD at a time, clear transitions)

Each iteration refines understanding. Failed attempts add context for the next try. Progress accumulates across iterations, making the system progressively smarter about your codebase.

## Prerequisites & Installation

**Requirements:**
- [Claude Code](https://code.claude.com/docs/en/overview)
- Homebrew (recommended) or Go 1.21+

**Install:**

```bash
# Recommended: Homebrew
brew install daydemir/tap/mil

# Or build from source
go install github.com/daydemir/milhouse/cmd/mil@latest
```

**Verify:**

```bash
mil version
```

## Quick Start

```bash
# Initialize project
cd your-project
mil init

# Create PRD interactively
mil chat

# Run first iteration
mil run 1

# Check status
mil status
```

**Output files** (in `.milhouse/`):
- `prd.json` — Product requirements
- `progress.md` — Iteration history
- `plans/` — Planner output per PRD
- `evidence/` — Builder and reviewer results

## Basic Commands

| Command | Purpose |
|---------|---------|
| `mil init` | Initialize a new Milhouse project |
| `mil chat` | Create or update PRDs interactively |
| `mil run N` | Execute N iterations of the full cycle |
| `mil status` | Show current progress and state |
| `mil config edit` | Edit configuration (model, tokens, etc.) |
| `mil config show` | Display current configuration |

**Common patterns:**

```bash
# Add new feature
mil chat "Add POST /api/users endpoint with validation"
mil run 3

# Refactor code
mil chat "Refactor auth module for readability"
mil run 2

# Use specific model
mil chat --model opus
mil run 1 --builder-model opus
```

For detailed configuration options, see [CONFIGURATION.md](docs/CONFIGURATION.md).

## Next Steps

- **Understand the system:** Read [ARCHITECTURE.md](docs/ARCHITECTURE.md) for the three-phase cycle
- **Optimize configuration:** See [CONFIGURATION.md](docs/CONFIGURATION.md) for all options
- **Troubleshoot issues:** Check [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for common problems
- **Contribute:** Found a bug or want to improve Milhouse? [Open an issue](https://github.com/daydemir/milhouse/issues)
- **Release guide:** Check [RELEASING.md](RELEASING.md) if you're a maintainer

## License

[MIT License](LICENSE)
