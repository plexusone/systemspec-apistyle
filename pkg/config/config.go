// Package config provides configuration file support for systemspec-apistyle.
//
// Configuration files are loaded from multiple locations in priority order:
//  1. Explicit path via --config flag
//  2. .api-style.yaml in current directory
//  3. .api-style.yml in current directory
//  4. api-style.yaml in current directory
package config

import "github.com/plexusone/systemspec-apistyle/pkg/types"

// Config represents the systemspec-apistyle configuration file.
type Config struct {
	// Profile is the style profile to use for linting.
	Profile string `yaml:"profile" json:"profile"`

	// Level is the target conformance level (bronze, silver, gold).
	Level string `yaml:"level" json:"level"`

	// Exceptions defines rule exceptions for this project.
	Exceptions []ExceptionConfig `yaml:"exceptions" json:"exceptions"`

	// SeverityOverrides allows changing rule severities.
	// Map of rule ID to severity (error, warn, info, hint).
	SeverityOverrides map[string]string `yaml:"severity-overrides" json:"severityOverrides"`

	// Include defines file patterns to include in linting.
	Include []string `yaml:"include" json:"include"`

	// Exclude defines file patterns to exclude from linting.
	Exclude []string `yaml:"exclude" json:"exclude"`
}

// ExceptionConfig defines an exception in the config file.
type ExceptionConfig struct {
	// RuleID is the rule being waived.
	RuleID string `yaml:"rule" json:"rule"`

	// Paths are glob patterns where the exception applies.
	Paths []string `yaml:"paths" json:"paths"`

	// Reason explains why this exception was granted.
	Reason string `yaml:"reason" json:"reason"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Profile: "default",
		Level:   "",
		Include: []string{
			"openapi.yaml",
			"openapi.yml",
			"openapi.json",
			"swagger.yaml",
			"swagger.yml",
			"swagger.json",
			"**/openapi.yaml",
			"**/openapi.yml",
			"**/openapi.json",
			"**/api.yaml",
			"**/api.yml",
		},
		Exclude: []string{
			"**/node_modules/**",
			"**/vendor/**",
			"**/.git/**",
		},
	}
}

// ToExceptions converts ExceptionConfigs to types.Exception.
func (c *Config) ToExceptions() []types.Exception {
	exceptions := make([]types.Exception, 0, len(c.Exceptions))
	for i, ec := range c.Exceptions {
		exc := types.Exception{
			ID:     ec.RuleID + "-exception-" + string(rune('0'+i)),
			RuleID: ec.RuleID,
			Reason: ec.Reason,
		}
		if len(ec.Paths) > 0 {
			exc.AppliesTo = &types.ExceptionScope{
				Paths: ec.Paths,
			}
		}
		exceptions = append(exceptions, exc)
	}
	return exceptions
}

// Merge merges CLI flags into this config. CLI flags take precedence.
func (c *Config) Merge(profile, level string, include, exclude []string) {
	if profile != "" && profile != "default" {
		c.Profile = profile
	}
	if level != "" {
		c.Level = level
	}
	if len(include) > 0 {
		c.Include = include
	}
	if len(exclude) > 0 {
		c.Exclude = exclude
	}
}
