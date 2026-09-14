package analyze

import (
	"context"
	"fmt"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/judge"
	"github.com/plexusone/systemspec-apistyle/pkg/lint"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Decision represents the GO/NO-GO recommendation.
type Decision string

const (
	// DecisionGo means the API spec passes quality gates.
	DecisionGo Decision = "GO"
	// DecisionNoGo means the API spec has blocking issues.
	DecisionNoGo Decision = "NO-GO"
	// DecisionWarning means non-blocking issues were found.
	DecisionWarning Decision = "WARNING"
)

// Analyzer orchestrates lint and LLM evaluation of API specifications.
type Analyzer struct {
	spec      *types.APIStyleSpec
	linter    *lint.VacuumLinter
	evaluator judge.Evaluator
}

// New creates a new Analyzer with the given style spec and optional LLM provider.
// If provider is nil, LLM evaluation is disabled.
func New(spec *types.APIStyleSpec, provider judge.Provider) *Analyzer {
	a := &Analyzer{
		spec:   spec,
		linter: lint.NewVacuumLinter(spec),
	}

	if provider != nil {
		a.evaluator = judge.NewClaudeEvaluator(provider, spec)
	}

	return a
}

// Options configures the analysis behavior.
type Options struct {
	// FileName is the name of the spec file (for reporting).
	FileName string

	// EnableLint enables deterministic linting (default: true).
	EnableLint bool

	// EnableEvaluate enables LLM evaluation (default: false).
	EnableEvaluate bool

	// ConformanceLevel is the target level (bronze, silver, gold).
	ConformanceLevel string

	// LintOptions are passed to the linter.
	LintOptions *lint.Options

	// EvaluateOptions are passed to the evaluator.
	EvaluateOptions *judge.Options

	// FailOnWarnings makes warnings cause NO-GO decision.
	FailOnWarnings bool

	// MinScore is the minimum overall score required for GO (0.0-1.0).
	// Only applies when EnableEvaluate is true.
	MinScore float64
}

// DefaultOptions returns options with sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		EnableLint:     true,
		EnableEvaluate: false,
		MinScore:       0.7,
	}
}

// AnalysisReport contains the combined results of all analysis types.
type AnalysisReport struct {
	// Decision is the GO/NO-GO recommendation.
	Decision Decision `json:"decision"`

	// Summary provides a human-readable overview.
	Summary string `json:"summary"`

	// LintReport contains deterministic linting results.
	LintReport *types.LintReport `json:"lintReport,omitempty"`

	// EvaluationReport contains LLM evaluation results.
	EvaluationReport *judge.EvaluationReport `json:"evaluationReport,omitempty"`

	// ConformanceLevel is the achieved conformance level.
	ConformanceLevel string `json:"conformanceLevel,omitempty"`

	// OverallScore is the combined score (0.0-1.0).
	OverallScore float64 `json:"overallScore"`

	// Metadata contains analysis context.
	Metadata AnalysisMetadata `json:"metadata"`
}

// AnalysisMetadata contains context about the analysis run.
type AnalysisMetadata struct {
	// FileName is the analyzed spec file.
	FileName string `json:"fileName,omitempty"`

	// ProfileName is the style profile used.
	ProfileName string `json:"profileName,omitempty"`

	// Duration is the total analysis time.
	Duration string `json:"duration"`

	// Timestamp is when analysis was performed.
	Timestamp string `json:"timestamp"`

	// LintEnabled indicates if linting was performed.
	LintEnabled bool `json:"lintEnabled"`

	// EvaluateEnabled indicates if LLM evaluation was performed.
	EvaluateEnabled bool `json:"evaluateEnabled"`

	// ToolVersion is the systemspec-apistyle version.
	ToolVersion string `json:"toolVersion,omitempty"`
}

