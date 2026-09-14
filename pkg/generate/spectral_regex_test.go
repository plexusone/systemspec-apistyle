package generate

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// TestSpectral_RegexEscaping_FromJSON tests the full pipeline from JSON profile
// to generated Spectral YAML. This catches double-escaping bugs where the JSON
// deserialization + YAML serialization pipeline produces quadruple backslashes.
//
// A customer style guide reported regex patterns like:
//
//	JSON source: "^(\\/[a-z][a-z0-9\\-]*(\\/{[^}]+})?)*$"
//
// After json.Unmarshal, Go string is:
//
//	`^(\/[a-z][a-z0-9\-]*(\/\{[^}]+})?)*$`
//
// When written to YAML double-quoted string, each `\` becomes `\\`:
//
//	Expected YAML: "^(\\/[a-z][a-z0-9\\-]*(\\/\\{[^}]+})?)*$"
//
// Bug produces YAML: "^(\\\\/[a-z][a-z0-9\\\\-]*(\\\\/{[^}]+})?)*$"
// (quadruple backslashes — Spectral interprets as literal `\\` in the regex)
func TestSpectral_RegexEscaping_FromJSON(t *testing.T) {
	// This JSON mimics the reported style guide's kebab-case rule
	profileJSON := `{
		"name": "test-escaping",
		"version": "1.0.0",
		"rules": [
			{
				"id": "EXT-012",
				"title": "Use kebab-case for path segments",
				"category": "urls",
				"severity": "error",
				"enforcement": {
					"type": "spectral",
					"function": "pattern",
					"given": {"paths": ["$.paths[*]~"]},
					"options": {
						"match": "^(\\/[a-z][a-z0-9\\-]*(\\/{[^}]+})?)*$"
					}
				}
			},
			{
				"id": "EXT-007",
				"title": "Use semantic versioning",
				"category": "versioning",
				"severity": "error",
				"enforcement": {
					"type": "spectral",
					"function": "pattern",
					"given": {"paths": ["$.info.version"]},
					"options": {
						"match": "^[0-9]+\\.[0-9]+\\.[0-9]+$"
					}
				}
			},
			{
				"id": "EXT-013",
				"title": "No trailing slash",
				"category": "urls",
				"severity": "error",
				"enforcement": {
					"type": "spectral",
					"function": "pattern",
					"given": {"paths": ["$.paths[*]~"]},
					"options": {
						"notMatch": "\\/$"
					}
				}
			}
		]
	}`

	// Step 1: Unmarshal from JSON (same as profile loader does)
	var spec types.APIStyleSpec
	if err := json.Unmarshal([]byte(profileJSON), &spec); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Step 2: Generate Spectral YAML
	yaml, err := Spectral(&spec, nil)
	if err != nil {
		t.Fatalf("Spectral() error = %v", err)
	}

	// Step 3: Verify NO quadruple backslashes exist (the double-escaping bug)
	if strings.Contains(yaml, `\\\\`) {
		t.Errorf("Generated YAML contains quadruple backslashes (\\\\\\\\), indicating double-escaping bug.\nGenerated YAML:\n%s", yaml)
	}

	// Step 4: Verify specific patterns are correctly escaped for YAML
	tests := []struct {
		name     string
		ruleID   string
		contains string
		absent   string
	}{
		{
			name:     "EXT-012 kebab-case regex has correct escaping",
			ruleID:   "ext-012",
			contains: `\/[a-z]`,
			absent:   `\\\\/[a-z]`,
		},
		{
			name:     "EXT-007 semver regex has correct dot escaping",
			ruleID:   "ext-007",
			contains: `[0-9]+\\.[0-9]+\\.[0-9]+`, // YAML `\\.` = regex `\.` (escaped dot) — CORRECT
			absent:   `[0-9]+\\\\.[0-9]+`,        // Quadruple-backslash would be wrong
		},
		{
			name:     "EXT-013 trailing slash regex has correct escaping",
			ruleID:   "ext-013",
			contains: `\/$`,
			absent:   `\\\\/$`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleIdx := strings.Index(yaml, tt.ruleID+":")
			if ruleIdx == -1 {
				t.Fatalf("Rule %s not found in generated YAML", tt.ruleID)
			}

			ruleSection := yaml[ruleIdx:]
			nextRule := strings.Index(ruleSection[1:], "\n  # ")
			if nextRule > 0 {
				ruleSection = ruleSection[:nextRule+1]
			}

			if !strings.Contains(ruleSection, tt.contains) {
				t.Errorf("Expected pattern containing %q not found.\nRule section:\n%s", tt.contains, ruleSection)
			}

			if tt.absent != "" && strings.Contains(ruleSection, tt.absent) {
				t.Errorf("Found double-escaped pattern %q (should not be present).\nRule section:\n%s", tt.absent, ruleSection)
			}
		})
	}
}

// TestSpectral_RegexRoundtrip_MatchesRealPaths verifies that the regex pattern
// from a JSON profile, after unmarshal, actually matches real API paths.
// This is the ultimate correctness check — if the regex matches in Go's regexp
// engine, it should also match in Spectral's JS regex engine for these patterns.
func TestSpectral_RegexRoundtrip_MatchesRealPaths(t *testing.T) {
	profileJSON := `{
		"name": "test",
		"version": "1.0.0",
		"rules": [{
			"id": "EXT-012",
			"title": "kebab-case",
			"category": "urls",
			"severity": "error",
			"enforcement": {
				"type": "spectral",
				"function": "pattern",
				"given": {"paths": ["$.paths[*]~"]},
				"options": {
					"match": "^(\\/[a-z][a-z0-9\\-]*(\\/{[^}]+})?)*$"
				}
			}
		}]
	}`

	var spec types.APIStyleSpec
	if err := json.Unmarshal([]byte(profileJSON), &spec); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Extract the regex after unmarshal — this is what Spectral() receives
	matchPattern := spec.Rules[0].Enforcement.Options.Match

	re, err := regexp.Compile(matchPattern)
	if err != nil {
		t.Fatalf("Regex from JSON unmarshal doesn't compile: %v\nPattern: %q", err, matchPattern)
	}

	// These paths should match the kebab-case regex
	validPaths := []string{
		"/users",
		"/user-accounts",
		"/api/observe/v1/ready",
		"/api/observe/v1/alive",
		"/access-requests/{requestId}/approvals",
		"/api/users/v2/accounts",
	}

	// These paths should NOT match
	invalidPaths := []string{
		"/userAccounts",  // camelCase
		"/User-Accounts", // uppercase
		"/user_accounts", // snake_case
	}

	for _, path := range validPaths {
		if !re.MatchString(path) {
			t.Errorf("Valid path %q should match kebab-case regex but doesn't.\nRegex: %q", path, matchPattern)
		}
	}

	for _, path := range invalidPaths {
		if re.MatchString(path) {
			t.Errorf("Invalid path %q should NOT match kebab-case regex but does.\nRegex: %q", path, matchPattern)
		}
	}
}
