package guide

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/generate"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Options configures HTML style guide generation.
type Options struct {
	Title       string
	Theme       string // "light" (default) or "dark"
	GeneratedAt time.Time

	IncludeTOC          bool
	IncludeExamples     bool
	IncludeRationale    bool
	IncludeReferences   bool
	IncludeConformance  bool
	IncludeMetadata     bool
	IncludeIntroduction bool
	IncludePrinciples   bool
	IncludePatterns     bool
	IncludeGlossary     bool
}

// DefaultOptions returns options with all sections enabled.
func DefaultOptions() *Options {
	return &Options{
		Theme:               "light",
		GeneratedAt:         time.Now(),
		IncludeTOC:          true,
		IncludeExamples:     true,
		IncludeRationale:    true,
		IncludeReferences:   true,
		IncludeConformance:  true,
		IncludeMetadata:     true,
		IncludeIntroduction: true,
		IncludePrinciples:   true,
		IncludePatterns:     true,
		IncludeGlossary:     true,
	}
}

// Generator creates HTML style guide documents.
type Generator struct {
	tmpl *template.Template
}

// New creates a new style guide generator.
func New() (*Generator, error) {
	tmpl, err := template.New("guide").Funcs(templateFuncs()).Parse(guideTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}
	return &Generator{tmpl: tmpl}, nil
}

// Generate creates an HTML style guide from an API style spec.
func (g *Generator) Generate(_ context.Context, spec *types.APIStyleSpec, opts *Options) ([]byte, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	categories := generate.GroupRulesByCategory(spec.Rules)
	categoryOrder := generate.GetCategoryOrder(spec.Categories, categories)

	data := &guideData{
		Spec:             spec,
		Options:          opts,
		GeneratedAt:      opts.GeneratedAt,
		CategorizedRules: categories,
		CategoryOrder:    categoryOrder,
	}

	if opts.Title == "" {
		data.Options.Title = spec.Name
	}

	var buf bytes.Buffer
	if err := g.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateFromFile reads a profile JSON file and generates an HTML style guide.
func (g *Generator) GenerateFromFile(_ context.Context, path string, opts *Options) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var spec types.APIStyleSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return g.Generate(context.Background(), &spec, opts)
}

type guideData struct {
	Spec             *types.APIStyleSpec
	Options          *Options
	GeneratedAt      time.Time
	CategorizedRules map[string][]types.Rule
	CategoryOrder    []string
}

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
		"slugify": func(s string) string {
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, " ", "-")
			s = strings.ReplaceAll(s, "&", "")
			s = strings.ReplaceAll(s, "'", "")
			return s
		},
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"catName": func(categories []types.Category, id string) string {
			return generate.GetCategoryName(categories, id)
		},
		"catDesc": func(categories []types.Category, id string) string {
			return generate.GetCategoryDescription(categories, id)
		},
		"join": func(items []string, sep string) string {
			return strings.Join(items, sep)
		},
		"hasGoodExamples": func(ex *types.Examples) bool {
			return ex != nil && len(ex.Good) > 0
		},
		"hasBadExamples": func(ex *types.Examples) bool {
			return ex != nil && len(ex.Bad) > 0
		},
		"hasDetailedExamples": func(ex *types.Examples) bool {
			return ex != nil && len(ex.Detailed) > 0
		},
		"levelOrder": func() []string {
			return []string{"bronze", "silver", "gold"}
		},
		"titleCase": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"levelColor": func(level string) string {
			switch level {
			case "bronze":
				return "#CD7F32"
			case "silver":
				return "#C0C0C0"
			case "gold":
				return "#FFD700"
			default:
				return "#94a3b8"
			}
		},
	}
}

