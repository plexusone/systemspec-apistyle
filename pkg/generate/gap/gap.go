package gap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Options configures gap analysis report generation.
type Options struct {
	Title       string
	Theme       string // "light" (default) or "dark"
	GeneratedAt time.Time
	Profile     *types.APIStyleSpec // optional, enables coverage analysis
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		Title:       "Gap Analysis Report",
		Theme:       "light",
		GeneratedAt: time.Now(),
	}
}

// Generator creates HTML gap analysis reports.
type Generator struct {
	tmpl *template.Template
}

// New creates a new gap analysis generator.
func New() (*Generator, error) {
	tmpl, err := template.New("gap").Funcs(templateFuncs()).Parse(gapTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}
	return &Generator{tmpl: tmpl}, nil
}

// Generate creates an HTML gap analysis from a multi-file lint report.
func (g *Generator) Generate(_ context.Context, report *types.MultiLintReport, opts *Options) ([]byte, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	data := buildGapData(report, opts)

	var buf bytes.Buffer
	if err := g.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateFromFile reads a lint JSON file and generates a gap analysis report.
// Auto-detects LintReport vs MultiLintReport format.
func (g *Generator) GenerateFromFile(_ context.Context, path string, opts *Options) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	report, err := parseLintJSON(data)
	if err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return g.Generate(context.Background(), report, opts)
}

func parseLintJSON(data []byte) (*types.MultiLintReport, error) {
	// Try MultiLintReport first
	var multi types.MultiLintReport
	if err := json.Unmarshal(data, &multi); err == nil && len(multi.FileReports) > 0 {
		return &multi, nil
	}

	// Fall back to single LintReport
	var single types.LintReport
	if err := json.Unmarshal(data, &single); err != nil {
		return nil, fmt.Errorf("not a valid LintReport or MultiLintReport: %w", err)
	}

	report := types.NewMultiLintReport()
	file := "spec"
	if single.Metadata != nil && single.Metadata.SpecFile != "" {
		file = single.Metadata.SpecFile
	}
	report.AddFileReport(file, &single)
	if single.Metadata != nil {
		report.Metadata = single.Metadata
	}
	return report, nil
}

// --- data types ---

type gapData struct {
	Report      *types.MultiLintReport
	Options     *Options
	GeneratedAt time.Time

	ViolationsByCategory []categoryViolations
	ViolationsBySeverity []severityGroup
	TopRules             []ruleCount
	FileResults          []fileResult
	MultiFile            bool

	Coverage *coverageData // nil when no profile
}

type categoryViolations struct {
	ID         string
	Name       string
	Violations []types.Violation
	ErrorCount int
	WarnCount  int
}

type severityGroup struct {
	Severity string
	Count    int
	Color    string
	Pct      float64
}

type ruleCount struct {
	RuleID   string
	Title    string
	Count    int
	Severity types.Severity
}

type fileResult struct {
	File       string
	Status     types.Status
	ErrorCount int
	WarnCount  int
	Total      int
}

type coverageData struct {
	TotalCategories int
	CoveredCount    int
	CleanCount      int
	UncoveredCount  int
	CoveragePercent float64
	Groups          []heatmapGroup // grouped heatmap (when categoryGroups defined)
	Heatmap         []heatmapCell  // flat heatmap (fallback)

	TotalRules       int
	EnforceableRules int
	TriggeredRules   int

	UncoveredCategories []string
	LLMOnlyRules        []string
}

type heatmapGroup struct {
	Title       string
	Description string
	Cells       []heatmapCell
	ErrorCount  int
	WarnCount   int
	CleanCount  int
}

type heatmapCell struct {
	ID             string
	Name           string
	RuleCount      int
	ViolationCount int
	Status         string // "clean", "errors", "warnings", "uncovered"
	Color          string
}

// --- data building ---

func buildGapData(report *types.MultiLintReport, opts *Options) *gapData {
	data := &gapData{
		Report:      report,
		Options:     opts,
		GeneratedAt: opts.GeneratedAt,
		MultiFile:   len(report.FileReports) > 1,
	}

	allViolations := collectAllViolations(report)

	data.ViolationsByCategory = groupByCategory(allViolations)
	data.ViolationsBySeverity = groupBySeverity(allViolations, report.Summary)
	data.TopRules = topViolatedRules(allViolations, 15)
	data.FileResults = buildFileResults(report)

	if opts.Profile != nil {
		data.Coverage = computeCoverage(opts.Profile, allViolations)
	}

	return data
}

