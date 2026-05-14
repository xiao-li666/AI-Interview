package repository

import (
	"context"

	"ai-interview/backend/internal/model"
)

type InterviewRepository interface {
	CreatePlan(ctx context.Context, plan model.InterviewPlan) (model.InterviewPlan, error)
	CreateSession(ctx context.Context, session model.InterviewSession, questions []model.InterviewQuestion) (model.SessionDetail, error)
	GetSession(ctx context.Context, sessionID int64) (model.SessionDetail, error)
	GetNextQuestion(ctx context.Context, sessionID int64) (model.InterviewQuestion, error)
	SaveAnswer(ctx context.Context, answer model.InterviewAnswer, feedback model.AnswerFeedback) (model.InterviewAnswer, model.AnswerFeedback, error)
	ListSessionAnswers(ctx context.Context, sessionID int64) ([]model.SessionAnswerItem, error)
	SaveReport(ctx context.Context, report model.InterviewReport) (model.InterviewReport, error)
	GetReport(ctx context.Context, sessionID int64) (model.InterviewReport, error)
	ListHistory(ctx context.Context, userID int64) ([]model.HistoryItem, error)
	GetUserAIConfig(ctx context.Context, userID int64) (model.UserAIProviderConfig, error)
	UpsertUserAIConfig(ctx context.Context, cfg model.UserAIProviderConfig) (model.UserAIProviderConfig, error)
}
