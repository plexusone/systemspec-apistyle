package fix

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func TestRuleFixer_SuggestFixes_PluralResources(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name: "test",
		Rules: []types.Rule{
			{
				ID:       "URI-001",
				Title:    "Use plural resource names",
				Category: "uri",
				Severity: types.SeverityError,
				Examples: &types.Examples{
					Good: []string{"/users", "/orders"},
					Bad:  []string{"/user", "/order"},
				},
			},
		},
	}

	fixer := NewRuleFixer(profile)

	violations := []types.Violation{
		{
			RuleID:   "URI-001",
			Severity: types.SeverityError,
			Message:  "Resource name should be plural",
			Path:     "$.paths['/user']",
		},
	}

	spec := []byte(`openapi: "3.0.0"
paths:
  /user:
    get:
      summary: Get user`)

	report, err := fixer.SuggestFixes(context.Background(), spec, violations, nil)
	require.NoError(t, err)

	assert.Equal(t, 1, report.FixedCount)
	assert.Equal(t, 0, report.UnfixedCount)
	require.Len(t, report.Suggestions, 1)
	assert.Equal(t, "URI-001", report.Suggestions[0].RuleID)
	assert.NotEmpty(t, report.Suggestions[0].SuggestedValue)
}

func TestRuleFixer_SuggestFixes_UnknownRule(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name:  "test",
		Rules: []types.Rule{},
	}

	fixer := NewRuleFixer(profile)

	violations := []types.Violation{
		{
			RuleID:  "UNKNOWN-001",
			Message: "Unknown rule",
			Path:    "$.paths['/test']",
		},
	}

	report, err := fixer.SuggestFixes(context.Background(), nil, violations, nil)
	require.NoError(t, err)

	assert.Equal(t, 0, report.FixedCount)
	assert.Equal(t, 1, report.UnfixedCount)
	assert.Contains(t, report.UnfixedRules, "UNKNOWN-001")
}

func TestRuleFixer_DesignCheck(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name: "test",
		Rules: []types.Rule{
			{
				ID:       "URI-001",
				Title:    "Use plural resources",
				Severity: types.SeverityError,
				Generate: &types.GenerationGuidance{
					Prompt:   "Use plural nouns for collection endpoints",
					Priority: 100,
				},
			},
			{
				ID:        "DOC-001",
				Title:     "Add descriptions",
				Severity:  types.SeverityWarn,
				Rationale: "Descriptions improve API documentation",
			},
		},
	}

	fixer := NewRuleFixer(profile)

	check, err := fixer.DesignCheck(context.Background(), "user", []string{"list", "create", "get"}, nil)
	require.NoError(t, err)

	assert.Len(t, check.Checklist, 2)
	// First item should be highest priority
	assert.Equal(t, "URI-001", check.Checklist[0].RuleID)
	assert.Equal(t, 100, check.Checklist[0].Priority)
	assert.True(t, check.Checklist[0].Required)

	// Template should have paths
	assert.NotNil(t, check.Template)
	paths, ok := check.Template["paths"].(map[string]any)
	assert.True(t, ok)
	assert.Contains(t, paths, "/users")
	assert.Contains(t, paths, "/users/{userId}")
}

func TestExtractPathSegment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$.paths['/users']", "/users"},
		{"$.paths['/user']", "/user"},
		{"$.paths./users", "/users"},
		{"$.info.title", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractPathSegment(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"userName", "user-name"},
		{"UserName", "user-name"},
		{"username", "username"},
		{"APIVersion", "a-p-i-version"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toKebabCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.Equal(t, "default", opts.Profile)
	assert.Equal(t, 50, opts.MaxSuggestions)
	assert.True(t, opts.IncludePatch)
	assert.False(t, opts.UseLLM)
}

func TestRuleFixer_ConformancePath(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name:  "test",
		Rules: []types.Rule{},
	}

	fixer := NewRuleFixer(profile)

	path, err := fixer.ConformancePath(context.Background(), nil, "gold", nil)
	require.NoError(t, err)

	assert.Equal(t, "none", path.CurrentLevel)
	assert.Equal(t, "gold", path.TargetLevel)
	assert.NotNil(t, path.Blockers)
	assert.NotNil(t, path.Warnings)
}

func TestRuleFixer_SuggestFixes_KebabCase(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name: "test",
		Rules: []types.Rule{
			{
				ID:       "KEBAB-001",
				Title:    "Use kebab-case for URLs",
				Category: "uri",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:    "spectral",
					Options: &types.EnforcementOptions{Type: "kebab"},
				},
			},
		},
	}

	fixer := NewRuleFixer(profile)

	violations := []types.Violation{
		{
			RuleID:  "KEBAB-001",
			Message: "URL should use kebab-case",
			Path:    "$.paths['/userAccounts']",
		},
	}

	report, err := fixer.SuggestFixes(context.Background(), nil, violations, nil)
	require.NoError(t, err)

	assert.Equal(t, 1, report.FixedCount)
	require.Len(t, report.Suggestions, 1)
	assert.Equal(t, "/user-accounts", report.Suggestions[0].SuggestedValue)
	assert.Contains(t, report.Suggestions[0].Reasoning, "kebab-case")
}

