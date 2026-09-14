# AI Agent Support Roadmap

## systemspec-apistyle v0.5.0+

This roadmap outlines the features planned for AI agent support in systemspec-apistyle.

---

## Completed (v0.4.0)

Foundation features already in place:

- [x] MCP server with lint, evaluate, analyze tools
- [x] list_rules and explain_rule tools
- [x] Claude Code hooks integration
- [x] Structured JSON output for all reports
- [x] SARIF output for IDE integration
- [x] Profile loading and validation

---

## v0.5.0: Enhanced Remediation

**Theme:** Make violations actionable for AI agents.

### 0.5.0-alpha: Violation Enhancement

- [ ] Add `ExampleFix` field to Violation type
- [ ] Add `RuleURL` field with documentation links
- [ ] Add `Confidence` field (0.0-1.0)
- [ ] Add `RelatedRules` field for dependencies
- [ ] Add `FixPriority` field for ordering
- [ ] Enrich violations during lint pipeline
- [ ] Update CLI text/JSON output with new fields

**Impact:** AI agents can suggest specific fixes instead of just reporting problems.

### 0.5.0-beta: Fix Suggestion Engine

- [ ] Create `pkg/fix` package
- [ ] Implement `RuleFixer` for rule-based suggestions
- [ ] Generate JSON Patch operations (RFC 6902)
- [ ] Add `suggest-fixes` CLI command
- [ ] Add `--suggest-fixes` flag to lint command

**Impact:** Dedicated tooling for fix generation, machine-applicable patches.

### 0.5.0: MCP Tool Integration

- [ ] Add `suggest_fixes` MCP tool
- [ ] Add `design_check` MCP tool
- [ ] Add `conformance_path` MCP tool
- [ ] Add exemplar MCP resources
- [ ] Integration tests for all tools

**Impact:** Full AI agent integration via MCP protocol.

---

## v0.6.0: Generation Support

**Theme:** Help AI agents generate conformant specs from scratch.

### Generation Guidance

- [ ] Add `GenerationGuidance` type to Rule
- [ ] Add `generate.prompt` to all error-severity rules
- [ ] Add `generate.template` for URI rules
- [ ] Add `generate.priority` for ordering
- [ ] Generate generation-focused rubrics

**Impact:** Rules include positive directives for spec creation.

### Exemplar Specs

- [ ] Create exemplar directory structure
- [ ] Add minimal exemplar to each profile
- [ ] Add CRUD API exemplar to default profile
- [ ] Exemplar loading via MCP resources
- [ ] CLI command to list/show exemplars

**Impact:** AI agents have reference implementations to learn from.

### Pattern Library

- [ ] Define Pattern type with templates
- [ ] Add CRUD collection pattern
- [ ] Add pagination pattern
- [ ] Add error response pattern
- [ ] Pattern expansion with variables
- [ ] `get_pattern` MCP tool

**Impact:** Reusable building blocks for spec generation.

---

## v0.7.0: LLM-Powered Fixes

**Theme:** Use LLM for complex fix suggestions.

### LLM Fix Generation

- [ ] Create `LLMFixer` implementation
- [ ] Generate fixes for semantic rules
- [ ] Confidence scoring for LLM suggestions
- [ ] Cost estimation before LLM calls
- [ ] Caching for repeated patterns

**Impact:** Fix suggestions for rules that can't be deterministically fixed.

### Interactive Fix Session

- [ ] Multi-turn fix conversation
- [ ] Context preservation across fixes
- [ ] Progress tracking toward conformance
- [ ] Batch fix application

**Impact:** Collaborative fixing workflow with AI agent.

---

## v0.8.0: Advanced Agent Features

**Theme:** Production-grade agent integration.

### Auto-Fix Mode

- [ ] `api-style fix openapi.yaml --auto`
- [ ] Apply high-confidence fixes automatically
- [ ] Interactive confirmation for low-confidence
- [ ] Dry-run mode with diff preview
- [ ] Rollback capability

**Impact:** Agents can fix simple issues without human intervention.

### Conformance Tracking

- [ ] Track conformance progress over time
- [ ] Show improvements per commit
- [ ] Visualize path to gold level
- [ ] CI integration for progress gates

**Impact:** Measurable API quality improvement.

### Agent Learning

- [ ] Store fix patterns that worked
- [ ] Learn from user corrections
- [ ] Suggest based on similar specs
- [ ] Profile-specific fix strategies

**Impact:** Fixes improve over time based on usage.

---

## Future Considerations

Ideas for longer-term exploration:

### Multi-Spec Analysis

- Analyze relationships between specs
- Detect inconsistencies across APIs
- Suggest standardization opportunities

### Breaking Change Prevention

- Integrate with diff command
- Warn about breaking fixes
- Suggest non-breaking alternatives
- Version migration paths

### IDE Deep Integration

- VS Code extension with fix suggestions
- Real-time conformance indicator
- Quick-fix code actions
- Inline exemplar snippets

### Team Analytics

- Fix patterns by team
- Common violation trends
- Training recommendations
- Quality scoring dashboards

---

## Version Summary

| Version | Theme | Key Features |
|---------|-------|--------------|
| v0.5.0 | Enhanced Remediation | Violation enrichment, suggest_fixes tool, MCP integration |
| v0.6.0 | Generation Support | Generation guidance, exemplars, pattern library |
| v0.7.0 | LLM-Powered Fixes | Semantic fix suggestions, interactive sessions |
| v0.8.0 | Advanced Agent | Auto-fix, conformance tracking, agent learning |

---

## Contributing

Feature requests and discussions welcome in [GitHub Issues](https://github.com/plexusone/systemspec-apistyle/issues).

To contribute implementations, see [CONTRIBUTING.md](../../../CONTRIBUTING.md).
