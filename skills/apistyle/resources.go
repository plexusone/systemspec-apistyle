package apistyle

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/plexusone/systemspec-apistyle/pkg/generate"
	"github.com/plexusone/systemspec-apistyle/pkg/judge"
	"github.com/plexusone/systemspec-apistyle/pkg/profile"
)

// ResourceRegistrar can register MCP resources.
type ResourceRegistrar interface {
	AddResource(res *mcp.Resource, h mcp.ResourceHandler)
	AddResourceTemplate(t *mcp.ResourceTemplate, h mcp.ResourceHandler)
}

// RegisterResources registers API style MCP resources on the given registrar.
// This enables MCP clients to access profiles, exemplars, patterns, and rubrics.
//
// Registered resources:
//   - profile://{name} - Style profile JSON
//   - exemplar://{name} - Exemplar OpenAPI spec
//   - pattern://{profile}/{id} - Pattern definition
//   - rubric://{profile}/{mode} - Generation rubric (mode: evaluation/generation)
func RegisterResources(r ResourceRegistrar) {
	// List available profiles as a resource
	r.AddResource(&mcp.Resource{
		URI:         "apistyle://profiles",
		Name:        "profiles",
		Description: "List of available API style profiles",
		MIMEType:    "application/json",
	}, handleProfileList)

	// List available exemplars as a resource
	r.AddResource(&mcp.Resource{
		URI:         "apistyle://exemplars",
		Name:        "exemplars",
		Description: "List of available exemplar OpenAPI specifications",
		MIMEType:    "application/json",
	}, handleExemplarList)

	// Profile template - access any profile by name
	r.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "apistyle://profile/{name}",
		Name:        "profile",
		Description: "API style profile specification",
		MIMEType:    "application/json",
	}, handleProfile)

	// Exemplar template - access any exemplar by name
	r.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "apistyle://exemplar/{name}",
		Name:        "exemplar",
		Description: "Exemplar OpenAPI specification that conforms to a style profile",
		MIMEType:    "application/x-yaml",
	}, handleExemplar)

	// Pattern template - access patterns by profile and ID
	r.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "apistyle://pattern/{profile}/{id}",
		Name:        "pattern",
		Description: "API design pattern definition",
		MIMEType:    "application/json",
	}, handlePattern)

	// Patterns list template - list patterns for a profile
	r.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "apistyle://patterns/{profile}",
		Name:        "patterns",
		Description: "List of API design patterns for a profile",
		MIMEType:    "application/json",
	}, handlePatternList)

	// Rubric template - access rubric by profile and mode
	r.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "apistyle://rubric/{profile}/{mode}",
		Name:        "rubric",
		Description: "Structured rubric for LLM evaluation or generation guidance",
		MIMEType:    "application/json",
	}, handleRubric)
}