// Analyze performs the full analysis of an API specification.
func (a *Analyzer) Analyze(ctx context.Context, specBytes []byte, opts *Options) (*AnalysisReport, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	start := time.Now()
	report := &AnalysisReport{
		Decision: DecisionGo,
		Metadata: AnalysisMetadata{
			FileName:        opts.FileName,
			ProfileName:     a.spec.Name,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
			LintEnabled:     opts.EnableLint,
			EvaluateEnabled: opts.EnableEvaluate,
		},
	}

	// Run linting if enabled
	if opts.EnableLint {
		lintOpts := opts.LintOptions
		if lintOpts == nil {
			lintOpts = &lint.Options{
				FileName:         opts.FileName,
				ConformanceLevel: opts.ConformanceLevel,
			}
		}

		lintReport, err := a.linter.Lint(ctx, specBytes, lintOpts)
		if err != nil {
			return nil, fmt.Errorf("linting: %w", err)
		}
		report.LintReport = lintReport

		// Update decision based on lint results
		if lintReport.HasBlockingViolations() {
			report.Decision = DecisionNoGo
		} else if lintReport.Summary.Warnings > 0 {
			if opts.FailOnWarnings {
				report.Decision = DecisionNoGo
			} else if report.Decision == DecisionGo {
				report.Decision = DecisionWarning
			}
		}
	}

	// Run LLM evaluation if enabled and provider available
	if opts.EnableEvaluate && a.evaluator != nil {
		evalOpts := opts.EvaluateOptions
		if evalOpts == nil {
			evalOpts = judge.DefaultOptions()
			evalOpts.FileName = opts.FileName
		}

		evalReport, err := a.evaluator.Evaluate(ctx, specBytes, evalOpts)
		if err != nil {
			return nil, fmt.Errorf("evaluating: %w", err)
		}
		report.EvaluationReport = evalReport

		// Update decision based on evaluation results
		switch {
		case evalReport.HasCriticalFailures():
			report.Decision = DecisionNoGo
		case evalReport.Summary.OverallScore < opts.MinScore:
			report.Decision = DecisionNoGo
		case evalReport.HasFailures() && report.Decision == DecisionGo:
			report.Decision = DecisionWarning
		}

		report.OverallScore = evalReport.Summary.OverallScore
	}

	// Calculate conformance level if defined
	if opts.ConformanceLevel != "" && report.LintReport != nil {
		report.ConformanceLevel = a.calculateConformanceLevel(report.LintReport, opts.ConformanceLevel)
	}

	// Generate summary
	report.Summary = a.generateSummary(report)

	// Finalize metadata
	report.Metadata.Duration = time.Since(start).String()

	return report, nil
}

// Lint performs only deterministic linting.
func (a *Analyzer) Lint(ctx context.Context, specBytes []byte, opts *Options) (*types.LintReport, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	lintOpts := opts.LintOptions
	if lintOpts == nil {
		lintOpts = &lint.Options{
			FileName:         opts.FileName,
			ConformanceLevel: opts.ConformanceLevel,
		}
	}

	return a.linter.Lint(ctx, specBytes, lintOpts)
}

// Evaluate performs only LLM evaluation.
func (a *Analyzer) Evaluate(ctx context.Context, specBytes []byte, opts *Options) (*judge.EvaluationReport, error) {
	if a.evaluator == nil {
		return nil, fmt.Errorf("LLM evaluator not configured")
	}

	if opts == nil {
		opts = DefaultOptions()
	}

	evalOpts := opts.EvaluateOptions
	if evalOpts == nil {
		evalOpts = judge.DefaultOptions()
		evalOpts.FileName = opts.FileName
	}

	return a.evaluator.Evaluate(ctx, specBytes, evalOpts)
}

// calculateConformanceLevel determines the achieved level based on violations.
func (a *Analyzer) calculateConformanceLevel(report *types.LintReport, _ string) string {
	levels := a.spec.ConformanceLevels
	if levels == nil {
		return ""
	}

	// Check each level from highest to lowest
	levelOrder := []string{"gold", "silver", "bronze"}

	for _, level := range levelOrder {
		levelDef, ok := levels[level]
		if !ok {
			continue
		}

		// Check if this level's requirements are met
		if a.meetsLevelRequirements(report, levelDef) {
			return level
		}
	}

	return "none"
}

// meetsLevelRequirements checks if the report meets a conformance level.
func (a *Analyzer) meetsLevelRequirements(report *types.LintReport, level types.ConformanceLevel) bool {
	// Build set of violated rules
	violatedRules := make(map[string]bool)
	for _, v := range report.Violations {
		violatedRules[v.RuleID] = true
	}

	// Check required rules
	for _, ruleID := range level.RequiredRules {
		if violatedRules[ruleID] {
			return false
		}
	}

	return true
}

// generateSummary creates a human-readable summary.
func (a *Analyzer) generateSummary(report *AnalysisReport) string {
	var summary string

	switch report.Decision {
	case DecisionGo:
		summary = "API specification passes all quality gates."
	case DecisionWarning:
		summary = "API specification has non-blocking issues that should be addressed."
	case DecisionNoGo:
		summary = "API specification has blocking issues that must be resolved."
	}

	if report.LintReport != nil {
		summary += fmt.Sprintf(" Lint: %d errors, %d warnings.",
			report.LintReport.Summary.Errors,
			report.LintReport.Summary.Warnings)
	}

	if report.EvaluationReport != nil {
		summary += fmt.Sprintf(" Evaluation score: %.1f%%.",
			report.EvaluationReport.Summary.OverallScore*100)
	}

	if report.ConformanceLevel != "" {
		summary += fmt.Sprintf(" Conformance level: %s.", report.ConformanceLevel)
	}

	return summary
}

// HasEvaluator returns true if an LLM evaluator is configured.
func (a *Analyzer) HasEvaluator() bool {
	return a.evaluator != nil
}
