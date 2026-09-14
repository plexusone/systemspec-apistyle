# Configuration File

Configure systemspec-apistyle defaults via `.api-style.yaml` in your project root.

## Overview

The configuration file lets you:

- Set default profile and conformance level
- Define file include/exclude patterns
- Configure rule exceptions
- Override rule severities

## File Locations

SystemSpec API Style searches for config files in this order:

1. Path specified via `--config` flag
2. `.api-style.yaml` in current directory
3. `.api-style.yml` in current directory
4. `api-style.yaml` in current directory

## Complete Example

```yaml
# .api-style.yaml

# Style profile to use (default, azure, google, zalando)
profile: azure

# Target conformance level (bronze, silver, gold)
level: silver

# File patterns to include
include:
  - "openapi.yaml"
  - "openapi.yml"
  - "openapi.json"
  - "**/openapi.yaml"
  - "**/api.yaml"
  - "specs/**/*.yaml"

# File patterns to exclude
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
  - "**/.git/**"
  - "**/generated/**"

# Rule exceptions
exceptions:
  - rule: URI-001
    paths:
      - "/legacy/**"
    reason: "Legacy API cannot be renamed"

  - rule: operation-description
    paths:
      - "/internal/**"
    reason: "Internal APIs have relaxed requirements"

# Override rule severities
severity-overrides:
  URI-002: warn
  info-description: info
```

## Options Reference

### profile

Style profile to use for linting.

| Value | Description |
|-------|-------------|
| `default` | Common REST API best practices |
| `azure` | Microsoft/Azure API guidelines |
| `google` | Google API Design Guide |
| `zalando` | Zalando RESTful Guidelines |

### level

Target conformance level.

| Value | Description |
|-------|-------------|
| `bronze` | Basic compliance |
| `silver` | Standard compliance |
| `gold` | Full compliance |

### include

File patterns to include in linting. Supports:

- Exact filenames: `openapi.yaml`
- Glob patterns: `*.yaml`
- Recursive patterns: `**/openapi.yaml`
- Directory patterns: `specs/**/*.yaml`

### exclude

File patterns to exclude from linting. Same pattern syntax as `include`.

Common exclusions:

```yaml
exclude:
  - "**/node_modules/**"
  - "**/vendor/**"
  - "**/.git/**"
  - "**/generated/**"
  - "**/mock/**"
  - "**/test/**"
```

### exceptions

Rule exceptions for specific paths.

```yaml
exceptions:
  - rule: <rule-id>
    paths:
      - "<glob-pattern>"
    reason: "<explanation>"
```

Each exception has:

| Field | Required | Description |
|-------|----------|-------------|
| `rule` | Yes | Rule ID to suppress (e.g., `URI-001`) |
| `paths` | Yes | Glob patterns where exception applies |
| `reason` | Yes | Explanation for the exception |

### severity-overrides

Override default rule severities.

```yaml
severity-overrides:
  <rule-id>: <severity>
```

| Severity | Description |
|----------|-------------|
| `error` | Blocking violation |
| `warn` | Non-blocking warning |
| `info` | Informational |
| `hint` | Suggestion |

## CLI Flag Precedence

CLI flags override config file settings:

```bash
# Config says profile: azure, but CLI overrides to google
api-style lint openapi.yaml --profile google
```

Precedence order:

1. CLI flags (highest)
2. Config file
3. Built-in defaults (lowest)

## JSON Format

Configuration can also be JSON:

```json
{
  "profile": "azure",
  "level": "silver",
  "include": ["openapi.yaml", "**/api.yaml"],
  "exclude": ["**/generated/**"],
  "exceptions": [
    {
      "rule": "URI-001",
      "paths": ["/legacy/**"],
      "reason": "Legacy API"
    }
  ],
  "severityOverrides": {
    "URI-002": "warn"
  }
}
```

Note: JSON uses `severityOverrides` (camelCase) while YAML uses `severity-overrides` (kebab-case).
