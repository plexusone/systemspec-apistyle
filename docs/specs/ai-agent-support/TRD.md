# Technical Requirements Document (TRD)

## AI Agent Support for systemspec-apistyle

**Version:** 0.5.0-draft
**Date:** 2026-07-14
**Status:** Draft

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        AI Agent Integration                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐              │
│  │  Claude Code │    │    Cursor    │    │     Kiro     │              │
│  │   (Agent)    │    │   (Agent)    │    │   (Agent)    │              │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘              │
│         │                   │                   │                       │
│         └───────────────────┼───────────────────┘                       │
│                             │                                           │
│                    ┌────────▼────────┐                                  │
│                    │   MCP Server    │                                  │
│                    │  (mcp-api-style)│                                  │
│                    └────────┬────────┘                                  │
│                             │                                           │
│    ┌────────────────────────┼────────────────────────┐                  │
│    │                        │                        │                  │
│    ▼                        ▼                        ▼                  │
│ ┌──────────┐         ┌──────────┐         ┌──────────────┐             │
│ │  lint    │         │ evaluate │         │ NEW: suggest │             │
│ │  tool    │         │   tool   │         │  _fixes tool │             │
│ └────┬─────┘         └────┬─────┘         └──────┬───────┘             │
│      │                    │                      │                      │
│      ▼                    ▼                      ▼                      │
│ ┌──────────┐         ┌──────────┐         ┌──────────────┐             │
│ │ pkg/lint │         │pkg/judge │         │ NEW: pkg/fix │             │
│ │ (vacuum) │         │ (Claude) │         │  (remediate) │             │
│ └──────────┘         └──────────┘         └──────────────┘             │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────┐      │
│  │                      pkg/types                                 │      │
│  │  Violation{..., ExampleFix, RuleURL, Confidence, Related}      │      │
│  │  Rule{..., Generate{prompt, template, priority, examples}}     │      │
│  └───────────────────────────────────────────────────────────────┘      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Type Changes

### Violation Type Enhancement

**File:** `pkg/types/report.go`

```go
// Violation represents a single rule violation.
type Violation struct {
    // Existing fields (unchanged)
    RuleID    string   `json:"ruleId"`
    Severity  Severity `json:"severity"`
    Message   string   `json:"message"`
    Path      string   `json:"path"`
    Line      int      `json:"line,omitempty"`
    Column    int      `json:"column,omitempty"`
    EndLine   int      `json:"endLine,omitempty"`
    EndColumn int      `json:"endColumn,omitempty"`
    Suggestion string  `json:"suggestion,omitempty"`  // Already exists
    RuleTitle  string  `json:"ruleTitle,omitempty"`   // Already exists
    Category   string  `json:"category,omitempty"`    // Already exists

    // NEW: Enhanced remediation fields

    // ExampleFix shows a code snippet demonstrating the fix.
    // Format: YAML or JSON depending on input spec format.
    ExampleFix string `json:"exampleFix,omitempty"`

    // RuleURL links to the rule's documentation page.
    RuleURL string `json:"ruleUrl,omitempty"`

    // Confidence indicates certainty of this violation (0.0-1.0).
    // 1.0 = deterministic match, <1.0 = heuristic or LLM-based.
    Confidence float64 `json:"confidence,omitempty"`

    // RelatedRules lists rule IDs that should be addressed first
    // or are commonly fixed together with this violation.
    RelatedRules []string `json:"relatedRules,omitempty"`

    // FixPriority indicates the recommended fix order (1 = fix first).
    // Derived from rule priority and dependencies.
    FixPriority int `json:"fixPriority,omitempty"`
}
```

### Rule Type Enhancement

**File:** `pkg/types/rule.go`

