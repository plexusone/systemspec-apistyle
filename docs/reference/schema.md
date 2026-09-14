# Specification Format

Complete reference for the systemspec-apistyle JSON schema.

## Overview

Style profiles are defined in JSON format and contain rules for API design guidelines.

```json
{
  "$schema": "https://plexusone.dev/systemspec-apistyle/schema/v1/api-style-spec.schema.json",
  "name": "my-style",
  "version": "1.0.0",
  "description": "My API Style Guidelines",
  "rules": [...]
}
```

## Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `$schema` | string | No | JSON Schema URL for validation |
| `name` | string | Yes | Unique profile identifier |
| `version` | string | Yes | Profile version (semver) |
| `description` | string | No | Human-readable description |
| `extends` | array | No | Base profiles to inherit from |
| `rules` | array | Yes | List of style rules |
| `overrides` | object | No | Override inherited rules |
| `metadata` | object | No | Additional metadata |

## Rule Schema

### Required Fields

```json
{
  "id": "URI-001",
  "title": "Use plural resource names",
  "category": "uri-design",
  "severity": "error"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique rule identifier |
| `title` | string | Short descriptive title |
| `category` | string | Rule category |
| `severity` | string | `error`, `warn`, `info`, `hint` |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `rationale` | string | Why this rule exists |
| `examples` | Examples | Good and bad examples |
| `enforcement` | Enforcement | Deterministic rule config |
| `judge` | Judge | LLM evaluation config |
| `conformanceLevel` | string | `bronze`, `silver`, `gold` |
| `references` | array | External links |
| `tags` | array | Searchable tags |

## Examples Object

```json
{
  "examples": {
    "good": [
      "/users",
      "/orders/{orderId}/items"
    ],
    "bad": [
      "/user",
      "/getOrder"
    ]
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `good` | array | Examples that pass the rule |
| `bad` | array | Examples that fail the rule |

## Enforcement Object

For deterministic, pattern-based rules:

```json
{
  "enforcement": {
    "type": "spectral",
    "given": "$.paths[*]~",
    "then": {
      "function": "pattern",
      "options": {
        "match": "^/[a-z-]+$"
      }
    }
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Rule engine: `spectral` |
| `given` | string/array | JSONPath to target elements |
| `then` | object | Assertion to apply |

### Spectral Functions

| Function | Description | Options |
|----------|-------------|---------|
| `truthy` | Value must be present | - |
| `falsy` | Value must be absent | - |
| `pattern` | Regex match | `match`, `notMatch` |
| `enumeration` | Value in list | `values` |
| `length` | Length check | `min`, `max` |
| `schema` | JSON Schema validation | `schema` |
| `alphabetical` | Sorted order | `keyedBy` |

### Given Path Examples

| Path | Targets |
|------|---------|
| `$.paths[*]~` | All path keys |
| `$.paths[*][*]` | All operations |
| `$.paths[*].get` | All GET operations |
| `$.components.schemas[*]` | All schemas |
| `$.info` | Info object |

## Judge Object

For LLM-based semantic evaluation:

```json
{
  "judge": {
    "criteria": "Does the API follow resource-oriented design?",
    "passCriteria": "All endpoints represent resources with standard CRUD operations",
    "failCriteria": "Endpoints use RPC-style naming or action verbs",
    "weight": 0.8,
    "rubric": [
      {"score": 5, "description": "Excellent resource design"},
      {"score": 3, "description": "Mostly resource-oriented"},
      {"score": 1, "description": "Mixed patterns"}
    ]
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `criteria` | string | What to evaluate |
| `passCriteria` | string | What passing looks like |
| `failCriteria` | string | What failing looks like |
| `weight` | number | Importance (0-1) |
| `rubric` | array | Scoring guidelines |

## References Array

```json
{
  "references": [
    {
      "title": "Microsoft REST API Guidelines",
      "url": "https://github.com/microsoft/api-guidelines"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Reference title |
| `url` | string | Reference URL |

## Extends Array

Inherit from base profiles:

```json
{
  "extends": ["default"],
  "rules": [
    // Additional rules
  ]
}
```

Rules from extended profiles are merged. Child rules with the same ID override parent rules.

## Overrides Object

Modify inherited rules without redefining them:

```json
{
  "extends": ["default"],
  "overrides": {
    "URI-001": {
      "severity": "warn"
    },
    "DOC-001": {
      "enabled": false
    }
  }
}
```

Supported override fields:

- `severity` - Change rule severity
- `enabled` - Enable/disable rule
- `conformanceLevel` - Change required level

## Categories

Standard category names:

| Category | Description |
|----------|-------------|
| `uri-design` | URL structure and naming |
| `http-methods` | HTTP verb usage |
| `status-codes` | Response status codes |
| `request-response` | Payload structure |
| `errors` | Error handling |
| `security` | Auth and security |
| `versioning` | API versioning |
| `documentation` | Descriptions and examples |
| `naming` | Field and property naming |
| `pagination` | Collection pagination |

## Severity Levels

| Severity | Description | Blocks CI |
|----------|-------------|-----------|
| `error` | Must fix | Yes |
| `warn` | Should fix | No |
| `info` | Consider | No |
| `hint` | Suggestion | No |

## Conformance Levels

| Level | Description |
|-------|-------------|
| `bronze` | Minimum viable quality |
| `silver` | Production-ready |
| `gold` | Best-in-class |

Rules tagged with a level are required at that level and above.

## Complete Example

```json
{
  "$schema": "https://plexusone.dev/systemspec-apistyle/schema/v1/api-style-spec.schema.json",
  "name": "my-company-api-style",
  "version": "1.0.0",
  "description": "My Company API Design Guidelines",
  "extends": ["default"],
  "rules": [
    {
      "id": "MYCO-001",
      "title": "Use company error format",
      "category": "errors",
      "severity": "error",
      "conformanceLevel": "bronze",
      "rationale": "Consistent error handling improves developer experience",
      "examples": {
        "good": ["{ \"error\": { \"code\": \"NOT_FOUND\", \"message\": \"...\" } }"],
        "bad": ["{ \"message\": \"Not found\" }"]
      },
      "enforcement": {
        "type": "spectral",
        "given": "$.components.responses[*].content['application/json'].schema",
        "then": {
          "field": "properties.error",
          "function": "truthy"
        }
      },
      "references": [
        {
          "title": "Company API Standards",
          "url": "https://docs.mycompany.com/api-standards"
        }
      ]
    }
  ],
  "overrides": {
    "URI-001": {
      "severity": "warn"
    }
  }
}
```
