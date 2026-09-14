// Command schemagen generates JSON Schema from Go types.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Find the schema directory
	schemaDir := "schema"
	if len(os.Args) > 1 {
		schemaDir = os.Args[1]
	}

	//nolint:gosec // G703: Path from CLI arg
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		return fmt.Errorf("creating schema directory: %w", err)
	}

	// Generate schemas for main types
	schemas := []struct {
		name     string
		typ      any
		schemaID string
	}{
		{
			name:     "api-style-spec.schema.json",
			typ:      &types.APIStyleSpec{},
			schemaID: "https://plexusone.dev/systemspec-apistyle/schema/v1/api-style-spec.schema.json",
		},
		{
			name:     "lint-report.schema.json",
			typ:      &types.LintReport{},
			schemaID: "https://plexusone.dev/systemspec-apistyle/schema/v1/lint-report.schema.json",
		},
	}

	for _, s := range schemas {
		if err := generateSchema(schemaDir, s.name, s.typ, s.schemaID); err != nil {
			return fmt.Errorf("generating %s: %w", s.name, err)
		}
		fmt.Printf("Generated %s\n", filepath.Join(schemaDir, s.name))
	}

	return nil
}

func generateSchema(dir, filename string, typ any, schemaID string) error {
	r := &jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true,
		ExpandedStruct:             true,
	}

	schema := r.Reflect(typ)
	schema.ID = jsonschema.ID(schemaID)
	schema.Version = "https://json-schema.org/draft/2020-12/schema"

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling schema: %w", err)
	}

	path := filepath.Join(dir, filename)
	//nolint:gosec // G703: Path from CLI arg
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing schema: %w", err)
	}

	return nil
}
