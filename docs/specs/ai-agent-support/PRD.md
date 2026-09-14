# Product Requirements Document (PRD)

## AI Agent Support for systemspec-apistyle

**Version:** 0.5.0-draft
**Date:** 2026-07-14
**Status:** Draft

## Overview

This document defines product requirements for enhancing systemspec-apistyle to fully support AI agents in two key workflows:

1. **Design**: AI agents generating OpenAPI specs that conform to a style specification
2. **Review**: AI agents reviewing existing OpenAPI specs for conformance

## Goals

1. **Enable Generation** - Provide AI agents with templates, patterns, and guidance to generate conformant OpenAPI specs
2. **Improve Remediation** - Give agents actionable fix suggestions, not just violation reports
3. **Prioritize Fixes** - Help agents understand which rules matter most and in what order to fix them
4. **Support Iteration** - Enable feedback loops where agents can incrementally improve specs

## Non-Goals

1. Replacing human API design decisions (agents assist, humans decide)
2. Auto-fixing all violations without review (suggestions only)
3. Training custom LLMs (use existing models with better prompts)
4. Real-time collaborative editing (single-agent workflows)

## User Stories

### AI Agent as API Designer

> As an AI coding assistant (Claude Code, Cursor, Kiro), I want to receive generation guidance from the style spec so that I can create OpenAPI specs that conform from the start.

**Acceptance Criteria:**

- Access generation-focused prompts for each rule
- Get exemplar OpenAPI specs for each profile
- Understand rule priority for conformance levels
- Receive pattern templates for common API structures (CRUD, pagination, etc.)

### AI Agent as API Reviewer

> As an AI coding assistant, I want to receive actionable remediation guidance when reviewing OpenAPI specs so that I can suggest specific fixes to the user.

**Acceptance Criteria:**

- Get suggested fixes for each violation
- Receive before/after code examples
- Understand fix priority and dependencies
- Access rule documentation URLs
- Know confidence level for each finding

### Platform Engineer Enabling AI Agents

> As a platform engineer, I want to configure AI agent behavior via the style spec so that all agents follow our API governance standards.

**Acceptance Criteria:**

- Define generation templates in profiles
- Set rule priorities and dependencies
- Configure conformance path guidance
- Provide company-specific fix examples

## Features

### F1: Violation Remediation Enhancement

**Priority:** P0 (Critical)

Extend the `Violation` type to include actionable remediation information:

```go
type Violation struct {
    // Existing fields...
    RuleID   string   `json:"ruleId"`
    Severity Severity `json:"severity"`
    Message  string   `json:"message"`
    Path     string   `json:"path"`
    Line     int      `json:"line,omitempty"`

    // NEW: Remediation fields
    Suggestion  string  `json:"suggestion,omitempty"`  // Human-readable fix description
    ExampleFix  string  `json:"exampleFix,omitempty"` // Code snippet showing the fix
    RuleURL     string  `json:"ruleUrl,omitempty"`    // Link to rule documentation
    Confidence  float64 `json:"confidence,omitempty"` // 0.0-1.0 certainty score
    RelatedRules []string `json:"relatedRules,omitempty"` // Rules to fix first
}
```

**Use Case:**

```json
{
  "ruleId": "URI-001",
  "severity": "error",
  "message": "Resource name 'user' should be plural",
  "path": "$.paths./user",
  "line": 42,
  "suggestion": "Rename path from /user to /users",
  "exampleFix": "paths:\n  /users:  # was: /user\n    get:",
  "ruleUrl": "https://plexusone.dev/systemspec-apistyle/profiles/default#uri-001",
  "confidence": 1.0,
  "relatedRules": []
}
```

### F2: Generation Guidance in Rules

**Priority:** P0 (Critical)

Add generation-focused metadata to rules:

```json
{
  "id": "URI-001",
  "title": "Use plural resource names",
  "severity": "error",

  "generate": {
    "prompt": "When creating API endpoints, always use plural nouns for resource collections (e.g., /users, /orders, /products). Use singular only for singleton resources.",
    "template": "/{resources}",
    "priority": 100,
    "examples": [
      {
        "description": "User collection endpoint",
        "openapi": "paths:\n  /users:\n    get:\n      operationId: listUsers"
      }
    ]
  }
}
```

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `prompt` | string | Instruction for LLM during generation |
| `template` | string | URI or schema template pattern |
| `priority` | int | 1-100, higher = apply first |
| `examples` | array | OpenAPI snippets showing correct usage |