const guideTemplate = `<!DOCTYPE html>
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
            --code-bg: #f1f5f9;
            --code-border: #e2e8f0;
            --good-bg: #f0fdf4;
            --good-border: #86efac;
            --bad-bg: #fef2f2;
            --bad-border: #fca5a5;
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
            --code-bg: #1e293b;
            --code-border: #475569;
            --good-bg: #052e16;
            --good-border: #166534;
            --bad-bg: #450a0a;
            --bad-border: #991b1b;
        }
        {{end}}

        * { margin: 0; padding: 0; box-sizing: border-box; }

        html { scroll-behavior: smooth; }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            line-height: 1.6;
        }

        .layout {
            display: grid;
            grid-template-columns: 260px 1fr;
            max-width: 1400px;
            margin: 0 auto;
        }

        @media (max-width: 900px) {
            .layout { grid-template-columns: 1fr; }
            .toc { display: none; }
        }

        @media print {
            .layout { display: block; }
            .toc { display: none !important; }
            details { open: true; }
            details[open] summary { display: none; }
            .rule-card { break-inside: avoid; }
            body { font-size: 10pt; }
        }

        /* TOC */
        .toc {
            position: sticky;
            top: 0;
            height: 100vh;
            overflow-y: auto;
            padding: 2rem 1rem;
            border-right: 1px solid var(--border-color);
            background: var(--bg-secondary);
            font-size: 0.85rem;
        }

        .toc h3 {
            font-size: 0.7rem;
            text-transform: uppercase;
            letter-spacing: 0.1em;
            color: var(--text-muted);
            margin-bottom: 0.75rem;
        }

        .toc ul { list-style: none; }

        .toc li { margin-bottom: 0.25rem; }

        .toc a {
            color: var(--text-secondary);
            text-decoration: none;
            display: block;
            padding: 0.2rem 0.5rem;
            border-radius: 4px;
            transition: background 0.15s;
        }

        .toc a:hover {
            background: var(--bg-tertiary);
            color: var(--text-primary);
        }

        .toc .section-label {
            font-weight: 600;
            color: var(--text-primary);
            margin-top: 1rem;
            margin-bottom: 0.25rem;
            font-size: 0.8rem;
        }

        /* Main content */
        main {
            padding: 2rem 3rem;
            max-width: 960px;
        }

        @media (max-width: 900px) {
            main { padding: 1.5rem; }
        }

        /* Header */
        .guide-header {
            margin-bottom: 2.5rem;
            padding-bottom: 1.5rem;
            border-bottom: 2px solid var(--border-color);
        }

        .guide-header h1 {
            font-size: 2rem;
            font-weight: 700;
            margin-bottom: 0.5rem;
        }

        .guide-header .description {
            color: var(--text-secondary);
            font-size: 1.05rem;
            margin-bottom: 1rem;
        }

        .meta-row {
            display: flex;
            flex-wrap: wrap;
            gap: 1rem;
            font-size: 0.85rem;
            color: var(--text-muted);
        }

        .meta-row span {
            display: inline-flex;
            align-items: center;
            gap: 0.3rem;
        }

        .version-badge {
            display: inline-block;
            padding: 0.15rem 0.5rem;
            border-radius: 9999px;
            background: var(--accent-color);
            color: white;
            font-size: 0.75rem;
            font-weight: 600;
        }

        /* Sections */
        section {
            margin-bottom: 3rem;
        }

        section > h2 {
            font-size: 1.4rem;
            font-weight: 700;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 1px solid var(--border-color);
        }

        section p {
            margin-bottom: 0.75rem;
            max-width: 70ch;
        }

        /* Severity badge */
        .severity-badge {
            display: inline-block;
            padding: 0.15rem 0.5rem;
            border-radius: 9999px;
            font-size: 0.7rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            color: white;
        }

        /* Rule cards */
        .rule-card {
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-radius: 0.5rem;
            margin-bottom: 1rem;
            overflow: hidden;
        }

        .rule-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.75rem 1rem;
            cursor: pointer;
        }

        .rule-header h3 {
            font-size: 0.95rem;
            font-weight: 600;
        }

        .rule-id {
            color: var(--text-muted);
            font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
            font-size: 0.85rem;
            margin-right: 0.5rem;
        }

        .rule-body {
            padding: 0 1rem 1rem;
            border-top: 1px solid var(--border-color);
        }

        .rule-body p { margin-top: 0.75rem; }

        details summary {
            list-style: none;
        }

        details summary::-webkit-details-marker {
            display: none;
        }

        details .rule-header::after {
            content: "\25B6";
            font-size: 0.65rem;
            color: var(--text-muted);
            transition: transform 0.15s;
        }

        details[open] .rule-header::after {
            transform: rotate(90deg);
        }

        /* Rationale */
        .rationale {
            margin-top: 0.75rem;
            padding: 0.75rem 1rem;
            border-left: 3px solid var(--accent-color);
            background: var(--bg-tertiary);
            border-radius: 0 4px 4px 0;
            font-size: 0.9rem;
            color: var(--text-secondary);
        }

        /* Examples */
        .examples { margin-top: 1rem; }

        .examples h4 {
            font-size: 0.8rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--text-muted);
            margin-bottom: 0.5rem;
        }

        .example-good, .example-bad {
            border-radius: 0.375rem;
            padding: 0.5rem 0.75rem;
            margin-bottom: 0.5rem;
            font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
            font-size: 0.85rem;
            overflow-x: auto;
        }

        .example-good {
            background: var(--good-bg);
            border: 1px solid var(--good-border);
        }

        .example-bad {
            background: var(--bad-bg);
            border: 1px solid var(--bad-border);
        }

        /* Code blocks */
        pre {
            background: var(--code-bg);
            border: 1px solid var(--code-border);
            border-radius: 0.375rem;
            padding: 0.75rem 1rem;
            overflow-x: auto;
            font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
            font-size: 0.85rem;
            margin: 0.5rem 0;
        }

        code {
            font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
            font-size: 0.85em;
            background: var(--code-bg);
            padding: 0.1em 0.3em;
            border-radius: 3px;
        }

        pre code {
            background: none;
            padding: 0;
        }

        /* Decision tables */
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 0.75rem 0;
            font-size: 0.9rem;
        }

        th, td {
            text-align: left;
            padding: 0.5rem 0.75rem;
            border-bottom: 1px solid var(--border-color);
        }

        th {
            background: var(--bg-tertiary);
            font-weight: 600;
            font-size: 0.8rem;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            color: var(--text-secondary);
        }

        /* References */
        .references {
            margin-top: 0.75rem;
            font-size: 0.85rem;
        }

        .references a {
            color: var(--accent-color);
            text-decoration: none;
        }

        .references a:hover {
            text-decoration: underline;
        }

        /* Principle cards */
        .principle-card {
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-radius: 0.5rem;
            padding: 1rem 1.25rem;
            margin-bottom: 1rem;
        }

        .principle-card h3 {
            font-size: 1rem;
            margin-bottom: 0.5rem;
        }

        .principle-card .related {
            margin-top: 0.5rem;
            font-size: 0.8rem;
            color: var(--text-muted);
        }

        /* Conformance */
        .conformance-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
            gap: 1rem;
        }

        .conformance-card {
            border-radius: 0.5rem;
            padding: 1.25rem;
            border: 2px solid;
        }

        .conformance-card h3 {
            font-size: 1.1rem;
            margin-bottom: 0.5rem;
        }

        .conformance-card .rules {
            margin-top: 0.75rem;
            font-size: 0.85rem;
        }

        .conformance-card .rules li {
            margin-left: 1rem;
            margin-bottom: 0.15rem;
        }

        /* Pattern cards */
        .pattern-card {
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-radius: 0.5rem;
            padding: 1.25rem;
            margin-bottom: 1.5rem;
        }

        .pattern-card h3 {
            font-size: 1.1rem;
            margin-bottom: 0.25rem;
        }

        .pattern-summary {
            color: var(--text-secondary);
            font-size: 0.9rem;
            margin-bottom: 0.75rem;
        }

        .pattern-section {
            margin-top: 0.75rem;
        }

        .pattern-section strong {
            font-size: 0.85rem;
            color: var(--text-secondary);
        }

        /* Glossary */
        .glossary-term {
            margin-bottom: 1rem;
        }

        .glossary-term dt {
            font-weight: 600;
            font-size: 0.95rem;
        }

        .glossary-term dt .alias {
            font-weight: 400;
            color: var(--text-muted);
            font-size: 0.85rem;
        }

        .glossary-term dd {
            margin-left: 1rem;
            color: var(--text-secondary);
        }

        /* Footer */
        footer {
            margin-top: 3rem;
            padding: 1.5rem 0;
            border-top: 1px solid var(--border-color);
            text-align: center;
            font-size: 0.85rem;
            color: var(--text-muted);
        }

        /* Rule count badge */
        .rule-count {
            display: inline-block;
            background: var(--bg-tertiary);
            color: var(--text-secondary);
            font-size: 0.75rem;
            padding: 0.1rem 0.4rem;
            border-radius: 9999px;
            margin-left: 0.5rem;
            font-weight: 500;
        }
    </style>
</head>
<body>
<div class="layout">

{{if .Options.IncludeTOC}}
<nav class="toc">
    <h3>Contents</h3>
    <ul>
    {{if .Options.IncludeMetadata}}<li><a href="#metadata">Metadata</a></li>{{end}}
    {{if and .Options.IncludeIntroduction .Spec.Introduction}}<li><a href="#introduction">Introduction</a></li>{{end}}
    {{if and .Options.IncludePrinciples .Spec.Principles}}<li><a href="#principles">Design Principles</a></li>{{end}}
    {{if and .Options.IncludeConformance .Spec.ConformanceLevels}}<li><a href="#conformance">Conformance Levels</a></li>{{end}}
    {{if and .Options.IncludePatterns .Spec.Patterns}}<li><a href="#patterns">Design Patterns</a></li>{{end}}
    <li class="section-label">Rules</li>
    {{range .CategoryOrder}}
    <li><a href="#cat-{{slugify .}}">{{catName $.Spec.Categories .}}<span class="rule-count">{{len (index $.CategorizedRules .)}}</span></a></li>
    {{end}}
    {{if and .Options.IncludeGlossary .Spec.Glossary}}<li><a href="#glossary">Glossary</a></li>{{end}}
    </ul>
</nav>
{{end}}

<main>
    <div class="guide-header">
        <h1>{{.Options.Title}}</h1>
        {{if .Spec.Description}}<p class="description">{{.Spec.Description}}</p>{{end}}
        <div class="meta-row">
            {{if .Spec.Version}}<span><span class="version-badge">v{{.Spec.Version}}</span></span>{{end}}
            {{if .Spec.Metadata}}
                {{if .Spec.Metadata.Author}}<span>Author: {{.Spec.Metadata.Author}}</span>{{end}}
                {{if .Spec.Metadata.LastUpdated}}<span>Updated: {{.Spec.Metadata.LastUpdated}}</span>{{end}}
                {{if .Spec.Metadata.License}}<span>License: {{.Spec.Metadata.License}}</span>{{end}}
            {{end}}
            <span>{{len .Spec.Rules}} rules</span>
        </div>
    </div>

    {{if and .Options.IncludeMetadata .Spec.Metadata}}
    <section id="metadata">
        <h2>Metadata</h2>
        <table>
            <tbody>
            {{if .Spec.Metadata.Author}}<tr><td><strong>Author</strong></td><td>{{.Spec.Metadata.Author}}</td></tr>{{end}}
            {{if .Spec.Metadata.URL}}<tr><td><strong>Source</strong></td><td><a href="{{.Spec.Metadata.URL}}">{{.Spec.Metadata.URL}}</a></td></tr>{{end}}
            {{if .Spec.Metadata.Repository}}<tr><td><strong>Repository</strong></td><td><a href="{{.Spec.Metadata.Repository}}">{{.Spec.Metadata.Repository}}</a></td></tr>{{end}}
            {{if .Spec.Metadata.License}}<tr><td><strong>License</strong></td><td>{{.Spec.Metadata.License}}</td></tr>{{end}}
            {{if .Spec.Metadata.LastUpdated}}<tr><td><strong>Last Updated</strong></td><td>{{.Spec.Metadata.LastUpdated}}</td></tr>{{end}}
            {{if .Spec.Metadata.Contact}}<tr><td><strong>Contact</strong></td><td>{{.Spec.Metadata.Contact}}</td></tr>{{end}}
            </tbody>
        </table>
    </section>
    {{end}}

    {{if and .Options.IncludeIntroduction .Spec.Introduction}}
    <section id="introduction">
        <h2>Introduction</h2>
        <p>{{.Spec.Introduction}}</p>
    </section>
    {{end}}

    {{if and .Options.IncludePrinciples .Spec.Principles}}
    <section id="principles">
        <h2>Design Principles</h2>
        {{range .Spec.Principles}}
        <div class="principle-card">
            <h3>{{.Title}}</h3>
            <p>{{.Description}}</p>
            {{if .RelatedRules}}
            <div class="related">Related: {{join .RelatedRules ", "}}</div>
            {{end}}
        </div>
        {{end}}
    </section>
    {{end}}

    {{if and .Options.IncludeConformance .Spec.ConformanceLevels}}
    <section id="conformance">
        <h2>Conformance Levels</h2>
        <div class="conformance-grid">
        {{range levelOrder}}
            {{$level := index $.Spec.ConformanceLevels .}}
            {{if $level.Description}}
            <div class="conformance-card" style="border-color: {{levelColor .}};">
                <h3 style="color: {{levelColor .}};">{{titleCase .}}</h3>
                <p>{{$level.Description}}</p>
                {{if $level.RequiredRules}}
                <div class="rules">
                    <strong>Required Rules:</strong>
                    <ul>
                    {{range $level.RequiredRules}}<li><code>{{.}}</code></li>{{end}}
                    </ul>
                </div>
                {{end}}
            </div>
            {{end}}
        {{end}}
        </div>
    </section>
    {{end}}

    {{if and .Options.IncludePatterns .Spec.Patterns}}
    <section id="patterns">
        <h2>Design Patterns</h2>
        {{range .Spec.Patterns}}
        <div class="pattern-card">
            <h3>{{.Name}}</h3>
            {{if .Summary}}<p class="pattern-summary">{{.Summary}}</p>{{end}}
            {{if .Problem}}<div class="pattern-section"><strong>Problem:</strong> {{.Problem}}</div>{{end}}
            {{if .Solution}}<div class="pattern-section"><strong>Solution:</strong> {{.Solution}}</div>{{end}}
            {{if .When}}<div class="pattern-section"><strong>When to Use:</strong> {{.When}}</div>{{end}}
            {{if .Examples}}
            <div class="examples">
                {{range .Examples}}
                <div class="pattern-section">
                    <strong>{{.Title}}</strong>
                    {{if .Description}}<p>{{.Description}}</p>{{end}}
                    {{if .Code}}<pre><code>{{.Code}}</code></pre>{{end}}
                </div>
                {{end}}
            </div>
            {{end}}
            {{if .RelatedRules}}
            <div class="pattern-section" style="margin-top: 0.75rem; font-size: 0.85rem; color: var(--text-muted);">
                Related Rules: {{join .RelatedRules ", "}}
            </div>
            {{end}}
            {{if .References}}
            <div class="references">
                {{range .References}}<a href="{{.URL}}">{{.Title}}</a> {{end}}
            </div>
            {{end}}
        </div>
        {{end}}
    </section>
    {{end}}

    {{range .CategoryOrder}}
    {{$catID := .}}
    {{$rules := index $.CategorizedRules $catID}}
    <section id="cat-{{slugify $catID}}">
        <h2>{{catName $.Spec.Categories $catID}} <span class="rule-count">{{len $rules}} rules</span></h2>
        {{$desc := catDesc $.Spec.Categories $catID}}
        {{if $desc}}<p>{{$desc}}</p>{{end}}

        {{range $rules}}
        <div class="rule-card">
            <details{{if eq .Severity "error"}} open{{end}}>
                <summary>
                    <div class="rule-header">
                        <h3><span class="rule-id">{{.ID}}</span>{{.Title}}</h3>
                        <span class="severity-badge" style="background: {{severityColor .Severity}};">{{severityLabel .Severity}}</span>
                    </div>
                </summary>
                <div class="rule-body">
                    {{if .Description}}<p>{{.Description}}</p>{{end}}

                    {{if and $.Options.IncludeRationale .Rationale}}
                    <div class="rationale">{{.Rationale}}</div>
                    {{end}}

                    {{if .DecisionTables}}
                    {{range .DecisionTables}}
                    {{if .Title}}<h4 style="margin-top: 0.75rem; font-size: 0.9rem;">{{.Title}}</h4>{{end}}
                    {{if and .Headers .Rows}}
                    <table>
                        <thead><tr>{{range .Headers}}<th>{{.}}</th>{{end}}</tr></thead>
                        <tbody>
                        {{range .Rows}}<tr>{{range .Values}}<td>{{.}}</td>{{end}}</tr>{{end}}
                        </tbody>
                    </table>
                    {{end}}
                    {{end}}
                    {{end}}

                    {{if $.Options.IncludeExamples}}
                    {{if hasGoodExamples .Examples}}
                    <div class="examples">
                        <h4>Good</h4>
                        {{range .Examples.Good}}
                        <div class="example-good">{{.}}</div>
                        {{end}}
                    </div>
                    {{end}}
                    {{if hasBadExamples .Examples}}
                    <div class="examples">
                        <h4>Bad</h4>
                        {{range .Examples.Bad}}
                        <div class="example-bad">{{.}}</div>
                        {{end}}
                    </div>
                    {{end}}
                    {{if hasDetailedExamples .Examples}}
                    <div class="examples">
                        {{range .Examples.Detailed}}
                        <div class="pattern-section">
                            <strong>{{.Title}}{{if eq .Type "good"}} (Correct){{else if eq .Type "bad"}} (Incorrect){{end}}</strong>
                            {{if .Description}}<p>{{.Description}}</p>{{end}}
                            {{if .Code}}<pre><code>{{.Code}}</code></pre>{{end}}
                            {{if and .Before .After}}
                            <p><strong>Before:</strong></p><pre><code>{{.Before}}</code></pre>
                            <p><strong>After:</strong></p><pre><code>{{.After}}</code></pre>
                            {{end}}
                        </div>
                        {{end}}
                    </div>
                    {{end}}
                    {{end}}

                    {{if and $.Options.IncludeReferences .References}}
                    <div class="references">
                        <strong>References:</strong>
                        {{range .References}} <a href="{{.URL}}">{{.Title}}</a>{{end}}
                    </div>
                    {{end}}
                </div>
            </details>
        </div>
        {{end}}
    </section>
    {{end}}

    {{if and .Options.IncludeGlossary .Spec.Glossary}}
    <section id="glossary">
        <h2>Glossary</h2>
        <dl>
        {{range .Spec.Glossary}}
        <div class="glossary-term">
            <dt>{{.Term}}{{if .Aliases}} <span class="alias">({{join .Aliases ", "}})</span>{{end}}</dt>
            <dd>{{.Definition}}</dd>
        </div>
        {{end}}
        </dl>
    </section>
    {{end}}

    <footer>
        Generated by systemspec-apistyle on {{formatDate .GeneratedAt}}
    </footer>
</main>

</div>
</body>
</html>
`
