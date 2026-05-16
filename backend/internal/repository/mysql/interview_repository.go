package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ai-interview/backend/internal/model"
	"ai-interview/backend/internal/repository"

	_ "github.com/go-sql-driver/mysql"
)

type InterviewRepository struct {
	db *sql.DB
}

func NewInterviewRepository(db *sql.DB) *InterviewRepository {
	return &InterviewRepository{db: db}
}

func (r *InterviewRepository) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (email, password_hash, nickname, avatar_url, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.Email,
		user.PasswordHash,
		user.Nickname,
		nullString(user.AvatarURL),
		defaultString(user.Status, "active"),
		now,
		now,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.User{}, fmt.Errorf("read user id: %w", err)
	}

	user.ID = id
	user.Status = defaultString(user.Status, "active")
	user.CreatedAt = now
	user.UpdatedAt = now
	return user, nil
}

func (r *InterviewRepository) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	return r.loadUser(ctx, "email = ?", email)
}

func (r *InterviewRepository) GetUserByID(ctx context.Context, userID int64) (model.User, error) {
	return r.loadUser(ctx, "id = ?", userID)
}

func (r *InterviewRepository) UpdateUser(ctx context.Context, user model.User) (model.User, error) {
	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE users
		SET nickname = ?, avatar_url = ?, status = ?, updated_at = ?
		WHERE id = ?`,
		user.Nickname,
		nullString(user.AvatarURL),
		defaultString(user.Status, "active"),
		now,
		user.ID,
	)
	if err != nil {
		return model.User{}, fmt.Errorf("update user: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return model.User{}, fmt.Errorf("read updated user rows: %w", err)
	}
	if affected == 0 {
		return model.User{}, repository.ErrNotFound
	}

	return r.GetUserByID(ctx, user.ID)
}

func (r *InterviewRepository) UpdateUserPassword(ctx context.Context, userID int64, passwordHash string) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE users
		SET password_hash = ?, updated_at = ?
		WHERE id = ?`,
		passwordHash,
		time.Now(),
		userID,
	)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated password rows: %w", err)
	}
	if affected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *InterviewRepository) CreateAdmin(ctx context.Context, admin model.Admin) (model.Admin, error) {
	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO admins (email, password_hash, nickname, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		admin.Email,
		admin.PasswordHash,
		admin.Nickname,
		defaultString(admin.Status, "active"),
		now,
		now,
	)
	if err != nil {
		return model.Admin{}, fmt.Errorf("insert admin: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Admin{}, fmt.Errorf("read admin id: %w", err)
	}

	admin.ID = id
	admin.Status = defaultString(admin.Status, "active")
	admin.CreatedAt = now
	admin.UpdatedAt = now
	return admin, nil
}

func (r *InterviewRepository) GetAdminByEmail(ctx context.Context, email string) (model.Admin, error) {
	return r.loadAdmin(ctx, "email = ?", email)
}

func (r *InterviewRepository) GetAdminByID(ctx context.Context, adminID int64) (model.Admin, error) {
	return r.loadAdmin(ctx, "id = ?", adminID)
}

func (r *InterviewRepository) CreatePlan(ctx context.Context, plan model.InterviewPlan) (model.InterviewPlan, error) {
	now := time.Now()
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO interview_plans
		(user_id, job_title, job_category, level_code, interview_type, interview_mode, difficulty_level, question_count, interviewer_style, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		plan.UserID,
		plan.JobTitle,
		plan.JobCategory,
		plan.LevelCode,
		plan.InterviewType,
		plan.InterviewMode,
		plan.DifficultyLevel,
		plan.QuestionCount,
		plan.InterviewerStyle,
		now,
		now,
	)
	if err != nil {
		return model.InterviewPlan{}, fmt.Errorf("insert interview plan: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.InterviewPlan{}, fmt.Errorf("read interview plan id: %w", err)
	}

	plan.ID = id
	plan.CreatedAt = now
	return plan, nil
}

func (r *InterviewRepository) CreateSession(ctx context.Context, session model.InterviewSession, questions []model.InterviewQuestion) (model.SessionDetail, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.SessionDetail{}, fmt.Errorf("begin create session tx: %w", err)
	}
	defer rollback(tx)

	now := time.Now()
	resumeID, err := r.upsertResumeTx(ctx, tx, session, now)
	if err != nil {
		return model.SessionDetail{}, err
	}

	jobTargetID, err := r.createJobTargetTx(ctx, tx, session, resumeID, now)
	if err != nil {
		return model.SessionDetail{}, err
	}

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO interview_sessions
		(user_id, job_target_id, resume_id, job_title, session_name, round_type, interview_mode, interviewer_style, difficulty_level, question_count, current_question_no, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.UserID,
		nullInt64(jobTargetID),
		nullInt64(resumeID),
		session.JobTitle,
		session.SessionName,
		session.RoundType,
		session.InterviewMode,
		session.InterviewerStyle,
		session.DifficultyLevel,
		session.QuestionCount,
		session.CurrentQuestionNo,
		session.Status,
		now,
		now,
	)
	if err != nil {
		return model.SessionDetail{}, fmt.Errorf("insert interview session: %w", err)
	}

	sessionID, err := result.LastInsertId()
	if err != nil {
		return model.SessionDetail{}, fmt.Errorf("read interview session id: %w", err)
	}

	createdQuestions := make([]model.InterviewQuestion, 0, len(questions))
	for _, question := range questions {
		expectedPoints, marshalErr := json.Marshal(question.ExpectedPoints)
		if marshalErr != nil {
			return model.SessionDetail{}, fmt.Errorf("marshal expected points: %w", marshalErr)
		}

		createdAt := question.CreatedAt
		if createdAt.IsZero() {
			createdAt = now
		}

		result, insertErr := tx.ExecContext(
			ctx,
			`INSERT INTO interview_questions
			(session_id, question_no, question_type, assessment_dimension, prompt, expected_points, source, is_follow_up, parent_question_id, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			sessionID,
			question.QuestionNo,
			question.QuestionType,
			question.AssessmentDimension,
			question.Prompt,
			expectedPoints,
			question.Source,
			boolToTinyInt(question.IsFollowUp),
			nullInt64FromPtr(question.ParentQuestionID),
			createdAt,
		)
		if insertErr != nil {
			return model.SessionDetail{}, fmt.Errorf("insert interview question: %w", insertErr)
		}

		questionID, idErr := result.LastInsertId()
		if idErr != nil {
			return model.SessionDetail{}, fmt.Errorf("read interview question id: %w", idErr)
		}

		question.ID = questionID
		question.SessionID = sessionID
		question.CreatedAt = createdAt
		createdQuestions = append(createdQuestions, question)
	}

	if err := tx.Commit(); err != nil {
		return model.SessionDetail{}, fmt.Errorf("commit create session tx: %w", err)
	}

	session.ID = sessionID
	session.JobTargetID = jobTargetID
	session.ResumeID = resumeID
	session.CreatedAt = now
	session.UpdatedAt = now

	return model.SessionDetail{
		Session:   session,
		Questions: createdQuestions,
	}, nil
}

func (r *InterviewRepository) GetSession(ctx context.Context, sessionID int64) (model.SessionDetail, error) {
	session, err := r.loadSession(ctx, sessionID)
	if err != nil {
		return model.SessionDetail{}, err
	}

	questions, err := r.loadQuestions(ctx, sessionID)
	if err != nil {
		return model.SessionDetail{}, err
	}

	return model.SessionDetail{
		Session:   session,
		Questions: questions,
	}, nil
}

func (r *InterviewRepository) GetNextQuestion(ctx context.Context, sessionID int64) (model.InterviewQuestion, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.InterviewQuestion{}, fmt.Errorf("begin next question tx: %w", err)
	}
	defer rollback(tx)

	var currentQuestionNo int
	var startedAt sql.NullTime

	row := tx.QueryRowContext(
		ctx,
		`SELECT current_question_no, started_at
		FROM interview_sessions
		WHERE id = ?
		FOR UPDATE`,
		sessionID,
	)
	if err := row.Scan(&currentQuestionNo, &startedAt); err != nil {
		if err == sql.ErrNoRows {
			return model.InterviewQuestion{}, repository.ErrNotFound
		}

		return model.InterviewQuestion{}, fmt.Errorf("load session for next question: %w", err)
	}

	questionNo := currentQuestionNo + 1
	question, err := r.loadQuestionByNumberTx(ctx, tx, sessionID, questionNo)
	if err != nil {
		return model.InterviewQuestion{}, err
	}

	now := time.Now()
	startedAtValue := any(now)
	if startedAt.Valid {
		startedAtValue = startedAt.Time
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE interview_sessions
		SET current_question_no = ?, status = ?, started_at = ?, updated_at = ?
		WHERE id = ?`,
		questionNo,
		"in_progress",
		startedAtValue,
		now,
		sessionID,
	); err != nil {
		return model.InterviewQuestion{}, fmt.Errorf("update session progress: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.InterviewQuestion{}, fmt.Errorf("commit next question tx: %w", err)
	}

	return question, nil
}

func (r *InterviewRepository) SaveAnswer(ctx context.Context, answer model.InterviewAnswer, feedback model.AnswerFeedback) (model.InterviewAnswer, model.AnswerFeedback, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, fmt.Errorf("begin save answer tx: %w", err)
	}
	defer rollback(tx)

	var questionCount int
	row := tx.QueryRowContext(
		ctx,
		`SELECT question_count
		FROM interview_sessions
		WHERE id = ?
		FOR UPDATE`,
		answer.SessionID,
	)
	if err := row.Scan(&questionCount); err != nil {
		if err == sql.ErrNoRows {
			return model.InterviewAnswer{}, model.AnswerFeedback{}, repository.ErrNotFound
		}

		return model.InterviewAnswer{}, model.AnswerFeedback{}, fmt.Errorf("load session for answer: %w", err)
	}

	submittedAt := answer.SubmittedAt
	if submittedAt.IsZero() {
		submittedAt = time.Now()
	}

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO interview_answers
		(session_id, question_id, answer_text, answer_audio_url, answer_duration_seconds, submitted_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		answer.SessionID,
		answer.QuestionID,
		answer.AnswerText,
		nullString(answer.AnswerAudioURL),
		nullInt(answer.AnswerDurationSeconds),
		submittedAt,
	)
	if err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, fmt.Errorf("insert interview answer: %w", err)
	}

	answerID, err := result.LastInsertId()
	if err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, fmt.Errorf("read interview answer id: %w", err)
	}

	answer.ID = answerID
	answer.SubmittedAt = submittedAt

	feedbackCreatedAt := feedback.CreatedAt
	if feedbackCreatedAt.IsZero() {
		feedbackCreatedAt = submittedAt
	}

	feedbackPayload, err := json.Marshal(map[string]any{
		"generatedBy": "rule-based-placeholder",
	})
	if err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, fmt.Errorf("marshal feedback payload: %w", err)
	}

	result, err = tx.ExecContext(
		ctx,
		`INSERT INTO answer_feedbacks
		(answer_id, overall_score, accuracy_score, clarity_score, depth_score, communication_score, strengths, issues, improvement_suggestion, follow_up_question, feedback_payload, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		answerID,
		feedback.OverallScore,
		feedback.AccuracyScore,
		feedback.ClarityScore,
		feedback.DepthScore,
		feedback.CommunicationScore,
		feedback.Strengths,
		feedback.Issues,
		feedback.ImprovementSuggestion,
		feedback.FollowUpQuestion,
		feedbackPayload,
		feedbackCreatedAt,
	)
	if err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, fmt.Errorf("insert answer feedback: %w", err)
	}

	feedbackID, err := result.LastInsertId()
	if err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, fmt.Errorf("read answer feedback id: %w", err)
	}

	feedback.ID = feedbackID
	feedback.AnswerID = answerID
	feedback.CreatedAt = feedbackCreatedAt

	var answerCount int
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM interview_answers
		WHERE session_id = ?`,
		answer.SessionID,
	).Scan(&answerCount); err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, fmt.Errorf("count session answers: %w", err)
	}

	if answerCount >= questionCount {
		if err := r.completeSessionTx(ctx, tx, answer.SessionID); err != nil {
			return model.InterviewAnswer{}, model.AnswerFeedback{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.InterviewAnswer{}, model.AnswerFeedback{}, fmt.Errorf("commit save answer tx: %w", err)
	}

	return answer, feedback, nil
}

func (r *InterviewRepository) ListSessionAnswers(ctx context.Context, sessionID int64) ([]model.SessionAnswerItem, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			iq.id,
			iq.question_no,
			iq.question_type,
			iq.assessment_dimension,
			iq.prompt,
			iq.expected_points,
			ia.id,
			ia.session_id,
			ia.question_id,
			ia.answer_text,
			ia.answer_audio_url,
			ia.answer_duration_seconds,
			ia.submitted_at,
			af.id,
			af.answer_id,
			af.overall_score,
			af.accuracy_score,
			af.clarity_score,
			af.depth_score,
			af.communication_score,
			af.strengths,
			af.issues,
			af.improvement_suggestion,
			af.follow_up_question,
			af.created_at
		FROM interview_answers ia
		INNER JOIN interview_questions iq ON iq.id = ia.question_id
		LEFT JOIN answer_feedbacks af ON af.answer_id = ia.id
		WHERE ia.session_id = ?
		ORDER BY iq.question_no, ia.submitted_at`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list session answers: %w", err)
	}
	defer rows.Close()

	items := make([]model.SessionAnswerItem, 0)
	for rows.Next() {
		item, err := scanSessionAnswerItem(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate session answers: %w", err)
	}

	if len(items) == 0 {
		if _, err := r.loadSession(ctx, sessionID); err != nil {
			return nil, err
		}
	}

	return items, nil
}

func (r *InterviewRepository) GetReport(ctx context.Context, sessionID int64) (model.InterviewReport, error) {
	var (
		report             model.InterviewReport
		nextPlanJSON       sql.NullString
		strengthsSummary   sql.NullString
		weaknessSummary    sql.NullString
		recommendedActions sql.NullString
	)

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, session_id, overall_score, performance_summary, strengths_summary, weakness_summary, recommended_actions, next_plan, created_at
		FROM interview_reports
		WHERE session_id = ?`,
		sessionID,
	)
	if err := row.Scan(
		&report.ID,
		&report.SessionID,
		&report.OverallScore,
		&report.PerformanceSummary,
		&strengthsSummary,
		&weaknessSummary,
		&recommendedActions,
		&nextPlanJSON,
		&report.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.InterviewReport{}, repository.ErrNotFound
		}

		return model.InterviewReport{}, fmt.Errorf("load interview report: %w", err)
	}

	report.StrengthsSummary = strengthsSummary.String
	report.WeaknessSummary = weaknessSummary.String
	report.RecommendedActions = recommendedActions.String
	report.NextPlan = decodeJSONObject(nextPlanJSON.String)

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT dimension_code, score, comment
		FROM report_dimension_scores
		WHERE report_id = ?
		ORDER BY id`,
		report.ID,
	)
	if err != nil {
		return model.InterviewReport{}, fmt.Errorf("load report dimension scores: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			item    model.ReportDimensionScore
			comment sql.NullString
		)

		if err := rows.Scan(&item.DimensionCode, &item.Score, &comment); err != nil {
			return model.InterviewReport{}, fmt.Errorf("scan report dimension score: %w", err)
		}

		item.Comment = comment.String
		report.DimensionScores = append(report.DimensionScores, item)
	}

	if err := rows.Err(); err != nil {
		return model.InterviewReport{}, fmt.Errorf("iterate report dimension scores: %w", err)
	}

	localizeInterviewReport(&report)

	return report, nil
}

func (r *InterviewRepository) SaveReport(ctx context.Context, report model.InterviewReport) (model.InterviewReport, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.InterviewReport{}, fmt.Errorf("begin save report tx: %w", err)
	}
	defer rollback(tx)

	var sessionExists int
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM interview_sessions
		WHERE id = ?`,
		report.SessionID,
	).Scan(&sessionExists); err != nil {
		return model.InterviewReport{}, fmt.Errorf("check interview session for report: %w", err)
	}
	if sessionExists == 0 {
		return model.InterviewReport{}, repository.ErrNotFound
	}

	nextPlanJSON, err := json.Marshal(report.NextPlan)
	if err != nil {
		return model.InterviewReport{}, fmt.Errorf("marshal report next plan: %w", err)
	}

	createdAt := report.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	var reportID int64
	loadErr := tx.QueryRowContext(
		ctx,
		`SELECT id
		FROM interview_reports
		WHERE session_id = ?`,
		report.SessionID,
	).Scan(&reportID)
	if loadErr != nil && loadErr != sql.ErrNoRows {
		return model.InterviewReport{}, fmt.Errorf("load existing interview report: %w", loadErr)
	}

	if loadErr == sql.ErrNoRows {
		result, execErr := tx.ExecContext(
			ctx,
			`INSERT INTO interview_reports
			(session_id, overall_score, performance_summary, strengths_summary, weakness_summary, recommended_actions, next_plan, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			report.SessionID,
			report.OverallScore,
			report.PerformanceSummary,
			nullString(report.StrengthsSummary),
			nullString(report.WeaknessSummary),
			nullString(report.RecommendedActions),
			nextPlanJSON,
			createdAt,
		)
		if execErr != nil {
			return model.InterviewReport{}, fmt.Errorf("insert interview report: %w", execErr)
		}

		reportID, err = result.LastInsertId()
		if err != nil {
			return model.InterviewReport{}, fmt.Errorf("read interview report id: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE interview_reports
			SET overall_score = ?, performance_summary = ?, strengths_summary = ?, weakness_summary = ?, recommended_actions = ?, next_plan = ?
			WHERE id = ?`,
			report.OverallScore,
			report.PerformanceSummary,
			nullString(report.StrengthsSummary),
			nullString(report.WeaknessSummary),
			nullString(report.RecommendedActions),
			nextPlanJSON,
			reportID,
		); err != nil {
			return model.InterviewReport{}, fmt.Errorf("update interview report: %w", err)
		}

		if _, err := tx.ExecContext(
			ctx,
			`DELETE FROM report_dimension_scores
			WHERE report_id = ?`,
			reportID,
		); err != nil {
			return model.InterviewReport{}, fmt.Errorf("delete report dimension scores: %w", err)
		}
	}

	for _, item := range report.DimensionScores {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO report_dimension_scores
			(report_id, dimension_code, score, comment)
			VALUES (?, ?, ?, ?)`,
			reportID,
			item.DimensionCode,
			item.Score,
			nullString(item.Comment),
		); err != nil {
			return model.InterviewReport{}, fmt.Errorf("insert report dimension score: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return model.InterviewReport{}, fmt.Errorf("commit save report tx: %w", err)
	}

	report.ID = reportID
	report.CreatedAt = createdAt
	return report, nil
}

func (r *InterviewRepository) ListHistory(ctx context.Context, userID int64) ([]model.HistoryItem, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT s.id, s.session_name, s.job_title, s.round_type, s.status, r.overall_score, s.created_at, s.ended_at
		FROM interview_sessions s
		LEFT JOIN interview_reports r ON r.session_id = s.id
		WHERE s.user_id = ?
		ORDER BY s.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}
	defer rows.Close()

	items := make([]model.HistoryItem, 0)
	for rows.Next() {
		var (
			item         model.HistoryItem
			overallScore sql.NullFloat64
			completedAt  sql.NullTime
		)

		if err := rows.Scan(
			&item.SessionID,
			&item.SessionName,
			&item.JobTitle,
			&item.RoundType,
			&item.Status,
			&overallScore,
			&item.CreatedAt,
			&completedAt,
		); err != nil {
			return nil, fmt.Errorf("scan history item: %w", err)
		}

		if overallScore.Valid {
			score := overallScore.Float64
			item.OverallScore = &score
		}
		if completedAt.Valid {
			item.CompletedAt = &completedAt.Time
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history items: %w", err)
	}

	return items, nil
}

func (r *InterviewRepository) GetUserAIConfig(ctx context.Context, userID int64) (model.UserAIProviderConfig, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, provider_code, api_key_encrypted, api_key_masked, model, base_url, is_enabled, last_test_ok, last_tested_at, created_at, updated_at
		FROM user_ai_provider_configs
		WHERE user_id = ?
		LIMIT 1`,
		userID,
	)

	var (
		cfg          model.UserAIProviderConfig
		lastTestedAt sql.NullTime
		isEnabled    int8
		lastTestOK   int8
	)
	if err := row.Scan(
		&cfg.ID,
		&cfg.UserID,
		&cfg.ProviderCode,
		&cfg.APIKey,
		&cfg.APIKeyMasked,
		&cfg.Model,
		&cfg.BaseURL,
		&isEnabled,
		&lastTestOK,
		&lastTestedAt,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.UserAIProviderConfig{}, repository.ErrNotFound
		}

		return model.UserAIProviderConfig{}, fmt.Errorf("load user ai config: %w", err)
	}

	cfg.IsEnabled = isEnabled == 1
	cfg.LastTestOK = lastTestOK == 1
	if lastTestedAt.Valid {
		cfg.LastTestedAt = &lastTestedAt.Time
	}

	return cfg, nil
}

