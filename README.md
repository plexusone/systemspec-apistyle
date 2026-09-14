# SystemSpec-APIStyle

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/systemspec-apistyle/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/systemspec-apistyle/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/systemspec-apistyle/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/systemspec-apistyle/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/systemspec-apistyle/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/systemspec-apistyle/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/systemspec-apistyle
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/systemspec-apistyle
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/systemspec-apistyle
 [viz-svg]: https://img.shields.io/badge/Go-visualizaton-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fsystemspec-apistyle
 [loc-svg]: https://tokei.rs/b1/github/plexusone/systemspec-apistyle
 [repo-url]: https://github.com/plexusone/systemspec-apistyle
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/systemspec-apistyle/blob/main/LICENSE

Machine-readable API style specification format that generates human documentation, linting rules, LLM evaluation rubrics, and AI agent instructions from a single source of truth.

> **Note:** This project is part of the PlexusOne `systemspec-{domain}` family of machine-readable system definitions (alongside `systemspec-designsystem`, `systemspec-architecture`, `systemspec-crypto`, and `systemspec-deploy`). It was formerly named `api-style-spec` and was renamed in September 2026; releases up to v0.5.0 were published under the old module path `github.com/plexusone/api-style-spec`.

## Overview

API style guides from Microsoft, Google, and Zalando have become industry standards, but they exist only as human-readable documents. **systemspec-apistyle** creates a machine-readable specification format that serves as the canonical source, generating all artifacts from one definition.

```
systemspec-apistyle (source of truth)
    ├── Human Style Guide (Markdown)
    ├── Deterministic Linters (Spectral/vacuum)
    ├── LLM Review Rubrics
    ├── AI Agent Instructions (Claude Code, Kiro)
    └── MCP Server Tools
```

## Features

