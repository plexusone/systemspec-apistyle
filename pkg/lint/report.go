package lint

import (
	"fmt"
	"strings"
	"time"

	vacuumModel "github.com/daveshanley/vacuum/model"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// convertVacuumResults converts vacuum results to our LintReport format.
func convertVacuumResults(
	results []vacuumModel.RuleFunctionResult,
	spec *types.APIStyleSpec,
	opts *Options,
	duration time.Duration,
) *types.LintReport {
	report := types.NewLintReport()

	// Build rule lookup for enrichment
	ruleLookup := make(map[string]*types.Rule)
	for i := range spec.Rules {
		ruleLookup[spec.Rules[i].ID] = &spec.Rules[i]
	}

	// Convert each result to a violation
	for i := range results {
		violation := convertVacuumResult(&results[i], ruleLookup)
		report.AddViolation(violation)
	}

	// Set metadata
	report.Metadata = &types.ReportMetadata{
		SpecFile:       opts.FileName,
		Profile:        opts.Profile,
		Duration:       duration,
		DurationMS:     duration.Milliseconds(),
		Timestamp:      time.Now(),
		RulesEvaluated: len(spec.Rules),
	}

	return report
}

// convertVacuumResult converts a single vacuum result to a Violation.
func convertVacuumResult(
	result *vacuumModel.RuleFunctionResult,
	ruleLookup map[string]*types.Rule,
) types.Violation {
	violation := types.Violation{
		RuleID:   result.RuleId,
		Message:  result.Message,
		Path:     result.Path,
		Severity: convertSeverity(result.RuleSeverity),
	}

	// Set location if available
	if result.StartNode != nil {
		violation.Line = result.StartNode.Line
		violation.Column = result.StartNode.Column
	}
	if result.EndNode != nil {
		violation.EndLine = result.EndNode.Line
		violation.EndColumn = result.EndNode.Column
	}

	// Enrich with rule information
	if rule, ok := ruleLookup[result.RuleId]; ok {
		enrichViolation(&violation, rule)
	}

	// Deterministic violations have full confidence
	violation.Confidence = 1.0

	return violation
}

// enrichViolation adds remediation metadata to a violation from rule information.
func enrichViolation(v *types.Violation, rule *types.Rule) {
	v.RuleTitle = rule.Title
	v.Category = rule.Category

	// Generate rule documentation URL
	v.RuleURL = fmt.Sprintf("https://plexusone.dev/systemspec-apistyle/rules/%s",
		strings.ToLower(rule.ID))

	// Add suggestion from migration guidance or examples
	if rule.Migration != nil && len(rule.Migration.Steps) > 0 {
		v.Suggestion = rule.Migration.Steps[0].Description
	} else if rule.Examples != nil && len(rule.Examples.Good) > 0 {
		v.Suggestion = "Consider: " + rule.Examples.Good[0]
	}

	// Add example fix from good examples
	if rule.Examples != nil && len(rule.Examples.Good) > 0 {
		v.ExampleFix = generateExampleFix(rule)
	}

	// Calculate fix priority from rule priority or generation priority
	// Higher priority value = fix first, lower value = fix later
	priority := rule.Priority
	if priority == 0 && rule.Generate != nil && rule.Generate.Priority > 0 {
		priority = rule.Generate.Priority
	}
	if priority > 0 {
		// Invert: higher priority value = lower fix priority number (fix first)
		v.FixPriority = 101 - priority
		if v.FixPriority < 1 {
			v.FixPriority = 1
		}
	} else {
		v.FixPriority = 50 // Default middle priority
	}

	// Extract related rules from relations
	for _, rel := range rule.Relations {
		if rel.Type == "requires" || rel.Type == "conflicts" || rel.Type == "related" {
			v.RelatedRules = append(v.RelatedRules, rel.RuleID)
		}
	}
}

// generateExampleFix creates a code snippet showing correct usage.
func generateExampleFix(rule *types.Rule) string {
	if rule.Examples == nil || len(rule.Examples.Good) == 0 {
		return ""
	}

	// If we have detailed examples with OpenAPI snippets, use those
	if len(rule.Examples.Detailed) > 0 {
		for _, ex := range rule.Examples.Detailed {
			if ex.Type == "good" && ex.Code != "" {
				return ex.Code
			}
		}
	}

	// Otherwise format the simple good examples
	var sb strings.Builder
	sb.WriteString("# Correct usage:\n")
	for _, good := range rule.Examples.Good {
		sb.WriteString("# - ")
		sb.WriteString(good)
		sb.WriteString("\n")
	}
	return sb.String()
}

// convertSeverity maps vacuum severity strings to our Severity type.
func convertSeverity(vacuumSeverity string) types.Severity {
	switch vacuumSeverity {
	case "error":
		return types.SeverityError
	case "warn", "warning":
		return types.SeverityWarn
	case "info", "information":
		return types.SeverityInfo
	case "hint":
		return types.SeverityHint
	default:
		return types.SeverityWarn
	}
}
