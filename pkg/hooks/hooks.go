// Package hooks provides integration with AI coding assistants via assistantkit.
//
// This package defines systemspec-apistyle hooks that can be exported to multiple
// AI assistant formats (Claude Code, Kiro, Cursor, Windsurf, etc.).
//
// Example hooks:
//   - AfterFileWrite: Auto-lint OpenAPI specs when saved
//   - BeforePrompt: Inject API style context when working on APIs
package hooks

import (
	"fmt"
	"path/filepath"

	"github.com/plexusone/assistantkit/hooks"
	"github.com/plexusone/assistantkit/hooks/core"
)

// Config represents the systemspec-apistyle hooks configuration.
type Config struct {
	// Profile is the style profile to use for linting.
	Profile string

	// AutoLint enables automatic linting on file save.
	AutoLint bool

	// AutoLintPatterns are glob patterns for files to auto-lint.
	// Defaults to common OpenAPI patterns if empty.
	AutoLintPatterns []string

	// InjectContext enables injecting style context before prompts.
	InjectContext bool
}

// DefaultConfig returns the default hooks configuration.
func DefaultConfig() *Config {
	return &Config{
		Profile:  "default",
		AutoLint: true,
		AutoLintPatterns: []string{
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
		InjectContext: false,
	}
}

// Generate creates an assistantkit hooks.Config from the systemspec-apistyle configuration.
func (c *Config) Generate() *hooks.Config {
	cfg := hooks.NewConfig()

	if c.AutoLint {
		c.addAutoLintHooks(cfg)
	}

	if c.InjectContext {
		c.addContextInjectionHooks(cfg)
	}

	return cfg
}

// addAutoLintHooks adds hooks for automatic linting on file save.
func (c *Config) addAutoLintHooks(cfg *hooks.Config) {
	// Create a matcher pattern for OpenAPI files
	matcher := "Write|Edit"

	// The command checks if the file matches OpenAPI patterns and runs lint
	command := c.buildLintCommand()

	hook := hooks.NewCommandHook(command).WithTimeout(30)

	cfg.AddHookWithMatcher(hooks.AfterFileWrite, matcher, hook)
}

// addContextInjectionHooks adds hooks for injecting API style context.
func (c *Config) addContextInjectionHooks(cfg *hooks.Config) {
	prompt := fmt.Sprintf(`When working with OpenAPI/Swagger specifications, remember to follow the "%s" API style guidelines.

Key principles:
- Use plural resource names (e.g., /users not /user)
- Use lowercase with hyphens in URIs (e.g., /user-accounts)
- Include proper HTTP status codes and error responses
- Document all operations with descriptions
- Use consistent naming conventions

Use the api-style MCP tools (lint, explain_rule) for guidance.`, c.Profile)

	hook := hooks.NewPromptHook(prompt)
	cfg.AddHook(hooks.BeforePrompt, hook)
}

// buildLintCommand creates the shell command for linting.
func (c *Config) buildLintCommand() string {
	// Build a shell script that checks if the file is an OpenAPI spec
	// and runs the linter if so
	return fmt.Sprintf(`#!/bin/bash
FILE="$CLAUDE_FILE_PATH"
if [[ -z "$FILE" ]]; then
  exit 0
fi

# Check if file matches OpenAPI patterns
case "$FILE" in
  */openapi.yaml|*/openapi.yml|*/openapi.json|*/swagger.yaml|*/swagger.yml|*/swagger.json|*/api.yaml|*/api.yml)
    # Check if mcp-api-style is available
    if command -v mcp-api-style &> /dev/null; then
      mcp-api-style lint -f "$FILE" -p %s 2>&1 | head -50
    fi
    ;;
esac
exit 0`, c.Profile)
}

// WriteToFile writes the hooks configuration to a file in the specified format.
func (c *Config) WriteToFile(path, format string) error {
	cfg := c.Generate()

	adapter, ok := hooks.GetAdapter(format)
	if !ok {
		return fmt.Errorf("unsupported format: %s (supported: %v)", format, hooks.AdapterNames())
	}

	// Filter to only events supported by the target
	filtered := cfg.FilterByTool(format)

	return adapter.WriteFile(filtered, path)
}

// MarshalFormat converts the configuration to a specific format.
func (c *Config) MarshalFormat(format string) ([]byte, error) {
	cfg := c.Generate()

	adapter, ok := hooks.GetAdapter(format)
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	// Filter to only events supported by the target
	filtered := cfg.FilterByTool(format)

	return adapter.Marshal(filtered)
}

// DefaultPaths returns the default output paths for each supported format.
func DefaultPaths() map[string]string {
	return map[string]string{
		"claude":   filepath.Join(".claude", "settings.json"),
		"cursor":   filepath.Join(".cursor", "hooks.json"),
		"windsurf": filepath.Join(".windsurf", "hooks.json"),
	}
}

// SupportedFormats returns the list of supported output formats.
func SupportedFormats() []string {
	return hooks.AdapterNames()
}

// EventSupport returns which events are supported by which tools.
func EventSupport() map[string][]core.Event {
	support := make(map[string][]core.Event)
	for _, name := range hooks.AdapterNames() {
		adapter, _ := hooks.GetAdapter(name)
		support[name] = adapter.SupportedEvents()
	}
	return support
}
