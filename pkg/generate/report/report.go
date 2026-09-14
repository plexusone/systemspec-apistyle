package report

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/judge"
	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// Options configures report generation.
type Options struct {
	// Title overrides the default report title.
	Title string

	// Logo is a base64-encoded logo image to include in the report.
	Logo string

	// IncludeRawJSON includes the raw JSON data in the report.
	IncludeRawJSON bool

	// Theme is the color theme (light, dark).
	Theme string

	// GeneratedAt overrides the generation timestamp.
	GeneratedAt time.Time
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() *Options {
	return &Options{
		Theme:       "light",
		GeneratedAt: time.Now(),
	}
}

// Generator creates HTML reports from evaluation data.
type Generator struct {
	tmpl *template.Template
}

// New creates a new report generator.
func New() (*Generator, error) {
	tmpl, err := template.New("report").Funcs(templateFuncs()).Parse(reportTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}
	return &Generator{tmpl: tmpl}, nil
}

// Generate creates an HTML report from an evaluation report.
func (g *Generator) Generate(_ context.Context, eval *types.EvaluationReport, opts *Options) ([]byte, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	data := &reportData{
		Evaluation:  eval,
		Options:     opts,
		GeneratedAt: opts.GeneratedAt,
	}

	if opts.Title == "" && eval.Metadata != nil {
		data.Options.Title = eval.Metadata.DocumentTitle
	}

	if opts.IncludeRawJSON {
		jsonBytes, err := json.MarshalIndent(eval, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshaling JSON: %w", err)
		}
		data.RawJSON = string(jsonBytes)
	}

	var buf bytes.Buffer
	if err := g.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateFromFile reads an evaluation JSON file and generates an HTML report.
// It accepts both structured-evaluation EvaluationReport JSON and score-profile
// StyleGuideReport JSON, converting the latter automatically.
func (g *Generator) GenerateFromFile(_ context.Context, path string, opts *Options) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	eval, err := ParseEvaluationJSON(data)
	if err != nil {
		return nil, err
	}

	return g.Generate(context.Background(), eval, opts)
}

// ParseEvaluationJSON parses evaluation JSON in either the
// structured-evaluation EvaluationReport format or the score-profile
// StyleGuideReport format, distinguished by their discriminator fields
// (reviewType vs profileName).
func ParseEvaluationJSON(data []byte) (*types.EvaluationReport, error) {
	var probe struct {
		ReviewType  string `json:"reviewType"`
		ProfileName string `json:"profileName"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if probe.ReviewType == "" && probe.ProfileName != "" {
		var sg judge.StyleGuideReport
		if err := json.Unmarshal(data, &sg); err != nil {
			return nil, fmt.Errorf("parsing style guide report JSON: %w", err)
		}
		return sg.ToEvaluationReport(), nil
	}

	var eval types.EvaluationReport
	if err := json.Unmarshal(data, &eval); err != nil {
		return nil, fmt.Errorf("parsing evaluation JSON: %w", err)
	}
	return &eval, nil
}

// reportData is the data passed to the template.
type reportData struct {
	Evaluation  *types.EvaluationReport
	Options     *Options
	GeneratedAt time.Time
	RawJSON     string
}

// templateFuncs returns template helper functions.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"scoreColor": func(score int) string {
			switch {
			case score >= 5:
				return "#22c55e" // green
			case score >= 4:
				return "#eab308" // yellow
			case score >= 3:
				return "#f97316" // orange
			default:
				return "#ef4444" // red
			}
		},
		"severityColor": func(severity string) string {
			switch severity {
			case "critical":
				return "#ef4444"
			case "high":
				return "#f97316"
			case "medium":
				return "#eab308"
			case "low":
				return "#22c55e"
			default:
				return "#94a3b8"
			}
		},
		"decisionColor": func(decision string) string {
			switch decision {
			case "pass":
				return "#22c55e"
			case "partial":
				return "#eab308"
			case "fail":
				return "#ef4444"
			default:
				return "#94a3b8"
			}
		},
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
		"upper": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s)
		},
	}
}

const reportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{if .Options.Title}}{{.Options.Title}}{{else}}Evaluation Report{{end}}</title>
    <style>
        :root {
            --bg-primary: #ffffff;
            --bg-secondary: #f8fafc;
            --text-primary: #1e293b;
            --text-secondary: #64748b;
            --border-color: #e2e8f0;
            --accent-color: #3b82f6;
        }

        {{if eq .Options.Theme "dark"}}
        :root {
            --bg-primary: #0f172a;
            --bg-secondary: #1e293b;
            --text-primary: #f1f5f9;
            --text-secondary: #94a3b8;
            --border-color: #334155;
        }
        {{end}}

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

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
            body {
                padding: 0;
                font-size: 10pt;
            }
            .page-break {
                page-break-before: always;
            }
            .no-print {
                display: none !important;
            }
        }

        header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 2rem;
            padding-bottom: 1.5rem;
            border-bottom: 2px solid var(--border-color);
        }

        .header-left h1 {
            font-size: 1.75rem;
            font-weight: 700;
            margin-bottom: 0.25rem;
        }

        .header-left .subtitle {
            color: var(--text-secondary);
            font-size: 0.95rem;
        }

        .header-right {
            text-align: right;
        }

        .decision-badge {
            display: inline-block;
            padding: 0.5rem 1.25rem;
            border-radius: 9999px;
            font-weight: 600;
            font-size: 1rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .meta-info {
            margin-top: 0.5rem;
            font-size: 0.85rem;
            color: var(--text-secondary);
        }

        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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
            font-size: 0.85rem;
            color: var(--text-secondary);
            margin-bottom: 0.25rem;
        }

        .summary-card .value {
            font-size: 1.5rem;
            font-weight: 700;
        }

        section {
            margin-bottom: 2.5rem;
        }

        section h2 {
            font-size: 1.25rem;
            font-weight: 600;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 1px solid var(--border-color);
        }

        table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 1rem;
        }

        th, td {
            text-align: left;
            padding: 0.75rem 1rem;
            border-bottom: 1px solid var(--border-color);
        }

        th {
            background: var(--bg-secondary);
            font-weight: 600;
            font-size: 0.85rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--text-secondary);
        }

        tr:hover {
            background: var(--bg-secondary);
        }

        .score-cell {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .score-dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
        }

        .severity-badge {
            display: inline-block;
            padding: 0.2rem 0.6rem;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 500;
            text-transform: uppercase;
        }

        .finding-card {
            background: var(--bg-secondary);
            border-radius: 0.5rem;
            padding: 1rem;
            margin-bottom: 0.75rem;
            border-left: 4px solid;
        }

        .finding-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 0.5rem;
        }

        .finding-category {
            font-size: 0.85rem;
            color: var(--text-secondary);
        }

        .finding-text {
            margin-bottom: 0.5rem;
        }

        .recommendation {
            font-size: 0.9rem;
            color: var(--text-secondary);
            padding-left: 1rem;
            border-left: 2px solid var(--accent-color);
        }

        .next-steps-list {
            list-style: none;
        }

        .next-steps-list li {
            padding: 0.75rem 0;
            border-bottom: 1px solid var(--border-color);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .effort-badge {
            font-size: 0.75rem;
            padding: 0.2rem 0.5rem;
            border-radius: 4px;
            background: var(--bg-secondary);
            color: var(--text-secondary);
        }

        .raw-json {
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            border-radius: 0.5rem;
            padding: 1rem;
            overflow-x: auto;
            font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
            font-size: 0.85rem;
            white-space: pre;
        }

        footer {
            margin-top: 3rem;
            padding-top: 1.5rem;
            border-top: 1px solid var(--border-color);
            text-align: center;
            font-size: 0.85rem;
            color: var(--text-secondary);
        }
    </style>
</head>
<body>
    <header>
        <div class="header-left">
            <h1>{{if .Options.Title}}{{.Options.Title}}{{else}}Evaluation Report{{end}}</h1>
            {{if .Evaluation.Metadata}}
            <div class="subtitle">{{.Evaluation.Metadata.Document}}</div>
            {{end}}
        </div>
        <div class="header-right">
            <span class="decision-badge" style="background: {{decisionColor .Evaluation.OverallDecision}}; color: white;">
                {{upper .Evaluation.OverallDecision}}
            </span>
            <div class="meta-info">
                {{if .Evaluation.Metadata}}
                Evaluated {{formatDate .Evaluation.Metadata.GeneratedAt}}<br>
                by {{.Evaluation.Metadata.GeneratedBy}}
                {{end}}
            </div>
        </div>
    </header>

    <div class="summary-grid">
        <div class="summary-card">
            <div class="label">Categories Passing</div>
            <div class="value" style="color: {{decisionColor .Evaluation.OverallDecision}};">
                {{.Evaluation.Decision.CategoryCounts.Pass}}/{{.Evaluation.Decision.CategoryCounts.Total}}
            </div>
        </div>
        <div class="summary-card">
            <div class="label">Critical/High Findings</div>
            <div class="value">
                {{with .Evaluation.Decision.FindingCounts}}{{.Critical}}/{{.High}}{{end}}
            </div>
        </div>
        <div class="summary-card">
            <div class="label">Total Findings</div>
            <div class="value">{{.Evaluation.Decision.FindingCounts.Total}}</div>
        </div>
        <div class="summary-card">
            <div class="label">Rubric Version</div>
            <div class="value" style="font-size: 1rem;">{{.Evaluation.RubricID}} v{{.Evaluation.RubricVersion}}</div>
        </div>
    </div>

    {{if .Evaluation.Summary}}
    <section>
        <h2>Executive Summary</h2>
        <p>{{.Evaluation.Summary}}</p>
    </section>
    {{end}}

    <section>
        <h2>Category Scores</h2>
        <table>
            <thead>
                <tr>
                    <th>Category</th>
                    <th>Score</th>
                    <th>Assessment</th>
                </tr>
            </thead>
            <tbody>
                {{range .Evaluation.Categories}}
                <tr>
                    <td>{{.Category}}</td>
                    <td>
                        <div class="score-cell">
                            <span class="score-dot" style="background: {{scoreColor .NumericScore}};"></span>
                            {{.NumericScore}}/5
                        </div>
                    </td>
                    <td>{{.Reasoning}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </section>

    {{if .Evaluation.Findings}}
    <section class="page-break">
        <h2>Findings</h2>
        {{range .Evaluation.Findings}}
        <div class="finding-card" style="border-color: {{severityColor .Severity}};">
            <div class="finding-header">
                <span class="severity-badge" style="background: {{severityColor .Severity}}20; color: {{severityColor .Severity}};">
                    {{upper .Severity}}
                </span>
                <span class="finding-category">{{.Category}}</span>
            </div>
            <div class="finding-text">{{.Finding}}</div>
            {{if .Recommendation}}
            <div class="recommendation">{{.Recommendation}}</div>
            {{end}}
        </div>
        {{end}}
    </section>
    {{end}}

    {{if .Evaluation.NextSteps}}
    <section>
        <h2>Recommended Next Steps</h2>
        {{if .Evaluation.NextSteps.Immediate}}
        <h3 style="color: #ef4444; margin-bottom: 0.5rem;">Immediate Actions</h3>
        <ul class="next-steps-list">
            {{range .Evaluation.NextSteps.Immediate}}
            <li>
                <span>{{.Action}}</span>
                <span class="effort-badge">{{.Effort}} effort</span>
            </li>
            {{end}}
        </ul>
        {{end}}

        {{if .Evaluation.NextSteps.Recommended}}
        <h3 style="margin: 1rem 0 0.5rem;">Recommended Improvements</h3>
        <ul class="next-steps-list">
            {{range .Evaluation.NextSteps.Recommended}}
            <li>
                <span>{{.Action}}</span>
                <span class="effort-badge">{{.Effort}} effort</span>
            </li>
            {{end}}
        </ul>
        {{end}}
    </section>
    {{end}}

    {{if .RawJSON}}
    <section class="no-print">
        <h2>Raw Data</h2>
        <pre class="raw-json">{{.RawJSON}}</pre>
    </section>
    {{end}}

    <footer>
        Generated by systemspec-apistyle on {{formatDate .GeneratedAt}}
    </footer>
</body>
</html>
`
