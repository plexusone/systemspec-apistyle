package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-apistyle/pkg/profile"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

var (
	patternProfile  string
	patternCategory string
)

var patternCmd = &cobra.Command{
	Use:   "pattern <command>",
	Short: "View API design patterns",
	Long: `View and explore API design patterns from style profiles.

Patterns are reusable solutions to common API design problems.
Each pattern includes problem/solution descriptions, examples,
and references to related rules.

Examples:
  api-style pattern list
  api-style pattern list --profile default --category pagination
  api-style pattern show cursor-pagination`,
}

var patternListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available patterns",
	Long: `List all patterns defined in a style profile.

Use --category to filter by pattern category.

Examples:
  api-style pattern list
  api-style pattern list --profile zalando
  api-style pattern list --category errors`,
	RunE: runPatternList,
}

var patternShowCmd = &cobra.Command{
	Use:   "show <pattern-id>",
	Short: "Show pattern details",
	Long: `Display detailed information about a specific pattern.

Includes problem/solution, examples, and related rules.

Examples:
  api-style pattern show cursor-pagination
  api-style pattern show rfc9457-errors --profile default`,
	Args: cobra.ExactArgs(1),
	RunE: runPatternShow,
}

func init() {
	patternListCmd.Flags().StringVarP(&patternProfile, "profile", "p", "default", "Style profile name")
	patternListCmd.Flags().StringVarP(&patternCategory, "category", "c", "", "Filter by category")

	patternShowCmd.Flags().StringVarP(&patternProfile, "profile", "p", "default", "Style profile name")

	patternCmd.AddCommand(patternListCmd)
	patternCmd.AddCommand(patternShowCmd)
}

func runPatternList(_ *cobra.Command, _ []string) error {
	spec, err := profile.Load(patternProfile)
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	patterns := spec.Patterns
	if len(patterns) == 0 {
		fmt.Printf("No patterns defined in profile %q\n", patternProfile)
		return nil
	}

	// Filter by category if specified
	if patternCategory != "" {
		var filtered []types.Pattern
		for _, p := range patterns {
			if strings.EqualFold(p.Category, patternCategory) {
				filtered = append(filtered, p)
			}
		}
		patterns = filtered
	}

	if len(patterns) == 0 {
		fmt.Printf("No patterns found for category %q\n", patternCategory)
		return nil
	}

	fmt.Printf("Patterns in %s:\n\n", patternProfile)

	// Group by category
	byCategory := make(map[string][]types.Pattern)
	for _, p := range patterns {
		cat := p.Category
		if cat == "" {
			cat = "general"
		}
		byCategory[cat] = append(byCategory[cat], p)
	}

	for cat, pats := range byCategory {
		fmt.Printf("[%s]\n", cat)
		for _, p := range pats {
			summary := p.Summary
			if len(summary) > 60 {
				summary = summary[:57] + "..."
			}
			fmt.Printf("  %-24s %s\n", p.ID, summary)
		}
		fmt.Println()
	}

	return nil
}

func runPatternShow(_ *cobra.Command, args []string) error {
	patternID := args[0]

	spec, err := profile.Load(patternProfile)
	if err != nil {
		return fmt.Errorf("loading profile: %w", err)
	}

	pattern := spec.GetPattern(patternID)
	if pattern == nil {
		return fmt.Errorf("pattern %q not found in profile %q", patternID, patternProfile)
	}

	// Print pattern details
	fmt.Printf("# %s\n\n", pattern.Name)

	if pattern.Summary != "" {
		fmt.Printf("%s\n\n", pattern.Summary)
	}

	if pattern.Category != "" {
		fmt.Printf("Category: %s\n\n", pattern.Category)
	}

	if pattern.Problem != "" {
		fmt.Println("## Problem")
		fmt.Printf("%s\n\n", pattern.Problem)
	}

	if pattern.Solution != "" {
		fmt.Println("## Solution")
		fmt.Printf("%s\n\n", pattern.Solution)
	}

	if pattern.When != "" {
		fmt.Println("## When to Use")
		fmt.Printf("%s\n\n", pattern.When)
	}

	if len(pattern.Examples) > 0 {
		fmt.Println("## Examples")
		for _, ex := range pattern.Examples {
			if ex.Title != "" {
				fmt.Printf("\n### %s\n", ex.Title)
			}
			if ex.Code != "" {
				fmt.Println("```")
				fmt.Println(ex.Code)
				fmt.Println("```")
			}
			if ex.Description != "" {
				fmt.Printf("%s\n", ex.Description)
			}
		}
		fmt.Println()
	}

	if len(pattern.RelatedRules) > 0 {
		fmt.Println("## Related Rules")
		fmt.Printf("%s\n\n", strings.Join(pattern.RelatedRules, ", "))
	}

	if len(pattern.RelatedPatterns) > 0 {
		fmt.Println("## Related Patterns")
		fmt.Printf("%s\n\n", strings.Join(pattern.RelatedPatterns, ", "))
	}

	if len(pattern.References) > 0 {
		fmt.Println("## References")
		for _, ref := range pattern.References {
			fmt.Printf("- [%s](%s)\n", ref.Title, ref.URL)
		}
		fmt.Println()
	}

	return nil
}
