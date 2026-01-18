# Troubleshooting Guide

This guide covers common issues and solutions when using Milhouse.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Runtime Issues](#runtime-issues)
- [PRD State Issues](#prd-state-issues)
- [Configuration Issues](#configuration-issues)
- [Performance Issues](#performance-issues)
- [Getting Help](#getting-help)

## Installation Issues

### `mil: command not found`

**Cause:** Milhouse CLI is not in your PATH.

**Solutions:**

1. **If installed via Homebrew:**
   ```bash
   brew install daydemir/tap/mil
   ```

2. **If installed via Go:**
   ```bash
   go install github.com/daydemir/milhouse/cmd/mil@latest
   ```
   Ensure `$GOPATH/bin` is in your PATH:
   ```bash
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```

3. **Verify installation:**
   ```bash
   mil version
   ```

### Build from source fails

**Cause:** Missing Go toolchain or incompatible version.

**Solution:**
- Ensure Go 1.21+ is installed: `go version`
- Update Go if needed: https://go.dev/dl/
- Clone and build:
  ```bash
  git clone https://github.com/daydemir/milhouse.git
  cd milhouse
  make build
  ```

## Runtime Issues

### API Errors or Authentication Failures

**Symptoms:**
- "API key not found"
- "Authentication failed"
- "Invalid credentials"

**Solutions:**

1. **Verify Claude Code CLI is installed:**
   ```bash
   claude-code --version
   ```

2. **Check authentication:**
   - Ensure you're logged into Claude Code
   - Milhouse uses the same credentials as `claude-code`

3. **Test Claude Code directly:**
   ```bash
   claude-code
   ```
   If this fails, fix Claude Code authentication first.

### Rate Limit Errors

**Symptoms:**
- "Rate limit exceeded"
- "Too many requests"

**Solutions:**

1. **Wait and retry:**
   - Claude API has rate limits
   - Wait 60 seconds and try again

2. **Reduce concurrent operations:**
   - Run fewer iterations: `mil run 1` instead of `mil run 10`
   - Increase delays between iterations (future feature)

3. **Check your Claude Code plan:**
   - Free tier has lower rate limits
   - Consider upgrading to Professional or Max plan

### Large Token Usage / High Costs

**Symptoms:**
- Unexpectedly high token consumption
- Higher costs than anticipated

**Solutions:**

1. **Check current configuration:**
   ```bash
   mil config show
   ```

2. **Reduce per-phase token limits:**
   ```bash
   mil config edit
   ```
   - Set lower `maxTokens` for each phase
   - Example: Planner: 40K, Builder: 60K, Reviewer: 40K

3. **Use cheaper models:**
   - Switch from `opus` to `sonnet` or `haiku`
   - Edit via `mil config edit` or use CLI flags:
     ```bash
     mil run 1 --builder-model sonnet
     ```

4. **Start with fewer iterations:**
   - Test with `mil run 1` first
   - Verify behavior before running `mil run 5` or higher

### Unexpected Output or Poor Quality

**Symptoms:**
- Builder produces incorrect code
- Reviewer fails to catch issues
- Plans don't match requirements

**Solutions:**

1. **Use higher quality model:**
   ```bash
   mil config edit
   # Set model to 'opus' for critical phases
   ```

2. **Improve PRD clarity:**
   ```bash
   mil chat
   ```
   - Add more specific acceptance criteria
   - Include examples of expected behavior
   - Reference existing code patterns in `prompt.md`

3. **Check progress history:**
   ```bash
   cat .milhouse/progress.md
   ```
   - Look for patterns in failures
   - Verify agents have enough context

4. **Review plans before execution:**
   ```bash
   cat .milhouse/plans/<prd-id>.md
   ```
   - Ensure planner understood requirements
   - Manually edit plans if needed (advanced)

## PRD State Issues

### `.milhouse/` Directory Not Found

**Cause:** Project not initialized.

**Solution:**
```bash
cd your-project
mil init
```

### PRDs Stuck in "Pending" State

**Cause:** Reviewer failed to complete or crashed.

**Solutions:**

1. **Check reviewer output:**
   ```bash
   cat .milhouse/evidence/<prd-id>-review.md
   ```

2. **Check progress log:**
   ```bash
   tail -50 .milhouse/progress.md
   ```

3. **Manually update PRD state (advanced):**
   - Edit `.milhouse/prd.json`
   - Change `status` field to `"active"` or `"open"`
   - Run `mil run 1` to retry

### PRDs Stuck in "Active" State

**Cause:** Builder failed to complete or crashed.

**Solutions:**

1. **Check builder output:**
   ```bash
   cat .milhouse/evidence/<prd-id>-build.md
   ```

2. **Check for partial work:**
   - Look for uncommitted changes: `git status`
   - Review recent commits: `git log`

3. **Retry iteration:**
   ```bash
   mil run 1
   ```

### Unable to Load prd.json

**Symptoms:**
- "Failed to load PRDs"
- "Invalid JSON"

**Solutions:**

1. **Validate JSON syntax:**
   ```bash
   cat .milhouse/prd.json | jq .
   ```

2. **Restore from backup:**
   ```bash
   cp .milhouse/prd.json.backup .milhouse/prd.json
   ```

3. **Recreate from scratch:**
   ```bash
   rm .milhouse/prd.json
   mil chat
   # Re-add your PRDs
   ```

## Configuration Issues

### Invalid Configuration After Edit

**Symptoms:**
- "Invalid configuration"
- Validation errors after `mil config edit`

**Solutions:**

1. **Check validation constraints:**
   - Model: must be `haiku`, `sonnet`, or `opus`
   - MaxTokens: must be between 10,000 and 200,000
   - ProgressLines: must be between 1 and 500

2. **View current config:**
   ```bash
   mil config show
   ```

3. **Reset to defaults:**
   ```bash
   rm .milhouse/config.yaml
   mil config show  # Will use defaults
   ```

### Configuration Not Taking Effect

**Cause:** CLI flags override config file.

**Solution:**

1. **Remove CLI flags to use config file:**
   ```bash
   # Instead of:
   mil run 1 --builder-model opus

   # Use:
   mil config edit  # Set builder model to opus
   mil run 1
   ```

2. **Verify config is loaded:**
   ```bash
   mil config show
   ```

## Performance Issues

### Slow Execution

**Cause:** Large context, slow model, or API latency.

**Solutions:**

1. **Use faster model:**
   ```bash
   mil config edit
   # Set model to 'haiku' for faster responses
   ```

2. **Reduce context size:**
   - Edit `prompt.md` to be more concise
   - Reduce `progressLines` in config
   - Clean up old evidence files

3. **Check network connectivity:**
   - Ensure stable internet connection
   - Test with: `claude-code`

### Out of Memory Errors

**Cause:** Large files or excessive context.

**Solutions:**

1. **Add files to .gitignore:**
   - Exclude large generated files
   - Milhouse respects .gitignore

2. **Reduce context files:**
   - Edit `.milhouse/config.yaml`
   - Remove unnecessary `contextFiles`

3. **Increase system resources:**
   - Close other applications
   - Ensure sufficient RAM available

## Getting Help

If you've tried the solutions above and still have issues:

1. **Check existing issues:**
   - https://github.com/daydemir/milhouse/issues
   - Search for your error message

2. **Open a new issue:**
   - Include: `mil version` output
   - Include: Relevant error messages
   - Include: Steps to reproduce
   - Include: Configuration (sanitize API keys!)

3. **Community support:**
   - Tag `@daydemir` in GitHub issues
   - Include context about your use case

## Debug Mode

For detailed debugging information:

```bash
# Enable verbose output (future feature)
export MILHOUSE_DEBUG=1
mil run 1
```

Check logs:
```bash
# Review progress for detailed agent output
cat .milhouse/progress.md

# Review individual phase outputs
cat .milhouse/plans/<prd-id>.md
cat .milhouse/evidence/<prd-id>-build.md
cat .milhouse/evidence/<prd-id>-review.md
```
