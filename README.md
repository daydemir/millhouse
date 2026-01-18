# Milhouse

> Claude Code Ralph Loops with Glasses

Milhouse is an autonomous multi-phase software development system powered by Claude AI. It automatically plans, builds, and reviews code through iterative cycles, enabling you to scale your development work without manual intervention.

## Table of Contents

- [Quick Overview](#quick-overview)
- [Why Milhouse](#why-milhouse)
- [Prerequisites & Installation](#prerequisites--installation)
- [Quick Start](#quick-start)
- [Basic Commands](#basic-commands)
- [Understanding the Workflow](#understanding-the-workflow)
- [Common Use Cases](#common-use-cases)
- [Token Usage & Costs](#token-usage--costs)
- [Troubleshooting Basics](#troubleshooting-basics)
- [Next Steps](#next-steps)
- [License](#license)

## Quick Overview

Milhouse automates the development cycle through three phases:

1. **Planner** — Analyzes requirements and creates a detailed plan
2. **Builder** — Implements changes based on the plan
3. **Reviewer** — Tests and validates the implementation

You provide a product requirement, Milhouse executes the cycle autonomously, and you review the results. Run multiple iterations (`mil run 1`, `mil run 5`, etc.) to refine the output incrementally.

## Why Milhouse

| Aspect | Manual Coding | Claude Only | Milhouse |
|--------|---|---|---|
| **Planning** | Manual, time-consuming | No structured output | Automatic PRD generation |
| **Iteration** | Requires human feedback | Single response | Built-in multi-phase loops |
| **Verification** | Manual testing | No verification | Automatic reviewer phase |
| **Cost Control** | N/A | Fixed per request | Configurable via phases/model |

Milhouse is ideal for:
- Scaling development workflows
- Rapid prototyping with automated iteration
- Batch processing of similar tasks
- Cost-controlled autonomous coding

Milhouse is **not** ideal for:
- Real-time interactive development
- Highly specialized or novel problem domains
- Projects requiring deep domain expertise

## Prerequisites & Installation

**Requirements:**
- Claude Code CLI (`claude-code`)
- Go 1.21+ (for building from source)

**Installation:**

```bash
go install github.com/daydemir/milhouse/cmd/mil@latest
```

**Verify installation:**

```bash
mil version
```

## Quick Start

1. **Initialize a project:**
   ```bash
   cd your-project
   mil init
   ```

2. **Define your requirement:**
   ```bash
   mil discuss "Add user authentication with JWT tokens"
   ```

3. **Run the first iteration:**
   ```bash
   mil run 1
   ```

4. **Check status:**
   ```bash
   mil status
   ```

Look in `.milhouse/` for:
- `prd.json` — Product requirement document
- `progress.md` — Iteration progress
- `plans/` — Planner output for each phase
- `evidence/` — Builder and reviewer results

## Basic Commands

| Command | Purpose |
|---------|---------|
| `mil init` | Initialize a new Milhouse project |
| `mil discuss` | Define or update the product requirement |
| `mil run N` | Execute N iterations of the full cycle |
| `mil status` | Show current progress and phase state |
| `mil config edit` | Edit project configuration (costs, model selection) |
| `mil config show` | Display current configuration settings |

For detailed options and flags, see [CONFIGURATION.md](docs/CONFIGURATION.md).

## Understanding the Workflow

```
Planner → Builder → Reviewer → Feedback Loop
   ↓        ↓         ↓
 Plan    Implement  Validate
```

**PRD State Flow:**

- **Open** — Initial state, ready for planning
- **Active** — Plan accepted, builder is working
- **Pending** — Builder done, reviewer is evaluating
- **Complete** — Cycle finished, ready for next iteration

Each iteration cycles through all three phases. The reviewer phase validates the builder's output and provides feedback for the next iteration.

For deeper technical understanding, see [ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Common Use Cases

**Adding a new API endpoint:**
```bash
mil discuss "Add POST /api/users endpoint with validation"
mil run 3
```

**Refactoring a module:**
```bash
mil discuss "Refactor authentication module for readability"
mil run 2
```

**Cost optimization focused work:**
```bash
mil config edit
# Set model to claude-3-5-sonnet for faster iterations
mil run 5
```

**Quality-focused development:**
```bash
mil config edit
# Set model to claude-opus-4-5 for higher quality
mil run 1
```

## Token Usage & Costs

> [!WARNING]
> Running autonomous loops consumes Claude API quota and incurs costs. Start small to observe behavior before scaling.

**Per-phase token estimates:**
- **Planner:** 15-30K tokens
- **Builder:** 30-80K tokens
- **Reviewer:** 10-20K tokens

**Real-world cost examples (using Claude 3.5 Sonnet @ $3/$15 per 1M tokens):**

| Command | Token Range | Cost |
|---------|---|---|
| `mil run 1` | 55-130K | ~$0.40-$1.30 |
| `mil run 5` | 275-650K | ~$2.00-$6.50 |
| `mil run 10` | 550-1.3M | ~$4.00-$13.00 |

**Optimize costs:**
- Use `claude-3-5-sonnet` for faster, cheaper iterations
- Use `claude-opus-4-5` for complex or critical work
- Start with `mil run 1` to validate the workflow
- Set `max_iterations` in config to limit per-phase work

See [CONFIGURATION.md](docs/CONFIGURATION.md) for all cost optimization options.

## Troubleshooting Basics

| Issue | Solution |
|-------|----------|
| `mil: command not found` | Verify installation: `go install github.com/daydemir/milhouse/cmd/mil@latest` |
| `.milhouse/` directory errors | Ensure current directory is writable; run `mil init` first |
| API errors or rate limits | Check Claude Code CLI credentials: `claude-code` |
| Large token usage | Reduce iterations: use `mil run 1` instead of `mil run 5` |
| Unexpected output quality | Check configuration: `mil config show`; try `claude-opus-4-5` model |

For complex issues, open a [GitHub issue](https://github.com/daydemir/milhouse/issues).

## Next Steps

- **Understand the system:** Read [ARCHITECTURE.md](docs/ARCHITECTURE.md) for the three-phase cycle
- **Optimize for your use case:** See [CONFIGURATION.md](docs/CONFIGURATION.md) for all options
- **Release guide:** Check [RELEASING.md](RELEASING.md) if you're a maintainer
- **Contribute:** Found a bug or want to improve Milhouse? [Open an issue](https://github.com/daydemir/milhouse/issues)

## License

MIT License — See [LICENSE](LICENSE) file for full text.

Milhouse is built on Claude AI and the Claude Code framework.
