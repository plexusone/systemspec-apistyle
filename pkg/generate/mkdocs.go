package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// MkDocsOptions configures MkDocs site generation.
type MkDocsOptions struct {
	// SiteName is the MkDocs site name.
	SiteName string

	// SiteURL is the base URL for the site.
	SiteURL string

	// RepoURL is the source repository URL.
	RepoURL string

	// Theme is the MkDocs theme (default: material).
	Theme string

	// IncludeSearch enables search functionality.
	IncludeSearch bool

	// SplitCategories creates separate pages per category.
	SplitCategories bool

	// SplitPatterns creates separate pages per pattern.
	SplitPatterns bool

	// MarkdownOptions for generating individual pages.
	MarkdownOptions *MarkdownOptions
}

// DefaultMkDocsOptions returns default MkDocs configuration.
func DefaultMkDocsOptions() *MkDocsOptions {
	return &MkDocsOptions{
		Theme:           "material",
		IncludeSearch:   true,
		SplitCategories: true,
		SplitPatterns:   false,
		MarkdownOptions: DefaultMarkdownOptions(),
	}
}

// MkDocsResult contains the generated MkDocs site files.
type MkDocsResult struct {
	// Config is the mkdocs.yml content.
	Config string

	// Pages maps file paths (relative to docs/) to content.
	Pages map[string]string
}

// MkDocs generates a complete MkDocs site structure from an APIStyleSpec.
func MkDocs(spec *types.APIStyleSpec, opts *MkDocsOptions) (*MkDocsResult, error) {
	if opts == nil {
		opts = DefaultMkDocsOptions()
	}
	if opts.MarkdownOptions == nil {
		opts.MarkdownOptions = DefaultMarkdownOptions()
	}
	if opts.SiteName == "" {
		opts.SiteName = spec.Name
	}

	result := &MkDocsResult{
		Pages: make(map[string]string),
	}

	// Generate index page
	result.Pages["index.md"] = generateIndexPage(spec, opts)

	// Generate introduction page (if exists)
	if spec.Introduction != "" {
		result.Pages["introduction.md"] = generateIntroductionPage(spec, opts)
	}

	// Generate principles page
	if len(spec.Principles) > 0 {
		result.Pages["principles.md"] = generatePrinciplesPage(spec, opts)
	}

	// Generate patterns pages
	if len(spec.Patterns) > 0 {
		if opts.SplitPatterns {
			for _, p := range spec.Patterns {
				filename := fmt.Sprintf("patterns/%s.md", sanitizeFilename(p.ID))
				result.Pages[filename] = generatePatternPage(p, opts)
			}
			result.Pages["patterns/index.md"] = generatePatternsIndexPage(spec, opts)
		} else {
			result.Pages["patterns.md"] = generatePatternsPage(spec, opts)
		}
	}

	// Generate conformance levels page
	if len(spec.ConformanceLevels) > 0 {
		result.Pages["conformance.md"] = generateConformancePage(spec, opts)
	}

	// Generate rules pages
	if opts.SplitCategories {
		categories := GroupRulesByCategory(spec.Rules)
		categoryOrder := GetCategoryOrder(spec.Categories, categories)

		for _, catID := range categoryOrder {
			rules := categories[catID]
			if len(rules) == 0 {
				continue
			}
			filename := fmt.Sprintf("rules/%s.md", sanitizeFilename(catID))
			result.Pages[filename] = generateCategoryPage(spec, catID, rules, opts)
		}
		result.Pages["rules/index.md"] = generateRulesIndexPage(spec, opts)
	} else {
		result.Pages["rules.md"] = generateAllRulesPage(spec, opts)
	}

	// Generate glossary page
	if len(spec.Glossary) > 0 {
		result.Pages["glossary.md"] = generateGlossaryPage(spec, opts)
	}

	// Generate mkdocs.yml
	result.Config = generateMkDocsConfig(spec, opts, result.Pages)

	return result, nil
}

