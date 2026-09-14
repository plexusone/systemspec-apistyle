// mcp-api-style is an MCP server for API style specification linting and evaluation.
// It can also be used as a CLI tool for testing and scripting.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtime "github.com/plexusone/omniskill/mcp/server"
	apistylespec "github.com/plexusone/systemspec-apistyle"
	"github.com/plexusone/systemspec-apistyle/pkg/profile"
	"github.com/plexusone/systemspec-apistyle/skills/apistyle"
	"github.com/spf13/cobra"
)

const serverName = "mcp-api-style"

var serverVersion = "v" + apistylespec.Version

var (
	// API key flags
	anthropicAPIKey string

	// Output format flag
	outputFormat string

	// Tool parameter flags
	profileName    string
	categoryFilter string
	severityFilter string
	lintOnly       bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "mcp-api-style",
	Short: "MCP server and CLI for API style specification linting and evaluation",
	Long: `An MCP (Model Context Protocol) server for linting OpenAPI specifications against
style guidelines and evaluating them using both deterministic rules and LLM-based
semantic analysis.

Running without a subcommand starts the MCP server (default behavior).

The server provides the tools listed below.`,
	Example: `  # Start MCP server (default)
  mcp-api-style

  # Start MCP server with LLM evaluation enabled
  ANTHROPIC_API_KEY=sk-ant-... mcp-api-style

  # CLI: Lint an OpenAPI spec
  mcp-api-style lint --file openapi.yaml --profile azure

  # CLI: List available profiles
  mcp-api-style list-profiles

  # CLI: Explain a rule
  mcp-api-style explain-rule URI-001 --profile default`,
	SilenceUsage: true,
	RunE:         runServer,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long:  "Start the MCP server using stdio transport for communication with MCP clients.",
	RunE:  runServer,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("%s %s\n", serverName, serverVersion)
	},
}

// CLI commands that invoke tools directly
var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint an OpenAPI specification",
	Long:  "Lint an OpenAPI specification against API style rules using deterministic checks.",
	RunE:  runLint,
}

var listProfilesCmd = &cobra.Command{
	Use:   "list-profiles",
	Short: "List available style profiles",
	Long:  "List all available style profiles with their descriptions and rule counts.",
	RunE:  runListProfiles,
}

var listRulesCmd = &cobra.Command{
	Use:   "list-rules",
	Short: "List rules from a profile",
	Long:  "List all rules from a style profile with their IDs, titles, categories, and severities.",
	RunE:  runListRules,
}

var explainRuleCmd = &cobra.Command{
	Use:   "explain-rule <rule-id>",
	Short: "Explain a specific rule",
	Long:  "Get detailed explanation of a specific rule including rationale, examples, and references.",
	Args:  cobra.ExactArgs(1),
	RunE:  runExplainRule,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze an OpenAPI specification",
	Long:  "Combined analysis: deterministic linting + LLM evaluation with GO/NO-GO decision.",
	RunE:  runAnalyze,
}

var (
	specFile string
)

func init() {
	// Environment variable defaults
	if anthropicAPIKey == "" {
		anthropicAPIKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	// Derive the tool list from the skill so help text never drifts
	// from the actual MCP surface.
	var tools strings.Builder
	tools.WriteString("\n\nTools:\n")
	for _, tool := range apistyle.New().Tools() {
		desc := tool.Description()
		if i := strings.Index(desc, ". "); i > 0 {
			desc = desc[:i+1]
		}
		fmt.Fprintf(&tools, "  - %s: %s\n", tool.Name(), desc)
	}
	rootCmd.Long += strings.TrimRight(tools.String(), "\n")

	// Persistent flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json",
		"output format: json, pretty (default: json)")

	// Lint command flags
	lintCmd.Flags().StringVarP(&specFile, "file", "f", "", "path to OpenAPI specification file")
	lintCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "style profile to use")
	_ = lintCmd.MarkFlagRequired("file")

	// List rules command flags
	listRulesCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "style profile to use")
	listRulesCmd.Flags().StringVarP(&categoryFilter, "category", "c", "", "filter by category")
	listRulesCmd.Flags().StringVarP(&severityFilter, "severity", "s", "", "filter by severity")

	// Explain rule command flags
	explainRuleCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "style profile to use")

	// Analyze command flags
	analyzeCmd.Flags().StringVarP(&specFile, "file", "f", "", "path to OpenAPI specification file")
	analyzeCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "style profile to use")
	analyzeCmd.Flags().BoolVar(&lintOnly, "lint-only", false, "skip LLM evaluation")
	_ = analyzeCmd.MarkFlagRequired("file")

	// Add commands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(listProfilesCmd)
	rootCmd.AddCommand(listRulesCmd)
	rootCmd.AddCommand(explainRuleCmd)
	rootCmd.AddCommand(analyzeCmd)
}

