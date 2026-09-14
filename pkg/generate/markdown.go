package generate

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

var titleCaser = cases.Title(language.English)

// MarkdownOptions configures Markdown generation.
type MarkdownOptions struct {
	// IncludeTOC adds a table of contents.
	IncludeTOC bool

	// IncludeExamples includes good/bad examples.
	IncludeExamples bool

	// IncludeRationale includes rule rationales.
	IncludeRationale bool

	// IncludeReferences includes external references.
	IncludeReferences bool

	// IncludeConformance includes conformance level details.
	IncludeConformance bool

	// IncludeMetadata includes spec metadata.
	IncludeMetadata bool

	// IncludeIntroduction includes the introduction section.
	IncludeIntroduction bool

	// IncludePrinciples includes design principles.
	IncludePrinciples bool

	// IncludePatterns includes API design patterns.
	IncludePatterns bool

	// IncludeGlossary includes the glossary.
	IncludeGlossary bool

	// IncludeDescription includes rule descriptions.
	IncludeDescription bool

	// SeverityEmojis uses emojis for severity indicators.
	SeverityEmojis bool
}

// DefaultMarkdownOptions returns options with all features enabled.
func DefaultMarkdownOptions() *MarkdownOptions {
	return &MarkdownOptions{
		IncludeTOC:          true,
		IncludeExamples:     true,
		IncludeRationale:    true,
		IncludeReferences:   true,
		IncludeConformance:  true,
		IncludeMetadata:     true,
		IncludeIntroduction: true,
		IncludePrinciples:   true,
		IncludePatterns:     true,
		IncludeGlossary:     true,
		IncludeDescription:  true,
		SeverityEmojis:      false,
	}
}

// Markdown generates Markdown documentation from an APIStyleSpec.
func Markdown(spec *types.APIStyleSpec, opts *MarkdownOptions) (string, error) {
	if opts == nil {
		opts = DefaultMarkdownOptions()
	}

	var sb strings.Builder

	// Title
	sb.WriteString("# ")
	sb.WriteString(spec.Name)
	sb.WriteString("\n\n")

	// Description
	if spec.Description != "" {
		sb.WriteString(spec.Description)
		sb.WriteString("\n\n")
	}

	// Metadata
	if opts.IncludeMetadata && spec.Metadata != nil {
		writeMetadata(&sb, spec.Metadata)
	}

	// Table of Contents
	if opts.IncludeTOC {
		writeTOC(&sb, spec, opts)
	}

	// Introduction
	if opts.IncludeIntroduction && spec.Introduction != "" {
		writeIntroduction(&sb, spec.Introduction)
	}

	// Principles
	if opts.IncludePrinciples && len(spec.Principles) > 0 {
		writePrinciples(&sb, spec.Principles)
	}

	// Conformance Levels
	if opts.IncludeConformance && len(spec.ConformanceLevels) > 0 {
		writeConformanceLevels(&sb, spec)
	}

	// Patterns
	if opts.IncludePatterns && len(spec.Patterns) > 0 {
		writePatterns(&sb, spec.Patterns)
	}

	// Rules by Category
	writeRulesByCategory(&sb, spec, opts)

	// Glossary
	if opts.IncludeGlossary && len(spec.Glossary) > 0 {
		writeGlossary(&sb, spec.Glossary)
	}

	return sb.String(), nil
}

func writeMetadata(sb *strings.Builder, meta *types.SpecMetadata) {
	sb.WriteString("## Metadata\n\n")

	if meta.Author != "" {
		fmt.Fprintf(sb, "- **Author:** %s\n", meta.Author)
	}
	if meta.URL != "" {
		fmt.Fprintf(sb, "- **Source:** [%s](%s)\n", meta.URL, meta.URL)
	}
	if meta.LastUpdated != "" {
		fmt.Fprintf(sb, "- **Last Updated:** %s\n", meta.LastUpdated)
	}
	if meta.License != "" {
		fmt.Fprintf(sb, "- **License:** %s\n", meta.License)
	}

	sb.WriteString("\n")
}

