# Custom Rules

Create custom API style rules to enforce your organization's specific guidelines.

## Rule Structure

Rules are defined in JSON format within a style profile:

```json
{
  "id": "URI-001",
  "title": "Use plural resource names",
  "category": "uri-design",
  "severity": "error",
  "rationale": "Plural resources improve consistency and make the API more intuitive.",
  "examples": {
    "good": ["/users", "/orders", "/products"],
    "bad": ["/user", "/order", "/product"]
  },
  "enforcement": {
    "type": "spectral",
    "given": "$.paths[*]~",
    "function": "pattern",
    "options": {
      "match": "^/[a-z]+s(/|$)"
    }
  }
}
```

## Rule Fields

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier (e.g., `URI-001`, `HTTP-002`) |
| `title` | string | Short, descriptive title |
| `category` | string | Rule category for grouping |
| `severity` | string | `error`, `warn`, `info`, or `hint` |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `rationale` | string | Why this rule exists |
| `examples` | object | Good and bad examples |
| `enforcement` | object | Spectral rule configuration |
| `judge` | object | LLM evaluation configuration |
| `conformanceLevel` | string | `bronze`, `silver`, or `gold` |
| `references` | array | External documentation links |

## Enforcement Types

### Spectral Rules

For deterministic, pattern-based checking:

```json
{
  "enforcement": {
    "type": "spectral",
    "given": "$.paths[*]~",
    "function": "pattern",
    "options": {
      "match": "^/[a-z-]+$"
    }
  }
}
```

Common Spectral functions:

| Function | Use Case |
|----------|----------|
| `pattern` | Regex matching |
| `truthy` | Value must be present |
| `falsy` | Value must be absent |
| `enumeration` | Value must be in list |
| `length` | String/array length |
| `schema` | JSON Schema validation |

### LLM Evaluation

For semantic analysis that can't be captured by patterns:

```json
{
  "judge": {
    "criteria": "Does the API use consistent terminology throughout?",
    "passCriteria": "All endpoints use the same terms for similar concepts",
    "failCriteria": "Different terms are used for the same concept (e.g., 'user' vs 'account')",
    "weight": 0.8
  }
}
```

## Creating a Custom Profile

### 1. Start with a Base Profile

```json
{
  "$schema": "https://plexusone.dev/systemspec-apistyle/schema/v1/api-style-spec.schema.json",
  "name": "my-company-style",
  "version": "1.0.0",
  "description": "My Company API Style Guidelines",
  "extends": ["default"],
  "rules": []
}
```

### 2. Add Custom Rules

```json
{
  "rules": [
    {
      "id": "MYCO-001",
      "title": "Use company standard error format",
      "category": "errors",
      "severity": "error",
      "rationale": "Consistent error handling across all APIs",
      "enforcement": {
        "type": "spectral",
        "given": "$.components.schemas[?(@.title == 'Error')]",
        "function": "truthy"
      }
    }
  ]
}
```

### 3. Override Inherited Rules

```json
{
  "extends": ["default"],
  "overrides": {
    "URI-001": {
      "severity": "warn"
    }
  }
}
```

## Rule Categories

Organize rules into logical categories:

| Category | Description |
|----------|-------------|
| `uri-design` | URL structure and naming |
| `http-methods` | HTTP verb usage |
| `status-codes` | Response status codes |
| `request-response` | Payload structure |
| `errors` | Error handling |
| `security` | Authentication/authorization |
| `versioning` | API version management |
| `documentation` | Descriptions and examples |

## Examples

### Require Operation IDs

```json
{
  "id": "DOC-001",
  "title": "Operations must have operationId",
  "category": "documentation",
  "severity": "error",
  "enforcement": {
    "type": "spectral",
    "given": "$.paths[*][get,post,put,patch,delete]",
    "then": {
      "field": "operationId",
      "function": "truthy"
    }
  }
}
```

### Enforce Semantic Versioning

```json
{
  "id": "VER-001",
  "title": "Use semantic versioning",
  "category": "versioning",
  "severity": "error",
  "enforcement": {
    "type": "spectral",
    "given": "$.info.version",
    "function": "pattern",
    "options": {
      "match": "^\\d+\\.\\d+\\.\\d+$"
    }
  }
}
```

### Require HTTPS

```json
{
  "id": "SEC-001",
  "title": "Use HTTPS only",
  "category": "security",
  "severity": "error",
  "enforcement": {
    "type": "spectral",
    "given": "$.servers[*].url",
    "function": "pattern",
    "options": {
      "match": "^https://"
    }
  }
}
```

## Testing Custom Rules

1. Create a test OpenAPI spec with known violations
2. Run the linter with your custom profile
3. Verify expected violations are reported

```bash
api-style lint test-spec.yaml --profile ./my-company-style.json
```

## Next Steps

- [Specification Format](../reference/schema.md) - Complete schema reference
- [CLI Reference](../reference/cli.md) - Command documentation
