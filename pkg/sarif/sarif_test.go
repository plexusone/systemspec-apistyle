package sarif

import (
	"encoding/json"
	"testing"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

func TestFromLintReport_Basic(t *testing.T) {
	report := &types.LintReport{
		Status: types.StatusFail,
		Summary: &types.ViolationSummary{
			Errors:   1,
			Warnings: 1,
			Total:    2,
		},
		Violations: []types.Violation{
			{
				RuleID:   "URI-001",
				Severity: types.SeverityError,
				Message:  "Use plural resource names",
				Path:     "$.paths./user",
				Line:     10,
				Column:   3,
			},
			{
				RuleID:   "DOC-001",
				Severity: types.SeverityWarn,
				Message:  "Missing description",
				Path:     "$.info.description",
			},
		},
	}

	log := FromLintReport(report, nil)

	if log.Schema != SchemaURI {
		t.Errorf("Schema = %q, want %q", log.Schema, SchemaURI)
	}
	if log.Version != Version {
		t.Errorf("Version = %q, want %q", log.Version, Version)
	}
	if len(log.Runs) != 1 {
		t.Fatalf("len(Runs) = %d, want 1", len(log.Runs))
	}

	run := log.Runs[0]
	if run.Tool.Driver.Name != "api-style" {
		t.Errorf("Tool.Driver.Name = %q, want %q", run.Tool.Driver.Name, "api-style")
	}
	if len(run.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2", len(run.Results))
	}
}

func TestFromLintReport_WithRules(t *testing.T) {
	report := &types.LintReport{
		Status:  types.StatusFail,
		Summary: &types.ViolationSummary{Errors: 1, Total: 1},
		Violations: []types.Violation{
			{
				RuleID:   "URI-001",
				Severity: types.SeverityError,
				Message:  "Use plural resource names",
				Path:     "/openapi.yaml",
				Line:     10,
			},
		},
	}

	rules := map[string]*types.Rule{
		"URI-001": {
			ID:        "URI-001",
			Title:     "Plural Resource Names",
			Category:  "uri-design",
			Severity:  types.SeverityError,
			Rationale: "Plural names are more consistent and RESTful.",
		},
	}

	opts := &Options{
		ToolName:     "api-style",
		ToolVersion:  "1.0.0",
		IncludeRules: true,
		Rules:        rules,
	}

	log := FromLintReport(report, opts)
	run := log.Runs[0]

	if len(run.Tool.Driver.Rules) != 1 {
		t.Fatalf("len(Rules) = %d, want 1", len(run.Tool.Driver.Rules))
	}

	rule := run.Tool.Driver.Rules[0]
	if rule.ID != "URI-001" {
		t.Errorf("Rule.ID = %q, want %q", rule.ID, "URI-001")
	}
	if rule.Name != "Plural Resource Names" {
		t.Errorf("Rule.Name = %q, want %q", rule.Name, "Plural Resource Names")
	}
	if rule.ShortDescr == nil || rule.ShortDescr.Text != "Plural Resource Names" {
		t.Error("Rule.ShortDescr not set correctly")
	}
	if rule.FullDescr == nil || rule.FullDescr.Text != "Plural names are more consistent and RESTful." {
		t.Error("Rule.FullDescr not set correctly")
	}
}

func TestSeverityToLevel(t *testing.T) {
	tests := []struct {
		severity types.Severity
		expected Level
	}{
		{types.SeverityError, LevelError},
		{types.SeverityWarn, LevelWarning},
		{types.SeverityInfo, LevelNote},
		{types.SeverityHint, LevelNote},
	}

	for _, tt := range tests {
		result := severityToLevel(tt.severity)
		if result != tt.expected {
			t.Errorf("severityToLevel(%v) = %v, want %v", tt.severity, result, tt.expected)
		}
	}
}

func TestNormalizeURI(t *testing.T) {
	tests := []struct {
		path     string
		baseURI  string
		expected string
	}{
		{"/Users/test/api.yaml", "", "file:///Users/test/api.yaml"},
		{"api.yaml", "file:///project", "file:///project/api.yaml"},
		{"file:///existing/path.yaml", "", "file:///existing/path.yaml"},
		{"https://example.com/api.yaml", "", "https://example.com/api.yaml"},
		{"$.paths./users", "", ""}, // JSON path should return empty
		{"$[0]", "", ""},           // JSON path should return empty
	}

	for _, tt := range tests {
		result := normalizeURI(tt.path, tt.baseURI)
		if result != tt.expected {
			t.Errorf("normalizeURI(%q, %q) = %q, want %q", tt.path, tt.baseURI, result, tt.expected)
		}
	}
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"api.yaml", "application/x-yaml"},
		{"api.yml", "application/x-yaml"},
		{"api.json", "application/json"},
		{"api.txt", "text/plain"},
		{"api", "text/plain"},
	}

	for _, tt := range tests {
		result := getMimeType(tt.filename)
		if result != tt.expected {
			t.Errorf("getMimeType(%q) = %q, want %q", tt.filename, result, tt.expected)
		}
	}
}

