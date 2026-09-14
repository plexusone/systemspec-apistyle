package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-apistyle/pkg/generate"
	"github.com/plexusone/systemspec-apistyle/pkg/generate/gap"
	"github.com/plexusone/systemspec-apistyle/pkg/generate/guide"
	"github.com/plexusone/systemspec-apistyle/pkg/generate/report"
	"github.com/plexusone/systemspec-apistyle/pkg/judge"
	"github.com/plexusone/systemspec-apistyle/pkg/profile"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

var (
	generateOutput  string
	generateProfile string
	generateFile    string

	// Rubric-specific flags
	rubricMode string

	// Report-specific flags
	reportInput   string
	reportTheme   string
	reportRawJSON bool

	// Guide HTML flags
	guideHTMLTheme string

	// Gap analysis flags
	gapInput string
	gapTheme string

	// MkDocs-specific flags
	mkdocsSiteName        string
	mkdocsSiteURL         string
	mkdocsRepoURL         string
	mkdocsTheme           string
	mkdocsSplitPatterns   bool
	mkdocsNoSplitCategory bool
	mkdocsNoSearch        bool
)

var generateCmd = &cobra.Command{
	Use:   "generate <type>",
	Short: "Generate outputs from a style profile",
	Long: `Generate various outputs from an API style specification profile.

Supported types:
  guide         - Generate single-page Markdown documentation (style guide)
  guide-html    - Generate standalone HTML style guide
  gap-analysis  - Generate HTML gap analysis from lint results
  mkdocs        - Generate MkDocs multi-page documentation site
  spectral      - Generate Spectral YAML ruleset
  rubric        - Generate structured-evaluation rubric for LLM-as-Judge
  report        - Generate HTML evaluation report

Examples:
  api-style generate guide --profile azure
  api-style generate guide-html --profile default --output guide.html
  api-style generate gap-analysis --input lint.json --profile default --output gap.html
  api-style generate mkdocs --profile zalando --output ./docs
  api-style generate spectral --profile zalando --output .spectral.yaml
  api-style generate rubric --profile zalando --output zalando.rubric.json`,
}

var generateGuideCmd = &cobra.Command{
	Use:   "guide",
	Short: "Generate single-page Markdown style guide",
	Long: `Generate a Markdown documentation file from a style profile.

The generated guide includes:
- Table of contents
- Introduction and design principles
- Design patterns with examples
- Rule categories and descriptions
- Good and bad examples
- Conformance level requirements
- Glossary

Examples:
  api-style generate guide
  api-style generate guide --profile azure
  api-style generate guide --file examples/zalando-rest.api-style.json --output zalando.md`,
	RunE: runGenerateGuide,
}

var generateMkDocsCmd = &cobra.Command{
	Use:   "mkdocs",
	Short: "Generate MkDocs multi-page documentation site",
	Long: `Generate a complete MkDocs documentation site from a style profile.

The generated site includes:
- mkdocs.yml configuration
- index.md with overview
- Separate pages for principles, patterns, rules (by category), glossary
- Material theme with search, dark mode, and code highlighting

Examples:
  api-style generate mkdocs --profile zalando --output ./docs
  api-style generate mkdocs --file examples/azure.api-style.json --output ./azure-docs
  api-style generate mkdocs --profile azure --site-name "My API Guidelines"`,
	RunE: runGenerateMkDocs,
}

var generateSpectralCmd = &cobra.Command{
	Use:   "spectral",
	Short: "Generate Spectral ruleset",
	Long: `Generate a Spectral-compatible YAML ruleset from a style profile.

The generated ruleset can be used with Spectral or vacuum for linting.
Only rules with Spectral enforcement configuration are included.

Examples:
  api-style generate spectral
  api-style generate spectral --profile azure
  api-style generate spectral --profile default --output .spectral.yaml`,
	RunE: runGenerateSpectral,
}

