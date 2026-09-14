package generate

import (
	"strings"
	"testing"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func TestMarkdown_Basic(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:        "Test Guidelines",
		Version:     "1.0.0",
		Description: "Test API style guidelines.",
		Categories: []types.Category{
			{ID: "uri", Title: "URI Design", Description: "URL conventions"},
		},
		Rules: []types.Rule{
			{
				ID:        "URI-001",
				Title:     "Use plural nouns",
				Category:  "uri",
				Severity:  types.SeverityError,
				Rationale: "Plural nouns are clearer.",
				Examples: &types.Examples{
					Good: []string{"/users", "/orders"},
					Bad:  []string{"/user", "/order"},
				},
			},
		},
	}

	md, err := Markdown(spec, nil)
	if err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	// Check title
	if !strings.Contains(md, "# Test Guidelines") {
		t.Error("Missing title")
	}

	// Check description
	if !strings.Contains(md, "Test API style guidelines.") {
		t.Error("Missing description")
	}

	// Check rule
	if !strings.Contains(md, "URI-001") {
		t.Error("Missing rule ID")
	}

	if !strings.Contains(md, "Use plural nouns") {
		t.Error("Missing rule title")
	}

	// Check examples
	if !strings.Contains(md, "/users") {
		t.Error("Missing good example")
	}

	if !strings.Contains(md, "/user") {
		t.Error("Missing bad example")
	}
}

func TestMarkdown_WithConformanceLevels(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		ConformanceLevels: map[string]types.ConformanceLevel{
			"bronze": {
				Description:   "Basic compliance",
				RequiredRules: []string{"RULE-001"},
			},
			"silver": {
				Description:   "Standard compliance",
				RequiredRules: []string{"RULE-001", "RULE-002"},
				Extends:       "bronze",
			},
		},
		Rules: []types.Rule{
			{ID: "RULE-001", Title: "Rule 1", Category: "test"},
			{ID: "RULE-002", Title: "Rule 2", Category: "test"},
		},
	}

	opts := DefaultMarkdownOptions()
	opts.IncludeConformance = true

	md, err := Markdown(spec, opts)
	if err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	if !strings.Contains(md, "Conformance Levels") {
		t.Error("Missing conformance levels section")
	}

	if !strings.Contains(md, "Bronze") {
		t.Error("Missing bronze level")
	}

	if !strings.Contains(md, "Silver") {
		t.Error("Missing silver level")
	}
}

func TestMarkdown_WithMetadata(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Metadata: &types.SpecMetadata{
			Author:      "Test Author",
			URL:         "https://example.com",
			LastUpdated: "2024-01-01",
		},
		Rules: []types.Rule{},
	}

	opts := DefaultMarkdownOptions()
	opts.IncludeMetadata = true

	md, err := Markdown(spec, opts)
	if err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	if !strings.Contains(md, "Test Author") {
		t.Error("Missing author")
	}

	if !strings.Contains(md, "https://example.com") {
		t.Error("Missing URL")
	}
}

func TestSpectral_Basic(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "Test Guidelines",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "URI-001",
				Title:    "Use lowercase in URIs",
				Category: "uri",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "pattern",
					Given:    types.NewGivenPath("$.paths[*]~"),
					Options: &types.EnforcementOptions{
						Match: "^[a-z0-9/{}.-]*$",
					},
				},
			},
		},
	}

	yaml, err := Spectral(spec, nil)
	if err != nil {
		t.Fatalf("Spectral() error = %v", err)
	}

	// Check header
	if !strings.Contains(yaml, "Generated from: Test Guidelines") {
		t.Error("Missing generated header")
	}

	// Check extends
	if !strings.Contains(yaml, "extends:") {
		t.Error("Missing extends section")
	}

	// Check rule
	if !strings.Contains(yaml, "uri-001:") {
		t.Error("Missing rule ID (should be kebab-case)")
	}

	if !strings.Contains(yaml, "severity: error") {
		t.Error("Missing severity")
	}

	if !strings.Contains(yaml, "function: pattern") {
		t.Error("Missing function")
	}

	if !strings.Contains(yaml, "match:") {
		t.Error("Missing functionOptions")
	}
}

