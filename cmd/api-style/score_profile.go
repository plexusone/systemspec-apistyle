package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-apistyle/pkg/judge"
	"github.com/plexusone/systemspec-apistyle/pkg/profile"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

var (
	scoreFormat string
	scoreOutput string
	scoreModel  string
)

var scoreProfileCmd = &cobra.Command{
	Use:   "score-profile <profile-name-or-file>",
	Short: "Score a style guide profile using LLM evaluation",
	Long: `Evaluate a style guide profile against quality criteria using LLM-as-a-Judge.

This command assesses how complete and well-structured a style guide is,
evaluating categories like content coverage, rule quality, examples, and more.

Requires ANTHROPIC_API_KEY environment variable to be set.

Examples:
  api-style score-profile default
  api-style score-profile azure --format json
  api-style score-profile ./custom-profile.json --output scores.json
  api-style score-profile zalando --model claude-sonnet-4`,
	Args: cobra.ExactArgs(1),
	RunE: runScoreProfile,
}

func init() {
	scoreProfileCmd.Flags().StringVarP(&scoreFormat, "format", "f", "text", "Output format: text, json")
	scoreProfileCmd.Flags().StringVarP(&scoreOutput, "output", "o", "", "Output file (default: stdout)")
	scoreProfileCmd.Flags().StringVarP(&scoreModel, "model", "m", "", "LLM model to use (default: claude-3-5-haiku)")
}

func runScoreProfile(_ *cobra.Command, args []string) error {
	profileArg := args[0]

	// Check for API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	// Load profile (either built-in name or file path)
	var spec *types.APIStyleSpec
	var err error

	if _, statErr := os.Stat(profileArg); statErr == nil {
		// It's a file path
		spec, err = profile.LoadFile(profileArg)
		if err != nil {
			return fmt.Errorf("loading profile file %q: %w", profileArg, err)
		}
	} else {
		// Try as built-in profile name
		spec, err = profile.Load(profileArg)
		if err != nil {
			return fmt.Errorf("loading profile %q: %w", profileArg, err)
		}
	}

	// Create provider
	provider := judge.NewAnthropicProvider("", nil)
	if scoreModel != "" {
		provider.SetDefaultModel(scoreModel)
	}

	// Create evaluator
	evaluator, err := judge.NewStyleGuideEvaluator(provider)
	if err != nil {
		return fmt.Errorf("creating evaluator: %w", err)
	}

	// Run evaluation
	ctx := context.Background()
	opts := &judge.StyleGuideEvalOptions{
		Model: scoreModel,
	}

	fmt.Fprintf(os.Stderr, "Evaluating profile %q against style guide quality rubric...\n", spec.Name)

	report, err := evaluator.Evaluate(ctx, spec, opts)
	if err != nil {
		return fmt.Errorf("evaluating profile: %w", err)
	}

	// Format output
	output, err := formatScoreReport(report, scoreFormat)
	if err != nil {
		return fmt.Errorf("formatting report: %w", err)
	}

	// Write output
	if scoreOutput != "" {
		if err := os.WriteFile(scoreOutput, []byte(output), 0o600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Report written to %s\n", scoreOutput)
	} else {
		fmt.Print(output)
	}

	return nil
}

func formatScoreReport(report *judge.StyleGuideReport, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data) + "\n", nil
	case "text":
		return formatScoreText(report), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

func formatScoreText(report *judge.StyleGuideReport) string {
	var sb strings.Builder

	sb.WriteString("Style Guide Quality Report\n")
	sb.WriteString("==========================\n\n")

	// Summary
	fmt.Fprintf(&sb, "Profile: %s\n", report.ProfileName)
	fmt.Fprintf(&sb, "Status: %s\n", strings.ToUpper(report.Status))
	fmt.Fprintf(&sb, "Overall Score: %.1f / 5.0\n", report.OverallScore)
	fmt.Fprintf(&sb, "Categories: %d passed, %d partial, %d failed\n\n",
		report.Summary.PassedCategories,
		report.Summary.PartialCategories,
		report.Summary.FailedCategories)

	// Category details
	sb.WriteString("Category Scores\n")
	sb.WriteString("---------------\n\n")

	for _, cat := range report.Categories {
		reqMarker := ""
		if cat.Required {
			reqMarker = " [required]"
		}

		fmt.Fprintf(&sb, "## %s%s\n", cat.Name, reqMarker)
		fmt.Fprintf(&sb, "   Score: %s (%d/5)\n", strings.ToUpper(cat.Score), cat.NumericScore)

		if cat.Reasoning != "" {
			fmt.Fprintf(&sb, "   %s\n", cat.Reasoning)
		}

		if len(cat.Strengths) > 0 {
			sb.WriteString("   Strengths:\n")
			for _, s := range cat.Strengths {
				fmt.Fprintf(&sb, "     + %s\n", s)
			}
		}

		if len(cat.Weaknesses) > 0 {
			sb.WriteString("   Weaknesses:\n")
			for _, w := range cat.Weaknesses {
				fmt.Fprintf(&sb, "     - %s\n", w)
			}
		}

		sb.WriteString("\n")
	}

	// Metadata
	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "Evaluated using: %s\n", report.Metadata.RubricName)
	fmt.Fprintf(&sb, "Model: %s\n", report.Metadata.Model)
	fmt.Fprintf(&sb, "Duration: %s\n", report.Metadata.Duration)

	return sb.String()
}