func writeTOC(sb *strings.Builder, spec *types.APIStyleSpec, opts *MarkdownOptions) {
	sb.WriteString("## Table of Contents\n\n")

	if opts.IncludeIntroduction && spec.Introduction != "" {
		sb.WriteString("- [Introduction](#introduction)\n")
	}

	if opts.IncludePrinciples && len(spec.Principles) > 0 {
		sb.WriteString("- [Design Principles](#design-principles)\n")
	}

	if opts.IncludeConformance && len(spec.ConformanceLevels) > 0 {
		sb.WriteString("- [Conformance Levels](#conformance-levels)\n")
	}

	if opts.IncludePatterns && len(spec.Patterns) > 0 {
		sb.WriteString("- [Design Patterns](#design-patterns)\n")
	}

	// Group rules by category
	categories := GroupRulesByCategory(spec.Rules)
	categoryOrder := GetCategoryOrder(spec.Categories, categories)

	sb.WriteString("- **Rules**\n")
	for _, catID := range categoryOrder {
		catName := GetCategoryName(spec.Categories, catID)
		anchor := strings.ToLower(strings.ReplaceAll(catName, " ", "-"))
		fmt.Fprintf(sb, "  - [%s](#%s)\n", catName, anchor)
	}

	if opts.IncludeGlossary && len(spec.Glossary) > 0 {
		sb.WriteString("- [Glossary](#glossary)\n")
	}

	sb.WriteString("\n")
}

func writeConformanceLevels(sb *strings.Builder, spec *types.APIStyleSpec) {
	sb.WriteString("## Conformance Levels\n\n")

	// Sort levels: bronze, silver, gold
	levelOrder := []string{"bronze", "silver", "gold"}

	for _, levelName := range levelOrder {
		level, ok := spec.ConformanceLevels[levelName]
		if !ok {
			continue
		}

		title := titleCaser.String(levelName)
		fmt.Fprintf(sb, "### %s\n\n", title)

		if level.Description != "" {
			sb.WriteString(level.Description)
			sb.WriteString("\n\n")
		}

		if len(level.RequiredRules) > 0 {
			sb.WriteString("**Required Rules:**\n\n")
			for _, ruleID := range level.RequiredRules {
				fmt.Fprintf(sb, "- %s\n", ruleID)
			}
			sb.WriteString("\n")
		}
	}
}

func writeRulesByCategory(sb *strings.Builder, spec *types.APIStyleSpec, opts *MarkdownOptions) {
	categories := GroupRulesByCategory(spec.Rules)
	categoryOrder := GetCategoryOrder(spec.Categories, categories)

	for _, catID := range categoryOrder {
		rules := categories[catID]
		if len(rules) == 0 {
			continue
		}

		catName := GetCategoryName(spec.Categories, catID)
		catDesc := GetCategoryDescription(spec.Categories, catID)

		fmt.Fprintf(sb, "## %s\n\n", catName)

		if catDesc != "" {
			sb.WriteString(catDesc)
			sb.WriteString("\n\n")
		}

		for _, rule := range rules {
			writeRule(sb, rule, opts)
		}
	}
}

func writeRule(sb *strings.Builder, rule types.Rule, opts *MarkdownOptions) {
	// Rule header
	fmt.Fprintf(sb, "### %s: %s\n\n", rule.ID, rule.Title)

	// Severity badge
	severityStr := formatSeverity(rule.Severity, opts.SeverityEmojis)
	fmt.Fprintf(sb, "**Severity:** %s\n\n", severityStr)

	// Description (extended prose)
	if opts.IncludeDescription && rule.Description != "" {
		sb.WriteString(rule.Description)
		sb.WriteString("\n\n")
	}

	// Rationale
	if opts.IncludeRationale && rule.Rationale != "" {
		if rule.Description == "" {
			// If no description, rationale is the main content
			sb.WriteString(rule.Rationale)
		} else {
			// If there's a description, rationale is secondary
			sb.WriteString("**Rationale:** ")
			sb.WriteString(rule.Rationale)
		}
		sb.WriteString("\n\n")
	}

	// Decision Tables
	if len(rule.DecisionTables) > 0 {
		for _, dt := range rule.DecisionTables {
			writeDecisionTable(sb, dt)
		}
	}

	// Examples
	if opts.IncludeExamples && rule.Examples != nil {
		writeExamples(sb, rule.Examples)
	}

	// References
	if opts.IncludeReferences && len(rule.References) > 0 {
		writeReferences(sb, rule.References)
	}

	sb.WriteString("---\n\n")
}