```go
// Rule defines a single API style guideline.
type Rule struct {
    // Existing fields (unchanged)...
    ID          string           `json:"id"`
    Title       string           `json:"title"`
    Category    string           `json:"category"`
    Severity    Severity         `json:"severity"`
    Priority    int              `json:"priority,omitempty"`    // Already exists
    Relations   []RuleRelation   `json:"relations,omitempty"`   // Already exists
    Migration   *MigrationGuidance `json:"migration,omitempty"` // Already exists
    // ...

    // NEW: Generation guidance for AI agents
    Generate *GenerationGuidance `json:"generate,omitempty"`
}

// GenerationGuidance provides instructions for AI agents generating OpenAPI specs.
type GenerationGuidance struct {
    // Prompt is the instruction for an LLM when generating spec content.
    // Written as a positive directive (e.g., "Use plural nouns for collections").
    Prompt string `json:"prompt"`

    // Template is a URI or schema pattern to follow.
    // Variables use {placeholder} syntax.
    Template string `json:"template,omitempty"`

    // Priority determines generation order (100 = apply first, 1 = last).
    // Used to ensure foundational rules are followed before details.
    Priority int `json:"priority,omitempty"`

    // Examples show OpenAPI snippets demonstrating correct usage.
    Examples []GenerationExample `json:"examples,omitempty"`

    // Checklist provides bullet points to verify compliance.
    Checklist []string `json:"checklist,omitempty"`
}

// GenerationExample shows correct OpenAPI usage for a rule.
type GenerationExample struct {
    // Description explains what this example demonstrates.
    Description string `json:"description"`

    // OpenAPI is a YAML/JSON snippet showing correct usage.
    OpenAPI string `json:"openapi"`

    // Context explains when this pattern applies.
    Context string `json:"context,omitempty"`
}
```

### New Types for Fix Suggestions

**File:** `pkg/types/fix.go` (NEW)

```go
package types

// FixSuggestion represents a proposed fix for a violation.
type FixSuggestion struct {
    // RuleID is the rule this fix addresses.
    RuleID string `json:"ruleId"`

    // Path is the JSONPath to the element to fix.
    Path string `json:"path"`

    // CurrentValue is the existing value (may be empty for missing fields).
    CurrentValue string `json:"currentValue,omitempty"`

    // SuggestedValue is the proposed replacement.
    SuggestedValue string `json:"suggestedValue"`

    // Diff shows the change in unified diff format.
    Diff string `json:"diff,omitempty"`

    // Confidence indicates certainty that this fix is correct (0.0-1.0).
    Confidence float64 `json:"confidence"`

    // Reasoning explains why this fix is suggested.
    Reasoning string `json:"reasoning,omitempty"`

    // Breaking indicates if this fix could break existing clients.
    Breaking bool `json:"breaking,omitempty"`

    // BreakingReason explains why the fix is breaking.
    BreakingReason string `json:"breakingReason,omitempty"`
}

// FixReport contains all fix suggestions for a spec.
type FixReport struct {
    // Suggestions are the proposed fixes.
    Suggestions []FixSuggestion `json:"suggestions"`

    // PatchOperations are JSON Patch operations (RFC 6902).
    PatchOperations []PatchOperation `json:"patchOperations,omitempty"`

    // FixedCount is how many violations have suggestions.
    FixedCount int `json:"fixedCount"`

    // UnfixedCount is how many violations could not be auto-fixed.
    UnfixedCount int `json:"unfixedCount"`

    // UnfixedRules lists rules that couldn't be auto-fixed.
    UnfixedRules []string `json:"unfixedRules,omitempty"`
}

// PatchOperation represents a JSON Patch operation (RFC 6902).
type PatchOperation struct {
    Op    string `json:"op"`              // "add", "remove", "replace", "move", "copy"
    Path  string `json:"path"`            // JSON Pointer
    From  string `json:"from,omitempty"`  // For move/copy
    Value any    `json:"value,omitempty"` // For add/replace
}

// ConformancePath shows the path to reach a conformance level.
type ConformancePath struct {
    // CurrentLevel is the current conformance level (or "none").
    CurrentLevel string `json:"currentLevel"`

    // TargetLevel is the requested conformance level.
    TargetLevel string `json:"targetLevel"`

    // Blockers are errors that must be fixed to reach target.
    Blockers []ConformanceBlocker `json:"blockers"`

    // Warnings are issues that should be addressed.
    Warnings []ConformanceBlocker `json:"warnings,omitempty"`

    // ProgressToTarget is a percentage (0.0-1.0) of completion.
    ProgressToTarget float64 `json:"progressToTarget"`

    // EstimatedFixes is the count of changes needed.
    EstimatedFixes int `json:"estimatedFixes"`
}

// ConformanceBlocker describes a barrier to conformance.
type ConformanceBlocker struct {
    // RuleID is the blocking rule.
    RuleID string `json:"ruleId"`

    // Count is how many violations of this rule exist.
    Count int `json:"count"`

    // Priority is the fix order (1 = fix first).
    Priority int `json:"priority"`

    // FixInstructions provides guidance for resolution.
    FixInstructions string `json:"fixInstructions"`
}

// DesignCheck provides pre-generation guidance.
type DesignCheck struct {
    // Checklist is ordered rules to follow.
    Checklist []DesignCheckItem `json:"checklist"`

    // Template is an OpenAPI skeleton to use.
    Template map[string]any `json:"template,omitempty"`

    // Warnings are potential issues to consider.
    Warnings []string `json:"warnings,omitempty"`
}

// DesignCheckItem is a single design guidance item.
type DesignCheckItem struct {
    // RuleID is the relevant rule.
    RuleID string `json:"ruleId"`

    // Instruction is what to do.
    Instruction string `json:"instruction"`

    // Priority determines order (100 = first).
    Priority int `json:"priority"`

    // Required indicates if this is mandatory.
    Required bool `json:"required"`
}
```

