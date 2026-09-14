package judge

import (
	"context"
	"errors"
	"testing"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func TestBuildRubricSet(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test-spec",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "URI-001",
				Title:    "Use plural resource names",
				Category: "uri-design",
				Severity: types.SeverityError,
				Judge: &types.JudgeCriteria{
					Prompt: "Check if all resource names in paths use plural form.",
					Weight: 1.0,
				},
			},
			{
				ID:       "URI-002",
				Title:    "Use lowercase in URIs",
				Category: "uri-design",
				Severity: types.SeverityWarn,
				// No Judge criteria - should be skipped
			},
			{
				ID:       "DOC-001",
				Title:    "Provide operation descriptions",
				Category: "documentation",
				Severity: types.SeverityInfo,
				Judge: &types.JudgeCriteria{
					Prompt: "Check if all operations have meaningful descriptions.",
					Weight: 0.5,
				},
			},
		},
	}

	rs := BuildRubricSet(spec)

	if rs.Name != "test-spec" {
		t.Errorf("Name = %q, want %q", rs.Name, "test-spec")
	}

	if rs.Size() != 2 {
		t.Errorf("Size() = %d, want 2 (rules with Judge criteria)", rs.Size())
	}

	// Check that URI-002 was skipped (no Judge criteria)
	if _, ok := rs.Criteria["URI-002"]; ok {
		t.Error("URI-002 should not be included (no Judge criteria)")
	}

	// Check URI-001
	if c, ok := rs.Criteria["URI-001"]; ok {
		if c.Weight != 1.0 {
			t.Errorf("URI-001 Weight = %f, want 1.0", c.Weight)
		}
		if c.Severity != types.SeverityError {
			t.Errorf("URI-001 Severity = %v, want Error", c.Severity)
		}
	} else {
		t.Error("URI-001 should be included")
	}

	// Check categories
	if len(rs.Categories["uri-design"]) != 1 {
		t.Errorf("uri-design category should have 1 criterion, got %d", len(rs.Categories["uri-design"]))
	}
	if len(rs.Categories["documentation"]) != 1 {
		t.Errorf("documentation category should have 1 criterion, got %d", len(rs.Categories["documentation"]))
	}
}

