# Project Constitution

Technology standards and architectural decisions for systemspec-apistyle.

## Core Principles

1. **Go-First Development** - Go is the primary implementation language
2. **Schema-Driven Design** - JSON Schema defines contracts, generated from Go types
3. **Static Typing Friendly** - All schemas must work well with statically typed languages
4. **Minimal Dependencies** - Prefer standard library; add dependencies only when justified
5. **Library-First** - Build as importable packages; CLI is a thin wrapper

## Technology Stack

### Backend

| Layer | Technology | Notes |
|-------|------------|-------|
| Language | Go 1.22+ | Primary implementation language |
| REST API | Huma + Chi | If REST API is needed |
| Database ORM | Ent | If database is needed |
| RDBMS | SQLite (dev/single), PostgreSQL (prod) | Prioritize SQLite for simplicity |
| Linting | vacuum (Go library) | Spectral-compatible, faster |
| LLM Evaluation | structured-evaluation | Rubric-based LLM-as-Judge |
| Agent Specs | multi-agent-spec | Agent/team definitions |
| Agent Generation | assistantkit | Claude Code/Kiro file generation |

### Frontend

| Layer | Technology | Notes |
|-------|------------|-------|
| Framework | Lit Web Components | TypeScript-based |
| Design System | PRISM (from web-tools) | CSS custom properties |
| Schema Validation | Zod | Generated from OpenAPI spec |
| Build | Vite | ES modules, TypeScript |
| Package Manager | pnpm | Workspace support |

### Schema Strategy

| Aspect | Approach |
|--------|----------|
| Source of Truth | Go structs with `json` tags |
| Schema Generation | `invopop/jsonschema` from Go types |
| Schema Validation | `schemago lint` for static-typing friendliness |
| Schema Embedding | `//go:embed` for runtime access |
| Frontend Types | Zod schemas generated from OpenAPI spec |

## JSON Schema Requirements

All JSON schemas MUST pass `schemago lint` validation:

1. **No ambiguous unions** - `anyOf`/`oneOf` must have discriminator fields
2. **No `interface{}` degradation** - All types must map cleanly to Go types
3. **Consistent discriminators** - Same field name across variants
4. **Required fields explicit** - No implicit optionality
5. **Enums as strings** - Not integers

### Schema Generation Workflow

```bash
# 1. Define Go types (source of truth)
# pkg/types/rule.go

# 2. Generate JSON Schema
go generate ./...

# 3. Validate schema is static-typing friendly
schemago lint schema/api-style-spec.schema.json

# 4. Embed for runtime use
# schema/embed.go with //go:embed
```

## API Design (If Applicable)

If a REST API is needed:

1. **OpenAPI 3.1** specification generated from Huma
2. **Consistent error format** following RFC 7807 (Problem Details)
3. **Versioning** via URL path (`/v1/...`)
4. **Authentication** via Bearer token (API key or JWT)

## Database Design (If Applicable)

If persistence is needed:

1. **Ent** for schema definition and migrations
2. **SQLite** for development and single-instance deployments
3. **PostgreSQL** for production multi-instance deployments
4. **Migrations** tracked in `ent/migrate/migrations/`

## Code Organization

```
systemspec-apistyle/
├── pkg/                    # Core library packages
│   ├── types/             # Go types (schema source of truth)
│   ├── lint/              # vacuum integration
│   ├── judge/             # LLM evaluation
│   ├── generate/          # Output generators
│   └── store/             # Persistence (if needed)
├── schema/                 # Generated JSON schemas + embed
├── cmd/
│   └── api-style/         # CLI binary
├── mcp/                    # MCP server
├── web/                    # Frontend (Lit components)
│   └── packages/
├── agents/                 # Agent definitions
│   ├── specs/             # multi-agent-spec format
│   └── plugins/           # Generated (assistantkit)
├── profiles/               # Pre-built style profiles
├── docs/
│   └── specs/             # Specification documents
└── examples/               # Example usage
```

## Testing Standards

1. **Unit tests** for all packages (`*_test.go`)
2. **Table-driven tests** preferred
3. **Golden files** for schema/output validation
4. **Integration tests** in `integration/` directory
5. **Coverage** minimum 70% for core packages

## Error Handling

Priority order (per CLAUDE.md):

1. **Panic** - Programming errors, invariant violations
2. **Return** - Propagate errors to caller
3. **Log** - Via `*slog.Logger` from context
4. **Report** - Inform user with remediation guidance

## Commit Standards

Follow Conventional Commits (per CLAUDE.md):

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation
- `refactor:` - Code restructuring
- `test:` - Test additions/changes
- `chore:` - Maintenance tasks

## Dependencies

### Go Dependencies (Verify Latest Versions)

| Module | Purpose | Verify URL |
|--------|---------|------------|
| `github.com/daveshanley/vacuum` | OpenAPI linting | https://github.com/daveshanley/vacuum/releases |
| `github.com/invopop/jsonschema` | Schema generation | https://github.com/invopop/jsonschema/releases |
| `github.com/spf13/cobra` | CLI framework | https://github.com/spf13/cobra/releases |
| `github.com/danielgtaylor/huma/v2` | REST API (if needed) | https://github.com/danielgtaylor/huma/releases |
| `entgo.io/ent` | Database ORM (if needed) | https://github.com/ent/ent/releases |

### Frontend Dependencies

| Package | Purpose |
|---------|---------|
| `lit` | Web components |
| `vite` | Build tooling |
| `zod` | Schema validation |
| `typescript` | Type safety |

## Review Checklist

Before merging:

- [ ] Go types defined with proper `json` tags
- [ ] JSON Schema generated and passes `schemago lint`
- [ ] Tests pass with adequate coverage
- [ ] `golangci-lint run` passes
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (if user-facing)
