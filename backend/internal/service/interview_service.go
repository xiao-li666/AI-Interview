package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"ai-interview/backend/internal/ai"
	"ai-interview/backend/internal/model"
	"ai-interview/backend/internal/repository"
)

var ErrNotFound = errors.New("resource not found")

type InterviewService struct {
	repo            repository.InterviewRepository
	providerFactory *ProviderFactory
}

type CreatePlanRequest struct {
	UserID           int64  `json:"userId"`
	JobTitle         string `json:"jobTitle"`
	JobCategory      string `json:"jobCategory"`
	LevelCode        string `json:"levelCode"`
	InterviewType    string `json:"interviewType"`
	CompanyName      string `json:"companyName"`
	ResumeFileName   string `json:"resumeFileName"`
	ResumeText       string `json:"resumeText"`
	InterviewMode    string `json:"interviewMode"`
	DifficultyLevel  string `json:"difficultyLevel"`
	QuestionCount    int    `json:"questionCount"`
	InterviewerStyle string `json:"interviewerStyle"`
}

type CreateSessionRequest struct {
	UserID           int64  `json:"userId"`
	JobTargetID      int64  `json:"jobTargetId"`
	JobTitle         string `json:"jobTitle"`
	JobCategory      string `json:"jobCategory"`
	LevelCode        string `json:"levelCode"`
	InterviewType    string `json:"interviewType"`
	CompanyName      string `json:"companyName"`
	ResumeFileName   string `json:"resumeFileName"`
	ResumeText       string `json:"resumeText"`
	SessionName      string `json:"sessionName"`
	RoundType        string `json:"roundType"`
	InterviewMode    string `json:"interviewMode"`
	DifficultyLevel  string `json:"difficultyLevel"`
	QuestionCount    int    `json:"questionCount"`
	InterviewerStyle string `json:"interviewerStyle"`
}

type SubmitAnswerRequest struct {
	SessionID             int64  `json:"sessionId"`
	QuestionID            int64  `json:"questionId"`
	AnswerText            string `json:"answerText"`
	AnswerAudioURL        string `json:"answerAudioUrl"`
	AnswerDurationSeconds int    `json:"answerDurationSeconds"`
}

func NewInterviewService(repo repository.InterviewRepository, providerFactory *ProviderFactory) *InterviewService {
	return &InterviewService{
		repo:            repo,
		providerFactory: providerFactory,
	}
}

func (s *InterviewService) CreatePlan(ctx context.Context, req CreatePlanRequest) (model.InterviewPlan, error) {
	if req.UserID == 0 || strings.TrimSpace(req.JobTitle) == "" {
		return model.InterviewPlan{}, errors.New("userId and jobTitle are required")
	}

	if req.QuestionCount <= 0 {
		req.QuestionCount = 5
	}

	plan := model.InterviewPlan{
		UserID:           req.UserID,
		JobTitle:         strings.TrimSpace(req.JobTitle),
		JobCategory:      defaultString(req.JobCategory, "backend"),
		LevelCode:        defaultString(req.LevelCode, "mid"),
		InterviewType:    defaultString(req.InterviewType, "technical_1"),
		InterviewMode:    defaultString(req.InterviewMode, "text"),
		DifficultyLevel:  defaultString(req.DifficultyLevel, "medium"),
		QuestionCount:    req.QuestionCount,
		InterviewerStyle: defaultString(req.InterviewerStyle, "balanced"),
		CreatedAt:        time.Now(),
	}

	return s.repo.CreatePlan(ctx, plan)
}

