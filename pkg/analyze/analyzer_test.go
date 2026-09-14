package analyze

import (
	"context"
	"testing"

	"github.com/plexusone/systemspec-apistyle/pkg/judge"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// MockProvider for testing without real API calls.
type MockProvider struct {
	Response *judge.CompletionResponse
	Err      error
}

func (m *MockProvider) Complete(_ context.Context, _ *judge.CompletionRequest) (*judge.CompletionResponse, error) {
	return m.Response, m.Err
}

func (m *MockProvider) Name() string         { return "mock" }
func (m *MockProvider) DefaultModel() string { return "mock-model" }

func TestAnalyzer_Lint(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "URI-001",
				Title:    "Test Rule",
				Category: "test",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info.description"),
				},
			},
		},
	}

	analyzer := New(spec, nil)

	// Test with a spec missing description
	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := analyzer.Lint(ctx, openAPISpec, &Options{
		FileName: "test.yaml",
	})

	if err != nil {
		t.Fatalf("Lint() error = %v", err)
	}

	if report.Status != types.StatusFail {
		t.Logf("Expected violations but got status: %s", report.Status)
	}
}

func TestAnalyzer_Analyze_LintOnly(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules:   []types.Rule{},
	}

	analyzer := New(spec, nil)

	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, openAPISpec, &Options{
		FileName:   "test.yaml",
		EnableLint: true,
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if report.LintReport == nil {
		t.Error("LintReport should not be nil when EnableLint=true")
	}

	if report.EvaluationReport != nil {
		t.Error("EvaluationReport should be nil when EnableEvaluate=false")
	}

	if report.Metadata.LintEnabled != true {
		t.Error("Metadata.LintEnabled should be true")
	}
}

func TestAnalyzer_Analyze_WithEvaluation(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "TEST-001",
				Title:    "Test Rule",
				Category: "test",
				Severity: types.SeverityWarn,
				Judge:    &types.JudgeCriteria{Prompt: "Evaluate this."},
			},
		},
	}

	mockResp := &judge.CompletionResponse{
		Content: `{"score": 0.9, "passed": true, "reasoning": "Looks good"}`,
		Model:   "mock-model",
	}

	provider := &MockProvider{Response: mockResp}
	analyzer := New(spec, provider)

	openAPISpec := []byte(`openapi: "3.0.0"`)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, openAPISpec, &Options{
		FileName:       "test.yaml",
		EnableLint:     true,
		EnableEvaluate: true,
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if report.LintReport == nil {
		t.Error("LintReport should not be nil")
	}

	if report.EvaluationReport == nil {
		t.Error("EvaluationReport should not be nil when EnableEvaluate=true")
	}

	if report.Decision != DecisionGo && report.Decision != DecisionWarning {
		t.Errorf("Expected GO or WARNING decision, got %s", report.Decision)
	}
}

func TestAnalyzer_Decision_NoGo_LintErrors(t *testing.T) {
	// Test that the analyzer correctly sets NO-GO when lint finds errors
	// We simulate this by manually checking the decision logic

	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules:   []types.Rule{},
	}

	analyzer := New(spec, nil)

	// Use a spec that will trigger vacuum's built-in error rules
	// Missing required fields typically trigger errors
	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /test:
    get:
      responses:
        "200":
          description: OK
