# Product Requirements Document (PRD)

## systemspec-apistyle

**Version:** 0.1.0-draft
**Date:** 2026-06-03
**Status:** Draft

## Overview

systemspec-apistyle is a machine-readable specification format for API style guides that generates human documentation, linting rules, LLM evaluation rubrics, and AI agent instructions from a single source of truth.

## Goals

1. **Unify** - One specification generates all artifacts
2. **Automate** - Deterministic + LLM evaluation in CI/CD
3. **Educate** - Clear documentation with examples and rationale
4. **Integrate** - Works with existing tools (Spectral, vacuum) and AI agents

## Non-Goals

1. API gateway runtime enforcement (separate concern)
2. API mocking or testing (separate tools exist)
3. API versioning/changelog (see Optic, openapi-diff)
4. Custom LLM training (use existing models)

## User Stories

### Platform Engineer

> As a platform engineer, I want to define our API standards once and generate linting rules, documentation, and agent instructions, so that enforcement is consistent across all teams and tools.

**Acceptance Criteria:**

- Define rules in JSON with schema validation
- Generate Spectral-compatible YAML
- Generate Markdown documentation
- Generate Claude Code/Kiro agent files

### API Designer

> As an API designer, I want to evaluate my OpenAPI spec against our style guide before review, so that I can fix issues early and learn best practices.

**Acceptance Criteria:**

- CLI command: `api-style analyze openapi.yaml`
- Combined lint + LLM evaluation
- Clear violation messages with locations
- Remediation suggestions

### API Reviewer

> As an API reviewer, I want AI assistance in reviewing specs against our guidelines, so that I can focus on architectural decisions rather than style nitpicks.

**Acceptance Criteria:**

- Claude Code agent that reviews specs
- Structured findings with severity
- Links to relevant guideline sections
- Prioritized remediation list

### Governance Lead

> As a governance lead, I want to track API quality trends across teams, so that I can identify training needs and measure improvement.

**Acceptance Criteria:**

- Historical evaluation storage
- Dashboard with trends
- Export to DevOps tools
- Conformance level reporting

## Features

### F1: Core Specification Format

**Priority:** P0 (Phase 1)

The systemspec-apistyle JSON format defines API style rules with:

```json
{
  "$schema": "https://plexusone.dev/systemspec-apistyle/schema/v1/api-style-spec.schema.json",
  "version": "1.0.0",
  "name": "acme-api-style",
  "extends": ["azure-v1"],

  "rules": [
    {
      "id": "URI-001",
      "title": "Use plural resource names",
      "category": "uri-design",
      "severity": "error",
      "scope": "path",

      "rationale": "Plural resources improve consistency and discoverability.",

      "examples": {
        "good": ["/users", "/orders/{orderId}/items"],
        "bad": ["/user", "/order/{orderId}/item"]
      },

      "enforcement": {
        "type": "spectral",
        "function": "pattern",
        "options": {
          "match": "^/[a-z]+s(/|$)"
        }
      },

      "judge": {
        "prompt": "Check if all resource names in paths use plural form.",
        "weight": 1.0
      },

      "references": [
        {"title": "Microsoft REST Guidelines", "url": "https://..."}
      ]
    }
  ],

  "lexicon": {
    "approved": ["customer", "order", "product"],
    "forbidden": [
      {"term": "user", "replaceWith": "customer", "reason": "Domain terminology"}
    ]
  },

  "conformanceLevels": {
    "bronze": {"requiredCategories": ["uri-design", "http-methods"]},
    "silver": {"requiredCategories": ["bronze", "naming", "errors"]},
    "gold": {"requiredCategories": ["silver", "security", "documentation"]}
  }
}
```

### F2: Deterministic Linting (vacuum)

**Priority:** P0 (Phase 1)

```bash
# Lint with default profile
api-style lint openapi.yaml

# Lint with specific profile and level
api-style lint openapi.yaml --profile azure --level silver

# Output formats
api-style lint openapi.yaml --format json --output report.json
api-style lint openapi.yaml --format sarif --output results.sarif
```

**Output:**

```json
{
  "status": "fail",
  "conformanceLevel": "bronze",
  "summary": {
    "errors": 3,
    "warnings": 7,
    "infos": 2
  },
  "violations": [
    {
      "ruleId": "URI-001",
      "severity": "error",
      "message": "Resource name 'user' should be plural 'users'",
      "path": "$.paths./user",
      "line": 42
    }
  ]
}
```

