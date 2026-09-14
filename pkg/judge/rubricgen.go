package judge

import (
	"github.com/plexusone/structured-evaluation/rubric"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// GenerateRubricSet converts an APIStyleSpec to a structured-evaluation RubricSet.
// This enables integration with the structured-evaluation framework for LLM-as-Judge.
func GenerateRubricSet(spec *types.APIStyleSpec) *rubric.RubricSet {
	rs := rubric.NewRubricSet(spec.Name, spec.Name+" Evaluation Rubric", spec.Version)

	if spec.Description != "" {
		rs.Description = spec.Description
	}

	// Group rules by category
	categoryRules := make(map[string][]types.Rule)
	for _, rule := range spec.Rules {
		categoryRules[rule.Category] = append(categoryRules[rule.Category], rule)
	}

	// Build categories from spec categories
	for _, cat := range spec.Categories {
		rules := categoryRules[cat.ID]
		if len(rules) == 0 {
			continue
		}

		category := rubric.NewCategory(cat.ID, cat.Title, cat.Description)

		// Determine if category is required (has any error-severity rules)
		category.SetRequired(hasErrorSeverityRule(rules))

		// Build pass/partial/fail criteria from rules
		passCriteria := aggregatePassCriteria(rules)
		partialCriteria := aggregatePartialCriteria(rules)
		failCriteria := aggregateFailCriteria(rules)

		if len(passCriteria) > 0 || len(partialCriteria) > 0 || len(failCriteria) > 0 {
			category.WithPassPartialFail(passCriteria, partialCriteria, failCriteria)
		}

		// Build evaluation prompt from rules
		if prompt := buildCategoryPrompt(rules); prompt != "" {
			category.SetEvaluationPrompt(prompt)
		}

		// Add few-shot examples if available
		if examples := aggregateExamples(rules); examples != nil {
			category.SetExamples(examples)
		}

		rs.AddCategory(*category)
	}

	// Set pass criteria based on spec
	rs.PassCriteria = rubric.RubricPassCriteria{
		MinCategoriesPassing: "all_required",
		MaxFindings: &rubric.FindingLimits{
			Critical: 0,
			High:     0,
			Medium:   -1, // Unlimited by default
		},
	}

	// Add metadata
	if spec.Metadata != nil {
		rs.Metadata = &rubric.RubricMetadata{
			Author: spec.Metadata.Author,
		}
	}

	return rs
}

// hasErrorSeverityRule returns true if any rule has error severity.
func hasErrorSeverityRule(rules []types.Rule) bool {
	for _, r := range rules {
		if r.Severity == types.SeverityError {
			return true
		}
	}
	return false
}

// aggregatePassCriteria builds pass criteria from rules with pass criteria.
func aggregatePassCriteria(rules []types.Rule) []string {
	var criteria []string
	for _, r := range rules {
		if r.Judge != nil && len(r.Judge.PassCriteria) > 0 {
			criteria = append(criteria, r.Judge.PassCriteria...)
		}
	}
	return criteria
}

// aggregatePartialCriteria builds partial criteria from rules.
func aggregatePartialCriteria(rules []types.Rule) []string {
	var criteria []string
	for _, r := range rules {
		if r.Judge != nil && len(r.Judge.PartialCriteria) > 0 {
			criteria = append(criteria, r.Judge.PartialCriteria...)
		}
	}
	return criteria
}

// aggregateFailCriteria builds fail criteria from rules.
func aggregateFailCriteria(rules []types.Rule) []string {
	var criteria []string
	for _, r := range rules {
		if r.Judge != nil && len(r.Judge.FailCriteria) > 0 {
			criteria = append(criteria, r.Judge.FailCriteria...)
		}
	}
	return criteria
}

// buildCategoryPrompt builds a combined evaluation prompt from rules.
func buildCategoryPrompt(rules []types.Rule) string {
	var prompts []string
	for _, r := range rules {
		if r.Judge != nil && r.Judge.Prompt != "" {
			prompts = append(prompts, r.Judge.Prompt)
		}
	}
	if len(prompts) == 0 {
		return ""
	}
	// Return first prompt for now; could be enhanced to merge
	return prompts[0]
}

// aggregateExamples builds CategoryExamples from rules with judge examples.
func aggregateExamples(rules []types.Rule) *rubric.CategoryExamples {
	var result rubric.CategoryExamples
	hasExamples := false

	for _, r := range rules {
		if r.Judge == nil || r.Judge.Examples == nil {
			continue
		}

		ex := r.Judge.Examples
		if ex.Pass != nil && result.Pass == nil {
			result.Pass = &rubric.Example{
				Excerpt:   ex.Pass.Excerpt,
				Reasoning: ex.Pass.Reasoning,
			}
			hasExamples = true
		}
		if ex.Partial != nil && result.Partial == nil {
			result.Partial = &rubric.Example{
				Excerpt:   ex.Partial.Excerpt,
				Reasoning: ex.Partial.Reasoning,
			}
			hasExamples = true
		}
		if ex.Fail != nil && result.Fail == nil {
			result.Fail = &rubric.Example{
				Excerpt:   ex.Fail.Excerpt,
				Reasoning: ex.Fail.Reasoning,
			}
			hasExamples = true
		}

		// Stop once we have all three
		if result.Pass != nil && result.Partial != nil && result.Fail != nil {
			break
		}
	}

	if !hasExamples {
		return nil
	}
	return &result
}

// RubricSetToJSON serializes a RubricSet to JSON bytes.
func RubricSetToJSON(rs *rubric.RubricSet) ([]byte, error) {
	return rs.ToJSON()
}
