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
	analyzeFormat       string
	analyzeOutput       string
	analyzeProfile      string
	analyzeLevel        string
	analyzeModel        string
	analyzeEnableEval   bool
	analyzeFailWarnings bool
	analyzeMinScore     float64
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze <openapi-spec>",
	Short: "Analyze an OpenAPI specification comprehensively",
	Long: `Analyze an OpenAPI specification using both linting and LLM evaluation.

This command combines deterministic linting (like 'api-style lint') with
AI-powered evaluation (like 'api-style evaluate') to provide comprehensive
API quality assessment.

Examples:
  api-style analyze openapi.yaml
  api-style analyze openapi.yaml --evaluate
  api-style analyze openapi.yaml --profile azure --level silver
  api-style analyze openapi.yaml --evaluate --min-score 0.8`,
	Args: cobra.ExactArgs(1),
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeFormat, "format", "f", "text", "Output format: text, json")
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "", "Output file (default: stdout)")
	analyzeCmd.Flags().StringVarP(&analyzeProfile, "profile", "p", "default", "Style profile to use")
	analyzeCmd.Flags().StringVarP(&analyzeLevel, "level", "l", "", "Target conformance level (bronze, silver, gold)")
	analyzeCmd.Flags().StringVarP(&analyzeModel, "model", "m", "", "LLM model for evaluation")
	analyzeCmd.Flags().BoolVarP(&analyzeEnableEval, "evaluate", "e", false, "Enable LLM evaluation")
	analyzeCmd.Flags().BoolVar(&analyzeFailWarnings, "fail-warnings", false, "Treat warnings as failures")
	analyzeCmd.Flags().Float64Var(&analyzeMinScore, "min-score", 0.7, "Minimum evaluation score for GO decision")
}

func runAnalyze(_ *cobra.Command, args []string) error {
	specFile := args[0]

	// Read the OpenAPI spec
	specBytes, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("reading spec file: %w", err)
	}

	// Load profile
	styleSpec, err := profile.Load(analyzeProfile)
	if err != nil {
		return fmt.Errorf("loading profile %q: %w", analyzeProfile, err)
	}

	// Create provider if evaluation is enabled
	var provider judge.Provider
	if analyzeEnableEval {
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY required for --evaluate")
		}
		p := judge.NewAnthropicProvider("", nil)
		if analyzeModel != "" {
			p.SetDefaultModel(analyzeModel)
		}
		provider = p
	}

	// Create analyzer
	analyzer := analyze.New(styleSpec, provider)

	// Run analysis
	ctx := context.Background()
	opts := &analyze.Options{
		FileName:         specFile,
		EnableLint:       true,
		EnableEvaluate:   analyzeEnableEval,
		ConformanceLevel: analyzeLevel,
		FailOnWarnings:   analyzeFailWarnings,
		MinScore:         analyzeMinScore,
	}

	if analyzeModel != "" {
		opts.EvaluateOptions = &judge.Options{
			FileName:         specFile,
			IncludeReasoning: true,
			Model:            analyzeModel,
		}
	}

	report, err := analyzer.Analyze(ctx, specBytes, opts)
	if err != nil {
		return fmt.Errorf("analyzing: %w", err)
	}

	// Format output
	output, err := formatAnalysisReport(report, analyzeFormat)
	if err != nil {
		return fmt.Errorf("formatting report: %w", err)
	}

	// Write output
	if analyzeOutput != "" {
		//nolint:gosec // G703: Path from CLI flag
		if err := os.WriteFile(analyzeOutput, []byte(output), 0o600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	// Exit with appropriate code
	switch report.Decision {
	case analyze.DecisionNoGo:
		os.Exit(1)
	case analyze.DecisionWarning:
		if analyzeFailWarnings {
			os.Exit(1)
		}
	}

	return nil
}

func formatAnalysisReport(report *analyze.AnalysisReport, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	case "text":
		return formatAnalysisText(report), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

func formatAnalysisText(report *analyze.AnalysisReport) string {
	var sb strings.Builder

	sb.WriteString("API Style Analysis Report\n")
	sb.WriteString("=========================\n\n")

	// Decision banner
	switch report.Decision {
	case analyze.DecisionGo:
		sb.WriteString("Decision: GO ✓\n")
	case analyze.DecisionWarning:
		sb.WriteString("Decision: WARNING ⚠\n")
	case analyze.DecisionNoGo:
		sb.WriteString("Decision: NO-GO ✗\n")
	}

	sb.WriteString("\n")
	sb.WriteString(report.Summary)
	sb.WriteString("\n\n")

	// Lint results
	if report.LintReport != nil {
		sb.WriteString("## Lint Results\n\n")
		fmt.Fprintf(&sb, "Status: %s\n", strings.ToUpper(string(report.LintReport.Status)))
		fmt.Fprintf(&sb, "Errors: %d, Warnings: %d, Info: %d, Hints: %d\n\n",
			report.LintReport.Summary.Errors,
			report.LintReport.Summary.Warnings,
			report.LintReport.Summary.Infos,
			report.LintReport.Summary.Hints)

		if len(report.LintReport.Violations) > 0 {
			// Show first 10 violations
			count := len(report.LintReport.Violations)
			if count > 10 {
				count = 10
			}

			sb.WriteString("Top Violations:\n")
			for _, v := range report.LintReport.Violations[:count] {
				fmt.Fprintf(&sb, "  [%s] %s: %s\n", v.Severity, v.RuleID, v.Message)
				fmt.Fprintf(&sb, "        at %s\n", v.Path)
			}

			if len(report.LintReport.Violations) > 10 {
				fmt.Fprintf(&sb, "\n  ... and %d more violations\n",
					len(report.LintReport.Violations)-10)
			}
			sb.WriteString("\n")
		}
	}

	// Evaluation results
	if report.EvaluationReport != nil {
		sb.WriteString("## Evaluation Results\n\n")
		fmt.Fprintf(&sb, "Overall Score: %.1f%%\n", report.EvaluationReport.Summary.OverallScore*100)
		fmt.Fprintf(&sb, "Rules Evaluated: %d (Passed: %d, Failed: %d)\n\n",
			report.EvaluationReport.Summary.TotalRules,
			report.EvaluationReport.Summary.PassedRules,
			report.EvaluationReport.Summary.FailedRules)

		if len(report.EvaluationReport.Summary.CategoryScores) > 0 {
			sb.WriteString("Category Scores:\n")
			for cat, score := range report.EvaluationReport.Summary.CategoryScores {
				fmt.Fprintf(&sb, "  %s: %.1f%%\n", cat, score*100)
			}
			sb.WriteString("\n")
		}
	}

	// Conformance level
	if report.ConformanceLevel != "" {
		fmt.Fprintf(&sb, "## Conformance Level: %s\n\n", strings.ToUpper(report.ConformanceLevel))
	}

	// Metadata
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "Profile: %s | Duration: %s | %s\n",
		report.Metadata.ProfileName,
		report.Metadata.Duration,
		report.Metadata.Timestamp)

	return sb.String()
}
