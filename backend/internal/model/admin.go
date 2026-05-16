package model

import "time"

type Admin struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Nickname     string    `json:"nickname"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type AdminSession struct {
	Token string `json:"token"`
	Admin Admin  `json:"admin"`
}

type AdminListQuery struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Keyword  string `json:"keyword"`
	Status   string `json:"status"`
}

type PagedResult[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type AdminUserListItem struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AdminUserDetail struct {
	User         User `json:"user"`
	SessionCount int  `json:"sessionCount"`
	ResumeCount  int  `json:"resumeCount"`
	ReportCount  int  `json:"reportCount"`
	AIConfigCount int `json:"aiConfigCount"`
}

type AdminUserSummary struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
}

type AdminInterviewSessionListItem struct {
	SessionID     int64      `json:"sessionId"`
	UserID        int64      `json:"userId"`
	UserEmail     string     `json:"userEmail"`
	UserNickname  string     `json:"userNickname"`
	SessionName   string     `json:"sessionName"`
	JobTitle      string     `json:"jobTitle"`
	CompanyName   string     `json:"companyName"`
	RoundType     string     `json:"roundType"`
	Status        string     `json:"status"`
	OverallScore  *float64   `json:"overallScore,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
}

type AdminInterviewSessionDetail struct {
	User      AdminUserSummary `json:"user"`
	Session   InterviewSession `json:"session"`
	Questions []InterviewQuestion `json:"questions"`
}

type AdminResumeListItem struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"userId"`
	UserEmail   string    `json:"userEmail"`
	UserNickname string   `json:"userNickname"`
	Name        string    `json:"name"`
	SourceType  string    `json:"sourceType"`
	IsDefault   bool      `json:"isDefault"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AdminResumeDetail struct {
	ID           int64      `json:"id"`
	User         AdminUserSummary `json:"user"`
	Name         string     `json:"name"`
	SourceType   string     `json:"sourceType"`
	FileURL      string     `json:"fileUrl"`
	RawText      string     `json:"rawText"`
	ParsedContent map[string]any `json:"parsedContent"`
	VersionNo    int        `json:"versionNo"`
	IsDefault    bool       `json:"isDefault"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type AdminAIConfigListItem struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"userId"`
	UserEmail    string     `json:"userEmail"`
	UserNickname string     `json:"userNickname"`
	ProviderCode string     `json:"provider"`
	APIKeyMasked string     `json:"apiKeyMasked"`
	Model        string     `json:"model"`
	BaseURL      string     `json:"baseUrl"`
	IsEnabled    bool       `json:"isEnabled"`
	LastTestOK   bool       `json:"lastTestOk"`
	LastTestedAt *time.Time `json:"lastTestedAt,omitempty"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type AdminOverview struct {
	AdminCount   int `json:"adminCount"`
	UserCount    int `json:"userCount"`
	SessionCount int `json:"sessionCount"`
	ResumeCount  int `json:"resumeCount"`
	ReportCount  int `json:"reportCount"`
	AIConfigCount int `json:"aiConfigCount"`
}
