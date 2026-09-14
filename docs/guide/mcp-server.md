# MCP Server

systemspec-apistyle provides an MCP (Model Context Protocol) server for integration with AI assistants like Claude.

## Overview

The MCP server exposes API style linting and evaluation as tools that AI assistants can use during API development and review.

## Starting the Server

```bash
# Start MCP server (stdio transport)
mcp-api-style

# Or explicitly
mcp-api-style serve
```

## Configuration

### Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "api-style": {
      "command": "mcp-api-style",
      "env": {
        "ANTHROPIC_API_KEY": "sk-ant-..."
      }
    }
  }
}
```

### Claude Code

Add to your project's `.claude/settings.json`:

```json
{
  "mcpServers": {
    "api-style": {
      "command": "mcp-api-style"
    }
  }
}
```

## Available Tools

### lint

Lint an OpenAPI specification against style rules.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `openapi_spec` | string | Yes | OpenAPI spec content (YAML or JSON) |
| `profile` | string | No | Style profile (default: `default`) |

**Example:**

```
Use the lint tool to check this OpenAPI spec against Azure guidelines:
[paste spec here]
```

### evaluate

LLM-based semantic evaluation of an API spec.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `openapi_spec` | string | Yes | OpenAPI spec content |
| `profile` | string | No | Style profile |
| `categories` | array | No | Categories to evaluate |

**Requires:** `ANTHROPIC_API_KEY` environment variable

### analyze

Combined lint + evaluate with GO/NO-GO decision.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `openapi_spec` | string | Yes | OpenAPI spec content |
| `profile` | string | No | Style profile |
| `lint_only` | boolean | No | Skip LLM evaluation |

### list_profiles

List available style profiles.

### list_rules

List rules from a profile.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `profile` | string | No | Style profile |
| `category` | string | No | Filter by category |
| `severity` | string | No | Filter by severity |

### explain_rule

Get detailed explanation of a rule.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `rule_id` | string | Yes | Rule ID (e.g., `URI-001`) |
| `profile` | string | No | Style profile |

## Resources

The MCP server provides resources for AI agents to access style data:

### Profile Resources

| Resource URI | Description |
|--------------|-------------|
| `apistyle://profiles` | List all available profiles with metadata |
| `apistyle://profile/{name}` | Get the full definition of a specific profile |

### Exemplar Resources

| Resource URI | Description |
|--------------|-------------|
| `apistyle://exemplars` | List all available exemplar specifications |
| `apistyle://exemplar/{name}` | Get exemplar OpenAPI content |

Exemplars are reference implementations that demonstrate best practices:

```
apistyle://exemplar/default-minimal
apistyle://exemplar/default-comprehensive
```

### Pattern Resources

| Resource URI | Description |
|--------------|-------------|
| `apistyle://patterns/{profile}` | List patterns for a profile |
| `apistyle://pattern/{profile}/{id}` | Get pattern definition with examples |

Patterns provide reusable solutions for common API design problems:

```
apistyle://pattern/default/cursor-pagination
apistyle://pattern/default/rfc9457-errors
apistyle://pattern/default/discriminated-unions
```

### Rubric Resources

| Resource URI | Description |
|--------------|-------------|
| `apistyle://rubric/{profile}/evaluation` | Get evaluation rubric for LLM-as-Judge |
| `apistyle://rubric/{profile}/generation` | Get generation rubric for API creation |

Use rubrics to guide AI agents during API design and review:

```
apistyle://rubric/default/evaluation
apistyle://rubric/default/generation
```

## Prompts

Pre-built prompt templates for common tasks:

### review_api

Comprehensive API review workflow:

1. Runs lint tool
2. Analyzes results
3. Provides recommendations
4. Gives GO/NO-GO assessment

### fix_violations

Suggests fixes for specific violations with before/after examples.

### explain_profile

Explains a profile's philosophy and key rules.

## CLI Usage

The MCP server binary also works as a standalone CLI:

```bash
# Lint a spec
mcp-api-style lint --file openapi.yaml --profile azure

# List profiles
mcp-api-style list-profiles

# Explain a rule
mcp-api-style explain-rule URI-001

# Analyze with GO/NO-GO decision
mcp-api-style analyze --file openapi.yaml --profile azure
```

## Example Workflow

When working with Claude on API design:

1. **Draft API**: Ask Claude to help design an API
2. **Lint Check**: "Use the lint tool to check this spec against Google guidelines"
3. **Fix Issues**: "Use explain_rule to understand URI-001 and suggest a fix"
4. **Final Review**: "Run analyze to get a GO/NO-GO decision"

## Next Steps

- [AI Assistant Hooks](../reference/hooks.md) - Auto-lint on file save
- [CLI Reference](../reference/cli.md) - Complete command documentation
