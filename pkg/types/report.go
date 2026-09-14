package types

import "time"

// Status indicates the overall result of linting or evaluation.
type Status string

const (
	// StatusPass means no blocking violations were found.
	StatusPass Status = "pass"
	// StatusFail means one or more blocking violations were found.
	StatusFail Status = "fail"
)

// ReportStatus is an alias for Status (deprecated, use Status).
type ReportStatus = Status

// Deprecated aliases for backwards compatibility.
const (
	ReportStatusPass = StatusPass
	ReportStatusFail = StatusFail
)

// LintReport contains the results of deterministic linting.
type LintReport struct {
	// Status is the overall pass/fail result.
	Status ReportStatus `json:"status"`

	// ConformanceLevel is the highest level achieved (if levels are defined).
	ConformanceLevel string `json:"conformanceLevel,omitempty"`

	// Summary provides violation counts by severity.
	Summary *ViolationSummary `json:"summary"`

	// Violations lists all findings.
	Violations []Violation `json:"violations"`

	// IgnoredViolations lists violations that were suppressed by exceptions.
	IgnoredViolations []Violation `json:"ignoredViolations,omitempty"`

	// Metadata includes timing, versions, and other context.
	Metadata *ReportMetadata `json:"metadata,omitempty"`
}

// ViolationSummary counts violations by severity.
type ViolationSummary struct {
	// Errors is the count of error-severity violations.
	Errors int `json:"errors"`
	// Warnings is the count of warning-severity violations.
	Warnings int `json:"warnings"`
	// Infos is the count of info-severity violations.
	Infos int `json:"infos"`
	// Hints is the count of hint-severity violations.
	Hints int `json:"hints"`
	// Total is the sum of all violations.
	Total int `json:"total"`
}

// Violation represents a single rule violation.
type Violation struct {
	// RuleID is the identifier of the violated rule.
	RuleID string `json:"ruleId"`

	// Severity indicates the importance of this violation.
	Severity Severity `json:"severity"`

	// Message describes what went wrong.
	Message string `json:"message"`

	// Path is the JSONPath to the violation location.
	Path string `json:"path"`

	// Line is the line number in the source file (1-indexed).
	Line int `json:"line,omitempty"`

	// Column is the column number in the source file (1-indexed).
	Column int `json:"column,omitempty"`

	// EndLine is the ending line for multi-line issues.
	EndLine int `json:"endLine,omitempty"`

	// EndColumn is the ending column for multi-line issues.
	EndColumn int `json:"endColumn,omitempty"`

	// Suggestion provides guidance for fixing the violation.
	Suggestion string `json:"suggestion,omitempty"`

	// RuleTitle is the human-readable rule name.
	RuleTitle string `json:"ruleTitle,omitempty"`

	// Category is the rule's category.
	Category string `json:"category,omitempty"`

	// ExampleFix shows a code snippet demonstrating the fix.
	// Format matches the input spec format (YAML or JSON).
	ExampleFix string `json:"exampleFix,omitempty"`

	// RuleURL links to the rule's documentation page.
	RuleURL string `json:"ruleUrl,omitempty"`

	// Confidence indicates certainty of this violation (0.0-1.0).
	// 1.0 = deterministic match, <1.0 = heuristic or LLM-based.
	Confidence float64 `json:"confidence,omitempty"`

	// RelatedRules lists rule IDs that should be addressed first
	// or are commonly fixed together with this violation.
	RelatedRules []string `json:"relatedRules,omitempty"`

	// FixPriority indicates the recommended fix order (1 = fix first).
	// Derived from rule priority and dependencies.
	FixPriority int `json:"fixPriority,omitempty"`
}