func writeExamples(sb *strings.Builder, examples *types.Examples) {
	hasSimple := len(examples.Good) > 0 || len(examples.Bad) > 0
	hasDetailed := len(examples.Detailed) > 0

	if !hasSimple && !hasDetailed {
		return
	}

	sb.WriteString("**Examples:**\n\n")

	// Simple examples (backward compatible)
	if len(examples.Good) > 0 {
		sb.WriteString("Good:\n\n")
		for _, ex := range examples.Good {
			fmt.Fprintf(sb, "- `%s`\n", ex)
		}
		sb.WriteString("\n")
	}

	if len(examples.Bad) > 0 {
		sb.WriteString("Bad:\n\n")
		for _, ex := range examples.Bad {
			fmt.Fprintf(sb, "- `%s`\n", ex)
		}
		sb.WriteString("\n")
	}

	// Detailed examples
	for _, ex := range examples.Detailed {
		writeDetailedExample(sb, ex)
	}
}

func writeReferences(sb *strings.Builder, refs []types.Reference) {
	sb.WriteString("**References:**\n\n")
	for _, ref := range refs {
		fmt.Fprintf(sb, "- [%s](%s)\n", ref.Title, ref.URL)
	}
	sb.WriteString("\n")
}

func formatSeverity(sev types.Severity, useEmoji bool) string {
	if useEmoji {
		switch sev {
		case types.SeverityError:
			return "Error"
		case types.SeverityWarn:
			return "Warning"
		case types.SeverityInfo:
			return "Info"
		case types.SeverityHint:
			return "Hint"
		default:
			return string(sev)
		}
	}
	return string(sev)
}

func GroupRulesByCategory(rules []types.Rule) map[string][]types.Rule {
	categories := make(map[string][]types.Rule)
	for _, rule := range rules {
		cat := rule.Category
		if cat == "" {
			cat = "other"
		}
		categories[cat] = append(categories[cat], rule)
	}
	return categories
}

func GetCategoryOrder(categories []types.Category, ruleCategories map[string][]types.Rule) []string {
	// First add categories in defined order
	var order []string
	seen := make(map[string]bool)

	for _, cat := range categories {
		if _, hasRules := ruleCategories[cat.ID]; hasRules {
			order = append(order, cat.ID)
			seen[cat.ID] = true
		}
	}

	// Then add any categories not in the defined list
	var extra []string
	for catID := range ruleCategories {
		if !seen[catID] {
			extra = append(extra, catID)
		}
	}
	sort.Strings(extra)
	order = append(order, extra...)

	return order
}

func GetCategoryName(categories []types.Category, id string) string {
	for _, cat := range categories {
		if cat.ID == id {
			return cat.Title
		}
	}
	// Fallback: title case the ID
	return titleCaser.String(strings.ReplaceAll(id, "-", " "))
}

func GetCategoryDescription(categories []types.Category, id string) string {
	for _, cat := range categories {
		if cat.ID == id {
			return cat.Description
		}
	}
	return ""
}

func writeIntroduction(sb *strings.Builder, intro string) {
	sb.WriteString("## Introduction\n\n")
	sb.WriteString(intro)
	sb.WriteString("\n\n")
}

