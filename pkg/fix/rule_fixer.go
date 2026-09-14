package fix

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// RuleFixer generates fixes using rule metadata.
type RuleFixer struct {
	profile *types.APIStyleSpec
}

// NewRuleFixer creates a fixer from a style profile.
func NewRuleFixer(profile *types.APIStyleSpec) *RuleFixer {
	return &RuleFixer{profile: profile}
}

// SuggestFixes implements Fixer.
func (f *RuleFixer) SuggestFixes(ctx context.Context, spec []byte, violations []types.Violation, opts *Options) (*types.FixReport, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	report := &types.FixReport{
		Suggestions: make([]types.FixSuggestion, 0, len(violations)),
	}

	// Build rule lookup
	ruleLookup := f.buildRuleLookup()

	// Process each violation
	for _, v := range violations {
		if opts.MaxSuggestions > 0 && report.FixedCount >= opts.MaxSuggestions {
			break
		}

		rule, ok := ruleLookup[v.RuleID]
		if !ok {
			report.UnfixedCount++
			if !contains(report.UnfixedRules, v.RuleID) {
				report.UnfixedRules = append(report.UnfixedRules, v.RuleID)
			}
			continue
		}

		suggestion, err := f.generateSuggestion(ctx, spec, v, rule, opts)
		if err != nil {
			report.UnfixedCount++
			if !contains(report.UnfixedRules, v.RuleID) {
				report.UnfixedRules = append(report.UnfixedRules, v.RuleID)
			}
			continue
		}

		report.Suggestions = append(report.Suggestions, *suggestion)
		report.FixedCount++
	}

	// Generate patch operations if requested
	if opts.IncludePatch {
		report.PatchOperations = f.generatePatch(report.Suggestions)
	}

	return report, nil
}

// DesignCheck implements Fixer.
func (f *RuleFixer) DesignCheck(_ context.Context, resource string, operations []string, _ *Options) (*types.DesignCheck, error) {
	check := &types.DesignCheck{
		Checklist: make([]types.DesignCheckItem, 0),
		Warnings:  make([]string, 0),
	}

	// Collect rules relevant to the operations
	for _, rule := range f.profile.Rules {
		if rule.Severity != types.SeverityError && rule.Severity != types.SeverityWarn {
			continue
		}

		item := types.DesignCheckItem{
			RuleID:   rule.ID,
			Priority: 50, // Default priority
			Required: rule.Severity == types.SeverityError,
		}

		// Use generation guidance if available
		if rule.Generate != nil {
			item.Instruction = rule.Generate.Prompt
			if rule.Generate.Priority > 0 {
				item.Priority = rule.Generate.Priority
			}
		} else if rule.Rationale != "" {
			item.Instruction = rule.Rationale
		} else {
			item.Instruction = rule.Title
		}

		check.Checklist = append(check.Checklist, item)
	}

	// Sort by priority (highest first)
	sort.Slice(check.Checklist, func(i, j int) bool {
		return check.Checklist[i].Priority > check.Checklist[j].Priority
	})

	// Generate template for the resource
	check.Template = f.generateResourceTemplate(resource, operations)

	return check, nil
}

// ConformancePath implements Fixer.
func (f *RuleFixer) ConformancePath(_ context.Context, _ []byte, targetLevel string, _ *Options) (*types.ConformancePath, error) {
	path := &types.ConformancePath{
		CurrentLevel: "none",
		TargetLevel:  targetLevel,
		Blockers:     make([]types.ConformanceBlocker, 0),
		Warnings:     make([]types.ConformanceBlocker, 0),
	}

	// This would normally lint the spec and analyze results
	// For now, return a basic structure
	path.ProgressToTarget = 0.0
	path.EstimatedFixes = 0

	return path, nil
}

// buildRuleLookup creates a map from rule ID to rule.
func (f *RuleFixer) buildRuleLookup() map[string]*types.Rule {
	lookup := make(map[string]*types.Rule)
	for i := range f.profile.Rules {
		lookup[f.profile.Rules[i].ID] = &f.profile.Rules[i]
	}
	return lookup
}