func (r *InterviewRepository) UpsertUserAIConfig(ctx context.Context, cfg model.UserAIProviderConfig) (model.UserAIProviderConfig, error) {
	now := time.Now()
	var existingID int64
	loadErr := r.db.QueryRowContext(
		ctx,
		`SELECT id
		FROM user_ai_provider_configs
		WHERE user_id = ? AND provider_code = ?
		LIMIT 1`,
		cfg.UserID,
		cfg.ProviderCode,
	).Scan(&existingID)
	if loadErr != nil && loadErr != sql.ErrNoRows {
		return model.UserAIProviderConfig{}, fmt.Errorf("load existing user ai config: %w", loadErr)
	}

	if loadErr == sql.ErrNoRows {
		result, err := r.db.ExecContext(
			ctx,
			`INSERT INTO user_ai_provider_configs
			(user_id, provider_code, api_key_encrypted, api_key_masked, model, base_url, is_enabled, last_test_ok, last_tested_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			cfg.UserID,
			cfg.ProviderCode,
			cfg.APIKey,
			cfg.APIKeyMasked,
			cfg.Model,
			cfg.BaseURL,
			boolToTinyInt(cfg.IsEnabled),
			boolToTinyInt(cfg.LastTestOK),
			cfg.LastTestedAt,
			now,
			now,
		)
		if err != nil {
			return model.UserAIProviderConfig{}, fmt.Errorf("insert user ai config: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return model.UserAIProviderConfig{}, fmt.Errorf("read user ai config id: %w", err)
		}

		cfg.ID = id
		cfg.CreatedAt = now
		cfg.UpdatedAt = now
		return cfg, nil
	}

	if _, err := r.db.ExecContext(
		ctx,
		`UPDATE user_ai_provider_configs
		SET api_key_encrypted = ?, api_key_masked = ?, model = ?, base_url = ?, is_enabled = ?, last_test_ok = ?, last_tested_at = ?, updated_at = ?
		WHERE id = ?`,
		cfg.APIKey,
		cfg.APIKeyMasked,
		cfg.Model,
		cfg.BaseURL,
		boolToTinyInt(cfg.IsEnabled),
		boolToTinyInt(cfg.LastTestOK),
		cfg.LastTestedAt,
		now,
		existingID,
	); err != nil {
		return model.UserAIProviderConfig{}, fmt.Errorf("update user ai config: %w", err)
	}

	cfg.ID = existingID
	cfg.UpdatedAt = now
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = now
	}
	return cfg, nil
}

func (r *InterviewRepository) ListAdminUsers(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminUserListItem], error) {
	whereClause, args := buildAdminUserFilter(query)
	total, err := r.countRows(ctx, "SELECT COUNT(*) FROM users u"+whereClause, args...)
	if err != nil {
		return model.PagedResult[model.AdminUserListItem]{}, fmt.Errorf("count admin users: %w", err)
	}

	offset := (query.Page - 1) * query.PageSize
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT u.id, u.email, u.nickname, u.status, u.created_at, u.updated_at
		FROM users u`+whereClause+`
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?`,
		append(args, query.PageSize, offset)...,
	)
	if err != nil {
		return model.PagedResult[model.AdminUserListItem]{}, fmt.Errorf("list admin users: %w", err)
	}
	defer rows.Close()

	items := make([]model.AdminUserListItem, 0)
	for rows.Next() {
		var item model.AdminUserListItem
		if err := rows.Scan(&item.ID, &item.Email, &item.Nickname, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return model.PagedResult[model.AdminUserListItem]{}, fmt.Errorf("scan admin user: %w", err)
		}
		items = append(items, item)
	}

	return model.PagedResult[model.AdminUserListItem]{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, rows.Err()
}

func (r *InterviewRepository) GetAdminUserDetail(ctx context.Context, userID int64) (model.AdminUserDetail, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return model.AdminUserDetail{}, err
	}

	sessionCount, err := r.countRows(ctx, "SELECT COUNT(*) FROM interview_sessions WHERE user_id = ?", userID)
	if err != nil {
		return model.AdminUserDetail{}, fmt.Errorf("count user sessions: %w", err)
	}

	resumeCount, err := r.countRows(ctx, "SELECT COUNT(*) FROM resumes WHERE user_id = ?", userID)
	if err != nil {
		return model.AdminUserDetail{}, fmt.Errorf("count user resumes: %w", err)
	}

	reportCount, err := r.countRows(ctx, `SELECT COUNT(*)
		FROM interview_reports ir
		INNER JOIN interview_sessions s ON s.id = ir.session_id
		WHERE s.user_id = ?`, userID)
	if err != nil {
		return model.AdminUserDetail{}, fmt.Errorf("count user reports: %w", err)
	}

	aiConfigCount, err := r.countRows(ctx, "SELECT COUNT(*) FROM user_ai_provider_configs WHERE user_id = ?", userID)
	if err != nil {
		return model.AdminUserDetail{}, fmt.Errorf("count user ai configs: %w", err)
	}

	return model.AdminUserDetail{
		User:          user,
		SessionCount:  sessionCount,
		ResumeCount:   resumeCount,
		ReportCount:   reportCount,
		AIConfigCount: aiConfigCount,
	}, nil
}

func (r *InterviewRepository) ListAdminInterviewSessions(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminInterviewSessionListItem], error) {
	whereClause, args := buildAdminSessionFilter(query)
	total, err := r.countRows(
		ctx,
		`SELECT COUNT(*)
		FROM interview_sessions s
		INNER JOIN users u ON u.id = s.user_id
		LEFT JOIN job_targets jt ON jt.id = s.job_target_id`+whereClause,
		args...,
	)
	if err != nil {
		return model.PagedResult[model.AdminInterviewSessionListItem]{}, fmt.Errorf("count admin sessions: %w", err)
	}

	offset := (query.Page - 1) * query.PageSize
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			s.id,
			s.user_id,
			u.email,
			u.nickname,
			s.session_name,
			s.job_title,
			COALESCE(jt.company_name, ''),
			s.round_type,
			s.status,
			r.overall_score,
			s.created_at,
			s.ended_at
		FROM interview_sessions s
		INNER JOIN users u ON u.id = s.user_id
		LEFT JOIN job_targets jt ON jt.id = s.job_target_id
		LEFT JOIN interview_reports r ON r.session_id = s.id`+whereClause+`
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?`,
		append(args, query.PageSize, offset)...,
	)
	if err != nil {
		return model.PagedResult[model.AdminInterviewSessionListItem]{}, fmt.Errorf("list admin sessions: %w", err)
	}
	defer rows.Close()

	items := make([]model.AdminInterviewSessionListItem, 0)
	for rows.Next() {
		var (
			item         model.AdminInterviewSessionListItem
			overallScore sql.NullFloat64
			completedAt  sql.NullTime
		)

		if err := rows.Scan(
			&item.SessionID,
			&item.UserID,
			&item.UserEmail,
			&item.UserNickname,
			&item.SessionName,
			&item.JobTitle,
			&item.CompanyName,
			&item.RoundType,
			&item.Status,
			&overallScore,
			&item.CreatedAt,
			&completedAt,
		); err != nil {
			return model.PagedResult[model.AdminInterviewSessionListItem]{}, fmt.Errorf("scan admin session: %w", err)
		}

		if overallScore.Valid {
			score := overallScore.Float64
			item.OverallScore = &score
		}
		if completedAt.Valid {
			item.CompletedAt = &completedAt.Time
		}
		items = append(items, item)
	}

	return model.PagedResult[model.AdminInterviewSessionListItem]{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, rows.Err()
}

func (r *InterviewRepository) ListAdminResumes(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminResumeListItem], error) {
	whereClause, args := buildAdminResumeFilter(query)
	total, err := r.countRows(
		ctx,
		`SELECT COUNT(*)
		FROM resumes rs
		INNER JOIN users u ON u.id = rs.user_id`+whereClause,
		args...,
	)
	if err != nil {
		return model.PagedResult[model.AdminResumeListItem]{}, fmt.Errorf("count admin resumes: %w", err)
	}

	offset := (query.Page - 1) * query.PageSize
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			rs.id,
			rs.user_id,
			u.email,
			u.nickname,
			rs.name,
			rs.source_type,
			rs.is_default,
			rs.updated_at
		FROM resumes rs
		INNER JOIN users u ON u.id = rs.user_id`+whereClause+`
		ORDER BY rs.updated_at DESC
		LIMIT ? OFFSET ?`,
		append(args, query.PageSize, offset)...,
	)
	if err != nil {
		return model.PagedResult[model.AdminResumeListItem]{}, fmt.Errorf("list admin resumes: %w", err)
	}
	defer rows.Close()

	items := make([]model.AdminResumeListItem, 0)
	for rows.Next() {
		var (
			item      model.AdminResumeListItem
			isDefault int8
		)
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.UserEmail,
			&item.UserNickname,
			&item.Name,
			&item.SourceType,
			&isDefault,
			&item.UpdatedAt,
		); err != nil {
			return model.PagedResult[model.AdminResumeListItem]{}, fmt.Errorf("scan admin resume: %w", err)
		}
		item.IsDefault = isDefault == 1
		items = append(items, item)
	}

	return model.PagedResult[model.AdminResumeListItem]{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, rows.Err()
}