// handleProfileList returns the list of available profiles.
func handleProfileList(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	names, err := profile.ListBuiltin()
	if err != nil {
		return nil, fmt.Errorf("listing profiles: %w", err)
	}

	result := make([]map[string]any, 0, len(names))
	for _, name := range names {
		spec, err := profile.Load(name)
		if err != nil {
			continue
		}
		result = append(result, map[string]any{
			"name":        name,
			"description": spec.Description,
			"version":     spec.Version,
			"ruleCount":   len(spec.Rules),
		})
	}

	data, err := json.MarshalIndent(map[string]any{"profiles": result}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling profiles: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      "apistyle://profiles",
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

// handleExemplarList returns the list of available exemplars.
func handleExemplarList(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	exemplars, err := profile.ListExemplars()
	if err != nil {
		return nil, fmt.Errorf("listing exemplars: %w", err)
	}

	result := make([]map[string]any, 0, len(exemplars))
	for _, e := range exemplars {
		result = append(result, map[string]any{
			"name":        e.Name,
			"profile":     e.Profile,
			"description": e.Description,
		})
	}

	data, err := json.MarshalIndent(map[string]any{"exemplars": result}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling exemplars: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      "apistyle://exemplars",
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

// handleProfile returns a specific profile's JSON content.
func handleProfile(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract profile name from URI: apistyle://profile/{name}
	name := extractURIParam(req.Params.URI, "apistyle://profile/")
	if name == "" {
		return nil, fmt.Errorf("profile name required")
	}

	spec, err := profile.Load(name)
	if err != nil {
		return nil, fmt.Errorf("loading profile %q: %w", name, err)
	}

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling profile: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

// handleExemplar returns a specific exemplar's content.
func handleExemplar(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract exemplar name from URI: apistyle://exemplar/{name}
	name := extractURIParam(req.Params.URI, "apistyle://exemplar/")
	if name == "" {
		return nil, fmt.Errorf("exemplar name required")
	}

	exemplar, err := profile.GetExemplar(name)
	if err != nil {
		return nil, fmt.Errorf("loading exemplar %q: %w", name, err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/x-yaml",
			Text:     string(exemplar.Content),
		}},
	}, nil
}

// handlePatternList returns the list of patterns for a profile.
func handlePatternList(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract profile name from URI: apistyle://patterns/{profile}
	profileName := extractURIParam(req.Params.URI, "apistyle://patterns/")
	if profileName == "" {
		profileName = "default"
	}

	spec, err := profile.Load(profileName)
	if err != nil {
		return nil, fmt.Errorf("loading profile %q: %w", profileName, err)
	}

	patterns := make([]map[string]any, 0, len(spec.Patterns))
	for _, p := range spec.Patterns {
		patterns = append(patterns, map[string]any{
			"id":       p.ID,
			"name":     p.Name,
			"category": p.Category,
			"summary":  p.Summary,
		})
	}

	data, err := json.MarshalIndent(map[string]any{
		"profile":  profileName,
		"patterns": patterns,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling patterns: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

// handlePattern returns a specific pattern's definition.
func handlePattern(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract profile and pattern ID from URI: apistyle://pattern/{profile}/{id}
	path := extractURIParam(req.Params.URI, "apistyle://pattern/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("pattern URI must be apistyle://pattern/{profile}/{id}")
	}

	profileName := parts[0]
	patternID := parts[1]

	spec, err := profile.Load(profileName)
	if err != nil {
		return nil, fmt.Errorf("loading profile %q: %w", profileName, err)
	}

	pattern := spec.GetPattern(patternID)
	if pattern == nil {
		return nil, fmt.Errorf("pattern %q not found in profile %q", patternID, profileName)
	}

	data, err := json.MarshalIndent(pattern, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling pattern: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

// handleRubric returns a structured rubric for evaluation or generation.
func handleRubric(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract profile and mode from URI: apistyle://rubric/{profile}/{mode}
	path := extractURIParam(req.Params.URI, "apistyle://rubric/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("rubric URI must be apistyle://rubric/{profile}/{mode}")
	}

	profileName := parts[0]
	mode := parts[1]

	spec, err := profile.Load(profileName)
	if err != nil {
		return nil, fmt.Errorf("loading profile %q: %w", profileName, err)
	}

	var data []byte

	switch strings.ToLower(mode) {
	case "generation":
		rubric := generate.GenerationRubricFromSpec(spec)
		data, err = rubric.ToJSON()
	case "evaluation":
		rubricSet := judge.GenerateRubricSet(spec)
		data, err = judge.RubricSetToJSON(rubricSet)
	default:
		return nil, fmt.Errorf("unknown rubric mode %q (use 'evaluation' or 'generation')", mode)
	}

	if err != nil {
		return nil, fmt.Errorf("generating rubric: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		}},
	}, nil
}

// extractURIParam extracts the parameter portion from a URI after the prefix.
func extractURIParam(uri, prefix string) string {
	if !strings.HasPrefix(uri, prefix) {
		return ""
	}
	return strings.TrimPrefix(uri, prefix)
}
