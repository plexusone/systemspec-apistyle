# CI/CD Integration

SystemSpec API Style is designed for CI/CD pipelines with machine-readable output formats and exit codes that signal pass/fail status.

## Exit Codes

All commands return appropriate exit codes for pipeline integration:

| Exit Code | Meaning |
|-----------|---------|
| `0` | Pass - no blocking violations |
| `1` | Fail - blocking violations or errors |

## Commands for CI/CD

### Lint Only (Deterministic)

Use `api-style lint` for fast, deterministic checks using Spectral rules:

```bash
api-style lint openapi.yaml \
  --profile plexusone-rest \
  --level standard \
  --format json
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--profile <name>` | Style profile to validate against |
| `--level <tier>` | Conformance level: `bronze`, `silver`, `gold` (or custom) |
| `--format <fmt>` | Output format: `text`, `json`, `sarif` |
| `--output <file>` | Write results to file |

### Full Analysis (Lint + LLM)

Use `api-style analyze` for comprehensive validation including AI-powered semantic checks:

```bash
api-style analyze openapi.yaml \
  --profile plexusone-rest \
  --level silver \
  --evaluate \
  --min-score 0.8 \
  --fail-warnings \
  --format json
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--evaluate` | Enable LLM-based evaluation |
| `--min-score <float>` | Minimum score threshold (0.0-1.0, default 0.7) |
| `--fail-warnings` | Treat warnings as failures (exit 1) |
| `--model <name>` | LLM model for evaluation |

## Output Formats

### JSON

Machine-readable format for programmatic processing:

```bash
api-style lint openapi.yaml --format json --output report.json
```

```json
{
  "files": [{
    "file": "openapi.yaml",
    "violations": [...],
    "summary": {
      "errors": 2,
      "warnings": 5,
      "info": 3
    }
  }],
  "decision": "no-go"
}
```

### SARIF

Standard format for GitHub Code Scanning and IDE integration:

```bash
api-style lint openapi.yaml --format sarif --output results.sarif
```

SARIF files can be uploaded to GitHub for inline PR annotations.

## GitHub Actions

### Basic Linting

```yaml
name: API Lint

on:
  pull_request:
    paths:
      - '**/*.yaml'
      - '**/*.json'

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install api-style
        run: go install github.com/plexusone/systemspec-apistyle/cmd/api-style@latest

      - name: Lint OpenAPI spec
        run: |
          api-style lint openapi.yaml \
            --profile plexusone-rest \
            --level standard \
            --format sarif \
            --output results.sarif

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

### Full Analysis with LLM

```yaml
name: API Analysis

on:
  pull_request:
    paths:
      - 'api/**'

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install api-style
        run: go install github.com/plexusone/systemspec-apistyle/cmd/api-style@latest

      - name: Analyze API spec
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          api-style analyze api/openapi.yaml \
            --profile plexusone-rest \
            --level silver \
            --evaluate \
            --min-score 0.8 \
            --format json \
            --output analysis.json

      - name: Upload analysis
        uses: actions/upload-artifact@v4
        with:
          name: api-analysis
          path: analysis.json
```

### Matrix Testing Multiple Specs

```yaml
name: API Lint Matrix

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        spec:
          - services/users/openapi.yaml
          - services/orders/openapi.yaml
          - services/payments/openapi.yaml
    steps:
      - uses: actions/checkout@v4

      - name: Install api-style
        run: go install github.com/plexusone/systemspec-apistyle/cmd/api-style@latest

      - name: Lint ${{ matrix.spec }}
        run: |
          api-style lint ${{ matrix.spec }} \
            --profile plexusone-rest \
            --level standard
```

## GitLab CI

```yaml
api-lint:
  image: golang:1.22
  stage: test
  script:
    - go install github.com/plexusone/systemspec-apistyle/cmd/api-style@latest
    - api-style lint openapi.yaml --profile plexusone-rest --level standard --format json
  artifacts:
    reports:
      codequality: report.json
  rules:
    - changes:
        - "**/*.yaml"
        - "**/*.json"
```

## Azure DevOps

```yaml
trigger:
  paths:
    include:
      - 'api/*'

pool:
  vmImage: 'ubuntu-latest'

steps:
  - task: GoTool@0
    inputs:
      version: '1.22'

  - script: |
      go install github.com/plexusone/systemspec-apistyle/cmd/api-style@latest
      api-style lint api/openapi.yaml \
        --profile plexusone-rest \
        --level standard \
        --format json \
        --output $(Build.ArtifactStagingDirectory)/lint-report.json
    displayName: 'Lint API Specification'

  - publish: $(Build.ArtifactStagingDirectory)/lint-report.json
    artifact: api-lint-report
```

## Conformance Levels

Use conformance levels to enforce graduated compliance:

| Level | Description | Use Case |
|-------|-------------|----------|
| `bronze` / `minimum` | Basic functionality | Early development |
| `silver` / `standard` | Production-ready | Pre-release |
| `gold` / `exemplary` | Best-in-class | Public APIs |

```bash
# Development - allow some issues
api-style lint openapi.yaml --level bronze

# Staging - stricter requirements
api-style lint openapi.yaml --level silver

# Production - full compliance
api-style lint openapi.yaml --level gold
```

## Pre-commit Hook

Add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: api-style-lint
        name: API Style Lint
        entry: api-style lint --profile plexusone-rest --level standard
        language: system
        files: '\.(yaml|json)$'
        types: [file]
```

## Makefile Integration

```makefile
.PHONY: lint-api analyze-api

PROFILE ?= plexusone-rest
LEVEL ?= standard

lint-api:
	api-style lint api/openapi.yaml \
		--profile $(PROFILE) \
		--level $(LEVEL) \
		--format text

analyze-api:
	api-style analyze api/openapi.yaml \
		--profile $(PROFILE) \
		--level $(LEVEL) \
		--evaluate \
		--min-score 0.8 \
		--format json

ci-lint:
	api-style lint api/openapi.yaml \
		--profile $(PROFILE) \
		--level $(LEVEL) \
		--format sarif \
		--output results.sarif
```

## Best Practices

1. **Start with bronze level** during development, increase as API matures
2. **Use SARIF format** for GitHub integration with inline annotations
3. **Cache the api-style binary** in CI to speed up builds
4. **Run lint on PRs** but full analysis on merges to main
5. **Store analysis artifacts** for audit trails
6. **Use secrets** for `ANTHROPIC_API_KEY` in LLM evaluation

## Troubleshooting

### Command not found

Ensure `$GOPATH/bin` is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### LLM evaluation fails

Check that `ANTHROPIC_API_KEY` is set:

```bash
echo $ANTHROPIC_API_KEY
```

### Timeout on large specs

Increase timeout for complex specifications:

```bash
api-style analyze large-spec.yaml --timeout 300
```
