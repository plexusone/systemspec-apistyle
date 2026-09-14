package lint

import (
	"github.com/daveshanley/vacuum/model"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// buildVacuumRuleSet converts an APIStyleSpec to a vacuum RuleSet.
func buildVacuumRuleSet(spec *types.APIStyleSpec) map[string]*model.Rule {
	rules := make(map[string]*model.Rule)

	for _, rule := range spec.Rules {
		if rule.Enforcement == nil {
			continue
		}
		if rule.Enforcement.Type != types.EnforcementSpectral {
			continue
		}

		vacuumRule := convertRule(&rule)
		if vacuumRule != nil {
			rules[rule.ID] = vacuumRule
		}
	}

	return rules
}

// convertRule converts a single systemspec-apistyle Rule to a vacuum Rule.
func convertRule(rule *types.Rule) *model.Rule {
	if rule.Enforcement == nil {
		return nil
	}

	// Build the Given paths
	var given any
	if rule.Enforcement.Given != nil && len(rule.Enforcement.Given.Paths) > 0 {
		if len(rule.Enforcement.Given.Paths) == 1 {
			given = rule.Enforcement.Given.Paths[0]
		} else {
			given = rule.Enforcement.Given.Paths
		}
	}

	// Build the Then action
	var then any
	if rule.Enforcement.Then != nil {
		ruleAction := &model.RuleAction{
			Field:    rule.Enforcement.Then.Field,
			Function: rule.Enforcement.Then.Function,
		}
		if rule.Enforcement.Then.FunctionOptions != nil {
			opts := make(map[string]any)
			for k, v := range rule.Enforcement.Then.FunctionOptions {
				opts[k] = v
			}
			ruleAction.FunctionOptions = opts
		}
		then = ruleAction
	} else if rule.Enforcement.Function != "" {
		// Shorthand: function directly on Enforcement
		ruleAction := &model.RuleAction{
			Function: rule.Enforcement.Function,
		}
		if rule.Enforcement.Options != nil {
			ruleAction.FunctionOptions = convertOptions(rule.Enforcement.Options)
		}
		then = ruleAction
	}

	return &model.Rule{
		Id:          rule.ID,
		Description: rule.Title,
		Message:     rule.Title,
		Severity:    string(rule.Severity),
		Given:       given,
		Then:        then,
		Recommended: rule.Recommended,
	}
}

// convertOptions converts EnforcementOptions to a map for vacuum.
func convertOptions(opts *types.EnforcementOptions) map[string]any {
	result := make(map[string]any)

	if opts.Match != "" {
		result["match"] = opts.Match
	}
	if opts.NotMatch != "" {
		result["notMatch"] = opts.NotMatch
	}
	if opts.Min != nil {
		result["min"] = *opts.Min
	}
	if opts.Max != nil {
		result["max"] = *opts.Max
	}
	if len(opts.Values) > 0 {
		result["values"] = opts.Values
	}
	if opts.Type != "" {
		result["type"] = opts.Type
	}
	if opts.Separator != "" {
		result["separator"] = opts.Separator
	}
	if opts.Schema != "" {
		result["schema"] = opts.Schema
	}

	return result
}