func (r *InterviewRepository) GetAdminResumeDetail(ctx context.Context, resumeID int64) (model.AdminResumeDetail, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT
			rs.id,
			u.id,
			u.email,
			u.nickname,
			u.status,
			rs.name,
			rs.source_type,
			COALESCE(rs.file_url, ''),
			COALESCE(rs.raw_text, ''),
			COALESCE(CAST(rs.parsed_content AS CHAR), ''),
			rs.version_no,
			rs.is_default,
			rs.created_at,
			rs.updated_at
		FROM resumes rs
		INNER JOIN users u ON u.id = rs.user_id
		WHERE rs.id = ?
		LIMIT 1`,
		resumeID,
	)

	var (
		detail        model.AdminResumeDetail
		user          model.AdminUserSummary
		parsedContent sql.NullString
		isDefault     int8
	)
	if err := row.Scan(
		&detail.ID,
		&user.ID,
		&user.Email,
		&user.Nickname,
		&user.Status,
		&detail.Name,
		&detail.SourceType,
		&detail.FileURL,
		&detail.RawText,
		&parsedContent,
		&detail.VersionNo,
		&isDefault,
		&detail.CreatedAt,
		&detail.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.AdminResumeDetail{}, repository.ErrNotFound
		}
		return model.AdminResumeDetail{}, fmt.Errorf("load admin resume detail: %w", err)
	}

	detail.User = user
	detail.IsDefault = isDefault == 1
	detail.ParsedContent = decodeJSONObject(parsedContent.String)
	return detail, nil
}

func (r *InterviewRepository) ListAdminAIConfigs(ctx context.Context, query model.AdminListQuery) (model.PagedResult[model.AdminAIConfigListItem], error) {
	whereClause, args := buildAdminAIConfigFilter(query)
	total, err := r.countRows(
		ctx,
		`SELECT COUNT(*)
		FROM user_ai_provider_configs cfg
		INNER JOIN users u ON u.id = cfg.user_id`+whereClause,
		args...,
	)
	if err != nil {
		return model.PagedResult[model.AdminAIConfigListItem]{}, fmt.Errorf("count admin ai configs: %w", err)
	}

	offset := (query.Page - 1) * query.PageSize
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			cfg.id,
			cfg.user_id,
			u.email,
			u.nickname,
			cfg.provider_code,
			cfg.api_key_masked,
			cfg.model,
			cfg.base_url,
			cfg.is_enabled,
			cfg.last_test_ok,
			cfg.last_tested_at,
			cfg.updated_at
		FROM user_ai_provider_configs cfg
		INNER JOIN users u ON u.id = cfg.user_id`+whereClause+`
		ORDER BY cfg.updated_at DESC
		LIMIT ? OFFSET ?`,
		append(args, query.PageSize, offset)...,
	)
	if err != nil {
		return model.PagedResult[model.AdminAIConfigListItem]{}, fmt.Errorf("list admin ai configs: %w", err)
	}
	defer rows.Close()

	items := make([]model.AdminAIConfigListItem, 0)
	for rows.Next() {
		var (
			item         model.AdminAIConfigListItem
			isEnabled    int8
			lastTestOK   int8
			lastTestedAt sql.NullTime
		)
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.UserEmail,
			&item.UserNickname,
			&item.ProviderCode,
			&item.APIKeyMasked,
			&item.Model,
			&item.BaseURL,
			&isEnabled,
			&lastTestOK,
			&lastTestedAt,
			&item.UpdatedAt,
		); err != nil {
			return model.PagedResult[model.AdminAIConfigListItem]{}, fmt.Errorf("scan admin ai config: %w", err)
		}
		item.IsEnabled = isEnabled == 1
		item.LastTestOK = lastTestOK == 1
		if lastTestedAt.Valid {
			item.LastTestedAt = &lastTestedAt.Time
		}
		items = append(items, item)
	}

	return model.PagedResult[model.AdminAIConfigListItem]{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, rows.Err()
}

