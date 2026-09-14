package profile

import (
	"testing"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func TestValidateJSONPath_Valid(t *testing.T) {
	validPaths := []string{
		"$.openapi",
		"$.info.title",
		"$.paths[*]~",
		"$.paths[*][*].responses",
		"$.components.schemas[*].properties[*]~",
		"$.paths[*].post.responses['201']", // quoted numeric key is ok
	}

	for _, path := range validPaths {
		if err := validateJSONPath(path); err != nil {
			t.Errorf("validateJSONPath(%q) returned error: %v", path, err)
		}
	}
}

func TestValidateJSONPath_Invalid(t *testing.T) {
	testCases := []struct {
		path   string
		reason string
	}{
		{"$.paths[?(@.type == 'object')]", "filter expression"},
		{"$.components.schemas[?(@.oneOf)]", "existence filter"},
		{"$.paths[*][*].responses['200'].content['application/json']", "quoted key with /"},
		{"$.paths[*][*].requestBody.content['application/json'].schema.$ref", "$ref key"},
	}

	for _, tc := range testCases {
		if err := validateJSONPath(tc.path); err == nil {
			t.Errorf("validateJSONPath(%q) should return error for %s", tc.path, tc.reason)
		}
	}
}

func TestValidateProfile(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "VALID-001",
				Title:    "Valid rule",
				Category: "test",
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info.title"),
				},
			},
			{
				ID:       "INVALID-001",
				Title:    "Invalid rule",
				Category: "test",
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.schemas[?(@.type == 'object')]"),
				},
			},
		},
	}

	result := ValidateProfile(spec)

	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}

	if len(result.Warnings) > 0 && result.Warnings[0].RuleID != "INVALID-001" {
		t.Errorf("expected warning for INVALID-001, got %s", result.Warnings[0].RuleID)
	}
}

func TestFilterInvalidRules(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "VALID-001",
				Title:    "Valid rule",
				Category: "test",
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info.title"),
				},
			},
			{
				ID:       "INVALID-001",
				Title:    "Invalid rule",
				Category: "test",
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.schemas[?(@.type == 'object')]"),
				},
			},
			{
				ID:       "VALID-002",
				Title:    "Another valid rule",
				Category: "test",
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "pattern",
					Given:    types.NewGivenPath("$.openapi"),
					Options:  &types.EnforcementOptions{Match: "^3\\."},
				},
			},
		},
	}

	filtered, result := FilterInvalidRules(spec)

	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}

	if len(filtered.Rules) != 2 {
		t.Errorf("expected 2 rules after filtering, got %d", len(filtered.Rules))
	}

	// Check that the invalid rule was removed
	for _, rule := range filtered.Rules {
		if rule.ID == "INVALID-001" {
			t.Error("INVALID-001 should have been filtered out")
		}
	}
}

func TestDisableInvalidRules(t *testing.T) {
	spec := &types.APIStyleSpec{
		Name:    "test",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "VALID-001",
				Title:    "Valid rule",
				Category: "test",
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.info.title"),
				},
			},
			{
				ID:       "INVALID-001",
				Title:    "Invalid rule",
				Category: "test",
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "truthy",
					Given:    types.NewGivenPath("$.schemas[?(@.type == 'object')]"),
				},
			},
		},
	}

	result := DisableInvalidRules(spec)

	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}

	// Check that the invalid rule was disabled
	var invalidRule *types.Rule
	for i := range spec.Rules {
		if spec.Rules[i].ID == "INVALID-001" {
			invalidRule = &spec.Rules[i]
			break
		}
	}

	if invalidRule == nil {
		t.Fatal("INVALID-001 should still be in the spec")
	}

	if invalidRule.Enforcement.Type != types.EnforcementNone {
		t.Errorf("INVALID-001 enforcement type should be 'none', got %s", invalidRule.Enforcement.Type)
	}

	// Valid rule should still be spectral
	var validRule *types.Rule
	for i := range spec.Rules {
		if spec.Rules[i].ID == "VALID-001" {
			validRule = &spec.Rules[i]
			break
		}
	}

	if validRule.Enforcement.Type != types.EnforcementSpectral {
		t.Errorf("VALID-001 enforcement type should still be 'spectral', got %s", validRule.Enforcement.Type)
	}
}