func TestSpectral_SkipsNonEnforceable(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Rules: []types.Rule{
			{
				ID:       "LLM-001",
				Title:    "LLM-only rule",
				Category: "test",
				Severity: types.SeverityWarn,
				// No enforcement - LLM only
				Judge: &types.JudgeCriteria{
					Prompt: "Check this rule",
				},
			},
		},
	}

	opts := DefaultSpectralOptions()
	opts.SkipNonEnforceable = true

	yaml, err := Spectral(spec, opts)
	if err != nil {
		t.Fatalf("Spectral() error = %v", err)
	}

	// Should not contain the rule
	if strings.Contains(yaml, "llm-001:") {
		t.Error("Should skip non-enforceable rules")
	}
}

func TestSpectral_MultipleGivenPaths(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Rules: []types.Rule{
			{
				ID:       "SEC-001",
				Title:    "Define security",
				Category: "security",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPaths("$.security", "$.components.securitySchemes"),
				},
			},
		},
	}

	yaml, err := Spectral(spec, nil)
	if err != nil {
		t.Fatalf("Spectral() error = %v", err)
	}

	// Should have multiple given paths
	if !strings.Contains(yaml, "given:") {
		t.Error("Missing given section")
	}

	if !strings.Contains(yaml, "$.security") {
		t.Error("Missing first given path")
	}

	if !strings.Contains(yaml, "$.components.securitySchemes") {
		t.Error("Missing second given path")
	}
}

// TestSpectral_TruthyWithMatch verifies that truthy function uses 'field'
// instead of 'functionOptions.match' when options.match is set.
// This is the correct Spectral format per the spec.
func TestSpectral_TruthyWithMatch(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Rules: []types.Rule{
			{
				ID:       "INFO-001",
				Title:    "API must have title",
				Category: "general",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info"),
					Options: &types.EnforcementOptions{
						Match: "title", // This should become field, not functionOptions.match
					},
				},
			},
		},
	}

	yaml, err := Spectral(spec, nil)
	if err != nil {
		t.Fatalf("Spectral() error = %v", err)
	}

	// Should have field: "title"
	if !strings.Contains(yaml, `field: "title"`) {
		t.Error("truthy function with options.match should output 'field', not 'functionOptions.match'")
	}

	// Should NOT have functionOptions.match
	if strings.Contains(yaml, "functionOptions:") && strings.Contains(yaml, `match: "title"`) {
		t.Error("truthy function should not have functionOptions.match - match should be field")
	}
}

func TestGroupRulesByCategory(t *testing.T) {
	rules := []types.Rule{
		{ID: "A-001", Category: "cat-a"},
		{ID: "A-002", Category: "cat-a"},
		{ID: "B-001", Category: "cat-b"},
		{ID: "C-001", Category: ""},
	}

	groups := GroupRulesByCategory(rules)

	if len(groups["cat-a"]) != 2 {
		t.Errorf("cat-a should have 2 rules, got %d", len(groups["cat-a"]))
	}

	if len(groups["cat-b"]) != 1 {
		t.Errorf("cat-b should have 1 rule, got %d", len(groups["cat-b"]))
	}

	if len(groups["other"]) != 1 {
		t.Errorf("other should have 1 rule, got %d", len(groups["other"]))
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"URI-001", "uri-001"},
		{"HTTP-METHOD-002", "http-method-002"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		result := toKebabCase(tt.input)
		if result != tt.expected {
			t.Errorf("toKebabCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestMarkdown_WithPrinciples(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Principles: []types.Principle{
			{
				ID:           "p1",
				Title:        "Consistency",
				Description:  "APIs should be consistent.",
				RelatedRules: []string{"RULE-001"},
			},
		},
		Rules: []types.Rule{
			{ID: "RULE-001", Title: "Rule 1", Category: "test"},
		},
	}

	opts := DefaultMarkdownOptions()
	opts.IncludePrinciples = true

	md, err := Markdown(spec, opts)
	if err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	if !strings.Contains(md, "Design Principles") {
		t.Error("Missing principles section")
	}

	if !strings.Contains(md, "Consistency") {
		t.Error("Missing principle title")
	}

	if !strings.Contains(md, "APIs should be consistent.") {
		t.Error("Missing principle description")
	}
}

func TestMarkdown_WithPatterns(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Patterns: []types.Pattern{
			{
				ID:       "cursor-pagination",
				Name:     "Cursor-Based Pagination",
				Summary:  "Use cursors for pagination.",
				Problem:  "Offset pagination has issues.",
				Solution: "Use opaque cursors.",
			},
		},
		Rules: []types.Rule{},
	}

	opts := DefaultMarkdownOptions()
	opts.IncludePatterns = true

	md, err := Markdown(spec, opts)
	if err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	if !strings.Contains(md, "Design Patterns") {
		t.Error("Missing patterns section")
	}

	if !strings.Contains(md, "Cursor-Based Pagination") {
		t.Error("Missing pattern name")
	}

	if !strings.Contains(md, "Problem:") {
		t.Error("Missing problem section")
	}
}

