package judge

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/plexusone/systemspec-apistyle/pkg/types"
)

// maxSpecChars caps the spec content sent to the LLM.
const maxSpecChars = 100000

// ClaudeEvaluator implements Evaluator using the Anthropic Claude API.
type ClaudeEvaluator struct {
	provider  Provider
	rubricSet *RubricSet
	builder   *PromptBuilder
}

// NewClaudeEvaluator creates a new evaluator with a provider and style spec.
func NewClaudeEvaluator(provider Provider, spec *types.APIStyleSpec) *ClaudeEvaluator {
	return &ClaudeEvaluator{
		provider:  provider,
		rubricSet: BuildRubricSet(spec),
		builder:   NewPromptBuilder(),
	}
}

// NewClaudeEvaluatorWithRubric creates an evaluator with a pre-built rubric set.
func NewClaudeEvaluatorWithRubric(provider Provider, rubricSet *RubricSet) *ClaudeEvaluator {
	return &ClaudeEvaluator{
		provider:  provider,
		rubricSet: rubricSet,
		builder:   NewPromptBuilder(),
	}
}

// Evaluate assesses an API specification against all criteria.
func (e *ClaudeEvaluator) Evaluate(ctx context.Context, specBytes []byte, opts *Options) (*EvaluationReport, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	start := time.Now()
	report := NewEvaluationReport()

	// Determine which criteria to evaluate
	criteria := e.selectCriteria(opts)
	if len(criteria) == 0 {
		report.Metadata = ReportMetadata{
			FileName:    opts.FileName,
			ProfileName: e.rubricSet.Name,
			Model:       e.getModel(opts),
			Duration:    time.Since(start).String(),
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
		return report, nil
	}

	// Prepare spec content
	specContent, truncated := TruncateSpec(string(specBytes), maxSpecChars)
	if truncated {
		slog.Warn("spec exceeds size limit; evaluating truncated content",
			"file", opts.FileName, "limit_chars", maxSpecChars)
	}

	// Evaluate each criterion
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, max(1, opts.MaxConcurrency))
	errChan := make(chan error, len(criteria))

	for _, c := range criteria {
		wg.Add(1)
		go func(criterion *Criterion) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			finding, err := e.evaluateCriterion(ctx, criterion, specContent, opts)
			if err != nil {
				errChan <- fmt.Errorf("evaluating %s: %w", criterion.RuleID, err)
				return
			}

			mu.Lock()
			report.AddFinding(*finding)
			mu.Unlock()
		}(c)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var evalErr error
	for err := range errChan {
		if evalErr == nil {
			evalErr = err
		}
	}

	// Calculate final scores
	report.CalculateScores()

	// Set metadata
	report.Metadata = ReportMetadata{
		FileName:      opts.FileName,
		ProfileName:   e.rubricSet.Name,
		Model:         e.getModel(opts),
		Duration:      time.Since(start).String(),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		SpecTruncated: truncated,
	}

	return report, evalErr
}

// EvaluateCategory evaluates a single category of rules.
func (e *ClaudeEvaluator) EvaluateCategory(ctx context.Context, specBytes []byte, category string, opts *Options) (*CategoryResult, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	criteria := e.rubricSet.FilterByCategory(category)
	if len(criteria) == 0 {
		return &CategoryResult{Name: category}, nil
	}

	specContent, truncated := TruncateSpec(string(specBytes), maxSpecChars)
	if truncated {
		slog.Warn("spec exceeds size limit; evaluating truncated content",
			"file", opts.FileName, "category", category, "limit_chars", maxSpecChars)
	}

	// Build category evaluation prompt
	prompt := e.builder.BuildCategoryEvaluation(category, criteria, specContent)

	req := &CompletionRequest{
		SystemPrompt: e.builder.SystemPrompt,
		UserPrompt:   prompt,
		Model:        e.getModel(opts),
		Temperature:  opts.Temperature,
		MaxTokens:    opts.MaxTokens,
	}

	resp, err := e.provider.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM completion: %w", err)
	}

	findings, err := ParseCategoryEvaluation(resp.Content, criteria)
	if err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	// Calculate category score
	var totalWeight, weightedSum float64
	for _, f := range findings {
		weight := f.Weight
		if weight == 0 {
			weight = 1.0
		}
		totalWeight += weight
		weightedSum += f.Score * weight
	}

	score := 0.0
	if totalWeight > 0 {
		score = weightedSum / totalWeight
	}

	return &CategoryResult{
		Name:     category,
		Score:    score,
		Findings: findings,
	}, nil
}

// evaluateCriterion evaluates a single criterion.
func (e *ClaudeEvaluator) evaluateCriterion(ctx context.Context, criterion *Criterion, specContent string, opts *Options) (*Finding, error) {
	prompt := e.builder.BuildSingleEvaluation(criterion, specContent)

	req := &CompletionRequest{
		SystemPrompt: e.builder.SystemPrompt,
		UserPrompt:   prompt,
		Model:        e.getModel(opts),
		Temperature:  opts.Temperature,
		MaxTokens:    opts.MaxTokens,
	}

	resp, err := e.provider.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM completion: %w", err)
	}

	finding, err := ParseSingleEvaluation(resp.Content, criterion)
	if err != nil {
		// Return a failed finding on parse error
		return &Finding{
			RuleID:    criterion.RuleID,
			RuleTitle: criterion.RuleTitle,
			Category:  criterion.Category,
			Score:     0,
			Passed:    false,
			Reasoning: fmt.Sprintf("Failed to parse LLM response: %v", err),
			Severity:  criterion.Severity,
			Weight:    criterion.Weight,
		}, nil
	}

	return finding, nil
}

// selectCriteria returns criteria based on options filters.
func (e *ClaudeEvaluator) selectCriteria(opts *Options) []*Criterion {
	if len(opts.RuleIDs) > 0 {
		return e.rubricSet.FilterByRuleIDs(opts.RuleIDs)
	}

	if len(opts.Categories) > 0 {
		var result []*Criterion
		for _, cat := range opts.Categories {
			result = append(result, e.rubricSet.FilterByCategory(cat)...)
		}
		return result
	}

	return e.rubricSet.AllCriteria()
}

// getModel returns the model to use, with fallback to provider default.
func (e *ClaudeEvaluator) getModel(opts *Options) string {
	if opts.Model != "" {
		return opts.Model
	}
	return e.provider.DefaultModel()
}

// RubricSet returns the underlying rubric set.
func (e *ClaudeEvaluator) RubricSet() *RubricSet {
	return e.rubricSet
}
