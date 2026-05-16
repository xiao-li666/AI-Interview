package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"ai-interview/backend/internal/model"
	"ai-interview/backend/internal/repository"
)

type InterviewRepository struct {
	mu                sync.RWMutex
	nextUserID        int64
	nextPlanID        int64
	nextSessionID     int64
	nextQuestionID    int64
	nextAnswerID      int64
	nextFeedbackID    int64
	nextReportID      int64
	usersByID         map[int64]model.User
	userIDsByEmail    map[string]int64
	plans             map[int64]model.InterviewPlan
	sessions          map[int64]model.InterviewSession
	questions         map[int64][]model.InterviewQuestion
	answers           map[int64][]model.InterviewAnswer
	feedbacks         map[int64]model.AnswerFeedback
	reports           map[int64]model.InterviewReport
	historyByUserID   map[int64][]model.HistoryItem
	aiConfigsByUserID map[int64]model.UserAIProviderConfig
}

func NewInterviewRepository() *InterviewRepository {
	return &InterviewRepository{
		nextUserID:        1,
		nextPlanID:        1,
		nextSessionID:     1,
		nextQuestionID:    1,
		nextAnswerID:      1,
		nextFeedbackID:    1,
		nextReportID:      1,
		usersByID:         make(map[int64]model.User),
		userIDsByEmail:    make(map[string]int64),
		plans:             make(map[int64]model.InterviewPlan),
		sessions:          make(map[int64]model.InterviewSession),
		questions:         make(map[int64][]model.InterviewQuestion),
		answers:           make(map[int64][]model.InterviewAnswer),
		feedbacks:         make(map[int64]model.AnswerFeedback),
		reports:           make(map[int64]model.InterviewReport),
		historyByUserID:   make(map[int64][]model.HistoryItem),
		aiConfigsByUserID: make(map[int64]model.UserAIProviderConfig),
	}
}

func (r *InterviewRepository) CreateUser(_ context.Context, user model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	email := strings.ToLower(strings.TrimSpace(user.Email))
	if email == "" {
		return model.User{}, repository.ErrNotFound
	}
	if _, exists := r.userIDsByEmail[email]; exists {
		return model.User{}, repository.ErrNotFound
	}

	now := time.Now()
	user.ID = r.nextUserID
	r.nextUserID++
	user.Email = email
	user.CreatedAt = now
	user.UpdatedAt = now

	r.usersByID[user.ID] = user
	r.userIDsByEmail[email] = user.ID
	return user, nil
}

func (r *InterviewRepository) GetUserByEmail(_ context.Context, email string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userID, ok := r.userIDsByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}

	user, ok := r.usersByID[userID]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}

	return user, nil
}

func (r *InterviewRepository) GetUserByID(_ context.Context, userID int64) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.usersByID[userID]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}

	return user, nil
}

func (r *InterviewRepository) UpdateUser(_ context.Context, user model.User) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.usersByID[user.ID]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}

	existing.Nickname = user.Nickname
	existing.AvatarURL = user.AvatarURL
	existing.Status = user.Status
	existing.UpdatedAt = time.Now()

	r.usersByID[user.ID] = existing
	return existing, nil
}

func (r *InterviewRepository) CreatePlan(_ context.Context, plan model.InterviewPlan) (model.InterviewPlan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	plan.ID = r.nextPlanID
	r.nextPlanID++
	r.plans[plan.ID] = plan

	return plan, nil
}

func (r *InterviewRepository) CreateSession(_ context.Context, session model.InterviewSession, questions []model.InterviewQuestion) (model.SessionDetail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session.ID = r.nextSessionID
	r.nextSessionID++
	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now
	r.sessions[session.ID] = session

	boundQuestions := make([]model.InterviewQuestion, 0, len(questions))
	for _, question := range questions {
		question.ID = r.nextQuestionID
		r.nextQuestionID++
		question.SessionID = session.ID
		boundQuestions = append(boundQuestions, question)
	}
	r.questions[session.ID] = boundQuestions

	r.historyByUserID[session.UserID] = append(r.historyByUserID[session.UserID], model.HistoryItem{
		SessionID:   session.ID,
		SessionName: session.SessionName,
		JobTitle:    session.JobTitle,
		RoundType:   session.RoundType,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt,
	})

	return model.SessionDetail{
		Session:   session,
		Questions: boundQuestions,
	}, nil
}

func (r *InterviewRepository) GetSession(_ context.Context, sessionID int64) (model.SessionDetail, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return model.SessionDetail{}, repository.ErrNotFound
	}

	return model.SessionDetail{
		Session:   session,
		Questions: append([]model.InterviewQuestion(nil), r.questions[sessionID]...),
	}, nil
}