var generateRubricCmd = &cobra.Command{
	Use:   "rubric",
	Short: "Generate structured-evaluation rubric",
	Long: `Generate a structured rubric from a style profile.

Supported modes:
  evaluation (default) - For LLM-as-Judge evaluation of existing specs
  generation           - For AI agents generating new specs

Evaluation mode includes:
- Categories derived from style profile categories
- Pass/partial/fail criteria from rule judge configurations
- Few-shot examples for calibration
- Evaluation prompts for each category

Generation mode includes:
- Directives ordered by generation priority
- Grouped by generation phase (info → paths → schemas)
- Templates and checklists for guidance
- Examples of good patterns

Examples:
  api-style generate rubric
  api-style generate rubric --profile azure
  api-style generate rubric --mode generation --profile default
  api-style generate rubric --file examples/zalando-rest.api-style.json --output zalando.rubric.json`,
	RunE: runGenerateRubric,
}

var generateReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate HTML evaluation report",
	Long: `Generate an HTML report from a profile evaluation JSON file.

Accepts both structured-evaluation JSON and the JSON emitted by
"api-style score-profile --format json".

The report includes:
- Executive summary with pass/fail decision
- Category scores with color-coded indicators
- Detailed findings with recommendations
- Recommended next steps

The generated HTML can be:
- Viewed directly in a browser
- Converted to PDF using a headless browser
- Printed with proper page breaks

Examples:
  api-style generate report --input evaluation.json
  api-style generate report --input evaluation.json --output report.html
  api-style generate report --input evaluation.json --theme dark
  api-style generate report --input evaluation.json --include-json
  api-style score-profile default --format json -o scores.json && \
    api-style generate report --input scores.json -o report.html`,
	RunE: runGenerateReport,
}

var generateGuideHTMLCmd = &cobra.Command{
	Use:   "guide-html",
	Short: "Generate standalone HTML style guide",
	Long: `Generate a standalone HTML document from a style profile.

The generated HTML includes:
- Sticky table of contents
- Design principles and conformance levels
- Rule categories with severity badges
- Collapsible rule details with examples
- Design patterns and glossary
- Light and dark theme support

Examples:
  api-style generate guide-html
  api-style generate guide-html --profile azure
  api-style generate guide-html --profile default --theme dark --output guide.html`,
	RunE: runGenerateGuideHTML,
}

var generateGapAnalysisCmd = &cobra.Command{
	Use:   "gap-analysis",
	Short: "Generate HTML gap analysis report",
	Long: `Generate an HTML gap analysis report from lint results.

The report includes:
- Severity distribution with visual bars
- Category coverage heatmap (when profile provided)
- Most violated rules ranked by count
- Violations grouped by category with details
- Per-file results (for multi-file reports)
- Uncovered areas and improvement opportunities

Accepts both single-file (LintReport) and multi-file (MultiLintReport) JSON.

Examples:
  api-style generate gap-analysis --input lint.json
  api-style generate gap-analysis --input lint.json --profile default --output gap.html
  api-style generate gap-analysis --input lint.json --theme dark`,
	RunE: runGenerateGapAnalysis,
}