func TestRuleFixer_SuggestFixes_MaxSuggestions(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name: "test",
		Rules: []types.Rule{
			{
				ID:       "URI-001",
				Title:    "Use plural resources",
				Severity: types.SeverityError,
				Examples: &types.Examples{
					Good: []string{"/users"},
				},
			},
		},
	}

	fixer := NewRuleFixer(profile)

	violations := []types.Violation{
		{RuleID: "URI-001", Path: "$.paths['/user1']"},
		{RuleID: "URI-001", Path: "$.paths['/user2']"},
		{RuleID: "URI-001", Path: "$.paths['/user3']"},
	}

	opts := &Options{MaxSuggestions: 2, IncludePatch: false}
	report, err := fixer.SuggestFixes(context.Background(), nil, violations, opts)
	require.NoError(t, err)

	assert.Equal(t, 2, report.FixedCount)
	assert.Len(t, report.Suggestions, 2)
}

func TestRuleFixer_SuggestFixes_WithPatch(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name: "test",
		Rules: []types.Rule{
			{
				ID:       "URI-001",
				Title:    "Use plural resources",
				Severity: types.SeverityError,
				Examples: &types.Examples{
					Good: []string{"/users"},
				},
			},
		},
	}

	fixer := NewRuleFixer(profile)

	violations := []types.Violation{
		{RuleID: "URI-001", Path: "$.paths['/user']"},
	}

	opts := &Options{IncludePatch: true}
	report, err := fixer.SuggestFixes(context.Background(), nil, violations, opts)
	require.NoError(t, err)

	assert.Len(t, report.PatchOperations, 1)
	assert.Equal(t, "replace", report.PatchOperations[0].Op)
}

func TestJsonPathToPointer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$.paths['/users'].get", "/paths/users/get"},
		{"$.info.title", "/info/title"},
		{"$.components.schemas.User", "/components/schemas/User"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := jsonPathToPointer(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRuleFixer_DesignCheck_SortsByPriority(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name: "test",
		Rules: []types.Rule{
			{
				ID:       "LOW-001",
				Title:    "Low priority rule",
				Severity: types.SeverityError,
				Generate: &types.GenerationGuidance{
					Prompt:   "Low priority instruction",
					Priority: 10,
				},
			},
			{
				ID:       "HIGH-001",
				Title:    "High priority rule",
				Severity: types.SeverityError,
				Generate: &types.GenerationGuidance{
					Prompt:   "High priority instruction",
					Priority: 100,
				},
			},
			{
				ID:       "MED-001",
				Title:    "Medium priority rule",
				Severity: types.SeverityWarn,
				Generate: &types.GenerationGuidance{
					Prompt:   "Medium priority instruction",
					Priority: 50,
				},
			},
		},
	}

	fixer := NewRuleFixer(profile)

	check, err := fixer.DesignCheck(context.Background(), "order", []string{"list", "get"}, nil)
	require.NoError(t, err)

	require.Len(t, check.Checklist, 3)
	// Should be sorted by priority descending
	assert.Equal(t, "HIGH-001", check.Checklist[0].RuleID)
	assert.Equal(t, 100, check.Checklist[0].Priority)
	assert.Equal(t, "MED-001", check.Checklist[1].RuleID)
	assert.Equal(t, "LOW-001", check.Checklist[2].RuleID)
}

func TestRuleFixer_DesignCheck_RequiredFromSeverity(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name: "test",
		Rules: []types.Rule{
			{
				ID:        "ERR-001",
				Title:     "Error severity rule",
				Severity:  types.SeverityError,
				Rationale: "Must do this",
			},
			{
				ID:        "WARN-001",
				Title:     "Warning severity rule",
				Severity:  types.SeverityWarn,
				Rationale: "Should do this",
			},
		},
	}

	fixer := NewRuleFixer(profile)

	check, err := fixer.DesignCheck(context.Background(), "user", []string{"list"}, nil)
	require.NoError(t, err)

	require.Len(t, check.Checklist, 2)

	// Find each rule and check Required
	for _, item := range check.Checklist {
		if item.RuleID == "ERR-001" {
			assert.True(t, item.Required, "Error severity should be required")
		}
		if item.RuleID == "WARN-001" {
			assert.False(t, item.Required, "Warning severity should not be required")
		}
	}
}

func TestRuleFixer_SuggestFixes_MigrationFallback(t *testing.T) {
	profile := &types.APIStyleSpec{
		Name: "test",
		Rules: []types.Rule{
			{
				ID:       "MIG-001",
				Title:    "Use new format",
				Severity: types.SeverityError,
				Migration: &types.MigrationGuidance{
					Steps: []types.MigrationStep{
						{
							Description: "Update to new format",
							Code:        "newFormat: true",
						},
					},
				},
			},
		},
	}

	fixer := NewRuleFixer(profile)

	violations := []types.Violation{
		{RuleID: "MIG-001", Path: "$.info.format"},
	}

	report, err := fixer.SuggestFixes(context.Background(), nil, violations, nil)
	require.NoError(t, err)

	assert.Equal(t, 1, report.FixedCount)
	require.Len(t, report.Suggestions, 1)
	assert.Equal(t, "newFormat: true", report.Suggestions[0].SuggestedValue)
	assert.Contains(t, report.Suggestions[0].Reasoning, "Update to new format")
}
