// Package apistyle provides an omniskill Skill for API style specification linting and evaluation.
//
// This package exposes the core systemspec-apistyle functionality as MCP tools:
//   - lint: Lint an OpenAPI spec against style rules
//   - evaluate: LLM-based semantic evaluation
//   - analyze: Combined lint + evaluate with GO/NO-GO decision
//   - list_rules: List all rules from a profile
//   - list_profiles: List available style profiles
//   - explain_rule: Get detailed explanation of a rule
//   - suggest_fixes: Generate fix suggestions for spec violations
//   - design_check: Get design guidance before generating a spec
//   - conformance_path: Show the path to reach a conformance level
//
// The Tools method is the single source of truth for the tool surface;
// consumers (MCP server, CLI) should derive tool metadata from it.
package apistyle

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/plexusone/omniskill/skill"
	apistylespec "github.com/plexusone/systemspec-apistyle"
	"github.com/plexusone/systemspec-apistyle/pkg/analyze"
	"github.com/plexusone/systemspec-apistyle/pkg/fix"
	"github.com/plexusone/systemspec-apistyle/pkg/judge"
	"github.com/plexusone/systemspec-apistyle/pkg/lint"
	"github.com/plexusone/systemspec-apistyle/pkg/profile"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Skill provides API style specification tools.
type Skill struct {
	// anthropicAPIKey is used for LLM evaluation (optional).
	anthropicAPIKey string
}

// Option configures the Skill.
type Option func(*Skill)

// WithAnthropicAPIKey sets the Anthropic API key for LLM evaluation.
func WithAnthropicAPIKey(key string) Option {
	return func(s *Skill) {
		s.anthropicAPIKey = key
	}
}

