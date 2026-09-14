package judge

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

//go:embed styleguide_rubric.json
var styleGuideRubricJSON []byte

// StyleGuideRubric defines the evaluation criteria for style guide quality.
type StyleGuideRubric struct {
	Schema       string                 `json:"$schema"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
	Categories   []StyleGuideCategory   `json:"categories"`
	PassCriteria StyleGuidePassCriteria `json:"passCriteria,omitempty"`
}

// StyleGuideCategory defines evaluation criteria for one aspect of a style guide.
type StyleGuideCategory struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	Weight           float64             `json:"weight"`
	Required         bool                `json:"required"`
	EvaluationPrompt string              `json:"evaluationPrompt"`
	PassCriteria     []string            `json:"passCriteria"`
	PartialCriteria  []string            `json:"partialCriteria"`
	FailCriteria     []string            `json:"failCriteria"`
	Examples         *StyleGuideExamples `json:"examples,omitempty"`
}

// StyleGuideExamples provides few-shot examples for evaluation.
type StyleGuideExamples struct {
	Pass    *StyleGuideExample `json:"pass,omitempty"`
	Partial *StyleGuideExample `json:"partial,omitempty"`
	Fail    *StyleGuideExample `json:"fail,omitempty"`
}

// StyleGuideExample is a single example for a category.
type StyleGuideExample struct {
	Excerpt   string `json:"excerpt"`
	Reasoning string `json:"reasoning"`
}

// StyleGuidePassCriteria defines overall pass requirements.
type StyleGuidePassCriteria struct {
	MinCategoriesPassing string `json:"minCategoriesPassing"`
	MaxFindings          struct {
		Critical int `json:"critical"`
		High     int `json:"high"`
		Medium   int `json:"medium"`
	} `json:"maxFindings"`
}

// StyleGuideEvaluator evaluates style guides against quality criteria.
type StyleGuideEvaluator struct {
	provider Provider
	rubric   *StyleGuideRubric
}

// StyleGuideReport contains the evaluation results for a style guide.
type StyleGuideReport struct {
	ProfileName  string                     `json:"profileName"`
	Status       string                     `json:"status"`
	OverallScore float64                    `json:"overallScore"`
	Categories   []StyleGuideCategoryResult `json:"categories"`
	Summary      StyleGuideSummary          `json:"summary"`
	Metadata     StyleGuideReportMetadata   `json:"metadata"`
}

// StyleGuideCategoryResult contains results for one category.
type StyleGuideCategoryResult struct {
	Category     string   `json:"category"`
	Name         string   `json:"name"`
	Score        string   `json:"score"`        // "pass", "partial", "fail"
	NumericScore int      `json:"numericScore"` // 1-5
	Weight       float64  `json:"weight"`
	Required     bool     `json:"required"`
	Reasoning    string   `json:"reasoning"`
	Strengths    []string `json:"strengths,omitempty"`
	Weaknesses   []string `json:"weaknesses,omitempty"`
}

// StyleGuideSummary provides aggregate statistics.
type StyleGuideSummary struct {
	TotalCategories   int     `json:"totalCategories"`
	PassedCategories  int     `json:"passedCategories"`
	PartialCategories int     `json:"partialCategories"`
	FailedCategories  int     `json:"failedCategories"`
	WeightedScore     float64 `json:"weightedScore"`
}

// StyleGuideReportMetadata contains evaluation context.
type StyleGuideReportMetadata struct {
	RubricName    string `json:"rubricName"`
	RubricVersion string `json:"rubricVersion"`
	Model         string `json:"model"`
	Duration      string `json:"duration"`
	Timestamp     string `json:"timestamp"`
}

// NewStyleGuideEvaluator creates a new style guide evaluator.
func NewStyleGuideEvaluator(provider Provider) (*StyleGuideEvaluator, error) {
	var rubric StyleGuideRubric
	if err := json.Unmarshal(styleGuideRubricJSON, &rubric); err != nil {
		return nil, fmt.Errorf("parsing embedded rubric: %w", err)
	}

	return &StyleGuideEvaluator{
		provider: provider,
		rubric:   &rubric,
	}, nil
}

// NewStyleGuideEvaluatorWithRubric creates an evaluator with a custom rubric.
func NewStyleGuideEvaluatorWithRubric(provider Provider, rubric *StyleGuideRubric) *StyleGuideEvaluator {
	return &StyleGuideEvaluator{
		provider: provider,
		rubric:   rubric,
	}
}

// Evaluate assesses a style guide against the quality rubric.
func (e *StyleGuideEvaluator) Evaluate(ctx context.Context, spec *types.APIStyleSpec, opts *StyleGuideEvalOptions) (*StyleGuideReport, error) {
	if opts == nil {
		opts = DefaultStyleGuideEvalOptions()
	}

	startTime := time.Now()

	report := &StyleGuideReport{
		ProfileName: spec.Name,
		Categories:  make([]StyleGuideCategoryResult, 0, len(e.rubric.Categories)),
		Metadata: StyleGuideReportMetadata{
			RubricName:    e.rubric.Name,
			RubricVersion: e.rubric.Version,
			Model:         opts.Model,
			Timestamp:     startTime.Format(time.RFC3339),
		},
	}

	// Serialize spec to JSON for evaluation
	specJSON, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serializing spec: %w", err)
	}

	// Evaluate each category
	for _, cat := range e.rubric.Categories {
		result, err := e.evaluateCategory(ctx, cat, string(specJSON), opts)
		if err != nil {
			return nil, fmt.Errorf("evaluating category %q: %w", cat.ID, err)
		}
		report.Categories = append(report.Categories, *result)
	}

	// Calculate summary
	e.calculateSummary(report)

	report.Metadata.Duration = time.Since(startTime).Round(time.Millisecond).String()

	return report, nil
}

// StyleGuideEvalOptions configures style guide evaluation.
type StyleGuideEvalOptions struct {
	Model       string
	MaxTokens   int
	Temperature float64
}

// DefaultStyleGuideEvalOptions returns default options.
func DefaultStyleGuideEvalOptions() *StyleGuideEvalOptions {
	return &StyleGuideEvalOptions{
		Model:       "",
		MaxTokens:   2048,
		Temperature: 0.3,
	}
}

func (e *StyleGuideEvaluator) evaluateCategory(ctx context.Context, cat StyleGuideCategory, specJSON string, opts *StyleGuideEvalOptions) (*StyleGuideCategoryResult, error) {
	systemPrompt := e.buildSystemPrompt(cat)
	userPrompt := e.buildUserPrompt(cat, specJSON)

	resp, err := e.provider.Complete(ctx, &CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Model:        opts.Model,
		MaxTokens:    opts.MaxTokens,
		Temperature:  opts.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM completion: %w", err)
	}

	return e.parseResponse(cat, resp.Content)
}

func (e *StyleGuideEvaluator) buildSystemPrompt(cat StyleGuideCategory) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "You are an expert evaluator of API style guides. Your task is to evaluate a style guide specification against the %q quality category.\n\n", cat.Name)
	sb.WriteString("You must respond with a JSON object in exactly this format:\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"score\": \"pass\" | \"partial\" | \"fail\",\n")
	sb.WriteString("  \"numericScore\": 1-5,\n")
	sb.WriteString("  \"reasoning\": \"Brief explanation of the score\",\n")
	sb.WriteString("  \"strengths\": [\"strength 1\", \"strength 2\"],\n")
	sb.WriteString("  \"weaknesses\": [\"weakness 1\", \"weakness 2\"]\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")
	sb.WriteString("Scoring guide:\n")
	sb.WriteString("- pass (4-5): Meets most or all criteria\n")
	sb.WriteString("- partial (2-3): Meets some criteria but has gaps\n")
	sb.WriteString("- fail (1): Does not meet criteria\n")

	return sb.String()
}

func (e *StyleGuideEvaluator) buildUserPrompt(cat StyleGuideCategory, specJSON string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Category: %s\n\n", cat.Name)
	fmt.Fprintf(&sb, "**Description:** %s\n\n", cat.Description)
	fmt.Fprintf(&sb, "**Evaluation Task:** %s\n\n", cat.EvaluationPrompt)

	sb.WriteString("### Pass Criteria (score 4-5):\n")
	for _, c := range cat.PassCriteria {
		fmt.Fprintf(&sb, "- %s\n", c)
	}

	sb.WriteString("\n### Partial Criteria (score 2-3):\n")
	for _, c := range cat.PartialCriteria {
		fmt.Fprintf(&sb, "- %s\n", c)
	}

	sb.WriteString("\n### Fail Criteria (score 1):\n")
	for _, c := range cat.FailCriteria {
		fmt.Fprintf(&sb, "- %s\n", c)
	}

	// Add examples if available
	if cat.Examples != nil {
		sb.WriteString("\n### Examples:\n")
		if cat.Examples.Pass != nil {
			fmt.Fprintf(&sb, "**Pass example:** %s\n*Reasoning:* %s\n\n", cat.Examples.Pass.Excerpt, cat.Examples.Pass.Reasoning)
		}
		if cat.Examples.Partial != nil {
			fmt.Fprintf(&sb, "**Partial example:** %s\n*Reasoning:* %s\n\n", cat.Examples.Partial.Excerpt, cat.Examples.Partial.Reasoning)
		}
		if cat.Examples.Fail != nil {
			fmt.Fprintf(&sb, "**Fail example:** %s\n*Reasoning:* %s\n\n", cat.Examples.Fail.Excerpt, cat.Examples.Fail.Reasoning)
		}
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString("## Style Guide Specification to Evaluate:\n\n")
	sb.WriteString("```json\n")
	// Truncate if too long (keep first 20000 chars)
	if len(specJSON) > 20000 {
		sb.WriteString(specJSON[:20000])
		sb.WriteString("\n... [truncated for length]\n")
	} else {
		sb.WriteString(specJSON)
	}
	sb.WriteString("\n```\n\n")
	sb.WriteString("Please evaluate this style guide against the criteria above and respond with JSON only.")

	return sb.String()
}

