// Package profile provides loading and management of API style profiles.
package profile

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
	"gopkg.in/yaml.v3"
)

//go:embed builtin/*.json
var builtinProfiles embed.FS

//go:embed builtin/exemplars/*.yaml
var builtinExemplars embed.FS

// Exemplar represents an exemplar OpenAPI specification.
type Exemplar struct {
	// Name is the exemplar identifier (e.g., "default-minimal")
	Name string `json:"name"`

	// Profile is the style profile this exemplar conforms to
	Profile string `json:"profile"`

	// Description is a human-readable description
	Description string `json:"description"`

	// Content is the raw OpenAPI specification content
	Content []byte `json:"-"`
}

// Loader loads and resolves API style profiles.
type Loader struct {
	// SearchPaths are directories to search for profiles.
	SearchPaths []string
}

// NewLoader creates a new profile loader.
func NewLoader() *Loader {
	return &Loader{
		SearchPaths: []string{
			".",
			"profiles",
			"~/.config/systemspec-apistyle/profiles",
		},
	}
}

// Load loads a profile by name.
// It first checks built-in profiles, then searches the filesystem.
func (l *Loader) Load(name string) (*types.APIStyleSpec, error) {
	// Try built-in profiles first
	spec, err := l.loadBuiltin(name)
	if err == nil {
		return spec, nil
	}

	// Try filesystem
	spec, err = l.loadFromFS(name)
	if err == nil {
		return spec, nil
	}

	return nil, fmt.Errorf("profile %q not found", name)
}

// loadBuiltin loads a built-in profile.
func (l *Loader) loadBuiltin(name string) (*types.APIStyleSpec, error) {
	filename := fmt.Sprintf("builtin/%s.api-style.json", name)
	data, err := builtinProfiles.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("built-in profile %q not found: %w", name, err)
	}

	return l.parseSpec(data, ".json")
}

// loadFromFS loads a profile from the filesystem.
func (l *Loader) loadFromFS(name string) (*types.APIStyleSpec, error) {
	// Try exact path first
	if filepath.IsAbs(name) || filepath.Ext(name) != "" {
		return l.loadFile(name)
	}

	// Search in paths
	for _, searchPath := range l.SearchPaths {
		candidates := []string{
			// JSON variants
			filepath.Join(searchPath, name+".api-style.json"),
			filepath.Join(searchPath, name+".json"),
			filepath.Join(searchPath, name, "api-style.json"),
			// YAML variants
			filepath.Join(searchPath, name+".api-style.yaml"),
			filepath.Join(searchPath, name+".api-style.yml"),
			filepath.Join(searchPath, name+".yaml"),
			filepath.Join(searchPath, name+".yml"),
			filepath.Join(searchPath, name, "api-style.yaml"),
			filepath.Join(searchPath, name, "api-style.yml"),
		}

		for _, candidate := range candidates {
			spec, err := l.loadFile(candidate)
			if err == nil {
				return spec, nil
			}
		}
	}

	return nil, fmt.Errorf("profile %q not found in search paths", name)
}

// loadFile loads a profile from a specific file path.
func (l *Loader) loadFile(path string) (*types.APIStyleSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	spec, err := l.parseSpec(data, ext)
	if err != nil {
		return nil, err
	}

	// Validate the loaded spec
	if err := l.validateSpec(spec); err != nil {
		return nil, fmt.Errorf("validating profile %q: %w", path, err)
	}

	return spec, nil
}

// parseSpec parses data into an APIStyleSpec based on file extension.
func (l *Loader) parseSpec(data []byte, ext string) (*types.APIStyleSpec, error) {
	var spec types.APIStyleSpec

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parsing YAML profile: %w", err)
		}
	case ".json", "":
		if err := json.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("parsing JSON profile: %w", err)
		}
	default:
		// Try JSON first, then YAML
		if err := json.Unmarshal(data, &spec); err != nil {
			if err2 := yaml.Unmarshal(data, &spec); err2 != nil {
				return nil, fmt.Errorf("parsing profile (tried JSON and YAML): JSON error: %w", err)
			}
		}
	}

	return &spec, nil
}

