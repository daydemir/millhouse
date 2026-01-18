# Configuration Guide

Milhouse supports extensive configuration options for model selection, token limits, and context settings per phase. This guide covers how to configure Milhouse for your specific needs.

## Overview

Configuration in Milhouse follows a precedence order from highest to lowest priority:

1. **CLI flags** - Command-line overrides (highest priority)
2. **Project config** - `.milhouse/config.yaml` (project-specific settings)
3. **User global config** - `~/.milhouse/config.yaml` (personal defaults)
4. **Built-in defaults** - Hardcoded defaults in Milhouse (lowest priority)

## Configuration File Format

Configuration files use YAML format. Create a `.milhouse/config.yaml` file in your project:

```yaml
# Global defaults (applied to all phases unless overridden)
global:
  model: "sonnet"          # haiku, sonnet, or opus
  maxTokens: 100000        # 10,000 to 200,000

# Phase-specific settings (override global defaults)
phases:
  planner:
    model: "sonnet"        # Model for planning phase
    maxTokens: 80000       # Token limit for planner
    progressLines: 20      # Lines of progress.md to include

  builder:
    model: "sonnet"        # Model for building phase
    maxTokens: 100000      # Token limit for builder
    progressLines: 20      # Lines of progress.md to include

  reviewer:
    model: "sonnet"        # Model for reviewing phase
    maxTokens: 80000       # Token limit for reviewer
    progressLines: 200     # Lines of progress.md to include (reviewers need more history)

  chat:
    model: "sonnet"        # Model for interactive chat sessions

# Optional: Additional context files to pass to agents
contextFiles:
  - "docs/ARCHITECTURE.md"
  - "CONTRIBUTING.md"
```

## Configuration Options

### Models

**Valid values:** `haiku`, `sonnet`, `opus` (or full model IDs like `claude-3-5-sonnet-20241022`)

Each phase can use a different model:
- **Haiku**: Fast, cheaper, good for simple tasks
- **Sonnet**: Balanced (default), good general-purpose model
- **Opus**: More capable, more expensive, best for complex reasoning

### Token Limits

**Valid range:** 10,000 to 200,000 tokens per phase

Token limits control how much context each agent can process before automatically stopping (bailout). Higher limits allow more code context but cost more.

**Default values:**
- Planner: 80,000 tokens
- Builder: 100,000 tokens (most generous for implementation)
- Reviewer: 80,000 tokens

### Progress Lines

**Valid range:** 10 to 1,000 lines per phase

Controls how many lines of `progress.md` each phase receives for context. Higher values give more historical context.

**Default values:**
- Planner: 20 lines
- Builder: 20 lines
- Reviewer: 200 lines (reviewers benefit from more history)

### Context Files

Optional additional documentation files to pass to agents. Paths are relative to the project root.

## Managing Configuration

### Interactive Editor

Open the interactive configuration editor with:

```bash
mil config edit
```

Navigate with arrow keys (↑/↓), edit values directly, and save with Ctrl+S or cancel with ESC.

### View Current Configuration

Display the effective configuration (merged from all sources):

```bash
mil config show
```

This helps you see which configuration is actually being used.

### Initialize Configuration

Create a `.milhouse/config.yaml` with defaults:

```bash
mil config init
```

## CLI Flag Overrides

Override configuration values for a single run using CLI flags:

```bash
# Override model selection
mil run 1 --planner-model haiku --builder-model sonnet --reviewer-model haiku

# Override token limits
mil run 1 --planner-max-tokens 50000 --builder-max-tokens 150000

# Combine overrides
mil run 2 --planner-model haiku --planner-max-tokens 60000
```

CLI flags take highest priority, so they override both project and global config files.

## Use Cases

### Cost Optimization

For large projects where cost is primary concern, use cheaper models for planning and reviewing:

```yaml
# .milhouse/config.yaml
phases:
  planner:
    model: "haiku"
    maxTokens: 50000      # Also reduce tokens to save more

  builder:
    model: "sonnet"       # Keep sonnet for complex implementation

  reviewer:
    model: "haiku"        # Haiku is fine for verification
```

### Quality Focus

When quality is more important than cost, use Opus and increase tokens:

```yaml
phases:
  planner:
    model: "opus"
    maxTokens: 150000     # More budget for exploration

  builder:
    model: "opus"
    maxTokens: 150000     # More budget for implementation

  reviewer:
    model: "sonnet"       # Sonnet is sufficient for verification
```

