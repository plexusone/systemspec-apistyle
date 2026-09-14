package gap

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func testReport() *types.MultiLintReport {
	report := types.NewMultiLintReport()
	report.AddFileReport("petstore.yaml", &types.LintReport{
		Violations: []types.Violation{
			{
				RuleID:    "naming-001",
				RuleTitle: "Use camelCase",
				Severity:  types.SeverityError,
				Category:  "naming",
				Message:   "Property 'user_name' must use camelCase",
				Path:      "$.paths./users.get.responses.200.content.application/json.schema.properties.user_name",
				Line:      42,
			},
			{
				RuleID:    "naming-001",
				RuleTitle: "Use camelCase",
				Severity:  types.SeverityError,
				Category:  "naming",
				Message:   "Property 'first_name' must use camelCase",
				Path:      "$.paths./users.get.responses.200.content.application/json.schema.properties.first_name",
				Line:      55,
			},
			{
				RuleID:    "errors-001",
				RuleTitle: "Use RFC 7807",
				Severity:  types.SeverityWarn,
				Category:  "errors",
				Message:   "Error response should follow RFC 7807",
				Path:      "$.paths./users.get.responses.400",
			},
			{
				RuleID:    "versioning-001",
				RuleTitle: "Version in URL",
				Severity:  types.SeverityInfo,
				Category:  "versioning",
				Message:   "Consider using URL versioning",
				Path:      "$.paths",
			},
		},
		Summary: &types.ViolationSummary{
			Total:    4,
			Errors:   2,
			Warnings: 1,
			Infos:    1,
		},
		Status: types.StatusFail,
	})
	return report
}

func testProfile() *types.APIStyleSpec {
	return &types.APIStyleSpec{
		Name:    "Test Profile",
		Version: "1.0.0",
		Categories: []types.Category{
			{ID: "naming", Title: "Naming Conventions"},
			{ID: "errors", Title: "Error Handling"},
			{ID: "versioning", Title: "Versioning"},
			{ID: "security", Title: "Security"},
		},
		CategoryGroups: []types.CategoryGroup{
			{ID: "fundamentals", Title: "Fundamentals", Categories: []string{"naming", "errors"}, Order: 1},
			{ID: "lifecycle", Title: "API Lifecycle", Categories: []string{"versioning"}, Order: 2},
			{ID: "operations", Title: "Operations", Categories: []string{"security"}, Order: 3},
		},
		Rules: []types.Rule{
			{ID: "naming-001", Category: "naming", Enforcement: &types.Enforcement{Type: types.EnforcementSpectral}},
			{ID: "naming-002", Category: "naming", Enforcement: &types.Enforcement{Type: types.EnforcementSpectral}},
			{ID: "errors-001", Category: "errors", Enforcement: &types.Enforcement{Type: types.EnforcementSpectral}},
			{ID: "security-001", Category: "security", Enforcement: &types.Enforcement{Type: types.EnforcementNone}},
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

	report := testReport()
	opts := &Options{
		Title:       "Test Gap Analysis",
		Theme:       "light",
		GeneratedAt: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	html, err := gen.Generate(context.Background(), report, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(html)

	checks := []struct {
		label    string
		contains string
	}{
		{"title", "Test Gap Analysis"},
		{"error count", ">2<"},
		{"warning count", ">1<"},
		{"total", ">4<"},
		{"rule naming-001", "naming-001"},
		{"rule title", "Use camelCase"},
		{"severity Error", "Error"},
		{"severity Warning", "Warning"},
		{"category naming", "Naming"},
		{"category errors", "Error"},
		{"date", "July 1, 2025"},
		{"violation message", "must use camelCase"},
	}

	for _, c := range checks {
		if !strings.Contains(result, c.contains) {
			t.Errorf("HTML missing %s: expected to contain %q", c.label, c.contains)
		}
	}
}

func TestGenerator_Generate_WithProfile(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	report := testReport()
	profile := testProfile()
	opts := &Options{
		Title:       "Coverage Analysis",
		GeneratedAt: time.Now(),
		Profile:     profile,
	}

	html, err := gen.Generate(context.Background(), report, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(html)

	checks := []struct {
		label    string
		contains string
	}{
		{"coverage section", "Category Coverage"},
		{"heatmap cell naming", "Naming Conventions"},
		{"heatmap cell errors", "Error Handling"},
		{"heatmap cell security", "Security"},
		{"uncovered areas", "Uncovered Areas"},
		{"LLM-only rule", "security-001"},
	}

	for _, c := range checks {
		if !strings.Contains(result, c.contains) {
			t.Errorf("HTML missing %s: expected to contain %q", c.label, c.contains)
		}
	}
}

func TestGenerator_Generate_WithProfile_GroupedHeatmap(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	report := testReport()
	profile := testProfile()
	opts := &Options{
		Title:       "Grouped Coverage",
		GeneratedAt: time.Now(),
		Profile:     profile,
	}

	html, err := gen.Generate(context.Background(), report, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	result := string(html)

	checks := []struct {
		label    string
		contains string
	}{
		{"group title fundamentals", "Fundamentals"},
		{"group title lifecycle", "API Lifecycle"},
		{"group title operations", "Operations"},
		{"group header class", "heatmap-group-header"},
		{"category in group", "Naming Conventions"},
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

	html, err := gen.Generate(context.Background(), testReport(), &Options{
		Title:       "Dark Report",
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

	html, err := gen.Generate(context.Background(), testReport(), nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(html) == 0 {
		t.Error("expected non-empty HTML with nil options")
	}
}

func TestGenerator_Generate_EmptyReport(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	report := types.NewMultiLintReport()
	html, err := gen.Generate(context.Background(), report, nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(html) == 0 {
		t.Error("expected non-empty HTML for empty report")
	}
}

func TestParseLintJSON_SingleReport(t *testing.T) {
	jsonData := []byte(`{
		"violations": [{"ruleId": "test-001", "severity": "error", "message": "test"}],
		"summary": {"total": 1, "errors": 1},
		"status": "fail",
		"metadata": {"specFile": "test.yaml"}
	}`)

	report, err := parseLintJSON(jsonData)
	if err != nil {
		t.Fatalf("parseLintJSON() error = %v", err)
	}

	if len(report.FileReports) != 1 {
		t.Fatalf("expected 1 file report, got %d", len(report.FileReports))
	}

	if report.FileReports[0].File != "test.yaml" {
		t.Errorf("expected file 'test.yaml', got %q", report.FileReports[0].File)
	}
}

func TestParseLintJSON_MultiReport(t *testing.T) {
	jsonData := []byte(`{
		"fileReports": [
			{
				"file": "a.yaml",
				"violations": [{"ruleId": "test-001", "severity": "error", "message": "a"}],
				"summary": {"total": 1, "errors": 1},
				"status": "fail"
			},
			{
				"file": "b.yaml",
				"violations": [],
				"summary": {"total": 0},
				"status": "pass"
			}
		],
		"summary": {"total": 1, "errors": 1},
		"status": "fail"
	}`)

	report, err := parseLintJSON(jsonData)
	if err != nil {
		t.Fatalf("parseLintJSON() error = %v", err)
	}

	if len(report.FileReports) != 2 {
		t.Fatalf("expected 2 file reports, got %d", len(report.FileReports))
	}
}
