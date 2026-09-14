// Package profile provides loading and management of API style profiles.
package profile

import (
	"fmt"
	"strings"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// ValidationError represents a validation error for a profile.
type ValidationError struct {
	RuleID  string
	Field   string
	Detail  string
	Warning bool // If true, this is a warning, not an error
}

func (e ValidationError) Error() string {
	severity := "error"
	if e.Warning {
		severity = "warning"
	}
	return fmt.Sprintf("[%s] rule %s: %s: %s", severity, e.RuleID, e.Field, e.Detail)
}

// ValidationResult contains the results of profile validation.
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
}

// HasErrors returns true if there are any errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any warnings.
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// ValidateProfile validates a profile and returns validation results.
// It checks for:
// - Invalid JSONPath expressions in enforcement.given.paths
// - Missing required fields
// - Semantic issues
func ValidateProfile(spec *types.APIStyleSpec) *ValidationResult {
	result := &ValidationResult{}

	for i := range spec.Rules {
		rule := &spec.Rules[i]
		validateRule(rule, result)
	}

	return result
}

// validateRule validates a single rule.
func validateRule(rule *types.Rule, result *ValidationResult) {
	if rule.Enforcement == nil {
		return
	}

	// Only validate spectral rules with JSONPath expressions
	if rule.Enforcement.Type != types.EnforcementSpectral {
		return
	}

	if rule.Enforcement.Given == nil {
		return
	}

	// Validate each JSONPath expression
	for _, path := range rule.Enforcement.Given.Paths {
		if err := validateJSONPath(path); err != nil {
			result.Warnings = append(result.Warnings, ValidationError{
				RuleID:  rule.ID,
				Field:   "enforcement.given.paths",
				Detail:  fmt.Sprintf("invalid JSONPath %q: %s", path, err),
				Warning: true,
			})
		}
	}
}

// validateJSONPath validates a JSONPath expression using heuristic checks
// for patterns known to cause issues with the vacuum JSONPath parser.
//
// Known unsupported patterns in vacuum's JSONPath implementation:
// - Filter expressions: [?(@...)]
// - Quoted keys with special characters: ['application/json']
// - Keys starting with $: .$ref
func validateJSONPath(path string) error {
	// Check for filter expressions [?(@...)]
	if strings.Contains(path, "[?(@") {
		return fmt.Errorf("filter expressions [?(@...)] not supported")
	}

	// Check for quoted keys with special characters
	// Pattern: ['...'] where content has / or @
	// Need to check ALL occurrences, not just the first
	remaining := path
	for {
		idx := strings.Index(remaining, "['")
		if idx == -1 {
			break
		}
		remaining = remaining[idx+2:]
		endIdx := strings.Index(remaining, "']")
		if endIdx == -1 {
			break
		}
		key := remaining[:endIdx]
		if strings.ContainsAny(key, "/@") {
			return fmt.Errorf("quoted keys with / or @ not supported: ['%s']", key)
		}
		remaining = remaining[endIdx+2:]
	}

	// Check for keys starting with $
	if strings.Contains(path, ".$") {
		return fmt.Errorf("keys starting with $ (like .$ref) not supported")
	}

	return nil
}

// FilterInvalidRules returns a new spec with invalid rules removed.
// Invalid rules are those with JSONPath expressions that cannot be parsed.
// This function also returns the validation result for reporting.
func FilterInvalidRules(spec *types.APIStyleSpec) (*types.APIStyleSpec, *ValidationResult) {
	result := ValidateProfile(spec)

	// Build a set of rule IDs that have JSONPath errors
	invalidRuleIDs := make(map[string]bool)
	for _, w := range result.Warnings {
		if w.Field == "enforcement.given.paths" {
			invalidRuleIDs[w.RuleID] = true
		}
	}

	// If no invalid rules, return the original spec
	if len(invalidRuleIDs) == 0 {
		return spec, result
	}

	// Create a copy with invalid rules removed
	newSpec := *spec // shallow copy
	newSpec.Rules = make([]types.Rule, 0, len(spec.Rules))

	for _, rule := range spec.Rules {
		if !invalidRuleIDs[rule.ID] {
			newSpec.Rules = append(newSpec.Rules, rule)
		}
	}

	return &newSpec, result
}

// DisableInvalidRules modifies the spec in-place to disable rules with
// invalid JSONPath expressions by changing their enforcement type to "none".
// This is an alternative to FilterInvalidRules that preserves the rules
// for LLM evaluation while skipping deterministic linting.
func DisableInvalidRules(spec *types.APIStyleSpec) *ValidationResult {
	result := ValidateProfile(spec)

	// Build a set of rule IDs that have JSONPath errors
	invalidRuleIDs := make(map[string]bool)
	for _, w := range result.Warnings {
		if w.Field == "enforcement.given.paths" {
			invalidRuleIDs[w.RuleID] = true
		}
	}

	// Disable invalid rules by changing enforcement type to "none"
	for i := range spec.Rules {
		if invalidRuleIDs[spec.Rules[i].ID] {
			if spec.Rules[i].Enforcement != nil {
				spec.Rules[i].Enforcement.Type = types.EnforcementNone
			}
		}
	}

	return result
}