// ReportMetadata contains context about the linting run.
type ReportMetadata struct {
	// SpecFile is the path to the linted specification.
	SpecFile string `json:"specFile,omitempty"`

	// SpecVersion is the OpenAPI version of the spec.
	SpecVersion string `json:"specVersion,omitempty"`

	// Profile is the style profile used.
	Profile string `json:"profile,omitempty"`

	// ProfileVersion is the version of the profile.
	ProfileVersion string `json:"profileVersion,omitempty"`

	// Duration is how long linting took.
	Duration time.Duration `json:"duration,omitempty"`

	// DurationMS is duration in milliseconds (for JSON serialization).
	DurationMS int64 `json:"durationMs,omitempty"`

	// Timestamp is when linting was performed.
	Timestamp time.Time `json:"timestamp"`

	// ToolVersion is the systemspec-apistyle version.
	ToolVersion string `json:"toolVersion,omitempty"`

	// RulesEvaluated is the count of rules that were checked.
	RulesEvaluated int `json:"rulesEvaluated,omitempty"`
}

// NewLintReport creates a new LintReport with initialized fields.
func NewLintReport() *LintReport {
	return &LintReport{
		Status:     ReportStatusPass,
		Summary:    &ViolationSummary{},
		Violations: []Violation{},
		Metadata: &ReportMetadata{
			Timestamp: time.Now(),
		},
	}
}

// AddViolation adds a violation and updates the summary.
func (r *LintReport) AddViolation(v Violation) {
	r.Violations = append(r.Violations, v)

	switch v.Severity {
	case SeverityError:
		r.Summary.Errors++
		r.Status = ReportStatusFail
	case SeverityWarn:
		r.Summary.Warnings++
	case SeverityInfo:
		r.Summary.Infos++
	case SeverityHint:
		r.Summary.Hints++
	}
	r.Summary.Total++
}

// HasBlockingViolations returns true if there are error-level violations.
func (r *LintReport) HasBlockingViolations() bool {
	return r.Summary.Errors > 0
}

// MultiLintReport contains results from linting multiple files.
type MultiLintReport struct {
	// Status is the overall pass/fail result across all files.
	Status Status `json:"status"`

	// Summary provides aggregate violation counts across all files.
	Summary *ViolationSummary `json:"summary"`

	// FileReports contains individual reports for each file.
	FileReports []FileLintReport `json:"fileReports"`

	// Metadata includes timing, versions, and other context.
	Metadata *ReportMetadata `json:"metadata,omitempty"`
}

// FileLintReport wraps a LintReport with file path information.
type FileLintReport struct {
	// File is the path to the linted specification.
	File string `json:"file"`

	// Report contains the lint results for this file.
	*LintReport
}

// NewMultiLintReport creates a new MultiLintReport with initialized fields.
func NewMultiLintReport() *MultiLintReport {
	return &MultiLintReport{
		Status:      StatusPass,
		Summary:     &ViolationSummary{},
		FileReports: []FileLintReport{},
		Metadata: &ReportMetadata{
			Timestamp: time.Now(),
		},
	}
}

// AddFileReport adds a file report and updates the aggregate summary.
func (r *MultiLintReport) AddFileReport(file string, report *LintReport) {
	r.FileReports = append(r.FileReports, FileLintReport{
		File:       file,
		LintReport: report,
	})

	// Aggregate summary
	if report.Summary != nil {
		r.Summary.Errors += report.Summary.Errors
		r.Summary.Warnings += report.Summary.Warnings
		r.Summary.Infos += report.Summary.Infos
		r.Summary.Hints += report.Summary.Hints
		r.Summary.Total += report.Summary.Total
	}

	// Update status if any file failed
	if report.Status == StatusFail {
		r.Status = StatusFail
	}
}

// HasBlockingViolations returns true if any file has error-level violations.
func (r *MultiLintReport) HasBlockingViolations() bool {
	return r.Summary.Errors > 0
}

// FileCount returns the number of files that were linted.
func (r *MultiLintReport) FileCount() int {
	return len(r.FileReports)
}

// FailedFileCount returns the number of files that failed linting.
func (r *MultiLintReport) FailedFileCount() int {
	count := 0
	for _, fr := range r.FileReports {
		if fr.Status == StatusFail {
			count++
		}
	}
	return count
}
