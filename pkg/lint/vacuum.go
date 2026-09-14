package lint

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	vacuumModel "github.com/daveshanley/vacuum/model"
	"github.com/daveshanley/vacuum/motor"
	"github.com/daveshanley/vacuum/rulesets"

	"github.com/plexusone/systemspec-apistyle/pkg/profile"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// VacuumLinter implements Linter using the vacuum library.
type VacuumLinter struct {
	spec           *types.APIStyleSpec
	validationDone bool
	skippedRules   []string
}

// NewVacuumLinter creates a new linter from an APIStyleSpec.
// It validates the profile and disables any rules with invalid JSONPath
// expressions, logging warnings for skipped rules.
func NewVacuumLinter(spec *types.APIStyleSpec) *VacuumLinter {
	linter := &VacuumLinter{spec: spec}

	// Validate and disable invalid rules
	result := profile.DisableInvalidRules(spec)
	linter.validationDone = true

	// Log warnings for skipped rules
	for _, w := range result.Warnings {
		slog.Warn("rule disabled due to invalid JSONPath",
			"rule", w.RuleID,
			"detail", w.Detail,
		)
		linter.skippedRules = append(linter.skippedRules, w.RuleID)
	}

	return linter
}

// SkippedRules returns the list of rules that were skipped due to
// invalid JSONPath expressions.
func (l *VacuumLinter) SkippedRules() []string {
	return l.skippedRules
}

// Lint analyzes an OpenAPI specification and returns violations.
func (l *VacuumLinter) Lint(_ context.Context, specBytes []byte, opts *Options) (*types.LintReport, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	start := time.Now()

	// Build ruleset from our spec
	customRules := buildVacuumRuleSet(l.spec)

	// Create a RuleSet that vacuum can use
	rs := l.buildRuleSet(customRules)

	// Configure execution
	timeout := time.Duration(opts.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	execution := &motor.RuleSetExecution{
		RuleSet:      rs,
		Spec:         specBytes,
		SpecFileName: opts.FileName,
		Timeout:      timeout,
	}

	// Run vacuum
	result := motor.ApplyRulesToRuleSet(execution)

	// Handle execution errors
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("linting failed: %w", result.Errors[0])
	}

	// Convert results
	duration := time.Since(start)
	report := convertVacuumResults(result.Results, l.spec, opts, duration)

	// Apply exceptions
	if len(opts.Exceptions) > 0 {
		report = l.applyExceptions(report, opts.Exceptions)
	}

	return report, nil
}

// buildRuleSet creates a vacuum RuleSet from our custom rules.
func (l *VacuumLinter) buildRuleSet(customRules map[string]*vacuumModel.Rule) *rulesets.RuleSet {
	return &rulesets.RuleSet{
		Description: l.spec.Name,
		Rules:       customRules,
	}
}

// applyExceptions filters violations based on exception rules.
func (l *VacuumLinter) applyExceptions(report *types.LintReport, exceptions []types.Exception) *types.LintReport {
	var filtered []types.Violation
	var ignored []types.Violation

	for _, v := range report.Violations {
		isExcepted := false
		for _, ex := range exceptions {
			if ex.Matches(v.RuleID, "", v.Path, "") {
				isExcepted = true
				break
			}
		}

		if isExcepted {
			ignored = append(ignored, v)
		} else {
			filtered = append(filtered, v)
		}
	}

	// Rebuild report with filtered violations
	newReport := types.NewLintReport()
	newReport.Metadata = report.Metadata

	for _, v := range filtered {
		newReport.AddViolation(v)
	}
	newReport.IgnoredViolations = ignored

	return newReport
}

// WithDefaults lints using vacuum's built-in recommended ruleset.
// This is useful when no custom APIStyleSpec is available.
func WithDefaults(_ context.Context, specBytes []byte, opts *Options) (*types.LintReport, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	start := time.Now()

	// Get vacuum's default recommended ruleset
	defaultRS := rulesets.BuildDefaultRuleSets()
	rs := defaultRS.GenerateOpenAPIRecommendedRuleSet()

	// Configure execution
	timeout := time.Duration(opts.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	execution := &motor.RuleSetExecution{
		RuleSet:      rs,
		Spec:         specBytes,
		SpecFileName: opts.FileName,
		Timeout:      timeout,
	}

	// Run vacuum
	result := motor.ApplyRulesToRuleSet(execution)

	// Handle execution errors
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("linting failed: %w", result.Errors[0])
	}

	// Convert results (use empty spec for enrichment)
	duration := time.Since(start)
	emptySpec := &types.APIStyleSpec{Rules: []types.Rule{}}
	report := convertVacuumResults(result.Results, emptySpec, opts, duration)

	return report, nil
}
