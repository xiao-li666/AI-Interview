package repository

import (
	"context"

	"ai-interview/backend/internal/model"
)

type InterviewRepository interface {
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
	GetUserByID(ctx context.Context, userID int64) (model.User, error)
	UpdateUser(ctx context.Context, user model.User) (model.User, error)
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

type AdminRepository interface {
	CreateAdmin(ctx context.Context, admin model.Admin) (model.Admin, error)
	GetAdminByEmail(ctx context.Context, email string) (model.Admin, error)
	GetAdminByID(ctx context.Context, adminID int64) (model.Admin, error)
	GetUserByID(ctx context.Context, userID int64) (model.User, error)
	UpdateUser(ctx context.Context, user model.User) (model.User, error)
	UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error
	GetSession(ctx context.Context, sessionID int64) (model.SessionDetail, error)
	ListSessionAnswers(ctx context.Context, sessionID int64) ([]model.SessionAnswerItem, error)
	GetReport(ctx context.Context, sessionID int64) (model.InterviewReport, error)
	ListAdminUsers(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminUserListItem], error)
	GetAdminUserDetail(ctx context.Context, userID int64) (model.AdminUserDetail, error)
	ListAdminInterviewSessions(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminInterviewSessionListItem], error)
	ListAdminResumes(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminResumeListItem], error)
	GetAdminResumeDetail(ctx context.Context, resumeID int64) (model.AdminResumeDetail, error)
	ListAdminAIConfigs(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminAIConfigListItem], error)
	GetAdminOverview(ctx context.Context) (model.AdminOverview, error)
}
