package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-apistyle/pkg/config"
	"github.com/plexusone/systemspec-apistyle/pkg/files"
	"github.com/plexusone/systemspec-apistyle/pkg/fix"
	"github.com/plexusone/systemspec-apistyle/pkg/lint"
	"github.com/plexusone/systemspec-apistyle/pkg/profile"
	"github.com/plexusone/systemspec-apistyle/pkg/sarif"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
	"github.com/plexusone/systemspec-apistyle/pkg/watch"
)

var (
	lintFormat       string
	lintOutput       string
	lintProfile      string
	lintLevel        string
	lintConfig       string
	lintRecursive    bool
	lintWatch        bool
	lintSuggestFixes bool
)

var lintCmd = &cobra.Command{
	Use:   "lint <openapi-spec> [<openapi-spec>...]",
	Short: "Lint an OpenAPI specification",
	Long: `Lint an OpenAPI specification against style rules.

Uses vacuum (a fast, Spectral-compatible linter) to check the specification
against the configured style rules.

Multiple files can be specified using glob patterns or by passing multiple
arguments. Use --recursive to search directories recursively.

Configuration can be provided via .api-style.yaml in the current directory
or explicitly with --config.

Examples:
  api-style lint openapi.yaml
  api-style lint openapi.yaml --format json
  api-style lint openapi.yaml --format json --output report.json
  api-style lint openapi.yaml --profile azure --level silver
  api-style lint api/*.yaml --recursive
  api-style lint . --recursive
  api-style lint openapi.yaml --watch
  api-style lint --config .api-style.yaml openapi.yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLint,
}

func init() {
	lintCmd.Flags().StringVarP(&lintFormat, "format", "f", "text", "Output format: text, json, sarif")
	lintCmd.Flags().StringVarP(&lintOutput, "output", "o", "", "Output file (default: stdout)")
	lintCmd.Flags().StringVarP(&lintProfile, "profile", "p", "", "Style profile to use (default from config or 'default')")
	lintCmd.Flags().StringVarP(&lintLevel, "level", "l", "", "Target conformance level (bronze, silver, gold)")
	lintCmd.Flags().StringVarP(&lintConfig, "config", "c", "", "Config file path (default: .api-style.yaml)")
	lintCmd.Flags().BoolVarP(&lintRecursive, "recursive", "r", false, "Search directories recursively")
	lintCmd.Flags().BoolVarP(&lintWatch, "watch", "w", false, "Watch files for changes and re-lint")
	lintCmd.Flags().BoolVar(&lintSuggestFixes, "suggest-fixes", false, "Include fix suggestions in output")
}

func runLint(_ *cobra.Command, args []string) error {
	// Load configuration
	var cfg *config.Config
	var err error
	if lintConfig != "" {
		cfg, err = config.LoadFile(lintConfig)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	} else {
		cfg, err = config.LoadOrDefault()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	}

	// Merge CLI flags into config (CLI takes precedence)
	cfg.Merge(lintProfile, lintLevel, nil, nil)

	// Resolve spec files
	specFiles, err := files.ResolveSpecs(args, lintRecursive, cfg.Include, cfg.Exclude)
	if err != nil {
		return fmt.Errorf("resolving files: %w", err)
	}

	if len(specFiles) == 0 {
		return fmt.Errorf("no OpenAPI spec files found")
	}

	// Watch mode
	if lintWatch {
		return runWatchMode(specFiles, cfg)
	}

	// Lint all files
	multiReport, styleSpec := lintFiles(specFiles, cfg)

	// Generate fix suggestions if requested
	var fixReport *types.FixReport
	if lintSuggestFixes && styleSpec != nil {
		fixReport = generateFixSuggestions(context.Background(), multiReport, styleSpec)
	}

	// Format and output
	if err := outputResults(multiReport, styleSpec, fixReport); err != nil {
		return err
	}

	// Exit with code 1 if there are blocking violations
	if multiReport.HasBlockingViolations() {
		os.Exit(1)
	}

	return nil
}

func lintFiles(specFiles []string, cfg *config.Config) (*types.MultiLintReport, *types.APIStyleSpec) {
	ctx := context.Background()
	multiReport := types.NewMultiLintReport()

	var styleSpec *types.APIStyleSpec

	// Load profile once for all files
	profileName := cfg.Profile
	if profileName == "" {
		profileName = "default"
	}

	if profileName != "" && profileName != "vacuum" {
		loadedSpec, loadErr := profile.Load(profileName)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: profile %q not found, using vacuum defaults\n", profileName)
		} else {
			styleSpec = loadedSpec
		}
	}

	// Lint each file
	for _, specFile := range specFiles {
		report, err := lintSingleFile(ctx, specFile, cfg, styleSpec)
		if err != nil {
			// Add error as a violation
			report = types.NewLintReport()
			report.Status = types.StatusFail
			report.AddViolation(types.Violation{
				RuleID:   "lint-error",
				Severity: types.SeverityError,
				Message:  fmt.Sprintf("Failed to lint: %v", err),
				Path:     specFile,
			})
		}

		multiReport.AddFileReport(specFile, report)
	}

	return multiReport, styleSpec
}

func lintSingleFile(ctx context.Context, specFile string, cfg *config.Config, styleSpec *types.APIStyleSpec) (*types.LintReport, error) {
	specBytes, err := os.ReadFile(specFile)
	if err != nil {
		return nil, fmt.Errorf("reading spec file: %w", err)
	}

	opts := &lint.Options{
		FileName:         specFile,
		Profile:          cfg.Profile,
		ConformanceLevel: cfg.Level,
		Exceptions:       cfg.ToExceptions(),
	}

	var report *types.LintReport

	if styleSpec != nil {
		linter := lint.NewVacuumLinter(styleSpec)
		report, err = linter.Lint(ctx, specBytes, opts)
	} else {
		report, err = lint.WithDefaults(ctx, specBytes, opts)
	}

	if err != nil {
		return nil, fmt.Errorf("linting: %w", err)
	}

	return report, nil
}

func runWatchMode(specFiles []string, cfg *config.Config) error {
	// Create watch config
	watchCfg := &watch.Config{
		Paths:    specFiles,
		Debounce: 200 * time.Millisecond,
		Include:  []string{"*.yaml", "*.yml", "*.json"},
	}

	watcher, err := watch.New(watchCfg)
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer func() {
		if closeErr := watcher.Close(); closeErr != nil {
			slog.Error("failed to close watcher", "error", closeErr)
		}
	}()

	// Load style spec once
	var styleSpec *types.APIStyleSpec
	profileName := cfg.Profile
	if profileName == "" {
		profileName = "default"
	}
	if profileName != "" && profileName != "vacuum" {
		if loadedSpec, loadErr := profile.Load(profileName); loadErr == nil {
			styleSpec = loadedSpec
		}
	}

	// Set up change callback
	watcher.OnChange(func(event watch.Event) error {
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("\n[%s] File changed: %s\n", timestamp, event.Path)

		ctx := context.Background()
		report, err := lintSingleFile(ctx, event.Path, cfg, styleSpec)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			return nil // Don't stop watching on error
		}

		// Print summary
		printWatchSummary(event.Path, report)
		return nil
	})

	// Initial lint
	fmt.Println("Starting watch mode...")
	fmt.Printf("Watching %d file(s). Press Ctrl+C to stop.\n\n", len(specFiles))

	for _, f := range specFiles {
		ctx := context.Background()
		report, err := lintSingleFile(ctx, f, cfg, styleSpec)
		if err != nil {
			fmt.Printf("%s: Error - %v\n", f, err)
			continue
		}
		printWatchSummary(f, report)
	}

	// Watch with graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	fmt.Println("\nWatching for changes...")

	if err := watcher.Watch(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("watching: %w", err)
	}

	fmt.Println("\nWatch mode stopped.")
	return nil
}

func printWatchSummary(file string, report *types.LintReport) {
	status := "PASS"
	if report.Status == types.StatusFail {
		status = "FAIL"
	}

	fmt.Printf("%s: %s (errors: %d, warnings: %d)\n",
		file, status,
		report.Summary.Errors,
		report.Summary.Warnings)

	// Print first few errors
	errorCount := 0
	for _, v := range report.Violations {
		if v.Severity == types.SeverityError && errorCount < 5 {
			fmt.Printf("  - [%s] %s\n", v.RuleID, v.Message)
			errorCount++
		}
	}
	if report.Summary.Errors > 5 {
		fmt.Printf("  ... and %d more errors\n", report.Summary.Errors-5)
	}
}

func generateFixSuggestions(ctx context.Context, multiReport *types.MultiLintReport, styleSpec *types.APIStyleSpec) *types.FixReport {
	// Collect all violations
	var allViolations []types.Violation
	for _, fr := range multiReport.FileReports {
		allViolations = append(allViolations, fr.Violations...)
	}

	if len(allViolations) == 0 {
		return nil
	}

	fixer := fix.NewRuleFixer(styleSpec)
	opts := &fix.Options{
		MaxSuggestions: 50,
		IncludePatch:   false,
	}

	report, err := fixer.SuggestFixes(ctx, nil, allViolations, opts)
	if err != nil {
		return nil
	}

	return report
}

func outputResults(multiReport *types.MultiLintReport, styleSpec *types.APIStyleSpec, fixReport *types.FixReport) error {
	var output string
	var err error

	// For single file, output the single report for backwards compatibility
	if len(multiReport.FileReports) == 1 {
		report := multiReport.FileReports[0].LintReport
		output, err = formatReport(report, styleSpec, lintFormat)
	} else {
		output, err = formatMultiReport(multiReport, styleSpec, lintFormat)
	}

	if err != nil {
		return fmt.Errorf("formatting report: %w", err)
	}

	// Append fix suggestions if available
	if fixReport != nil && fixReport.FixedCount > 0 {
		fixOutput := formatFixSuggestions(fixReport, lintFormat)
		output += fixOutput
	}

	// Write output
	if lintOutput != "" {
		//nolint:gosec // G703: Path from CLI flag
		if err := os.WriteFile(lintOutput, []byte(output), 0o600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	return nil
}

func formatFixSuggestions(report *types.FixReport, format string) string {
	switch strings.ToLower(format) {
	case "json":
		// For JSON, return a separate fixSuggestions object
		data, err := json.MarshalIndent(map[string]any{
			"fixSuggestions": report.Suggestions,
			"fixedCount":     report.FixedCount,
			"unfixedCount":   report.UnfixedCount,
			"unfixedRules":   report.UnfixedRules,
		}, "", "  ")
		if err != nil {
			return ""
		}
		return "\n" + string(data) + "\n"
	case "text":
		var sb strings.Builder
		sb.WriteString("\nFix Suggestions\n")
		sb.WriteString("---------------\n\n")
		fmt.Fprintf(&sb, "Violations with fixes: %d, without fixes: %d\n\n",
			report.FixedCount, report.UnfixedCount)

		for i, s := range report.Suggestions {
			fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, s.RuleID, s.Path)
			if s.CurrentValue != "" {
				fmt.Fprintf(&sb, "   Current:   %s\n", s.CurrentValue)
			}
			if s.SuggestedValue != "" {
				fmt.Fprintf(&sb, "   Suggested: %s\n", s.SuggestedValue)
			}
			if s.Reasoning != "" {
				fmt.Fprintf(&sb, "   Reason:    %s\n", s.Reasoning)
			}
			sb.WriteString("\n")
		}

		if len(report.UnfixedRules) > 0 {
			sb.WriteString("Rules without automatic fixes:\n")
			for _, r := range report.UnfixedRules {
				fmt.Fprintf(&sb, "  - %s\n", r)
			}
		}

		return sb.String()
	default:
		return ""
	}
}

func formatReport(report *types.LintReport, styleSpec *types.APIStyleSpec, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return formatJSON(report)
	case "sarif":
		return formatSARIF(report, styleSpec)
	case "text":
		return formatText(report), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

func formatMultiReport(report *types.MultiLintReport, styleSpec *types.APIStyleSpec, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return formatMultiJSON(report)
	case "sarif":
		return formatMultiSARIF(report, styleSpec)
	case "text":
		return formatMultiText(report), nil
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}

func formatJSON(report *types.LintReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

func formatMultiJSON(report *types.MultiLintReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

func formatText(report *types.LintReport) string {
	var sb strings.Builder

	// Header
	sb.WriteString("API Style Lint Report\n")
	sb.WriteString("=====================\n\n")

	// Summary
	fmt.Fprintf(&sb, "Status: %s\n", strings.ToUpper(string(report.Status)))
	fmt.Fprintf(&sb, "Errors: %d, Warnings: %d, Info: %d, Hints: %d\n\n",
		report.Summary.Errors,
		report.Summary.Warnings,
		report.Summary.Infos,
		report.Summary.Hints,
	)

	if len(report.Violations) == 0 {
		sb.WriteString("No violations found.\n")
		return sb.String()
	}

	// Group by severity
	byLevel := map[types.Severity][]types.Violation{
		types.SeverityError: {},
		types.SeverityWarn:  {},
		types.SeverityInfo:  {},
		types.SeverityHint:  {},
	}

	for _, v := range report.Violations {
		byLevel[v.Severity] = append(byLevel[v.Severity], v)
	}

	// Print violations by severity
	printViolations(&sb, "Errors", byLevel[types.SeverityError])
	printViolations(&sb, "Warnings", byLevel[types.SeverityWarn])
	printViolations(&sb, "Info", byLevel[types.SeverityInfo])
	printViolations(&sb, "Hints", byLevel[types.SeverityHint])

	return sb.String()
}

func formatMultiText(report *types.MultiLintReport) string {
	var sb strings.Builder

	// Header
	sb.WriteString("API Style Lint Report\n")
	sb.WriteString("=====================\n\n")

	// Overall Summary
	fmt.Fprintf(&sb, "Files: %d (%d failed)\n", report.FileCount(), report.FailedFileCount())
	fmt.Fprintf(&sb, "Status: %s\n", strings.ToUpper(string(report.Status)))
	fmt.Fprintf(&sb, "Total Errors: %d, Warnings: %d, Info: %d, Hints: %d\n\n",
		report.Summary.Errors,
		report.Summary.Warnings,
		report.Summary.Infos,
		report.Summary.Hints,
	)

	// Per-file results
	for _, fr := range report.FileReports {
		fmt.Fprintf(&sb, "--- %s ---\n", fr.File)
		fmt.Fprintf(&sb, "Status: %s (errors: %d, warnings: %d)\n",
			strings.ToUpper(string(fr.Status)),
			fr.Summary.Errors,
			fr.Summary.Warnings)

		if len(fr.Violations) == 0 {
			sb.WriteString("No violations.\n\n")
			continue
		}

		// Print violations (limited to first 10 per file for readability)
		count := 0
		for _, v := range fr.Violations {
			if count >= 10 {
				fmt.Fprintf(&sb, "  ... and %d more violations\n", len(fr.Violations)-10)
				break
			}
			location := v.Path
			if v.Line > 0 {
				location = fmt.Sprintf("%s (line %d)", v.Path, v.Line)
			}
			fmt.Fprintf(&sb, "  [%s] %s: %s\n    %s\n", v.Severity, v.RuleID, v.Message, location)
			// Show key metadata in multi-file mode
			if v.RuleTitle != "" {
				fmt.Fprintf(&sb, "    Rule: %s\n", v.RuleTitle)
			}
			if v.FixPriority > 0 {
				fmt.Fprintf(&sb, "    Fix Priority: %d\n", v.FixPriority)
			}
			count++
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func printViolations(sb *strings.Builder, level string, violations []types.Violation) {
	if len(violations) == 0 {
		return
	}

	fmt.Fprintf(sb, "%s:\n", level)
	for _, v := range violations {
		location := v.Path
		if v.Line > 0 {
			location = fmt.Sprintf("%s (line %d)", v.Path, v.Line)
		}
		fmt.Fprintf(sb, "  - [%s] %s\n    %s\n", v.RuleID, v.Message, location)

		// Print additional metadata when available
		if v.RuleTitle != "" {
			fmt.Fprintf(sb, "    Rule: %s\n", v.RuleTitle)
		}
		if v.Category != "" {
			fmt.Fprintf(sb, "    Category: %s\n", v.Category)
		}
		if v.FixPriority > 0 {
			fmt.Fprintf(sb, "    Fix Priority: %d\n", v.FixPriority)
		}
		if len(v.RelatedRules) > 0 {
			fmt.Fprintf(sb, "    Related: %s\n", strings.Join(v.RelatedRules, ", "))
		}
		if v.ExampleFix != "" {
			fmt.Fprintf(sb, "    Example Fix:\n")
			for _, line := range strings.Split(v.ExampleFix, "\n") {
				fmt.Fprintf(sb, "      %s\n", line)
			}
		}
		if v.RuleURL != "" {
			fmt.Fprintf(sb, "    Docs: %s\n", v.RuleURL)
		}
	}
	sb.WriteString("\n")
}

// formatSARIF outputs in SARIF format for IDE integration
func formatSARIF(report *types.LintReport, styleSpec *types.APIStyleSpec) (string, error) {
	opts := sarif.DefaultOptions()

	// Add rule metadata if we have a style spec
	if styleSpec != nil {
		opts.Rules = make(map[string]*types.Rule)
		for i := range styleSpec.Rules {
			rule := &styleSpec.Rules[i]
			opts.Rules[rule.ID] = rule
		}
	}

	return sarif.FormatLintReport(report, opts)
}

func formatMultiSARIF(report *types.MultiLintReport, styleSpec *types.APIStyleSpec) (string, error) {
	// For SARIF, combine all file reports into one
	// This is a simplified approach - a full implementation would create
	// separate runs for each file
	combined := types.NewLintReport()
	for _, fr := range report.FileReports {
		for _, v := range fr.Violations {
			// Prefix path with file for clarity
			v.Path = fr.File + ":" + v.Path
			combined.AddViolation(v)
		}
	}

	return formatSARIF(combined, styleSpec)
}