func (r *InterviewRepository) GetAdminOverview(ctx context.Context) (model.AdminOverview, error) {
	adminCount, err := r.countRows(ctx, "SELECT COUNT(*) FROM admins")
	if err != nil {
		return model.AdminOverview{}, fmt.Errorf("count admins: %w", err)
	}
	userCount, err := r.countRows(ctx, "SELECT COUNT(*) FROM users")
	if err != nil {
		return model.AdminOverview{}, fmt.Errorf("count users: %w", err)
	}
	sessionCount, err := r.countRows(ctx, "SELECT COUNT(*) FROM interview_sessions")
	if err != nil {
		return model.AdminOverview{}, fmt.Errorf("count sessions: %w", err)
	}
	resumeCount, err := r.countRows(ctx, "SELECT COUNT(*) FROM resumes")
	if err != nil {
		return model.AdminOverview{}, fmt.Errorf("count resumes: %w", err)
	}
	reportCount, err := r.countRows(ctx, "SELECT COUNT(*) FROM interview_reports")
	if err != nil {
		return model.AdminOverview{}, fmt.Errorf("count reports: %w", err)
	}
	aiConfigCount, err := r.countRows(ctx, "SELECT COUNT(*) FROM user_ai_provider_configs")
	if err != nil {
		return model.AdminOverview{}, fmt.Errorf("count ai configs: %w", err)
	}

	return model.AdminOverview{
		AdminCount:    adminCount,
		UserCount:     userCount,
		SessionCount:  sessionCount,
		ResumeCount:   resumeCount,
		ReportCount:   reportCount,
		AIConfigCount: aiConfigCount,
	}, nil
}

