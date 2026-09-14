package judge

import (
	"context"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Evaluator runs LLM-based evaluation of API specifications.
type Evaluator interface {
	// Evaluate assesses an API specification against the configured rubric.
	Evaluate(ctx context.Context, specBytes []byte, opts *Options) (*EvaluationReport, error)

	// EvaluateCategory evaluates a single category of rules.
	EvaluateCategory(ctx context.Context, specBytes []byte, category string, opts *Options) (*CategoryResult, error)
}

// Options configures the evaluation behavior.
type Options struct {
	// Categories limits evaluation to specific categories.
	// If empty, all categories are evaluated.
	Categories []string

	// RuleIDs limits evaluation to specific rules.
	// If empty, all rules with Judge criteria are evaluated.
	RuleIDs []string

	// FileName is the name of the spec file (for context in prompts).
	FileName string

	// IncludeReasoning enables detailed reasoning in results.
	IncludeReasoning bool

	// MaxConcurrency limits parallel LLM calls (default: 1).
	MaxConcurrency int

	// Model specifies the LLM model to use.
	// Provider-specific (e.g., "claude-3-haiku-20240307" for Anthropic).
	Model string

	// Temperature controls response randomness (0.0-1.0).
	Temperature float64

	// MaxTokens limits response length.
	MaxTokens int
}

// DefaultOptions returns options with sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		IncludeReasoning: true,
		MaxConcurrency:   1,
		Temperature:      0.3,
		MaxTokens:        2048,
	}
}

// EvaluationReport contains the full results of an LLM evaluation.
type EvaluationReport struct {
	// Status indicates overall evaluation outcome.
	Status types.Status `json:"status"`

	// Summary provides aggregate statistics.
	Summary EvaluationSummary `json:"summary"`

	// Categories contains results grouped by category.
	Categories []CategoryResult `json:"categories"`

	// Findings contains all individual rule evaluations.
	Findings []Finding `json:"findings"`

	// Metadata contains evaluation context.
	Metadata ReportMetadata `json:"metadata"`
}

// EvaluationSummary provides aggregate statistics.
type EvaluationSummary struct {
	// TotalRules is the number of rules evaluated.
	TotalRules int `json:"totalRules"`

	// PassedRules is the number of rules that passed.
	PassedRules int `json:"passedRules"`

	// FailedRules is the number of rules that failed.
	FailedRules int `json:"failedRules"`

	// SkippedRules is the number of rules skipped.
	SkippedRules int `json:"skippedRules"`

	// OverallScore is the weighted average score (0.0-1.0).
	OverallScore float64 `json:"overallScore"`

	// CategoryScores maps category names to their scores.
	CategoryScores map[string]float64 `json:"categoryScores,omitempty"`
}

// CategoryResult contains evaluation results for a single category.
type CategoryResult struct {
	// Name is the category identifier.
	Name string `json:"name"`

	// Score is the category's aggregate score (0.0-1.0).
	Score float64 `json:"score"`

	// Findings contains rule evaluations in this category.
	Findings []Finding `json:"findings"`
}

// Finding represents an individual rule evaluation result.
type Finding struct {
	// RuleID is the evaluated rule's identifier.
	RuleID string `json:"ruleId"`

	// RuleTitle is the rule's display name.
	RuleTitle string `json:"ruleTitle"`

	// Category is the rule's category.
	Category string `json:"category"`

	// Score is the evaluation score (0.0-1.0).
	Score float64 `json:"score"`

	// Passed indicates if the rule passed (score >= 0.5).
	Passed bool `json:"passed"`

	// Reasoning explains the evaluation decision.
	Reasoning string `json:"reasoning,omitempty"`

	// Examples are specific instances found in the spec.
	Examples []string `json:"examples,omitempty"`

	// Suggestions provides improvement recommendations.
	Suggestions []string `json:"suggestions,omitempty"`

	// Locations identifies paths in the spec related to findings.
	Locations []string `json:"locations,omitempty"`

	// Severity reflects the rule's configured severity.
	Severity types.Severity `json:"severity"`

	// Weight is the rule's evaluation weight.
	Weight float64 `json:"weight"`
}

// ReportMetadata contains evaluation context information.
type ReportMetadata struct {
	// FileName is the evaluated spec file name.
	FileName string `json:"fileName,omitempty"`

	// ProfileName is the style profile used.
	ProfileName string `json:"profileName,omitempty"`

	// Model is the LLM model used.
	Model string `json:"model,omitempty"`

	// Duration is the total evaluation time.
	Duration string `json:"duration,omitempty"`

	// Timestamp is when the evaluation was performed.
	Timestamp string `json:"timestamp,omitempty"`

	// SpecTruncated indicates the spec exceeded the size limit and was
	// evaluated partially.
	SpecTruncated bool `json:"specTruncated,omitempty"`
}

// NewEvaluationReport creates a new empty evaluation report.
func NewEvaluationReport() *EvaluationReport {
	return &EvaluationReport{
		Status:     types.StatusPass,
		Categories: []CategoryResult{},
		Findings:   []Finding{},
		Summary: EvaluationSummary{
			CategoryScores: make(map[string]float64),
		},
	}
}

// AddFinding adds a finding and updates summary statistics.
func (r *EvaluationReport) AddFinding(f Finding) {
	r.Findings = append(r.Findings, f)
	r.Summary.TotalRules++

	if f.Passed {
		r.Summary.PassedRules++
	} else {
		r.Summary.FailedRules++
		// Fail the report if any error-severity rule fails
		if f.Severity == types.SeverityError {
			r.Status = types.StatusFail
		}
	}
}

// CalculateScores computes final scores after all findings are added.
func (r *EvaluationReport) CalculateScores() {
	if len(r.Findings) == 0 {
		return
	}

	// Calculate overall weighted score
	var totalWeight float64
	var weightedSum float64

	categoryWeights := make(map[string]float64)
	categorySums := make(map[string]float64)

	for _, f := range r.Findings {
		weight := f.Weight
		if weight == 0 {
			weight = 1.0
		}

		totalWeight += weight
		weightedSum += f.Score * weight

		categoryWeights[f.Category] += weight
		categorySums[f.Category] += f.Score * weight
	}

	if totalWeight > 0 {
		r.Summary.OverallScore = weightedSum / totalWeight
	}

	// Calculate category scores
	for category, weight := range categoryWeights {
		if weight > 0 {
			r.Summary.CategoryScores[category] = categorySums[category] / weight
		}
	}

	// Group findings by category
	categoryMap := make(map[string]*CategoryResult)
	for _, f := range r.Findings {
		cr, ok := categoryMap[f.Category]
		if !ok {
			cr = &CategoryResult{Name: f.Category}
			categoryMap[f.Category] = cr
		}
		cr.Findings = append(cr.Findings, f)
	}

	// Set category scores and build result slice
	r.Categories = make([]CategoryResult, 0, len(categoryMap))
	for name, cr := range categoryMap {
		cr.Score = r.Summary.CategoryScores[name]
		r.Categories = append(r.Categories, *cr)
	}
}

// HasFailures returns true if any findings failed.
func (r *EvaluationReport) HasFailures() bool {
	return r.Summary.FailedRules > 0
}

// HasCriticalFailures returns true if any error-severity rules failed.
func (r *EvaluationReport) HasCriticalFailures() bool {
	for _, f := range r.Findings {
		if !f.Passed && f.Severity == types.SeverityError {
			return true
		}
	}
	return false
}
