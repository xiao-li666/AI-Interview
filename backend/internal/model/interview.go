package model

import "time"

type InterviewPlan struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"userId"`
	JobTitle         string    `json:"jobTitle"`
	JobCategory      string    `json:"jobCategory"`
	LevelCode        string    `json:"levelCode"`
	InterviewType    string    `json:"interviewType"`
	InterviewMode    string    `json:"interviewMode"`
	DifficultyLevel  string    `json:"difficultyLevel"`
	QuestionCount    int       `json:"questionCount"`
	InterviewerStyle string    `json:"interviewerStyle"`
	CreatedAt        time.Time `json:"createdAt"`
}

type InterviewSession struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"userId"`
	CandidateProfileID int64     `json:"candidateProfileId"`
	JobTargetID       int64      `json:"jobTargetId"`
	ResumeID          int64      `json:"resumeId"`
	JobTitle          string     `json:"jobTitle"`
	JobCategory       string     `json:"jobCategory"`
	LevelCode         string     `json:"levelCode"`
	InterviewType     string     `json:"interviewType"`
	CompanyName       string     `json:"companyName"`
	ResumeFileName    string     `json:"resumeFileName"`
	ResumeText        string     `json:"resumeText"`
	SessionName       string     `json:"sessionName"`
	RoundType         string     `json:"roundType"`
	InterviewMode     string     `json:"interviewMode"`
	InterviewerStyle  string     `json:"interviewerStyle"`
	DifficultyLevel   string     `json:"difficultyLevel"`
	QuestionCount     int        `json:"questionCount"`
	CurrentQuestionNo int        `json:"currentQuestionNo"`
	Status            string     `json:"status"`
	StartedAt         *time.Time `json:"startedAt,omitempty"`
	EndedAt           *time.Time `json:"endedAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type InterviewQuestion struct {
	ID                  int64     `json:"id"`
	SessionID           int64     `json:"sessionId"`
	QuestionNo          int       `json:"questionNo"`
	QuestionType        string    `json:"questionType"`
	AssessmentDimension string    `json:"assessmentDimension"`
	Prompt              string    `json:"prompt"`
	ExpectedPoints      []string  `json:"expectedPoints"`
	Source              string    `json:"source"`
	IsFollowUp          bool      `json:"isFollowUp"`
	ParentQuestionID    *int64    `json:"parentQuestionId,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
}

type InterviewAnswer struct {
	ID                    int64     `json:"id"`
	SessionID             int64     `json:"sessionId"`
	QuestionID            int64     `json:"questionId"`
	AnswerText            string    `json:"answerText"`
	AnswerAudioURL        string    `json:"answerAudioUrl,omitempty"`
	AnswerDurationSeconds int       `json:"answerDurationSeconds"`
	SubmittedAt           time.Time `json:"submittedAt"`
}

type AnswerFeedback struct {
	ID                    int64     `json:"id"`
	AnswerID              int64     `json:"answerId"`
	OverallScore          float64   `json:"overallScore"`
	AccuracyScore         float64   `json:"accuracyScore"`
	ClarityScore          float64   `json:"clarityScore"`
	DepthScore            float64   `json:"depthScore"`
	CommunicationScore    float64   `json:"communicationScore"`
	Strengths             string    `json:"strengths"`
	Issues                string    `json:"issues"`
	ImprovementSuggestion string    `json:"improvementSuggestion"`
	FollowUpQuestion      string    `json:"followUpQuestion"`
	CreatedAt             time.Time `json:"createdAt"`
	Source                string    `json:"source,omitempty"`
}

type SessionAnswerItem struct {
	QuestionID          int64           `json:"questionId"`
	QuestionNo          int             `json:"questionNo"`
	QuestionType        string          `json:"questionType"`
	AssessmentDimension string          `json:"assessmentDimension"`
	Prompt              string          `json:"prompt"`
	ExpectedPoints      []string        `json:"expectedPoints"`
	Answer              InterviewAnswer `json:"answer"`
	Feedback            *AnswerFeedback `json:"feedback,omitempty"`
}

type QuestionReviewItem struct {
	QuestionID          int64    `json:"questionId"`
	QuestionNo          int      `json:"questionNo"`
	QuestionType        string   `json:"questionType"`
	AssessmentDimension string   `json:"assessmentDimension"`
	Prompt              string   `json:"prompt"`
	ExpectedPoints      []string `json:"expectedPoints"`
	ReviewType          string   `json:"reviewType"`
	Title               string   `json:"title"`
	Content             string   `json:"content"`
}

type ReportDimensionScore struct {
	DimensionCode string  `json:"dimensionCode"`
	Score         float64 `json:"score"`
	Comment       string  `json:"comment"`
}

type InterviewReport struct {
	ID                 int64                  `json:"id"`
	SessionID          int64                  `json:"sessionId"`
	OverallScore       float64                `json:"overallScore"`
	PerformanceSummary string                 `json:"performanceSummary"`
	StrengthsSummary   string                 `json:"strengthsSummary"`
	WeaknessSummary    string                 `json:"weaknessSummary"`
	RecommendedActions string                 `json:"recommendedActions"`
	NextPlan           map[string]any         `json:"nextPlan"`
	DimensionScores    []ReportDimensionScore `json:"dimensionScores"`
	CreatedAt          time.Time              `json:"createdAt"`
	Source             string                 `json:"source,omitempty"`
}

type SessionDetail struct {
	Session   InterviewSession    `json:"session"`
	Questions []InterviewQuestion `json:"questions"`
}

type HistoryItem struct {
	SessionID    int64      `json:"sessionId"`
	SessionName  string     `json:"sessionName"`
	JobTitle     string     `json:"jobTitle"`
	RoundType    string     `json:"roundType"`
	Status       string     `json:"status"`
	OverallScore *float64   `json:"overallScore,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
}
