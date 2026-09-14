# Implementation Plan

## AI Agent Support for systemspec-apistyle

**Version:** 0.5.0
**Date:** 2026-07-14

## Overview

This plan organizes implementation into phases, with each phase delivering incremental value. Dependencies between tasks are noted.

## Phase 1: Enhanced Violation Remediation

**Goal:** Violations include actionable fix guidance for AI agents.

### Task 1.1: Extend Violation Type

**Files:** `pkg/types/report.go`

Add new fields to `Violation`:

```go
type Violation struct {
    // Existing fields...

    // NEW fields
    ExampleFix   string   `json:"exampleFix,omitempty"`
    RuleURL      string   `json:"ruleUrl,omitempty"`
    Confidence   float64  `json:"confidence,omitempty"`
    RelatedRules []string `json:"relatedRules,omitempty"`
    FixPriority  int      `json:"fixPriority,omitempty"`
}
```

**Acceptance:**

- [ ] Fields added with JSON tags
- [ ] Existing tests pass
- [ ] JSON output includes new fields when populated

### Task 1.2: Enrich Violations During Linting

**Files:** `pkg/lint/results.go`, `pkg/lint/vacuum.go`

**Depends on:** Task 1.1

Populate new fields when converting vacuum results:

```go
func enrichViolation(v *types.Violation, rule *types.Rule, profile *types.APIStyleSpec) {
    v.RuleURL = generateRuleURL(profile.Name, rule.ID)
    v.Confidence = 1.0 // Deterministic
    v.FixPriority = calculateFixPriority(rule)
    v.RelatedRules = extractRelatedRules(rule)

    if rule.Examples != nil && len(rule.Examples.Good) > 0 {
        v.ExampleFix = generateExampleFix(v, rule)
    }
}
```

**Acceptance:**

- [ ] All lint violations have RuleURL populated
- [ ] Violations have Confidence = 1.0
- [ ] Rules with examples have ExampleFix populated
- [ ] Rules with relations have RelatedRules populated

### Task 1.3: Update CLI Output Formats

**Files:** `cmd/api-style/lint.go`

**Depends on:** Task 1.2

Update text, JSON, and SARIF formatters to include new fields:

```
[URI-001] Resource name 'user' should be plural
  Path: $.paths./user (line 42)
  Suggestion: Rename path from /user to /users
  Example: /users:
             get:
               operationId: listUsers
  Fix Priority: 1 (fix first)
  Docs: https://plexusone.dev/systemspec-apistyle/profiles/default#uri-001
```

**Acceptance:**

- [ ] Text format shows suggestion and example
- [ ] JSON format includes all new fields
- [ ] SARIF format maps fields appropriately

### Task 1.4: Add --suggest-fixes Flag

**Files:** `cmd/api-style/lint.go`

**Depends on:** Task 1.3

Add flag to include enhanced remediation:

```bash
api-style lint openapi.yaml --suggest-fixes
```

When enabled:

- Include ExampleFix for all violations
- Sort violations by FixPriority
- Group related violations

**Acceptance:**

- [ ] Flag is documented in --help
- [ ] Violations sorted by priority when flag enabled
- [ ] Related violations grouped together

---

## Phase 2: Fix Suggestion Engine

**Goal:** Dedicated package and CLI for generating fix suggestions.

### Task 2.1: Create pkg/fix Package

**Files:** `pkg/fix/fixer.go`, `pkg/fix/rule_fixer.go`

Create the fix suggestion interface and rule-based implementation:

```go
type Fixer interface {
    SuggestFixes(ctx context.Context, spec []byte, violations []types.Violation, opts *Options) (*types.FixReport, error)
}

type RuleFixer struct {
    profile *types.APIStyleSpec
}
```

**Acceptance:**

- [ ] Interface defined
- [ ] Rule-based fixer generates suggestions from rule examples
- [ ] Unit tests for common violation patterns

### Task 2.2: Create pkg/types/fix.go

**Files:** `pkg/types/fix.go`

Define fix-related types:

```go
type FixSuggestion struct {
    RuleID         string  `json:"ruleId"`
    Path           string  `json:"path"`
    CurrentValue   string  `json:"currentValue,omitempty"`
    SuggestedValue string  `json:"suggestedValue"`
    Diff           string  `json:"diff,omitempty"`
    Confidence     float64 `json:"confidence"`
    Reasoning      string  `json:"reasoning,omitempty"`
}

type FixReport struct {
    Suggestions     []FixSuggestion  `json:"suggestions"`
    PatchOperations []PatchOperation `json:"patchOperations,omitempty"`
    FixedCount      int              `json:"fixedCount"`
    UnfixedCount    int              `json:"unfixedCount"`
}
```

**Acceptance:**

- [ ] Types defined with JSON tags
- [ ] Types compile successfully
- [ ] Integration with existing types package

