package ai

import (
	"context"
	"fmt"

	"ai-interview/backend/internal/model"
)

type Provider interface {
	Name() string
	GenerateQuestions(ctx context.Context, input GenerateQuestionsInput) ([]model.InterviewQuestion, error)
	EvaluateAnswer(ctx context.Context, input EvaluateAnswerInput) (model.AnswerFeedback, error)
	GenerateReport(ctx context.Context, input GenerateReportInput) (model.InterviewReport, error)
}

type GenerateQuestionsInput struct {
	JobTitle         string `json:"jobTitle"`
	RoundType        string `json:"roundType"`
	InterviewMode    string `json:"interviewMode"`
	DifficultyLevel  string `json:"difficultyLevel"`
	InterviewerStyle string `json:"interviewerStyle"`
	QuestionCount    int    `json:"questionCount"`
}

type EvaluateAnswerInput struct {
	Session       model.InterviewSession    `json:"session"`
	Question      model.InterviewQuestion   `json:"question"`
	AnswerText    string                    `json:"answerText"`
	AnswerHistory []model.SessionAnswerItem `json:"answerHistory"`
}

type GenerateReportInput struct {
	Session       model.InterviewSession    `json:"session"`
	Questions     []model.InterviewQuestion `json:"questions"`
	AnswerHistory []model.SessionAnswerItem `json:"answerHistory"`
}

type FallbackProvider struct {
	primary  Provider
	fallback Provider
}

func NewFallbackProvider(primary Provider, fallback Provider) Provider {
	if primary == nil {
		return fallback
	}
	if fallback == nil {
		return primary
	}

	return &FallbackProvider{
		primary:  primary,
		fallback: fallback,
	}
}

func (p *FallbackProvider) Name() string {
	if p.primary != nil {
		return p.primary.Name()
	}
	if p.fallback != nil {
		return p.fallback.Name()
	}

	return "unknown"
}

func (p *FallbackProvider) GenerateQuestions(ctx context.Context, input GenerateQuestionsInput) ([]model.InterviewQuestion, error) {
	items, err := p.primary.GenerateQuestions(ctx, input)
	if err == nil {
		return items, nil
	}

	fallbackItems, fallbackErr := p.fallback.GenerateQuestions(ctx, input)
	if fallbackErr != nil {
		return nil, fmt.Errorf("primary provider failed: %v; fallback provider failed: %w", err, fallbackErr)
	}

	return fallbackItems, nil
}

func (p *FallbackProvider) EvaluateAnswer(ctx context.Context, input EvaluateAnswerInput) (model.AnswerFeedback, error) {
	feedback, err := p.primary.EvaluateAnswer(ctx, input)
	if err == nil {
		return feedback, nil
	}

	fallbackFeedback, fallbackErr := p.fallback.EvaluateAnswer(ctx, input)
	if fallbackErr != nil {
		return model.AnswerFeedback{}, fmt.Errorf("primary provider failed: %v; fallback provider failed: %w", err, fallbackErr)
	}

	return fallbackFeedback, nil
}

func (p *FallbackProvider) GenerateReport(ctx context.Context, input GenerateReportInput) (model.InterviewReport, error) {
	report, err := p.primary.GenerateReport(ctx, input)
	if err == nil {
		return report, nil
	}

	fallbackReport, fallbackErr := p.fallback.GenerateReport(ctx, input)
	if fallbackErr != nil {
		return model.InterviewReport{}, fmt.Errorf("primary provider failed: %v; fallback provider failed: %w", err, fallbackErr)
	}

	return fallbackReport, nil
}