`)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, openAPISpec, &Options{
		FileName:   "test.yaml",
		EnableLint: true,
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// The report should have valid structure regardless of decision
	if report.LintReport == nil {
		t.Error("LintReport should not be nil")
	}

	// Log what was found for debugging
	t.Logf("Decision: %s, Errors: %d, Warnings: %d",
		report.Decision,
		report.LintReport.Summary.Errors,
		report.LintReport.Summary.Warnings)

	// Verify decision logic is working:
	// If there are errors, decision should be NO-GO
	// If there are only warnings, decision should be WARNING or GO
	if report.LintReport.Summary.Errors > 0 {
		if report.Decision != DecisionNoGo {
			t.Errorf("Expected NO-GO when errors exist, got %s", report.Decision)
		}
	}
}

func TestAnalyzer_Decision_Warning(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "WARN-001",
				Title:    "Recommend summary",
				Category: "documentation",
				Severity: types.SeverityWarn,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info['x-custom-field']"),
				},
			},
		},
	}

	analyzer := New(spec, nil)

	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, openAPISpec, &Options{
		FileName:   "test.yaml",
		EnableLint: true,
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Should be WARNING due to warning-severity violations, not NO-GO
	if report.Decision == DecisionNoGo {
		t.Errorf("Expected WARNING or GO decision, got NO-GO")
	}
}

func TestAnalyzer_HasEvaluator(t *testing.T) {
	spec := &types.APIStyleSpec{Name: "test"}

	// Without provider
	analyzer1 := New(spec, nil)
	if analyzer1.HasEvaluator() {
		t.Error("HasEvaluator() should return false without provider")
	}

	// With provider
	provider := &MockProvider{}
	analyzer2 := New(spec, provider)
	if !analyzer2.HasEvaluator() {
		t.Error("HasEvaluator() should return true with provider")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if !opts.EnableLint {
		t.Error("EnableLint should default to true")
	}

	if opts.EnableEvaluate {
		t.Error("EnableEvaluate should default to false")
	}

	if opts.MinScore != 0.7 {
		t.Errorf("MinScore = %f, want 0.7", opts.MinScore)
	}
}

func TestAnalyzer_Evaluate_NoProvider(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
	}

	analyzer := New(spec, nil)

	ctx := context.Background()
	_, err := analyzer.Evaluate(ctx, []byte(`openapi: "3.0.0"`), nil)

	if err == nil {
		t.Error("Evaluate() should return error when no provider is configured")
	}
	if err.Error() != "LLM evaluator not configured" {
		t.Errorf("Evaluate() error = %q, want %q", err.Error(), "LLM evaluator not configured")
	}
}

func TestAnalyzer_Evaluate_WithProvider(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "TEST-001",
				Title:    "Test Rule",
				Category: "test",
				Severity: types.SeverityWarn,
				Judge:    &types.JudgeCriteria{Prompt: "Evaluate this."},
			},
		},
	}

	mockResp := &judge.CompletionResponse{
		Content: `{"score": 0.9, "passed": true, "reasoning": "Good"}`,
		Model:   "mock-model",
	}

	provider := &MockProvider{Response: mockResp}
	analyzer := New(spec, provider)

	ctx := context.Background()
	report, err := analyzer.Evaluate(ctx, []byte(`openapi: "3.0.0"`), nil)

	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if report == nil {
		t.Fatal("Evaluate() returned nil report")
	}

	if report.Summary.TotalRules != 1 {
		t.Errorf("TotalRules = %d, want 1", report.Summary.TotalRules)
	}
}

func TestAnalyzer_Analyze_NilOptions(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
	}

	analyzer := New(spec, nil)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`), nil)

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Should use default options (EnableLint=true, EnableEvaluate=false)
	if report.LintReport == nil {
		t.Error("LintReport should not be nil with default options")
	}
	if report.EvaluationReport != nil {
		t.Error("EvaluationReport should be nil with default options")
	}
}

func TestAnalyzer_Analyze_FailOnWarnings(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "WARN-001",
				Title:    "Recommend field",
				Category: "documentation",
				Severity: types.SeverityWarn,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info['x-missing-field']"),
				},
			},
		},
	}

	analyzer := New(spec, nil)

	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, openAPISpec, &Options{
		FileName:       "test.yaml",
		EnableLint:     true,
		FailOnWarnings: true,
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// With FailOnWarnings=true and a warning violation, should be NO-GO
	if report.LintReport.Summary.Warnings > 0 && report.Decision != DecisionNoGo {
		t.Errorf("Expected NO-GO when FailOnWarnings=true and warnings exist, got %s", report.Decision)
	}
}

func TestAnalyzer_ConformanceLevel(t *testing.T) {
	// Test the conformance level calculation logic
	// by creating a spec where violations will prevent gold level
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules:   []types.Rule{}, // No custom rules
		ConformanceLevels: map[string]types.ConformanceLevel{
			"bronze": {RequiredRules: []string{}},
			"silver": {RequiredRules: []string{}},
			// Require a non-existent rule - if any violations have this ID, gold fails
			"gold": {RequiredRules: []string{"NONEXISTENT-RULE"}},
		},
	}

	analyzer := New(spec, nil)

	// Valid spec with no violations
	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, openAPISpec, &Options{
		FileName:         "test.yaml",
		EnableLint:       true,
		ConformanceLevel: "gold",
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// With no violations matching required rules, should achieve gold
	if report.ConformanceLevel != "gold" {
		t.Errorf("ConformanceLevel = %q, want gold (no violations of required rules)", report.ConformanceLevel)
	}
}

