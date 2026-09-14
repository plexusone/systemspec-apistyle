# Profile Scoring & Coverage

systemspec-apistyle includes tools to evaluate the quality and coverage of API style guides using LLM-as-a-Judge.

## Overview

Two rubrics are available for evaluating style guides:

| Rubric | Purpose | Categories |
|--------|---------|------------|
| **Style Guide Quality** | Evaluates structure, clarity, examples | 9 |
| **API Coverage** | Evaluates topic completeness | 8 |

## Scoring a Profile

Use the `score-profile` command to evaluate any profile:

```bash
# Score a built-in profile
api-style score-profile comprehensive

# Score with JSON output
api-style score-profile zalando --format json

# Score a custom profile file
api-style score-profile ./my-profile.json

# Use a specific model
api-style score-profile microsoft-rest --model claude-sonnet-4
```

## Quality Rubric Categories

The **Style Guide Quality Rubric** evaluates how well a style guide is written:

| Category | Weight | What It Measures |
|----------|--------|------------------|
| Content Coverage | 1.0 | Covers 6 essential domains |
| Rule Quality & Clarity | 1.0 | Clear, actionable rules |
| Structure & Navigation | 0.9 | Organization and findability |
| Internal Consistency | 0.9 | No contradictions |
| Completeness & Depth | 0.8 | Edge cases, complex scenarios |
| Examples & Code Samples | 0.8 | Good/bad examples |
| Enforceability & Tooling | 0.7 | Linting support |
| Guide Versioning | 0.6 | Changelog, deprecation |
| Accessibility & Tone | 0.5 | Appropriate for audience |

## Coverage Rubric Categories

The **API Coverage Rubric** evaluates what topics are addressed:

| Category | Weight | Required | Topics |
|----------|--------|----------|--------|
| Fundamentals | 1.0 | Yes | general, naming, urls, http-methods, http-status |
| Request/Response | 0.9 | Yes | request-response, headers, errors, json, schema |
| Collections & Querying | 0.8 | Yes | pagination, filtering, collections |
| API Lifecycle | 0.8 | Yes | versioning, compatibility, deprecation |
| Security | 1.0 | Yes | authentication, authorization, HTTPS |
| Documentation | 0.7 | No | OpenAPI requirements, examples |
| Advanced Patterns | 0.6 | No | long-running, conditional, batch, actions |
| Ecosystem | 0.5 | No | events, hypermedia, performance, throttling |

## Scoring Levels

### Quality Score

| Level | Score | Meaning |
|-------|-------|---------|
| Excellent | 4.5+ | Best-in-class guide |
| Good | 3.5-4.4 | Production-ready |
| Adequate | 2.5-3.4 | Usable with gaps |
| Minimal | 1.5-2.4 | Basic only |
| Insufficient | <1.5 | Major gaps |

### Coverage Score

| Level | Score | Meaning |
|-------|-------|---------|
| Comprehensive | 90%+ | All topics covered |
| Strong | 70-89% | Most topics covered |
| Moderate | 50-69% | Core topics covered |
| Limited | 30-49% | Significant gaps |
| Minimal | <30% | Major gaps |

## Current Profile Scores

Based on LLM evaluation:

| Profile | Quality | Coverage |
|---------|---------|----------|
| comprehensive | 4.8 | 100% |
| default | 4.7 | 100% |
| zalando | 4.7 | 75% |
| microsoft-rest | 4.7 | 70% |
| microsoft-graph | 4.2 | 40% |
| azure | 3.8 | 45% |
| google | 3.5 | 35% |
| minimal | 3.0 | 30% |

## Improving Coverage

To identify gaps in your profile:

```bash
# Get detailed scoring report
api-style score-profile my-profile --format json > report.json

# View categories with low scores
jq '.categories[] | select(.numericScore < 4)' report.json
```

### Common Gaps

Most profiles are missing:

1. **Events/Webhooks** - Only Zalando covers thoroughly
2. **Deprecation** - Often overlooked
3. **Conditional Requests** - ETags, optimistic concurrency
4. **Batch Operations** - Bulk create/update patterns
5. **Performance** - Caching, compression guidance

### Adding Missing Coverage

Extend your profile with custom rules:

```yaml
# custom-rules.yaml
rules:
  - id: CUSTOM-EVENT-001
    title: Use CloudEvents format
    category: events
    severity: warn
    description: Webhook payloads SHOULD follow CloudEvents spec.
```

## Programmatic Scoring

The scoring rubrics are JSON files that can be used programmatically:

```go
import "github.com/plexusone/systemspec-apistyle/pkg/judge"

// Load quality rubric
rubric, _ := judge.LoadStyleGuideRubric()

// Create evaluator
evaluator, _ := judge.NewStyleGuideEvaluator(provider)

// Score a profile
report, _ := evaluator.Evaluate(ctx, spec, nil)
fmt.Printf("Score: %.1f/5.0\n", report.OverallScore)
```

## Rubric Files

The rubric JSON files are located in `schema/`:

| File | Purpose |
|------|---------|
| `style-guide-quality.rubric.json` | Quality evaluation |
| `api-coverage.rubric.json` | Coverage evaluation |

These follow the [structured-evaluation](https://github.com/grokify/structured-evaluation) rubric schema.

## See Also

- [Profile Comparison](profile-comparison.md) - Compare all profiles
- [Using Profiles](profiles.md) - Basic profile usage
- [Custom Rules](custom-rules.md) - Extend profiles
