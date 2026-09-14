# Roadmap

## systemspec-apistyle

**Version:** 0.1.0-draft
**Date:** 2026-06-03
**Status:** Draft

## Overview

This roadmap outlines the phased development of systemspec-apistyle, from core types through full ecosystem integration.

## Phase Summary

| Phase | Focus | Key Deliverables |
|-------|-------|------------------|
| **Phase 1** | Foundation | Core types, JSON Schema, vacuum integration, basic CLI |
| **Phase 2** | Intelligence | LLM evaluation, profiles, documentation generation |
| **Phase 3** | Integration | Web UI, MCP server, agent generation |
| **Phase 4** | Enterprise | Observability, conformance levels, breaking change detection |
| **Phase 5** | Ecosystem | Community profiles, advanced features, polish |

---

## Phase 1: Foundation

**Goal:** Establish core types and deterministic linting capability.

### Milestones

#### P1.1: Project Setup

- [ ] Initialize Go module
- [ ] Set up directory structure per TRD
- [ ] Configure CI/CD (GitHub Actions)
- [ ] Set up linting (golangci-lint)
- [ ] Create initial README

#### P1.2: Core Types

- [ ] Define `APIStyleSpec` root type
- [ ] Define `Rule` and related types
- [ ] Define `LintReport` and `Violation` types
- [ ] Define `Profile` and `Extends` types
- [ ] Generate JSON Schema from Go types
- [ ] Validate schema with `schemago lint`
- [ ] Embed schema via `//go:embed`

#### P1.3: vacuum Integration

- [ ] Add vacuum as dependency
- [ ] Implement `Linter` interface
- [ ] Convert `APIStyleSpec` rules to vacuum ruleset
- [ ] Convert vacuum results to `LintReport`
- [ ] Support custom Spectral-compatible rules
- [ ] Unit tests with golden files

#### P1.4: Basic CLI

- [ ] Set up cobra CLI structure
- [ ] Implement `api-style lint` command
- [ ] Support `--format json|text|sarif`
- [ ] Support `--output` file path
- [ ] Exit codes (0=pass, 1=violations, 2=error)
- [ ] Basic `--help` documentation

#### P1.5: Default Profile

- [ ] Create `default.api-style.json` with ~30 rules
- [ ] Cover categories: uri-design, http-methods, naming, errors
- [ ] Embed as built-in profile
- [ ] Document each rule with rationale and examples

### Deliverables

- `pkg/types/` - Core Go types
- `schema/` - Generated JSON Schema
- `pkg/lint/` - vacuum integration
- `cmd/api-style/` - CLI with `lint` command
- `profiles/default.api-style.json` - Built-in profile
- Unit tests with >70% coverage

---

## Phase 2: Intelligence

**Goal:** Add LLM evaluation and documentation generation.

### Milestones

#### P2.1: structured-evaluation Integration

- [ ] Add structured-evaluation as dependency
- [ ] Implement `Evaluator` interface
- [ ] Build `RubricSet` from `APIStyleSpec`
- [ ] Generate evaluation prompts from rules
- [ ] Map LLM responses to `rubric.Rubric`

#### P2.2: Anthropic Integration

- [ ] Implement Claude client wrapper
- [ ] Category-based evaluation
- [ ] Configurable model (haiku/sonnet/opus)
- [ ] Temperature and token control
- [ ] Error handling and retries

#### P2.3: Combined Analysis

- [ ] Implement `Analyzer` orchestrator
- [ ] Combine lint + evaluate results
- [ ] Generate `SummaryReport` (GO/NO-GO)
- [ ] Implement `api-style evaluate` command
- [ ] Implement `api-style analyze` command

#### P2.4: Profiles

- [ ] Create `azure.api-style.json` (~80 rules)
- [ ] Create `google.api-style.json` (~60 rules)
- [ ] Create `zalando.api-style.json` (~70 rules)
- [ ] Implement profile loading and resolution
- [ ] Support `extends` inheritance
- [ ] Support `overrides` customization
- [ ] Implement `api-style profile list|show|validate`

#### P2.5: Documentation Generation

- [ ] Implement Markdown generator
- [ ] Generate table of contents
- [ ] Generate rule categories
- [ ] Include examples (good/bad)
- [ ] Include rationale and references
- [ ] Implement `api-style generate guide`

#### P2.6: Spectral Generation

- [ ] Implement Spectral YAML generator
- [ ] Map enforcement rules to Spectral format
- [ ] Support custom functions
- [ ] Implement `api-style generate spectral`

### Deliverables

- `pkg/judge/` - LLM evaluation
- `pkg/analyze/` - Combined orchestration
- `pkg/generate/` - Output generators
- `pkg/profile/` - Profile management
- `profiles/*.api-style.json` - Industry profiles
- CLI commands: `evaluate`, `analyze`, `generate guide|spectral`, `profile`

