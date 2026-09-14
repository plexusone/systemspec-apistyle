// Package lint provides deterministic linting for OpenAPI specifications.
package lint

import (
	"context"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Linter defines the interface for OpenAPI specification linting.
type Linter interface {
	// Lint analyzes an OpenAPI specification and returns violations.
	Lint(ctx context.Context, spec []byte, opts *Options) (*types.LintReport, error)
}

// Options configures linting behavior.
type Options struct {
	// FileName is the path to the specification file (for error reporting).
	FileName string

	// Profile is the style profile name being used.
	Profile string

	// ConformanceLevel is the target conformance level.
	ConformanceLevel string

	// FailFast stops on first error if true.
	FailFast bool

	// Timeout is the maximum duration for linting.
	Timeout int

	// Exceptions is a list of exceptions to apply.
	Exceptions []types.Exception
}

// DefaultOptions returns options with sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		FileName: "openapi.yaml",
		Profile:  "default",
		Timeout:  30000, // 30 seconds
	}
}