func writePrinciples(sb *strings.Builder, principles []types.Principle) {
	sb.WriteString("## Design Principles\n\n")

	for _, p := range principles {
		fmt.Fprintf(sb, "### %s\n\n", p.Title)
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
}

func writePatterns(sb *strings.Builder, patterns []types.Pattern) {
	sb.WriteString("## Design Patterns\n\n")

	for _, p := range patterns {
		fmt.Fprintf(sb, "### %s\n\n", p.Name)

		if p.Summary != "" {
			sb.WriteString(p.Summary)
			sb.WriteString("\n\n")
		}

		if p.Problem != "" {
			sb.WriteString("**Problem:** ")
			sb.WriteString(p.Problem)
			sb.WriteString("\n\n")
		}

		if p.Solution != "" {
			sb.WriteString("**Solution:** ")
			sb.WriteString(p.Solution)
			sb.WriteString("\n\n")
		}

		if p.When != "" {
			sb.WriteString("**When to Use:** ")
			sb.WriteString(p.When)
			sb.WriteString("\n\n")
		}

		if p.Description != "" {
			sb.WriteString(p.Description)
			sb.WriteString("\n\n")
		}

		// Examples
		for _, ex := range p.Examples {
			writeDetailedExample(sb, ex)
		}

		// Diagrams
		for _, d := range p.Diagrams {
			writeDiagram(sb, d)
		}

		// Related rules
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

		// References
		if len(p.References) > 0 {
			writeReferences(sb, p.References)
		}

		sb.WriteString("---\n\n")
	}
}

func writeGlossary(sb *strings.Builder, terms []types.GlossaryTerm) {
	sb.WriteString("## Glossary\n\n")

	for _, t := range terms {
		fmt.Fprintf(sb, "**%s**", t.Term)
		if len(t.Aliases) > 0 {
			sb.WriteString(" (")
			for i, alias := range t.Aliases {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(alias)
			}
			sb.WriteString(")")
		}
		sb.WriteString("\n: ")
		sb.WriteString(t.Definition)
		sb.WriteString("\n\n")
	}
}

func writeDetailedExample(sb *strings.Builder, ex types.DetailedExample) {
	// Title with type indicator
	typeLabel := ""
	switch ex.Type {
	case "good":
		typeLabel = " (Correct)"
	case "bad":
		typeLabel = " (Incorrect)"
	case "context":
		typeLabel = ""
	}
	fmt.Fprintf(sb, "**%s%s**\n\n", ex.Title, typeLabel)

	if ex.Description != "" {
		sb.WriteString(ex.Description)
		sb.WriteString("\n\n")
	}

	// Code block with language
	lang := ex.Language
	if lang == "" {
		lang = "text"
	}

	// Before/After for migration examples
	if ex.Before != "" && ex.After != "" {
		sb.WriteString("Before:\n\n")
		fmt.Fprintf(sb, "```%s\n%s\n```\n\n", lang, ex.Before)
		sb.WriteString("After:\n\n")
		fmt.Fprintf(sb, "```%s\n%s\n```\n\n", lang, ex.After)
	} else if ex.Code != "" {
		fmt.Fprintf(sb, "```%s\n%s\n```\n\n", lang, ex.Code)
	}

	// Annotations as a list
	if len(ex.Annotations) > 0 {
		sb.WriteString("Notes:\n\n")
		for _, ann := range ex.Annotations {
			lineRef := ""
			if ann.Line > 0 {
				if ann.EndLine > 0 && ann.EndLine != ann.Line {
					lineRef = fmt.Sprintf("Lines %d-%d: ", ann.Line, ann.EndLine)
				} else {
					lineRef = fmt.Sprintf("Line %d: ", ann.Line)
				}
			}
			fmt.Fprintf(sb, "- %s%s\n", lineRef, ann.Text)
		}
		sb.WriteString("\n")
	}
}

func writeDiagram(sb *strings.Builder, d types.Diagram) {
	fmt.Fprintf(sb, "**%s**\n\n", d.Title)

	switch d.Type {
	case "mermaid":
		fmt.Fprintf(sb, "```mermaid\n%s\n```\n\n", d.Content)
	case "plantuml":
		fmt.Fprintf(sb, "```plantuml\n%s\n```\n\n", d.Content)
	case "url":
		if d.Alt != "" {
			fmt.Fprintf(sb, "![%s](%s)\n\n", d.Alt, d.Content)
		} else {
			fmt.Fprintf(sb, "![%s](%s)\n\n", d.Title, d.Content)
		}
	default:
		sb.WriteString(d.Content)
		sb.WriteString("\n\n")
	}
}

func writeDecisionTable(sb *strings.Builder, dt types.DecisionTable) {
	if dt.Title != "" {
		fmt.Fprintf(sb, "**%s**\n\n", dt.Title)
	}

	if dt.Description != "" {
		sb.WriteString(dt.Description)
		sb.WriteString("\n\n")
	}

	if len(dt.Headers) == 0 || len(dt.Rows) == 0 {
		return
	}

	// Header row
	sb.WriteString("| ")
	for i, h := range dt.Headers {
		if i > 0 {
			sb.WriteString(" | ")
		}
		sb.WriteString(h)
	}
	sb.WriteString(" |\n")

	// Separator row
	sb.WriteString("|")
	for range dt.Headers {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range dt.Rows {
		sb.WriteString("| ")
		for i, v := range row.Values {
			if i > 0 {
				sb.WriteString(" | ")
			}
			sb.WriteString(v)
		}
		sb.WriteString(" |\n")
	}
	sb.WriteString("\n")
}