func (r *InterviewRepository) GetNextQuestion(_ context.Context, sessionID int64) (model.InterviewQuestion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return model.InterviewQuestion{}, repository.ErrNotFound
	}

	questionIndex := session.CurrentQuestionNo
	if questionIndex >= len(r.questions[sessionID]) {
		return model.InterviewQuestion{}, repository.ErrNotFound
	}

	nextQuestion := r.questions[sessionID][questionIndex]
	session.CurrentQuestionNo++
	session.Status = "in_progress"
	now := time.Now()
	if session.StartedAt == nil {
		session.StartedAt = &now
	}
	session.UpdatedAt = now
	r.sessions[sessionID] = session
	r.updateHistoryLocked(session)

	return nextQuestion, nil
}

func (r *InterviewRepository) SaveAnswer(_ context.Context, answer model.InterviewAnswer, feedback model.AnswerFeedback) (model.InterviewAnswer, model.AnswerFeedback, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[answer.SessionID]
	if !ok {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, repository.ErrNotFound
	}

	answer.ID = r.nextAnswerID
	r.nextAnswerID++
	r.answers[answer.SessionID] = append(r.answers[answer.SessionID], answer)

	feedback.ID = r.nextFeedbackID
	r.nextFeedbackID++
	feedback.AnswerID = answer.ID
	r.feedbacks[answer.ID] = feedback

	if len(r.answers[answer.SessionID]) >= session.QuestionCount {
		session.Status = "completed"
		now := time.Now()
		session.EndedAt = &now
		session.UpdatedAt = now
		r.sessions[answer.SessionID] = session
		r.updateHistoryLocked(session)
	}

	return answer, feedback, nil
}

func (r *InterviewRepository) ListSessionAnswers(_ context.Context, sessionID int64) ([]model.SessionAnswerItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.sessions[sessionID]; !ok {
		return nil, repository.ErrNotFound
	}

	questionIndex := make(map[int64]model.InterviewQuestion, len(r.questions[sessionID]))
	for _, question := range r.questions[sessionID] {
		questionIndex[question.ID] = question
	}

	items := make([]model.SessionAnswerItem, 0, len(r.answers[sessionID]))
	for _, answer := range r.answers[sessionID] {
		question := questionIndex[answer.QuestionID]
		feedback, ok := r.feedbacks[answer.ID]

		item := model.SessionAnswerItem{
			QuestionID:          question.ID,
			QuestionNo:          question.QuestionNo,
			QuestionType:        question.QuestionType,
			AssessmentDimension: question.AssessmentDimension,
			Prompt:              question.Prompt,
			ExpectedPoints:      append([]string(nil), question.ExpectedPoints...),
			Answer:              answer,
		}
		if ok {
			feedbackCopy := feedback
			item.Feedback = &feedbackCopy
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *InterviewRepository) SaveReport(_ context.Context, report model.InterviewReport) (model.InterviewReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[report.SessionID]
	if !ok {
		return model.InterviewReport{}, repository.ErrNotFound
	}

	if existing, ok := r.reports[report.SessionID]; ok {
		report.ID = existing.ID
	}
	if report.ID == 0 {
		report.ID = r.nextReportID
		r.nextReportID++
	}
	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}

	r.reports[report.SessionID] = report
	r.updateHistoryLocked(session)
	return report, nil
}

func (r *InterviewRepository) GetReport(_ context.Context, sessionID int64) (model.InterviewReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	report, ok := r.reports[sessionID]
	if !ok {
		return model.InterviewReport{}, repository.ErrNotFound
	}

	return report, nil
}

func (r *InterviewRepository) ListHistory(_ context.Context, userID int64) ([]model.HistoryItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := append([]model.HistoryItem(nil), r.historyByUserID[userID]...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	return items, nil
}

func (r *InterviewRepository) GetUserAIConfig(_ context.Context, userID int64) (model.UserAIProviderConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, ok := r.aiConfigsByUserID[userID]
	if !ok {
		return model.UserAIProviderConfig{}, repository.ErrNotFound
	}

	return cfg, nil
}

func (r *InterviewRepository) UpsertUserAIConfig(_ context.Context, cfg model.UserAIProviderConfig) (model.UserAIProviderConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.aiConfigsByUserID[cfg.UserID]
	if ok {
		cfg.ID = existing.ID
		cfg.CreatedAt = existing.CreatedAt
	} else {
		cfg.ID = r.nextPlanID + 100000
		cfg.CreatedAt = time.Now()
	}
	cfg.UpdatedAt = time.Now()
	r.aiConfigsByUserID[cfg.UserID] = cfg
	return cfg, nil
}

func (r *InterviewRepository) updateHistoryLocked(session model.InterviewSession) {
	items := r.historyByUserID[session.UserID]
	for index := range items {
		if items[index].SessionID == session.ID {
			items[index].Status = session.Status
			items[index].CompletedAt = session.EndedAt
			if report, ok := r.reports[session.ID]; ok {
				score := report.OverallScore
				items[index].OverallScore = &score
			}
			r.historyByUserID[session.UserID] = items
			return
		}
	}
}