### Large Codebases

For projects with large codebases, increase the context window for each phase:

```yaml
phases:
  planner:
    progressLines: 50     # More history

  builder:
    progressLines: 50     # More history

  reviewer:
    progressLines: 500    # Significantly more history for verification
```

### Custom Documentation

Include project-specific documentation:

```yaml
contextFiles:
  - "docs/ARCHITECTURE.md"
  - "docs/API.md"
  - "docs/CONVENTIONS.md"
  - "CONTRIBUTING.md"
```

### Multi-Phase Strategy

Different settings for different workloads - fast iteration for initial planning, higher quality for final verification:

```yaml
phases:
  planner:
    model: "haiku"        # Fast planning
    maxTokens: 60000

  builder:
    model: "opus"         # High-quality implementation
    maxTokens: 150000

  reviewer:
    model: "sonnet"       # Balanced verification
    maxTokens: 100000
```

### Chat Configuration

The chat phase is used for interactive sessions (`mil chat`). Unlike other phases, chat runs in interactive mode without token limits or progress line tracking.

```yaml
phases:
  chat:
    model: "opus"  # Use opus for high-quality interactive sessions
```

**Note:** Chat phase only supports `model` configuration. It does not use `maxTokens` or `progressLines` because it runs in interactive mode.

You can also override the chat model via CLI flag:

```bash
mil chat --model opus
```

## File Locations

### Project Config

**Location:** `.milhouse/config.yaml` (in project root)

**Scope:** Applies only to this project

**Use for:** Project-specific settings, team preferences

### User Global Config

**Location:** `~/.milhouse/config.yaml` (in home directory)

**Scope:** Applies to all Milhouse projects on this machine

**Use for:** Personal preferences (preferred model, token budget)

## Configuration Precedence

When the same setting appears in multiple places, this is the precedence order:

```
CLI flag (--planner-model haiku)
    ↓
Project config (.milhouse/config.yaml)
    ↓
User global config (~/.milhouse/config.yaml)
    ↓
Built-in defaults
```

Example: If your project config uses `sonnet`, but you run `mil run 1 --planner-model haiku`, then Haiku is used.

## Validation

Configuration values are validated on load:

- **Invalid models** trigger an error
- **Out-of-range tokens** (< 10K or > 200K) trigger an error
- **Invalid progress lines** (< 10 or > 1000) trigger an error

The editor shows validation errors in red and prevents saving invalid configurations.

## Default Configuration

If no configuration files exist, these built-in defaults are used:

```yaml
global:
  model: sonnet
  maxTokens: 100000

phases:
  planner:
    model: sonnet
    maxTokens: 80000
    progressLines: 20

  builder:
    model: sonnet
    maxTokens: 100000
    progressLines: 20

  reviewer:
    model: sonnet
    maxTokens: 80000
    progressLines: 200
```

## Tips and Best Practices

1. **Start with defaults** - The built-in defaults are reasonable for most projects
2. **Use project config** for team settings - Commit `.milhouse/config.yaml` to version control
3. **Use user config** for personal preferences - Keep `~/.milhouse/config.yaml` local
4. **Use CLI flags** for one-off experiments - `mil run 1 --planner-model haiku`
5. **Use `mil config show`** to verify your effective configuration before running
6. **Document your strategy** - Add comments to `.milhouse/config.yaml` explaining your choices
7. **Monitor costs** - Higher token limits and better models cost more; monitor usage

## Troubleshooting

### Config not being used

Check what configuration is actually being used:

```bash
mil config show
```

Remember the precedence order - CLI flags override file configs.

### Invalid configuration error

Configuration values are validated when loaded. Check:

- Model names are `haiku`, `sonnet`, or `opus` (case-sensitive)
- Token limits are between 10,000 and 200,000
- Progress lines are between 10 and 1,000

Use `mil config edit` to fix values interactively with validation.

### Permission denied on config file

Ensure the config file is readable:

```bash
chmod 644 .milhouse/config.yaml
chmod 755 ~/.milhouse/
chmod 644 ~/.milhouse/config.yaml
```

## Future Enhancements

Planned features for future versions:

- Environment variable support
- Per-PRD configuration overrides
- Configuration profiles (dev/prod/test)
- Customization of allowed tools per phase
- Cost estimation and budgeting