// WriteMkDocs writes the MkDocs site to a directory.
func WriteMkDocs(result *MkDocsResult, outputDir string) error {
	// Create docs directory
	docsDir := filepath.Join(outputDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Write mkdocs.yml
	configPath := filepath.Join(outputDir, "mkdocs.yml")
	if err := os.WriteFile(configPath, []byte(result.Config), 0600); err != nil {
		return fmt.Errorf("failed to write mkdocs.yml: %w", err)
	}

	// Write pages
	for relPath, content := range result.Pages {
		fullPath := filepath.Join(docsDir, relPath)

		// Create parent directories
		parentDir := filepath.Dir(fullPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", relPath, err)
		}
	}

	return nil
}

func generateIndexPage(spec *types.APIStyleSpec, opts *MkDocsOptions) string {
	var sb strings.Builder

	sb.WriteString("# ")
	sb.WriteString(spec.Name)
	sb.WriteString("\n\n")

	if spec.Description != "" {
		sb.WriteString(spec.Description)
		sb.WriteString("\n\n")
	}

	// Quick stats
	sb.WriteString("## Overview\n\n")

	ruleCount := len(spec.Rules)
	catCount := len(spec.Categories)
	patternCount := len(spec.Patterns)

	sb.WriteString("| Metric | Count |\n")
	sb.WriteString("|--------|-------|\n")
	fmt.Fprintf(&sb, "| Rules | %d |\n", ruleCount)
	fmt.Fprintf(&sb, "| Categories | %d |\n", catCount)
	if patternCount > 0 {
		fmt.Fprintf(&sb, "| Patterns | %d |\n", patternCount)
	}
	sb.WriteString("\n")

	// Navigation
	sb.WriteString("## Quick Links\n\n")

	if spec.Introduction != "" {
		sb.WriteString("- [Introduction](introduction.md)\n")
	}
	if len(spec.Principles) > 0 {
		sb.WriteString("- [Design Principles](principles.md)\n")
	}
	if len(spec.Patterns) > 0 {
		if opts.SplitPatterns {
			sb.WriteString("- [Design Patterns](patterns/index.md)\n")
		} else {
			sb.WriteString("- [Design Patterns](patterns.md)\n")
		}
	}
	if len(spec.ConformanceLevels) > 0 {
		sb.WriteString("- [Conformance Levels](conformance.md)\n")
	}
	if opts.SplitCategories {
		sb.WriteString("- [Rules](rules/index.md)\n")
	} else {
		sb.WriteString("- [Rules](rules.md)\n")
	}
	if len(spec.Glossary) > 0 {
		sb.WriteString("- [Glossary](glossary.md)\n")
	}

	// Metadata
	if spec.Metadata != nil {
		sb.WriteString("\n## About\n\n")
		if spec.Metadata.Author != "" {
			fmt.Fprintf(&sb, "- **Author:** %s\n", spec.Metadata.Author)
		}
		if spec.Metadata.URL != "" {
			fmt.Fprintf(&sb, "- **Source:** [%s](%s)\n", spec.Metadata.URL, spec.Metadata.URL)
		}
		if spec.Metadata.License != "" {
			fmt.Fprintf(&sb, "- **License:** %s\n", spec.Metadata.License)
		}
		if spec.Version != "" {
			fmt.Fprintf(&sb, "- **Version:** %s\n", spec.Version)
		}
	}

	return sb.String()
}

func generateIntroductionPage(spec *types.APIStyleSpec, _ *MkDocsOptions) string {
	var sb strings.Builder

	sb.WriteString("# Introduction\n\n")
	sb.WriteString(spec.Introduction)
	sb.WriteString("\n")

	return sb.String()
}

func generatePrinciplesPage(spec *types.APIStyleSpec, _ *MkDocsOptions) string {
	var sb strings.Builder

	sb.WriteString("# Design Principles\n\n")

	for _, p := range spec.Principles {
		fmt.Fprintf(&sb, "## %s\n\n", p.Title)
		sb.WriteString(p.Description)
		sb.WriteString("\n\n")

		if len(p.RelatedRules) > 0 {
			sb.WriteString("**Related Rules:** ")
			for i, ruleID := range p.RelatedRules {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(ruleID)
			}
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

func generatePatternsPage(spec *types.APIStyleSpec, opts *MkDocsOptions) string {
	var sb strings.Builder

	sb.WriteString("# Design Patterns\n\n")

	for _, p := range spec.Patterns {
		writePatternContent(&sb, p, opts)
	}

	return sb.String()
}

func generatePatternsIndexPage(spec *types.APIStyleSpec, _ *MkDocsOptions) string {
	var sb strings.Builder

	sb.WriteString("# Design Patterns\n\n")

	// Group by category if present
	byCategory := make(map[string][]types.Pattern)
	for _, p := range spec.Patterns {
		cat := p.Category
		if cat == "" {
			cat = "General"
		}
		byCategory[cat] = append(byCategory[cat], p)
	}

	// Sort categories
	var cats []string
	for cat := range byCategory {
		cats = append(cats, cat)
	}
	sort.Strings(cats)

	for _, cat := range cats {
		patterns := byCategory[cat]
		fmt.Fprintf(&sb, "## %s\n\n", titleCaser.String(cat))

		for _, p := range patterns {
			filename := sanitizeFilename(p.ID)
			fmt.Fprintf(&sb, "- [%s](%s.md): %s\n", p.Name, filename, p.Summary)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func generatePatternPage(p types.Pattern, opts *MkDocsOptions) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# %s\n\n", p.Name)
	writePatternContent(&sb, p, opts)

	return sb.String()
}

func writePatternContent(sb *strings.Builder, p types.Pattern, _ *MkDocsOptions) {
	if p.Summary != "" {
		sb.WriteString(p.Summary)
		sb.WriteString("\n\n")
	}

	if p.Problem != "" {
		sb.WriteString("## Problem\n\n")
		sb.WriteString(p.Problem)
		sb.WriteString("\n\n")
	}

	if p.Solution != "" {
		sb.WriteString("## Solution\n\n")
		sb.WriteString(p.Solution)
		sb.WriteString("\n\n")
	}

	if p.When != "" {
		sb.WriteString("## When to Use\n\n")
		sb.WriteString(p.When)
		sb.WriteString("\n\n")
	}

	if p.Description != "" {
		sb.WriteString("## Details\n\n")
		sb.WriteString(p.Description)
		sb.WriteString("\n\n")
	}

	// Examples
	if len(p.Examples) > 0 {
		sb.WriteString("## Examples\n\n")
		for _, ex := range p.Examples {
			writeDetailedExample(sb, ex)
		}
	}

	// Diagrams
	for _, d := range p.Diagrams {
		writeDiagram(sb, d)
	}

	// Related rules
	if len(p.RelatedRules) > 0 {
		sb.WriteString("## Related Rules\n\n")
		for _, ruleID := range p.RelatedRules {
			fmt.Fprintf(sb, "- %s\n", ruleID)
		}
		sb.WriteString("\n")
	}

	// References
	if len(p.References) > 0 {
		sb.WriteString("## References\n\n")
		for _, ref := range p.References {
			fmt.Fprintf(sb, "- [%s](%s)\n", ref.Title, ref.URL)
		}
		sb.WriteString("\n")
	}
}

func generateConformancePage(spec *types.APIStyleSpec, _ *MkDocsOptions) string {
	var sb strings.Builder

	sb.WriteString("# Conformance Levels\n\n")

	levelOrder := []string{"bronze", "silver", "gold"}

	for _, levelName := range levelOrder {
		level, ok := spec.ConformanceLevels[levelName]
		if !ok {
			continue
		}

		title := titleCaser.String(levelName)
		fmt.Fprintf(&sb, "## %s\n\n", title)

		if level.Description != "" {
			sb.WriteString(level.Description)
			sb.WriteString("\n\n")
		}

		if len(level.RequiredRules) > 0 {
			sb.WriteString("### Required Rules\n\n")
			for _, ruleID := range level.RequiredRules {
				fmt.Fprintf(&sb, "- %s\n", ruleID)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func generateRulesIndexPage(spec *types.APIStyleSpec, _ *MkDocsOptions) string {
	var sb strings.Builder

	sb.WriteString("# Rules\n\n")

	categories := GroupRulesByCategory(spec.Rules)
	categoryOrder := GetCategoryOrder(spec.Categories, categories)

	// Summary table
	sb.WriteString("| Category | Rules |\n")
	sb.WriteString("|----------|-------|\n")

	for _, catID := range categoryOrder {
		rules := categories[catID]
		catName := GetCategoryName(spec.Categories, catID)
		filename := sanitizeFilename(catID)
		fmt.Fprintf(&sb, "| [%s](%s.md) | %d |\n", catName, filename, len(rules))
	}

	sb.WriteString("\n")

	// Category descriptions
	sb.WriteString("## Categories\n\n")

	for _, catID := range categoryOrder {
		catName := GetCategoryName(spec.Categories, catID)
		catDesc := GetCategoryDescription(spec.Categories, catID)
		filename := sanitizeFilename(catID)

		fmt.Fprintf(&sb, "### [%s](%s.md)\n\n", catName, filename)

		if catDesc != "" {
			sb.WriteString(catDesc)
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

func generateCategoryPage(spec *types.APIStyleSpec, catID string, rules []types.Rule, opts *MkDocsOptions) string {
	var sb strings.Builder

	catName := GetCategoryName(spec.Categories, catID)
	catDesc := GetCategoryDescription(spec.Categories, catID)

	fmt.Fprintf(&sb, "# %s\n\n", catName)

	if catDesc != "" {
		sb.WriteString(catDesc)
		sb.WriteString("\n\n")
	}

	// Sort rules by priority then ID
	sortedRules := make([]types.Rule, len(rules))
	copy(sortedRules, rules)
	sort.Slice(sortedRules, func(i, j int) bool {
		if sortedRules[i].Priority != sortedRules[j].Priority {
			return sortedRules[i].Priority < sortedRules[j].Priority
		}
		return sortedRules[i].ID < sortedRules[j].ID
	})

	for _, rule := range sortedRules {
		writeRule(&sb, rule, opts.MarkdownOptions)
	}

	return sb.String()
}

func generateAllRulesPage(spec *types.APIStyleSpec, opts *MkDocsOptions) string {
	var sb strings.Builder

	sb.WriteString("# Rules\n\n")

	categories := GroupRulesByCategory(spec.Rules)
	categoryOrder := GetCategoryOrder(spec.Categories, categories)

	for _, catID := range categoryOrder {
		rules := categories[catID]
		if len(rules) == 0 {
			continue
		}

		catName := GetCategoryName(spec.Categories, catID)
		catDesc := GetCategoryDescription(spec.Categories, catID)

		fmt.Fprintf(&sb, "## %s\n\n", catName)

		if catDesc != "" {
			sb.WriteString(catDesc)
			sb.WriteString("\n\n")
		}

		for _, rule := range rules {
			writeRule(&sb, rule, opts.MarkdownOptions)
		}
	}

	return sb.String()
}

func generateGlossaryPage(spec *types.APIStyleSpec, _ *MkDocsOptions) string {
	var sb strings.Builder

	sb.WriteString("# Glossary\n\n")

	// Sort terms alphabetically
	terms := make([]types.GlossaryTerm, len(spec.Glossary))
	copy(terms, spec.Glossary)
	sort.Slice(terms, func(i, j int) bool {
		return strings.ToLower(terms[i].Term) < strings.ToLower(terms[j].Term)
	})

	for _, t := range terms {
		fmt.Fprintf(&sb, "## %s\n\n", t.Term)

		if len(t.Aliases) > 0 {
			sb.WriteString("*Also known as: ")
			for i, alias := range t.Aliases {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(alias)
			}
			sb.WriteString("*\n\n")
		}

		sb.WriteString(t.Definition)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func generateMkDocsConfig(spec *types.APIStyleSpec, opts *MkDocsOptions, pages map[string]string) string {
	var sb strings.Builder

	siteName := opts.SiteName
	if siteName == "" {
		siteName = spec.Name
	}

	fmt.Fprintf(&sb, "site_name: %s\n", siteName)

	if opts.SiteURL != "" {
		fmt.Fprintf(&sb, "site_url: %s\n", opts.SiteURL)
	}

	if opts.RepoURL != "" {
		fmt.Fprintf(&sb, "repo_url: %s\n", opts.RepoURL)
	} else if spec.Metadata != nil && spec.Metadata.Repository != "" {
		fmt.Fprintf(&sb, "repo_url: %s\n", spec.Metadata.Repository)
	}

	if spec.Description != "" {
		fmt.Fprintf(&sb, "site_description: %s\n", spec.Description)
	}

	if spec.Metadata != nil && spec.Metadata.Author != "" {
		fmt.Fprintf(&sb, "site_author: %s\n", spec.Metadata.Author)
	}

	sb.WriteString("\n")

	// Theme
	sb.WriteString("theme:\n")
	fmt.Fprintf(&sb, "  name: %s\n", opts.Theme)

	if opts.Theme == "material" {
		sb.WriteString("  features:\n")
		sb.WriteString("    - navigation.tabs\n")
		sb.WriteString("    - navigation.sections\n")
		sb.WriteString("    - navigation.expand\n")
		sb.WriteString("    - search.suggest\n")
		sb.WriteString("    - search.highlight\n")
		sb.WriteString("    - content.tabs.link\n")
		sb.WriteString("    - content.code.copy\n")
		sb.WriteString("  palette:\n")
		sb.WriteString("    - scheme: default\n")
		sb.WriteString("      primary: indigo\n")
		sb.WriteString("      accent: indigo\n")
		sb.WriteString("      toggle:\n")
		sb.WriteString("        icon: material/brightness-7\n")
		sb.WriteString("        name: Switch to dark mode\n")
		sb.WriteString("    - scheme: slate\n")
		sb.WriteString("      primary: indigo\n")
		sb.WriteString("      accent: indigo\n")
		sb.WriteString("      toggle:\n")
		sb.WriteString("        icon: material/brightness-4\n")
		sb.WriteString("        name: Switch to light mode\n")
	}

	sb.WriteString("\n")

	// Plugins
	if opts.IncludeSearch {
		sb.WriteString("plugins:\n")
		sb.WriteString("  - search\n")
		sb.WriteString("\n")
	}

	// Markdown extensions
	sb.WriteString("markdown_extensions:\n")
	sb.WriteString("  - tables\n")
	sb.WriteString("  - admonition\n")
	sb.WriteString("  - pymdownx.details\n")
	sb.WriteString("  - pymdownx.superfences:\n")
	sb.WriteString("      custom_fences:\n")
	sb.WriteString("        - name: mermaid\n")
	sb.WriteString("          class: mermaid\n")
	sb.WriteString("          format: !!python/name:pymdownx.superfences.fence_code_format\n")
	sb.WriteString("  - pymdownx.highlight:\n")
	sb.WriteString("      anchor_linenums: true\n")
	sb.WriteString("  - pymdownx.inlinehilite\n")
	sb.WriteString("  - pymdownx.snippets\n")
	sb.WriteString("  - def_list\n")
	sb.WriteString("  - attr_list\n")
	sb.WriteString("  - md_in_html\n")

	sb.WriteString("\n")

	// Navigation
	sb.WriteString("nav:\n")
	sb.WriteString("  - Home: index.md\n")

	if _, ok := pages["introduction.md"]; ok {
		sb.WriteString("  - Introduction: introduction.md\n")
	}

	if _, ok := pages["principles.md"]; ok {
		sb.WriteString("  - Principles: principles.md\n")
	}

	if _, ok := pages["conformance.md"]; ok {
		sb.WriteString("  - Conformance: conformance.md\n")
	}

	// Patterns
	if opts.SplitPatterns {
		if _, ok := pages["patterns/index.md"]; ok {
			sb.WriteString("  - Patterns:\n")
			sb.WriteString("    - Overview: patterns/index.md\n")
			for _, p := range spec.Patterns {
				filename := sanitizeFilename(p.ID)
				fmt.Fprintf(&sb, "    - %s: patterns/%s.md\n", p.Name, filename)
			}
		}
	} else {
		if _, ok := pages["patterns.md"]; ok {
			sb.WriteString("  - Patterns: patterns.md\n")
		}
	}

	// Rules
	if opts.SplitCategories {
		if _, ok := pages["rules/index.md"]; ok {
			sb.WriteString("  - Rules:\n")
			sb.WriteString("    - Overview: rules/index.md\n")

			categories := GroupRulesByCategory(spec.Rules)
			categoryOrder := GetCategoryOrder(spec.Categories, categories)

			for _, catID := range categoryOrder {
				catName := GetCategoryName(spec.Categories, catID)
				filename := sanitizeFilename(catID)
				fmt.Fprintf(&sb, "    - %s: rules/%s.md\n", catName, filename)
			}
		}
	} else {
		if _, ok := pages["rules.md"]; ok {
			sb.WriteString("  - Rules: rules.md\n")
		}
	}

	if _, ok := pages["glossary.md"]; ok {
		sb.WriteString("  - Glossary: glossary.md\n")
	}

	return sb.String()
}

func sanitizeFilename(s string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove any characters that aren't alphanumeric or hyphens
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