func collectAllViolations(report *types.MultiLintReport) []types.Violation {
	var all []types.Violation
	for _, fr := range report.FileReports {
		all = append(all, fr.Violations...)
	}
	return all
}

func groupByCategory(violations []types.Violation) []categoryViolations {
	cats := make(map[string]*categoryViolations)
	var order []string

	for _, v := range violations {
		cat := v.Category
		if cat == "" {
			cat = "other"
		}
		if _, ok := cats[cat]; !ok {
			cats[cat] = &categoryViolations{ID: cat, Name: formatCategoryName(cat)}
			order = append(order, cat)
		}
		c := cats[cat]
		c.Violations = append(c.Violations, v)
		switch v.Severity {
		case types.SeverityError:
			c.ErrorCount++
		case types.SeverityWarn:
			c.WarnCount++
		}
	}

	sort.Slice(order, func(i, j int) bool {
		ci, cj := cats[order[i]], cats[order[j]]
		if ci.ErrorCount != cj.ErrorCount {
			return ci.ErrorCount > cj.ErrorCount
		}
		return len(ci.Violations) > len(cj.Violations)
	})

	result := make([]categoryViolations, len(order))
	for i, id := range order {
		result[i] = *cats[id]
	}
	return result
}

func groupBySeverity(violations []types.Violation, summary *types.ViolationSummary) []severityGroup {
	total := len(violations)
	if total == 0 {
		total = 1
	}
	groups := []severityGroup{
		{Severity: "Error", Count: summary.Errors, Color: "#ef4444", Pct: float64(summary.Errors) / float64(total) * 100},
		{Severity: "Warning", Count: summary.Warnings, Color: "#f97316", Pct: float64(summary.Warnings) / float64(total) * 100},
		{Severity: "Info", Count: summary.Infos, Color: "#3b82f6", Pct: float64(summary.Infos) / float64(total) * 100},
		{Severity: "Hint", Count: summary.Hints, Color: "#22c55e", Pct: float64(summary.Hints) / float64(total) * 100},
	}
	var result []severityGroup
	for _, g := range groups {
		if g.Count > 0 {
			result = append(result, g)
		}
	}
	return result
}

func topViolatedRules(violations []types.Violation, limit int) []ruleCount {
	counts := make(map[string]*ruleCount)
	for _, v := range violations {
		if _, ok := counts[v.RuleID]; !ok {
			counts[v.RuleID] = &ruleCount{
				RuleID:   v.RuleID,
				Title:    v.RuleTitle,
				Severity: v.Severity,
			}
		}
		counts[v.RuleID].Count++
	}

	rules := make([]ruleCount, 0, len(counts))
	for _, rc := range counts {
		rules = append(rules, *rc)
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Count > rules[j].Count
	})

	if len(rules) > limit {
		rules = rules[:limit]
	}
	return rules
}

func buildFileResults(report *types.MultiLintReport) []fileResult {
	results := make([]fileResult, len(report.FileReports))
	for i, fr := range report.FileReports {
		results[i] = fileResult{
			File:   fr.File,
			Status: fr.Status,
		}
		if fr.Summary != nil {
			results[i].ErrorCount = fr.Summary.Errors
			results[i].WarnCount = fr.Summary.Warnings
			results[i].Total = fr.Summary.Total
		}
	}
	return results
}