---

## Phase 3: Integration

**Goal:** Enable IDE and web-based usage.

### Milestones

#### P3.1: MCP Server

- [ ] Add mcp-go as dependency
- [ ] Implement MCP server structure
- [ ] Register `api_style_lint` tool
- [ ] Register `api_style_evaluate` tool
- [ ] Register `api_style_analyze` tool
- [ ] Register `api_style_explain_rule` tool
- [ ] Register `api_style_list_rules` tool
- [ ] Implement `api-style serve mcp`

#### P3.2: Agent Generation

- [ ] Create `api-style-reviewer.md` spec (multi-agent-spec)
- [ ] Create `api-style-guide-generator.md` spec
- [ ] Implement agent instruction generator
- [ ] Integrate assistantkit for Claude Code output
- [ ] Integrate assistantkit for Kiro output
- [ ] Implement `api-style generate agent`

#### P3.3: Web UI - Core

- [ ] Set up web/packages/api-style-analyzer
- [ ] Implement `api-style-analyzer` main component
- [ ] Implement `asa-url-input` component
- [ ] Implement tab navigation (Evaluate/Guide/Rules)
- [ ] Set up Vite build configuration
- [ ] Integrate PRISM design system

#### P3.4: Web UI - Evaluate Tab

- [ ] URL inputs for OpenAPI, Style Spec, Spectral
- [ ] "Evaluate" button with loading state
- [ ] Implement `asa-evaluation-report` component
- [ ] Display violations with severity icons
- [ ] Display LLM findings (if applicable)
- [ ] Export to JSON button
- [ ] Export to PDF button

#### P3.5: Web UI - Style Guide Tab

- [ ] URL input for Style Spec
- [ ] Load and parse spec
- [ ] Render as Markdown (reuse markdown-editor patterns)
- [ ] Download as Markdown button
- [ ] Download as PDF button (reuse export-pdf utility)

#### P3.6: Web UI - Rules Browser Tab

- [ ] Implement `asa-rule-browser` component
- [ ] Search/filter by keyword
- [ ] Filter by category
- [ ] Filter by severity
- [ ] Rule detail view with examples

#### P3.7: Web Server (Optional)

- [ ] Implement Go server for Web UI
- [ ] Serve static assets
- [ ] API endpoints for lint/evaluate (optional)
- [ ] Implement `api-style serve web`

### Deliverables

- `mcp/` - MCP server
- `agents/specs/` - Agent definitions
- `agents/plugins/` - Generated Claude Code/Kiro files
- `web/packages/api-style-analyzer/` - Web UI
- `web/server/` - Optional Go server
- CLI commands: `serve mcp`, `serve web`, `generate agent`

---

## Phase 4: Enterprise

**Goal:** Add governance and observability features.

### Milestones

#### P4.1: Conformance Levels

- [ ] Define `ConformanceLevel` type
- [ ] Implement level calculation from violations
- [ ] Add `--level bronze|silver|gold` option
- [ ] Fail if level not achieved
- [ ] Display achieved level in reports

#### P4.2: Exception Registry

- [ ] Define `Exception` type
- [ ] Implement exception loading
- [ ] Filter violations by exceptions
- [ ] Track exception expiration
- [ ] Warn on expired exceptions

#### P4.3: Lexicon/Terminology

- [ ] Define `Lexicon` type
- [ ] Implement terminology checker
- [ ] Check approved terms
- [ ] Flag forbidden terms with replacements
- [ ] Integrate into lint pipeline

#### P4.4: Breaking Change Detection

- [ ] Add openapi-diff or implement custom differ
- [ ] Define `CompatibilityReport` type
- [ ] Detect breaking changes (removed endpoints, required params, etc.)
- [ ] Implement `api-style diff` command
- [ ] Support `--fail-on-breaking` flag

#### P4.5: Evaluation Storage

- [ ] Define storage interface
- [ ] Implement SQLite backend
- [ ] Implement PostgreSQL backend
- [ ] Store evaluation results
- [ ] Implement `api-style history` command

#### P4.6: Observability Export

- [ ] Define metrics interface
- [ ] Implement Datadog exporter
- [ ] Implement generic StatsD exporter
- [ ] Export lint counts, decision, conformance level
- [ ] Add `--export datadog|statsd` option

### Deliverables

- `pkg/types/conformance.go` - Conformance levels
- `pkg/types/exception.go` - Exception registry
- `pkg/types/lexicon.go` - Terminology
- `pkg/diff/` - Breaking change detection
- `pkg/store/` - Evaluation storage
- CLI commands: `diff`, `history`

---

## Phase 5: Ecosystem

**Goal:** Community growth and advanced features.

### Milestones

#### P5.1: Community Profiles

