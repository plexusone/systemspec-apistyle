package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigFileNames are the supported config file names in priority order.
var ConfigFileNames = []string{
	".api-style.yaml",
	".api-style.yml",
	"api-style.yaml",
	"api-style.yml",
	".api-style.json",
	"api-style.json",
}

// Loader loads systemspec-apistyle configuration files.
type Loader struct {
	// SearchPaths are directories to search for config files.
	SearchPaths []string
}

// NewLoader creates a new config loader.
func NewLoader() *Loader {
	return &Loader{
		SearchPaths: []string{
			".",
		},
	}
}

// Load searches for and loads a configuration file.
// Returns nil config (not an error) if no config file is found.
func (l *Loader) Load() (*Config, error) {
	for _, searchPath := range l.SearchPaths {
		for _, name := range ConfigFileNames {
			path := filepath.Join(searchPath, name)
			if _, err := os.Stat(path); err == nil {
				return l.LoadFile(path)
			}
		}
	}
	return nil, nil
}

// LoadFile loads configuration from a specific file path.
func (l *Loader) LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	cfg, err := l.parse(data, ext)
	if err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}

// parse parses configuration data based on file extension.
func (l *Loader) parse(data []byte, ext string) (*Config, error) {
	var cfg Config

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parsing JSON: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			if err2 := json.Unmarshal(data, &cfg); err2 != nil {
				return nil, fmt.Errorf("parsing config (tried YAML and JSON): %w", err)
			}
		}
	}

	return &cfg, nil
}

// Load loads configuration using default search paths.
// Returns nil config (not an error) if no config file is found.
func Load() (*Config, error) {
	loader := NewLoader()
	return loader.Load()
}

// LoadFile loads configuration from a specific file path.
func LoadFile(path string) (*Config, error) {
	loader := NewLoader()
	return loader.LoadFile(path)
}

// LoadOrDefault loads configuration or returns default if not found.
func LoadOrDefault() (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return Default(), nil
	}
	return cfg, nil
}