func computeCoverage(spec *types.APIStyleSpec, violations []types.Violation) *coverageData {
	cov := &coverageData{
		TotalCategories: len(spec.Categories),
		TotalRules:      len(spec.Rules),
	}

	// Count enforceable rules
	for _, r := range spec.Rules {
		if r.Enforcement != nil && r.Enforcement.Type != types.EnforcementNone {
			cov.EnforceableRules++
		} else {
			cov.LLMOnlyRules = append(cov.LLMOnlyRules, r.ID)
		}
	}

	// Track triggered rules
	triggeredRules := make(map[string]bool)
	violationsByCat := make(map[string][]types.Violation)
	for _, v := range violations {
		triggeredRules[v.RuleID] = true
		cat := v.Category
		if cat == "" {
			cat = "other"
		}
		violationsByCat[cat] = append(violationsByCat[cat], v)
	}
	cov.TriggeredRules = len(triggeredRules)

	// Build cell per category
	cellMap := make(map[string]heatmapCell)
	for _, cat := range spec.Categories {
		rules := spec.RulesForCategory(cat.ID)
		catViolations := violationsByCat[cat.ID]

		cell := heatmapCell{
			ID:             cat.ID,
			Name:           cat.Title,
			RuleCount:      len(rules),
			ViolationCount: len(catViolations),
		}

		if len(rules) == 0 {
			cell.Status = "uncovered"
			cell.Color = "#94a3b8"
			cov.UncoveredCount++
			cov.UncoveredCategories = append(cov.UncoveredCategories, cat.Title)
		} else if len(catViolations) == 0 {
			cell.Status = "clean"
			cell.Color = "#22c55e"
			cov.CleanCount++
			cov.CoveredCount++
		} else {
			hasErrors := false
			for _, v := range catViolations {
				if v.Severity == types.SeverityError {
					hasErrors = true
					break
				}
			}
			if hasErrors {
				cell.Status = "errors"
				cell.Color = "#ef4444"
			} else {
				cell.Status = "warnings"
				cell.Color = "#f97316"
			}
			cov.CoveredCount++
		}

		cellMap[cat.ID] = cell
	}

	// Organize into groups or flat heatmap
	if len(spec.CategoryGroups) > 0 {
		grouped := make(map[string]bool)
		for _, g := range spec.CategoryGroups {
			group := heatmapGroup{
				Title:       g.Title,
				Description: g.Description,
			}
			for _, catID := range g.Categories {
				if cell, ok := cellMap[catID]; ok {
					group.Cells = append(group.Cells, cell)
					grouped[catID] = true
					switch cell.Status {
					case "errors":
						group.ErrorCount++
					case "warnings":
						group.WarnCount++
					case "clean":
						group.CleanCount++
					}
				}
			}
			if len(group.Cells) > 0 {
				cov.Groups = append(cov.Groups, group)
			}
		}
		// Collect ungrouped categories
		var ungrouped []heatmapCell
		for _, cat := range spec.Categories {
			if !grouped[cat.ID] {
				if cell, ok := cellMap[cat.ID]; ok {
					ungrouped = append(ungrouped, cell)
				}
			}
		}
		if len(ungrouped) > 0 {
			cov.Groups = append(cov.Groups, heatmapGroup{
				Title: "Other",
				Cells: ungrouped,
			})
		}
	} else {
		for _, cat := range spec.Categories {
			if cell, ok := cellMap[cat.ID]; ok {
				cov.Heatmap = append(cov.Heatmap, cell)
			}
		}
	}

	if cov.TotalCategories > 0 {
		cov.CoveragePercent = float64(cov.CoveredCount) / float64(cov.TotalCategories) * 100
	}

	return cov
}