### Task 2.3: JSON Patch Generation

**Files:** `pkg/fix/patch.go`

Generate RFC 6902 JSON Patch operations:

```go
func GeneratePatch(suggestions []types.FixSuggestion) []types.PatchOperation {
    // Convert suggestions to JSON Patch operations
}
```

**Acceptance:**

- [ ] Valid RFC 6902 operations generated
- [ ] Handles add, remove, replace, move
- [ ] Tests verify patch application

### Task 2.4: Add suggest-fixes CLI Command

**Files:** `cmd/api-style/suggest.go`

**Depends on:** Task 2.1, 2.2

New standalone command:

```bash
api-style suggest-fixes openapi.yaml --profile azure
api-style suggest-fixes openapi.yaml --output fixes.json
api-style suggest-fixes openapi.yaml --patch # Output JSON Patch
```

**Acceptance:**

- [ ] Command documented in --help
- [ ] Outputs fix suggestions in JSON format
- [ ] --patch flag outputs JSON Patch operations

---

## Phase 3: MCP Tool Integration

**Goal:** AI agents can access fix suggestions via MCP.

### Task 3.1: Add suggest_fixes Tool

**Files:** `cmd/mcp-api-style/tools.go`

**Depends on:** Task 2.1

Register new MCP tool:

```go
var suggestFixesTool = mcp.Tool{
    Name: "suggest_fixes",
    Description: "Generate fix suggestions for violations",
    // ...
}
```

**Acceptance:**

- [ ] Tool registered in MCP server
- [ ] Input schema documented
- [ ] Returns FixReport JSON
- [ ] Integration test passes

### Task 3.2: Add design_check Tool

**Files:** `cmd/mcp-api-style/tools.go`, `pkg/fix/design.go`

**Depends on:** Task 2.1

Pre-generation guidance tool:

```go
var designCheckTool = mcp.Tool{
    Name: "design_check",
    Description: "Get design guidance before generating OpenAPI spec",
    // ...
}
```

**Acceptance:**

- [ ] Tool generates checklist from profile rules
- [ ] Returns template for common patterns
- [ ] Ordered by rule priority

### Task 3.3: Add conformance_path Tool

**Files:** `cmd/mcp-api-style/tools.go`, `pkg/fix/conformance.go`

**Depends on:** Task 2.1

Show path to conformance level:

```go
var conformancePathTool = mcp.Tool{
    Name: "conformance_path",
    Description: "Show path to reach conformance level",
    // ...
}
```

**Acceptance:**

- [ ] Calculates blockers and warnings
- [ ] Shows progress percentage
- [ ] Orders fixes by priority

### Task 3.4: Update MCP Resources

**Files:** `cmd/mcp-api-style/resources.go`

Add new resources:

- `apistyle://profiles/{name}/exemplars`
- `apistyle://profiles/{name}/exemplars/{id}`
- `apistyle://profiles/{name}/patterns`

**Acceptance:**

- [ ] Resources return correct data
- [ ] 404 for unknown exemplars/patterns
- [ ] Integration tests pass

---

## Phase 4: Generation Guidance

**Goal:** Rules include generation-focused metadata.

### Task 4.1: Add GenerationGuidance Type

**Files:** `pkg/types/rule.go`

```go
type GenerationGuidance struct {
    Prompt    string              `json:"prompt"`
    Template  string              `json:"template,omitempty"`
    Priority  int                 `json:"priority,omitempty"`
    Examples  []GenerationExample `json:"examples,omitempty"`
    Checklist []string            `json:"checklist,omitempty"`
}
```

**Acceptance:**

- [ ] Type defined with JSON tags
- [ ] Rule.Generate field added
- [ ] Existing tests pass

### Task 4.2: Add Generation Metadata to Profiles

**Files:** `profiles/default/profile.json`, `profiles/azure/profile.json`, etc.

**Depends on:** Task 4.1

Add `generate` field to high-priority rules:

```json
{
  "id": "URI-001",
  "generate": {
    "prompt": "Use plural nouns for collection endpoints",
    "template": "/{resources}",
    "priority": 100
  }
}
```

**Acceptance:**

- [ ] All error-severity rules have generate.prompt
- [ ] Templates provided for URI rules
- [ ] Priority set for ordering

### Task 4.3: Generate Generation Rubric

**Files:** `pkg/generate/rubric.go`

**Depends on:** Task 4.1

Add `--mode generation` to rubric generation:

```bash
api-style generate rubric --mode generation --profile azure
```

**Acceptance:**

- [ ] Outputs rules ordered by generate.priority
- [ ] Includes prompts as directives
- [ ] Groups by generation phase (info → paths → schemas)

---

## Phase 5: Exemplar Specs

**Goal:** Each profile has reference OpenAPI specs.