func (r *InterviewRepository) loadSession(ctx context.Context, sessionID int64) (model.InterviewSession, error) {
	var (
		session            model.InterviewSession
		candidateProfileID sql.NullInt64
		jobTargetID        sql.NullInt64
		resumeID           sql.NullInt64
		companyName        sql.NullString
		resumeFileName     sql.NullString
		resumeText         sql.NullString
		startedAt          sql.NullTime
		endedAt            sql.NullTime
	)

	row := r.db.QueryRowContext(
		ctx,
		`SELECT
			s.id,
			s.user_id,
			s.candidate_profile_id,
			s.job_target_id,
			s.resume_id,
			s.job_title,
			COALESCE(jt.job_category, ''),
			COALESCE(jt.level_code, ''),
			COALESCE(jt.interview_type, ''),
			COALESCE(jt.company_name, ''),
			COALESCE(r.name, ''),
			COALESCE(r.raw_text, ''),
			s.session_name,
			s.round_type,
			s.interview_mode,
			s.interviewer_style,
			s.difficulty_level,
			s.question_count,
			s.current_question_no,
			s.status,
			s.started_at,
			s.ended_at,
			s.created_at,
			s.updated_at
		FROM interview_sessions s
		LEFT JOIN job_targets jt ON jt.id = s.job_target_id
		LEFT JOIN resumes r ON r.id = s.resume_id
		WHERE s.id = ?`,
		sessionID,
	)
	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&candidateProfileID,
		&jobTargetID,
		&resumeID,
		&session.JobTitle,
		&session.JobCategory,
		&session.LevelCode,
		&session.InterviewType,
		&companyName,
		&resumeFileName,
		&resumeText,
		&session.SessionName,
		&session.RoundType,
		&session.InterviewMode,
		&session.InterviewerStyle,
		&session.DifficultyLevel,
		&session.QuestionCount,
		&session.CurrentQuestionNo,
		&session.Status,
		&startedAt,
		&endedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.InterviewSession{}, repository.ErrNotFound
		}

		return model.InterviewSession{}, fmt.Errorf("load interview session: %w", err)
	}

	if candidateProfileID.Valid {
		session.CandidateProfileID = candidateProfileID.Int64
	}
	if jobTargetID.Valid {
		session.JobTargetID = jobTargetID.Int64
	}
	if resumeID.Valid {
		session.ResumeID = resumeID.Int64
	}
	session.CompanyName = companyName.String
	session.ResumeFileName = resumeFileName.String
	session.ResumeText = resumeText.String
	if startedAt.Valid {
		session.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}

	return session, nil
}

