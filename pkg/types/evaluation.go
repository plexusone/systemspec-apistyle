package types

import "time"

// EvaluationReport contains the results of LLM-based style guide evaluation.
// This matches the structured-evaluation JSON format.
type EvaluationReport struct {
	// Schema is the JSON Schema URI for validation.
	Schema string `json:"$schema,omitempty"`

	// Metadata contains context about the evaluation.
	Metadata *EvaluationMetadata `json:"metadata"`

	// ReviewType identifies the type of review performed.
	ReviewType string `json:"reviewType"`

	// RubricID is the identifier of the evaluation rubric used.
	RubricID string `json:"rubricId"`

	// RubricVersion is the version of the rubric.
	RubricVersion string `json:"rubricVersion"`

	// Categories contains scored results for each evaluation category.
	Categories []CategoryResult `json:"categories"`

	// Findings lists all issues found during evaluation.
	Findings []EvaluationFinding `json:"findings"`

	// PassCriteria defines what constitutes a passing evaluation.
	PassCriteria *PassCriteria `json:"passCriteria,omitempty"`

	// Decision is the overall pass/fail determination.
	Decision *EvaluationDecision `json:"decision"`

	// OverallDecision is a simple pass/fail string for quick access.
	OverallDecision string `json:"overallDecision"`

	// NextSteps provides recommended actions.
	NextSteps *NextSteps `json:"nextSteps,omitempty"`

	// Summary is a brief textual summary of the evaluation.
	Summary string `json:"summary,omitempty"`
}

// EvaluationMetadata contains context about the evaluation run.
type EvaluationMetadata struct {
	// Document is the path or identifier of the evaluated document.
	Document string `json:"document"`

	// DocumentTitle is the human-readable title of the document.
	DocumentTitle string `json:"documentTitle,omitempty"`

	// GeneratedAt is when the evaluation was performed.
	GeneratedAt time.Time `json:"generatedAt"`

	// GeneratedBy identifies who/what performed the evaluation.
	GeneratedBy string `json:"generatedBy,omitempty"`

	// ToolVersion is the systemspec-apistyle version.
	ToolVersion string `json:"toolVersion,omitempty"`
}

// CategoryResult contains the evaluation result for a single category.
type CategoryResult struct {
	// Category is the name of the evaluation category.
	Category string `json:"category"`

	// Score is the categorical score (pass, partial, fail).
	Score string `json:"score"`

	// NumericScore is the numeric score (typically 1-5).
	NumericScore int `json:"numericScore"`

	// Weight is the category's weight in overall scoring (0.0-1.0).
	Weight float64 `json:"weight,omitempty"`

	// Required indicates if this category must pass.
	Required bool `json:"required,omitempty"`

	// Reasoning explains why this score was given.
	Reasoning string `json:"reasoning"`

	// Findings are issues specific to this category.
	Findings []EvaluationFinding `json:"findings,omitempty"`
}

// EvaluationFinding represents a single finding from evaluation.
type EvaluationFinding struct {
	// Severity is the importance of this finding.
	Severity string `json:"severity"` // critical, high, medium, low

	// Category is which evaluation category this finding belongs to.
	Category string `json:"category"`

	// Finding is the description of what was found.
	Finding string `json:"finding"`

	// Recommendation is how to address this finding.
	Recommendation string `json:"recommendation,omitempty"`

	// Location is where in the document this finding applies.
	Location string `json:"location,omitempty"`

	// RuleID links to a specific rule if applicable.
	RuleID string `json:"ruleId,omitempty"`
}

// PassCriteria defines what constitutes a passing evaluation.
type PassCriteria struct {
	// MinCategoriesPassing is how many categories must pass.
	// Can be a number or "all_required".
	MinCategoriesPassing string `json:"minCategoriesPassing,omitempty"`

	// MaxFindings is the maximum allowed findings by severity.
	MaxFindings *FindingLimits `json:"maxFindings,omitempty"`
}

