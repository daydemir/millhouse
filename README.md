# Milhouse

> autonomous claude code ralph loops, but with glasses

<p align="center">
  <img src="assets/milhouse.webp" alt="milhouse" height="300">
</p>

> [!CAUTION]
> **Use at your own risk.** Milhouse runs with `--dangerously-skip-permissions` under the hood.
>
> **Recommended:**
> - Run in a sandboxed environment
> - Use [Claude Code Safety Net](https://github.com/kenryu42/claude-code-safety-net) for protection
> - Claude Code Max plan recommended (consumes significant tokens)
> - Start with small iterations to understand behavior

Milhouse is a self-improving autonomous Claude Code runner built on the concept of Ralph Loops. You create Product Requirement Documents (PRDs) by calling `mil chat` and discussing what PRDs you want to make. Then call `mil run` and Milhouse spawns agents to plan, build, and review work to complete your PRDs. As it reviews its work, Milhouse adds more context and guidance to improve the likelihood for successful completion in future iterations.

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
- [How Milhouse Works](#how-milhouse-works)
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

## How Milhouse Works

### Creating PRDs with `mil chat`

Use `mil chat` to define what needs to be built. Milhouse acts as your product manager, asking clarifying questions to create well-structured PRDs with clear acceptance criteria. You can also edit existing PRDs or update system prompts to teach Milhouse about your codebase patterns.

### Autonomous Execution with `mil run`

Once PRDs are defined, `mil run` orchestrates three autonomous agents in a resilient feedback loop. Each iteration completes a full cycle, with work flowing through PRD states: **Open → Active → Pending → Complete**.

### The Planner Agent

Selects one open PRD based on dependencies and priority, explores the codebase to understand implementation context, and creates a detailed plan mapping acceptance criteria to specific steps. Plans are written to `.milhouse/plans/{prd-id}-plan.md`.

### The Builder Agent

Executes the Planner's steps sequentially, verifying each step and committing changes incrementally. Documents discoveries in `progress.md` and creates evidence files showing what was completed. The Builder gracefully bails out at ~100K tokens to preserve context, documenting progress for resumption in the next iteration.

### The Reviewer Agent

**Core mandate:** Never leave state unchanged—always modify something to improve the situation.

Verifies pending PRDs meet all acceptance criteria (not just "build succeeded"), updates plans when the Builder bails out, and prevents stuck cycles by ensuring state and context is always changing. Cross-pollinates learnings from one PRD to all future PRDs via `progress.md`.

### Resilience & Iteration

The system handles interruptions and improves over time:

- **Token-based bailout**: Agents stop before hitting limits, preserving context for resumption
- **Plan lifecycle**: Plans are updated with actual progress, not discarded on failure
- **Loop prevention**: Reviewer prevents stuck cycles by ensuring state and context always changes
- **Cross-pollination**: Discoveries from one PRD improve all future PRDs via `progress.md`

Each iteration refines understanding. Failed attempts add context for the next try, making the system progressively smarter about your codebase.

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
