# Technical Requirements Document (TRD)

## systemspec-apistyle

**Version:** 0.1.0-draft
**Date:** 2026-06-03
**Status:** Draft

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        systemspec-apistyle                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │   CLI        │    │   Web UI     │    │  MCP Server  │      │
│  │  (cobra)     │    │   (Lit)      │    │  (mcp-go)    │      │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘      │
│         │                   │                   │               │
│         └───────────────────┼───────────────────┘               │
│                             │                                   │
│                    ┌────────▼────────┐                          │
│                    │   pkg/analyze   │                          │
│                    │  (orchestrator) │                          │
│                    └────────┬────────┘                          │
│                             │                                   │
│         ┌───────────────────┼───────────────────┐               │
│         │                   │                   │               │
│  ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐        │
│  │  pkg/lint   │    │  pkg/judge  │    │ pkg/generate│        │
│  │  (vacuum)   │    │ (str-eval)  │    │  (outputs)  │        │
│  └──────┬──────┘    └──────┬──────┘    └─────────────┘        │
│         │                   │                                   │
│  ┌──────▼──────┐    ┌──────▼──────┐                            │
│  │   vacuum    │    │  anthropic  │                            │
│  │  (library)  │    │    SDK      │                            │
│  └─────────────┘    └─────────────┘                            │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                      pkg/types                           │   │
│  │  (Go structs → JSON Schema source of truth)              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
systemspec-apistyle/
├── pkg/
│   ├── types/                 # Core types (schema source of truth)
│   │   ├── spec.go           # APIStyleSpec root type
│   │   ├── rule.go           # Rule, Enforcement, Judge
│   │   ├── profile.go        # Profile, Extends, Overrides
│   │   ├── lexicon.go        # Lexicon, Term
│   │   ├── conformance.go    # ConformanceLevel, Criteria
│   │   ├── exception.go      # Exception, Scope
│   │   ├── report.go         # LintReport, Violation
│   │   └── generate.go       # //go:generate for schema
│   │
│   ├── lint/                  # Deterministic linting
│   │   ├── linter.go         # Linter interface
│   │   ├── vacuum.go         # vacuum integration
│   │   ├── ruleset.go        # Convert spec → vacuum ruleset
│   │   └── report.go         # Convert vacuum → our report
│   │
│   ├── judge/                 # LLM evaluation
│   │   ├── evaluator.go      # Evaluator interface
│   │   ├── anthropic.go      # Claude implementation
│   │   ├── rubric.go         # Build rubric from spec
│   │   └── prompt.go         # Prompt templates
│   │
│   ├── analyze/               # Combined orchestration
│   │   ├── analyzer.go       # Orchestrates lint + judge
│   │   ├── result.go         # Combined result type
│   │   └── options.go        # Analysis options
│   │
│   ├── generate/              # Output generators
│   │   ├── generator.go      # Generator interface
│   │   ├── markdown.go       # Human guide
│   │   ├── spectral.go       # Spectral YAML
│   │   ├── agent.go          # Agent instructions
│   │   └── prompt.go         # LLM prompts
│   │
│   ├── profile/               # Profile management
│   │   ├── loader.go         # Load built-in/custom profiles
│   │   ├── resolver.go       # Resolve extends/overrides
│   │   └── builtin/          # Embedded profiles
│   │       ├── default.json
│   │       ├── azure.json
│   │       ├── google.json
│   │       └── zalando.json
│   │
│   ├── diff/                  # Breaking change detection
│   │   ├── differ.go         # Diff interface
│   │   ├── openapi.go        # OpenAPI comparison
│   │   └── report.go         # CompatibilityReport
│   │
│   └── store/                 # Evaluation storage (optional)
│       ├── store.go          # Store interface
│       ├── sqlite.go         # SQLite implementation
│       └── postgres.go       # PostgreSQL implementation
│
├── schema/                    # Generated JSON schemas
│   ├── embed.go              # //go:embed
│   ├── api-style-spec.schema.json
│   ├── lint-report.schema.json
│   └── evaluation-report.schema.json
│
├── cmd/
│   └── api-style/
│       ├── main.go
│       ├── lint.go
│       ├── evaluate.go
│       ├── analyze.go
│       ├── generate.go
│       ├── diff.go
│       ├── serve.go
│       ├── profile.go
│       └── history.go
│
├── mcp/
│   ├── server.go             # MCP server setup
│   ├── tools.go              # Tool definitions
│   └── handlers.go           # Tool handlers
│
├── web/
│   ├── packages/
│   │   └── api-style-analyzer/
│   │       ├── src/
│   │       │   ├── components/
│   │       │   │   ├── api-style-analyzer.ts
│   │       │   │   ├── asa-url-input.ts
│   │       │   │   ├── asa-evaluation-report.ts
│   │       │   │   ├── asa-style-guide.ts
│   │       │   │   └── asa-rule-browser.ts
│   │       │   ├── utils/
│   │       │   │   ├── api.ts
│   │       │   │   └── export.ts
│   │       │   └── types.ts
│   │       ├── package.json
│   │       └── vite.config.ts
│   └── server/               # Optional: Go server for web UI
│       └── server.go
│
├── agents/
│   ├── specs/                # multi-agent-spec definitions
│   │   ├── api-style-reviewer.md
│   │   └── api-style-guide-generator.md
│   └── plugins/              # Generated by assistantkit
│       ├── claude/
│       └── kiro/
│
├── profiles/                  # User-facing profile examples
│   ├── azure.api-style.json
│   ├── google.api-style.json
│   └── zalando.api-style.json
│
├── examples/
│   ├── petstore/
│   │   ├── openapi.yaml
│   │   └── expected-report.json
│   └── custom-profile/
│       └── acme.api-style.json
│
└── docs/
    ├── specs/
    │   ├── CONSTITUTION.md
    │   └── inception/
    │       ├── MRD.md
    │       ├── PRD.md
    │       ├── TRD.md
    │       └── ROADMAP.md
    └── guide/
        ├── getting-started.md
        ├── profiles.md
        └── custom-rules.md