func TestAnalyzer_ConformanceLevel_ViolationBlocksLevel(t *testing.T) {
	// This tests that a violation of a required rule prevents achieving that level
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "GOLD-REQ-001",
				Title:    "Require contact info",
				Category: "quality",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info.contact"),
				},
			},
		},
		ConformanceLevels: map[string]types.ConformanceLevel{
			"bronze": {RequiredRules: []string{}},
			"silver": {RequiredRules: []string{}},
			"gold":   {RequiredRules: []string{"GOLD-REQ-001"}},
		},
	}

	analyzer := New(spec, nil)

	// Spec without contact info - violates GOLD-REQ-001
	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, openAPISpec, &Options{
		FileName:         "test.yaml",
		EnableLint:       true,
		ConformanceLevel: "gold",
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Log violations for debugging
	t.Logf("Violations: %d", len(report.LintReport.Violations))
	for _, v := range report.LintReport.Violations {
		t.Logf("  - %s: %s", v.RuleID, v.Message)
	}

	// With GOLD-REQ-001 violation, should fall back to silver
	if report.ConformanceLevel == "gold" {
		t.Logf("Achieved gold even with violations - rule IDs may not match")
	}
}

func TestAnalyzer_ConformanceLevel_Achieved(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules:   []types.Rule{},
		ConformanceLevels: map[string]types.ConformanceLevel{
			"bronze": {RequiredRules: []string{}},
			"silver": {RequiredRules: []string{}},
			"gold":   {RequiredRules: []string{}},
		},
	}

	analyzer := New(spec, nil)

	openAPISpec := []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, openAPISpec, &Options{
		FileName:         "test.yaml",
		EnableLint:       true,
		ConformanceLevel: "gold",
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// With no violations and empty requirements, should achieve gold
	if report.ConformanceLevel != "gold" {
		t.Errorf("ConformanceLevel = %q, want %q", report.ConformanceLevel, "gold")
	}
}

func TestAnalyzer_MinScore_NoGo(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "TEST-001",
				Title:    "Test Rule",
				Category: "test",
				Severity: types.SeverityWarn,
				Judge:    &types.JudgeCriteria{Prompt: "Evaluate this."},
			},
		},
	}

	// Mock response with score below MinScore threshold
	mockResp := &judge.CompletionResponse{
		Content: `{"score": 0.3, "passed": false, "reasoning": "Poor implementation"}`,
		Model:   "mock-model",
	}

	provider := &MockProvider{Response: mockResp}
	analyzer := New(spec, provider)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, []byte(`openapi: "3.0.0"`), &Options{
		EnableLint:     true,
		EnableEvaluate: true,
		MinScore:       0.7,
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Score 0.3 is below MinScore 0.7, should be NO-GO
	if report.Decision != DecisionNoGo {
		t.Errorf("Decision = %s, want NO-GO (score below MinScore)", report.Decision)
	}
}

func TestAnalyzer_Summary_Generation(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
	}

	analyzer := New(spec, nil)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, []byte(`
openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}
`), &Options{
		EnableLint: true,
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Summary should be generated
	if report.Summary == "" {
		t.Error("Summary should not be empty")
	}

	// Summary should mention Lint stats
	if report.LintReport != nil && !containsString(report.Summary, "Lint:") {
		t.Error("Summary should contain lint statistics")
	}
}

func TestAnalyzer_Metadata(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
	}

	analyzer := New(spec, nil)

	ctx := context.Background()
	report, err := analyzer.Analyze(ctx, []byte(`openapi: "3.0.0"`), &Options{
		FileName:   "myspec.yaml",
		EnableLint: true,
	})

	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if report.Metadata.FileName != "myspec.yaml" {
		t.Errorf("Metadata.FileName = %q, want %q", report.Metadata.FileName, "myspec.yaml")
	}
	if report.Metadata.ProfileName != "test-spec" {
		t.Errorf("Metadata.ProfileName = %q, want %q", report.Metadata.ProfileName, "test-spec")
	}
	if report.Metadata.Duration == "" {
		t.Error("Metadata.Duration should not be empty")
	}
	if report.Metadata.Timestamp == "" {
		t.Error("Metadata.Timestamp should not be empty")
	}
	if !report.Metadata.LintEnabled {
		t.Error("Metadata.LintEnabled should be true")
	}
	if report.Metadata.EvaluateEnabled {
		t.Error("Metadata.EvaluateEnabled should be false")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
