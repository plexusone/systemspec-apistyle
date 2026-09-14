package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	hookscore "github.com/plexusone/assistantkit/hooks/core"
	"github.com/plexusone/systemspec-apistyle/pkg/hooks"
	"github.com/spf13/cobra"
)

var (
	hooksProfile     string
	hooksFormat      string
	hooksOutput      string
	hooksAutoLint    bool
	hooksInjectCtx   bool
	hooksListFormats bool
	hooksForce       bool
	hooksLevel       string
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Generate AI assistant hooks configuration",
	Long: `Generate hooks configuration for AI coding assistants.

Hooks enable automatic API linting when files are saved and can inject
API style context into prompts.

Supported output formats:
  - claude:   Claude Code (.claude/settings.json)
  - cursor:   Cursor IDE (.cursor/hooks.json)
  - windsurf: Windsurf/Codeium (.windsurf/hooks.json)

Examples:
  # Generate Claude Code hooks with default settings
  api-style hooks --format claude

  # Generate hooks for all supported formats
  api-style hooks --format all

  # Generate with a specific profile
  api-style hooks --format claude --profile azure

  # Write to a specific output file
  api-style hooks --format claude --output .claude/settings.json

  # Install git pre-commit hook
  api-style hooks init`,
	RunE: runHooks,
}

var hooksGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate hooks configuration files",
	Long:  "Generate hooks configuration files for AI coding assistants.",
	RunE:  runHooks,
}

var hooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported hooks formats and events",
	Long:  "List all supported hooks formats and the events they support.",
	RunE:  runHooksList,
}

var hooksInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize git pre-commit hook",
	Long: `Initialize a git pre-commit hook that lints OpenAPI specifications.

The hook will automatically lint staged OpenAPI/Swagger files before each commit.
If any linting errors are found, the commit will be blocked.

Examples:
  # Install pre-commit hook with default settings
  api-style hooks init

  # Install with a specific profile
  api-style hooks init --profile azure

  # Install with a specific conformance level
  api-style hooks init --level silver

  # Overwrite existing hook
  api-style hooks init --force`,
	RunE: runHooksInit,
}

func init() {
	// Hooks command flags
	hooksCmd.Flags().StringVarP(&hooksFormat, "format", "f", "claude", "output format: claude, cursor, windsurf, all")
	hooksCmd.Flags().StringVarP(&hooksOutput, "output", "o", "", "output file path (default: format-specific)")
	hooksCmd.Flags().StringVarP(&hooksProfile, "profile", "p", "default", "style profile to use")
	hooksCmd.Flags().BoolVar(&hooksAutoLint, "auto-lint", true, "enable auto-linting on file save")
	hooksCmd.Flags().BoolVar(&hooksInjectCtx, "inject-context", false, "inject style context before prompts")
	hooksCmd.Flags().BoolVar(&hooksListFormats, "list", false, "list supported formats")

	// Generate subcommand inherits parent flags
	hooksGenerateCmd.Flags().StringVarP(&hooksFormat, "format", "f", "claude", "output format: claude, cursor, windsurf, all")
	hooksGenerateCmd.Flags().StringVarP(&hooksOutput, "output", "o", "", "output file path")
	hooksGenerateCmd.Flags().StringVarP(&hooksProfile, "profile", "p", "default", "style profile to use")
	hooksGenerateCmd.Flags().BoolVar(&hooksAutoLint, "auto-lint", true, "enable auto-linting")
	hooksGenerateCmd.Flags().BoolVar(&hooksInjectCtx, "inject-context", false, "inject style context")

	// Init subcommand flags
	hooksInitCmd.Flags().StringVarP(&hooksProfile, "profile", "p", "default", "style profile to use for linting")
	hooksInitCmd.Flags().StringVarP(&hooksLevel, "level", "l", "", "conformance level to enforce")
	hooksInitCmd.Flags().BoolVar(&hooksForce, "force", false, "overwrite existing pre-commit hook")

	hooksCmd.AddCommand(hooksGenerateCmd)
	hooksCmd.AddCommand(hooksListCmd)
	hooksCmd.AddCommand(hooksInitCmd)
}

func runHooks(cmd *cobra.Command, _ []string) error {
	if hooksListFormats {
		return runHooksList(cmd, nil)
	}

	cfg := &hooks.Config{
		Profile:       hooksProfile,
		AutoLint:      hooksAutoLint,
		InjectContext: hooksInjectCtx,
	}

	formats := []string{hooksFormat}
	if hooksFormat == "all" {
		formats = hooks.SupportedFormats()
	}

	defaultPaths := hooks.DefaultPaths()

	for _, format := range formats {
		outputPath := hooksOutput
		if outputPath == "" {
			outputPath = defaultPaths[format]
		}

		// For "all" format, use format-specific paths
		if hooksFormat == "all" {
			outputPath = defaultPaths[format]
		}

		// Create directory if needed
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := cfg.WriteToFile(outputPath, format); err != nil {
			return fmt.Errorf("failed to write %s hooks: %w", format, err)
		}

		fmt.Printf("Generated %s hooks: %s\n", format, outputPath)
	}

	return nil
}

func runHooksList(_ *cobra.Command, _ []string) error {
	fmt.Println("Supported Hooks Formats:")
	fmt.Println()

	support := hooks.EventSupport()
	for _, format := range hooks.SupportedFormats() {
		events := support[format]
		path := hooks.DefaultPaths()[format]

		fmt.Printf("  %s:\n", format)
		fmt.Printf("    Config: %s\n", path)
		fmt.Printf("    Events: %s\n", formatEvents(events))
		fmt.Println()
	}

	fmt.Println("Use --format <name> to generate configuration for a specific format.")
	fmt.Println("Use --format all to generate for all supported formats.")

	return nil
}

func runHooksInit(_ *cobra.Command, _ []string) error {
	opts := &hooks.InstallOptions{
		PreCommitOptions: &hooks.PreCommitOptions{
			Profile: hooksProfile,
			Level:   hooksLevel,
		},
		Force: hooksForce,
	}

	if err := hooks.InstallPreCommit(opts); err != nil {
		return fmt.Errorf("installing pre-commit hook: %w", err)
	}

	gitDir, _ := hooks.FindGitDir()
	hookPath := filepath.Join(gitDir, "hooks", "pre-commit")

	fmt.Printf("Installed pre-commit hook: %s\n", hookPath)
	fmt.Println()
	fmt.Println("The hook will automatically lint staged OpenAPI/Swagger files before commit.")
	fmt.Println("To skip the hook for a single commit, use: git commit --no-verify")

	return nil
}

func formatEvents(events []hookscore.Event) string {
	if len(events) == 0 {
		return "(none)"
	}
	strs := make([]string, len(events))
	for i, e := range events {
		strs[i] = string(e)
	}
	if len(strs) > 5 {
		return fmt.Sprintf("%s... (%d total)", strings.Join(strs[:5], ", "), len(strs))
	}
	return strings.Join(strs, ", ")
}
