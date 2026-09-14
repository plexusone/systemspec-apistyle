# Documentation Generation

SystemSpec API Style can generate human-readable documentation from your style profiles. This allows you to maintain a single source of truth (the JSON profile) while producing documentation in multiple formats.

## Output Formats

| Format | Description | Use Case |
|--------|-------------|----------|
| Single-page Markdown | One comprehensive `.md` file | GitHub wikis, simple docs |
| MkDocs site | Multi-page documentation site | Full documentation websites |
| Spectral ruleset | YAML ruleset for linting | CI/CD pipelines |

## Single-Page Markdown

Generate a complete style guide as a single Markdown file:

```bash
api-style generate guide --profile zalando --output zalando-guide.md
```

The generated Markdown includes:

- **Title and description**
- **Table of contents**
- **Introduction** (if defined)
- **Design principles**
- **Conformance levels**
- **Design patterns** with examples
- **All rules** organized by category
- **Glossary**

### Example Output

```markdown
# Zalando REST API Guidelines

Comprehensive style rules based on Zalando RESTful API Guidelines...

## Table of Contents

- [Introduction](#introduction)
- [Design Principles](#design-principles)
- [Design Patterns](#design-patterns)
- **Rules**
  - [General Guidelines](#general-guidelines)
  - [JSON Guidelines](#json-guidelines)
  ...
- [Glossary](#glossary)

## Introduction

These guidelines define standards for designing APIs at Zalando...

## Design Principles

### API First
Design the API contract before implementation...
```

### Options

Control what's included in the output:

```bash
# Include everything (default)
api-style generate guide --profile azure

# Skip examples
api-style generate guide --profile azure --no-examples

# Skip patterns section
api-style generate guide --profile azure --no-patterns

# Include severity emojis
api-style generate guide --profile azure --emojis
```

## MkDocs Multi-Page Site

Generate a complete MkDocs documentation site:

```bash
api-style generate mkdocs --profile zalando --output ./docs-site
```

This creates a ready-to-use MkDocs project:

```
docs-site/
├── mkdocs.yml          # MkDocs configuration
└── docs/
    ├── index.md        # Home page with overview
    ├── introduction.md # Full introduction
    ├── principles.md   # Design principles
    ├── patterns.md     # Design patterns
    ├── conformance.md  # Conformance levels
    ├── glossary.md     # Glossary terms
    └── rules/
        ├── index.md    # Rules overview
        ├── general.md  # General rules
        ├── json.md     # JSON guidelines
        ├── urls.md     # URL design rules
        └── ...         # One file per category
```

### Building the Site

After generating, build and serve the documentation:

```bash
cd docs-site
pip install mkdocs-material
mkdocs serve
```

Visit `http://localhost:8000` to view the documentation.

### MkDocs Options

```bash
# Split patterns into individual pages
api-style generate mkdocs --profile azure --split-patterns

# Keep all rules in one page
api-style generate mkdocs --profile azure --no-split-categories

# Custom site name
api-style generate mkdocs --profile azure --site-name "My API Guidelines"

# Set repository URL
api-style generate mkdocs --profile azure --repo-url https://github.com/myorg/api-guidelines
```

### Generated mkdocs.yml

The generated configuration includes Material theme with:

- **Navigation tabs** for major sections
- **Search** with highlighting
- **Dark/light mode** toggle
- **Code copy** buttons
- **Mermaid diagrams** support
- **Syntax highlighting**

Example:

```yaml
site_name: Zalando REST API Guidelines
repo_url: https://github.com/zalando/restful-api-guidelines

theme:
  name: material
  features:
    - navigation.tabs
    - navigation.sections
    - search.highlight
    - content.code.copy

nav:
  - Home: index.md
  - Introduction: introduction.md
  - Principles: principles.md
  - Patterns: patterns.md
  - Rules:
    - Overview: rules/index.md
    - General: rules/general.md
    - JSON: rules/json.md
    ...
  - Glossary: glossary.md
```

## Spectral Ruleset