func init() {
	generateCmd.PersistentFlags().StringVarP(&generateOutput, "output", "o", "", "Output file/directory (default: stdout for guide/spectral, ./docs for mkdocs)")
	generateCmd.PersistentFlags().StringVarP(&generateProfile, "profile", "p", "", "Style profile name to use (built-in or from search paths)")
	generateCmd.PersistentFlags().StringVarP(&generateFile, "file", "f", "", "Path to style profile JSON/YAML file")

	// MkDocs-specific flags
	generateMkDocsCmd.Flags().StringVar(&mkdocsSiteName, "site-name", "", "MkDocs site name (default: from profile)")
	generateMkDocsCmd.Flags().StringVar(&mkdocsSiteURL, "site-url", "", "MkDocs site URL")
	generateMkDocsCmd.Flags().StringVar(&mkdocsRepoURL, "repo-url", "", "Repository URL (default: from profile)")
	generateMkDocsCmd.Flags().StringVar(&mkdocsTheme, "theme", "material", "MkDocs theme")
	generateMkDocsCmd.Flags().BoolVar(&mkdocsSplitPatterns, "split-patterns", false, "Create separate pages per pattern")
	generateMkDocsCmd.Flags().BoolVar(&mkdocsNoSplitCategory, "no-split-categories", false, "Keep all rules in one page")
	generateMkDocsCmd.Flags().BoolVar(&mkdocsNoSearch, "no-search", false, "Disable search plugin")

	// Rubric-specific flags
	generateRubricCmd.Flags().StringVarP(&rubricMode, "mode", "m", "evaluation", "Rubric mode: evaluation, generation")

	// Report-specific flags
	generateReportCmd.Flags().StringVarP(&reportInput, "input", "i", "", "Path to evaluation JSON file (required)")
	generateReportCmd.Flags().StringVar(&reportTheme, "theme", "light", "Color theme: light, dark")
	generateReportCmd.Flags().BoolVar(&reportRawJSON, "include-json", false, "Include raw JSON data in report")
	_ = generateReportCmd.MarkFlagRequired("input")

	// Guide HTML flags
	generateGuideHTMLCmd.Flags().StringVar(&guideHTMLTheme, "theme", "light", "Color theme: light, dark")

	// Gap analysis flags
	generateGapAnalysisCmd.Flags().StringVarP(&gapInput, "input", "i", "", "Path to lint JSON file (required)")
	generateGapAnalysisCmd.Flags().StringVar(&gapTheme, "theme", "light", "Color theme: light, dark")
	_ = generateGapAnalysisCmd.MarkFlagRequired("input")

	generateCmd.AddCommand(generateGuideCmd)
	generateCmd.AddCommand(generateGuideHTMLCmd)
	generateCmd.AddCommand(generateMkDocsCmd)
	generateCmd.AddCommand(generateSpectralCmd)
	generateCmd.AddCommand(generateRubricCmd)
	generateCmd.AddCommand(generateReportCmd)
	generateCmd.AddCommand(generateGapAnalysisCmd)
}

func loadProfile() (*types.APIStyleSpec, error) {
	// Prefer --file over --profile
	if generateFile != "" {
		spec, err := profile.LoadFile(generateFile)
		if err != nil {
			return nil, fmt.Errorf("loading file %q: %w", generateFile, err)
		}
		return spec, nil
	}

	// Use profile name (default to "default" if neither specified)
	profileName := generateProfile
	if profileName == "" {
		profileName = "default"
	}

	spec, err := profile.Load(profileName)
	if err != nil {
		return nil, fmt.Errorf("loading profile %q: %w", profileName, err)
	}
	return spec, nil
}

func runGenerateGuide(_ *cobra.Command, _ []string) error {
	// Load profile
	spec, err := loadProfile()
	if err != nil {
		return err
	}

	// Generate Markdown
	opts := generate.DefaultMarkdownOptions()
	md, err := generate.Markdown(spec, opts)
	if err != nil {
		return fmt.Errorf("generating markdown: %w", err)
	}

	// Write output
	return writeOutput(md)
}

func runGenerateMkDocs(_ *cobra.Command, _ []string) error {
	// Load profile
	spec, err := loadProfile()
	if err != nil {
		return err
	}

	// Configure options
	opts := generate.DefaultMkDocsOptions()
	if mkdocsSiteName != "" {
		opts.SiteName = mkdocsSiteName
	}
	if mkdocsSiteURL != "" {
		opts.SiteURL = mkdocsSiteURL
	}
	if mkdocsRepoURL != "" {
		opts.RepoURL = mkdocsRepoURL
	}
	if mkdocsTheme != "" {
		opts.Theme = mkdocsTheme
	}
	opts.SplitPatterns = mkdocsSplitPatterns
	opts.SplitCategories = !mkdocsNoSplitCategory
	opts.IncludeSearch = !mkdocsNoSearch

	// Generate MkDocs site
	result, err := generate.MkDocs(spec, opts)
	if err != nil {
		return fmt.Errorf("generating mkdocs: %w", err)
	}

	// Determine output directory
	outputDir := generateOutput
	if outputDir == "" {
		outputDir = "./docs"
	}

	// Write to filesystem
	if err := generate.WriteMkDocs(result, outputDir); err != nil {
		return fmt.Errorf("writing mkdocs: %w", err)
	}

	fmt.Printf("Generated MkDocs site: %s\n", outputDir)
	fmt.Printf("  - %d pages created\n", len(result.Pages))
	fmt.Println("\nTo build and serve:")
	fmt.Printf("  cd %s && pip install mkdocs-material && mkdocs serve\n", outputDir)

	return nil
}

