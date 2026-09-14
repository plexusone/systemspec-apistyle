// Package fix provides fix suggestion generation for API style violations.
package fix

import (
	"context"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Fixer generates fix suggestions for violations.
type Fixer interface {
	// SuggestFixes generates fix suggestions for the given violations.
	SuggestFixes(ctx context.Context, spec []byte, violations []types.Violation, opts *Options) (*types.FixReport, error)

	// DesignCheck provides pre-generation guidance for designing an API resource.
	DesignCheck(ctx context.Context, resource string, operations []string, opts *Options) (*types.DesignCheck, error)

	// ConformancePath shows the path to reach a conformance level.
	ConformancePath(ctx context.Context, spec []byte, targetLevel string, opts *Options) (*types.ConformancePath, error)
}

// Options configures fix generation.
type Options struct {
	// Profile is the style profile to use.
	Profile string

	// MaxSuggestions limits the number of suggestions returned.
	MaxSuggestions int

	// IncludePatch generates JSON Patch operations.
	IncludePatch bool

	// UseLLM enables LLM-based fix generation for complex rules.
	UseLLM bool
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		Profile:        "default",
		MaxSuggestions: 50,
		IncludePatch:   true,
		UseLLM:         false,
	}
}
