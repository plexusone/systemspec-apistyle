package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-apistyle/pkg/fix"
	"github.com/plexusone/systemspec-apistyle/pkg/lint"
	"github.com/plexusone/systemspec-apistyle/pkg/profile"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

var (
	suggestFormat  string
	suggestOutput  string
	suggestProfile string
	suggestMax     int
	suggestPatch   bool
	suggestVerbose bool
)

var suggestCmd = &cobra.Command{
	Use:   "suggest-fixes <openapi-spec>",
	Short: "Generate fix suggestions for API style violations",
	Long: `Generate fix suggestions for violations found in an OpenAPI specification.

This command lints the specification and then generates actionable fix suggestions
for each violation, including:
  - Suggested replacement values
  - Before/after diffs
  - Confidence scores
  - JSON Patch operations (optional)

Examples:
  api-style suggest-fixes openapi.yaml
  api-style suggest-fixes openapi.yaml --profile azure
  api-style suggest-fixes openapi.yaml --format json --output fixes.json
  api-style suggest-fixes openapi.yaml --patch  # Include JSON Patch operations
  api-style suggest-fixes openapi.yaml --max 10  # Limit to 10 suggestions`,
	Args: cobra.ExactArgs(1),
	RunE: runSuggestFixes,
}

func init() {
	suggestCmd.Flags().StringVarP(&suggestFormat, "format", "f", "text", "Output format: text, json")
	suggestCmd.Flags().StringVarP(&suggestOutput, "output", "o", "", "Output file (default: stdout)")
	suggestCmd.Flags().StringVarP(&suggestProfile, "profile", "p", "", "Style profile to use (default from config or 'default')")
	suggestCmd.Flags().IntVarP(&suggestMax, "max", "m", 50, "Maximum number of suggestions")
	suggestCmd.Flags().BoolVar(&suggestPatch, "patch", false, "Include JSON Patch operations in output")
	suggestCmd.Flags().BoolVarP(&suggestVerbose, "verbose", "v", false, "Show detailed output")
}

func runSuggestFixes(_ *cobra.Command, args []string) error {
	specFile := args[0]

	// Read spec file
	specBytes, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("reading spec file: %w", err)
	}

	// Determine profile
	profileName := suggestProfile
	if profileName == "" {
		profileName = "default"
	}

	// Load profile
	styleSpec, err := profile.Load(profileName)
	if err != nil {
		return fmt.Errorf("loading profile %q: %w", profileName, err)
	}

	ctx := context.Background()

	// Lint to get violations
	linter := lint.NewVacuumLinter(styleSpec)
	lintReport, err := linter.Lint(ctx, specBytes, nil)
	if err != nil {
		return fmt.Errorf("linting: %w", err)
	}

	if len(lintReport.Violations) == 0 {
		fmt.Println("No violations found. Nothing to fix.")
		return nil
	}

	// Generate fix suggestions
	fixer := fix.NewRuleFixer(styleSpec)
	opts := &fix.Options{
		Profile:        profileName,
		MaxSuggestions: suggestMax,
		IncludePatch:   suggestPatch,
	}

	report, err := fixer.SuggestFixes(ctx, specBytes, lintReport.Violations, opts)
	if err != nil {
		return fmt.Errorf("generating fixes: %w", err)
	}

	// Format output
	var output string
	switch strings.ToLower(suggestFormat) {
	case "json":
		output, err = formatSuggestJSON(report)
		if err != nil {
			return fmt.Errorf("formatting output: %w", err)
		}
	case "text":
		output = formatSuggestText(report, lintReport, suggestVerbose)
	default:
		return fmt.Errorf("unknown format: %s", suggestFormat)
	}

	// Write output
	if suggestOutput != "" {
		// Path is from user-provided CLI flag, not untrusted input
		cleanPath := filepath.Clean(suggestOutput)
		if err := os.WriteFile(cleanPath, []byte(output), 0o600); err != nil { //nolint:gosec // G703: user-controlled CLI flag
			return fmt.Errorf("writing output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	return nil
}

func formatSuggestJSON(report *types.FixReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

func formatSuggestText(report *types.FixReport, lintReport *types.LintReport, verbose bool) string {
	var sb strings.Builder

	sb.WriteString("Fix Suggestions\n")
	sb.WriteString("===============\n\n")

	// Summary
	fmt.Fprintf(&sb, "Violations: %d total, %d with fixes, %d unfixable\n\n",
		len(lintReport.Violations),
		report.FixedCount,
		report.UnfixedCount)

	if report.FixedCount == 0 {
		sb.WriteString("No automatic fixes available.\n")
		if len(report.UnfixedRules) > 0 {
			sb.WriteString("\nRules without fixes:\n")
			for _, r := range report.UnfixedRules {
				fmt.Fprintf(&sb, "  - %s\n", r)
			}
		}
		return sb.String()
	}

	// Suggestions
	sb.WriteString("Suggestions:\n\n")

	for i, s := range report.Suggestions {
		fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, s.RuleID, s.Path)

		if s.CurrentValue != "" {
			fmt.Fprintf(&sb, "   Current:   %s\n", s.CurrentValue)
		}
		if s.SuggestedValue != "" {
			fmt.Fprintf(&sb, "   Suggested: %s\n", s.SuggestedValue)
		}

		if verbose {
			if s.Reasoning != "" {
				fmt.Fprintf(&sb, "   Reason:    %s\n", s.Reasoning)
			}
			fmt.Fprintf(&sb, "   Confidence: %.0f%%\n", s.Confidence*100)
		}

		if s.Diff != "" && verbose {
			sb.WriteString("   Diff:\n")
			for _, line := range strings.Split(s.Diff, "\n") {
				fmt.Fprintf(&sb, "     %s\n", line)
			}
		}

		sb.WriteString("\n")
	}

	// Patch operations
	if len(report.PatchOperations) > 0 {
		sb.WriteString("JSON Patch Operations:\n")
		for _, op := range report.PatchOperations {
			fmt.Fprintf(&sb, "  {\"op\": %q, \"path\": %q, \"value\": %v}\n",
				op.Op, op.Path, op.Value)
		}
		sb.WriteString("\n")
	}

	// Unfixed rules
	if len(report.UnfixedRules) > 0 {
		sb.WriteString("Rules without automatic fixes:\n")
		for _, r := range report.UnfixedRules {
			fmt.Fprintf(&sb, "  - %s\n", r)
		}
	}

	return sb.String()
}
