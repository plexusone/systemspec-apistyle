package guide

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func testSpec() *types.APIStyleSpec {
	return &types.APIStyleSpec{
		Name:         "Test API Style",
		Version:      "1.0.0",
		Description:  "A test API style specification.",
		Introduction: "This guide defines the rules for test APIs.",
		Principles: []types.Principle{
			{ID: "consistency", Title: "Consistency", Description: "Be consistent."},
		},
		Categories: []types.Category{
			{ID: "naming", Title: "Naming Conventions", Description: "Rules about naming."},
			{ID: "errors", Title: "Error Handling", Description: "Rules about errors."},
		},
		Rules: []types.Rule{
			{
				ID:          "naming-001",
				Title:       "Use camelCase",
				Category:    "naming",
				Severity:    types.SeverityError,
				Description: "All properties must use camelCase.",
				Rationale:   "Consistency improves developer experience.",
				Examples: &types.Examples{
					Good: []string{"userName is correct"},
					Bad:  []string{"user_name is incorrect"},
				},
			},
			{
				ID:          "errors-001",
				Title:       "Use RFC 7807",
				Category:    "errors",
				Severity:    types.SeverityWarn,
				Description: "Errors must follow RFC 7807 format.",
			},
		},
		ConformanceLevels: map[string]types.ConformanceLevel{
			"bronze": {
				Description:        "Minimum compliance",
				RequiredRules:      []string{"naming-001"},
				RequiredCategories: []string{"naming"},
			},
		},
		Patterns: []types.Pattern{
			{
				ID:      "pagination",
				Name:    "Pagination",
				Summary: "Cursor-based pagination pattern.",
				Problem: "Large result sets need pagination.",
			},
		},
		Glossary: []types.GlossaryTerm{
			{Term: "API", Definition: "Application Programming Interface"},
		},
	}
}

func TestNew(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if gen == nil {
		t.Fatal("New() returned nil generator")
	}
}

func TestGenerator_Generate(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	spec := testSpec()
	opts := &Options{
		Title:       "Test Guide",
		Theme:       "light",
		GeneratedAt: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),

		IncludeTOC:          true,
		IncludeExamples:     true,
		IncludeRationale:    true,
		IncludeReferences:   true,
		IncludeConformance:  true,
		IncludeMetadata:     true,
		IncludeIntroduction: true,
		IncludePrinciples:   true,
		IncludePatterns:     true,
		IncludeGlossary:     true,
	}

	html, err := gen.Generate(context.Background(), spec, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(html)

	checks := []struct {
		label    string
		contains string
	}{
		{"title", "Test Guide"},
		{"version", "1.0.0"},
		{"introduction", "This guide defines the rules"},
		{"principle", "Consistency"},
		{"category naming", "Naming Conventions"},
		{"category errors", "Error Handling"},
		{"rule naming-001", "naming-001"},
		{"rule title", "Use camelCase"},
		{"rule errors-001", "errors-001"},
		{"severity error", "Error"},
		{"severity warn", "Warning"},
		{"good example", "userName is correct"},
		{"bad example", "user_name is incorrect"},
		{"conformance", "Bronze"},
		{"pattern", "Pagination"},
		{"glossary term", "API"},
		{"glossary def", "Application Programming Interface"},
		{"date", "July 1, 2025"},
	}

	for _, c := range checks {
		if !strings.Contains(result, c.contains) {
			t.Errorf("HTML missing %s: expected to contain %q", c.label, c.contains)
		}
	}
}

func TestGenerator_Generate_DarkTheme(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	html, err := gen.Generate(context.Background(), testSpec(), &Options{
		Title:       "Dark Guide",
		Theme:       "dark",
		GeneratedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(html)
	if !strings.Contains(result, "--bg-primary: #0f172a") {
		t.Error("dark theme CSS vars not applied")
	}
}

func TestGenerator_Generate_NilOptions(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	html, err := gen.Generate(context.Background(), testSpec(), nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(html) == 0 {
		t.Error("expected non-empty HTML with nil options")
	}
}

func TestGenerator_Generate_MinimalSpec(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	spec := &types.APIStyleSpec{
		Name:    "Minimal",
		Version: "0.1.0",
	}

	html, err := gen.Generate(context.Background(), spec, nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(html)
	if !strings.Contains(result, "Minimal") {
		t.Error("HTML should contain spec name")
	}
}