- [ ] Accept community profile contributions
- [ ] Profile validation in CI
- [ ] Profile discovery/registry
- [ ] Documentation for creating profiles

#### P5.2: IDE Extensions

- [ ] VS Code extension (consumes MCP)
- [ ] JetBrains plugin (consumes MCP)
- [ ] Real-time linting feedback

#### P5.3: CI/CD Templates

- [ ] GitHub Actions workflow template
- [ ] GitLab CI template
- [ ] Azure DevOps template
- [ ] Documentation for CI integration

#### P5.4: Advanced LLM Features

- [ ] Multi-judge evaluation (consensus)
- [ ] Fine-tuned judge calibration
- [ ] Cost optimization (tiered evaluation)

#### P5.5: API Format Extensions

- [ ] AsyncAPI support (experimental)
- [ ] GraphQL support (experimental)
- [ ] gRPC/protobuf support (experimental)

#### P5.6: Documentation & Polish

- [ ] Comprehensive user guide
- [ ] API reference documentation
- [ ] Tutorial videos
- [ ] Example repositories
- [ ] Blog posts / announcements

### Deliverables

- Community contribution guidelines
- IDE extensions
- CI/CD templates
- Extended format support
- Production-ready documentation

---

## Feature Matrix by Phase

| Feature | P1 | P2 | P3 | P4 | P5 |
|---------|:--:|:--:|:--:|:--:|:--:|
| Core Types | ✅ | | | | |
| JSON Schema | ✅ | | | | |
| vacuum Integration | ✅ | | | | |
| CLI (lint) | ✅ | | | | |
| Default Profile | ✅ | | | | |
| LLM Evaluation | | ✅ | | | |
| Industry Profiles | | ✅ | | | |
| Documentation Generation | | ✅ | | | |
| Spectral Generation | | ✅ | | | |
| CLI (evaluate, analyze, generate) | | ✅ | | | |
| MCP Server | | | ✅ | | |
| Agent Generation | | | ✅ | | |
| Web UI | | | ✅ | | |
| Conformance Levels | | | | ✅ | |
| Exception Registry | | | | ✅ | |
| Lexicon/Terminology | | | | ✅ | |
| Breaking Change Detection | | | | ✅ | |
| Evaluation Storage | | | | ✅ | |
| Observability Export | | | | ✅ | |
| Community Profiles | | | | | ✅ |
| IDE Extensions | | | | | ✅ |
| CI/CD Templates | | | | | ✅ |
| Extended Format Support | | | | | ✅ |

---

## Release Schedule

| Version | Phase | Focus |
|---------|-------|-------|
| v0.1.0 | P1 | Foundation - lint capability |
| v0.2.0 | P2.1-P2.3 | LLM evaluation |
| v0.3.0 | P2.4-P2.6 | Profiles and generation |
| v0.4.0 | P3.1-P3.2 | MCP and agents |
| v0.5.0 | P3.3-P3.7 | Web UI |
| v0.6.0 | P4.1-P4.3 | Conformance and lexicon |
| v0.7.0 | P4.4-P4.6 | Diff, storage, observability |
| v1.0.0 | P5 | Production release |

---

## Success Metrics

| Metric | Target | Phase |
|--------|--------|-------|
| Lint <500ms (1MB spec) | ✓ | P1 |
| Default profile covers common rules | 30+ rules | P1 |
| Industry profiles available | 3 profiles | P2 |
| LLM evaluation cost | <$0.10/spec | P2 |
| Web UI functional | All tabs working | P3 |
| MCP server functional | All tools working | P3 |
| GitHub stars | 100+ | P5 |
| Community profiles | 5+ | P5 |

---

## Dependencies

### External

| Dependency | Version | Used In |
|------------|---------|---------|
| vacuum | latest | P1 |
| structured-evaluation | latest | P2 |
| multi-agent-spec | latest | P3 |
| assistantkit | latest | P3 |
| mcp-go | latest | P3 |
| huma/v2 | latest | P3 (optional) |
| ent | latest | P4 (optional) |

### Internal (from grokify/plexusone ecosystem)

| Project | Purpose |
|---------|---------|
| structured-evaluation | LLM evaluation framework |
| multi-agent-spec | Agent definitions |
| assistantkit | Agent file generation |
| web-tools | Web UI components |
| schemago | JSON Schema validation |

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| vacuum API changes | Pin version, abstract behind interface |
| LLM cost overruns | Tiered evaluation, caching, cost alerts |
| Scope creep | Strict phase boundaries, defer to backlog |
| Community adoption | Early outreach, good documentation |

---

## Next Steps

1. **Approve this roadmap** - Review with stakeholders
2. **Begin Phase 1** - Start with project setup and core types
3. **Iterate** - Adjust roadmap based on learnings