// generateSuggestion creates a fix suggestion for a single violation.
func (f *RuleFixer) generateSuggestion(_ context.Context, _ []byte, v types.Violation, rule *types.Rule, _ *Options) (*types.FixSuggestion, error) {
	suggestion := &types.FixSuggestion{
		RuleID:     v.RuleID,
		Path:       v.Path,
		Confidence: 0.8, // Default confidence for rule-based fixes
	}

	// Try to extract current value from path
	suggestion.CurrentValue = extractPathSegment(v.Path)

	// Try pattern-based fix for common rules
	if fix := f.tryPatternFix(v, rule); fix != nil {
		suggestion.SuggestedValue = fix.Value
		suggestion.Reasoning = fix.Reasoning
		suggestion.Confidence = fix.Confidence
		suggestion.Diff = fmt.Sprintf("- %s\n+ %s", suggestion.CurrentValue, suggestion.SuggestedValue)
		return suggestion, nil
	}

	// Try example-based fix
	if fix := f.tryExampleFix(v, rule); fix != nil {
		suggestion.SuggestedValue = fix.Value
		suggestion.Reasoning = fix.Reasoning
		suggestion.Confidence = fix.Confidence
		suggestion.Diff = fmt.Sprintf("- %s\n+ %s", suggestion.CurrentValue, suggestion.SuggestedValue)
		return suggestion, nil
	}

	// Fall back to migration guidance
	if rule.Migration != nil && len(rule.Migration.Steps) > 0 {
		suggestion.Reasoning = rule.Migration.Steps[0].Description
		suggestion.SuggestedValue = rule.Migration.Steps[0].Code
		suggestion.Confidence = 0.6
		return suggestion, nil
	}

	return nil, fmt.Errorf("no fix strategy available for rule %s", v.RuleID)
}

// tryPatternFix attempts to fix based on rule enforcement patterns.
func (f *RuleFixer) tryPatternFix(v types.Violation, rule *types.Rule) *fixResult {
	if rule.Enforcement == nil || rule.Enforcement.Options == nil {
		return nil
	}

	// Handle common pattern transformations
	switch {
	case strings.Contains(strings.ToLower(rule.ID), "plural"):
		// Plural resource fix
		current := extractPathSegment(v.Path)
		if current != "" && !strings.HasSuffix(current, "s") {
			return &fixResult{
				Value:      current + "s",
				Reasoning:  "Added plural suffix 's' to make resource name plural",
				Confidence: 0.9,
			}
		}

	case strings.Contains(strings.ToLower(rule.ID), "kebab"):
		// Kebab-case fix
		current := extractPathSegment(v.Path)
		if current != "" {
			fixed := toKebabCase(current)
			if fixed != current {
				return &fixResult{
					Value:      fixed,
					Reasoning:  "Converted to kebab-case",
					Confidence: 0.95,
				}
			}
		}

	case strings.Contains(strings.ToLower(rule.ID), "lowercase"):
		// Lowercase fix
		current := extractPathSegment(v.Path)
		if current != "" {
			fixed := strings.ToLower(current)
			if fixed != current {
				return &fixResult{
					Value:      fixed,
					Reasoning:  "Converted to lowercase",
					Confidence: 1.0,
				}
			}
		}
	}

	return nil
}

// tryExampleFix attempts to fix based on good examples.
func (f *RuleFixer) tryExampleFix(_ types.Violation, rule *types.Rule) *fixResult {
	if rule.Examples == nil || len(rule.Examples.Good) == 0 {
		return nil
	}

	// Use first good example as template
	return &fixResult{
		Value:      rule.Examples.Good[0],
		Reasoning:  fmt.Sprintf("Based on good example from rule %s", rule.ID),
		Confidence: 0.7,
	}
}