func (e *StyleGuideEvaluator) parseResponse(cat StyleGuideCategory, content string) (*StyleGuideCategoryResult, error) {
	result := &StyleGuideCategoryResult{
		Category: cat.ID,
		Name:     cat.Name,
		Weight:   cat.Weight,
		Required: cat.Required,
	}

	// Extract JSON from response (handle markdown code blocks)
	jsonContent := content
	if idx := strings.Index(content, "```json"); idx >= 0 {
		jsonContent = content[idx+7:]
		if endIdx := strings.Index(jsonContent, "```"); endIdx >= 0 {
			jsonContent = jsonContent[:endIdx]
		}
	} else if idx := strings.Index(content, "```"); idx >= 0 {
		jsonContent = content[idx+3:]
		if endIdx := strings.Index(jsonContent, "```"); endIdx >= 0 {
			jsonContent = jsonContent[:endIdx]
		}
	}

	jsonContent = strings.TrimSpace(jsonContent)

	var parsed struct {
		Score        string   `json:"score"`
		NumericScore int      `json:"numericScore"`
		Reasoning    string   `json:"reasoning"`
		Strengths    []string `json:"strengths"`
		Weaknesses   []string `json:"weaknesses"`
	}

	if parseErr := json.Unmarshal([]byte(jsonContent), &parsed); parseErr != nil {
		// Log parse error and fall back to extracting score from text
		slog.Debug("JSON parse failed, extracting score from text", "error", parseErr, "category", cat.ID)
		result.Score = extractScoreFromText(content)
		result.NumericScore = scoreToNumeric(result.Score)
		result.Reasoning = "Unable to parse structured response: " + content[:min(200, len(content))]
		return result, nil
	}

	result.Score = parsed.Score
	result.NumericScore = parsed.NumericScore
	result.Reasoning = parsed.Reasoning
	result.Strengths = parsed.Strengths
	result.Weaknesses = parsed.Weaknesses

	// Ensure numeric score is set
	if result.NumericScore == 0 {
		result.NumericScore = scoreToNumeric(result.Score)
	}

	return result, nil
}

