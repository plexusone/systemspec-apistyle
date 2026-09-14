// Package sarif provides SARIF (Static Analysis Results Interchange Format) output
// for API style linting results. SARIF 2.1.0 is supported.
//
// SARIF enables integration with IDEs (VS Code, JetBrains), GitHub Code Scanning,
// and other static analysis tools.
package sarif

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/grokify/mogo/path/filepathutil"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Version is the SARIF schema version.
const Version = "2.1.0"

// SchemaURI is the URI for the SARIF 2.1.0 JSON schema.
const SchemaURI = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"

// Log is the top-level SARIF object containing one or more runs.
type Log struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

// Run represents a single invocation of an analysis tool.
type Run struct {
	Tool        Tool         `json:"tool"`
	Results     []Result     `json:"results,omitempty"`
	Invocations []Invocation `json:"invocations,omitempty"`
	Artifacts   []Artifact   `json:"artifacts,omitempty"`
}

// Tool describes the analysis tool that produced the results.
type Tool struct {
	Driver ToolComponent `json:"driver"`
}

// ToolComponent provides metadata about the tool.
type ToolComponent struct {
	Name            string              `json:"name"`
	Version         string              `json:"version,omitempty"`
	SemanticVersion string              `json:"semanticVersion,omitempty"`
	InformationURI  string              `json:"informationUri,omitempty"`
	Rules           []ReportingDescr    `json:"rules,omitempty"`
	Organization    string              `json:"organization,omitempty"`
	FullName        string              `json:"fullName,omitempty"`
	ShortDescr      *MultiformatMessage `json:"shortDescription,omitempty"`
}

// ReportingDescr describes a rule that the tool can report.
type ReportingDescr struct {
	ID               string              `json:"id"`
	Name             string              `json:"name,omitempty"`
	ShortDescr       *MultiformatMessage `json:"shortDescription,omitempty"`
	FullDescr        *MultiformatMessage `json:"fullDescription,omitempty"`
	HelpURI          string              `json:"helpUri,omitempty"`
	Help             *MultiformatMessage `json:"help,omitempty"`
	DefaultConfig    *ReportingConfig    `json:"defaultConfiguration,omitempty"`
	Properties       PropertyBag         `json:"properties,omitempty"`
	DeprecatedIDs    []string            `json:"deprecatedIds,omitempty"`
	DeprecatedNames  []string            `json:"deprecatedNames,omitempty"`
	RelationshipList []Relationship      `json:"relationships,omitempty"`
}

// ReportingConfig specifies the default severity and other settings.
type ReportingConfig struct {
	Enabled bool    `json:"enabled,omitempty"`
	Level   Level   `json:"level,omitempty"`
	Rank    float64 `json:"rank,omitempty"`
}

