package judge

import (
	"fmt"
	"strings"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// ToEvaluationReport converts a StyleGuideReport (score-profile output) into
// the structured-evaluation EvaluationReport format consumed by the HTML
// report generator, so score-profile JSON can feed "generate report" directly.
func (r *StyleGuideReport) ToEvaluationReport() *types.EvaluationReport {
	eval := &types.EvaluationReport{
		Metadata: &types.EvaluationMetadata{
			Document:      r.ProfileName,
			DocumentTitle: r.ProfileName,
			GeneratedBy:   fmt.Sprintf("%s (LLM-as-Judge)", r.Metadata.Model),
		},
		ReviewType:      "style-guide-quality",
		RubricID:        r.Metadata.RubricName,
		RubricVersion:   r.Metadata.RubricVersion,
		OverallDecision: r.Status,
	}

	if ts, err := time.Parse(time.RFC3339, r.Metadata.Timestamp); err == nil {
		eval.Metadata.GeneratedAt = ts
	}

	findingCounts := &types.FindingCounts{}
	for _, cat := range r.Categories {
		reasoning := cat.Reasoning
		if len(cat.Strengths) > 0 {
			reasoning = strings.TrimRight(reasoning, " ") +
				" Strengths: " + strings.Join(cat.Strengths, "; ") + "."
		}

		severity := severityForScore(cat.Score)
		catFindings := make([]types.EvaluationFinding, 0, len(cat.Weaknesses))
		for _, weakness := range cat.Weaknesses {
			finding := types.EvaluationFinding{
				Severity: severity,
				Category: cat.Name,
				Finding:  weakness,
			}
			catFindings = append(catFindings, finding)
			eval.Findings = append(eval.Findings, finding)
			countFinding(findingCounts, severity)
		}

		eval.Categories = append(eval.Categories, types.CategoryResult{
			Category:     cat.Name,
			Score:        cat.Score,
			NumericScore: cat.NumericScore,
			Weight:       cat.Weight,
			Required:     cat.Required,
			Reasoning:    reasoning,
			Findings:     catFindings,
		})
	}

	summary := fmt.Sprintf("Overall score %.1f/5.0: %d passed, %d partial, %d failed categories.",
		r.OverallScore,
		r.Summary.PassedCategories,
		r.Summary.PartialCategories,
		r.Summary.FailedCategories)

	eval.Summary = summary
	eval.Decision = &types.EvaluationDecision{
		Status:    r.Status,
		Reasoning: summary,
		CategoryCounts: &types.CategoryCounts{
			Pass:    r.Summary.PassedCategories,
			Partial: r.Summary.PartialCategories,
			Fail:    r.Summary.FailedCategories,
			Total:   r.Summary.TotalCategories,
		},
		FindingCounts: findingCounts,
	}

	return eval
}

// severityForScore maps a category score to a finding severity for that
// category's weaknesses.
func severityForScore(score string) string {
	switch score {
	case "fail":
		return "high"
	case "partial":
		return "medium"
	default:
		return "low"
	}
}

func countFinding(counts *types.FindingCounts, severity string) {
	switch severity {
	case "critical":
		counts.Critical++
	case "high":
		counts.High++
	case "medium":
		counts.Medium++
	case "low":
		counts.Low++
	}
	counts.Total++
}
