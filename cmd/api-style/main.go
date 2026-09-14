// Command api-style is a CLI for API style specification linting and evaluation.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	apistylespec "github.com/plexusone/systemspec-apistyle"
)

var rootCmd = &cobra.Command{
	Use:   "api-style",
	Short: "API style specification linter and evaluator",
	Long: `api-style is a tool for linting OpenAPI specifications against
style guidelines and evaluating them using both deterministic rules
and LLM-based semantic analysis.

It supports machine-readable style specifications that can generate
human documentation, linting rules, and AI agent instructions from
a single source of truth.`,
	Version: apistylespec.Version,
}

func init() {
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(evaluateCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(exemplarCmd)
	rootCmd.AddCommand(patternCmd)
	rootCmd.AddCommand(hooksCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(scoreProfileCmd)
	rootCmd.AddCommand(suggestCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("api-style version %s\n", apistylespec.Version)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