func runGenerateSpectral(_ *cobra.Command, _ []string) error {
	// Load profile
	spec, err := loadProfile()
	if err != nil {
		return err
	}

	// Generate Spectral YAML
	opts := generate.DefaultSpectralOptions()
	yaml, err := generate.Spectral(spec, opts)
	if err != nil {
		return fmt.Errorf("generating spectral: %w", err)
	}

	// Write output
	return writeOutput(yaml)
}

func runGenerateRubric(_ *cobra.Command, _ []string) error {
	// Load profile
	spec, err := loadProfile()
	if err != nil {
		return err
	}

	var jsonBytes []byte

	switch strings.ToLower(rubricMode) {
	case "generation":
		// Generate generation rubric (for AI agents creating specs)
		genRubric := generate.GenerationRubricFromSpec(spec)
		jsonBytes, err = genRubric.ToJSON()
		if err != nil {
			return fmt.Errorf("serializing generation rubric: %w", err)
		}
	case "evaluation", "":
		// Generate evaluation rubric (for LLM-as-Judge)
		rubricSet := judge.GenerateRubricSet(spec)
		jsonBytes, err = judge.RubricSetToJSON(rubricSet)
		if err != nil {
			return fmt.Errorf("serializing rubric: %w", err)
		}
	default:
		return fmt.Errorf("unknown rubric mode: %s (use 'evaluation' or 'generation')", rubricMode)
	}

	// Write output
	return writeOutput(string(jsonBytes))
}

func runGenerateReport(_ *cobra.Command, _ []string) error {
	// Create report generator
	gen, err := report.New()
	if err != nil {
		return fmt.Errorf("creating report generator: %w", err)
	}

	// Configure options
	opts := &report.Options{
		Theme:          reportTheme,
		IncludeRawJSON: reportRawJSON,
	}

	// Generate HTML report from evaluation JSON
	html, err := gen.GenerateFromFile(context.Background(), reportInput, opts)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	// Write output
	if generateOutput != "" {
		if err := os.WriteFile(generateOutput, html, 0o600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Printf("Generated report: %s\n", generateOutput)
	} else {
		fmt.Print(string(html))
	}

	return nil
}

func runGenerateGuideHTML(_ *cobra.Command, _ []string) error {
	spec, err := loadProfile()
	if err != nil {
		return err
	}

	gen, err := guide.New()
	if err != nil {
		return fmt.Errorf("creating guide generator: %w", err)
	}

	opts := guide.DefaultOptions()
	opts.Theme = guideHTMLTheme

	html, err := gen.Generate(context.Background(), spec, opts)
	if err != nil {
		return fmt.Errorf("generating guide HTML: %w", err)
	}

	if generateOutput != "" {
		if err := os.WriteFile(generateOutput, html, 0o600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Printf("Generated guide: %s\n", generateOutput)
	} else {
		fmt.Print(string(html))
	}

	return nil
}

func runGenerateGapAnalysis(_ *cobra.Command, _ []string) error {
	gen, err := gap.New()
	if err != nil {
		return fmt.Errorf("creating gap analysis generator: %w", err)
	}

	opts := gap.DefaultOptions()
	opts.Theme = gapTheme

	// Optionally load profile for coverage analysis
	if generateProfile != "" || generateFile != "" {
		spec, err := loadProfile()
		if err != nil {
			return err
		}
		opts.Profile = spec
	}

	html, err := gen.GenerateFromFile(context.Background(), gapInput, opts)
	if err != nil {
		return fmt.Errorf("generating gap analysis: %w", err)
	}

	if generateOutput != "" {
		if err := os.WriteFile(generateOutput, html, 0o600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Printf("Generated gap analysis: %s\n", generateOutput)
	} else {
		fmt.Print(string(html))
	}

	return nil
}

func writeOutput(content string) error {
	// Ensure trailing newline
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if generateOutput != "" {
		if err := os.WriteFile(generateOutput, []byte(content), 0o600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Printf("Generated: %s\n", generateOutput)
	} else {
		fmt.Print(content)
	}

	return nil
}
