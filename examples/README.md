# Examples

This directory contains example OpenAPI specifications and custom profiles for testing and learning systemspec-apistyle.

## Directory Structure

```
examples/
├── specs/                    # Example OpenAPI specifications
│   ├── petstore.yaml        # Classic Swagger Petstore example
│   └── ecommerce.yaml       # E-commerce API example
├── profiles/                 # Custom style profiles
│   └── custom-company.yaml  # Example company-specific profile
└── README.md
```

## OpenAPI Specifications

### Petstore (`specs/petstore.yaml`)

The classic Swagger Petstore API, updated for OpenAPI 3.1. Demonstrates:

- Pet management (CRUD operations)
- Store orders
- User accounts
- OAuth2 and API key authentication
- Pagination with cursors
- RFC 7807 error format

### E-Commerce (`specs/ecommerce.yaml`)

A comprehensive e-commerce API. Demonstrates:

- Product catalog with filtering and sorting
- Shopping cart management
- Order processing
- Customer profiles
- Product reviews
- JWT authentication
- Nested resources (products/{id}/reviews)

## Custom Profiles

### Custom Company (`profiles/custom-company.yaml`)

An example custom profile that:

- Extends the `default` profile
- Adds company-specific rules
- Overrides inherited rules
- Defines custom conformance levels

## Usage Examples

### Basic Linting

```bash
# Lint PetStore with default profile
api-style lint examples/specs/petstore.yaml

# Lint E-Commerce with Azure profile
api-style lint examples/specs/ecommerce.yaml --profile azure

# Lint with Google profile
api-style lint examples/specs/petstore.yaml --profile google
```

### Multi-File Linting

```bash
# Lint all YAML files in specs directory
api-style lint examples/specs/*.yaml

# Recursive search
api-style lint examples/ --recursive

# Multiple specific files
api-style lint examples/specs/petstore.yaml examples/specs/ecommerce.yaml
```

### Watch Mode

```bash
# Watch a file for changes
api-style lint examples/specs/petstore.yaml --watch

# Watch with specific profile
api-style lint examples/specs/ecommerce.yaml --profile azure --watch
```

### Using Configuration Files

```bash
# Create config in examples directory
cat > examples/.api-style.yaml << 'EOF'
profile: azure
level: silver
include:
  - "specs/*.yaml"
exclude:
  - "**/bad-api/**"
EOF

# Run lint (config auto-detected)
cd examples && api-style lint specs/petstore.yaml
```

### Git Pre-Commit Hook

```bash
# Install pre-commit hook (in a git repository)
api-style hooks init

# With specific profile
api-style hooks init --profile azure --level silver

# Test by staging an OpenAPI file
git add examples/specs/petstore.yaml
git commit -m "Test pre-commit"
```

### Using Custom Profiles

```bash
# Lint with a custom profile (by file path)
api-style lint examples/specs/ecommerce.yaml --profile examples/profiles/custom-company.yaml

# Or add to search path and use by name
export API_STYLE_PROFILES=examples/profiles
api-style lint examples/specs/ecommerce.yaml --profile custom-company
```

### JSON Output

```bash
# Get JSON report for CI integration
api-style lint examples/specs/petstore.yaml --format json --output report.json

# Pretty print JSON
api-style lint examples/specs/ecommerce.yaml --format json | jq .
```

### Conformance Levels

```bash
# Check bronze level compliance
api-style lint examples/specs/petstore.yaml --level bronze

# Check silver level compliance
api-style lint examples/specs/ecommerce.yaml --profile azure --level silver
```

### Full Analysis

```bash
# Combined lint + LLM evaluation (requires ANTHROPIC_API_KEY)
export ANTHROPIC_API_KEY=sk-ant-...
api-style analyze examples/specs/ecommerce.yaml --profile default

# Lint-only analysis (no LLM)
api-style analyze examples/specs/petstore.yaml --lint-only
```

### Using the MCP Server

```bash
# Start MCP server
mcp-api-style serve

# Or use CLI commands
mcp-api-style lint --file examples/specs/petstore.yaml --profile default
mcp-api-style list-rules --profile azure
mcp-api-style explain-rule URI-001
```

## Expected Results

### Petstore with Default Profile

The Petstore example demonstrates common API patterns. Expected:

- **Status**: PASS or minor warnings
- **Errors**: 0
- **Warnings**: ~2-3 (documentation completeness)

### E-Commerce with Default Profile

The E-Commerce example follows best practices. Expected:

- **Status**: PASS
- **Errors**: 0
- **Warnings**: 0-2

### Bad API Example

The `bad-api/openapi.yaml` contains intentional violations to demonstrate linting:

```bash
api-style lint examples/bad-api/openapi.yaml
```

Expected output:

```
Status: FAIL
Errors: 8, Warnings: 1

Errors:
  - [NAMING-001] Use camelCase for JSON properties
  - [URI-002] Use kebab-case for path segments
  - [URI-001] Use plural resource names

Warnings:
  - [URI-003] Avoid verbs in paths
```

Violations include:

| Violation | Example | Fix |
|-----------|---------|-----|
| Singular resource | `/user` | `/users` |
| Verb in path | `/getUserById` | `/users/{id}` |
| Snake_case path | `/order_items` | `/order-items` |
| Snake_case property | `first_name` | `firstName` |

### Common Violations to Explore

Try creating violations to see how they're reported:

1. **URI-001**: Change `/pets` to `/pet` (singular)
2. **HTTP-001**: Add `requestBody` to a GET operation
3. **DOC-001**: Remove the `info.description` field
4. **SEC-004**: Change server URL from `https://` to `http://`

## Creating Your Own Examples

1. Copy `specs/ecommerce.yaml` as a starting point
2. Modify to match your API design
3. Run linting to identify issues:
   ```bash
   api-style lint your-api.yaml --format json | jq '.violations'
   ```
4. Fix issues and iterate

## Creating Custom Profiles

1. Copy `profiles/custom-company.yaml` as a template
2. Modify rules, overrides, and conformance levels
3. Test with your specs:
   ```bash
   api-style lint your-api.yaml --profile your-profile.yaml
   ```

See the [Custom Rules Guide](../docs/guide/custom-rules.md) for detailed documentation.
