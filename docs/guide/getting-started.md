# Getting Started

This guide walks you through installing systemspec-apistyle and linting your first OpenAPI specification.

## Installation

### From Source (Go)

```bash
go install github.com/plexusone/systemspec-apistyle/cmd/api-style@latest
```

### Verify Installation

```bash
api-style version
```

## Your First Lint

Create a simple OpenAPI specification:

```yaml
# openapi.yaml
openapi: "3.1.0"
info:
  title: My API
  version: "1.0.0"
paths:
  /user:
    get:
      responses:
        "200":
          description: OK
```

Run the linter:

```bash
api-style lint openapi.yaml
```

You'll see output like:

```
API Style Lint Report
=====================

Status: FAIL
Errors: 1, Warnings: 2, Info: 0, Hints: 0

Errors:
  - [URI-001] Use plural resource names
    $.paths['/user']

Warnings:
  - [operation-description] operation method `GET` is missing a description
    $.paths['/user'].get
```

## Understanding Violations

Each violation includes:

| Field | Description |
|-------|-------------|
| Rule ID | Unique identifier (e.g., `URI-001`) |
| Message | Human-readable explanation |
| Path | JSON path to the violation location |
| Severity | `error`, `warn`, `info`, or `hint` |

## Fixing Violations

Update your spec to fix the violations:

```yaml
# openapi.yaml
openapi: "3.1.0"
info:
  title: My API
  version: "1.0.0"
  description: A sample API
paths:
  /users:  # Changed from /user to /users (plural)
    get:
      summary: List users
      description: Returns a list of all users
      operationId: listUsers
      responses:
        "200":
          description: OK
```

Run again:

```bash
api-style lint openapi.yaml
```

```
API Style Lint Report
=====================

Status: PASS
Errors: 0, Warnings: 0, Info: 0, Hints: 0

No violations found.
```

## Using Different Profiles

systemspec-apistyle includes profiles based on industry guidelines:

```bash
# Use Azure/Microsoft guidelines
api-style lint openapi.yaml --profile azure

# Use Google API Design Guide
api-style lint openapi.yaml --profile google

# Use Zalando RESTful Guidelines
api-style lint openapi.yaml --profile zalando
```

See [Using Profiles](profiles.md) for details on each profile.

## Output Formats

### Text (default)

```bash
api-style lint openapi.yaml
```

### JSON

```bash
api-style lint openapi.yaml --format json
```

### SARIF (for IDE integration)

```bash
api-style lint openapi.yaml --format sarif --output report.sarif
```

## Combined Analysis

For comprehensive review including LLM-based semantic analysis:

```bash
# Requires ANTHROPIC_API_KEY
export ANTHROPIC_API_KEY=sk-ant-...
api-style analyze openapi.yaml --profile azure
```

This combines:

1. Deterministic linting (fast, CI-friendly)
2. LLM evaluation (semantic analysis of design quality)
3. GO/NO-GO decision for production readiness

## Configuration File

Create `.api-style.yaml` in your project root to configure default options:

```yaml
# .api-style.yaml
profile: azure
level: silver

include:
  - "openapi.yaml"
  - "**/api.yaml"

exclude:
  - "**/generated/**"

exceptions:
  - rule: URI-001
    paths: ["/legacy/**"]
    reason: "Legacy API cannot be changed"
```

CLI flags override config file settings. See the [Configuration Reference](../reference/config.md) for all options.

## Multi-File Linting

Lint multiple files at once:

```bash
# Glob patterns
api-style lint api/*.yaml

# Recursive directory search
api-style lint . --recursive

# Multiple arguments
api-style lint openapi.yaml swagger.yaml
```

## Watch Mode

Continuously lint as you edit:

```bash
api-style lint openapi.yaml --watch
```

The watcher:

- Runs initial lint on startup
- Re-lints when files change
- Shows timestamps for each run
- Exits cleanly with Ctrl+C

## Git Pre-Commit Hook

Block commits with API style violations:

```bash
api-style hooks init
```

This installs a pre-commit hook that lints staged OpenAPI files. See [Hooks Reference](../reference/hooks.md) for details.

## Starting from Exemplars

Use exemplar specifications as templates for new APIs:

```bash
# List available exemplars
api-style exemplar list

# View an exemplar
api-style exemplar show default-minimal

# Copy as starting point
api-style exemplar copy default-minimal ./my-api.yaml
```

Exemplars demonstrate best practices for each style profile.

## Using Design Patterns

Explore reusable solutions for common API design problems:

```bash
# List patterns for a profile
api-style pattern list --profile default

# View pattern details with examples
api-style pattern show cursor-pagination
```

Patterns include problem descriptions, solutions, code examples, and related rules.

## Getting Fix Suggestions

Get AI-powered suggestions to fix violations:

```bash
# Include suggestions in lint output
api-style lint openapi.yaml --suggest-fixes

# Or generate suggestions from violations
api-style lint openapi.yaml --format json --output violations.json
api-style suggest-fixes violations.json
```

Suggestions include the recommended fix, reasoning, and JSON Patch operations for automated fixing.

## Next Steps

- [Automated API Governance](automated-api-governance.md) - Set up AI-first API design with 100% automated enforcement
- [Using Profiles](profiles.md) - Understand built-in style profiles
- [Documentation Generation](documentation-generation.md) - Generate Markdown and MkDocs documentation
- [Custom Rules](custom-rules.md) - Create your own rules
- [MCP Server](mcp-server.md) - AI agent integration
- [CLI Reference](../reference/cli.md) - Complete command documentation