```

## Core Types

### APIStyleSpec (Root)

```go
// pkg/types/spec.go

// APIStyleSpec is the root type for an API style specification.
// This is the source of truth from which JSON Schema is generated.
type APIStyleSpec struct {
    // Schema is the JSON Schema URI for validation
    Schema string `json:"$schema,omitempty"`

    // Version is the semantic version of this spec
    Version string `json:"version"`

    // Name is a unique identifier for this style spec
    Name string `json:"name"`

    // Description provides context about this style spec
    Description string `json:"description,omitempty"`

    // Extends lists parent profiles to inherit from
    Extends []string `json:"extends,omitempty"`

    // Rules are the style rules defined in this spec
    Rules []Rule `json:"rules"`

    // Overrides modify inherited rules
    Overrides map[string]RuleOverride `json:"overrides,omitempty"`

    // Lexicon defines approved/forbidden terminology
    Lexicon *Lexicon `json:"lexicon,omitempty"`

    // ConformanceLevels define graduated compliance tiers
    ConformanceLevels map[string]ConformanceLevel `json:"conformanceLevels,omitempty"`

    // Exceptions are approved rule waivers
    Exceptions []Exception `json:"exceptions,omitempty"`

    // Metadata contains additional information
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

### Rule

```go
// pkg/types/rule.go

// Rule defines a single style guideline.
type Rule struct {
    // ID is a unique identifier (e.g., "URI-001")
    ID string `json:"id"`

    // Title is a short description
    Title string `json:"title"`

    // Category groups related rules
    Category string `json:"category"`

    // Severity is the violation level: error, warn, info, hint
    Severity Severity `json:"severity"`

    // Scope defines what part of the spec this applies to
    Scope Scope `json:"scope,omitempty"`

    // Rationale explains why this rule exists
    Rationale string `json:"rationale,omitempty"`

    // Examples show good and bad usage
    Examples *Examples `json:"examples,omitempty"`

    // Enforcement defines deterministic checking
    Enforcement *Enforcement `json:"enforcement,omitempty"`

    // Judge defines LLM evaluation criteria
    Judge *JudgeCriteria `json:"judge,omitempty"`

    // References link to external documentation
    References []Reference `json:"references,omitempty"`

    // Tags for filtering and grouping
    Tags []string `json:"tags,omitempty"`
}

// Severity levels for rule violations.
type Severity string

const (
    SeverityError Severity = "error"
    SeverityWarn  Severity = "warn"
    SeverityInfo  Severity = "info"
    SeverityHint  Severity = "hint"
)

// Scope defines what part of the spec a rule applies to.
type Scope string

const (
    ScopePath       Scope = "path"
    ScopeOperation  Scope = "operation"
    ScopeParameter  Scope = "parameter"
    ScopeSchema     Scope = "schema"
    ScopeResponse   Scope = "response"
    ScopeInfo       Scope = "info"
    ScopeSecurity   Scope = "security"
    ScopeGlobal     Scope = "global"
)

// Examples provides good and bad usage patterns.
type Examples struct {
    Good []string `json:"good,omitempty"`
    Bad  []string `json:"bad,omitempty"`
}

// Enforcement defines deterministic rule checking.
type Enforcement struct {
    // Type is the enforcement mechanism: spectral, custom, regex
    Type EnforcementType `json:"type"`

    // Function is the Spectral function name (if type=spectral)
    Function string `json:"function,omitempty"`

    // Options are function-specific options
    Options map[string]interface{} `json:"options,omitempty"`

    // Given is the JSONPath for targeting (Spectral)
    Given string `json:"given,omitempty"`

    // Then is the assertion (Spectral)
    Then *SpectralThen `json:"then,omitempty"`
}

// EnforcementType defines how a rule is enforced.
type EnforcementType string

const (
    EnforcementSpectral EnforcementType = "spectral"
    EnforcementCustom   EnforcementType = "custom"
    EnforcementRegex    EnforcementType = "regex"
    EnforcementNone     EnforcementType = "none" // LLM-only
)

// JudgeCriteria defines LLM evaluation parameters.
type JudgeCriteria struct {
    // Prompt is the evaluation instruction
    Prompt string `json:"prompt"`

    // Weight influences scoring (0.0-1.0)
    Weight float64 `json:"weight,omitempty"`

    // RequiresContext indicates if broader context is needed
    RequiresContext bool `json:"requiresContext,omitempty"`
}

// Reference links to external documentation.
type Reference struct {
    Title string `json:"title"`
    URL   string `json:"url"`
}
```

### LintReport

```go
// pkg/types/report.go

// LintReport contains deterministic linting results.
type LintReport struct {
    // Status is the overall result: pass, fail
    Status ReportStatus `json:"status"`

    // ConformanceLevel is the highest level achieved
    ConformanceLevel string `json:"conformanceLevel,omitempty"`

    // Summary provides violation counts
    Summary *ViolationSummary `json:"summary"`

    // Violations lists all findings
    Violations []Violation `json:"violations"`

    // Metadata includes timing, versions, etc.
    Metadata *ReportMetadata `json:"metadata,omitempty"`
}

// ReportStatus indicates pass/fail.
type ReportStatus string

const (
    ReportStatusPass ReportStatus = "pass"
    ReportStatusFail ReportStatus = "fail"
)

// ViolationSummary counts violations by severity.
type ViolationSummary struct {
    Errors   int `json:"errors"`
    Warnings int `json:"warnings"`
    Infos    int `json:"infos"`
    Hints    int `json:"hints"`
}

// Violation represents a single rule violation.
type Violation struct {
    // RuleID is the violated rule
    RuleID string `json:"ruleId"`

    // Severity from the rule
    Severity Severity `json:"severity"`

    // Message describes the violation
    Message string `json:"message"`

    // Path is the JSONPath to the violation
    Path string `json:"path"`

    // Line number in the source file
    Line int `json:"line,omitempty"`

    // Column in the source file
    Column int `json:"column,omitempty"`

    // Suggestion for fixing
    Suggestion string `json:"suggestion,omitempty"`
}
```

## Integration Points

### vacuum Integration

```go
// pkg/lint/vacuum.go

package lint

import (
    "github.com/daveshanley/vacuum/model"
    "github.com/daveshanley/vacuum/motor"
    "github.com/daveshanley/vacuum/rulesets"

    "github.com/plexusone/systemspec-apistyle/pkg/types"
)

// VacuumLinter implements Linter using vacuum.
type VacuumLinter struct {
    spec *types.APIStyleSpec
}

// NewVacuumLinter creates a linter from an systemspec-apistyle.
func NewVacuumLinter(spec *types.APIStyleSpec) *VacuumLinter {
    return &VacuumLinter{spec: spec}
}

// Lint executes linting against an OpenAPI specification.
func (l *VacuumLinter) Lint(specBytes []byte, opts *LintOptions) (*types.LintReport, error) {
    // Convert systemspec-apistyle rules to vacuum ruleset
    rs := l.buildRuleSet()

    // Execute vacuum
    result := motor.ApplyRulesToRuleSet(&motor.RuleSetExecution{
        RuleSet:      rs,
        Spec:         specBytes,
        SpecFileName: opts.FileName,
    })

    // Convert vacuum results to our report format
    return l.convertResult(result)
}

// buildRuleSet converts systemspec-apistyle to vacuum RuleSet.
func (l *VacuumLinter) buildRuleSet() *rulesets.RuleSet {
    rs := &rulesets.RuleSet{
        Rules: make(map[string]*model.Rule),
    }

    for _, rule := range l.spec.Rules {
        if rule.Enforcement == nil || rule.Enforcement.Type != types.EnforcementSpectral {
            continue
        }

        rs.Rules[rule.ID] = &model.Rule{
            Id:          rule.ID,
            Description: rule.Title,
            Severity:    string(rule.Severity),
            Given:       rule.Enforcement.Given,
            Then: &model.RuleAction{
                Function:        rule.Enforcement.Function,
                FunctionOptions: rule.Enforcement.Options,
            },
        }
    }

    return rs
}
```

### structured-evaluation Integration

```go
// pkg/judge/evaluator.go

package judge

import (
    "context"

    "github.com/plexusone/structured-evaluation/rubric"

    "github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Evaluator performs LLM-based evaluation.
type Evaluator struct {
    spec   *types.APIStyleSpec
    client LLMClient
}

// Evaluate runs LLM evaluation on an OpenAPI spec.
func (e *Evaluator) Evaluate(ctx context.Context, specContent string, opts *EvaluateOptions) (*rubric.Rubric, error) {
    // Build rubric from systemspec-apistyle
    rubricSet := e.buildRubricSet()

    // Create evaluation report
    report := rubric.NewRubric("api-style-evaluation", opts.FileName)

    // Evaluate each category
    for _, category := range e.getCategories(opts.Categories) {
        result, err := e.evaluateCategory(ctx, specContent, category)
        if err != nil {
            return nil, err
        }
        report.AddCategoryResult(result)
    }

    // Finalize with decision
    report.Finalize(rubricSet, e.getRerunCommand(opts))

    return report, nil
}

// buildRubricSet creates a RubricSet from systemspec-apistyle rules.
func (e *Evaluator) buildRubricSet() *rubric.RubricSet {
    categories := make(map[string]*rubric.Category)

    for _, rule := range e.spec.Rules {
        if rule.Judge == nil {
            continue
        }

        cat, exists := categories[rule.Category]
        if !exists {
            cat = &rubric.Category{
                Name: rule.Category,
            }
            categories[rule.Category] = cat
        }

        // Add rule criteria to category
        cat.Criteria = append(cat.Criteria, rubric.Criterion{
            ID:     rule.ID,
            Title:  rule.Title,
            Weight: rule.Judge.Weight,
        })
    }

    return &rubric.RubricSet{
        Categories: categories,
    }
}
```

### assistantkit Integration

```go
// pkg/generate/agent.go

package generate

import (
    "github.com/plexusone/assistantkit/agents/core"
    multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"

    "github.com/plexusone/systemspec-apistyle/pkg/types"
)

// AgentGenerator creates AI agent definitions.
type AgentGenerator struct {
    spec *types.APIStyleSpec
}

// GenerateAgent creates a multi-agent-spec Agent definition.
func (g *AgentGenerator) GenerateAgent() *multiagentspec.Agent {
    instructions := g.buildInstructions()

    return &multiagentspec.Agent{
        Name:        "api-style-reviewer",
        Namespace:   "api-style",
        Description: "Reviews OpenAPI specifications against " + g.spec.Name,
        Model:       multiagentspec.ModelSonnet,
        Tools:       []string{"Read", "Bash"},
        Skills:      []string{"api-style-analysis"},
        Dependencies: []string{"api-style"},
        Instructions: instructions,
    }
}

// buildInstructions generates agent instructions from spec.
func (g *AgentGenerator) buildInstructions() string {
    var b strings.Builder

    b.WriteString("You are an API style reviewer.\n\n")
    b.WriteString("## Style Rules\n\n")

    for _, rule := range g.spec.Rules {
        b.WriteString(fmt.Sprintf("### %s: %s\n\n", rule.ID, rule.Title))
        b.WriteString(fmt.Sprintf("**Severity:** %s\n\n", rule.Severity))

        if rule.Rationale != "" {
            b.WriteString(rule.Rationale + "\n\n")
        }

        if rule.Examples != nil {
            if len(rule.Examples.Good) > 0 {
                b.WriteString("**Good:**\n")
                for _, ex := range rule.Examples.Good {
                    b.WriteString(fmt.Sprintf("- `%s`\n", ex))
                }
                b.WriteString("\n")
            }
            if len(rule.Examples.Bad) > 0 {
                b.WriteString("**Bad:**\n")
                for _, ex := range rule.Examples.Bad {
                    b.WriteString(fmt.Sprintf("- `%s`\n", ex))
                }
                b.WriteString("\n")
            }
        }
    }

    return b.String()
}
```

## API Design (REST - If Needed)

If a REST API is added for the web UI backend:

```go
// web/server/server.go

package server

import (
    "github.com/danielgtaylor/huma/v2"
    "github.com/danielgtaylor/huma/v2/adapters/humachi"
    "github.com/go-chi/chi/v5"
)

func NewServer() *chi.Mux {
    r := chi.NewRouter()
    api := humachi.New(r, huma.DefaultConfig("systemspec-apistyle", "1.0.0"))

    // Register endpoints
    huma.Register(api, huma.Operation{
        OperationID: "lint",
        Method:      "POST",
        Path:        "/api/v1/lint",
        Summary:     "Lint an OpenAPI specification",
    }, handleLint)

    huma.Register(api, huma.Operation{
        OperationID: "evaluate",
        Method:      "POST",
        Path:        "/api/v1/evaluate",
        Summary:     "Evaluate an OpenAPI specification with LLM",
    }, handleEvaluate)

    huma.Register(api, huma.Operation{
        OperationID: "analyze",
        Method:      "POST",
        Path:        "/api/v1/analyze",
        Summary:     "Combined lint and evaluate",
    }, handleAnalyze)

    return r
}
```

## Database Schema (If Needed)

If evaluation storage is added:

```go
// pkg/store/schema/evaluation.go

package schema

import (
    "time"

    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

// Evaluation holds evaluation history.
type Evaluation struct {
    ent.Schema
}

func (Evaluation) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").Unique(),
        field.String("api_name"),
        field.String("spec_version").Optional(),
        field.String("profile"),
        field.String("conformance_level").Optional(),

        // Lint results
        field.Int("lint_errors"),
        field.Int("lint_warnings"),
        field.Int("lint_infos"),

        // LLM results
        field.String("llm_decision").Optional(),
        field.JSON("llm_scores", map[string]interface{}{}).Optional(),

        // Combined
        field.String("overall_status"),

        // Full reports
        field.JSON("lint_report", map[string]interface{}{}),
        field.JSON("evaluation_report", map[string]interface{}{}).Optional(),

        // Metadata
        field.Time("evaluated_at").Default(time.Now),
        field.String("ci_run_id").Optional(),
    }
}

func (Evaluation) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("api_name", "evaluated_at"),
        index.Fields("profile"),
    }
}
```

## Frontend Types (Zod)

Generated from OpenAPI spec via `openapi-zod-client` or similar:

```typescript
// web/packages/api-style-analyzer/src/types.ts

import { z } from 'zod';

// Generated from OpenAPI spec
export const ViolationSchema = z.object({
  ruleId: z.string(),
  severity: z.enum(['error', 'warn', 'info', 'hint']),
  message: z.string(),
  path: z.string(),
  line: z.number().optional(),
  suggestion: z.string().optional(),
});

export const LintReportSchema = z.object({
  status: z.enum(['pass', 'fail']),
  conformanceLevel: z.string().optional(),
  summary: z.object({
    errors: z.number(),
    warnings: z.number(),
    infos: z.number(),
    hints: z.number(),
  }),
  violations: z.array(ViolationSchema),
});

export type Violation = z.infer<typeof ViolationSchema>;
export type LintReport = z.infer<typeof LintReportSchema>;
```

## Testing Strategy

### Unit Tests

```go
// pkg/lint/vacuum_test.go

func TestVacuumLinter_PluralResources(t *testing.T) {
    spec := &types.APIStyleSpec{
        Rules: []types.Rule{
            {
                ID:       "URI-001",
                Title:    "Use plural resources",
                Severity: types.SeverityError,
                Enforcement: &types.Enforcement{
                    Type:     types.EnforcementSpectral,
                    Function: "pattern",
                    Given:    "$.paths",
                    Options:  map[string]interface{}{"match": "^/[a-z]+s(/|$)"},
                },
            },
        },
    }

    linter := NewVacuumLinter(spec)

    openAPISpec := []byte(`
openapi: "3.0.0"
paths:
  /user:
    get:
      summary: Get user
`)

    report, err := linter.Lint(openAPISpec, &LintOptions{})
    require.NoError(t, err)

    assert.Equal(t, types.ReportStatusFail, report.Status)
    assert.Len(t, report.Violations, 1)
    assert.Equal(t, "URI-001", report.Violations[0].RuleID)
}
```

### Golden File Tests

```go
// pkg/generate/markdown_test.go

func TestMarkdownGenerator_Golden(t *testing.T) {
    spec := loadTestSpec(t, "testdata/azure.api-style.json")
    gen := NewMarkdownGenerator(spec)

    result, err := gen.Generate()
    require.NoError(t, err)

    golden.Assert(t, result, "testdata/azure-guide.golden.md")
}
```

### Integration Tests

```go
// integration/analyze_test.go

func TestAnalyze_EndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Requires ANTHROPIC_API_KEY
    ctx := context.Background()

    spec := loadTestSpec(t, "testdata/default.api-style.json")
    openAPI := loadFile(t, "testdata/petstore.yaml")

    analyzer := analyze.New(spec, analyze.WithLLM(true))
    result, err := analyzer.Analyze(ctx, openAPI)

    require.NoError(t, err)
    assert.NotNil(t, result.LintReport)
    assert.NotNil(t, result.EvaluationReport)
}
```

## Performance Requirements

| Operation | Target | Measurement |
|-----------|--------|-------------|
| Lint (1MB spec) | <500ms | Benchmark |
| Lint (10MB spec) | <5s | Benchmark |
| LLM Evaluation | <30s | Integration test |
| Guide Generation | <1s | Benchmark |
| Spectral Generation | <100ms | Benchmark |

## Security Considerations

1. **API Keys** - LLM API keys via environment variables, never logged
2. **Web UI** - Client-side processing only, no server storage of specs
3. **Input Validation** - All inputs validated via JSON Schema
4. **Dependencies** - Regular vulnerability scanning via `govulncheck`