### F3: suggest_fixes MCP Tool

**Priority:** P0 (Critical)

New MCP tool that generates fix suggestions for violations:

**Tool Definition:**

```json
{
  "name": "suggest_fixes",
  "description": "Generate fix suggestions for OpenAPI spec violations",
  "inputSchema": {
    "type": "object",
    "properties": {
      "spec": {
        "type": "string",
        "description": "The OpenAPI specification content"
      },
      "violations": {
        "type": "array",
        "description": "List of violations to fix",
        "items": {
          "type": "object",
          "properties": {
            "ruleId": { "type": "string" },
            "path": { "type": "string" }
          }
        }
      },
      "profile": {
        "type": "string",
        "description": "Style profile to use for fix suggestions"
      }
    },
    "required": ["spec", "violations"]
  }
}
```

**Output:**

```json
{
  "suggestions": [
    {
      "ruleId": "URI-001",
      "path": "$.paths./user",
      "currentValue": "/user",
      "suggestedValue": "/users",
      "diff": "- /user\n+ /users",
      "confidence": 0.95,
      "reasoning": "Plural form 'users' is required for collection endpoints"
    }
  ],
  "patchOperations": [
    {
      "op": "move",
      "from": "/paths/~1user",
      "path": "/paths/~1users"
    }
  ]
}
```

### F4: design_check MCP Tool

**Priority:** P1 (High)

Pre-generation validation tool:

```json
{
  "name": "design_check",
  "description": "Get design guidance before generating an OpenAPI spec",
  "inputSchema": {
    "type": "object",
    "properties": {
      "resourceName": {
        "type": "string",
        "description": "The resource being designed (e.g., 'user', 'order')"
      },
      "operations": {
        "type": "array",
        "description": "Planned operations (e.g., ['list', 'create', 'get', 'update', 'delete'])"
      },
      "profile": {
        "type": "string",
        "description": "Style profile to check against"
      }
    },
    "required": ["resourceName", "profile"]
  }
}
```

**Output:**

```json
{
  "checklist": [
    {
      "ruleId": "URI-001",
      "instruction": "Use plural form: /users (not /user)",
      "priority": 100
    },
    {
      "ruleId": "HTTP-001",
      "instruction": "Use GET for list, POST for create",
      "priority": 95
    }
  ],
  "template": {
    "paths": {
      "/users": {
        "get": { "operationId": "listUsers" },
        "post": { "operationId": "createUser" }
      },
      "/users/{userId}": {
        "get": { "operationId": "getUser" },
        "put": { "operationId": "updateUser" },
        "delete": { "operationId": "deleteUser" }
      }
    }
  }
}
```

### F5: conformance_path MCP Tool

**Priority:** P1 (High)

Shows the path to achieve a conformance level:

```json
{
  "name": "conformance_path",
  "description": "Get the path to reach a conformance level",
  "inputSchema": {
    "type": "object",
    "properties": {
      "spec": {
        "type": "string",
        "description": "The OpenAPI specification content"
      },
      "targetLevel": {
        "type": "string",
        "enum": ["bronze", "silver", "gold"],
        "description": "Target conformance level"
      },
      "profile": {
        "type": "string",
        "description": "Style profile to evaluate against"
      }
    },
    "required": ["spec", "targetLevel", "profile"]
  }
}
```

**Output:**

```json
{
  "currentLevel": "none",
  "targetLevel": "bronze",
  "blockers": [
    {
      "ruleId": "URI-001",
      "count": 3,
      "priority": 1,
      "fixInstructions": "Rename /user to /users, /order to /orders, /product to /products"
    }
  ],
  "warnings": [
    {
      "ruleId": "DOC-001",
      "count": 5,
      "priority": 2,
      "fixInstructions": "Add descriptions to 5 operations"
    }
  ],
  "progressToTarget": 0.6,
  "estimatedFixes": 8
}
```

### F6: Exemplar Specs per Profile

**Priority:** P1 (High)

Each profile includes reference OpenAPI specs:

