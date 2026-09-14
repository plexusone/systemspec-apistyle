// Package types defines the core data types for systemspec-apistyle.
//
// These Go types are the source of truth for the systemspec-apistyle format.
// JSON Schema is generated from these types using invopop/jsonschema.
//
// Main types:
//   - APIStyleSpec: Root type for a style specification
//   - Rule: Individual style rule with enforcement and judge criteria
//   - LintReport: Results from deterministic linting
//   - Violation: A single rule violation
package types

//go:generate go run ../../cmd/schemagen/main.go ../../schema
