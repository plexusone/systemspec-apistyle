package generate

import (
	"fmt"
	"strings"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// SpectralOptions configures Spectral ruleset generation.
type SpectralOptions struct {
	// IncludeDisabled includes disabled rules (commented out).
	IncludeDisabled bool

	// IncludeDescriptions adds rule descriptions as comments.
	IncludeDescriptions bool

	// SkipNonEnforceable skips rules without enforcement config.
	SkipNonEnforceable bool
}

// DefaultSpectralOptions returns options with sensible defaults.
func DefaultSpectralOptions() *SpectralOptions {
	return &SpectralOptions{
		IncludeDisabled:     false,
		IncludeDescriptions: true,
		SkipNonEnforceable:  true,
	}
}

// Spectral generates a Spectral-compatible YAML ruleset from an APIStyleSpec.
func Spectral(spec *types.APIStyleSpec, opts *SpectralOptions) (string, error) {
	if opts == nil {
		opts = DefaultSpectralOptions()
	}

	var sb strings.Builder

	// Header
	sb.WriteString("# Spectral Ruleset\n")
	fmt.Fprintf(&sb, "# Generated from: %s\n", spec.Name)
	if spec.Version != "" {
		fmt.Fprintf(&sb, "# Version: %s\n", spec.Version)
	}
	sb.WriteString("#\n")
	sb.WriteString("# This file was auto-generated from an systemspec-apistyle profile.\n")
	sb.WriteString("# Do not edit manually - modify the source profile instead.\n\n")

	// Extends section (if applicable)
	sb.WriteString("extends:\n")
	sb.WriteString("  - spectral:oas\n\n")

	// Rules section
	sb.WriteString("rules:\n")

	for _, rule := range spec.Rules {
		writeSpectralRule(&sb, rule, opts)
	}

	return sb.String(), nil
}

func writeSpectralRule(sb *strings.Builder, rule types.Rule, opts *SpectralOptions) {
	// Skip rules without enforcement
	if rule.Enforcement == nil {
		if opts.SkipNonEnforceable {
			return
		}
		// Write as comment
		fmt.Fprintf(sb, "  # %s: (no enforcement - LLM-only rule)\n", rule.ID)
		return
	}

	// Skip non-spectral enforcement types
	if rule.Enforcement.Type != types.EnforcementSpectral {
		if opts.SkipNonEnforceable {
			return
		}
		fmt.Fprintf(sb, "  # %s: (enforcement type: %s)\n", rule.ID, rule.Enforcement.Type)
		return
	}

	// Rule description as comment
	if opts.IncludeDescriptions {
		fmt.Fprintf(sb, "\n  # %s: %s\n", rule.ID, rule.Title)
		if rule.Rationale != "" {
			// Truncate long rationales
			rationale := rule.Rationale
			if len(rationale) > 80 {
				rationale = rationale[:77] + "..."
			}
			fmt.Fprintf(sb, "  # %s\n", rationale)
		}
	}

	// Rule ID (converted to kebab-case for Spectral convention)
	ruleID := toKebabCase(rule.ID)
	fmt.Fprintf(sb, "  %s:\n", ruleID)

	// Description
	if rule.Title != "" {
		fmt.Fprintf(sb, "    description: %q\n", rule.Title)
	}

	// Message
	message := rule.Title
	if rule.Rationale != "" && len(rule.Rationale) < 100 {
		message = rule.Rationale
	}
	fmt.Fprintf(sb, "    message: %q\n", message)

	// Severity
	fmt.Fprintf(sb, "    severity: %s\n", mapSeverity(rule.Severity))

	// Given
	if rule.Enforcement.Given != nil && len(rule.Enforcement.Given.Paths) > 0 {
		if len(rule.Enforcement.Given.Paths) == 1 {
			fmt.Fprintf(sb, "    given: %q\n", rule.Enforcement.Given.Paths[0])
		} else {
			sb.WriteString("    given:\n")
			for _, path := range rule.Enforcement.Given.Paths {
				fmt.Fprintf(sb, "      - %q\n", path)
			}
		}
	}

	// Then
	sb.WriteString("    then:\n")
	writeThen(sb, rule.Enforcement)
}

func writeThen(sb *strings.Builder, enf *types.Enforcement) {
	// Determine the function
	function := enf.Function
	if enf.Then != nil && enf.Then.Function != "" {
		function = enf.Then.Function
	}

	// Field (if specified in Then)
	if enf.Then != nil && enf.Then.Field != "" {
		fmt.Fprintf(sb, "      field: %q\n", enf.Then.Field)
	} else if isBooleanFunction(function) && enf.Options != nil && enf.Options.Match != "" {
		// For truthy/falsy functions, options.match should be output as field
		// (match is only valid for pattern function)
		fmt.Fprintf(sb, "      field: %q\n", enf.Options.Match)
	}

	// Function
	if function != "" {
		fmt.Fprintf(sb, "      function: %s\n", function)
	}

	// Function options (not for boolean functions that use field)
	if enf.Options != nil {
		writeFunctionOptions(sb, enf.Options, function)
	} else if enf.Then != nil && enf.Then.FunctionOptions != nil {
		sb.WriteString("      functionOptions:\n")
		for k, v := range enf.Then.FunctionOptions {
			fmt.Fprintf(sb, "        %s: %q\n", k, v)
		}
	}
}

// isBooleanFunction returns true for functions that check existence (truthy/falsy)
// These functions don't use match/notMatch - they use field to specify the target.
func isBooleanFunction(function string) bool {
	return function == "truthy" || function == "falsy" || function == "defined" || function == "undefined"
}

func writeFunctionOptions(sb *strings.Builder, opts *types.EnforcementOptions, function string) {
	// For boolean functions (truthy/falsy), match is handled as field, not functionOptions
	skipMatch := isBooleanFunction(function)

	hasOptions := (!skipMatch && opts.Match != "") || opts.NotMatch != "" ||
		opts.Min != nil || opts.Max != nil ||
		len(opts.Values) > 0 || opts.Type != ""

	if !hasOptions {
		return
	}

	sb.WriteString("      functionOptions:\n")

	if opts.Match != "" && !skipMatch {
		fmt.Fprintf(sb, "        match: %q\n", opts.Match)
	}
	if opts.NotMatch != "" {
		fmt.Fprintf(sb, "        notMatch: %q\n", opts.NotMatch)
	}
	if opts.Min != nil {
		fmt.Fprintf(sb, "        min: %d\n", *opts.Min)
	}
	if opts.Max != nil {
		fmt.Fprintf(sb, "        max: %d\n", *opts.Max)
	}
	if len(opts.Values) > 0 {
		sb.WriteString("        values:\n")
		for _, v := range opts.Values {
			fmt.Fprintf(sb, "          - %q\n", v)
		}
	}
	if opts.Type != "" {
		fmt.Fprintf(sb, "        type: %s\n", opts.Type)
	}
	if opts.Separator != "" {
		fmt.Fprintf(sb, "        separator: %q\n", opts.Separator)
	}
}

func mapSeverity(sev types.Severity) string {
	switch sev {
	case types.SeverityError:
		return "error"
	case types.SeverityWarn:
		return "warn"
	case types.SeverityInfo:
		return "info"
	case types.SeverityHint:
		return "hint"
	default:
		return "warn"
	}
}

func toKebabCase(s string) string {
	// Convert rule IDs like "URI-001" to "uri-001"
	return strings.ToLower(s)
}