## New Package: pkg/fix

**Purpose:** Generate fix suggestions for violations using rules and LLM.

### Interface

```go
// pkg/fix/fixer.go

package fix

import (
    "context"

    "github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Fixer generates fix suggestions for violations.
type Fixer interface {
    // SuggestFixes generates fix suggestions for the given violations.
    SuggestFixes(ctx context.Context, spec []byte, violations []types.Violation, opts *Options) (*types.FixReport, error)

    // DesignCheck provides pre-generation guidance.
    DesignCheck(ctx context.Context, resource string, operations []string, opts *Options) (*types.DesignCheck, error)

    // ConformancePath shows the path to a conformance level.
    ConformancePath(ctx context.Context, spec []byte, targetLevel string, opts *Options) (*types.ConformancePath, error)
}

// Options configures fix generation.
type Options struct {
    // Profile is the style profile to use.
    Profile string

    // MaxSuggestions limits the number of suggestions returned.
    MaxSuggestions int

    // IncludePatch generates JSON Patch operations.
    IncludePatch bool

    // UseLLM enables LLM-based fix generation for complex rules.
    UseLLM bool
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() *Options {
    return &Options{
        Profile:        "default",
        MaxSuggestions: 50,
        IncludePatch:   true,
        UseLLM:         true,
    }
}
```

### Implementation Strategy

```go
// pkg/fix/rule_fixer.go

package fix

import (
    "context"
    "fmt"
    "strings"

    "github.com/plexusone/systemspec-apistyle/pkg/types"
)

// RuleFixer generates fixes using rule metadata.
type RuleFixer struct {
    profile *types.APIStyleSpec
}

// NewRuleFixer creates a fixer from a style profile.
func NewRuleFixer(profile *types.APIStyleSpec) *RuleFixer {
    return &RuleFixer{profile: profile}
}

// SuggestFixes implements Fixer.
func (f *RuleFixer) SuggestFixes(ctx context.Context, spec []byte, violations []types.Violation, opts *Options) (*types.FixReport, error) {
    report := &types.FixReport{
        Suggestions: make([]types.FixSuggestion, 0, len(violations)),
    }

    for _, v := range violations {
        rule := f.findRule(v.RuleID)
        if rule == nil {
            report.UnfixedCount++
            report.UnfixedRules = append(report.UnfixedRules, v.RuleID)
            continue
        }

        suggestion, err := f.generateSuggestion(ctx, spec, v, rule, opts)
        if err != nil {
            report.UnfixedCount++
            continue
        }

        report.Suggestions = append(report.Suggestions, *suggestion)
        report.FixedCount++
    }

    if opts.IncludePatch {
        report.PatchOperations = f.generatePatch(report.Suggestions)
    }

    return report, nil
}

// generateSuggestion creates a fix suggestion for a single violation.
func (f *RuleFixer) generateSuggestion(ctx context.Context, spec []byte, v types.Violation, rule *types.Rule, opts *Options) (*types.FixSuggestion, error) {
    suggestion := &types.FixSuggestion{
        RuleID:     v.RuleID,
        Path:       v.Path,
        Confidence: 1.0,
    }

    // Try rule-based fix first
    if fix := f.tryRuleBasedFix(spec, v, rule); fix != nil {
        suggestion.SuggestedValue = fix.Value
        suggestion.Reasoning = fix.Reasoning
        return suggestion, nil
    }

    // Fall back to LLM if enabled
    if opts.UseLLM {
        return f.llmSuggestFix(ctx, spec, v, rule)
    }

    return nil, fmt.Errorf("no fix available for rule %s", v.RuleID)
}
```