func TestMarkdown_WithGlossary(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Glossary: []types.GlossaryTerm{
			{
				Term:       "Resource",
				Definition: "An entity that can be manipulated via the API.",
				Aliases:    []string{"Entity"},
			},
		},
		Rules: []types.Rule{},
	}

	opts := DefaultMarkdownOptions()
	opts.IncludeGlossary = true

	md, err := Markdown(spec, opts)
	if err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	if !strings.Contains(md, "Glossary") {
		t.Error("Missing glossary section")
	}

	if !strings.Contains(md, "Resource") {
		t.Error("Missing term")
	}

	if !strings.Contains(md, "Entity") {
		t.Error("Missing alias")
	}
}

func TestMarkdown_WithDetailedExamples(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Rules: []types.Rule{
			{
				ID:       "TEST-001",
				Title:    "Test Rule",
				Category: "test",
				Examples: &types.Examples{
					Detailed: []types.DetailedExample{
						{
							Title:       "Collection Response",
							Description: "How to return a list",
							Type:        "good",
							Language:    "json",
							Code:        `{"items": []}`,
						},
					},
				},
			},
		},
	}

	opts := DefaultMarkdownOptions()
	opts.IncludeExamples = true

	md, err := Markdown(spec, opts)
	if err != nil {
		t.Fatalf("Markdown() error = %v", err)
	}

	if !strings.Contains(md, "Collection Response (Correct)") {
		t.Error("Missing detailed example title")
	}

	if !strings.Contains(md, "```json") {
		t.Error("Missing code block with language")
	}
}

func TestMkDocs_Basic(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:        "Test Guidelines",
		Version:     "1.0.0",
		Description: "Test API guidelines.",
		Categories: []types.Category{
			{ID: "uri", Title: "URI Design"},
		},
		Rules: []types.Rule{
			{ID: "URI-001", Title: "Use plural nouns", Category: "uri", Severity: types.SeverityError},
		},
	}

	result, err := MkDocs(spec, nil)
	if err != nil {
		t.Fatalf("MkDocs() error = %v", err)
	}

	// Check config
	if !strings.Contains(result.Config, "site_name: Test Guidelines") {
		t.Error("Missing site_name in config")
	}

	if !strings.Contains(result.Config, "theme:") {
		t.Error("Missing theme in config")
	}

	// Check index page
	if _, ok := result.Pages["index.md"]; !ok {
		t.Error("Missing index.md")
	}

	// Check rules index (split by default)
	if _, ok := result.Pages["rules/index.md"]; !ok {
		t.Error("Missing rules/index.md")
	}
}

func TestMkDocs_WithPatternsAndGlossary(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Patterns: []types.Pattern{
			{ID: "p1", Name: "Pattern 1", Summary: "Test pattern"},
		},
		Glossary: []types.GlossaryTerm{
			{Term: "API", Definition: "Application Programming Interface"},
		},
		Rules: []types.Rule{
			{ID: "R-001", Title: "Rule", Category: "general"},
		},
	}

	result, err := MkDocs(spec, nil)
	if err != nil {
		t.Fatalf("MkDocs() error = %v", err)
	}

	if _, ok := result.Pages["patterns.md"]; !ok {
		t.Error("Missing patterns.md")
	}

	if _, ok := result.Pages["glossary.md"]; !ok {
		t.Error("Missing glossary.md")
	}
}