func (s *InterviewService) CreateSession(ctx context.Context, req CreateSessionRequest) (model.SessionDetail, error) {
	if req.UserID == 0 || strings.TrimSpace(req.SessionName) == "" {
		return model.SessionDetail{}, errors.New("userId and sessionName are required")
	}

	if req.QuestionCount <= 0 {
		req.QuestionCount = 5
	}

	now := time.Now()
	session := model.InterviewSession{
		UserID:            req.UserID,
		JobTargetID:       req.JobTargetID,
		JobTitle:          defaultString(strings.TrimSpace(req.JobTitle), "未命名岗位"),
		JobCategory:       defaultString(req.JobCategory, "backend"),
		LevelCode:         defaultString(req.LevelCode, "mid"),
		InterviewType:     defaultString(req.InterviewType, "technical_1"),
		CompanyName:       strings.TrimSpace(req.CompanyName),
		ResumeFileName:    strings.TrimSpace(req.ResumeFileName),
		ResumeText:        strings.TrimSpace(req.ResumeText),
		SessionName:       strings.TrimSpace(req.SessionName),
		RoundType:         defaultString(req.RoundType, "technical_1"),
		InterviewMode:     defaultString(req.InterviewMode, "text"),
		InterviewerStyle:  defaultString(req.InterviewerStyle, "balanced"),
		DifficultyLevel:   defaultString(req.DifficultyLevel, "medium"),
		QuestionCount:     req.QuestionCount,
		CurrentQuestionNo: 0,
		Status:            "draft",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	provider, err := s.resolveProviderForUser(ctx, session.UserID)
	if err != nil {
		return model.SessionDetail{}, err
	}

	questions, err := provider.GenerateQuestions(ctx, ai.GenerateQuestionsInput{
		JobTitle:            session.JobTitle,
		JobCategory:         session.JobCategory,
		LevelCode:           session.LevelCode,
		InterviewType:       session.InterviewType,
		CompanyName:         session.CompanyName,
		ResumeFileName:      session.ResumeFileName,
		ResumeText:          session.ResumeText,
		CompanyQuestionBank: BuildCompanyQuestionBank(session.CompanyName),
		RoundType:           session.RoundType,
		InterviewMode:       session.InterviewMode,
		DifficultyLevel:     session.DifficultyLevel,
		InterviewerStyle:    session.InterviewerStyle,
		QuestionCount:       session.QuestionCount,
	})
	if err != nil {
		return model.SessionDetail{}, err
	}

	session.QuestionCount = len(questions)
	return s.repo.CreateSession(ctx, session, questions)
}

func (s *InterviewService) GetSession(ctx context.Context, userID int64, sessionID int64) (model.SessionDetail, error) {
	detail, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.SessionDetail{}, ErrNotFound
		}

		return model.SessionDetail{}, err
	}

	if detail.Session.UserID != userID {
		return model.SessionDetail{}, ErrNotFound
	}

	return detail, nil
}

func (s *InterviewService) GetNextQuestion(ctx context.Context, userID int64, sessionID int64) (model.InterviewQuestion, error) {
	detail, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.InterviewQuestion{}, ErrNotFound
		}

		return model.InterviewQuestion{}, err
	}
	if detail.Session.UserID != userID {
		return model.InterviewQuestion{}, ErrNotFound
	}

	question, err := s.repo.GetNextQuestion(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.InterviewQuestion{}, ErrNotFound
		}

		return model.InterviewQuestion{}, err
	}

	return question, nil
}

func (s *InterviewService) SubmitAnswer(ctx context.Context, userID int64, req SubmitAnswerRequest) (model.InterviewAnswer, model.AnswerFeedback, error) {
	if req.SessionID == 0 || req.QuestionID == 0 || strings.TrimSpace(req.AnswerText) == "" {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, errors.New("sessionId, questionId and answerText are required")
	}

	detail, err := s.repo.GetSession(ctx, req.SessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.InterviewAnswer{}, model.AnswerFeedback{}, ErrNotFound
		}

		return model.InterviewAnswer{}, model.AnswerFeedback{}, err
	}
	if detail.Session.UserID != userID {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, ErrNotFound
	}

	question, found := findQuestion(detail.Questions, req.QuestionID)
	if !found {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, ErrNotFound
	}

	answerHistory, err := s.repo.ListSessionAnswers(ctx, req.SessionID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, err
	}

	answer := model.InterviewAnswer{
		SessionID:             req.SessionID,
		QuestionID:            req.QuestionID,
		AnswerText:            req.AnswerText,
		AnswerAudioURL:        req.AnswerAudioURL,
		AnswerDurationSeconds: req.AnswerDurationSeconds,
		SubmittedAt:           time.Now(),
	}

	provider, err := s.resolveProviderForUser(ctx, detail.Session.UserID)
	if err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, err
	}

	feedback, err := provider.EvaluateAnswer(ctx, ai.EvaluateAnswerInput{
		Session:       detail.Session,
		Question:      question,
		AnswerText:    req.AnswerText,
		AnswerHistory: answerHistory,
	})
	if err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, err
	}

	savedAnswer, savedFeedback, err := s.repo.SaveAnswer(ctx, answer, feedback)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.InterviewAnswer{}, model.AnswerFeedback{}, ErrNotFound
		}

		return model.InterviewAnswer{}, model.AnswerFeedback{}, err
	}

	updatedDetail, err := s.repo.GetSession(ctx, req.SessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return savedAnswer, savedFeedback, nil
		}

		return model.InterviewAnswer{}, model.AnswerFeedback{}, err
	}

	if updatedDetail.Session.Status == "completed" {
		updatedHistory, historyErr := s.repo.ListSessionAnswers(ctx, req.SessionID)
		if historyErr != nil && !errors.Is(historyErr, repository.ErrNotFound) {
			return model.InterviewAnswer{}, model.AnswerFeedback{}, historyErr
		}

		report, reportErr := provider.GenerateReport(ctx, ai.GenerateReportInput{
			Session:       updatedDetail.Session,
			Questions:     updatedDetail.Questions,
			AnswerHistory: updatedHistory,
		})
		if reportErr != nil {
			return model.InterviewAnswer{}, model.AnswerFeedback{}, reportErr
		}

		if _, saveErr := s.repo.SaveReport(ctx, report); saveErr != nil {
			return model.InterviewAnswer{}, model.AnswerFeedback{}, saveErr
		}
	}

	return savedAnswer, savedFeedback, nil
}

