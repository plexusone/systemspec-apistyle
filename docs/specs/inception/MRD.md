# Market Requirements Document (MRD)

## systemspec-apistyle

**Version:** 0.1.0-draft
**Date:** 2026-06-03
**Status:** Draft

## Executive Summary

API style guides from Microsoft, Google, and Zalando have become industry standards, but they exist only as human-readable documents. Organizations struggle to enforce these guidelines consistently across teams, tools, and AI agents. **systemspec-apistyle** creates a machine-readable specification format that serves as the single source of truth, generating human documentation, linting rules, LLM evaluation rubrics, and agent instructions from one canonical definition.

## Market Context

### The Problem

1. **Fragmented Enforcement** - Style guides exist as Markdown/HTML documents separate from the Spectral rules that partially enforce them
2. **Incomplete Coverage** - Spectral can only lint ~40% of typical style guide rules; semantic and architectural guidance requires human review
3. **AI Agent Gap** - Claude Code, Kiro, Codex, and other AI coding assistants lack structured API design knowledge
4. **Inconsistent Tooling** - Teams maintain separate configurations for CI linters, IDE hints, and documentation
5. **No Historical Tracking** - Organizations cannot measure API quality trends over time

### Market Opportunity

| Segment | Need | Current Solution | Gap |
|---------|------|------------------|-----|
| Platform Teams | Enforce API standards | Spectral + manual review | No LLM evaluation, no agent integration |
| API Designers | Learn and apply guidelines | Read documentation | No interactive feedback |
| AI Agents | Design compliant APIs | None | No structured knowledge |
| Governance | Track compliance trends | Manual audits | No automated reporting |

### Competitive Landscape

| Solution | Strengths | Limitations |
|----------|-----------|-------------|
| Spectral | Industry standard, extensible | Slow, no LLM, CLI-only |
| vacuum | Fast, Go library | No style guide generation |
| Optic | API diff/changelog | No style enforcement |
| Redocly | Documentation focus | Limited linting |
| **systemspec-apistyle** | Unified spec, LLM+lint, multi-platform | New entrant |

## Target Users

### Primary

1. **Platform Engineers** - Define and enforce API standards across organization
2. **API Designers** - Learn guidelines, get feedback during design
3. **AI Agent Developers** - Embed API design knowledge into agents

### Secondary

1. **Technical Writers** - Generate consistent API documentation
2. **Security Teams** - Enforce security-related API patterns
3. **Compliance Officers** - Audit API conformance for regulatory requirements

## Market Requirements

### MR-1: Machine-Readable Style Specification

**Need:** Organizations need a canonical, version-controlled format for API style rules that machines can consume.

**Acceptance Criteria:**

- JSON format with JSON Schema validation
- Covers all rule types: syntax, structural, semantic, architectural
- Supports severity levels, rationale, examples
- Extensible via profiles and custom rules

### MR-2: Deterministic Linting

**Need:** Fast, consistent, CI-friendly linting that catches objective violations.

**Acceptance Criteria:**

- Sub-second linting for typical specs (<1MB)
- Spectral-compatible ruleset format
- CI exit codes (0=pass, 1=violations, 2=error)
- Multiple output formats (JSON, SARIF, text)

### MR-3: LLM-Based Evaluation

**Need:** Evaluate semantic and architectural guidance that cannot be linted deterministically.

**Acceptance Criteria:**

- Rubric-based evaluation with categorical scoring
- Findings with severity and remediation
- Reproducible results (temperature=0, prompt versioning)
- Cost-effective (target <$0.05 per evaluation)

### MR-4: Human-Readable Documentation

**Need:** Generate clear, well-organized style guides that humans can read and learn from.

**Acceptance Criteria:**

- Markdown output with proper structure
- PDF export for offline/print use
- Searchable, categorized rules
- Good/bad examples for each rule

### MR-5: AI Agent Integration

**Need:** Embed API style knowledge into AI coding assistants.

**Acceptance Criteria:**

- Claude Code agent definitions
- Kiro CLI agent definitions
- MCP server for IDE integration
- Structured prompts optimized for LLM consumption

### MR-6: Web-Based Analysis

**Need:** Allow ad-hoc evaluation without CLI installation.

**Acceptance Criteria:**

- URL-based spec input
- Interactive results display
- Export to JSON/PDF
- No server-side storage of specs (privacy)

### MR-7: Observability and Trends

**Need:** Track API quality over time for governance and improvement.

**Acceptance Criteria:**

- Historical evaluation storage
- Trend visualization
- DevOps integration (Datadog, etc.)
- Compliance reporting

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Lint Performance | <500ms for 1MB spec | Benchmark suite |
| Rule Coverage | 80% of Microsoft guidelines | Rule mapping audit |
| LLM Accuracy | 90% agreement with human review | Calibration study |
| Adoption | 100+ GitHub stars in 6 months | GitHub metrics |
| Agent Effectiveness | 50% reduction in API review cycles | User survey |

## Constraints

1. **Open Source** - MIT license, public repository
2. **Privacy** - No server-side spec storage in web UI
3. **Cost** - LLM evaluation must be cost-effective (<$0.10/spec)
4. **Performance** - Deterministic lint <1s, LLM evaluation <30s
5. **Compatibility** - Must consume existing Spectral rulesets

## Timeline

| Phase | Duration | Focus |
|-------|----------|-------|
| Phase 1 | 4 weeks | Core types, vacuum integration, CLI |
| Phase 2 | 4 weeks | LLM evaluation, profiles |
| Phase 3 | 4 weeks | Web UI, MCP server |
| Phase 4 | 4 weeks | Agents, observability |
| Phase 5 | Ongoing | Profiles, community, polish |

## References

- [Microsoft REST API Guidelines](https://github.com/microsoft/api-guidelines)
- [Azure API Style Guide (Spectral)](https://github.com/Azure/azure-api-style-guide)
- [Google API Design Guide](https://cloud.google.com/apis/design)
- [Zalando RESTful API Guidelines](https://opensource.zalando.com/restful-api-guidelines/)
- [vacuum](https://github.com/daveshanley/vacuum)
- [Spectral](https://github.com/stoplightio/spectral)