func TestLogMarshal(t *testing.T) {
	log := &Log{
		Schema:  SchemaURI,
		Version: Version,
		Runs: []Run{
			{
				Tool: Tool{
					Driver: ToolComponent{
						Name:    "test-tool",
						Version: "1.0.0",
					},
				},
				Results: []Result{},
			},
		},
	}

	// Test pretty print
	data, err := log.Marshal(true)
	if err != nil {
		t.Fatalf("Marshal(true) error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse marshaled JSON: %v", err)
	}

	if parsed["version"] != Version {
		t.Errorf("Parsed version = %v, want %v", parsed["version"], Version)
	}

	// Test compact
	dataCompact, err := log.Marshal(false)
	if err != nil {
		t.Fatalf("Marshal(false) error = %v", err)
	}

	if len(dataCompact) >= len(data) {
		t.Error("Compact output should be smaller than pretty output")
	}
}

func TestLogString(t *testing.T) {
	log := &Log{
		Schema:  SchemaURI,
		Version: Version,
		Runs:    []Run{},
	}

	str := log.String()
	if str == "" {
		t.Error("String() returned empty string")
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(str), &parsed); err != nil {
		t.Fatalf("String() output is not valid JSON: %v", err)
	}
}

func TestFormatLintReport(t *testing.T) {
	report := &types.LintReport{
		Status:  types.StatusPass,
		Summary: &types.ViolationSummary{},
		Violations: []types.Violation{
			{
				RuleID:   "TEST-001",
				Severity: types.SeverityWarn,
				Message:  "Test violation",
				Path:     "test.yaml",
				Line:     5,
			},
		},
	}

	output, err := FormatLintReport(report, nil)
	if err != nil {
		t.Fatalf("FormatLintReport() error = %v", err)
	}

	if output == "" {
		t.Error("FormatLintReport() returned empty string")
	}

	// Verify it's valid JSON
	var parsed Log
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("FormatLintReport() output is not valid SARIF JSON: %v", err)
	}

	if parsed.Version != Version {
		t.Errorf("Parsed version = %q, want %q", parsed.Version, Version)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.ToolName != "api-style" {
		t.Errorf("ToolName = %q, want %q", opts.ToolName, "api-style")
	}
	if opts.ToolVersion != "0.1.0" {
		t.Errorf("ToolVersion = %q, want %q", opts.ToolVersion, "0.1.0")
	}
	if !opts.IncludeRules {
		t.Error("IncludeRules should default to true")
	}
	if !opts.PrettyPrint {
		t.Error("PrettyPrint should default to true")
	}
}

func TestViolationWithSuggestion(t *testing.T) {
	report := &types.LintReport{
		Status:  types.StatusFail,
		Summary: &types.ViolationSummary{Errors: 1, Total: 1},
		Violations: []types.Violation{
			{
				RuleID:     "URI-001",
				Severity:   types.SeverityError,
				Message:    "Use plural resource names",
				Path:       "/user",
				Suggestion: "Rename to /users",
				Category:   "uri-design",
			},
		},
	}

	log := FromLintReport(report, nil)
	result := log.Runs[0].Results[0]

	if result.Properties == nil {
		t.Fatal("Properties should not be nil when suggestion is present")
	}
	if result.Properties["suggestion"] != "Rename to /users" {
		t.Errorf("Properties[suggestion] = %v, want %q", result.Properties["suggestion"], "Rename to /users")
	}
	if result.Properties["category"] != "uri-design" {
		t.Errorf("Properties[category] = %v, want %q", result.Properties["category"], "uri-design")
	}
}