// generatePatch creates JSON Patch operations from suggestions.
func (f *RuleFixer) generatePatch(suggestions []types.FixSuggestion) []types.PatchOperation {
	operations := make([]types.PatchOperation, 0, len(suggestions))

	for _, s := range suggestions {
		if s.SuggestedValue == "" {
			continue
		}

		// Convert JSONPath to JSON Pointer
		pointer := jsonPathToPointer(s.Path)

		op := types.PatchOperation{
			Op:    "replace",
			Path:  pointer,
			Value: s.SuggestedValue,
		}
		operations = append(operations, op)
	}

	return operations
}

// generateResourceTemplate creates an OpenAPI template for a resource.
func (f *RuleFixer) generateResourceTemplate(resource string, operations []string) map[string]any {
	// Pluralize resource name
	plural := resource
	if !strings.HasSuffix(resource, "s") {
		plural = resource + "s"
	}

	paths := make(map[string]any)
	collectionPath := "/" + plural
	itemPath := "/" + plural + "/{" + resource + "Id}"

	// Collection endpoints
	collectionOps := make(map[string]any)
	if contains(operations, "list") {
		collectionOps["get"] = map[string]any{
			"operationId": "list" + strings.Title(plural),
			"summary":     "List " + plural,
		}
	}
	if contains(operations, "create") {
		collectionOps["post"] = map[string]any{
			"operationId": "create" + strings.Title(resource),
			"summary":     "Create a " + resource,
		}
	}
	if len(collectionOps) > 0 {
		paths[collectionPath] = collectionOps
	}

	// Item endpoints
	itemOps := make(map[string]any)
	if contains(operations, "get") {
		itemOps["get"] = map[string]any{
			"operationId": "get" + strings.Title(resource),
			"summary":     "Get a " + resource,
		}
	}
	if contains(operations, "update") {
		itemOps["put"] = map[string]any{
			"operationId": "update" + strings.Title(resource),
			"summary":     "Update a " + resource,
		}
	}
	if contains(operations, "delete") {
		itemOps["delete"] = map[string]any{
			"operationId": "delete" + strings.Title(resource),
			"summary":     "Delete a " + resource,
		}
	}
	if len(itemOps) > 0 {
		paths[itemPath] = itemOps
	}

	return map[string]any{
		"paths": paths,
	}
}

// fixResult holds the result of a fix attempt.
type fixResult struct {
	Value      string
	Reasoning  string
	Confidence float64
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractPathSegment(jsonPath string) string {
	// Extract the last path segment from a JSONPath like $.paths['/users']
	re := regexp.MustCompile(`\['([^']+)'\]$`)
	matches := re.FindStringSubmatch(jsonPath)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try extracting from $.paths./users format
	re = regexp.MustCompile(`\./([^/\[]+)`)
	matches = re.FindStringSubmatch(jsonPath)
	if len(matches) > 1 {
		return "/" + matches[1]
	}

	return ""
}

func toKebabCase(s string) string {
	// Convert camelCase or PascalCase to kebab-case
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func jsonPathToPointer(jsonPath string) string {
	// Convert JSONPath to JSON Pointer
	// $.paths['/users'].get -> /paths/~1users/get
	pointer := strings.TrimPrefix(jsonPath, "$.")
	pointer = strings.ReplaceAll(pointer, "['", "/")
	pointer = strings.ReplaceAll(pointer, "']", "")
	pointer = strings.ReplaceAll(pointer, "[", "/")
	pointer = strings.ReplaceAll(pointer, "]", "")
	pointer = strings.ReplaceAll(pointer, ".", "/")
	pointer = strings.ReplaceAll(pointer, "//", "/")

	// Escape special characters per RFC 6901
	pointer = strings.ReplaceAll(pointer, "~", "~0")
	pointer = strings.ReplaceAll(pointer, "/", "~1")

	// Restore path separators
	pointer = "/" + strings.ReplaceAll(pointer, "~1", "/")
	pointer = strings.ReplaceAll(pointer, "~0", "~")

	return pointer
}
