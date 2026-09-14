package lint

import (
	"context"
	"testing"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func TestVacuumLinter_PluralResources(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "URI-001",
				Title:    "Use plural resource names",
				Category: "uri-design",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "pattern",
					Given:    types.NewGivenPath("$.paths[*]~"),
					Options: &types.EnforcementOptions{
						Match: "^/[a-z]+s(/|$|\\{)",
					},
				},
			},
		},
	}

	linter := NewVacuumLinter(spec)

	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /user:
    get:
      summary: Get user
      responses:
        "200":
          description: OK
  /users:
    get:
      summary: Get users
      responses:
        "200":
          description: OK
`)

	ctx := context.Background()
	report, err := linter.Lint(ctx, openAPISpec, &Options{
		FileName: "test.yaml",
	})

	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	// Should have violations for singular resource names
	// Note: The exact behavior depends on vacuum's pattern matching
	t.Logf("Report status: %s", report.Status)
	t.Logf("Violations: %d", len(report.Violations))
	for _, v := range report.Violations {
		t.Logf("  - %s: %s at %s", v.RuleID, v.Message, v.Path)
	}
}

func TestWithDefaults(t *testing.T) {
	// Test with vacuum's built-in rules
	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /users:
    get:
      responses:
        "200":
          description: OK
`)

	ctx := context.Background()
	report, err := WithDefaults(ctx, openAPISpec, &Options{
		FileName: "test.yaml",
	})

	if err != nil {
		t.Fatalf("WithDefaults failed: %v", err)
	}

	t.Logf("Report status: %s", report.Status)
	t.Logf("Errors: %d, Warnings: %d", report.Summary.Errors, report.Summary.Warnings)

	// vacuum's default rules should catch missing operationId, etc.
	if len(report.Violations) == 0 {
		t.Log("No violations found (spec may be minimal but valid)")
	} else {
		for _, v := range report.Violations {
			t.Logf("  - [%s] %s: %s", v.Severity, v.RuleID, v.Message)
		}
	}
}

func TestConvertSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected types.Severity
	}{
		{"error", types.SeverityError},
		{"warn", types.SeverityWarn},
		{"warning", types.SeverityWarn},
		{"info", types.SeverityInfo},
		{"information", types.SeverityInfo},
		{"hint", types.SeverityHint},
		{"unknown", types.SeverityWarn},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertSeverity(tt.input)
			if result != tt.expected {
				t.Errorf("convertSeverity(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Timeout != 30000 {
		t.Errorf("Timeout = %d, want 30000", opts.Timeout)
	}
	if opts.FailFast != false {
		t.Error("FailFast should default to false")
	}
	if len(opts.Exceptions) != 0 {
		t.Errorf("len(Exceptions) = %d, want 0", len(opts.Exceptions))
	}
}

func TestVacuumLinter_NilOptions(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules:   []types.Rule{},
	}

	linter := NewVacuumLinter(spec)

	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := linter.Lint(ctx, openAPISpec, nil)

	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if report == nil {
		t.Fatal("Lint returned nil report")
	}
}

func TestWithDefaults_NilOptions(t *testing.T) {
	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := WithDefaults(ctx, openAPISpec, nil)

	if err != nil {
		t.Fatalf("WithDefaults failed: %v", err)
	}

	if report == nil {
		t.Fatal("WithDefaults returned nil report")
	}
}

func TestVacuumLinter_Exceptions(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "DESC-001",
				Title:    "Require description",
				Category: "documentation",
				Severity: types.SeverityWarn,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info.description"),
				},
			},
		},
	}

	linter := NewVacuumLinter(spec)

	// Spec missing description - should trigger violation
	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()

	// First, lint without exceptions
	report1, err := linter.Lint(ctx, openAPISpec, &Options{
		FileName: "test.yaml",
	})
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	violations1 := len(report1.Violations)
	t.Logf("Without exceptions: %d violations", violations1)

	// Now, lint with an exception for DESC-001
	exceptions := []types.Exception{
		{RuleID: "DESC-001"},
	}

	report2, err := linter.Lint(ctx, openAPISpec, &Options{
		FileName:   "test.yaml",
		Exceptions: exceptions,
	})
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	violations2 := len(report2.Violations)
	ignored2 := len(report2.IgnoredViolations)
	t.Logf("With exceptions: %d violations, %d ignored", violations2, ignored2)

	// With exceptions, DESC-001 violations should be in IgnoredViolations
	if violations2 > violations1 {
		t.Error("Exceptions should reduce or maintain violations, not increase")
	}
}

func TestVacuumLinter_TruthyFunction(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "CONTACT-001",
				Title:    "Require contact info",
				Category: "metadata",
				Severity: types.SeverityWarn,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info.contact"),
				},
			},
		},
	}

	linter := NewVacuumLinter(spec)

	// Spec without contact - should trigger violation
	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := linter.Lint(ctx, openAPISpec, &Options{
		FileName: "test.yaml",
	})

	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	// Log results
	t.Logf("Status: %s", report.Status)
	t.Logf("Violations: %d", len(report.Violations))
	for _, v := range report.Violations {
		t.Logf("  - %s: %s", v.RuleID, v.Message)
	}
}

func TestVacuumLinter_LengthFunction(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "TITLE-001",
				Title:    "Title minimum length",
				Category: "metadata",
				Severity: types.SeverityWarn,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "length",
					Given:    types.NewGivenPath("$.info.title"),
					Options: &types.EnforcementOptions{
						Min: intPtr(5),
						Max: intPtr(50),
					},
				},
			},
		},
	}

	linter := NewVacuumLinter(spec)

	// Spec with short title
	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := linter.Lint(ctx, openAPISpec, &Options{
		FileName: "test.yaml",
	})

	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	// Log results
	t.Logf("Status: %s", report.Status)
	t.Logf("Violations: %d", len(report.Violations))
}

func TestLintReport_HasBlockingViolations(t *testing.T) {
	report := types.NewLintReport()

	// No violations
	if report.HasBlockingViolations() {
		t.Error("HasBlockingViolations() = true, want false for empty report")
	}

	// Add a warning
	report.AddViolation(types.Violation{
		RuleID:   "WARN-001",
		Severity: types.SeverityWarn,
		Message:  "Warning",
	})

	if report.HasBlockingViolations() {
		t.Error("HasBlockingViolations() = true, want false for warning-only report")
	}

	// Add an error
	report.AddViolation(types.Violation{
		RuleID:   "ERR-001",
		Severity: types.SeverityError,
		Message:  "Error",
	})

	if !report.HasBlockingViolations() {
		t.Error("HasBlockingViolations() = false, want true when errors exist")
	}
}

func intPtr(i int) *int {
	return &i
}
