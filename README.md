# Millhouse

Autonomous multi-phase software development using Claude AI. Millhouse orchestrates three specialized agents (planner, builder, reviewer) that work together to implement software requirements.

## Features

- **Three-phase workflow**: Planner → Builder → Reviewer
- **Autonomous execution**: Run multiple iterations to completion
- **Interactive sessions**: Use `mill discuss` for real-time collaboration
- **Flexible configuration**: Customize models, token limits, and context per phase
- **Cost optimization**: Choose different models for different phases
- **Progress tracking**: Automatic progress.md logging of all work

## Configuration

Millhouse supports extensive configuration for model selection, token limits, and context settings:

```yaml
# .millhouse/config.yaml
phases:
  planner:
    model: "sonnet"
    maxTokens: 80000
    progressLines: 20

  builder:
    model: "sonnet"
    maxTokens: 100000
    progressLines: 20

  reviewer:
    model: "sonnet"
    maxTokens: 80000
    progressLines: 200
```

### Quick Configuration

**Manage configuration interactively:**

```bash
mill config edit        # Open interactive TUI editor
mill config show        # Show effective configuration
mill config init        # Initialize with defaults
```

**Override for single run:**

```bash
mill run 1 --planner-model haiku --builder-model opus
mill run 2 --planner-max-tokens 60000 --builder-max-tokens 150000
```

**See also:** [Configuration Guide](docs/CONFIGURATION.md) for detailed options, use cases, and best practices.

## Usage

### Initialize a project

```bash
mill init
```

### Add requirements

Edit `.millhouse/prd.json` to add requirements, or use:

```bash
mill discuss
```

### Run autonomously

```bash
mill run 1      # 1 iteration
mill run 5      # 5 iterations
```

### Check status

```bash
mill status
```

## Commands

| Command | Purpose |
|---------|---------|
| `mill init` | Initialize Millhouse in current directory |
| `mill run N` | Execute N autonomous iterations |
| `mill discuss` | Start interactive session |
| `mill status` | Show PRD status |
| `mill config edit` | Edit configuration interactively |
| `mill config show` | Display effective configuration |
| `mill config init` | Initialize config file with defaults |

## Configuration

Millhouse follows a configuration precedence order:

1. CLI flags (highest priority)
2. Project config (`.millhouse/config.yaml`)
3. User global config (`~/.millhouse/config.yaml`)
4. Built-in defaults (lowest priority)

**Example configurations:**

Cost optimization:
```yaml
phases:
  planner: {model: "haiku", maxTokens: 50000}
  builder: {model: "sonnet", maxTokens: 100000}
  reviewer: {model: "haiku", maxTokens: 50000}
```

Quality focus:
```yaml
phases:
  planner: {model: "opus", maxTokens: 150000}
  builder: {model: "opus", maxTokens: 150000}
  reviewer: {model: "sonnet", maxTokens: 100000}
```

See [Configuration Guide](docs/CONFIGURATION.md) for more use cases and detailed documentation.