func TestViolationWithLineInfo(t *testing.T) {
	report := &types.LintReport{
		Status:  types.StatusFail,
		Summary: &types.ViolationSummary{Errors: 1, Total: 1},
		Violations: []types.Violation{
			{
				RuleID:    "TEST-001",
				Severity:  types.SeverityError,
				Message:   "Test",
				Path:      "/test/file.yaml",
				Line:      10,
				Column:    5,
				EndLine:   12,
				EndColumn: 10,
			},
		},
	}

	log := FromLintReport(report, nil)
	result := log.Runs[0].Results[0]

	if len(result.Locations) != 1 {
		t.Fatalf("len(Locations) = %d, want 1", len(result.Locations))
	}

	loc := result.Locations[0]
	if loc.PhysicalLocation == nil {
		t.Fatal("PhysicalLocation should not be nil")
	}

	region := loc.PhysicalLocation.Region
	if region == nil {
		t.Fatal("Region should not be nil when line info is present")
	}
	if region.StartLine != 10 {
		t.Errorf("StartLine = %d, want 10", region.StartLine)
	}
	if region.StartColumn != 5 {
		t.Errorf("StartColumn = %d, want 5", region.StartColumn)
	}
	if region.EndLine != 12 {
		t.Errorf("EndLine = %d, want 12", region.EndLine)
	}
	if region.EndColumn != 10 {
		t.Errorf("EndColumn = %d, want 10", region.EndColumn)
	}
}

func TestLogicalLocationForJSONPath(t *testing.T) {
	report := &types.LintReport{
		Status:  types.StatusFail,
		Summary: &types.ViolationSummary{Errors: 1, Total: 1},
		Violations: []types.Violation{
			{
				RuleID:   "TEST-001",
				Severity: types.SeverityError,
				Message:  "Test",
				Path:     "$.paths./users.get",
			},
		},
	}

	log := FromLintReport(report, nil)
	result := log.Runs[0].Results[0]
	loc := result.Locations[0]

	if len(loc.LogicalLocations) != 1 {
		t.Fatalf("len(LogicalLocations) = %d, want 1", len(loc.LogicalLocations))
	}

	logLoc := loc.LogicalLocations[0]
	if logLoc.FullyQualifiedName != "$.paths./users.get" {
		t.Errorf("FullyQualifiedName = %q, want %q", logLoc.FullyQualifiedName, "$.paths./users.get")
	}
	if logLoc.Kind != "jsonpath" {
		t.Errorf("Kind = %q, want %q", logLoc.Kind, "jsonpath")
	}
}

func TestEmptyReport(t *testing.T) {
	report := &types.LintReport{
		Status:     types.StatusPass,
		Summary:    &types.ViolationSummary{},
		Violations: []types.Violation{},
	}

	log := FromLintReport(report, nil)

	if len(log.Runs[0].Results) != 0 {
		t.Errorf("len(Results) = %d, want 0 for empty report", len(log.Runs[0].Results))
	}
	if len(log.Runs[0].Tool.Driver.Rules) != 0 {
		t.Errorf("len(Rules) = %d, want 0 for empty report", len(log.Runs[0].Tool.Driver.Rules))
	}
}

func TestBuildArtifacts(t *testing.T) {
	report := &types.LintReport{
		Status:     types.StatusPass,
		Summary:    &types.ViolationSummary{},
		Violations: []types.Violation{},
		Metadata: &types.ReportMetadata{
			SpecFile: "/path/to/openapi.yaml",
		},
	}

	log := FromLintReport(report, nil)

	if len(log.Runs[0].Artifacts) != 1 {
		t.Fatalf("len(Artifacts) = %d, want 1", len(log.Runs[0].Artifacts))
	}

	artifact := log.Runs[0].Artifacts[0]
	if artifact.MimeType != "application/x-yaml" {
		t.Errorf("MimeType = %q, want %q", artifact.MimeType, "application/x-yaml")
	}
}