Generate a Spectral-compatible ruleset for use with [Spectral](https://github.com/stoplightio/spectral) or [vacuum](https://github.com/daveshanley/vacuum):

```bash
api-style generate spectral --profile azure --output .spectral.yaml
```

This exports rules that have deterministic enforcement defined:

```yaml
# Generated from: Azure REST API Guidelines
# Version: 1.0.0

extends: []

rules:
  uri-001:
    description: Use plural resource names
    message: "{{description}}"
    severity: error
    given: "$.paths[*]~"
    then:
      function: pattern
      functionOptions:
        match: "^/[a-z]+s(/.*)?$"
```

## Programmatic Usage

Use the generation functions in Go code:

### Single-Page Markdown

```go
import (
    "github.com/plexusone/systemspec-apistyle/pkg/generate"
    "github.com/plexusone/systemspec-apistyle/pkg/profile"
)

// Load profile
spec, err := profile.Load("azure")
if err != nil {
    log.Fatal(err)
}

// Generate Markdown with default options
md, err := generate.Markdown(spec, nil)
if err != nil {
    log.Fatal(err)
}

// Write to file
os.WriteFile("guide.md", []byte(md), 0644)
```

### Custom Options

```go
opts := &generate.MarkdownOptions{
    IncludeTOC:          true,
    IncludeExamples:     true,
    IncludeRationale:    true,
    IncludeReferences:   true,
    IncludeConformance:  true,
    IncludeMetadata:     true,
    IncludeIntroduction: true,
    IncludePrinciples:   true,
    IncludePatterns:     true,
    IncludeGlossary:     true,
    IncludeDescription:  true,
    SeverityEmojis:      false,
}

md, err := generate.Markdown(spec, opts)
```

### MkDocs Site

```go
import "github.com/plexusone/systemspec-apistyle/pkg/generate"

// Generate MkDocs site structure
opts := &generate.MkDocsOptions{
    SiteName:        "My API Guidelines",
    SiteURL:         "https://api.example.com/guidelines",
    RepoURL:         "https://github.com/example/api-guidelines",
    Theme:           "material",
    IncludeSearch:   true,
    SplitCategories: true,
    SplitPatterns:   false,
}

result, err := generate.MkDocs(spec, opts)
if err != nil {
    log.Fatal(err)
}

// Write to filesystem
err = generate.WriteMkDocs(result, "./output")
if err != nil {
    log.Fatal(err)
}
```

### Access Individual Pages

```go
result, _ := generate.MkDocs(spec, nil)

// Access generated files
for path, content := range result.Pages {
    fmt.Printf("Generated: %s (%d bytes)\n", path, len(content))
}

// Access mkdocs.yml
fmt.Println(result.Config)
```

## Content in Generated Documentation

### Design Principles

Principles defined in the profile are rendered with their descriptions and related rules:

```json
{
  "principles": [
    {
      "id": "api-first",
      "title": "API First",
      "description": "Design the API contract before implementation...",
      "relatedRules": ["DESIGN-001", "DESIGN-002"]
    }
  ]
}
```

### Design Patterns

Patterns include problem/solution descriptions, examples, and diagrams:

```json
{
  "patterns": [
    {
      "id": "cursor-pagination",
      "name": "Cursor-Based Pagination",
      "summary": "Use opaque cursors for efficient pagination",
      "problem": "Offset pagination has performance issues...",
      "solution": "Use opaque cursor tokens that encode position...",
      "examples": [...],
      "diagrams": [
        {
          "title": "Cursor Flow",
          "type": "mermaid",
          "content": "sequenceDiagram..."
        }
      ]
    }
  ]
}
```

### Detailed Examples

Rules can include rich examples with code blocks and annotations:

```json
{
  "examples": {
    "detailed": [
      {
        "title": "Paginated Response",
        "type": "good",
        "language": "json",
        "code": "{\n  \"items\": [...],\n  \"next\": \"cursor123\"\n}",
        "annotations": [
          {"line": 2, "text": "Use 'items' for the collection array"},
          {"line": 3, "text": "Opaque cursor for next page"}
        ]
      }
    ]
  }
}
```

### Decision Tables

Render structured decision guidance:

```json
{
  "decisionTables": [
    {
      "title": "HTTP Method Selection",
      "headers": ["Operation", "Method", "Idempotent"],
      "rows": [
        {"values": ["Create", "POST", "No"]},
        {"values": ["Read", "GET", "Yes"]},
        {"values": ["Replace", "PUT", "Yes"]}
      ]
    }
  ]
}
```

## CI/CD Integration

### Generate Documentation on Release

```yaml
# .github/workflows/docs.yml
name: Generate Documentation

on:
  release:
    types: [published]

jobs:
  docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install api-style
        run: go install github.com/plexusone/systemspec-apistyle/cmd/api-style@latest

      - name: Generate MkDocs site
        run: api-style generate mkdocs --profile ./my-profile.json --output ./site

      - name: Build site
        run: |
          pip install mkdocs-material
          cd site && mkdocs build

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v4
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./site/site
```

### Keep Docs in Sync

```yaml
# .github/workflows/sync-docs.yml
name: Sync Documentation

on:
  push:
    paths:
      - 'profiles/*.json'

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Regenerate docs
        run: |
          api-style generate guide --profile profiles/company.json --output docs/api-guidelines.md

      - name: Commit changes
        run: |
          git config user.name "github-actions"
          git config user.email "github-actions@github.com"
          git add docs/
          git commit -m "docs: regenerate API guidelines" || exit 0
          git push
```

## Next Steps

- [Using Profiles](profiles.md) - Learn about built-in profiles
- [Custom Rules](custom-rules.md) - Create your own rules
- [CLI Reference](../reference/cli.md) - Complete command documentation