func TestRubricSetFilters(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "test-spec",
		Rules: []types.Rule{
			{ID: "A-001", Category: "cat-a", Judge: &types.JudgeCriteria{Prompt: "test"}},
			{ID: "A-002", Category: "cat-a", Judge: &types.JudgeCriteria{Prompt: "test"}},
			{ID: "B-001", Category: "cat-b", Judge: &types.JudgeCriteria{Prompt: "test"}},
		},
	}

	rs := BuildRubricSet(spec)

	// Test FilterByCategory
	catA := rs.FilterByCategory("cat-a")
	if len(catA) != 2 {
		t.Errorf("FilterByCategory('cat-a') returned %d, want 2", len(catA))
	}

	catB := rs.FilterByCategory("cat-b")
	if len(catB) != 1 {
		t.Errorf("FilterByCategory('cat-b') returned %d, want 1", len(catB))
	}

	// Test FilterByRuleIDs
	filtered := rs.FilterByRuleIDs([]string{"A-001", "B-001"})
	if len(filtered) != 2 {
		t.Errorf("FilterByRuleIDs returned %d, want 2", len(filtered))
	}

	// Test CategoryNames
	names := rs.CategoryNames()
	if len(names) != 2 {
		t.Errorf("CategoryNames() returned %d names, want 2", len(names))
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "json in code block",
			input: `Here's my evaluation:

` + "```json" + `
{
  "score": 0.8,
  "passed": true,
  "reasoning": "Most endpoints use plural names",
  "examples": ["/users", "/orders"],
  "suggestions": ["Consider renaming /user to /users"],
  "locations": ["$.paths./user"]
}
` + "```",
			wantErr: false,
		},
		{
			name: "raw json",
			input: `{
  "score": 0.5,
  "passed": true,
  "reasoning": "test"
}`,
			wantErr: false,
		},
		{
			name:    "no json",
			input:   "This is just plain text without any JSON.",
			wantErr: true,
		},
	}

	criterion := &Criterion{
		RuleID:   "TEST-001",
		Category: "test",
		Severity: types.SeverityWarn,
		Weight:   1.0,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSingleEvaluation(tt.input, criterion)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSingleEvaluation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseSingleEvaluation(t *testing.T) {
	input := `{
  "score": 0.75,
  "passed": true,
  "reasoning": "The API mostly follows the rule",
  "examples": ["/users endpoint is correct"],
  "suggestions": ["Consider adding more documentation"],
  "locations": ["$.paths./users"]
}`

	criterion := &Criterion{
		RuleID:    "TEST-001",
		RuleTitle: "Test Rule",
		Category:  "test",
		Severity:  types.SeverityError,
		Weight:    0.8,
	}

	finding, err := ParseSingleEvaluation(input, criterion)
	if err != nil {
		t.Fatalf("ParseSingleEvaluation() error = %v", err)
	}

	if finding.RuleID != "TEST-001" {
		t.Errorf("RuleID = %q, want %q", finding.RuleID, "TEST-001")
	}
	if finding.Score != 0.75 {
		t.Errorf("Score = %f, want 0.75", finding.Score)
	}
	if !finding.Passed {
		t.Error("Passed = false, want true")
	}
	if finding.Severity != types.SeverityError {
		t.Errorf("Severity = %v, want Error", finding.Severity)
	}
	if finding.Weight != 0.8 {
		t.Errorf("Weight = %f, want 0.8", finding.Weight)
	}
	if len(finding.Examples) != 1 {
		t.Errorf("len(Examples) = %d, want 1", len(finding.Examples))
	}
}

func TestEvaluationReport(t *testing.T) {
	report := NewEvaluationReport()

	// Add findings
	report.AddFinding(Finding{
		RuleID:   "A-001",
		Category: "cat-a",
		Score:    0.8,
		Passed:   true,
		Severity: types.SeverityError,
		Weight:   1.0,
	})
	report.AddFinding(Finding{
		RuleID:   "A-002",
		Category: "cat-a",
		Score:    0.6,
		Passed:   true,
		Severity: types.SeverityWarn,
		Weight:   1.0,
	})
	report.AddFinding(Finding{
		RuleID:   "B-001",
		Category: "cat-b",
		Score:    0.3,
		Passed:   false,
		Severity: types.SeverityError,
		Weight:   1.0,
	})

	report.CalculateScores()

	// Check summary
	if report.Summary.TotalRules != 3 {
		t.Errorf("TotalRules = %d, want 3", report.Summary.TotalRules)
	}
	if report.Summary.PassedRules != 2 {
		t.Errorf("PassedRules = %d, want 2", report.Summary.PassedRules)
	}
	if report.Summary.FailedRules != 1 {
		t.Errorf("FailedRules = %d, want 1", report.Summary.FailedRules)
	}

	// Check overall score (average of 0.8, 0.6, 0.3 = ~0.567)
	expectedScore := (0.8 + 0.6 + 0.3) / 3
	if report.Summary.OverallScore < expectedScore-0.01 || report.Summary.OverallScore > expectedScore+0.01 {
		t.Errorf("OverallScore = %f, want ~%f", report.Summary.OverallScore, expectedScore)
	}

	// Check status (should fail due to error-severity failure)
	if report.Status != types.StatusFail {
		t.Errorf("Status = %v, want Fail", report.Status)
	}

	if !report.HasFailures() {
		t.Error("HasFailures() = false, want true")
	}
	if !report.HasCriticalFailures() {
		t.Error("HasCriticalFailures() = false, want true")
	}
}

func TestPromptBuilder(t *testing.T) {
	pb := NewPromptBuilder()

	criterion := &Criterion{
		RuleID:       "TEST-001",
		RuleTitle:    "Test Rule",
		Category:     "test",
		Prompt:       "Evaluate this rule.",
		Severity:     types.SeverityWarn,
		GoodExamples: []string{"/users"},
		BadExamples:  []string{"/user"},
		Rationale:    "Plural names are clearer.",
	}

	specContent := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"`

	prompt := pb.BuildSingleEvaluation(criterion, specContent)

	// Check that key elements are present
	if !contains(prompt, "TEST-001") {
		t.Error("Prompt should contain rule ID")
	}
	if !contains(prompt, "Test Rule") {
		t.Error("Prompt should contain rule title")
	}
	if !contains(prompt, "Evaluate this rule.") {
		t.Error("Prompt should contain evaluation criteria")
	}
	if !contains(prompt, "Plural names are clearer.") {
		t.Error("Prompt should contain rationale")
	}
	if !contains(prompt, "/users") {
		t.Error("Prompt should contain good example")
	}
	if !contains(prompt, "Test API") {
		t.Error("Prompt should contain spec content")
	}
}

func TestAggregateScores(t *testing.T) {
	tests := []struct {
		name     string
		scores   []float64
		weights  []float64
		expected float64
	}{
		{
			name:     "equal weights (empty)",
			scores:   []float64{0.8, 0.6, 0.4},
			weights:  []float64{},
			expected: 0.6,
		},
		{
			name:     "with weights",
			scores:   []float64{1.0, 0.0},
			weights:  []float64{3.0, 1.0},
			expected: 0.75,
		},
		{
			name:     "empty scores",
			scores:   []float64{},
			weights:  []float64{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateScores(tt.scores, tt.weights)
			if result != tt.expected {
				t.Errorf("AggregateScores() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestTruncateSpec(t *testing.T) {
	short := "short content"
	truncated, wasTruncated := TruncateSpec(short, 100)
	if wasTruncated {
		t.Error("Short content should not be truncated")
	}
	if truncated != short {
		t.Error("Short content should be unchanged")
	}

	long := "line1\nline2\nline3\nline4\nline5"
	truncated, wasTruncated = TruncateSpec(long, 15)
	if !wasTruncated {
		t.Error("Long content should be truncated")
	}
	if !contains(truncated, "truncated") {
		t.Error("Truncated content should contain truncation marker")
	}
}

// MockProvider for testing without real API calls.
type MockProvider struct {
	Response *CompletionResponse
	Err      error
}

func (m *MockProvider) Complete(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
	return m.Response, m.Err
}

func (m *MockProvider) Name() string         { return "mock" }
func (m *MockProvider) DefaultModel() string { return "mock-model" }

func TestClaudeEvaluator_Evaluate(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "test-spec",
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

	mockResp := &CompletionResponse{
		Content: `{"score": 0.8, "passed": true, "reasoning": "Looks good"}`,
		Model:   "mock-model",
	}

	provider := &MockProvider{Response: mockResp}
	evaluator := NewClaudeEvaluator(provider, spec)

	ctx := context.Background()
	specBytes := []byte(`openapi: "3.0.0"`)

	report, err := evaluator.Evaluate(ctx, specBytes, nil)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if report.Summary.TotalRules != 1 {
		t.Errorf("TotalRules = %d, want 1", report.Summary.TotalRules)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("len(Findings) = %d, want 1", len(report.Findings))
	}
	if report.Findings[0].Score != 0.8 {
		t.Errorf("Score = %f, want 0.8", report.Findings[0].Score)
	}
}

func TestClaudeEvaluator_Evaluate_SpecTruncated(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "test-spec",
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

	mockResp := &CompletionResponse{
		Content: `{"score": 0.8, "passed": true, "reasoning": "Looks good"}`,
		Model:   "mock-model",
	}
	evaluator := NewClaudeEvaluator(&MockProvider{Response: mockResp}, spec)
	ctx := context.Background()

	report, err := evaluator.Evaluate(ctx, []byte(`openapi: "3.0.0"`), nil)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if report.Metadata.SpecTruncated {
		t.Error("SpecTruncated = true for small spec, want false")
	}

	huge := make([]byte, maxSpecChars+1)
	for i := range huge {
		huge[i] = 'a'
	}
	report, err = evaluator.Evaluate(ctx, huge, nil)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !report.Metadata.SpecTruncated {
		t.Error("SpecTruncated = false for oversized spec, want true")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNewClaudeEvaluatorWithRubric(t *testing.T) {
	rubricSet := &RubricSet{
		Name:       "custom-rubric",
		Criteria:   make(map[string]*Criterion),
		Categories: make(map[string][]*Criterion),
	}
	rubricSet.Criteria["TEST-001"] = &Criterion{
		RuleID:   "TEST-001",
		Category: "test",
	}
	rubricSet.Categories["test"] = []*Criterion{rubricSet.Criteria["TEST-001"]}

	provider := &MockProvider{}
	evaluator := NewClaudeEvaluatorWithRubric(provider, rubricSet)

	if evaluator.RubricSet().Name != "custom-rubric" {
		t.Errorf("RubricSet().Name = %q, want %q", evaluator.RubricSet().Name, "custom-rubric")
	}
	if evaluator.RubricSet().Size() != 1 {
		t.Errorf("RubricSet().Size() = %d, want 1", evaluator.RubricSet().Size())
	}
}

func TestClaudeEvaluator_Evaluate_EmptyCriteria(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:  "empty-spec",
		Rules: []types.Rule{}, // No rules with Judge criteria
	}

	provider := &MockProvider{}
	evaluator := NewClaudeEvaluator(provider, spec)

	ctx := context.Background()
	report, err := evaluator.Evaluate(ctx, []byte(`openapi: "3.0.0"`), &Options{
		FileName: "test.yaml",
	})

	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if report.Summary.TotalRules != 0 {
		t.Errorf("TotalRules = %d, want 0", report.Summary.TotalRules)
	}
	if report.Metadata.FileName != "test.yaml" {
		t.Errorf("Metadata.FileName = %q, want %q", report.Metadata.FileName, "test.yaml")
	}
}

func TestClaudeEvaluator_Evaluate_FilterByRuleIDs(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "filter-spec",
		Rules: []types.Rule{
			{ID: "A-001", Judge: &types.JudgeCriteria{Prompt: "test"}},
			{ID: "A-002", Judge: &types.JudgeCriteria{Prompt: "test"}},
			{ID: "B-001", Judge: &types.JudgeCriteria{Prompt: "test"}},
		},
	}

	mockResp := &CompletionResponse{
		Content: `{"score": 0.9, "passed": true, "reasoning": "Good"}`,
		Model:   "mock-model",
	}

	provider := &MockProvider{Response: mockResp}
	evaluator := NewClaudeEvaluator(provider, spec)

	ctx := context.Background()
	report, err := evaluator.Evaluate(ctx, []byte(`openapi: "3.0.0"`), &Options{
		RuleIDs: []string{"A-001"}, // Only evaluate A-001
	})

	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if report.Summary.TotalRules != 1 {
		t.Errorf("TotalRules = %d, want 1 (filtered by RuleIDs)", report.Summary.TotalRules)
	}
}

func TestClaudeEvaluator_Evaluate_FilterByCategories(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "filter-spec",
		Rules: []types.Rule{
			{ID: "A-001", Category: "cat-a", Judge: &types.JudgeCriteria{Prompt: "test"}},
			{ID: "A-002", Category: "cat-a", Judge: &types.JudgeCriteria{Prompt: "test"}},
			{ID: "B-001", Category: "cat-b", Judge: &types.JudgeCriteria{Prompt: "test"}},
		},
	}

	mockResp := &CompletionResponse{
		Content: `{"score": 0.9, "passed": true, "reasoning": "Good"}`,
		Model:   "mock-model",
	}

	provider := &MockProvider{Response: mockResp}
	evaluator := NewClaudeEvaluator(provider, spec)

	ctx := context.Background()
	report, err := evaluator.Evaluate(ctx, []byte(`openapi: "3.0.0"`), &Options{
		Categories: []string{"cat-a"}, // Only evaluate cat-a
	})

	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if report.Summary.TotalRules != 2 {
		t.Errorf("TotalRules = %d, want 2 (filtered by Categories)", report.Summary.TotalRules)
	}
}

func TestClaudeEvaluator_EvaluateCategory(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "category-spec",
		Rules: []types.Rule{
			{ID: "A-001", Category: "cat-a", Judge: &types.JudgeCriteria{Prompt: "test"}},
			{ID: "A-002", Category: "cat-a", Judge: &types.JudgeCriteria{Prompt: "test"}},
		},
	}

	mockResp := &CompletionResponse{
		Content: `{"findings": [{"ruleId": "A-001", "score": 0.8, "passed": true, "reasoning": "Good"}, {"ruleId": "A-002", "score": 0.6, "passed": true, "reasoning": "OK"}]}`,
		Model:   "mock-model",
	}

	provider := &MockProvider{Response: mockResp}
	evaluator := NewClaudeEvaluator(provider, spec)

	ctx := context.Background()
	result, err := evaluator.EvaluateCategory(ctx, []byte(`openapi: "3.0.0"`), "cat-a", nil)

	if err != nil {
		t.Fatalf("EvaluateCategory() error = %v", err)
	}

	if result.Name != "cat-a" {
		t.Errorf("Name = %q, want %q", result.Name, "cat-a")
	}
	if len(result.Findings) != 2 {
		t.Errorf("len(Findings) = %d, want 2", len(result.Findings))
	}
	// Score should be average of 0.8 and 0.6 = 0.7
	if result.Score < 0.69 || result.Score > 0.71 {
		t.Errorf("Score = %f, want ~0.7", result.Score)
	}
}

func TestClaudeEvaluator_EvaluateCategory_Empty(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:  "empty-spec",
		Rules: []types.Rule{},
	}

	provider := &MockProvider{}
	evaluator := NewClaudeEvaluator(provider, spec)

	ctx := context.Background()
	result, err := evaluator.EvaluateCategory(ctx, []byte(`openapi: "3.0.0"`), "nonexistent", nil)

	if err != nil {
		t.Fatalf("EvaluateCategory() error = %v", err)
	}

	if result.Name != "nonexistent" {
		t.Errorf("Name = %q, want %q", result.Name, "nonexistent")
	}
	if len(result.Findings) != 0 {
		t.Errorf("len(Findings) = %d, want 0", len(result.Findings))
	}
}

func TestParseCategoryEvaluation(t *testing.T) {
	input := `{"findings": [
		{"ruleId": "A-001", "score": 0.8, "passed": true, "reasoning": "Good"},
		{"ruleId": "A-002", "score": 0.6, "passed": true, "reasoning": "OK"},
		{"ruleId": "UNKNOWN", "score": 0.5, "passed": true, "reasoning": "Skip this"}
	]}`

	criteria := []*Criterion{
		{RuleID: "A-001", RuleTitle: "Rule A1", Category: "test", Severity: types.SeverityWarn, Weight: 1.0},
		{RuleID: "A-002", RuleTitle: "Rule A2", Category: "test", Severity: types.SeverityInfo, Weight: 0.5},
	}

	findings, err := ParseCategoryEvaluation(input, criteria)
	if err != nil {
		t.Fatalf("ParseCategoryEvaluation() error = %v", err)
	}

	// Should only have 2 findings (UNKNOWN is skipped)
	if len(findings) != 2 {
		t.Errorf("len(findings) = %d, want 2", len(findings))
	}

	// Check A-001
	for _, f := range findings {
		if f.RuleID == "A-001" {
			if f.Score != 0.8 {
				t.Errorf("A-001 Score = %f, want 0.8", f.Score)
			}
			if f.RuleTitle != "Rule A1" {
				t.Errorf("A-001 RuleTitle = %q, want %q", f.RuleTitle, "Rule A1")
			}
		}
	}
}

func TestExtractJSON_Array(t *testing.T) {
	input := `Here's the data: [1, 2, 3]`

	// Test that array extraction works
	result, err := extractJSON(input)
	if err != nil {
		t.Fatalf("extractJSON() error = %v", err)
	}
	if result != "[1, 2, 3]" {
		t.Errorf("extractJSON() = %q, want %q", result, "[1, 2, 3]")
	}
}

func TestNormalizeScore(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{-0.5, 0},
		{0, 0},
		{0.5, 0.5},
		{1.0, 1.0},
		{1.5, 1.0},
	}

	for _, tt := range tests {
		result := normalizeScore(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeScore(%f) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestScoreToStatus(t *testing.T) {
	tests := []struct {
		score    float64
		expected types.Status
	}{
		{0.0, types.StatusFail},
		{0.49, types.StatusFail},
		{0.5, types.StatusPass},
		{0.51, types.StatusPass},
		{1.0, types.StatusPass},
	}

	for _, tt := range tests {
		result := ScoreToStatus(tt.score)
		if result != tt.expected {
			t.Errorf("ScoreToStatus(%f) = %v, want %v", tt.score, result, tt.expected)
		}
	}
}

func TestDefaultOptions_Judge(t *testing.T) {
	opts := DefaultOptions()

	if !opts.IncludeReasoning {
		t.Error("IncludeReasoning should default to true")
	}
	if opts.MaxConcurrency != 1 {
		t.Errorf("MaxConcurrency = %d, want 1", opts.MaxConcurrency)
	}
	if opts.Temperature != 0.3 {
		t.Errorf("Temperature = %f, want 0.3", opts.Temperature)
	}
	if opts.MaxTokens != 2048 {
		t.Errorf("MaxTokens = %d, want 2048", opts.MaxTokens)
	}
}

func TestClaudeEvaluator_GetModel(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:  "test-spec",
		Rules: []types.Rule{},
	}

	provider := &MockProvider{}
	evaluator := NewClaudeEvaluator(provider, spec)

	// Test with custom model
	opts := &Options{Model: "custom-model"}
	model := evaluator.getModel(opts)
	if model != "custom-model" {
		t.Errorf("getModel() = %q, want %q", model, "custom-model")
	}

	// Test with default model
	opts = &Options{}
	model = evaluator.getModel(opts)
	if model != "mock-model" {
		t.Errorf("getModel() = %q, want %q (provider default)", model, "mock-model")
	}
}

func TestRubricSet_AllCriteria(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "test-spec",
		Rules: []types.Rule{
			{ID: "A-001", Category: "cat-a", Judge: &types.JudgeCriteria{Prompt: "test"}},
			{ID: "A-002", Category: "cat-a", Judge: &types.JudgeCriteria{Prompt: "test"}},
			{ID: "B-001", Category: "cat-b", Judge: &types.JudgeCriteria{Prompt: "test"}},
		},
	}

	rs := BuildRubricSet(spec)
	all := rs.AllCriteria()

	if len(all) != 3 {
		t.Errorf("AllCriteria() returned %d, want 3", len(all))
	}
}

func TestEvaluationReport_NoFindings(t *testing.T) {
	report := NewEvaluationReport()
	report.CalculateScores()

	if report.Summary.OverallScore != 0 {
		t.Errorf("OverallScore = %f, want 0 for empty report", report.Summary.OverallScore)
	}
	if report.HasFailures() {
		t.Error("HasFailures() = true, want false for empty report")
	}
	if report.HasCriticalFailures() {
		t.Error("HasCriticalFailures() = true, want false for empty report")
	}
}

func TestClaudeEvaluator_Evaluate_ProviderError(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name: "test-spec",
		Rules: []types.Rule{
			{ID: "TEST-001", Judge: &types.JudgeCriteria{Prompt: "test"}},
		},
	}

	provider := &MockProvider{
		Err: errors.New("provider error"),
	}
	evaluator := NewClaudeEvaluator(provider, spec)

	ctx := context.Background()
	_, err := evaluator.Evaluate(ctx, []byte(`openapi: "3.0.0"`), nil)

	if err == nil {
		t.Error("Evaluate() should return error when provider fails")
	}
}