### Task 5.1: Create Exemplar Directory Structure

**Files:** `profiles/*/exemplars/*.yaml`

Create exemplar specs:

```
profiles/
├── default/
│   └── exemplars/
│       ├── minimal.yaml     # Minimal conformant spec
│       └── crud-api.yaml    # CRUD API example
├── azure/
│   └── exemplars/
│       └── resource-api.yaml
└── zalando/
    └── exemplars/
        └── ecommerce-api.yaml
```

**Acceptance:**

- [ ] Each profile has at least 1 exemplar
- [ ] Exemplars pass lint with their profile
- [ ] Exemplars are well-documented

### Task 5.2: Exemplar Loading

**Files:** `pkg/profile/exemplars.go`

Load exemplars from embedded files:

```go
func ListExemplars(profileName string) ([]ExemplarInfo, error)
func GetExemplar(profileName, exemplarID string) ([]byte, error)
```

**Acceptance:**

- [ ] Exemplars embedded at compile time
- [ ] API returns exemplar content
- [ ] Error handling for missing exemplars

### Task 5.3: CLI List Exemplars

**Files:** `cmd/api-style/profile.go`

```bash
api-style profile exemplars default
api-style profile exemplars azure --show crud-api
```

**Acceptance:**

- [ ] Lists exemplars for a profile
- [ ] --show displays exemplar content
- [ ] Documented in --help

---

## Phase 6: Pattern Library

**Goal:** Reusable patterns in profiles.

### Task 6.1: Define Pattern Type

**Files:** `pkg/types/pattern.go`

```go
type Pattern struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Template    map[string]any `json:"template"`
    Variables   []PatternVar   `json:"variables,omitempty"`
}
```

**Acceptance:**

- [ ] Type defined
- [ ] Supports variable substitution
- [ ] JSON schema validation

### Task 6.2: Add Patterns to Profiles

**Files:** `profiles/*/profile.json`

**Depends on:** Task 6.1

```json
{
  "patterns": [
    {
      "id": "crud-collection",
      "name": "CRUD Collection",
      "template": { ... }
    }
  ]
}
```

**Acceptance:**

- [ ] Common patterns defined
- [ ] Patterns include CRUD, pagination, error responses
- [ ] Templates are valid OpenAPI fragments

### Task 6.3: Pattern Expansion

**Files:** `pkg/profile/patterns.go`

**Depends on:** Task 6.2

```go
func ExpandPattern(pattern *Pattern, vars map[string]string) (map[string]any, error)
```

**Acceptance:**

- [ ] Variables substituted correctly
- [ ] Invalid patterns return errors
- [ ] Tests cover common patterns

### Task 6.4: MCP get_pattern Tool

**Files:** `cmd/mcp-api-style/tools.go`

**Depends on:** Task 6.3

```go
var getPatternTool = mcp.Tool{
    Name: "get_pattern",
    Description: "Get a pattern template with variable substitution",
}
```

**Acceptance:**

- [ ] Tool returns expanded template
- [ ] Variables can be provided
- [ ] Integration test passes

---

## Implementation Order

```
Phase 1 (Week 1)
├── Task 1.1: Extend Violation Type
├── Task 1.2: Enrich Violations
├── Task 1.3: Update CLI Output
└── Task 1.4: Add --suggest-fixes

Phase 2 (Week 2)
├── Task 2.1: Create pkg/fix
├── Task 2.2: Create types/fix.go
├── Task 2.3: JSON Patch Generation
└── Task 2.4: suggest-fixes CLI

Phase 3 (Week 3)
├── Task 3.1: suggest_fixes MCP Tool
├── Task 3.2: design_check MCP Tool
├── Task 3.3: conformance_path MCP Tool
└── Task 3.4: Update MCP Resources

Phase 4 (Week 4)
├── Task 4.1: GenerationGuidance Type
├── Task 4.2: Add to Profiles
└── Task 4.3: Generation Rubric

Phase 5 (Week 5)
├── Task 5.1: Exemplar Directory
├── Task 5.2: Exemplar Loading
└── Task 5.3: CLI List Exemplars

Phase 6 (Week 6)
├── Task 6.1: Pattern Type
├── Task 6.2: Add to Profiles
├── Task 6.3: Pattern Expansion
└── Task 6.4: get_pattern MCP Tool
```

## Testing Requirements

Each phase must include:

1. **Unit tests** for new functions
2. **Integration tests** for CLI commands
3. **MCP integration tests** for new tools
4. **Golden file tests** for output formats
5. **Profile validation** ensuring exemplars pass lint

## Documentation Updates

For each phase, update:

- `docs/reference/cli.md` - New commands and flags
- `docs/guide/mcp-server.md` - New tools and resources
- `docs/guide/profiles.md` - Exemplars and patterns
- `CHANGELOG.json` - Feature entries