func formatCategoryName(id string) string {
	name := strings.ReplaceAll(id, "-", " ")
	if len(name) == 0 {
		return id
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// --- template functions ---

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"severityColor": func(sev types.Severity) string {
			switch sev {
			case types.SeverityError:
				return "#ef4444"
			case types.SeverityWarn:
				return "#f97316"
			case types.SeverityInfo:
				return "#3b82f6"
			case types.SeverityHint:
				return "#22c55e"
			default:
				return "#94a3b8"
			}
		},
		"severityLabel": func(sev types.Severity) string {
			switch sev {
			case types.SeverityError:
				return "Error"
			case types.SeverityWarn:
				return "Warning"
			case types.SeverityInfo:
				return "Info"
			case types.SeverityHint:
				return "Hint"
			default:
				return string(sev)
			}
		},
		"statusColor": func(s types.Status) string {
			if s == types.StatusPass {
				return "#22c55e"
			}
			return "#ef4444"
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
		"pct": func(v float64) string {
			return fmt.Sprintf("%.0f%%", v)
		},
		"barWidth": func(pct float64) string {
			if pct < 2 {
				return "2%"
			}
			return fmt.Sprintf("%.0f%%", pct)
		},
	}
}

// --- template ---

const gapTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Options.Title}}</title>
    <style>
        :root {
            --bg-primary: #ffffff;
            --bg-secondary: #f8fafc;
            --bg-tertiary: #f1f5f9;
            --text-primary: #1e293b;
            --text-secondary: #64748b;
            --text-muted: #94a3b8;
            --border-color: #e2e8f0;
            --accent-color: #3b82f6;
        }

        {{if eq .Options.Theme "dark"}}
        :root {
            --bg-primary: #0f172a;
            --bg-secondary: #1e293b;
            --bg-tertiary: #334155;
            --text-primary: #f1f5f9;
            --text-secondary: #94a3b8;
            --text-muted: #64748b;
            --border-color: #334155;
        }
        {{end}}

        * { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            line-height: 1.6;
            padding: 2rem;
            max-width: 1200px;
            margin: 0 auto;
        }

        @media print {
            body { padding: 0; font-size: 10pt; }
            .no-print { display: none !important; }
            details[open] summary { display: none; }
            .page-break { page-break-before: always; }
        }

        header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 2rem;
            padding-bottom: 1.5rem;
            border-bottom: 2px solid var(--border-color);
        }

        header h1 { font-size: 1.75rem; font-weight: 700; margin-bottom: 0.25rem; }

        .subtitle { color: var(--text-secondary); font-size: 0.95rem; }

        .status-badge {
            display: inline-block;
            padding: 0.5rem 1.25rem;
            border-radius: 9999px;
            font-weight: 600;
            font-size: 1rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: white;
        }

        .meta-info {
            margin-top: 0.5rem;
            font-size: 0.85rem;
            color: var(--text-muted);
        }

        /* Summary cards */
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }

        .summary-card {
            background: var(--bg-secondary);
            border-radius: 0.75rem;
            padding: 1.25rem;
            border: 1px solid var(--border-color);
        }

        .summary-card .label {
            font-size: 0.8rem;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.25rem;
        }

        .summary-card .value { font-size: 1.5rem; font-weight: 700; }

        .summary-card .detail { font-size: 0.8rem; color: var(--text-secondary); margin-top: 0.25rem; }

        /* Sections */
        section { margin-bottom: 2.5rem; }

        section h2 {
            font-size: 1.25rem;
            font-weight: 600;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 1px solid var(--border-color);
        }

        /* Severity bars */
        .severity-bars { margin-bottom: 1.5rem; }

        .severity-bar-row {
            display: flex;
            align-items: center;
            margin-bottom: 0.5rem;
            gap: 0.75rem;
        }

        .severity-bar-label {
            width: 80px;
            font-size: 0.85rem;
            font-weight: 500;
            text-align: right;
        }

        .severity-bar-track {
            flex: 1;
            height: 24px;
            background: var(--bg-tertiary);
            border-radius: 4px;
            overflow: hidden;
        }

        .severity-bar-fill {
            height: 100%;
            border-radius: 4px;
            display: flex;
            align-items: center;
            padding-left: 0.5rem;
            font-size: 0.75rem;
            font-weight: 600;
            color: white;
            min-width: fit-content;
        }

        /* Heatmap */
        .heatmap-group {
            margin-bottom: 1.25rem;
        }

        .heatmap-group-header {
            display: flex;
            justify-content: space-between;
            align-items: baseline;
            margin-bottom: 0.5rem;
        }

        .heatmap-group-title {
            font-size: 0.9rem;
            font-weight: 600;
        }

        .heatmap-group-stats {
            font-size: 0.75rem;
            color: var(--text-muted);
        }

        .heatmap-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
            gap: 0.5rem;
        }

        .heatmap-cell {
            border-radius: 0.5rem;
            padding: 0.75rem;
            text-align: center;
            color: white;
            font-size: 0.8rem;
        }

        .heatmap-cell .cell-name {
            font-weight: 600;
            font-size: 0.75rem;
            margin-bottom: 0.25rem;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .heatmap-cell .cell-stats { font-size: 0.7rem; opacity: 0.9; }

        .heatmap-legend {
            display: flex;
            gap: 1.5rem;
            font-size: 0.8rem;
            color: var(--text-secondary);
            margin-top: 0.75rem;
        }

        .legend-item {
            display: flex;
            align-items: center;
            gap: 0.35rem;
        }

        .legend-dot {
            width: 12px;
            height: 12px;
            border-radius: 3px;
        }

        /* Tables */
        table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 1rem;
        }

        th, td {
            text-align: left;
            padding: 0.6rem 0.75rem;
            border-bottom: 1px solid var(--border-color);
        }

        th {
            background: var(--bg-secondary);
            font-weight: 600;
            font-size: 0.8rem;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            color: var(--text-secondary);
        }

        tr:hover { background: var(--bg-secondary); }

        .severity-badge {
            display: inline-block;
            padding: 0.15rem 0.5rem;
            border-radius: 9999px;
            font-size: 0.7rem;
            font-weight: 600;
            text-transform: uppercase;
            color: white;
        }

        .rule-id {
            font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
            font-size: 0.85rem;
            color: var(--text-secondary);
        }

        .path-cell {
            font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
            font-size: 0.8rem;
            color: var(--text-muted);
            max-width: 300px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        /* Collapsible sections */
        details { margin-bottom: 0.75rem; }

        details summary {
            cursor: pointer;
            padding: 0.75rem 1rem;
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-radius: 0.5rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
            list-style: none;
        }

        details summary::-webkit-details-marker { display: none; }

        details summary::after {
            content: "\25B6";
            font-size: 0.65rem;
            color: var(--text-muted);
            transition: transform 0.15s;
        }

        details[open] summary::after { transform: rotate(90deg); }

        details[open] summary {
            border-radius: 0.5rem 0.5rem 0 0;
            border-bottom: none;
        }

        .detail-content {
            border: 1px solid var(--border-color);
            border-top: none;
            border-radius: 0 0 0.5rem 0.5rem;
            overflow-x: auto;
        }

        .detail-content table { margin: 0; }

        /* File results */
        .file-header {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            font-weight: 500;
        }

        .file-status {
            display: inline-block;
            width: 8px;
            height: 8px;
            border-radius: 50%;
        }

        .file-count {
            font-size: 0.85rem;
            color: var(--text-muted);
        }

        /* Uncovered areas */
        .uncovered-list {
            list-style: none;
            margin-top: 0.5rem;
        }

        .uncovered-list li {
            padding: 0.5rem 0;
            border-bottom: 1px solid var(--border-color);
            display: flex;
            justify-content: space-between;
            font-size: 0.9rem;
        }

        .uncovered-label { color: var(--text-secondary); font-size: 0.8rem; }

        /* Improvement items */
        .improvement-item {
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-left: 4px solid var(--accent-color);
            border-radius: 0 0.5rem 0.5rem 0;
            padding: 0.75rem 1rem;
            margin-bottom: 0.5rem;
        }

        .improvement-item .fix-rule {
            font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
            font-size: 0.85rem;
            color: var(--accent-color);
        }

        .improvement-item .fix-desc { font-size: 0.9rem; margin-top: 0.25rem; }

        .improvement-item .fix-meta {
            font-size: 0.8rem;
            color: var(--text-muted);
            margin-top: 0.25rem;
        }

        footer {
            margin-top: 3rem;
            padding-top: 1.5rem;
            border-top: 1px solid var(--border-color);
            text-align: center;
            font-size: 0.85rem;
            color: var(--text-muted);
        }
    </style>
</head>
<body>
    <header>
        <div>
            <h1>{{.Options.Title}}</h1>
            <div class="subtitle">
                {{if .Report.Metadata}}
                    {{if .Report.Metadata.Profile}}Profile: {{.Report.Metadata.Profile}}{{end}}
                    {{if .Report.Metadata.SpecFile}} &middot; {{.Report.Metadata.SpecFile}}{{end}}
                {{end}}
                {{if .MultiFile}} &middot; {{len .FileResults}} files{{end}}
            </div>
        </div>
        <div style="text-align: right;">
            <span class="status-badge" style="background: {{statusColor .Report.Status}};">
                {{upper (print .Report.Status)}}
            </span>
            <div class="meta-info">{{formatDate .GeneratedAt}}</div>
        </div>
    </header>

    <div class="summary-grid">
        <div class="summary-card">
            <div class="label">Errors</div>
            <div class="value" style="color: #ef4444;">{{.Report.Summary.Errors}}</div>
        </div>
        <div class="summary-card">
            <div class="label">Warnings</div>
            <div class="value" style="color: #f97316;">{{.Report.Summary.Warnings}}</div>
        </div>
        <div class="summary-card">
            <div class="label">Total Violations</div>
            <div class="value">{{.Report.Summary.Total}}</div>
            <div class="detail">{{.Report.Summary.Infos}} info, {{.Report.Summary.Hints}} hint</div>
        </div>
        {{if .Coverage}}
        <div class="summary-card">
            <div class="label">Category Coverage</div>
            <div class="value" style="color: var(--accent-color);">{{pct .Coverage.CoveragePercent}}</div>
            <div class="detail">{{.Coverage.CoveredCount}}/{{.Coverage.TotalCategories}} categories</div>
        </div>
        {{end}}
        {{if .MultiFile}}
        <div class="summary-card">
            <div class="label">Files</div>
            <div class="value">{{len .FileResults}}</div>
        </div>
        {{end}}
    </div>

    {{if .ViolationsBySeverity}}
    <section>
        <h2>Severity Distribution</h2>
        <div class="severity-bars">
        {{range .ViolationsBySeverity}}
            <div class="severity-bar-row">
                <span class="severity-bar-label">{{.Severity}}</span>
                <div class="severity-bar-track">
                    <div class="severity-bar-fill" style="width: {{barWidth .Pct}}; background: {{.Color}};">{{.Count}}</div>
                </div>
            </div>
        {{end}}
        </div>
    </section>
    {{end}}

    {{if .Coverage}}
    <section>
        <h2>Category Coverage</h2>
        {{if .Coverage.Groups}}
        {{range .Coverage.Groups}}
        <div class="heatmap-group">
            <div class="heatmap-group-header">
                <span class="heatmap-group-title">{{.Title}}</span>
                <span class="heatmap-group-stats">{{len .Cells}} categories{{if .ErrorCount}} &middot; {{.ErrorCount}} with errors{{end}}{{if .WarnCount}} &middot; {{.WarnCount}} with warnings{{end}}</span>
            </div>
            <div class="heatmap-grid">
            {{range .Cells}}
                <div class="heatmap-cell" style="background: {{.Color}};">
                    <div class="cell-name" title="{{.Name}}">{{.Name}}</div>
                    <div class="cell-stats">{{.RuleCount}} rules &middot; {{.ViolationCount}} violations</div>
                </div>
            {{end}}
            </div>
        </div>
        {{end}}
        {{else}}
        <div class="heatmap-grid">
        {{range .Coverage.Heatmap}}
            <div class="heatmap-cell" style="background: {{.Color}};">
                <div class="cell-name" title="{{.Name}}">{{.Name}}</div>
                <div class="cell-stats">{{.RuleCount}} rules &middot; {{.ViolationCount}} violations</div>
            </div>
        {{end}}
        </div>
        {{end}}
        <div class="heatmap-legend">
            <div class="legend-item"><div class="legend-dot" style="background: #22c55e;"></div> Clean (passing)</div>
            <div class="legend-item"><div class="legend-dot" style="background: #f97316;"></div> Warnings</div>
            <div class="legend-item"><div class="legend-dot" style="background: #ef4444;"></div> Errors</div>
            <div class="legend-item"><div class="legend-dot" style="background: #94a3b8;"></div> Uncovered</div>
        </div>
    </section>
    {{end}}

    {{if .TopRules}}
    <section>
        <h2>Most Violated Rules</h2>
        <table>
            <thead>
                <tr>
                    <th>Rule</th>
                    <th>Title</th>
                    <th>Severity</th>
                    <th style="text-align:right;">Count</th>
                </tr>
            </thead>
            <tbody>
            {{range .TopRules}}
                <tr>
                    <td class="rule-id">{{.RuleID}}</td>
                    <td>{{.Title}}</td>
                    <td><span class="severity-badge" style="background: {{severityColor .Severity}};">{{severityLabel .Severity}}</span></td>
                    <td style="text-align:right; font-variant-numeric: tabular-nums;">{{.Count}}</td>
                </tr>
            {{end}}
            </tbody>
        </table>
    </section>
    {{end}}

    {{if .ViolationsByCategory}}
    <section class="page-break">
        <h2>Violations by Category</h2>
        {{range .ViolationsByCategory}}
        <details{{if .ErrorCount}} open{{end}}>
            <summary>
                <span class="file-header">
                    {{.Name}}
                    <span class="file-count">{{len .Violations}} violations</span>
                </span>
            </summary>
            <div class="detail-content">
                <table>
                    <thead>
                        <tr>
                            <th style="width:80px;">Severity</th>
                            <th style="width:100px;">Rule</th>
                            <th>Message</th>
                            <th>Path</th>
                        </tr>
                    </thead>
                    <tbody>
                    {{range .Violations}}
                        <tr>
                            <td><span class="severity-badge" style="background: {{severityColor .Severity}};">{{severityLabel .Severity}}</span></td>
                            <td class="rule-id">{{.RuleID}}</td>
                            <td>{{.Message}}{{if .Suggestion}}<br><em style="color:var(--text-muted);font-size:0.85rem;">{{.Suggestion}}</em>{{end}}</td>
                            <td class="path-cell" title="{{.Path}}">{{.Path}}{{if .Line}} :{{.Line}}{{end}}</td>
                        </tr>
                    {{end}}
                    </tbody>
                </table>
            </div>
        </details>
        {{end}}
    </section>
    {{end}}

    {{if .MultiFile}}
    <section>
        <h2>Per-File Results</h2>
        {{range .FileResults}}
        <details>
            <summary>
                <span class="file-header">
                    <span class="file-status" style="background: {{statusColor .Status}};"></span>
                    {{.File}}
                    <span class="file-count">{{.Total}} violations ({{.ErrorCount}} errors, {{.WarnCount}} warnings)</span>
                </span>
            </summary>
        </details>
        {{end}}
    </section>
    {{end}}

    {{if .Coverage}}
    {{if or .Coverage.UncoveredCategories .Coverage.LLMOnlyRules}}
    <section>
        <h2>Uncovered Areas</h2>
        {{if .Coverage.UncoveredCategories}}
        <h3 style="font-size: 1rem; margin-bottom: 0.5rem; color: var(--text-secondary);">Categories Without Enforceable Rules</h3>
        <ul class="uncovered-list">
        {{range .Coverage.UncoveredCategories}}
            <li>{{.}} <span class="uncovered-label">no enforceable rules</span></li>
        {{end}}
        </ul>
        {{end}}

        {{if .Coverage.LLMOnlyRules}}
        <h3 style="font-size: 1rem; margin: 1rem 0 0.5rem; color: var(--text-secondary);">LLM-Only Rules (No Deterministic Enforcement)</h3>
        <p style="font-size: 0.9rem; color: var(--text-muted); margin-bottom: 0.5rem;">
            These rules require LLM evaluation and are not checked by the linter.
            Run <code>api-style evaluate</code> for semantic analysis.
        </p>
        <ul class="uncovered-list">
        {{range .Coverage.LLMOnlyRules}}
            <li><span class="rule-id">{{.}}</span></li>
        {{end}}
        </ul>
        {{end}}
    </section>
    {{end}}
    {{end}}

    {{if .TopRules}}
    <section>
        <h2>Improvement Opportunities</h2>
        <p style="color: var(--text-secondary); margin-bottom: 1rem; font-size: 0.9rem;">
            Fixing these rules will have the highest impact on your conformance score.
        </p>
        {{range .TopRules}}
        <div class="improvement-item" style="border-left-color: {{severityColor .Severity}};">
            <span class="fix-rule">{{.RuleID}}</span>
            {{if .Title}}<div class="fix-desc">{{.Title}}</div>{{end}}
            <div class="fix-meta">{{.Count}} violations &middot; {{severityLabel .Severity}}</div>
        </div>
        {{end}}
    </section>
    {{end}}

    <footer>
        Generated by systemspec-apistyle on {{formatDate .GeneratedAt}}
    </footer>
</body>
</html>
`
