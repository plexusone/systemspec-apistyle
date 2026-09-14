# AI Assistant Hooks

Configure automatic API linting when working with AI coding assistants and git pre-commit hooks.

## Overview

Hooks enable automatic API style checking in two contexts:

1. **AI Assistants** - Auto-lint when you save OpenAPI files in Claude Code, Cursor, or Windsurf
2. **Git Pre-Commit** - Block commits that contain API style violations

## Supported Assistants

| Assistant | Config File | Events |
|-----------|-------------|--------|
| Claude Code | `.claude/settings.json` | 16 events |
| Cursor | `.cursor/hooks.json` | 12 events |
| Windsurf | `.windsurf/hooks.json` | 9 events |

## Quick Setup

```bash
# Generate hooks for Claude Code
api-style hooks --format claude

# Generate for all assistants
api-style hooks --format all

# See supported formats
api-style hooks list

# Install git pre-commit hook
api-style hooks init
```

## Hook Types

### Auto-Lint on Save

Automatically lint OpenAPI files when they're saved:

```bash
api-style hooks --format claude --auto-lint
```

This creates a hook that:

1. Triggers when files are written (Edit/Write tools)
2. Checks if the file matches OpenAPI patterns
3. Runs `mcp-api-style lint` if it matches
4. Returns violations to the assistant

**Matched Patterns:**

- `openapi.yaml`, `openapi.yml`, `openapi.json`
- `swagger.yaml`, `swagger.yml`, `swagger.json`
- `**/openapi.yaml`, `**/api.yaml`, etc.

### Context Injection

Inject API style guidelines before prompts:

```bash
api-style hooks --format claude --inject-context
```

This adds context about the active style profile to help the AI assistant write better APIs.

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `--format` | `claude` | Target assistant |
| `--profile` | `default` | Style profile for linting |
| `--auto-lint` | `true` | Enable auto-lint on save |
| `--inject-context` | `false` | Inject style context |
| `--output` | auto | Output file path |

## Generated Configuration

### Claude Code

`.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "#!/bin/bash\nFILE=\"$CLAUDE_FILE_PATH\"\n...",
            "timeout": 30
          }
        ]
      }
    ]
  }
}
```

### Cursor

`.cursor/hooks.json`:

```json
{
  "hooks": {
    "after_file_write": [
      {
        "pattern": "Write|Edit",
        "actions": [
          {
            "type": "shell",
            "command": "mcp-api-style lint ...",
            "timeout": 30000
          }
        ]
      }
    ]
  }
}
```

### Windsurf

`.windsurf/hooks.json`:

```json
{
  "hooks": {
    "onFileSave": [
      {
        "glob": "**/*.yaml",
        "command": "mcp-api-style lint ..."
      }
    ]
  }
}
```

## Prerequisites

For hooks to work, the `mcp-api-style` binary must be in your PATH:

```bash
go install github.com/plexusone/systemspec-apistyle/cmd/mcp-api-style@latest
```

## Manual Configuration

### Claude Code

Add to `.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "mcp-api-style lint -f \"$CLAUDE_FILE_PATH\" -p default 2>&1 | head -50",
            "timeout": 30
          }
        ]
      }
    ]
  }
}
```

### Using with MCP Server

For full integration, combine hooks with the MCP server:

1. Configure MCP server in settings
2. Generate hooks for auto-linting
3. AI assistant can use both automatic checks and on-demand tools

## Workflow Example

1. **Setup:**
   ```bash
   api-style hooks --format claude
   ```

2. **Working with Claude:**
   - Ask Claude to create an OpenAPI spec
   - When Claude writes the file, hooks run automatically
   - Violations appear in the conversation
   - Claude can fix issues immediately

3. **On-Demand Tools:**
   - "Use the lint tool to check this against Azure guidelines"
   - "Explain rule URI-001"
   - "Run analyze for a GO/NO-GO decision"

## Troubleshooting

### Hooks Not Triggering

1. Verify the config file exists in the correct location
2. Check that `mcp-api-style` is in PATH
3. Ensure the file matches OpenAPI patterns

### Timeout Issues

Increase the timeout in the generated config:

```json
{
  "timeout": 60
}
```

### Profile Not Found

Ensure the profile name is valid:

```bash
mcp-api-style list-profiles
```

## Git Pre-Commit Hook

Install a git pre-commit hook that lints staged OpenAPI files before each commit:

```bash
api-style hooks init
```

### How It Works

1. When you run `git commit`, the hook triggers
2. It finds staged files matching OpenAPI patterns (`openapi.yaml`, `swagger.json`, etc.)
3. Each file is linted using `api-style lint`
4. If errors are found, the commit is blocked
5. Warnings are displayed but don't block commits

### Options

```bash
# Use Azure profile
api-style hooks init --profile azure

# Enforce silver conformance level
api-style hooks init --level silver

# Overwrite existing hook
api-style hooks init --force
```

### Bypassing the Hook

For emergency commits, bypass the hook with:

```bash
git commit --no-verify
```

### Uninstalling

Remove the hook manually:

```bash
rm .git/hooks/pre-commit
```

## Next Steps

- [MCP Server](../guide/mcp-server.md) - Full MCP integration
- [CLI Reference](cli.md) - Command documentation