func newSkill() *apistyle.Skill {
	var opts []apistyle.Option
	if anthropicAPIKey != "" {
		opts = append(opts, apistyle.WithAnthropicAPIKey(anthropicAPIKey))
	}
	return apistyle.New(opts...)
}

func runServer(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	// Create omniskill Runtime
	rt := runtime.New(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	// Create and initialize skill
	skill := newSkill()
	if err := skill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize apistyle skill: %w", err)
	}
	defer func() {
		_ = skill.Close()
	}()

	// Register skill with the runtime
	rt.RegisterSkill(skill)

	// Register MCP resources
	registerResources(rt)

	// Register MCP prompts
	registerPrompts(rt)

	// Run server with stdio transport
	if err := rt.ServeStdio(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// registerResources adds MCP resources for style profiles.
func registerResources(rt *runtime.Runtime) {
	// Resource: List all profiles
	rt.AddResource(&mcp.Resource{
		URI:         "apistyle://profiles",
		Name:        "API Style Profiles",
		Description: "List of all available API style profiles",
		MIMEType:    "application/json",
	}, func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		names, err := profile.ListBuiltin()
		if err != nil {
			return nil, fmt.Errorf("listing profiles: %w", err)
		}

		profiles := make([]map[string]any, 0, len(names))
		for _, name := range names {
			spec, err := profile.Load(name)
			if err != nil {
				continue
			}
			profiles = append(profiles, map[string]any{
				"name":        name,
				"description": spec.Description,
				"version":     spec.Version,
				"rule_count":  len(spec.Rules),
			})
		}

		data, err := json.MarshalIndent(map[string]any{"profiles": profiles}, "", "  ")
		if err != nil {
			return nil, err
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      "apistyle://profiles",
				MIMEType: "application/json",
				Text:     string(data),
			}},
		}, nil
	})

	// Resource template: Get specific profile
	rt.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "apistyle://profiles/{name}",
		Name:        "API Style Profile",
		Description: "Full definition of a specific API style profile",
		MIMEType:    "application/json",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Extract profile name from URI
		uri := req.Params.URI
		name := strings.TrimPrefix(uri, "apistyle://profiles/")
		if name == "" || name == uri {
			return nil, fmt.Errorf("invalid profile URI: %s", uri)
		}

		spec, err := profile.Load(name)
		if err != nil {
			return nil, fmt.Errorf("profile %q not found: %w", name, err)
		}

		data, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return nil, err
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      uri,
				MIMEType: "application/json",
				Text:     string(data),
			}},
		}, nil
	})
}