func (s *InterviewService) GetSessionAnswers(ctx context.Context, userID int64, sessionID int64) ([]model.SessionAnswerItem, error) {
	if sessionID == 0 {
		return nil, errors.New("sessionId is required")
	}

	detail, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if detail.Session.UserID != userID {
		return nil, ErrNotFound
	}

	items, err := s.repo.ListSessionAnswers(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return items, nil
}

func (s *InterviewService) GetQuestionReviews(ctx context.Context, userID int64, sessionID int64) ([]model.QuestionReviewItem, error) {
	if sessionID == 0 {
		return nil, errors.New("sessionId is required")
	}

	detail, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}
	if detail.Session.UserID != userID {
		return nil, ErrNotFound
	}

	answerHistory, err := s.repo.ListSessionAnswers(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	provider, err := s.resolveProviderForUser(ctx, detail.Session.UserID)
	if err != nil {
		return nil, err
	}

	items, err := provider.GenerateQuestionReviews(ctx, ai.GenerateQuestionReviewsInput{
		Session:       detail.Session,
		Questions:     detail.Questions,
		AnswerHistory: answerHistory,
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *InterviewService) GetReport(ctx context.Context, userID int64, sessionID int64) (model.InterviewReport, error) {
	detail, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.InterviewReport{}, ErrNotFound
		}

		return model.InterviewReport{}, err
	}
	if detail.Session.UserID != userID {
		return model.InterviewReport{}, ErrNotFound
	}

	report, err := s.repo.GetReport(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.InterviewReport{}, ErrNotFound
		}

		return model.InterviewReport{}, err
	}

	return report, nil
}

func (s *InterviewService) ListHistory(ctx context.Context, userID int64) ([]model.HistoryItem, error) {
	if userID == 0 {
		return nil, errors.New("userId is required")
	}

	return s.repo.ListHistory(ctx, userID)
}

func (s *InterviewService) resolveProviderForUser(ctx context.Context, userID int64) (ai.Provider, error) {
	cfg, err := s.repo.GetUserAIConfig(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return s.providerFactory.Default(), nil
		}
		return nil, err
	}

	if !cfg.IsEnabled || strings.TrimSpace(cfg.APIKey) == "" {
		return s.providerFactory.Default(), nil
	}

	provider, err := s.providerFactory.BuildFromUserInput(cfg.ProviderCode, cfg.APIKey, cfg.Model, cfg.BaseURL)
	if err != nil {
		return s.providerFactory.Default(), nil
	}

	return ai.NewFallbackProvider(provider, s.providerFactory.Default()), nil
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return strings.TrimSpace(value)
}

func findQuestion(items []model.InterviewQuestion, questionID int64) (model.InterviewQuestion, bool) {
	for _, item := range items {
		if item.ID == questionID {
			return item, true
		}
	}

	return model.InterviewQuestion{}, false
}