// validateSpec validates a loaded profile spec.
func (l *Loader) validateSpec(spec *types.APIStyleSpec) error {
	if spec.Name == "" {
		return fmt.Errorf("profile name is required")
	}
	if spec.Version == "" {
		return fmt.Errorf("profile version is required")
	}
	// Rules array can be empty if profile only extends others
	return nil
}

// LoadFile loads a profile from a specific file path.
func LoadFile(path string) (*types.APIStyleSpec, error) {
	loader := NewLoader()
	return loader.loadFile(path)
}

// Load loads a profile by name using default search paths.
func Load(name string) (*types.APIStyleSpec, error) {
	loader := NewLoader()
	return loader.Load(name)
}

// ListBuiltin returns the names of all built-in profiles.
func ListBuiltin() ([]string, error) {
	entries, err := builtinProfiles.ReadDir("builtin")
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Remove .api-style.json suffix (16 chars)
		const suffix = ".api-style.json"
		if len(name) > len(suffix) && name[len(name)-len(suffix):] == suffix {
			names = append(names, name[:len(name)-len(suffix)])
		}
	}

	return names, nil
}

// ListExemplars returns all built-in exemplar specifications.
func ListExemplars() ([]Exemplar, error) {
	entries, err := builtinExemplars.ReadDir("builtin/exemplars")
	if err != nil {
		return nil, err
	}

	var exemplars []Exemplar
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}

		// Parse exemplar info from filename (format: profile-variant.yaml)
		baseName := strings.TrimSuffix(name, ".yaml")
		parts := strings.SplitN(baseName, "-", 2)
		profile := parts[0]
		variant := ""
		if len(parts) > 1 {
			variant = parts[1]
		}

		// Load content
		content, err := builtinExemplars.ReadFile("builtin/exemplars/" + name)
		if err != nil {
			continue
		}

		// Extract description from info.description
		description := extractDescription(content)

		exemplars = append(exemplars, Exemplar{
			Name:        baseName,
			Profile:     profile,
			Description: description,
			Content:     content,
		})
		_ = variant // for future use
	}

	return exemplars, nil
}

// ListExemplarsForProfile returns exemplars for a specific profile.
func ListExemplarsForProfile(profileName string) ([]Exemplar, error) {
	all, err := ListExemplars()
	if err != nil {
		return nil, err
	}

	var filtered []Exemplar
	for _, e := range all {
		if e.Profile == profileName {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

// GetExemplar loads a specific exemplar by name.
func GetExemplar(name string) (*Exemplar, error) {
	// Try exact match first
	filename := "builtin/exemplars/" + name + ".yaml"
	content, err := builtinExemplars.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("exemplar %q not found: %w", name, err)
	}

	// Parse info
	parts := strings.SplitN(name, "-", 2)
	profile := parts[0]

	return &Exemplar{
		Name:        name,
		Profile:     profile,
		Description: extractDescription(content),
		Content:     content,
	}, nil
}

// extractDescription extracts the description from OpenAPI content.
func extractDescription(content []byte) string {
	var spec struct {
		Info struct {
			Description string `yaml:"description"`
			Title       string `yaml:"title"`
		} `yaml:"info"`
	}

	if err := yaml.Unmarshal(content, &spec); err != nil {
		return ""
	}

	// Return first paragraph of description, or title if no description
	desc := spec.Info.Description
	if desc == "" {
		return spec.Info.Title
	}

	// Take first paragraph (up to double newline or 200 chars)
	if idx := strings.Index(desc, "\n\n"); idx > 0 && idx < 200 {
		desc = desc[:idx]
	} else if len(desc) > 200 {
		desc = desc[:200] + "..."
	}

	return strings.TrimSpace(desc)
}
