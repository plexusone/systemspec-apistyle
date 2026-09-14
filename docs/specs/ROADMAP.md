# systemspec-apistyle Roadmap

This document outlines planned features and enhancements for systemspec-apistyle, organized by priority.

## Completed

- [x] Core linting engine with vacuum integration
- [x] LLM evaluation with Claude
- [x] MCP server with resources and prompts
- [x] Claude Code hooks integration
- [x] Custom profile loading (YAML/JSON)
- [x] Example specs (PetStore, E-Commerce)
- [x] Web UI with Lit + Vite
- [x] REST API backend for Web UI
- [x] MkDocs documentation structure
- [x] SARIF output for IDE/CI integration
- [x] Test coverage improvements (82%+ on core packages)

---

## Priority 1: Essential Developer Experience

High-impact features that make the tool immediately useful for daily development.

### 1.1 Config File Support

Project-level configuration via `.api-style.yaml`:

```yaml
profile: azure
level: silver
exceptions:
  - rule: INFO-001
    paths: ["$.info.x-internal"]
severity-overrides:
  NAMING-001: warn
include:
  - "api/**/*.yaml"
exclude:
  - "**/examples/**"
```

**Impact:** Enables team-wide settings, reduces CLI flag repetition.

### 1.2 Watch Mode

Re-lint automatically on file changes:

```bash
api-style lint openapi.yaml --watch
```

**Impact:** Immediate feedback during API development.

### 1.3 Multi-File Support

Lint multiple specs or directories:

```bash
api-style lint api/*.yaml
api-style lint ./specs --recursive
```

**Impact:** Supports monorepos and multi-service architectures.

### 1.4 Pre-Commit Hook Generator

Generate git hooks for automatic linting:

```bash
api-style hooks init
# Creates .git/hooks/pre-commit that runs api-style lint
```

**Impact:** Prevents violations from being committed.

---

## Priority 2: CI/CD Integration

Features that enable seamless integration with CI/CD pipelines.

### 2.1 GitHub Action

Reusable action for GitHub workflows:

```yaml
- uses: plexusone/api-style-action@v1
  with:
    spec: openapi.yaml
    profile: azure
    fail-on: error
```

**Impact:** One-line integration for GitHub users.

### 2.2 JUnit XML Output

Test result format for CI systems:

```bash
api-style lint openapi.yaml --format junit --output results.xml
```

**Impact:** Native integration with Jenkins, GitLab CI, Azure DevOps.

### 2.3 Baseline Mode

Ignore existing violations, only fail on new ones:

```bash
# Generate baseline
api-style lint openapi.yaml --format json > .api-style-baseline.json

# Lint with baseline (only new violations fail)
api-style lint openapi.yaml --baseline .api-style-baseline.json
```

**Impact:** Enables gradual adoption without fixing all legacy issues.

### 2.4 Diff Mode

Only report violations on changed lines:

```bash
api-style lint openapi.yaml --diff HEAD~1
api-style lint openapi.yaml --diff main
```

**Impact:** Focused feedback on PR changes, reduces noise.

### 2.5 GitHub PR Comments

Post violations as PR review comments:

```bash
api-style lint openapi.yaml --github-pr 123
```

**Impact:** Inline feedback directly on pull requests.

---

## Priority 3: Profile Management

Features for creating, managing, and sharing style profiles.

### 3.1 Profile Validation

Validate custom profiles before use:

```bash
api-style profile validate my-profile.yaml
```

Checks:

- Required fields (name, version)
- Valid rule syntax
- Referenced base profiles exist
- Spectral function compatibility

**Impact:** Catches profile errors early.

### 3.2 Profile Diff

Compare two profiles to see differences:

```bash
api-style profile diff azure google
```

Output:

- Rules unique to each profile
- Severity differences for shared rules
- Configuration differences

**Impact:** Helps teams choose or merge profiles.

### 3.3 Profile Generator

Generate a profile from an existing "golden" API spec:

```bash
api-style profile generate --from golden-api.yaml --output company-profile.yaml
```

Analyzes the spec to infer:

- Naming conventions
- URI patterns
- Required fields
- Common structures

**Impact:** Bootstrap profiles from existing best-practice APIs.

---

## Priority 4: Web UI Enhancements

Features that improve the web-based linting experience.