## MCP Tool Additions

### suggest_fixes Tool

**File:** `cmd/mcp-api-style/tools.go` (extend)

```go
// Tool definition
var suggestFixesTool = mcp.Tool{
    Name:        "suggest_fixes",
    Description: "Generate fix suggestions for OpenAPI spec violations. Returns specific changes to make, with diffs and JSON Patch operations.",
    InputSchema: mcp.InputSchema{
        Type: "object",
        Properties: map[string]mcp.Property{
            "spec": {
                Type:        "string",
                Description: "The OpenAPI specification content (YAML or JSON)",
            },
            "violations": {
                Type:        "array",
                Description: "List of violations to fix. If empty, lints first and fixes all.",
                Items: &mcp.Property{
                    Type: "object",
                    Properties: map[string]mcp.Property{
                        "ruleId": {Type: "string"},
                        "path":   {Type: "string"},
                    },
                },
            },
            "profile": {
                Type:        "string",
                Description: "Style profile to use (default: 'default')",
            },
            "maxSuggestions": {
                Type:        "integer",
                Description: "Maximum suggestions to return (default: 50)",
            },
        },
        Required: []string{"spec"},
    },
}

// Handler
func handleSuggestFixes(ctx context.Context, args map[string]any) (*mcp.ToolResult, error) {
    spec := args["spec"].(string)
    profile := getStringOrDefault(args, "profile", "default")
    maxSuggestions := getIntOrDefault(args, "maxSuggestions", 50)

    // Load profile
    styleSpec, err := profile.Load(profile)
    if err != nil {
        return nil, fmt.Errorf("loading profile: %w", err)
    }

    // Get violations (from args or by linting)
    violations, err := getViolations(ctx, spec, args, styleSpec)
    if err != nil {
        return nil, err
    }

    // Generate fixes
    fixer := fix.NewRuleFixer(styleSpec)
    opts := &fix.Options{
        Profile:        profile,
        MaxSuggestions: maxSuggestions,
        IncludePatch:   true,
        UseLLM:         true,
    }

    report, err := fixer.SuggestFixes(ctx, []byte(spec), violations, opts)
    if err != nil {
        return nil, err
    }

    return &mcp.ToolResult{
        Content: []mcp.Content{{
            Type: "text",
            Text: mustJSON(report),
        }},
    }, nil
}
```

### design_check Tool

```go
var designCheckTool = mcp.Tool{
    Name:        "design_check",
    Description: "Get design guidance before generating an OpenAPI spec. Returns a checklist of rules to follow and a template.",
    InputSchema: mcp.InputSchema{
        Type: "object",
        Properties: map[string]mcp.Property{
            "resourceName": {
                Type:        "string",
                Description: "The resource being designed (e.g., 'user', 'order')",
            },
            "operations": {
                Type:        "array",
                Description: "Planned operations (e.g., ['list', 'create', 'get', 'update', 'delete'])",
                Items:       &mcp.Property{Type: "string"},
            },
            "profile": {
                Type:        "string",
                Description: "Style profile to check against",
            },
            "level": {
                Type:        "string",
                Description: "Target conformance level (bronze, silver, gold)",
            },
        },
        Required: []string{"resourceName", "profile"},
    },
}

func handleDesignCheck(ctx context.Context, args map[string]any) (*mcp.ToolResult, error) {
    resourceName := args["resourceName"].(string)
    profile := getStringOrDefault(args, "profile", "default")
    operations := getStringArrayOrDefault(args, "operations", []string{"list", "create", "get", "update", "delete"})
    level := getStringOrDefault(args, "level", "bronze")

    styleSpec, err := profile.Load(profile)
    if err != nil {
        return nil, err
    }

    fixer := fix.NewRuleFixer(styleSpec)
    opts := &fix.Options{Profile: profile}

    check, err := fixer.DesignCheck(ctx, resourceName, operations, opts)
    if err != nil {
        return nil, err
    }

    return &mcp.ToolResult{
        Content: []mcp.Content{{
            Type: "text",
            Text: mustJSON(check),
        }},
    }, nil
}
```