- 📋 **Unified Specification** - Define rules once, generate all artifacts
- ✅ **Deterministic Linting** - Fast, CI-friendly checks via [vacuum](https://github.com/daveshanley/vacuum)
- 🧠 **LLM Evaluation** - Semantic analysis for rules that can't be linted
- 🏢 **Industry Profiles** - Pre-built profiles based on Microsoft, Google, Zalando guidelines
- 🏆 **Conformance Levels** - Graduated compliance (bronze/silver/gold)
- 🌐 **Multi-Platform** - CLI, Web UI, MCP server, AI agents
- 📖 **Exemplar Specs** - Reference OpenAPI specifications demonstrating best practices
- 🧩 **Pattern Library** - Reusable solutions for common API design problems
- 🔧 **Fix Suggestions** - AI-powered suggestions to fix style violations

## Installation

```bash
go install github.com/plexusone/systemspec-apistyle/cmd/api-style@latest
```

## Quick Start

```bash
# Lint an OpenAPI specification
api-style lint openapi.yaml

# Lint with a specific profile
api-style lint openapi.yaml --profile azure

# Lint with fix suggestions
api-style lint openapi.yaml --suggest-fixes

# Lint multiple files with glob patterns
api-style lint api/*.yaml --recursive

# Watch mode for continuous linting
api-style lint openapi.yaml --watch

# Combined lint + LLM evaluation
api-style analyze openapi.yaml --profile azure --level silver

# Generate human-readable style guide
api-style generate guide --spec my-style.json --output docs/

# View exemplar specifications
api-style exemplar list
api-style exemplar show default-minimal
api-style exemplar copy default-minimal ./my-api.yaml

# Explore design patterns
api-style pattern list
api-style pattern show cursor-pagination

# Get fix suggestions for violations
api-style suggest-fixes violations.json --profile default
```

## Configuration

Create `.api-style.yaml` in your project root:

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

severity-overrides:
  URI-002: warn
```

See [.api-style.yaml.example](.api-style.yaml.example) for a complete example.

## Specification Format

```json
{
  "$schema": "https://plexusone.dev/systemspec-apistyle/schema/v1/api-style-spec.schema.json",
  "version": "1.0.0",
  "name": "my-api-style",
  "extends": ["default"],

  "rules": [
    {
      "id": "URI-001",
      "title": "Use plural resource names",
      "category": "uri-design",
      "severity": "error",
      "rationale": "Plural resources improve consistency.",
      "examples": {
        "good": ["/users", "/orders"],
        "bad": ["/user", "/order"]
      },
      "enforcement": {
        "type": "spectral",
        "function": "pattern",
        "options": {"match": "^/[a-z]+s(/|$)"}
      }
    }
  ]
}
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `api-style lint` | Deterministic linting (supports glob patterns, `--watch`, `--recursive`, `--suggest-fixes`) |
| `api-style evaluate` | LLM-based evaluation |
| `api-style analyze` | Combined lint + evaluate |
| `api-style suggest-fixes` | Generate fix suggestions for violations |
| `api-style exemplar list` | List available exemplar specifications |
| `api-style exemplar show` | Display an exemplar specification |
| `api-style exemplar copy` | Copy an exemplar to a local file |
| `api-style pattern list` | List available design patterns |
| `api-style pattern show` | Display pattern details with examples |
| `api-style score-profile` | Score a style profile using LLM evaluation |
| `api-style generate guide` | Generate Markdown documentation |
| `api-style generate mkdocs` | Generate MkDocs multi-page site |
| `api-style generate spectral` | Generate Spectral ruleset |
| `api-style generate rubric` | Generate LLM evaluation rubric |
| `api-style hooks` | Generate AI assistant hooks |
| `api-style hooks init` | Install git pre-commit hook |
| `api-style diff` | Breaking change detection |
| `api-style serve mcp` | Start MCP server |
| `api-style serve web` | Start Web UI |

## Built-in Profiles

| Profile | Rules | Categories | Focus |
|---------|-------|------------|-------|
| `default` | 106 | 27 | Industry-leading, SDK-optimized (ogen) |
| `comprehensive` | 88 | 26 | Full coverage, all best practices |
| `zalando` | 147 | 13 | E-commerce, events |
| `microsoft-rest` | 123 | 15 | Enterprise REST APIs |
| `microsoft-graph` | 82 | 12 | OData/Graph APIs |
| `azure` | 23 | 9 | Azure cloud services |
| `google` | 20 | 7 | Resource-oriented design |
| `minimal` | 29 | 7 | Basic API hygiene |

**Default Profile Highlights:**

- 100% evaluable with LLM-as-Judge criteria
- 34% deterministic Spectral enforcement
- SDK-optimized for ogen, openapi-generator
- Multi-tenancy patterns with `~` alias
- RFC 9457 Problem Details for errors
- Discriminated unions for polymorphism

## Exemplar Specifications

Exemplars are reference OpenAPI specifications that demonstrate best practices for a style profile. Use them as starting points or learning resources.

```bash
# List all exemplars
api-style exemplar list

# Show a specific exemplar
api-style exemplar show default-minimal

# Copy to local file as starting point
api-style exemplar copy default-minimal ./my-api.yaml
```

| Exemplar | Profile | Description |
|----------|---------|-------------|
| `default-minimal` | default | Minimal CRUD API demonstrating core patterns |
| `default-comprehensive` | default | Full-featured API with pagination, errors, versioning |

## Pattern Library

Design patterns are reusable solutions to common API design problems. Each pattern includes problem/solution descriptions, code examples, and related rules.

```bash
# List patterns for a profile
api-style pattern list --profile default

# Show pattern details
api-style pattern show cursor-pagination
```

| Pattern | Category | Description |
|---------|----------|-------------|
| `cursor-pagination` | pagination | Cursor-based pagination for large datasets |
| `rfc9457-errors` | errors | RFC 9457 Problem Details for error responses |
| `discriminated-unions` | schemas | Type-safe polymorphism with discriminator fields |

## MCP Server Resources

The MCP server exposes API style resources for AI agents:

| Resource URI | Description |
|--------------|-------------|
| `apistyle://profiles` | List available profiles |
| `apistyle://profile/{name}` | Get profile specification |
| `apistyle://exemplars` | List available exemplars |
| `apistyle://exemplar/{name}` | Get exemplar content |
| `apistyle://patterns/{profile}` | List patterns for a profile |
| `apistyle://pattern/{profile}/{id}` | Get pattern definition |
| `apistyle://rubric/{profile}/{mode}` | Get evaluation/generation rubric |

## AI Agent Integration

systemspec-apistyle integrates with AI assistants for automated API design:

```bash
# Generate Claude Code hooks
api-style hooks --format claude-code > .claude/CLAUDE.md

# Install as pre-commit hook
api-style hooks init

# Start MCP server for AI agents
api-style serve mcp
```

AI agents can use MCP resources to:

- Access style profiles and rules during API generation
- Retrieve exemplar specs as reference implementations
- Look up design patterns for specific problems
- Get structured rubrics for self-evaluation

## Documentation

- [Getting Started](docs/guide/getting-started.md)
- [Automated API Governance](docs/guide/automated-api-governance.md) - AI-first API design workflow
- [CI/CD Integration](docs/guide/ci-cd.md) - Pipeline integration and pre-commit hooks
- [Documentation Generation](docs/guide/documentation-generation.md) - Generate Markdown and MkDocs sites
- [Creating Custom Profiles](docs/guide/profiles.md)
- [Writing Custom Rules](docs/guide/custom-rules.md)
- [MRD](docs/specs/inception/MRD.md) | [PRD](docs/specs/inception/PRD.md) | [TRD](docs/specs/inception/TRD.md) | [Roadmap](docs/specs/inception/ROADMAP.md)

## Related Projects

- [vacuum](https://github.com/daveshanley/vacuum) - Fast OpenAPI linter (used internally)
- [structured-evaluation](https://github.com/plexusone/structured-evaluation) - LLM evaluation framework
- [multi-agent-spec](https://github.com/plexusone/multi-agent-spec) - Agent definitions
- [assistantkit](https://github.com/plexusone/assistantkit) - AI assistant file generation

## License

MIT
