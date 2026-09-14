# systemspec-apistyle

Machine-readable API style specification format that generates human documentation, linting rules, LLM evaluation rubrics, and AI agent instructions from a single source of truth.

## Overview

API style guides from Microsoft, Google, and Zalando have become industry standards, but they exist only as human-readable documents. **systemspec-apistyle** creates a machine-readable specification format that serves as the canonical source, generating all artifacts from one definition.

```
systemspec-apistyle (source of truth)
    ├── Human Style Guide
    │   ├── Single-page Markdown
    │   └── MkDocs Multi-page Site
    ├── Deterministic Linters (Spectral/vacuum)
    ├── LLM Review Rubrics
    ├── AI Agent Instructions (Claude Code, Cursor)
    └── MCP Server Tools
```

## Features

- **Unified Specification** - Define rules once, generate all artifacts
- **Deterministic Linting** - Fast, CI-friendly checks via [vacuum](https://github.com/daveshanley/vacuum)
- **LLM Evaluation** - Semantic analysis for rules that can't be linted
- **Industry Profiles** - Pre-built profiles based on Microsoft, Google, Zalando guidelines
- **Conformance Levels** - Graduated compliance (bronze/silver/gold)
- **Multi-Platform** - CLI, MCP server, AI assistant hooks
- **Exemplar Specs** - Reference OpenAPI implementations demonstrating best practices
- **Pattern Library** - Reusable solutions for common API design problems
- **Fix Suggestions** - AI-powered recommendations to fix style violations

## Quick Start

### Installation

```bash
go install github.com/plexusone/systemspec-apistyle/cmd/api-style@latest
```

### Basic Usage

```bash
# Lint an OpenAPI specification
api-style lint openapi.yaml

# Lint with a specific profile
api-style lint openapi.yaml --profile azure

# Lint with fix suggestions
api-style lint openapi.yaml --suggest-fixes

# Combined lint + LLM evaluation
api-style analyze openapi.yaml --profile azure

# Start from an exemplar
api-style exemplar copy default-minimal ./my-api.yaml

# Explore design patterns
api-style pattern list
api-style pattern show cursor-pagination
```

## Built-in Profiles

| Profile | Based On | Focus |
|---------|----------|-------|
| [default](profiles/default.md) | Common best practices | General REST API design |
| [azure](profiles/azure.md) | Microsoft/Azure guidelines | Enterprise cloud APIs |
| [google](profiles/google.md) | Google API Design Guide | Resource-oriented design |
| [zalando](profiles/zalando.md) | Zalando RESTful Guidelines | E-commerce APIs |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    systemspec-apistyle                           │
├─────────────────────────────────────────────────────────────┤
│  Style Profile (JSON)                                       │
│  ├── Rules with enforcement config                          │
│  ├── LLM evaluation rubrics                                 │
│  └── Conformance levels                                     │
├─────────────────────────────────────────────────────────────┤
│  Outputs                                                    │
│  ├── Deterministic Lint (vacuum)                            │
│  ├── LLM Evaluation (Claude)                                │
│  ├── Markdown Documentation                                 │
│  ├── Spectral Ruleset                                       │
│  └── AI Agent Instructions                                  │
└─────────────────────────────────────────────────────────────┘
```

## Next Steps

- [Getting Started](guide/getting-started.md) - Installation and first steps
- [Using Profiles](guide/profiles.md) - Work with built-in style profiles
- [Documentation Generation](guide/documentation-generation.md) - Generate Markdown and MkDocs sites
- [CLI Reference](reference/cli.md) - Complete command documentation
- [MCP Server](guide/mcp-server.md) - Use with AI assistants