```
profiles/
├── default/
│   ├── profile.json
│   └── exemplars/
│       ├── minimal.yaml      # Minimal conformant spec
│       ├── crud-api.yaml     # Basic CRUD API
│       └── full-api.yaml     # Comprehensive example
├── azure/
│   └── exemplars/
│       └── resource-api.yaml
└── zalando/
    └── exemplars/
        └── ecommerce-api.yaml
```

**MCP Resource:**

```
apistyle://profiles/{name}/exemplars/{exemplar}
```

### F7: Rule Priority and Dependencies

**Priority:** P2 (Medium)

Add priority and dependency metadata to rules:

```json
{
  "id": "URI-002",
  "title": "Use lowercase paths",
  "priority": 90,
  "dependencies": ["URI-001"],
  "enabledAt": "bronze",
  "blockingFor": ["silver", "gold"]
}
```

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `priority` | int | Fix order (100 = first, 1 = last) |
| `dependencies` | []string | Rules that must be fixed first |
| `enabledAt` | string | Minimum conformance level |
| `blockingFor` | []string | Levels this rule blocks if violated |

### F8: Pattern Library

**Priority:** P2 (Medium)

Reusable patterns in profiles:

```json
{
  "patterns": [
    {
      "id": "crud-collection",
      "name": "CRUD Collection Pattern",
      "description": "Standard endpoints for a resource collection",
      "template": {
        "paths": {
          "/{resources}": {
            "get": { "summary": "List {resources}" },
            "post": { "summary": "Create {resource}" }
          },
          "/{resources}/{id}": {
            "get": { "summary": "Get {resource}" },
            "put": { "summary": "Update {resource}" },
            "delete": { "summary": "Delete {resource}" }
          }
        }
      }
    },
    {
      "id": "pagination",
      "name": "Pagination Pattern",
      "description": "Standard pagination query parameters",
      "template": {
        "parameters": [
          { "name": "page", "in": "query", "schema": { "type": "integer", "default": 1 } },
          { "name": "pageSize", "in": "query", "schema": { "type": "integer", "default": 20 } }
        ]
      }
    }
  ]
}
```

### F9: Generation Rubrics

**Priority:** P2 (Medium)

Generate rubrics optimized for spec generation (inverse of evaluation rubrics):

```bash
api-style generate rubric --mode generation --profile azure
```

**Output format differs from evaluation:**

- Positive instructions ("DO this") vs negative ("Check for violations")
- Ordered by generation sequence
- Includes templates and examples
- Groups by API structure (info → paths → schemas)

## User Interface

### CLI Additions

```
api-style
├── lint
│   └── --suggest-fixes    # Include fix suggestions in output
├── suggest-fixes          # Standalone fix suggestion command
├── design-check           # Pre-generation guidance
└── conformance-path       # Path to conformance level
```

### MCP Server Additions

| Tool | Description |
|------|-------------|
| `suggest_fixes` | Generate fix suggestions for violations |
| `design_check` | Pre-generation design guidance |
| `conformance_path` | Path to reach conformance level |
| `get_pattern` | Get pattern template by ID |
| `list_exemplars` | List exemplar specs for a profile |

### MCP Resource Additions

| Resource | Description |
|----------|-------------|
| `apistyle://profiles/{name}/exemplars` | List exemplar specs |
| `apistyle://profiles/{name}/exemplars/{id}` | Get exemplar content |
| `apistyle://profiles/{name}/patterns` | List patterns |
| `apistyle://profiles/{name}/patterns/{id}` | Get pattern template |

## Success Criteria

| Metric | Target |
|--------|--------|
| Violation with suggestion | 80% of violations include actionable suggestions |
| Generation guidance | 100% of error-severity rules have `generate` field |
| Exemplar coverage | All built-in profiles have at least 1 exemplar |
| MCP tool adoption | suggest_fixes used in 50%+ of review workflows |
| Fix accuracy | 90% of suggested fixes resolve the violation |

## Dependencies

- Existing MCP server infrastructure
- Rule definitions in profiles
- LLM integration for suggest_fixes tool

## Risks

| Risk | Mitigation |
|------|------------|
| LLM suggestions may be wrong | Include confidence scores, require human review |
| Too many suggestions overwhelm agents | Prioritize by rule priority, limit batch size |
| Exemplars become outdated | Include exemplars in CI validation |
| Pattern templates too rigid | Allow customization via profile overrides |

## Timeline

See [ROADMAP.md](./ROADMAP.md) for implementation timeline.
