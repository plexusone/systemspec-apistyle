# Using Profiles

Profiles are collections of API style rules. systemspec-apistyle includes profiles based on industry-standard guidelines.

## Built-in Profiles

| Profile | Based On | Rules | Categories | Focus |
|---------|----------|-------|------------|-------|
| `default` | Microsoft, Zalando, Google, PayPal | 79 | 26 | Full coverage, recommended |
| `comprehensive` | All major guides | 88 | 26 | Maximum coverage |
| `zalando` | Zalando RESTful Guidelines | 147 | 13 | E-commerce, events |
| `microsoft-rest` | Microsoft REST Guidelines | 123 | 15 | Enterprise REST APIs |
| `microsoft-graph` | Microsoft Graph Guidelines | 82 | 12 | OData/Graph APIs |
| `minimal` | Core essentials | 29 | 7 | Basic API hygiene |
| `azure` | Azure REST Guidelines | 23 | 9 | Azure cloud services |
| `google` | Google API Design Guide | 20 | 7 | Resource-oriented design |

## Selecting a Profile

Use the `--profile` flag with any command:

```bash
# Lint with Azure profile
api-style lint openapi.yaml --profile azure

# Analyze with Google profile
api-style analyze openapi.yaml --profile google

# Generate documentation for Zalando profile
api-style generate guide --profile zalando
```

## Profile Details

### Default Profile

The default profile provides complete REST API coverage synthesized from Microsoft, Zalando, Google, and PayPal guidelines:

- **URI Design** - kebab-case, plural resources, version prefix
- **HTTP Methods** - Proper verb usage and idempotency
- **Status Codes** - Complete response code coverage
- **Request/Response** - JSON conventions, ISO 8601 dates
- **Headers** - Content-Type, Accept, request tracing
- **Errors** - RFC 9457 Problem Details format
- **Pagination** - Cursor-based with collection envelopes
- **Versioning** - URL versioning with deprecation lifecycle
- **Security** - Bearer tokens, HTTPS, OAuth 2.0
- **Performance** - Caching, compression, rate limiting
- **Events** - CloudEvents format, webhook signatures
- **Long-Running** - 202 Accepted with status polling

```bash
api-style lint openapi.yaml --profile default
```

### Azure Profile

Based on [Microsoft Azure REST API Guidelines](https://github.com/microsoft/api-guidelines):

- **Consistency** - Uniform interface patterns
- **Versioning** - Date-based API versions
- **Error Handling** - Standard error response format
- **Pagination** - Consistent collection handling
- **Long-running Operations** - Async patterns

```bash
api-style lint openapi.yaml --profile azure
```

### Google Profile

Based on [Google API Design Guide](https://cloud.google.com/apis/design):

- **Resource-Oriented Design** - Everything is a resource
- **Standard Methods** - List, Get, Create, Update, Delete
- **Naming Conventions** - Consistent field names
- **Error Model** - Structured error details
- **Documentation** - Comprehensive descriptions

```bash
api-style lint openapi.yaml --profile google
```

### Zalando Profile

Based on [Zalando RESTful API Guidelines](https://opensource.zalando.com/restful-api-guidelines/):

- **Pragmatic REST** - Practical over dogmatic
- **Hypermedia** - HATEOAS where appropriate
- **Events** - Async event patterns
- **Compatibility** - Evolution strategies
- **Security** - OAuth2 patterns

```bash
api-style lint openapi.yaml --profile zalando
```

## Listing Profile Rules

View all rules in a profile:

```bash
# List all rules
api-style list-rules --profile azure

# Filter by category
api-style list-rules --profile azure --category uri-design

# Filter by severity
api-style list-rules --profile azure --severity error
```

## Comparing Profiles

Different profiles may have different opinions on the same topic:

| Topic | Default | Zalando | Microsoft | Azure | Google |
|-------|---------|---------|-----------|-------|--------|
| Property case | camelCase | snake_case | camelCase | camelCase | snake_case |
| URL case | kebab-case | kebab-case | - | - | - |
| Versioning | URL path | URL/Header | URL path | Date-based | URL path |
| Date format | ISO 8601 | ISO 8601 | ISO 8601 | ISO 8601 | RFC 3339 |
| Error format | RFC 9457 | RFC 9457 | Custom | Azure format | Google format |
| Pagination | Cursor | Cursor | Offset | OData | Token |

For detailed coverage comparison, see [Profile Comparison](profile-comparison.md).

## Conformance Levels

Profiles support graduated compliance levels:

| Level | Description | Errors Allowed | Warnings Allowed |
|-------|-------------|----------------|------------------|
| Minimum | Basic API functionality | 5 | 20 |
| Standard | Production-ready APIs | 0 | 10 |
| Exemplary | Best-in-class design | 0 | 0 |

```bash
# Check against standard level
api-style lint openapi.yaml --profile comprehensive --level standard
```

Rules are tagged with minimum conformance levels. Higher levels include all rules from lower levels.

## Scoring Profiles

Evaluate profile quality and coverage:

```bash
# Score a profile against quality rubric
api-style score-profile comprehensive

# JSON output for analysis
api-style score-profile zalando --format json
```

See [Profile Scoring](profile-scoring.md) for details.

## Generating Documentation

Convert any profile into human-readable documentation:

```bash
# Single-page Markdown
api-style generate guide --profile zalando --output zalando-guide.md

# MkDocs multi-page site
api-style generate mkdocs --profile zalando --output ./zalando-docs
```

See [Documentation Generation](documentation-generation.md) for details.

## Next Steps

- [Profile Comparison](profile-comparison.md) - Detailed coverage analysis
- [Profile Scoring](profile-scoring.md) - Evaluate profile quality
- [Documentation Generation](documentation-generation.md) - Generate Markdown and MkDocs docs
- [Custom Rules](custom-rules.md) - Create your own rules
