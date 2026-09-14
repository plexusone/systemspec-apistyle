package generate

import (
	"encoding/json"
	"sort"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// GenerationRubric contains guidance for AI agents generating OpenAPI specs.
type GenerationRubric struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Version     string              `json:"version,omitempty"`
	Phases      []GenerationPhase   `json:"phases"`
	Metadata    *GenerationMetadata `json:"metadata,omitempty"`
}

// GenerationPhase groups directives by generation phase.
type GenerationPhase struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Order       int                   `json:"order"`
	Directives  []GenerationDirective `json:"directives"`
}

// GenerationDirective is a single generation instruction.
type GenerationDirective struct {
	RuleID    string             `json:"ruleId"`
	Title     string             `json:"title"`
	Priority  int                `json:"priority"`
	Prompt    string             `json:"prompt"`
	Template  string             `json:"template,omitempty"`
	Checklist []string           `json:"checklist,omitempty"`
	Examples  []DirectiveExample `json:"examples,omitempty"`
	Required  bool               `json:"required"`
}

// DirectiveExample shows good and bad patterns.
type DirectiveExample struct {
	Good string `json:"good"`
	Bad  string `json:"bad,omitempty"`
}

// GenerationMetadata contains rubric metadata.
type GenerationMetadata struct {
	Author     string `json:"author,omitempty"`
	TotalRules int    `json:"totalRules"`
}

// phaseMapping maps categories to generation phases.
var phaseMapping = map[string]struct {
	phaseID string
	order   int
}{
	// Info phase (first)
	"general":    {"info", 1},
	"versioning": {"info", 1},

	// Paths phase (second)
	"urls":         {"paths", 2},
	"uri-design":   {"paths", 2},
	"http-methods": {"paths", 2},
	"operations":   {"paths", 2},

	// Request/Response phase (third)
	"request-response": {"request-response", 3},
	"http-status":      {"request-response", 3},
	"headers":          {"request-response", 3},
	"errors":           {"request-response", 3},
	"pagination":       {"request-response", 3},
	"filtering":        {"request-response", 3},

	// Schemas phase (fourth)
	"naming":      {"schemas", 4},
	"schema":      {"schemas", 4},
	"schemas":     {"schemas", 4},
	"types":       {"schemas", 4},
	"sdk-codegen": {"schemas", 4},

	// Security phase (fifth)
	"security":       {"security", 5},
	"authentication": {"security", 5},
}

// phaseNames provides human-readable phase names.
var phaseNames = map[string]string{
	"info":             "API Info & Metadata",
	"paths":            "URL Paths & Operations",
	"request-response": "Request & Response",
	"schemas":          "Data Schemas",
	"security":         "Security",
	"other":            "Other Guidelines",
}

// GenerationRubricFromSpec builds a generation rubric from an APIStyleSpec.
func GenerationRubricFromSpec(spec *types.APIStyleSpec) *GenerationRubric {
	rubric := &GenerationRubric{
		Name:        spec.Name,
		Description: spec.Description,
		Version:     spec.Version,
	}

	// Collect directives by phase
	phaseDirectives := make(map[string][]GenerationDirective)

	for _, rule := range spec.Rules {
		// Skip rules without generation guidance
		if rule.Generate == nil || rule.Generate.Prompt == "" {
			continue
		}

		directive := GenerationDirective{
			RuleID:    rule.ID,
			Title:     rule.Title,
			Priority:  rule.Generate.Priority,
			Prompt:    rule.Generate.Prompt,
			Template:  rule.Generate.Template,
			Checklist: rule.Generate.Checklist,
			Required:  rule.Severity == types.SeverityError,
		}

		// Convert examples from GenerationExample struct
		for _, ex := range rule.Generate.Examples {
			directive.Examples = append(directive.Examples, DirectiveExample{
				Good: ex.OpenAPI,
				Bad:  ex.Description, // Use description as context
			})
		}

		// Determine phase from category
		phaseID := "other"
		if mapping, ok := phaseMapping[rule.Category]; ok {
			phaseID = mapping.phaseID
		}

		phaseDirectives[phaseID] = append(phaseDirectives[phaseID], directive)
	}

	// Build phases with sorted directives
	phaseOrder := []string{"info", "paths", "request-response", "schemas", "security", "other"}

	for i, phaseID := range phaseOrder {
		directives := phaseDirectives[phaseID]
		if len(directives) == 0 {
			continue
		}

		// Sort directives by priority (higher first)
		sort.Slice(directives, func(a, b int) bool {
			return directives[a].Priority > directives[b].Priority
		})

		phaseName := phaseNames[phaseID]
		if phaseName == "" {
			phaseName = phaseID
		}

		phase := GenerationPhase{
			ID:         phaseID,
			Name:       phaseName,
			Order:      i + 1,
			Directives: directives,
		}

		rubric.Phases = append(rubric.Phases, phase)
	}

	// Add metadata
	if spec.Metadata != nil {
		rubric.Metadata = &GenerationMetadata{
			Author:     spec.Metadata.Author,
			TotalRules: len(spec.Rules),
		}
	} else {
		rubric.Metadata = &GenerationMetadata{
			TotalRules: len(spec.Rules),
		}
	}

	return rubric
}

// ToJSON serializes the generation rubric to JSON.
func (r *GenerationRubric) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
