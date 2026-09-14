package report

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func TestGenerator_Generate(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	eval := &types.EvaluationReport{
		Metadata: &types.EvaluationMetadata{
			Document:      "test-profile.json",
			DocumentTitle: "Test API Guidelines",
			GeneratedAt:   time.Date(2025, 6, 17, 0, 0, 0, 0, time.UTC),
			GeneratedBy:   "Test Evaluator",
		},
		ReviewType:    "style-guide-quality",
		RubricID:      "api-style-guide-quality",
		RubricVersion: "1.0.0",
		Categories: []types.CategoryResult{
			{
				Category:     "Content Coverage",
				Score:        "pass",
				NumericScore: 5,
				Reasoning:    "Covers all essential domains.",
			},
			{
				Category:     "Structure & Navigation",
				Score:        "pass",
				NumericScore: 4,
				Reasoning:    "Good structure with room for improvement.",
			},
		},
		Findings: []types.EvaluationFinding{
			{
				Severity:       "low",
				Category:       "Examples",
				Finding:        "Some rules lack examples",
				Recommendation: "Add examples to all rules",
			},
		},
		Decision: &types.EvaluationDecision{
			Status:    "pass",
			Reasoning: "All categories pass.",
			CategoryCounts: &types.CategoryCounts{
				Pass:  2,
				Total: 2,
			},
			FindingCounts: &types.FindingCounts{
				Low:   1,
				Total: 1,
			},
		},
		OverallDecision: "pass",
		Summary:         "Test guidelines: excellent coverage.",
	}

	html, err := gen.Generate(context.Background(), eval, nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	htmlStr := string(html)

	// Check essential elements
	checks := []string{
		"<!DOCTYPE html>",
		"Test API Guidelines",
		"PASS",
		"Content Coverage",
		"5/5",
		"Structure &amp; Navigation",
		"4/5",
		"Some rules lack examples",
		"Add examples to all rules",
		"api-style-guide-quality",
	}

	for _, check := range checks {
		if !strings.Contains(htmlStr, check) {
			t.Errorf("HTML missing expected content: %q", check)
		}
	}
}

func TestGenerator_GenerateDarkTheme(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	eval := &types.EvaluationReport{
		OverallDecision: "pass",
		Decision: &types.EvaluationDecision{
			CategoryCounts: &types.CategoryCounts{},
			FindingCounts:  &types.FindingCounts{},
		},
	}

	opts := &Options{
		Theme: "dark",
		Title: "Dark Theme Test",
	}

	html, err := gen.Generate(context.Background(), eval, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	htmlStr := string(html)

	// Check dark theme CSS is present
	if !strings.Contains(htmlStr, "--bg-primary: #0f172a") {
		t.Error("Dark theme CSS not applied")
	}
}

func TestGenerator_GenerateWithRawJSON(t *testing.T) {
	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	eval := &types.EvaluationReport{
		OverallDecision: "pass",
		RubricID:        "test-rubric",
		Decision: &types.EvaluationDecision{
			CategoryCounts: &types.CategoryCounts{},
			FindingCounts:  &types.FindingCounts{},
		},
	}

	opts := &Options{
		IncludeRawJSON: true,
	}

	html, err := gen.Generate(context.Background(), eval, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	htmlStr := string(html)

	// Check raw JSON section is present
	if !strings.Contains(htmlStr, "Raw Data") {
		t.Error("Raw JSON section not present")
	}
	// Check that rubricId appears in the JSON (exact formatting may vary)
	if !strings.Contains(htmlStr, "test-rubric") {
		t.Error("Raw JSON content not present")
	}
}

func TestScoreEmoji(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{5, "🟢"},
		{4, "🟡"},
		{3, "🟠"},
		{2, "🔴"},
		{1, "🔴"},
	}

	for _, tt := range tests {
		got := types.ScoreEmoji(tt.score)
		if got != tt.expected {
			t.Errorf("ScoreEmoji(%d) = %s, want %s", tt.score, got, tt.expected)
		}
	}
}

func TestSeverityEmoji(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "🔴"},
		{"high", "🟠"},
		{"medium", "🟡"},
		{"low", "🟢"},
		{"unknown", "⚪"},
	}

	for _, tt := range tests {
		got := types.SeverityEmoji(tt.severity)
		if got != tt.expected {
			t.Errorf("SeverityEmoji(%q) = %s, want %s", tt.severity, got, tt.expected)
		}
	}
}

func TestParseEvaluationJSON_StyleGuideReport(t *testing.T) {
	data := []byte(`{
		"profileName": "test-profile",
		"status": "pass",
		"overallScore": 4.5,
		"categories": [
			{"category": "content", "name": "Content Coverage", "score": "pass",
			 "numericScore": 5, "reasoning": "Good.", "weaknesses": ["Minor gap"]}
		],
		"summary": {"totalCategories": 1, "passedCategories": 1},
		"metadata": {"rubricName": "style-guide-quality", "rubricVersion": "1.0.0",
			"model": "claude-sonnet-5", "timestamp": "2026-08-30T12:00:00Z"}
	}`)

	eval, err := ParseEvaluationJSON(data)
	if err != nil {
		t.Fatalf("ParseEvaluationJSON() error = %v", err)
	}
	if eval.ReviewType != "style-guide-quality" {
		t.Errorf("ReviewType = %q, want style-guide-quality (converted)", eval.ReviewType)
	}
	if len(eval.Findings) != 1 {
		t.Errorf("len(Findings) = %d, want 1", len(eval.Findings))
	}

	gen, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	html, err := gen.Generate(context.Background(), eval, nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	for _, want := range []string{"test-profile", "Content Coverage", "Minor gap"} {
		if !strings.Contains(string(html), want) {
			t.Errorf("HTML missing %q", want)
		}
	}
}

func TestParseEvaluationJSON_EvaluationReport(t *testing.T) {
	data := []byte(`{"reviewType": "style-guide-quality", "rubricId": "r1",
		"categories": [{"category": "C1", "score": "pass", "numericScore": 5, "reasoning": "ok"}],
		"overallDecision": "pass"}`)

	eval, err := ParseEvaluationJSON(data)
	if err != nil {
		t.Fatalf("ParseEvaluationJSON() error = %v", err)
	}
	if eval.RubricID != "r1" || len(eval.Categories) != 1 {
		t.Errorf("native EvaluationReport not parsed: %+v", eval)
	}
}