func (r *InterviewRepository) upsertResumeTx(ctx context.Context, tx *sql.Tx, session model.InterviewSession, now time.Time) (int64, error) {
	if session.UserID == 0 || strings.TrimSpace(session.ResumeText) == "" {
		return 0, nil
	}

	resumeName := strings.TrimSpace(session.ResumeFileName)
	if resumeName == "" {
		resumeName = session.JobTitle + " 简历"
	}

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO resumes
		(user_id, name, source_type, raw_text, is_default, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		session.UserID,
		resumeName,
		"manual",
		session.ResumeText,
		0,
		now,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("insert resume: %w", err)
	}

	resumeID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read resume id: %w", err)
	}

	return resumeID, nil
}

func (r *InterviewRepository) createJobTargetTx(ctx context.Context, tx *sql.Tx, session model.InterviewSession, resumeID int64, now time.Time) (int64, error) {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO job_targets
		(user_id, resume_id, company_name, job_title, job_category, level_code, interview_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.UserID,
		nullInt64(resumeID),
		nullString(session.CompanyName),
		session.JobTitle,
		defaultString(session.JobCategory, "backend"),
		defaultString(session.LevelCode, "mid"),
		defaultString(session.InterviewType, "technical_1"),
		now,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("insert job target: %w", err)
	}

	jobTargetID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read job target id: %w", err)
	}

	return jobTargetID, nil
}

func (r *InterviewRepository) loadQuestions(ctx context.Context, sessionID int64) ([]model.InterviewQuestion, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, session_id, question_no, question_type, assessment_dimension, prompt, expected_points, source, is_follow_up, parent_question_id, created_at
		FROM interview_questions
		WHERE session_id = ?
		ORDER BY question_no`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("load interview questions: %w", err)
	}
	defer rows.Close()

	questions := make([]model.InterviewQuestion, 0)
	for rows.Next() {
		question, err := scanQuestion(rows)
		if err != nil {
			return nil, err
		}

		questions = append(questions, question)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate interview questions: %w", err)
	}

	return questions, nil
}

func (r *InterviewRepository) loadQuestionByNumberTx(ctx context.Context, tx *sql.Tx, sessionID int64, questionNo int) (model.InterviewQuestion, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT id, session_id, question_no, question_type, assessment_dimension, prompt, expected_points, source, is_follow_up, parent_question_id, created_at
		FROM interview_questions
		WHERE session_id = ? AND question_no = ?`,
		sessionID,
		questionNo,
	)

	return scanQuestion(row)
}