### conformance_path Tool

```go
var conformancePathTool = mcp.Tool{
    Name:        "conformance_path",
    Description: "Show the path to reach a conformance level. Returns blockers, warnings, and progress.",
    InputSchema: mcp.InputSchema{
        Type: "object",
        Properties: map[string]mcp.Property{
            "spec": {
                Type:        "string",
                Description: "The OpenAPI specification content",
            },
            "targetLevel": {
                Type:        "string",
                Enum:        []string{"bronze", "silver", "gold"},
                Description: "Target conformance level",
            },
            "profile": {
                Type:        "string",
                Description: "Style profile to evaluate against",
            },
        },
        Required: []string{"spec", "targetLevel", "profile"},
    },
}

func handleConformancePath(ctx context.Context, args map[string]any) (*mcp.ToolResult, error) {
    spec := args["spec"].(string)
    targetLevel := args["targetLevel"].(string)
    profile := getStringOrDefault(args, "profile", "default")

    styleSpec, err := profile.Load(profile)
    if err != nil {
        return nil, err
    }

    fixer := fix.NewRuleFixer(styleSpec)
    opts := &fix.Options{Profile: profile}

    path, err := fixer.ConformancePath(ctx, []byte(spec), targetLevel, opts)
    if err != nil {
        return nil, err
    }

    return &mcp.ToolResult{
        Content: []mcp.Content{{
            Type: "text",
            Text: mustJSON(path),
        }},
    }, nil
}
```

## MCP Resource Additions

### Exemplar Specs

```go
// Resource: apistyle://profiles/{name}/exemplars
func handleListExemplars(uri string) (*mcp.Resource, error) {
    profileName := extractProfileName(uri)

    exemplars, err := profile.ListExemplars(profileName)
    if err != nil {
        return nil, err
    }

    return &mcp.Resource{
        URI:      uri,
        MimeType: "application/json",
        Contents: mustJSON(exemplars),
    }, nil
}

// Resource: apistyle://profiles/{name}/exemplars/{id}
func handleGetExemplar(uri string) (*mcp.Resource, error) {
    profileName, exemplarID := extractProfileAndExemplar(uri)

    content, err := profile.GetExemplar(profileName, exemplarID)
    if err != nil {
        return nil, err
    }

    return &mcp.Resource{
        URI:      uri,
        MimeType: "application/x-yaml",
        Contents: content,
    }, nil
}

// Resource: apistyle://profiles/{name}/patterns
func handleListPatterns(uri string) (*mcp.Resource, error) {
    profileName := extractProfileName(uri)

    patterns, err := profile.ListPatterns(profileName)
    if err != nil {
        return nil, err
    }

    return &mcp.Resource{
        URI:      uri,
        MimeType: "application/json",
        Contents: mustJSON(patterns),
    }, nil
}
```

## Enrichment in Lint Pipeline

Populate new Violation fields during linting:

```go
// pkg/lint/results.go

// enrichViolation adds remediation metadata to a violation.
func enrichViolation(v *types.Violation, rule *types.Rule, profile *types.APIStyleSpec) {
    // Add rule metadata
    v.RuleTitle = rule.Title
    v.Category = rule.Category

    // Generate rule URL
    v.RuleURL = fmt.Sprintf("https://plexusone.dev/systemspec-apistyle/profiles/%s#%s",
        strings.ToLower(profile.Name),
        strings.ToLower(rule.ID))

    // Add suggestion from rule examples or migration guidance
    if v.Suggestion == "" && rule.Migration != nil {
        v.Suggestion = rule.Migration.Steps[0]
    }

    // Add example fix from rule examples
    if rule.Examples != nil && len(rule.Examples.Good) > 0 {
        v.ExampleFix = generateExampleFix(v, rule.Examples.Good[0])
    }

    // Set confidence (1.0 for deterministic, lower for heuristics)
    v.Confidence = 1.0

    // Find related rules from dependencies
    for _, rel := range rule.Relations {
        if rel.Type == "depends-on" || rel.Type == "conflicts-with" {
            v.RelatedRules = append(v.RelatedRules, rel.RuleID)
        }
    }

    // Calculate fix priority from rule priority
    v.FixPriority = 100 - rule.Priority // Invert: low priority number = high fix priority
}
```

