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
- [Why Milhouse](#why-milhouse)
- [Prerequisites & Installation](#prerequisites--installation)
- [Quick Start](#quick-start)
- [Basic Commands](#basic-commands)
- [Token Usage & Costs](#token-usage--costs)
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

## Token Usage & Costs

**Per-phase estimates:**
- **Planner:** 15-30K tokens
- **Builder:** 30-80K tokens
- **Reviewer:** 10-20K tokens

**Real-world costs** (Claude 3.5 Sonnet @ $3/$15 per 1M tokens):

| Command | Token Range | Cost |
|---------|---|---|
| `mil run 1` | 55-130K | ~$0.40-$1.30 |
| `mil run 5` | 275-650K | ~$2.00-$6.50 |
| `mil run 10` | 550-1.3M | ~$4.00-$13.00 |

**Optimize costs:**
- Use `sonnet` for faster, cheaper iterations
- Use `opus` for complex or critical work
- Start with `mil run 1` to validate before scaling
- Configure `maxTokens` per phase in config

See [CONFIGURATION.md](docs/CONFIGURATION.md) for all optimization options.

## Next Steps

- **Understand the system:** Read [ARCHITECTURE.md](docs/ARCHITECTURE.md) for the three-phase cycle
- **Optimize configuration:** See [CONFIGURATION.md](docs/CONFIGURATION.md) for all options
- **Troubleshoot issues:** Check [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for common problems
- **Contribute:** Found a bug or want to improve Milhouse? [Open an issue](https://github.com/daydemir/milhouse/issues)
- **Release guide:** Check [RELEASING.md](RELEASING.md) if you're a maintainer

## License

[MIT License](LICENSE)