// FindingLimits specifies maximum allowed findings per severity.
type FindingLimits struct {
	Critical int `json:"critical,omitempty"`
	High     int `json:"high,omitempty"`
	Medium   int `json:"medium,omitempty"`
	Low      int `json:"low,omitempty"`
}

// EvaluationDecision contains the final pass/fail determination.
type EvaluationDecision struct {
	// Status is the overall decision (pass, fail, partial).
	Status string `json:"status"`

	// Reasoning explains why this decision was made.
	Reasoning string `json:"reasoning"`

	// CategoryCounts summarizes category results.
	CategoryCounts *CategoryCounts `json:"categoryCounts"`

	// FindingCounts summarizes findings by severity.
	FindingCounts *FindingCounts `json:"findingCounts"`
}

// CategoryCounts tallies category results.
type CategoryCounts struct {
	Pass    int `json:"pass"`
	Partial int `json:"partial"`
	Fail    int `json:"fail"`
	Total   int `json:"total"`
}

// FindingCounts tallies findings by severity.
type FindingCounts struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Total    int `json:"total"`
}

// NextSteps provides recommended actions after evaluation.
type NextSteps struct {
	// Immediate are actions that must be taken before proceeding.
	Immediate []ActionItem `json:"immediate,omitempty"`

	// Recommended are suggested improvements.
	Recommended []ActionItem `json:"recommended,omitempty"`
}

// ActionItem is a single recommended action.
type ActionItem struct {
	// Action is what should be done.
	Action string `json:"action"`

	// Category is which evaluation category this relates to.
	Category string `json:"category,omitempty"`

	// Effort estimates the work required (low, medium, high).
	Effort string `json:"effort,omitempty"`

	// Priority indicates urgency (1 = most urgent).
	Priority int `json:"priority,omitempty"`
}

// NewEvaluationReport creates a new EvaluationReport with initialized fields.
func NewEvaluationReport() *EvaluationReport {
	return &EvaluationReport{
		Metadata: &EvaluationMetadata{
			GeneratedAt: time.Now(),
		},
		Categories: []CategoryResult{},
		Findings:   []EvaluationFinding{},
		Decision: &EvaluationDecision{
			CategoryCounts: &CategoryCounts{},
			FindingCounts:  &FindingCounts{},
		},
	}
}

// AddCategory adds a category result and updates counts.
func (r *EvaluationReport) AddCategory(cat CategoryResult) {
	r.Categories = append(r.Categories, cat)

	if r.Decision.CategoryCounts == nil {
		r.Decision.CategoryCounts = &CategoryCounts{}
	}

	switch cat.Score {
	case "pass":
		r.Decision.CategoryCounts.Pass++
	case "partial":
		r.Decision.CategoryCounts.Partial++
	case "fail":
		r.Decision.CategoryCounts.Fail++
	}
	r.Decision.CategoryCounts.Total++
}

// AddFinding adds a finding and updates counts.
func (r *EvaluationReport) AddFinding(f EvaluationFinding) {
	r.Findings = append(r.Findings, f)

	if r.Decision.FindingCounts == nil {
		r.Decision.FindingCounts = &FindingCounts{}
	}

	switch f.Severity {
	case "critical":
		r.Decision.FindingCounts.Critical++
	case "high":
		r.Decision.FindingCounts.High++
	case "medium":
		r.Decision.FindingCounts.Medium++
	case "low":
		r.Decision.FindingCounts.Low++
	}
	r.Decision.FindingCounts.Total++
}

// IsPassing returns true if the evaluation passed.
func (r *EvaluationReport) IsPassing() bool {
	return r.OverallDecision == "pass" || r.Decision.Status == "pass"
}

// ScoreEmoji returns an emoji representing the numeric score.
func ScoreEmoji(score int) string {
	switch {
	case score >= 5:
		return "🟢"
	case score >= 4:
		return "🟡"
	case score >= 3:
		return "🟠"
	default:
		return "🔴"
	}
}

// SeverityEmoji returns an emoji representing finding severity.
func SeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}
