package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-apistyle/pkg/analyze"
	"github.com/plexusone/systemspec-apistyle/pkg/judge"
	"github.com/plexusone/systemspec-apistyle/pkg/profile"
)

var (
	evalFormat  string
	evalOutput  string
	evalProfile string
	evalModel   string
)

var evaluateCmd = &cobra.Command{
	Use:   "evaluate <openapi-spec>",
	Short: "Evaluate an OpenAPI specification using LLM",
	Long: `Evaluate an OpenAPI specification using AI-powered analysis.

This command uses an LLM (Claude by default) to assess the API specification
against style rules that cannot be checked deterministically.

Requires ANTHROPIC_API_KEY environment variable to be set.

Examples:
  api-style evaluate openapi.yaml
  api-style evaluate openapi.yaml --format json
  api-style evaluate openapi.yaml --profile azure --model claude-sonnet-4`,
	Args: cobra.ExactArgs(1),
	RunE: runEvaluate,
}

func init() {
	evaluateCmd.Flags().StringVarP(&evalFormat, "format", "f", "text", "Output format: text, json")
	evaluateCmd.Flags().StringVarP(&evalOutput, "output", "o", "", "Output file (default: stdout)")
	evaluateCmd.Flags().StringVarP(&evalProfile, "profile", "p", "default", "Style profile to use")
	evaluateCmd.Flags().StringVarP(&evalModel, "model", "m", "", "LLM model to use (default: claude-3-5-haiku)")
}

func runEvaluate(_ *cobra.Command, args []string) error {
	specFile := args[0]

	// Check for API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	// Read the OpenAPI spec
	specBytes, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("reading spec file: %w", err)
	}

	// Load profile
	styleSpec, err := profile.Load(evalProfile)
	if err != nil {
		return fmt.Errorf("loading profile %q: %w", evalProfile, err)
	}

	// Create provider
	provider := judge.NewAnthropicProvider("", nil)
	if evalModel != "" {
		provider.SetDefaultModel(evalModel)
	}

	// Create analyzer
	analyzer := analyze.New(styleSpec, provider)

	// Run evaluation
	ctx := context.Background()
	opts := &analyze.Options{
		FileName:       specFile,
		EnableLint:     false,
		EnableEvaluate: true,
		EvaluateOptions: &judge.Options{
			FileName:         specFile,
			IncludeReasoning: true,
			Model:            evalModel,
		},
	}

	report, err := analyzer.Evaluate(ctx, specBytes, opts)
	if err != nil {
		return fmt.Errorf("evaluating: %w", err)
	}

	// Format output
	output, err := formatEvaluationReport(report, evalFormat)
	if err != nil {
		return fmt.Errorf("formatting report: %w", err)
	}

	// Write output
	if evalOutput != "" {
		//nolint:gosec // G703: Path from CLI flag
		if err := os.WriteFile(evalOutput, []byte(output), 0o600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	// Exit with code 1 if there are critical failures
	if report.HasCriticalFailures() {
		os.Exit(1)
	}

	return nil
}

func formatEvaluationReport(report *judge.EvaluationReport, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	case "text":
		return formatEvaluationText(report), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

func formatEvaluationText(report *judge.EvaluationReport) string {
	var sb strings.Builder

	sb.WriteString("API Style Evaluation Report\n")
	sb.WriteString("===========================\n\n")

	// Summary
	fmt.Fprintf(&sb, "Status: %s\n", strings.ToUpper(string(report.Status)))
	fmt.Fprintf(&sb, "Overall Score: %.1f%%\n", report.Summary.OverallScore*100)
	fmt.Fprintf(&sb, "Rules Evaluated: %d (Passed: %d, Failed: %d)\n\n",
		report.Summary.TotalRules,
		report.Summary.PassedRules,
		report.Summary.FailedRules)

	if len(report.Findings) == 0 {
		sb.WriteString("No rules were evaluated.\n")
		return sb.String()
	}

	// Findings by category
	for _, cat := range report.Categories {
		fmt.Fprintf(&sb, "## %s (Score: %.1f%%)\n\n", cat.Name, cat.Score*100)

		for _, f := range cat.Findings {
			status := "PASS"
			if !f.Passed {
				status = "FAIL"
			}

			fmt.Fprintf(&sb, "  [%s] %s (%s)\n", status, f.RuleTitle, f.RuleID)
			fmt.Fprintf(&sb, "        Score: %.1f%% | Severity: %s\n", f.Score*100, f.Severity)

			if f.Reasoning != "" {
				fmt.Fprintf(&sb, "        %s\n", f.Reasoning)
			}

			if len(f.Suggestions) > 0 {
				sb.WriteString("        Suggestions:\n")
				for _, s := range f.Suggestions {
					fmt.Fprintf(&sb, "          - %s\n", s)
				}
			}

			sb.WriteString("\n")
		}
	}

	return sb.String()
}
