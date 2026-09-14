package judge

import (
	"fmt"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// RubricSet is a collection of evaluation criteria built from an APIStyleSpec.
type RubricSet struct {
	// Name is the rubric set identifier (from spec name).
	Name string

	// Criteria contains all evaluation criteria keyed by rule ID.
	Criteria map[string]*Criterion

	// Categories groups criteria by category name.
	Categories map[string][]*Criterion
}

// Criterion defines how a single rule should be evaluated by an LLM.
type Criterion struct {
	// RuleID links back to the source rule.
	RuleID string

	// RuleTitle is the rule's display name.
	RuleTitle string

	// Category is the rule's category for grouping.
	Category string

	// Prompt is the evaluation instruction for the LLM.
	Prompt string

	// Weight influences scoring (0.0-1.0, default 1.0).
	Weight float64

	// Severity is the rule's severity level.
	Severity types.Severity

	// RequiresContext indicates if broader context is needed.
	RequiresContext bool

	// Examples from the rule (if available).
	GoodExamples []string
	BadExamples  []string

	// Rationale from the rule (if available).
	Rationale string

	// References from the rule.
	References []types.Reference
}

// BuildRubricSet creates a RubricSet from an APIStyleSpec.
// Only rules with JudgeCriteria are included.
func BuildRubricSet(spec *types.APIStyleSpec) *RubricSet {
	rs := &RubricSet{
		Name:       spec.Name,
		Criteria:   make(map[string]*Criterion),
		Categories: make(map[string][]*Criterion),
	}

	for i := range spec.Rules {
		rule := &spec.Rules[i]

		// Skip rules without judge criteria
		if rule.Judge == nil || rule.Judge.Prompt == "" {
			continue
		}

		criterion := &Criterion{
			RuleID:          rule.ID,
			RuleTitle:       rule.Title,
			Category:        rule.Category,
			Prompt:          rule.Judge.Prompt,
			Weight:          rule.Judge.Weight,
			Severity:        rule.Severity,
			RequiresContext: rule.Judge.RequiresContext,
			Rationale:       rule.Rationale,
			References:      rule.References,
		}

		// Default weight to 1.0
		if criterion.Weight == 0 {
			criterion.Weight = 1.0
		}

		// Use judge category override if specified
		if rule.Judge.Category != "" {
			criterion.Category = rule.Judge.Category
		}

		// Include examples if available
		if rule.Examples != nil {
			criterion.GoodExamples = rule.Examples.Good
			criterion.BadExamples = rule.Examples.Bad
		}

		rs.Criteria[rule.ID] = criterion
		rs.Categories[criterion.Category] = append(rs.Categories[criterion.Category], criterion)
	}

	return rs
}

// FilterByCategory returns criteria for a specific category.
func (rs *RubricSet) FilterByCategory(category string) []*Criterion {
	return rs.Categories[category]
}

// FilterByRuleIDs returns criteria for specific rule IDs.
func (rs *RubricSet) FilterByRuleIDs(ruleIDs []string) []*Criterion {
	result := make([]*Criterion, 0, len(ruleIDs))
	for _, id := range ruleIDs {
		if c, ok := rs.Criteria[id]; ok {
			result = append(result, c)
		}
	}
	return result
}

// CategoryNames returns all category names in the rubric set.
func (rs *RubricSet) CategoryNames() []string {
	names := make([]string, 0, len(rs.Categories))
	for name := range rs.Categories {
		names = append(names, name)
	}
	return names
}

// AllCriteria returns all criteria as a slice.
func (rs *RubricSet) AllCriteria() []*Criterion {
	result := make([]*Criterion, 0, len(rs.Criteria))
	for _, c := range rs.Criteria {
		result = append(result, c)
	}
	return result
}

// Size returns the number of criteria in the rubric set.
func (rs *RubricSet) Size() int {
	return len(rs.Criteria)
}

// String returns a summary of the rubric set.
func (rs *RubricSet) String() string {
	return fmt.Sprintf("RubricSet{name=%q, criteria=%d, categories=%d}",
		rs.Name, len(rs.Criteria), len(rs.Categories))
}