func extractScoreFromText(text string) string {
	text = strings.ToLower(text)
	if strings.Contains(text, "\"score\": \"pass\"") || strings.Contains(text, "score: pass") {
		return "pass"
	}
	if strings.Contains(text, "\"score\": \"partial\"") || strings.Contains(text, "score: partial") {
		return "partial"
	}
	if strings.Contains(text, "\"score\": \"fail\"") || strings.Contains(text, "score: fail") {
		return "fail"
	}
	// Default to partial if unclear
	return "partial"
}

func scoreToNumeric(score string) int {
	switch strings.ToLower(score) {
	case "pass":
		return 4
	case "partial":
		return 3
	case "fail":
		return 1
	default:
		return 3
	}
}

func (e *StyleGuideEvaluator) calculateSummary(report *StyleGuideReport) {
	var totalWeight, weightedSum float64
	var passed, partial, failed int

	for _, cat := range report.Categories {
		weight := cat.Weight
		if weight == 0 {
			weight = 1.0
		}
		totalWeight += weight
		weightedSum += float64(cat.NumericScore) * weight

		switch cat.Score {
		case "pass":
			passed++
		case "partial":
			partial++
		case "fail":
			failed++
		}
	}

	report.Summary = StyleGuideSummary{
		TotalCategories:   len(report.Categories),
		PassedCategories:  passed,
		PartialCategories: partial,
		FailedCategories:  failed,
	}

	if totalWeight > 0 {
		report.Summary.WeightedScore = weightedSum / totalWeight
		report.OverallScore = report.Summary.WeightedScore
	}

	// Determine overall status
	switch {
	case failed > 0:
		report.Status = "fail"
	case partial > 0:
		report.Status = "partial"
	default:
		report.Status = "pass"
	}
}

// LoadStyleGuideRubric loads the rubric from embedded JSON.
func LoadStyleGuideRubric() (*StyleGuideRubric, error) {
	var rubric StyleGuideRubric
	if err := json.Unmarshal(styleGuideRubricJSON, &rubric); err != nil {
		return nil, fmt.Errorf("parsing embedded rubric: %w", err)
	}
	return &rubric, nil
}

// LoadStyleGuideRubricFromBytes loads a rubric from JSON bytes.
func LoadStyleGuideRubricFromBytes(data []byte) (*StyleGuideRubric, error) {
	var rubric StyleGuideRubric
	if err := json.Unmarshal(data, &rubric); err != nil {
		return nil, fmt.Errorf("parsing rubric: %w", err)
	}
	return &rubric, nil
}

// NumericScoreFromText extracts a numeric score from various text formats.
func NumericScoreFromText(text string) int {
	// Try to find patterns like "4/5", "4 out of 5", "score: 4"
	patterns := []string{
		`(\d)/5`,
		`(\d)\s*out\s*of\s*5`,
		`score[:\s]+(\d)`,
		`numericScore[:\s]+(\d)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			if score, err := strconv.Atoi(matches[1]); err == nil && score >= 1 && score <= 5 {
				return score
			}
		}
	}

	return 0
}