// New creates a new API style skill.
func New(opts ...Option) *Skill {
	s := &Skill{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Name returns the skill identifier.
func (s *Skill) Name() string {
	return "apistyle"
}

// Description returns what this skill does.
func (s *Skill) Description() string {
	return "Lint and evaluate OpenAPI specifications against style guidelines using deterministic rules and LLM-based semantic analysis"
}

// Version returns the skill version.
func (s *Skill) Version() string {
	return apistylespec.Version
}

// Init initializes the skill (no-op for this skill).
func (s *Skill) Init(_ context.Context) error {
	return nil
}

// Close releases resources (no-op for this skill).
func (s *Skill) Close() error {
	return nil
}

// Tools returns all tools provided by this skill.
func (s *Skill) Tools() []skill.Tool {
	return []skill.Tool{
		s.lintTool(),
		s.evaluateTool(),
		s.analyzeTool(),
		s.listRulesTool(),
		s.listProfilesTool(),
		s.explainRuleTool(),
		s.suggestFixesTool(),
		s.designCheckTool(),
		s.conformancePathTool(),
	}
}

// Ensure Skill implements skill.Skill.
var _ skill.Skill = (*Skill)(nil)

func (s *Skill) lintTool() skill.Tool {
	return skill.NewTool(
		"lint",
		"Lint an OpenAPI specification against API style rules using deterministic checks. Returns violations with severity, rule ID, and location.",
		map[string]skill.Parameter{
			"openapi_spec": {
				Type:        "string",
				Description: "The OpenAPI specification content (YAML or JSON)",
				Required:    true,
			},
			"profile": {
				Type:        "string",
				Description: "Style profile to use (default, minimal, comprehensive, azure, google, microsoft-rest, microsoft-graph, zalando)",
				Required:    false,
				Default:     "default",
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			specContent, _ := params["openapi_spec"].(string)
			if specContent == "" {
				return nil, fmt.Errorf("openapi_spec is required")
			}

			profileName, _ := params["profile"].(string)
			if profileName == "" {
				profileName = "default"
			}

			// Load the style profile
			styleSpec, err := profile.Load(profileName)
			if err != nil {
				return nil, fmt.Errorf("failed to load profile %q: %w", profileName, err)
			}

			// Create linter and run
			linter := lint.NewVacuumLinter(styleSpec)

			report, err := linter.Lint(ctx, []byte(specContent), nil)
			if err != nil {
				return nil, fmt.Errorf("linting failed: %w", err)
			}

			return formatLintReport(report, profileName), nil
		},
	)
}

func (s *Skill) evaluateTool() skill.Tool {
	return skill.NewTool(
		"evaluate",
		"Evaluate an OpenAPI specification using LLM-based semantic analysis. Requires ANTHROPIC_API_KEY. Returns findings with confidence scores.",
		map[string]skill.Parameter{
			"openapi_spec": {
				Type:        "string",
				Description: "The OpenAPI specification content (YAML or JSON)",
				Required:    true,
			},
			"profile": {
				Type:        "string",
				Description: "Style profile to use (default, minimal, comprehensive, azure, google, microsoft-rest, microsoft-graph, zalando)",
				Required:    false,
				Default:     "default",
			},
			"categories": {
				Type:        "array",
				Description: "Categories to evaluate (empty = all)",
				Required:    false,
				Items:       &skill.Parameter{Type: "string"},
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			if s.anthropicAPIKey == "" {
				return nil, fmt.Errorf("ANTHROPIC_API_KEY not configured; LLM evaluation unavailable")
			}

			specContent, _ := params["openapi_spec"].(string)
			if specContent == "" {
				return nil, fmt.Errorf("openapi_spec is required")
			}

			profileName, _ := params["profile"].(string)
			if profileName == "" {
				profileName = "default"
			}

			// Load the style profile
			styleSpec, err := profile.Load(profileName)
			if err != nil {
				return nil, fmt.Errorf("failed to load profile %q: %w", profileName, err)
			}

			// Create evaluator
			provider := judge.NewAnthropicProvider(s.anthropicAPIKey, nil)
			evaluator := judge.NewClaudeEvaluator(provider, styleSpec)

			opts := &judge.Options{}

			// Handle categories
			if cats, ok := params["categories"].([]any); ok && len(cats) > 0 {
				categories := make([]string, len(cats))
				for i, c := range cats {
					categories[i], _ = c.(string)
				}
				opts.Categories = categories
			}

			report, err := evaluator.Evaluate(ctx, []byte(specContent), opts)
			if err != nil {
				return nil, fmt.Errorf("evaluation failed: %w", err)
			}

			return formatEvaluationReport(report), nil
		},
	)
}

func (s *Skill) analyzeTool() skill.Tool {
	return skill.NewTool(
		"analyze",
		"Combined analysis: deterministic linting + LLM evaluation with GO/NO-GO decision. Best for comprehensive API review.",
		map[string]skill.Parameter{
			"openapi_spec": {
				Type:        "string",
				Description: "The OpenAPI specification content (YAML or JSON)",
				Required:    true,
			},
			"profile": {
				Type:        "string",
				Description: "Style profile to use (default, minimal, comprehensive, azure, google, microsoft-rest, microsoft-graph, zalando)",
				Required:    false,
				Default:     "default",
			},
			"lint_only": {
				Type:        "boolean",
				Description: "Skip LLM evaluation, only run deterministic linting",
				Required:    false,
				Default:     false,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			specContent, _ := params["openapi_spec"].(string)
			if specContent == "" {
				return nil, fmt.Errorf("openapi_spec is required")
			}

			profileName, _ := params["profile"].(string)
			if profileName == "" {
				profileName = "default"
			}

			lintOnly, _ := params["lint_only"].(bool)

			// Load the style profile
			styleSpec, err := profile.Load(profileName)
			if err != nil {
				return nil, fmt.Errorf("failed to load profile %q: %w", profileName, err)
			}

			// Create analyzer
			var provider judge.Provider
			if !lintOnly && s.anthropicAPIKey != "" {
				provider = judge.NewAnthropicProvider(s.anthropicAPIKey, nil)
			}

			analyzer := analyze.New(styleSpec, provider)

			opts := &analyze.Options{
				EnableLint:     true,
				EnableEvaluate: !lintOnly && s.anthropicAPIKey != "",
			}

			report, err := analyzer.Analyze(ctx, []byte(specContent), opts)
			if err != nil {
				return nil, fmt.Errorf("analysis failed: %w", err)
			}

			return formatAnalysisReport(report), nil
		},
	)
}

func (s *Skill) listRulesTool() skill.Tool {
	return skill.NewTool(
		"list_rules",
		"List all rules from a style profile with their IDs, titles, categories, and severities.",
		map[string]skill.Parameter{
			"profile": {
				Type:        "string",
				Description: "Style profile to use (default, minimal, comprehensive, azure, google, microsoft-rest, microsoft-graph, zalando)",
				Required:    false,
				Default:     "default",
			},
			"category": {
				Type:        "string",
				Description: "Filter rules by category",
				Required:    false,
			},
			"severity": {
				Type:        "string",
				Description: "Filter rules by severity (error, warn, info, hint)",
				Required:    false,
			},
		},
		func(_ context.Context, params map[string]any) (any, error) {
			profileName, _ := params["profile"].(string)
			if profileName == "" {
				profileName = "default"
			}

			categoryFilter, _ := params["category"].(string)
			severityFilter, _ := params["severity"].(string)

			// Load the style profile
			styleSpec, err := profile.Load(profileName)
			if err != nil {
				return nil, fmt.Errorf("failed to load profile %q: %w", profileName, err)
			}

			var rules []map[string]any
			for _, rule := range styleSpec.Rules {
				// Apply filters
				if categoryFilter != "" && rule.Category != categoryFilter {
					continue
				}
				if severityFilter != "" && string(rule.Severity) != severityFilter {
					continue
				}

				rules = append(rules, map[string]any{
					"id":       rule.ID,
					"title":    rule.Title,
					"category": rule.Category,
					"severity": rule.Severity,
				})
			}

			return map[string]any{
				"profile":     profileName,
				"rule_count":  len(rules),
				"total_rules": len(styleSpec.Rules),
				"rules":       rules,
			}, nil
		},
	)
}

func (s *Skill) listProfilesTool() skill.Tool {
	return skill.NewTool(
		"list_profiles",
		"List all available style profiles with their descriptions and rule counts.",
		map[string]skill.Parameter{},
		func(_ context.Context, _ map[string]any) (any, error) {
			names, err := profile.ListBuiltin()
			if err != nil {
				return nil, fmt.Errorf("listing profiles: %w", err)
			}

			result := make([]map[string]any, 0, len(names))
			for _, name := range names {
				spec, err := profile.Load(name)
				if err != nil {
					continue
				}

				result = append(result, map[string]any{
					"name":        name,
					"description": spec.Description,
					"version":     spec.Version,
					"rule_count":  len(spec.Rules),
				})
			}

			return map[string]any{
				"profiles": result,
			}, nil
		},
	)
}

func (s *Skill) explainRuleTool() skill.Tool {
	return skill.NewTool(
		"explain_rule",
		"Get detailed explanation of a specific rule including rationale, examples, and references.",
		map[string]skill.Parameter{
			"rule_id": {
				Type:        "string",
				Description: "The rule ID to explain (e.g., PO-001)",
				Required:    true,
			},
			"profile": {
				Type:        "string",
				Description: "Style profile to use (default, minimal, comprehensive, azure, google, microsoft-rest, microsoft-graph, zalando)",
				Required:    false,
				Default:     "default",
			},
		},
		func(_ context.Context, params map[string]any) (any, error) {
			ruleID, _ := params["rule_id"].(string)
			if ruleID == "" {
				return nil, fmt.Errorf("rule_id is required")
			}

			profileName, _ := params["profile"].(string)
			if profileName == "" {
				profileName = "default"
			}

			// Load the style profile
			styleSpec, err := profile.Load(profileName)
			if err != nil {
				return nil, fmt.Errorf("failed to load profile %q: %w", profileName, err)
			}

			// Find the rule
			var rule *types.Rule
			for i := range styleSpec.Rules {
				if strings.EqualFold(styleSpec.Rules[i].ID, ruleID) {
					rule = &styleSpec.Rules[i]
					break
				}
			}

			if rule == nil {
				return nil, fmt.Errorf("rule %q not found in profile %q", ruleID, profileName)
			}

			result := map[string]any{
				"id":        rule.ID,
				"title":     rule.Title,
				"category":  rule.Category,
				"severity":  rule.Severity,
				"rationale": rule.Rationale,
			}

			if rule.Examples != nil {
				result["examples"] = map[string]any{
					"good": rule.Examples.Good,
					"bad":  rule.Examples.Bad,
				}
			}

			if len(rule.References) > 0 {
				refs := make([]map[string]string, len(rule.References))
				for i, ref := range rule.References {
					refs[i] = map[string]string{
						"title": ref.Title,
						"url":   ref.URL,
					}
				}
				result["references"] = refs
			}

			if rule.Enforcement != nil {
				result["enforcement"] = map[string]any{
					"type":     rule.Enforcement.Type,
					"function": rule.Enforcement.Function,
				}
			}

			if rule.Judge != nil {
				result["llm_judgeable"] = true
			}

			return result, nil
		},
	)
}

func (s *Skill) suggestFixesTool() skill.Tool {
	return skill.NewTool(
		"suggest_fixes",
		"Generate fix suggestions for OpenAPI spec violations. Returns specific changes to make, with diffs and JSON Patch operations.",
		map[string]skill.Parameter{
			"openapi_spec": {
				Type:        "string",
				Description: "The OpenAPI specification content (YAML or JSON)",
				Required:    true,
			},
			"violations": {
				Type:        "array",
				Description: "List of violations to fix. If empty, lints first and fixes all violations.",
				Required:    false,
				Items:       &skill.Parameter{Type: "object"},
			},
			"profile": {
				Type:        "string",
				Description: "Style profile to use (default, minimal, comprehensive, azure, google, microsoft-rest, microsoft-graph, zalando)",
				Required:    false,
				Default:     "default",
			},
			"max_suggestions": {
				Type:        "integer",
				Description: "Maximum number of suggestions to return (default: 50)",
				Required:    false,
				Default:     50,
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			specContent, _ := params["openapi_spec"].(string)
			if specContent == "" {
				return nil, fmt.Errorf("openapi_spec is required")
			}

			profileName, _ := params["profile"].(string)
			if profileName == "" {
				profileName = "default"
			}

			maxSuggestions := 50
			if ms, ok := params["max_suggestions"].(float64); ok {
				maxSuggestions = int(ms)
			}

			// Load the style profile
			styleSpec, err := profile.Load(profileName)
			if err != nil {
				return nil, fmt.Errorf("failed to load profile %q: %w", profileName, err)
			}

			// Get violations from params or by linting
			var violations []types.Violation
			if v, ok := params["violations"].([]any); ok && len(v) > 0 {
				violations = parseViolations(v)
			} else {
				// Lint to get violations
				linter := lint.NewVacuumLinter(styleSpec)
				report, err := linter.Lint(ctx, []byte(specContent), nil)
				if err != nil {
					return nil, fmt.Errorf("linting failed: %w", err)
				}
				violations = report.Violations
			}

			// Generate fix suggestions
			fixer := fix.NewRuleFixer(styleSpec)
			opts := &fix.Options{
				Profile:        profileName,
				MaxSuggestions: maxSuggestions,
				IncludePatch:   true,
			}

			report, err := fixer.SuggestFixes(ctx, []byte(specContent), violations, opts)
			if err != nil {
				return nil, fmt.Errorf("generating fixes: %w", err)
			}

			return formatFixReport(report), nil
		},
	)
}

func (s *Skill) designCheckTool() skill.Tool {
	return skill.NewTool(
		"design_check",
		"Get design guidance before generating an OpenAPI spec. Returns a checklist of rules to follow and a template.",
		map[string]skill.Parameter{
			"resource_name": {
				Type:        "string",
				Description: "The resource being designed (e.g., 'user', 'order')",
				Required:    true,
			},
			"operations": {
				Type:        "array",
				Description: "Planned operations (e.g., ['list', 'create', 'get', 'update', 'delete'])",
				Required:    false,
				Items:       &skill.Parameter{Type: "string"},
				Default:     []string{"list", "create", "get", "update", "delete"},
			},
			"profile": {
				Type:        "string",
				Description: "Style profile to use",
				Required:    false,
				Default:     "default",
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			resourceName, _ := params["resource_name"].(string)
			if resourceName == "" {
				return nil, fmt.Errorf("resource_name is required")
			}

			profileName, _ := params["profile"].(string)
			if profileName == "" {
				profileName = "default"
			}

			operations := []string{"list", "create", "get", "update", "delete"}
			if ops, ok := params["operations"].([]any); ok && len(ops) > 0 {
				operations = make([]string, len(ops))
				for i, op := range ops {
					operations[i], _ = op.(string)
				}
			}

			// Load the style profile
			styleSpec, err := profile.Load(profileName)
			if err != nil {
				return nil, fmt.Errorf("failed to load profile %q: %w", profileName, err)
			}

			// Generate design check
			fixer := fix.NewRuleFixer(styleSpec)
			check, err := fixer.DesignCheck(ctx, resourceName, operations, nil)
			if err != nil {
				return nil, fmt.Errorf("design check failed: %w", err)
			}

			return formatDesignCheck(check), nil
		},
	)
}

func (s *Skill) conformancePathTool() skill.Tool {
	return skill.NewTool(
		"conformance_path",
		"Show the path to reach a conformance level. Returns blockers, warnings, and progress.",
		map[string]skill.Parameter{
			"openapi_spec": {
				Type:        "string",
				Description: "The OpenAPI specification content",
				Required:    true,
			},
			"target_level": {
				Type:        "string",
				Description: "Target conformance level (bronze, silver, gold)",
				Required:    true,
			},
			"profile": {
				Type:        "string",
				Description: "Style profile to evaluate against",
				Required:    false,
				Default:     "default",
			},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			specContent, _ := params["openapi_spec"].(string)
			if specContent == "" {
				return nil, fmt.Errorf("openapi_spec is required")
			}

			targetLevel, _ := params["target_level"].(string)
			if targetLevel == "" {
				return nil, fmt.Errorf("target_level is required")
			}

			profileName, _ := params["profile"].(string)
			if profileName == "" {
				profileName = "default"
			}

			// Load the style profile
			styleSpec, err := profile.Load(profileName)
			if err != nil {
				return nil, fmt.Errorf("failed to load profile %q: %w", profileName, err)
			}

			// Generate conformance path
			fixer := fix.NewRuleFixer(styleSpec)
			path, err := fixer.ConformancePath(ctx, []byte(specContent), targetLevel, nil)
			if err != nil {
				return nil, fmt.Errorf("conformance path failed: %w", err)
			}

			return formatConformancePath(path), nil
		},
	)
}

// parseViolations converts a slice of any to typed Violations.
func parseViolations(v []any) []types.Violation {
	violations := make([]types.Violation, 0, len(v))
	for _, item := range v {
		if m, ok := item.(map[string]any); ok {
			violation := types.Violation{}
			if ruleID, ok := m["ruleId"].(string); ok {
				violation.RuleID = ruleID
			} else if ruleID, ok := m["rule_id"].(string); ok {
				violation.RuleID = ruleID
			}
			if path, ok := m["path"].(string); ok {
				violation.Path = path
			}
			if msg, ok := m["message"].(string); ok {
				violation.Message = msg
			}
			violations = append(violations, violation)
		}
	}
	return violations
}

// formatFixReport converts a fix report to a map for JSON serialization.
func formatFixReport(report *types.FixReport) map[string]any {
	suggestions := make([]map[string]any, len(report.Suggestions))
	for i, s := range report.Suggestions {
		suggestions[i] = map[string]any{
			"rule_id":         s.RuleID,
			"path":            s.Path,
			"current_value":   s.CurrentValue,
			"suggested_value": s.SuggestedValue,
			"diff":            s.Diff,
			"confidence":      s.Confidence,
			"reasoning":       s.Reasoning,
		}
	}

	patchOps := make([]map[string]any, len(report.PatchOperations))
	for i, p := range report.PatchOperations {
		patchOps[i] = map[string]any{
			"op":    p.Op,
			"path":  p.Path,
			"value": p.Value,
		}
		if p.From != "" {
			patchOps[i]["from"] = p.From
		}
	}

	return map[string]any{
		"suggestions":      suggestions,
		"patch_operations": patchOps,
		"fixed_count":      report.FixedCount,
		"unfixed_count":    report.UnfixedCount,
		"unfixed_rules":    report.UnfixedRules,
	}
}

// formatDesignCheck converts a design check to a map for JSON serialization.
func formatDesignCheck(check *types.DesignCheck) map[string]any {
	checklist := make([]map[string]any, len(check.Checklist))
	for i, item := range check.Checklist {
		checklist[i] = map[string]any{
			"rule_id":     item.RuleID,
			"instruction": item.Instruction,
			"priority":    item.Priority,
			"required":    item.Required,
		}
	}

	return map[string]any{
		"checklist": checklist,
		"template":  check.Template,
		"warnings":  check.Warnings,
	}
}

// formatConformancePath converts a conformance path to a map for JSON serialization.
func formatConformancePath(path *types.ConformancePath) map[string]any {
	blockers := make([]map[string]any, len(path.Blockers))
	for i, b := range path.Blockers {
		blockers[i] = map[string]any{
			"rule_id":          b.RuleID,
			"count":            b.Count,
			"priority":         b.Priority,
			"fix_instructions": b.FixInstructions,
		}
	}

	warnings := make([]map[string]any, len(path.Warnings))
	for i, w := range path.Warnings {
		warnings[i] = map[string]any{
			"rule_id":          w.RuleID,
			"count":            w.Count,
			"priority":         w.Priority,
			"fix_instructions": w.FixInstructions,
		}
	}

	return map[string]any{
		"current_level":      path.CurrentLevel,
		"target_level":       path.TargetLevel,
		"blockers":           blockers,
		"warnings":           warnings,
		"progress_to_target": path.ProgressToTarget,
		"estimated_fixes":    path.EstimatedFixes,
	}
}

// formatLintReport converts a lint report to a map for JSON serialization.
func formatLintReport(report *types.LintReport, profileName string) map[string]any {
	violations := make([]map[string]any, len(report.Violations))
	for i, v := range report.Violations {
		violations[i] = map[string]any{
			"rule_id":  v.RuleID,
			"severity": v.Severity,
			"message":  v.Message,
			"path":     v.Path,
			"line":     v.Line,
		}
	}

	return map[string]any{
		"status":          report.Status,
		"violation_count": len(report.Violations),
		"violations":      violations,
		"profile":         profileName,
		"summary": map[string]any{
			"errors":   report.Summary.Errors,
			"warnings": report.Summary.Warnings,
			"infos":    report.Summary.Infos,
			"hints":    report.Summary.Hints,
			"total":    report.Summary.Total,
		},
	}
}

// formatEvaluationReport converts an evaluation report to a map for JSON serialization.
func formatEvaluationReport(report *judge.EvaluationReport) map[string]any {
	findings := make([]map[string]any, len(report.Findings))
	for i, f := range report.Findings {
		findings[i] = map[string]any{
			"rule_id":     f.RuleID,
			"passed":      f.Passed,
			"score":       f.Score,
			"reasoning":   f.Reasoning,
			"suggestions": f.Suggestions,
		}
	}

	categories := make([]map[string]any, len(report.Categories))
	for i, c := range report.Categories {
		categories[i] = map[string]any{
			"name":  c.Name,
			"score": c.Score,
		}
	}

	return map[string]any{
		"status":        report.Status,
		"summary":       report.Summary,
		"categories":    categories,
		"findings":      findings,
		"finding_count": len(report.Findings),
	}
}

// formatAnalysisReport converts an analysis report to a map for JSON serialization.
func formatAnalysisReport(report *analyze.AnalysisReport) map[string]any {
	result := map[string]any{
		"decision": report.Decision,
		"summary":  report.Summary,
	}

	if report.LintReport != nil {
		result["lint"] = formatLintReport(report.LintReport, "")
	}

	if report.EvaluationReport != nil {
		result["evaluation"] = formatEvaluationReport(report.EvaluationReport)
	}

	return result
}

// MarshalJSON implements json.Marshaler for Skill metadata.
func (s *Skill) MarshalJSON() ([]byte, error) {
	tools := s.Tools()
	toolNames := make([]string, len(tools))
	for i, t := range tools {
		toolNames[i] = t.Name()
	}

	return json.Marshal(map[string]any{
		"name":        s.Name(),
		"description": s.Description(),
		"tools":       toolNames,
	})
}
