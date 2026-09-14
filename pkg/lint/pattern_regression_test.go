package lint

import (
	"context"
	"testing"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// These tests guard against externally reported pattern-rule false positives:
// valid semver, kebab-case paths, query-free paths, and body-less GETs were
// flagged as violations.
// The defect does not reproduce with vacuum >= v0.30.0; these tests pin the
// correct behavior in both directions (no false positives, no false negatives).

func patternRegressionSpec() *types.APIStyleSpec {
	return &types.APIStyleSpec{
		Name:    "pattern-regression",
		Version: "1.0.0",
		Rules: []types.Rule{
			{
				ID:       "EXT-007",
				Title:    "Use semantic versioning",
				Category: "meta",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "pattern",
					Given:    types.NewGivenPath("$.info.version"),
					Options:  &types.EnforcementOptions{Match: `^[0-9]+\.[0-9]+\.[0-9]+$`},
				},
			},
			{
				ID:       "EXT-012",
				Title:    "Use kebab-case for path segments",
				Category: "naming",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "pattern",
					Given:    types.NewGivenPath("$.paths[*]~"),
					Options:  &types.EnforcementOptions{Match: `^(/[a-z0-9-{}]+)+$`},
				},
			},
			{
				ID:       "EXT-016",
				Title:    "No query parameters in paths",
				Category: "naming",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:     types.EnforcementSpectral,
					Function: "pattern",
					Given:    types.NewGivenPath("$.paths[*]~"),
					Options:  &types.EnforcementOptions{NotMatch: `\?`},
				},
			},
			{
				ID:       "EXT-020",
				Title:    "GET must not have request body",
				Category: "operations",
				Severity: types.SeverityError,
				Enforcement: &types.Enforcement{
					Type:  types.EnforcementSpectral,
					Given: types.NewGivenPath("$.paths[*].get"),
					Then: &types.SpectralThen{
						Field:    "requestBody",
						Function: "falsy",
					},
				},
			},
		},
	}
}

func lintSpec(t *testing.T, openAPISpec string) *types.LintReport {
	t.Helper()
	linter := NewVacuumLinter(patternRegressionSpec())
	report, err := linter.Lint(context.Background(), []byte(openAPISpec), nil)
	if err != nil {
		t.Fatalf("Lint() error = %v", err)
	}
	return report
}

func violatedRules(report *types.LintReport) map[string]bool {
	rules := make(map[string]bool)
	for _, v := range report.Violations {
		rules[v.RuleID] = true
	}
	return rules
}

func TestPatternRules_NoFalsePositivesOnConformantSpec(t *testing.T) {
	report := lintSpec(t, `
openapi: "3.1.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /api/observe/v1/alive:
    get:
      operationId: getAlive
      responses:
        "200":
          description: OK
`)

	if len(report.Violations) != 0 {
		t.Errorf("conformant spec produced %d violations (want 0): %v",
			len(report.Violations), violatedRules(report))
	}
}

func TestPatternRules_FireOnViolations(t *testing.T) {
	report := lintSpec(t, `
openapi: "3.1.0"
info:
  title: Test API
  version: "v1.0-beta"
paths:
  /api/Observe_v1/Alive:
    get:
      operationId: getAlive
      requestBody:
        content: {}
      responses:
        "200":
          description: OK
`)

	rules := violatedRules(report)
	for _, want := range []string{"EXT-007", "EXT-012", "EXT-020"} {
		if !rules[want] {
			t.Errorf("expected violation of %s, got violations: %v", want, rules)
		}
	}
}

func TestPatternRules_NotMatchFiresOnQueryInPath(t *testing.T) {
	report := lintSpec(t, `
openapi: "3.1.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /api/users?filter=x:
    get:
      operationId: listUsers
      responses:
        "200":
          description: OK
`)

	if !violatedRules(report)["EXT-016"] {
		t.Errorf("expected EXT-016 violation for path containing '?', got: %v",
			violatedRules(report))
	}
}