### F3: LLM Evaluation

**Priority:** P0 (Phase 2)

```bash
# Evaluate with LLM
api-style evaluate openapi.yaml --judge claude-sonnet

# Evaluate specific categories
api-style evaluate openapi.yaml --categories naming,semantics
```

**Output (structured-evaluation format):**

```json
{
  "type": "rubric",
  "document": "openapi.yaml",
  "categories": [
    {
      "name": "naming-conventions",
      "score": "partial",
      "reasoning": "Most resources use plural names, but 'user' endpoint is singular.",
      "findings": [
        {
          "severity": "medium",
          "title": "Singular resource name",
          "description": "The /user endpoint uses singular naming.",
          "path": "$.paths./user",
          "recommendation": "Rename to /users for consistency."
        }
      ]
    }
  ],
  "decision": "conditional",
  "nextSteps": {
    "immediate": ["Fix singular resource names"],
    "recommended": ["Add operation summaries"]
  }
}
```

### F4: Combined Analysis

**Priority:** P0 (Phase 2)

```bash
# Full analysis (lint + evaluate)
api-style analyze openapi.yaml --profile azure --level silver

# Output to directory
api-style analyze openapi.yaml --output results/
```

Generates:

- `results/lint-report.json` - Deterministic findings
- `results/evaluation-report.json` - LLM findings
- `results/summary-report.json` - Combined GO/NO-GO

### F5: Documentation Generation

**Priority:** P1 (Phase 2)

```bash
# Generate human-readable guide
api-style generate guide --spec acme.api-style.json --output docs/

# Generate with specific format
api-style generate guide --format markdown --output api-style-guide.md
api-style generate guide --format html --output api-style-guide.html
```

**Output Structure:**

```markdown
# ACME API Style Guide

## Quick Reference

| Category | Rules | Severity |
|----------|-------|----------|
| URI Design | 12 | 4 error, 8 warn |
| Naming | 8 | 2 error, 6 warn |
...

## URI Design

### URI-001: Use Plural Resource Names

**Severity:** Error

Plural resources improve consistency and discoverability.

**Good:**
- `/users`
- `/orders/{orderId}/items`

**Bad:**
- `/user`
- `/order/{orderId}/item`

**References:**
- [Microsoft REST Guidelines](https://...)
```

### F6: Spectral Generation

**Priority:** P1 (Phase 2)

```bash
# Generate Spectral ruleset
api-style generate spectral --spec acme.api-style.json --output .spectral.yaml
```

**Output:**

```yaml
extends: [[spectral:oas, recommended]]

rules:
  uri-001-plural-resources:
    description: Resource names should be plural
    severity: error
    given: $.paths
    then:
      function: pattern
      functionOptions:
        match: "^/[a-z]+s(/|$)"
```

### F7: Agent Generation

**Priority:** P1 (Phase 3)

```bash
# Generate Claude Code agent
api-style generate agent --platform claude-code --output agents/

# Generate Kiro agent
api-style generate agent --platform kiro-cli --output agents/
```

**Output (Claude Code):**

```markdown
---
name: api-style-reviewer
description: Reviews OpenAPI specs against ACME API style guide
model: sonnet
tools: [Read, Bash]
---

You are an API style reviewer for ACME Corp.

## Style Rules

{Generated from systemspec-apistyle rules}

## Evaluation Process

1. Read the OpenAPI specification
2. Run `api-style lint <path>` for deterministic checks
3. Evaluate semantic aspects against the rules below
4. Generate structured findings report
```

### F8: Web UI

**Priority:** P1 (Phase 3)

**Features:**

1. **Evaluate Tab**
   - URL inputs for: OpenAPI spec, Style spec, Spectral rules
   - "Evaluate" button
   - Results display with violations
   - Export to JSON/PDF

2. **Style Guide Tab**
   - URL input for Style spec
   - Rendered Markdown preview
   - Download as Markdown/PDF

3. **Rules Browser Tab**
   - Search/filter rules
   - Category grouping
   - Rule details with examples

### F9: MCP Server

**Priority:** P1 (Phase 3)

**Tools:**

| Tool | Description |
|------|-------------|
| `api_style_lint` | Lint OpenAPI spec |
| `api_style_evaluate` | LLM evaluation |
| `api_style_analyze` | Combined analysis |
| `api_style_explain_rule` | Get rule details |
| `api_style_list_rules` | List available rules |
| `api_style_generate_guide` | Generate documentation |

### F10: Profiles