// registerPrompts adds MCP prompts for common API review tasks.
func registerPrompts(rt *runtime.Runtime) {
	// Prompt: Review API
	rt.AddPrompt(&mcp.Prompt{
		Name:        "review_api",
		Description: "Comprehensive API review against a style profile",
		Arguments: []*mcp.PromptArgument{
			{Name: "openapi_spec", Description: "The OpenAPI specification content (YAML or JSON)", Required: true},
			{Name: "profile", Description: "Style profile to use (default, azure, google, zalando)", Required: false},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		spec := req.Params.Arguments["openapi_spec"]
		prof := req.Params.Arguments["profile"]
		if prof == "" {
			prof = "default"
		}

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Review API against %s style profile", prof),
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: fmt.Sprintf(`Please review this OpenAPI specification against the "%s" API style guidelines.

First, use the 'lint' tool to check for deterministic rule violations.
Then, analyze the results and provide:
1. A summary of critical issues (errors)
2. A summary of warnings and suggestions
3. Specific recommendations for improvement
4. An overall assessment (GO/NO-GO for production)

OpenAPI Specification:
%s`, prof, spec),
					},
				},
			},
		}, nil
	})

	// Prompt: Fix Violations
	rt.AddPrompt(&mcp.Prompt{
		Name:        "fix_violations",
		Description: "Suggest fixes for specific API style violations",
		Arguments: []*mcp.PromptArgument{
			{Name: "openapi_spec", Description: "The OpenAPI specification content", Required: true},
			{Name: "violations", Description: "JSON array of violations to fix", Required: true},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		spec := req.Params.Arguments["openapi_spec"]
		violations := req.Params.Arguments["violations"]

		return &mcp.GetPromptResult{
			Description: "Generate fixes for API style violations",
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: fmt.Sprintf(`Please suggest specific fixes for these API style violations.

For each violation, provide:
1. The exact change needed
2. Before/after code snippets
3. Explanation of why this fix is recommended

Violations:
%s

OpenAPI Specification:
%s`, violations, spec),
					},
				},
			},
		}, nil
	})

	// Prompt: Explain Profile
	rt.AddPrompt(&mcp.Prompt{
		Name:        "explain_profile",
		Description: "Explain a style profile's philosophy and key rules",
		Arguments: []*mcp.PromptArgument{
			{Name: "profile", Description: "Profile name to explain", Required: true},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		prof := req.Params.Arguments["profile"]

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Explain the %s API style profile", prof),
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: fmt.Sprintf(`Please explain the "%s" API style profile.

First, use the 'list_rules' tool with profile="%s" to get all rules.
Then, provide:
1. The philosophy and goals of this style guide
2. Key categories of rules and their importance
3. The most critical rules (error severity)
4. Common patterns and conventions
5. When to use this profile vs others`, prof, prof),
					},
				},
			},
		}, nil
	})
}

func runLint(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	specContent, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	skill := newSkill()
	if err := skill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize skill: %w", err)
	}
	defer func() {
		_ = skill.Close()
	}()

	return runTool(ctx, skill, "lint", map[string]any{
		"openapi_spec": string(specContent),
		"profile":      profileName,
	})
}

func runListProfiles(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	skill := newSkill()
	if err := skill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize skill: %w", err)
	}
	defer func() {
		_ = skill.Close()
	}()

	return runTool(ctx, skill, "list_profiles", map[string]any{})
}

func runListRules(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	skill := newSkill()
	if err := skill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize skill: %w", err)
	}
	defer func() {
		_ = skill.Close()
	}()

	params := map[string]any{
		"profile": profileName,
	}
	if categoryFilter != "" {
		params["category"] = categoryFilter
	}
	if severityFilter != "" {
		params["severity"] = severityFilter
	}

	return runTool(ctx, skill, "list_rules", params)
}

func runExplainRule(_ *cobra.Command, args []string) error {
	ctx := context.Background()

	skill := newSkill()
	if err := skill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize skill: %w", err)
	}
	defer func() {
		_ = skill.Close()
	}()

	return runTool(ctx, skill, "explain_rule", map[string]any{
		"rule_id": args[0],
		"profile": profileName,
	})
}

func runAnalyze(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	specContent, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	skill := newSkill()
	if err := skill.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize skill: %w", err)
	}
	defer func() {
		_ = skill.Close()
	}()

	return runTool(ctx, skill, "analyze", map[string]any{
		"openapi_spec": string(specContent),
		"profile":      profileName,
		"lint_only":    lintOnly,
	})
}

func runTool(ctx context.Context, s *apistyle.Skill, toolName string, params map[string]any) error {
	for _, tool := range s.Tools() {
		if tool.Name() == toolName {
			result, err := tool.Call(ctx, params)
			if err != nil {
				return err
			}
			return outputResult(result)
		}
	}
	return fmt.Errorf("tool not found: %s", toolName)
}

func outputResult(result any) error {
	var data []byte
	var err error

	switch outputFormat {
	case "pretty":
		data, err = json.MarshalIndent(result, "", "  ")
	default:
		data, err = json.Marshal(result)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	fmt.Println(string(data))
	return nil
}
