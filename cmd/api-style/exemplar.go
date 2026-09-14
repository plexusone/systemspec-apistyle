package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-apistyle/pkg/profile"
)

var (
	exemplarProfile string
)

var exemplarCmd = &cobra.Command{
	Use:   "exemplar <command>",
	Short: "Work with exemplar OpenAPI specifications",
	Long: `View and copy exemplar OpenAPI specifications that conform to style profiles.

Exemplars are reference implementations that demonstrate best practices
and can be used as starting points for new APIs.

Examples:
  api-style exemplar list
  api-style exemplar list --profile default
  api-style exemplar show default-minimal
  api-style exemplar copy default-minimal ./openapi.yaml`,
}

var exemplarListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available exemplars",
	Long: `List all available exemplar specifications.

Use --profile to filter by style profile.

Examples:
  api-style exemplar list
  api-style exemplar list --profile default`,
	RunE: runExemplarList,
}

var exemplarShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show exemplar content",
	Long: `Display the content of an exemplar specification.

Examples:
  api-style exemplar show default-minimal
  api-style exemplar show default-crud`,
	Args: cobra.ExactArgs(1),
	RunE: runExemplarShow,
}

var exemplarCopyCmd = &cobra.Command{
	Use:   "copy <name> <destination>",
	Short: "Copy exemplar to a file",
	Long: `Copy an exemplar specification to a local file.

Examples:
  api-style exemplar copy default-minimal ./openapi.yaml
  api-style exemplar copy default-crud ./api/orders.yaml`,
	Args: cobra.ExactArgs(2),
	RunE: runExemplarCopy,
}

func init() {
	exemplarListCmd.Flags().StringVarP(&exemplarProfile, "profile", "p", "", "Filter by profile name")

	exemplarCmd.AddCommand(exemplarListCmd)
	exemplarCmd.AddCommand(exemplarShowCmd)
	exemplarCmd.AddCommand(exemplarCopyCmd)
}

func runExemplarList(_ *cobra.Command, _ []string) error {
	var exemplars []profile.Exemplar
	var err error

	if exemplarProfile != "" {
		exemplars, err = profile.ListExemplarsForProfile(exemplarProfile)
	} else {
		exemplars, err = profile.ListExemplars()
	}
	if err != nil {
		return fmt.Errorf("listing exemplars: %w", err)
	}

	if len(exemplars) == 0 {
		if exemplarProfile != "" {
			fmt.Printf("No exemplars found for profile %q\n", exemplarProfile)
		} else {
			fmt.Println("No exemplars found")
		}
		return nil
	}

	fmt.Println("Available Exemplars:")
	fmt.Println()

	// Group by profile if no filter
	if exemplarProfile == "" {
		byProfile := make(map[string][]profile.Exemplar)
		for _, e := range exemplars {
			byProfile[e.Profile] = append(byProfile[e.Profile], e)
		}

		for prof, exs := range byProfile {
			fmt.Printf("[%s]\n", prof)
			for _, e := range exs {
				printExemplarSummary(e)
			}
			fmt.Println()
		}
	} else {
		for _, e := range exemplars {
			printExemplarSummary(e)
		}
	}

	return nil
}

func printExemplarSummary(e profile.Exemplar) {
	// Truncate description to first line
	desc := e.Description
	if idx := strings.Index(desc, "\n"); idx > 0 {
		desc = desc[:idx]
	}
	if len(desc) > 60 {
		desc = desc[:57] + "..."
	}

	fmt.Printf("  %-20s %s\n", e.Name, desc)
}

func runExemplarShow(_ *cobra.Command, args []string) error {
	name := args[0]

	exemplar, err := profile.GetExemplar(name)
	if err != nil {
		return fmt.Errorf("loading exemplar: %w", err)
	}

	fmt.Print(string(exemplar.Content))
	return nil
}

func runExemplarCopy(_ *cobra.Command, args []string) error {
	name := args[0]
	dest := args[1]

	exemplar, err := profile.GetExemplar(name)
	if err != nil {
		return fmt.Errorf("loading exemplar: %w", err)
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(dest)
	if destDir != "." && destDir != "" {
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
	}

	// Write file with user read/write permissions
	if err := os.WriteFile(dest, exemplar.Content, 0o600); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Printf("Copied %s to %s\n", name, dest)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %s for your API\n", dest)
	fmt.Printf("  2. Run: api-style lint %s --profile %s\n", dest, exemplar.Profile)

	return nil
}