### 4.1 Monaco Editor Integration

Replace textarea with VS Code's Monaco editor:

- Syntax highlighting for YAML/JSON
- Error squiggles at violation locations
- Auto-completion for OpenAPI keywords
- Bracket matching and folding

**Impact:** Professional editing experience.

### 4.2 Real-Time Linting

Lint as you type with debouncing:

- 300ms debounce after typing stops
- Loading indicator during lint
- Incremental updates

**Impact:** Immediate feedback without manual "Lint" button.

### 4.3 Rule Explanations

Click violation to see full details:

- Rule rationale
- Good/bad examples
- Links to documentation
- Suggested fixes

**Impact:** Educational, helps developers understand why rules exist.

### 4.4 Share via URL

Encode spec in URL for sharing:

```
https://api-style.example.com/?spec=base64encoded...&profile=azure
```

- Compress with gzip before base64
- Short URLs via hash storage
- Copy-to-clipboard button

**Impact:** Easy sharing for code reviews and discussions.

### 4.5 Side-by-Side View

Spec editor and results panel side by side:

- Click violation to highlight in editor
- Synchronized scrolling option
- Collapsible panels

**Impact:** Better workflow for fixing violations.

### 4.6 Dark Mode Toggle

Manual theme switching:

- Light/dark/system preference
- Persist preference in localStorage
- Smooth transition animation

**Impact:** User comfort preference.

---

## Priority 5: LLM Evaluation Features

Advanced AI-powered analysis capabilities.

### 5.1 Comparison Mode

Compare two API versions for breaking changes:

```bash
api-style compare v1/openapi.yaml v2/openapi.yaml
```

Detects:

- Removed endpoints
- Changed request/response schemas
- Modified authentication requirements
- Semantic breaking changes (via LLM)

**Impact:** Prevents accidental breaking changes.

### 5.2 Custom Evaluation Prompts

User-defined evaluation criteria:

```yaml
# .api-style.yaml
custom-evaluations:
  - id: COMPANY-001
    prompt: "Check if all endpoints follow our naming convention: /api/v{version}/{domain}/{resource}"
    severity: error
```

**Impact:** Extend LLM evaluation with company-specific rules.

### 5.3 Batch Evaluation

Evaluate multiple specs efficiently:

```bash
api-style evaluate ./specs/*.yaml --parallel 4
```

- Progress bar for batch jobs
- Summary report across all specs
- Cost estimation before running

**Impact:** Supports large-scale API governance.

### 5.4 Cost Estimation

Estimate API costs before LLM evaluation:

```bash
api-style evaluate openapi.yaml --estimate
# Estimated cost: ~$0.12 (15k input tokens, 2k output tokens)
```

**Impact:** Budget awareness for LLM features.

---

## Priority 6: Editor Integration

Native IDE support for real-time feedback.

### 6.1 VS Code Extension

Real-time linting in VS Code:

- Uses SARIF output internally
- Diagnostics panel integration
- Quick-fix suggestions
- Status bar indicator
- Settings UI for configuration

**Impact:** Most popular editor, huge reach.

### 6.2 Language Server Protocol (LSP)

Generic LSP server for any editor:

```bash
api-style lsp
```

Features:

- Real-time diagnostics
- Hover information for rules
- Code actions for fixes
- Workspace configuration

**Impact:** Supports Vim, Emacs, Sublime, and other LSP clients.

### 6.3 JetBrains Plugin

IntelliJ/WebStorm integration:

- Inspections for OpenAPI files
- Intention actions for fixes
- Tool window for results

**Impact:** Popular among enterprise Java/Kotlin developers.

---

## Future Considerations

Ideas for longer-term exploration:

- **API Registry Integration** - Sync with Backstage, Apigee, Kong
- **GraphQL Support** - Extend beyond REST/OpenAPI
- **AsyncAPI Support** - Event-driven API specifications
- **Spectral Compatibility** - Import existing Spectral rulesets
- **Team Analytics** - Dashboard for org-wide API quality trends
- **AI Fix Suggestions** - LLM-powered auto-fix recommendations
- **Versioned Profiles** - Profile versioning and migration guides

---

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for how to propose new features or contribute implementations.

Feature requests and discussions are welcome in [GitHub Issues](https://github.com/plexusone/systemspec-apistyle/issues).