// MultiformatMessage provides text in multiple formats.
type MultiformatMessage struct {
	Text     string `json:"text,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

// Result represents a single finding from the analysis.
type Result struct {
	RuleID     string        `json:"ruleId"`
	RuleIndex  int           `json:"ruleIndex,omitempty"`
	Level      Level         `json:"level,omitempty"`
	Kind       ResultKind    `json:"kind,omitempty"`
	Message    Message       `json:"message"`
	Locations  []Location    `json:"locations,omitempty"`
	Fixes      []Fix         `json:"fixes,omitempty"`
	Properties PropertyBag   `json:"properties,omitempty"`
	RelatedLoc []Location    `json:"relatedLocations,omitempty"`
	CodeFlows  []CodeFlow    `json:"codeFlows,omitempty"`
	Stacks     []Stack       `json:"stacks,omitempty"`
	Suppressed []Suppression `json:"suppressions,omitempty"`
}

// Level indicates the severity of a result.
type Level string

// Level constants for SARIF result severity.
const (
	LevelNone    Level = "none"
	LevelNote    Level = "note"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

// ResultKind indicates the nature of a result.
type ResultKind string

// ResultKind constants for SARIF result classification.
const (
	KindNotApplicable ResultKind = "notApplicable"
	KindPass          ResultKind = "pass"
	KindFail          ResultKind = "fail"
	KindReview        ResultKind = "review"
	KindOpen          ResultKind = "open"
	KindInformational ResultKind = "informational"
)

// Message provides the text of a result message.
type Message struct {
	Text      string   `json:"text,omitempty"`
	Markdown  string   `json:"markdown,omitempty"`
	ID        string   `json:"id,omitempty"`
	Arguments []string `json:"arguments,omitempty"`
}

// Location specifies where a result was detected.
type Location struct {
	ID               int          `json:"id,omitempty"`
	PhysicalLocation *PhysicalLoc `json:"physicalLocation,omitempty"`
	LogicalLocations []LogicalLoc `json:"logicalLocations,omitempty"`
	Message          *Message     `json:"message,omitempty"`
	Properties       PropertyBag  `json:"properties,omitempty"`
}

// PhysicalLoc identifies a file and region within it.
type PhysicalLoc struct {
	ArtifactLocation *ArtifactLoc `json:"artifactLocation,omitempty"`
	Region           *Region      `json:"region,omitempty"`
	ContextRegion    *Region      `json:"contextRegion,omitempty"`
}

// ArtifactLoc identifies a file.
type ArtifactLoc struct {
	URI       string `json:"uri,omitempty"`
	URIBaseID string `json:"uriBaseId,omitempty"`
	Index     int    `json:"index,omitempty"`
}

// Region identifies a portion of a file.
type Region struct {
	StartLine   int      `json:"startLine,omitempty"`
	StartColumn int      `json:"startColumn,omitempty"`
	EndLine     int      `json:"endLine,omitempty"`
	EndColumn   int      `json:"endColumn,omitempty"`
	CharOffset  int      `json:"charOffset,omitempty"`
	CharLength  int      `json:"charLength,omitempty"`
	ByteOffset  int      `json:"byteOffset,omitempty"`
	ByteLength  int      `json:"byteLength,omitempty"`
	Snippet     *Snippet `json:"snippet,omitempty"`
	Message     *Message `json:"message,omitempty"`
}

// Snippet contains source code text.
type Snippet struct {
	Text     string              `json:"text,omitempty"`
	Rendered *MultiformatMessage `json:"rendered,omitempty"`
}

// LogicalLoc identifies a logical location (like a function or JSON path).
type LogicalLoc struct {
	Name               string `json:"name,omitempty"`
	Index              int    `json:"index,omitempty"`
	FullyQualifiedName string `json:"fullyQualifiedName,omitempty"`
	DecoratedName      string `json:"decoratedName,omitempty"`
	Kind               string `json:"kind,omitempty"`
	ParentIndex        int    `json:"parentIndex,omitempty"`
}

// Fix describes a proposed fix for a result.
type Fix struct {
	Description     *Message         `json:"description,omitempty"`
	ArtifactChanges []ArtifactChange `json:"artifactChanges,omitempty"`
}

// ArtifactChange describes changes to a single artifact.
type ArtifactChange struct {
	ArtifactLocation *ArtifactLoc  `json:"artifactLocation"`
	Replacements     []Replacement `json:"replacements"`
}

// Replacement describes a replacement in a file.
type Replacement struct {
	DeletedRegion   *Region          `json:"deletedRegion"`
	InsertedContent *ArtifactContent `json:"insertedContent,omitempty"`
}

// ArtifactContent contains the content to insert.
type ArtifactContent struct {
	Text   string `json:"text,omitempty"`
	Binary string `json:"binary,omitempty"`
}

// Invocation describes a single invocation of the tool.
type Invocation struct {
	CommandLine         string       `json:"commandLine,omitempty"`
	ExecutionSuccessful bool         `json:"executionSuccessful"`
	StartTimeUTC        string       `json:"startTimeUtc,omitempty"`
	EndTimeUTC          string       `json:"endTimeUtc,omitempty"`
	ExitCode            int          `json:"exitCode,omitempty"`
	WorkingDirectory    *ArtifactLoc `json:"workingDirectory,omitempty"`
}

// Artifact describes a file that was analyzed.
type Artifact struct {
	Location    *ArtifactLoc `json:"location,omitempty"`
	Length      int          `json:"length,omitempty"`
	MimeType    string       `json:"mimeType,omitempty"`
	Encoding    string       `json:"encoding,omitempty"`
	Description *Message     `json:"description,omitempty"`
}

// CodeFlow describes execution paths through the code.
type CodeFlow struct {
	ThreadFlows []ThreadFlow `json:"threadFlows"`
	Message     *Message     `json:"message,omitempty"`
}

// ThreadFlow describes a sequence of locations.
type ThreadFlow struct {
	Locations []ThreadFlowLoc `json:"locations"`
}

// ThreadFlowLoc is a location in a thread flow.
type ThreadFlowLoc struct {
	Location *Location `json:"location,omitempty"`
}

// Stack describes a call stack.
type Stack struct {
	Frames  []StackFrame `json:"frames"`
	Message *Message     `json:"message,omitempty"`
}

// StackFrame describes a single frame in a stack.
type StackFrame struct {
	Location *Location `json:"location,omitempty"`
	Module   string    `json:"module,omitempty"`
}

// Suppression describes a suppressed result.
type Suppression struct {
	Kind          string `json:"kind"`
	Status        string `json:"status,omitempty"`
	Justification string `json:"justification,omitempty"`
}

// Relationship describes a relationship between rules.
type Relationship struct {
	Target      *ReportingDescrRef `json:"target"`
	Kinds       []string           `json:"kinds,omitempty"`
	Description *Message           `json:"description,omitempty"`
}

// ReportingDescrRef references a reporting descriptor.
type ReportingDescrRef struct {
	ID            string            `json:"id,omitempty"`
	Index         int               `json:"index,omitempty"`
	ToolComponent *ToolComponentRef `json:"toolComponent,omitempty"`
}

// ToolComponentRef references a tool component.
type ToolComponentRef struct {
	Name  string `json:"name,omitempty"`
	Index int    `json:"index,omitempty"`
}

// PropertyBag is a set of key-value pairs for custom properties.
type PropertyBag map[string]any

// Options configures SARIF output generation.
type Options struct {
	// ToolName overrides the default tool name.
	ToolName string

	// ToolVersion specifies the tool version.
	ToolVersion string

	// ToolURI is a URL for more information about the tool.
	ToolURI string

	// IncludeRules adds rule definitions to the output.
	IncludeRules bool

	// Rules provides rule metadata for the rules array.
	Rules map[string]*types.Rule

	// BaseURI is the base URI for artifact locations.
	BaseURI string

	// PrettyPrint enables indented JSON output.
	PrettyPrint bool
}

// DefaultOptions returns options with sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		ToolName:     "api-style",
		ToolVersion:  "0.1.0",
		ToolURI:      "https://github.com/plexusone/systemspec-apistyle",
		IncludeRules: true,
		PrettyPrint:  true,
	}
}

// FromLintReport converts a LintReport to a SARIF Log.
func FromLintReport(report *types.LintReport, opts *Options) *Log {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Build tool component
	driver := ToolComponent{
		Name:           opts.ToolName,
		Version:        opts.ToolVersion,
		InformationURI: opts.ToolURI,
	}

	// Build rule definitions if requested
	ruleIndex := make(map[string]int)
	if opts.IncludeRules {
		driver.Rules = buildRuleDescriptors(report.Violations, opts.Rules, ruleIndex)
	}

	// Convert violations to results
	results := make([]Result, 0, len(report.Violations))
	for _, v := range report.Violations {
		result := violationToResult(v, opts, ruleIndex)
		results = append(results, result)
	}

	// Build artifacts list
	artifacts := buildArtifacts(report, opts)

	// Build invocation
	invocations := []Invocation{
		{
			ExecutionSuccessful: true,
		},
	}

	return &Log{
		Schema:  SchemaURI,
		Version: Version,
		Runs: []Run{
			{
				Tool:        Tool{Driver: driver},
				Results:     results,
				Invocations: invocations,
				Artifacts:   artifacts,
			},
		},
	}
}

// buildRuleDescriptors creates SARIF rule definitions from violations.
func buildRuleDescriptors(violations []types.Violation, rules map[string]*types.Rule, ruleIndex map[string]int) []ReportingDescr {
	seen := make(map[string]bool)
	descriptors := []ReportingDescr{}

	for _, v := range violations {
		if seen[v.RuleID] {
			continue
		}
		seen[v.RuleID] = true

		rd := ReportingDescr{
			ID: v.RuleID,
		}

		// Add rule metadata if available
		if rules != nil {
			if rule, ok := rules[v.RuleID]; ok {
				rd.Name = rule.Title
				if rule.Title != "" {
					rd.ShortDescr = &MultiformatMessage{Text: rule.Title}
				}
				if rule.Rationale != "" {
					rd.FullDescr = &MultiformatMessage{Text: rule.Rationale}
				}
				rd.DefaultConfig = &ReportingConfig{
					Level: severityToLevel(rule.Severity),
				}
				if rule.Category != "" {
					rd.Properties = PropertyBag{"category": rule.Category}
				}
			}
		}

		// Use violation data as fallback
		if rd.Name == "" && v.RuleTitle != "" {
			rd.Name = v.RuleTitle
			rd.ShortDescr = &MultiformatMessage{Text: v.RuleTitle}
		}
		if rd.DefaultConfig == nil {
			rd.DefaultConfig = &ReportingConfig{
				Level: severityToLevel(v.Severity),
			}
		}

		ruleIndex[v.RuleID] = len(descriptors)
		descriptors = append(descriptors, rd)
	}

	return descriptors
}

// violationToResult converts a single violation to a SARIF result.
func violationToResult(v types.Violation, opts *Options, ruleIndex map[string]int) Result {
	result := Result{
		RuleID:  v.RuleID,
		Level:   severityToLevel(v.Severity),
		Kind:    KindFail,
		Message: Message{Text: v.Message},
	}

	// Add rule index if available
	if idx, ok := ruleIndex[v.RuleID]; ok {
		result.RuleIndex = idx
	}

	// Build location
	loc := Location{}

	// Physical location (file + line/column)
	physLoc := &PhysicalLoc{
		ArtifactLocation: &ArtifactLoc{
			URI: normalizeURI(v.Path, opts.BaseURI),
		},
	}

	if v.Line > 0 {
		physLoc.Region = &Region{
			StartLine:   v.Line,
			StartColumn: maxInt(1, v.Column),
		}
		if v.EndLine > 0 {
			physLoc.Region.EndLine = v.EndLine
			physLoc.Region.EndColumn = v.EndColumn
		}
	}
	loc.PhysicalLocation = physLoc

	// Logical location (JSON path)
	if strings.HasPrefix(v.Path, "$.") || strings.HasPrefix(v.Path, "$[") {
		loc.LogicalLocations = []LogicalLoc{
			{
				FullyQualifiedName: v.Path,
				Kind:               "jsonpath",
			},
		}
		// For JSON path, use the file from metadata if available
		if loc.PhysicalLocation != nil && loc.PhysicalLocation.ArtifactLocation != nil {
			// Keep the JSON path in logical location, use file in physical
			loc.PhysicalLocation.ArtifactLocation.URI = ""
		}
	}

	result.Locations = []Location{loc}

	// Add suggestion as a fix if available
	if v.Suggestion != "" {
		result.Properties = PropertyBag{"suggestion": v.Suggestion}
	}

	// Add category as property
	if v.Category != "" {
		if result.Properties == nil {
			result.Properties = PropertyBag{}
		}
		result.Properties["category"] = v.Category
	}

	return result
}

// buildArtifacts creates the artifacts array from the report.
func buildArtifacts(report *types.LintReport, opts *Options) []Artifact {
	if report.Metadata == nil || report.Metadata.SpecFile == "" {
		return nil
	}

	return []Artifact{
		{
			Location: &ArtifactLoc{
				URI: normalizeURI(report.Metadata.SpecFile, opts.BaseURI),
			},
			MimeType: getMimeType(report.Metadata.SpecFile),
		},
	}
}

// severityToLevel converts our severity to SARIF level.
func severityToLevel(s types.Severity) Level {
	switch s {
	case types.SeverityError:
		return LevelError
	case types.SeverityWarn:
		return LevelWarning
	case types.SeverityInfo:
		return LevelNote
	case types.SeverityHint:
		return LevelNote
	default:
		return LevelWarning
	}
}

// normalizeURI creates a proper URI from a path.
func normalizeURI(path, baseURI string) string {
	if path == "" {
		return ""
	}

	// If it's already a URI, return as-is
	if strings.HasPrefix(path, "file://") || strings.HasPrefix(path, "http") {
		return path
	}

	// If it's a JSON path, return empty (will be in logical location)
	if strings.HasPrefix(path, "$.") || strings.HasPrefix(path, "$[") {
		return ""
	}

	// Convert to absolute path if relative
	if !filepathutil.IsAbsAny(path) && baseURI != "" {
		return baseURI + "/" + path
	}

	// Convert to file URI
	if filepathutil.IsAbsAny(path) {
		return "file://" + path
	}

	return path
}

// getMimeType returns the MIME type for a file.
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		return "application/x-yaml"
	case ".json":
		return "application/json"
	default:
		return "text/plain"
	}
}

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Marshal converts a SARIF Log to JSON bytes.
func (l *Log) Marshal(pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(l, "", "  ")
	}
	return json.Marshal(l)
}

// String returns the SARIF Log as a JSON string.
func (l *Log) String() string {
	data, err := l.Marshal(true)
	if err != nil {
		return fmt.Sprintf("error marshaling SARIF: %v", err)
	}
	return string(data)
}

// FormatLintReport converts a LintReport to SARIF JSON string.
func FormatLintReport(report *types.LintReport, opts *Options) (string, error) {
	log := FromLintReport(report, opts)
	data, err := log.Marshal(opts == nil || opts.PrettyPrint)
	if err != nil {
		return "", fmt.Errorf("marshaling SARIF: %w", err)
	}
	return string(data) + "\n", nil
}