## File Structure

```
systemspec-apistyle/
├── pkg/
│   ├── types/
│   │   ├── report.go      # Enhanced Violation
│   │   ├── rule.go        # Add GenerationGuidance
│   │   └── fix.go         # NEW: FixSuggestion, ConformancePath, etc.
│   │
│   ├── fix/               # NEW PACKAGE
│   │   ├── fixer.go       # Fixer interface
│   │   ├── rule_fixer.go  # Rule-based fix generation
│   │   ├── llm_fixer.go   # LLM-based fix generation
│   │   ├── patch.go       # JSON Patch generation
│   │   └── templates.go   # Pattern templates
│   │
│   ├── lint/
│   │   └── results.go     # Add enrichViolation()
│   │
│   └── profile/
│       ├── exemplars.go   # NEW: Exemplar loading
│       └── patterns.go    # NEW: Pattern loading
│
├── cmd/
│   ├── api-style/
│   │   └── suggest.go     # NEW: suggest-fixes command
│   │
│   └── mcp-api-style/
│       └── tools.go       # Add new tools
│
├── profiles/
│   ├── default/
│   │   ├── profile.json
│   │   └── exemplars/     # NEW
│   │       ├── minimal.yaml
│   │       └── crud-api.yaml
│   └── azure/
│       └── exemplars/
│
└── docs/
    └── specs/
        └── ai-agent-support/
            ├── PRD.md
            ├── TRD.md
            ├── PLAN.md
            └── ROADMAP.md
```

## Testing Strategy

### Unit Tests

```go
// pkg/fix/fixer_test.go

func TestRuleFixer_SuggestFixes_PluralResources(t *testing.T) {
    profile := &types.APIStyleSpec{
        Rules: []types.Rule{{
            ID:       "URI-001",
            Title:    "Use plural resources",
            Severity: types.SeverityError,
            Examples: &types.Examples{
                Good: []string{"/users", "/orders"},
                Bad:  []string{"/user", "/order"},
            },
            Migration: &types.MigrationGuidance{
                Steps: []string{"Rename singular path to plural form"},
            },
        }},
    }

    fixer := NewRuleFixer(profile)

    violations := []types.Violation{{
        RuleID: "URI-001",
        Path:   "$.paths./user",
    }}

    spec := []byte(`paths:
  /user:
    get:
      summary: Get user`)

    report, err := fixer.SuggestFixes(context.Background(), spec, violations, DefaultOptions())
    require.NoError(t, err)

    require.Len(t, report.Suggestions, 1)
    assert.Equal(t, "URI-001", report.Suggestions[0].RuleID)
    assert.Contains(t, report.Suggestions[0].SuggestedValue, "/users")
}
```

### Integration Tests

```go
// integration/suggest_fixes_test.go

func TestMCP_SuggestFixes(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    server := startMCPServer(t)
    defer server.Close()

    result, err := server.CallTool("suggest_fixes", map[string]any{
        "spec": testOpenAPISpec,
        "profile": "default",
    })

    require.NoError(t, err)

    var report types.FixReport
    require.NoError(t, json.Unmarshal([]byte(result.Content[0].Text), &report))

    assert.Greater(t, report.FixedCount, 0)
}
```

## Performance Requirements

| Operation | Target | Measurement |
|-----------|--------|-------------|
| SuggestFixes (rule-based, 50 violations) | <500ms | Benchmark |
| SuggestFixes (LLM, 10 violations) | <10s | Integration |
| DesignCheck | <200ms | Benchmark |
| ConformancePath | <1s | Benchmark |
| Violation enrichment per item | <1ms | Benchmark |

## Security Considerations

1. **LLM API Keys** - Never logged, use environment variables
2. **Spec Content** - Treat as potentially sensitive, don't log full content
3. **Suggested Fixes** - Mark as "suggestions" requiring human review
4. **Patch Operations** - Validate before applying to prevent injection
