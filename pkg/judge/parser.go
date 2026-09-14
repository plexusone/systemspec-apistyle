package judge

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// ErrNoJSONFound indicates no JSON block was found in the response.
var ErrNoJSONFound = errors.New("no JSON block found in response")

// SingleEvaluationResponse is the expected JSON structure for single rule evaluation.
type SingleEvaluationResponse struct {
	Score       float64  `json:"score"`
	Passed      bool     `json:"passed"`
	Reasoning   string   `json:"reasoning"`
	Examples    []string `json:"examples"`
	Suggestions []string `json:"suggestions"`
	Locations   []string `json:"locations"`
}

// CategoryEvaluationResponse is the expected JSON structure for category evaluation.
type CategoryEvaluationResponse struct {
	Findings []FindingResponse `json:"findings"`
}

// FindingResponse is the expected JSON structure for individual findings.
type FindingResponse struct {
	RuleID      string   `json:"ruleId"`
	Score       float64  `json:"score"`
	Passed      bool     `json:"passed"`
	Reasoning   string   `json:"reasoning"`
	Examples    []string `json:"examples"`
	Suggestions []string `json:"suggestions"`
	Locations   []string `json:"locations"`
}

// jsonBlockPattern matches JSON code blocks in markdown.
var jsonBlockPattern = regexp.MustCompile("(?s)```json\\s*\\n(.+?)\\n```")

// ParseSingleEvaluation extracts a single evaluation result from LLM response.
func ParseSingleEvaluation(response string, criterion *Criterion) (*Finding, error) {
	jsonStr, err := extractJSON(response)
	if err != nil {
		return nil, fmt.Errorf("extracting JSON: %w", err)
	}

	var evalResp SingleEvaluationResponse
	if err := json.Unmarshal([]byte(jsonStr), &evalResp); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return &Finding{
		RuleID:      criterion.RuleID,
		RuleTitle:   criterion.RuleTitle,
		Category:    criterion.Category,
		Score:       normalizeScore(evalResp.Score),
		Passed:      evalResp.Passed || evalResp.Score >= 0.5,
		Reasoning:   evalResp.Reasoning,
		Examples:    evalResp.Examples,
		Suggestions: evalResp.Suggestions,
		Locations:   evalResp.Locations,
		Severity:    criterion.Severity,
		Weight:      criterion.Weight,
	}, nil
}

// ParseCategoryEvaluation extracts multiple findings from LLM response.
func ParseCategoryEvaluation(response string, criteria []*Criterion) ([]Finding, error) {
	jsonStr, err := extractJSON(response)
	if err != nil {
		return nil, fmt.Errorf("extracting JSON: %w", err)
	}

	var catResp CategoryEvaluationResponse
	if err := json.Unmarshal([]byte(jsonStr), &catResp); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	// Build lookup for criteria
	criteriaMap := make(map[string]*Criterion)
	for _, c := range criteria {
		criteriaMap[c.RuleID] = c
	}

	findings := make([]Finding, 0, len(catResp.Findings))
	for _, fr := range catResp.Findings {
		c, ok := criteriaMap[fr.RuleID]
		if !ok {
			// Unknown rule ID, skip
			continue
		}

		findings = append(findings, Finding{
			RuleID:      fr.RuleID,
			RuleTitle:   c.RuleTitle,
			Category:    c.Category,
			Score:       normalizeScore(fr.Score),
			Passed:      fr.Passed || fr.Score >= 0.5,
			Reasoning:   fr.Reasoning,
			Examples:    fr.Examples,
			Suggestions: fr.Suggestions,
			Locations:   fr.Locations,
			Severity:    c.Severity,
			Weight:      c.Weight,
		})
	}

	return findings, nil
}

// extractJSON finds and extracts JSON content from the response.
func extractJSON(response string) (string, error) {
	// Try to find JSON in code block first
	matches := jsonBlockPattern.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}

	// Try to find raw JSON (object or array)
	response = strings.TrimSpace(response)

	// Look for JSON object
	start := strings.Index(response, "{")
	if start >= 0 {
		// Find matching closing brace
		depth := 0
		for i := start; i < len(response); i++ {
			switch response[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return response[start : i+1], nil
				}
			}
		}
	}

	// Look for JSON array
	start = strings.Index(response, "[")
	if start >= 0 {
		depth := 0
		for i := start; i < len(response); i++ {
			switch response[i] {
			case '[':
				depth++
			case ']':
				depth--
				if depth == 0 {
					return response[start : i+1], nil
				}
			}
		}
	}

	return "", ErrNoJSONFound
}

// normalizeScore ensures score is within [0.0, 1.0].
func normalizeScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

// ScoreToStatus converts a numeric score to a pass/fail status.
func ScoreToStatus(score float64) types.Status {
	if score >= 0.5 {
		return types.StatusPass
	}
	return types.StatusFail
}

// AggregateScores computes an aggregate score from multiple scores with weights.
func AggregateScores(scores []float64, weights []float64) float64 {
	if len(scores) == 0 {
		return 0
	}

	if len(weights) == 0 || len(weights) != len(scores) {
		// Equal weights
		var sum float64
		for _, s := range scores {
			sum += s
		}
		return sum / float64(len(scores))
	}

	var weightedSum, totalWeight float64
	for i, s := range scores {
		weightedSum += s * weights[i]
		totalWeight += weights[i]
	}

	if totalWeight == 0 {
		return 0
	}
	return weightedSum / totalWeight
}