func (r *InterviewRepository) completeSessionTx(ctx context.Context, tx *sql.Tx, sessionID int64) error {
	now := time.Now()
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE interview_sessions
		SET status = ?, ended_at = ?, updated_at = ?
		WHERE id = ?`,
		"completed",
		now,
		now,
		sessionID,
	); err != nil {
		return fmt.Errorf("complete interview session: %w", err)
	}

	var reportCount int
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM interview_reports
		WHERE session_id = ?`,
		sessionID,
	).Scan(&reportCount); err != nil {
		return fmt.Errorf("count interview reports: %w", err)
	}

	if reportCount > 0 {
		return nil
	}

	var (
		overallScore       float64
		accuracyScore      float64
		clarityScore       float64
		depthScore         float64
		communicationScore float64
	)
	if err := tx.QueryRowContext(
		ctx,
		`SELECT
			COALESCE(AVG(overall_score), 0),
			COALESCE(AVG(accuracy_score), 0),
			COALESCE(AVG(clarity_score), 0),
			COALESCE(AVG(depth_score), 0),
			COALESCE(AVG(communication_score), 0)
		FROM answer_feedbacks af
		INNER JOIN interview_answers ia ON ia.id = af.answer_id
		WHERE ia.session_id = ?`,
		sessionID,
	).Scan(&overallScore, &accuracyScore, &clarityScore, &depthScore, &communicationScore); err != nil {
		return fmt.Errorf("aggregate answer scores: %w", err)
	}

	nextPlan, err := json.Marshal(map[string]any{
		"focusAreas":               []string{"high_concurrency", "cache_consistency", "project_depth"},
		"recommendedQuestionCount": 6,
	})
	if err != nil {
		return fmt.Errorf("marshal next plan: %w", err)
	}

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO interview_reports
		(session_id, overall_score, performance_summary, strengths_summary, weakness_summary, recommended_actions, next_plan, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID,
		overallScore,
		"整体回答覆盖面比较完整，主问题链路清楚，但实现细节还可以再具体一些。",
		"回答结构比较稳定，能够清晰覆盖核心问题链路。",
		"在缓存一致性、幂等控制和故障恢复等关键细节上，深度还不够充分。",
		"下一轮建议重点训练高并发设计、缓存一致性、重试补偿，以及项目表达的完整度。",
		nextPlan,
		now,
	)
	if err != nil {
		return fmt.Errorf("insert interview report: %w", err)
	}

	reportID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read interview report id: %w", err)
	}

	dimensionItems := []model.ReportDimensionScore{
		{DimensionCode: "accuracy", Score: accuracyScore, Comment: "技术准确性整体稳定。"},
		{DimensionCode: "clarity", Score: clarityScore, Comment: "回答结构清晰，表达比较顺畅。"},
		{DimensionCode: "depth", Score: depthScore, Comment: "技术深度还有提升空间。"},
		{DimensionCode: "communication", Score: communicationScore, Comment: "沟通表达自然，节奏较好。"},
	}

	for _, item := range dimensionItems {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO report_dimension_scores
			(report_id, dimension_code, score, comment)
			VALUES (?, ?, ?, ?)`,
			reportID,
			item.DimensionCode,
			item.Score,
			item.Comment,
		); err != nil {
			return fmt.Errorf("insert report dimension score: %w", err)
		}
	}

	return nil
}

func localizeInterviewReport(report *model.InterviewReport) {
	report.PerformanceSummary = localizeReportText(report.PerformanceSummary)
	report.StrengthsSummary = localizeReportText(report.StrengthsSummary)
	report.WeaknessSummary = localizeReportText(report.WeaknessSummary)
	report.RecommendedActions = localizeReportText(report.RecommendedActions)
	report.NextPlan = localizeNextPlan(report.NextPlan)

	for index := range report.DimensionScores {
		item := &report.DimensionScores[index]
		item.Comment = localizeDimensionComment(item.DimensionCode, item.Comment)
	}
}

func localizeReportText(text string) string {
	switch text {
	case "Overall coverage is solid, but implementation details can still be more concrete.":
		return "整体回答覆盖面比较完整，主问题链路清楚，但实现细节还可以再具体一些。"
	case "Answer structure is stable and the main problem path is covered clearly.":
		return "回答结构比较稳定，能够清晰覆盖核心问题链路。"
	case "Depth around cache consistency, idempotency, and failure recovery is still light.":
		return "在缓存一致性、幂等控制和故障恢复等关键细节上，深度还不够充分。"
	case "Next round should focus on concurrency design, cache consistency, retries, and project storytelling.":
		return "下一轮建议重点训练高并发设计、缓存一致性、重试补偿，以及项目表达的完整度。"
	default:
		return text
	}
}

func localizeDimensionComment(dimensionCode string, comment string) string {
	switch comment {
	case "Technical accuracy is stable.":
		return "技术准确性整体稳定。"
	case "Answer clarity is good.":
		return "回答结构清晰，表达比较顺畅。"
	case "Depth still has room to improve.":
		return "技术深度还有提升空间。"
	case "Communication feels natural.":
		return "沟通表达自然，节奏较好。"
	}

	switch dimensionCode {
	case "accuracy":
		if comment == "" {
			return "技术准确性整体稳定。"
		}
	case "clarity":
		if comment == "" {
			return "回答结构清晰，表达比较顺畅。"
		}
	case "depth":
		if comment == "" {
			return "技术深度还有提升空间。"
		}
	case "communication":
		if comment == "" {
			return "沟通表达自然，节奏较好。"
		}
	}

	return comment
}

func localizeNextPlan(plan map[string]any) map[string]any {
	if len(plan) == 0 {
		return plan
	}

	rawAreas, ok := plan["focusAreas"]
	if !ok {
		return plan
	}

	areas, ok := rawAreas.([]any)
	if !ok {
		if items, castOK := rawAreas.([]string); castOK {
			converted := make([]any, 0, len(items))
			for _, item := range items {
				converted = append(converted, localizeFocusArea(item))
			}
			plan["focusAreas"] = converted
		}
		return plan
	}

	localized := make([]any, 0, len(areas))
	for _, item := range areas {
		text, _ := item.(string)
		localized = append(localized, localizeFocusArea(text))
	}
	plan["focusAreas"] = localized
	return plan
}

func localizeFocusArea(value string) string {
	switch value {
	case "high_concurrency":
		return "高并发设计"
	case "cache_consistency":
		return "缓存一致性"
	case "project_depth":
		return "项目深度表达"
	default:
		return value
	}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanQuestion(s scanner) (model.InterviewQuestion, error) {
	var (
		question         model.InterviewQuestion
		expectedPoints   sql.NullString
		isFollowUp       int8
		parentQuestionID sql.NullInt64
	)

	if err := s.Scan(
		&question.ID,
		&question.SessionID,
		&question.QuestionNo,
		&question.QuestionType,
		&question.AssessmentDimension,
		&question.Prompt,
		&expectedPoints,
		&question.Source,
		&isFollowUp,
		&parentQuestionID,
		&question.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.InterviewQuestion{}, repository.ErrNotFound
		}

		return model.InterviewQuestion{}, fmt.Errorf("scan interview question: %w", err)
	}

	if expectedPoints.Valid && expectedPoints.String != "" {
		if err := json.Unmarshal([]byte(expectedPoints.String), &question.ExpectedPoints); err != nil {
			return model.InterviewQuestion{}, fmt.Errorf("unmarshal expected points: %w", err)
		}
	}
	question.IsFollowUp = isFollowUp == 1

	if parentQuestionID.Valid {
		question.ParentQuestionID = &parentQuestionID.Int64
	}

	return question, nil
}

func scanSessionAnswerItem(s scanner) (model.SessionAnswerItem, error) {
	var (
		item                  model.SessionAnswerItem
		expectedPoints        sql.NullString
		answerAudioURL        sql.NullString
		answerDurationSeconds sql.NullInt64
		feedbackID            sql.NullInt64
		feedbackAnswerID      sql.NullInt64
		overallScore          sql.NullFloat64
		accuracyScore         sql.NullFloat64
		clarityScore          sql.NullFloat64
		depthScore            sql.NullFloat64
		communicationScore    sql.NullFloat64
		strengths             sql.NullString
		issues                sql.NullString
		improvementSuggestion sql.NullString
		followUpQuestion      sql.NullString
		feedbackCreatedAt     sql.NullTime
	)

	if err := s.Scan(
		&item.QuestionID,
		&item.QuestionNo,
		&item.QuestionType,
		&item.AssessmentDimension,
		&item.Prompt,
		&expectedPoints,
		&item.Answer.ID,
		&item.Answer.SessionID,
		&item.Answer.QuestionID,
		&item.Answer.AnswerText,
		&answerAudioURL,
		&answerDurationSeconds,
		&item.Answer.SubmittedAt,
		&feedbackID,
		&feedbackAnswerID,
		&overallScore,
		&accuracyScore,
		&clarityScore,
		&depthScore,
		&communicationScore,
		&strengths,
		&issues,
		&improvementSuggestion,
		&followUpQuestion,
		&feedbackCreatedAt,
	); err != nil {
		return model.SessionAnswerItem{}, fmt.Errorf("scan session answer item: %w", err)
	}

	if expectedPoints.Valid && expectedPoints.String != "" {
		if err := json.Unmarshal([]byte(expectedPoints.String), &item.ExpectedPoints); err != nil {
			return model.SessionAnswerItem{}, fmt.Errorf("unmarshal session answer expected points: %w", err)
		}
	}

	if answerAudioURL.Valid {
		item.Answer.AnswerAudioURL = answerAudioURL.String
	}
	if answerDurationSeconds.Valid {
		item.Answer.AnswerDurationSeconds = int(answerDurationSeconds.Int64)
	}

	if feedbackID.Valid {
		feedback := &model.AnswerFeedback{
			ID:                    feedbackID.Int64,
			AnswerID:              feedbackAnswerID.Int64,
			OverallScore:          overallScore.Float64,
			AccuracyScore:         accuracyScore.Float64,
			ClarityScore:          clarityScore.Float64,
			DepthScore:            depthScore.Float64,
			CommunicationScore:    communicationScore.Float64,
			Strengths:             strengths.String,
			Issues:                issues.String,
			ImprovementSuggestion: improvementSuggestion.String,
			FollowUpQuestion:      followUpQuestion.String,
		}
		if feedbackCreatedAt.Valid {
			feedback.CreatedAt = feedbackCreatedAt.Time
		}
		item.Feedback = feedback
	}

	return item, nil
}

func decodeJSONObject(raw string) map[string]any {
	if raw == "" {
		return map[string]any{}
	}

	decoded := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return map[string]any{}
	}

	return decoded
}

func nullInt64(value int64) any {
	if value == 0 {
		return nil
	}

	return value
}

func nullInt64FromPtr(value *int64) any {
	if value == nil || *value == 0 {
		return nil
	}

	return *value
}

func nullInt(value int) any {
	if value == 0 {
		return nil
	}

	return value
}

func nullString(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func boolToTinyInt(value bool) int {
	if value {
		return 1
	}

	return 0
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return strings.TrimSpace(value)
}

func (r *InterviewRepository) loadAdmin(ctx context.Context, where string, value any) (model.Admin, error) {
	row := r.db.QueryRowContext(
		ctx,
		fmt.Sprintf(
			`SELECT id, email, password_hash, nickname, status, created_at, updated_at
			FROM admins
			WHERE %s
			LIMIT 1`,
			where,
		),
		value,
	)

	var admin model.Admin
	if err := row.Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.Nickname,
		&admin.Status,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.Admin{}, repository.ErrNotFound
		}

		return model.Admin{}, fmt.Errorf("load admin: %w", err)
	}

	return admin, nil
}

func (r *InterviewRepository) loadUser(ctx context.Context, where string, value any) (model.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		fmt.Sprintf(
			`SELECT id, email, password_hash, nickname, avatar_url, status, created_at, updated_at
			FROM users
			WHERE %s
			LIMIT 1`,
			where,
		),
		value,
	)

	var (
		user      model.User
		avatarURL sql.NullString
	)
	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Nickname,
		&avatarURL,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.User{}, repository.ErrNotFound
		}

		return model.User{}, fmt.Errorf("load user: %w", err)
	}

	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}

	return user, nil
}

func (r *InterviewRepository) countRows(ctx context.Context, query string, args ...any) (int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func buildAdminUserFilter(query model.AdminListQuery) (string, []any) {
	clauses := make([]string, 0)
	args := make([]any, 0)
	if query.Keyword != "" {
		clauses = append(clauses, "(u.email LIKE ? OR u.nickname LIKE ?)")
		like := "%" + query.Keyword + "%"
		args = append(args, like, like)
	}
	if query.Status != "" {
		clauses = append(clauses, "u.status = ?")
		args = append(args, query.Status)
	}
	return buildWhereClause(clauses), args
}

func buildAdminSessionFilter(query model.AdminListQuery) (string, []any) {
	clauses := make([]string, 0)
	args := make([]any, 0)
	if query.Keyword != "" {
		clauses = append(clauses, "(u.email LIKE ? OR u.nickname LIKE ? OR s.session_name LIKE ? OR s.job_title LIKE ? OR COALESCE(jt.company_name, '') LIKE ?)")
		like := "%" + query.Keyword + "%"
		args = append(args, like, like, like, like, like)
	}
	if query.Status != "" {
		clauses = append(clauses, "s.status = ?")
		args = append(args, query.Status)
	}
	return buildWhereClause(clauses), args
}

func buildAdminResumeFilter(query model.AdminListQuery) (string, []any) {
	clauses := make([]string, 0)
	args := make([]any, 0)
	if query.Keyword != "" {
		clauses = append(clauses, "(u.email LIKE ? OR u.nickname LIKE ? OR rs.name LIKE ?)")
		like := "%" + query.Keyword + "%"
		args = append(args, like, like, like)
	}
	return buildWhereClause(clauses), args
}

func buildAdminAIConfigFilter(query model.AdminListQuery) (string, []any) {
	clauses := make([]string, 0)
	args := make([]any, 0)
	if query.Keyword != "" {
		clauses = append(clauses, "(u.email LIKE ? OR u.nickname LIKE ? OR cfg.provider_code LIKE ? OR cfg.model LIKE ?)")
		like := "%" + query.Keyword + "%"
		args = append(args, like, like, like, like)
	}
	if query.Status != "" {
		switch query.Status {
		case "active":
			clauses = append(clauses, "cfg.is_enabled = 1")
		case "disabled":
			clauses = append(clauses, "cfg.is_enabled = 0")
		}
	}
	return buildWhereClause(clauses), args
}

func buildWhereClause(clauses []string) string {
	if len(clauses) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(clauses, " AND ")
}