**Priority:** P1 (Phase 2)

Pre-built profiles based on industry standards:

| Profile | Based On | Rules |
|---------|----------|-------|
| `default` | Common best practices | ~50 |
| `azure` | Microsoft/Azure guidelines | ~80 |
| `google` | Google API Design Guide | ~60 |
| `zalando` | Zalando RESTful Guidelines | ~70 |

```bash
# Use built-in profile
api-style lint openapi.yaml --profile azure

# Extend profile
{
  "extends": ["azure"],
  "rules": [
    {"id": "CUSTOM-001", ...}
  ],
  "overrides": {
    "URI-001": {"severity": "warn"}
  }
}
```

### F11: Conformance Levels

**Priority:** P2 (Phase 4)

Graduated compliance similar to WCAG:

```json
{
  "conformanceLevels": {
    "bronze": {
      "description": "Minimum viable API",
      "requiredCategories": ["uri-design", "http-methods"],
      "maxErrors": 0,
      "maxWarnings": 10
    },
    "silver": {
      "description": "Production-ready API",
      "requiredCategories": ["bronze", "naming", "errors", "pagination"],
      "maxErrors": 0,
      "maxWarnings": 5
    },
    "gold": {
      "description": "Exemplary API",
      "requiredCategories": ["silver", "security", "documentation", "versioning"],
      "maxErrors": 0,
      "maxWarnings": 0
    }
  }
}
```

### F12: Exception Registry

**Priority:** P2 (Phase 4)

```json
{
  "exceptions": [
    {
      "id": "EX-2026-001",
      "ruleId": "URI-001",
      "appliesTo": {
        "path": "/legacy/user"
      },
      "reason": "Legacy public contract, cannot change",
      "approvedBy": "api-review-board",
      "expiresOn": "2027-12-31"
    }
  ]
}
```

### F13: Compatibility/Breaking Change Detection

**Priority:** P2 (Phase 4)

```bash
# Compare specs for breaking changes
api-style diff old.yaml new.yaml --fail-on-breaking

# Output
{
  "compatible": false,
  "breakingChanges": [
    {
      "type": "removed-endpoint",
      "path": "/users/{id}",
      "method": "DELETE"
    },
    {
      "type": "required-parameter-added",
      "path": "/orders",
      "parameter": "customerId"
    }
  ]
}
```

### F14: Observability

**Priority:** P2 (Phase 4)

**Local Storage:**

```bash
# Store evaluation result
api-style analyze openapi.yaml --store

# Query history
api-style history --api orders-api --since 2026-01-01
```

**DevOps Export:**

```bash
# Export metrics to Datadog
api-style analyze openapi.yaml --export datadog

# Metrics exported:
# - api_style.lint.errors{api=orders,profile=azure}
# - api_style.lint.warnings{api=orders,profile=azure}
# - api_style.evaluation.decision{api=orders,decision=pass}
# - api_style.conformance.level{api=orders,level=silver}
```

## User Interface

### CLI

```
api-style
├── lint          # Deterministic linting
├── evaluate      # LLM evaluation
├── analyze       # Combined lint + evaluate
├── diff          # Breaking change detection
├── generate
│   ├── guide     # Human documentation
│   ├── spectral  # Spectral ruleset
│   └── agent     # AI agent files
├── serve
│   ├── web       # Web UI server
│   └── mcp       # MCP server
├── profile
│   ├── list      # List available profiles
│   ├── show      # Show profile details
│   └── validate  # Validate custom profile
└── history       # Query evaluation history
```

### Web UI

Three-tab interface:

1. **Evaluate** - Input URLs, run analysis, view results
2. **Style Guide** - View rendered documentation, export
3. **Rules** - Browse and search rules

### MCP

Standard MCP server with tools for IDE integration.

## Technical Constraints

1. **Performance:** Lint <500ms for 1MB spec
2. **Privacy:** Web UI processes client-side only (no server storage)
3. **Compatibility:** Consume existing Spectral rulesets
4. **Cost:** LLM evaluation <$0.10 per spec

## Dependencies

See [CONSTITUTION.md](../CONSTITUTION.md) for technology standards.

## Success Criteria

| Metric | Target |
|--------|--------|
| Phase 1 complete | Core types, vacuum integration, basic CLI |
| Phase 2 complete | LLM evaluation, profiles, documentation generation |
| Phase 3 complete | Web UI, MCP server, agent generation |
| Phase 4 complete | Observability, conformance levels, breaking change detection |