func TestMkDocs_SplitPatterns(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "Test",
		Patterns: []types.Pattern{
			{ID: "cursor-pagination", Name: "Cursor Pagination", Summary: "Test"},
		},
		Rules: []types.Rule{
			{ID: "R-001", Title: "Rule", Category: "general"},
		},
	}

	opts := DefaultMkDocsOptions()
	opts.SplitPatterns = true

	result, err := MkDocs(spec, opts)
	if err != nil {
		t.Fatalf("MkDocs() error = %v", err)
	}

	if _, ok := result.Pages["patterns/index.md"]; !ok {
		t.Error("Missing patterns/index.md")
	}

	if _, ok := result.Pages["patterns/cursor-pagination.md"]; !ok {
		t.Error("Missing individual pattern page")
	}
}

// TestSpectral_RegexEscaping verifies that regex patterns with backslashes
// are correctly escaped in YAML output (not double-escaped).
// Bug: https://github.com/plexusone/systemspec-apistyle/issues/BUG-spectral-regex-escaping
func TestSpectral_RegexEscaping(t *testing.T) {
	// JSON profile contains: "match": "^(\\/[a-z][a-z0-9\\-]*(\\/\\{[^}]+})?)*$"
	// After JSON unmarshaling:
	//   - JSON `\\` becomes Go `\` (single backslash)
	//   - JSON `\/` is just `/` (forward slashes can be escaped in JSON, but decode to just /)
	//
	// Wait - JSON `\\/` is actually `\\` + `/` = backslash + forward-slash after parsing
	//
	// So JSON "^(\\/[a-z]..." decodes to Go string `^(\/[a-z]...` where:
	//   - `\/` = backslash followed by forward-slash
	//   - `\-` = backslash followed by hyphen
	//
	// When written to YAML with %q:
	//   - Go `\` becomes YAML `\\`
	//
	// Expected YAML: "^(\\/[a-z][a-z0-9\\-]*(\\/\\{[^}]+})?)*$"

	spec := &types.APIStyleSpec{
		Name: "Test",
		Rules: []types.Rule{
			{
				ID:       "EXT-012",
				Title:    "Use kebab-case paths",
				Category: "uri",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "pattern",
					Given:    types.NewGivenPath("$.paths[*]~"),
					Options: &types.EnforcementOptions{
						// This simulates the Go string AFTER JSON parsing of:
						// JSON: "^(\\/[a-z][a-z0-9\\-]*(\\/\\{[^}]+})?)*$"
						// Go raw string with equivalent content:
						Match: `^(\/[a-z][a-z0-9\-]*(\/\{[^}]+})?)*$`,
					},
				},
			},
		},
	}

	yaml, err := Spectral(spec, nil)
	if err != nil {
		t.Fatalf("Spectral() error = %v", err)
	}

	// The YAML should have double-backslash (\\) for each single backslash in the pattern.
	// This is correct YAML escaping for double-quoted strings.
	// Go string `\/` → YAML `\\/`
	// Go string `\{` → YAML `\\{`
	expectedMatch := `"^(\\/[a-z][a-z0-9\\-]*(\\/\\{[^}]+})?)*$"`

	if !strings.Contains(yaml, expectedMatch) {
		// Find what we actually got
		lines := strings.Split(yaml, "\n")
		var actualMatch string
		for _, line := range lines {
			if strings.Contains(line, "match:") {
				actualMatch = strings.TrimSpace(line)
				break
			}
		}
		t.Errorf("Regex escaping incorrect.\nExpected match: %s\nActual line: %s", expectedMatch, actualMatch)
	}

	// Should NOT have quadruple backslashes (\\\\) - that would be double-escaping bug
	if strings.Contains(yaml, `\\\\`) {
		t.Error("Found quadruple backslashes (\\\\\\\\) - this indicates double-escaping bug")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"uri-design", "uri-design"},
		{"URI Design", "uri-design"},
		{"http_methods", "http-methods"},
		{"test@#$file", "testfile"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
