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

type SaveAIConfigRequest struct {
	UserID    int64  `json:"userId"`
	Provider  string `json:"provider"`
	APIKey    string `json:"apiKey"`
	Model     string `json:"model"`
	BaseURL   string `json:"baseUrl"`
	IsEnabled bool   `json:"isEnabled"`
}

type TestAIConfigRequest struct {
	UserID   int64  `json:"userId"`
	Provider string `json:"provider"`
	APIKey   string `json:"apiKey"`
	Model    string `json:"model"`
	BaseURL  string `json:"baseUrl"`
}

type TestAIConfigResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (s *InterviewService) GetUserAIConfig(ctx context.Context, userID int64) (model.UserAIProviderConfig, error) {
	if userID == 0 {
		return model.UserAIProviderConfig{}, errors.New("userId is required")
	}

	cfg, err := s.repo.GetUserAIConfig(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.UserAIProviderConfig{}, ErrNotFound
		}
		return model.UserAIProviderConfig{}, err
	}

	cfg.APIKey = ""
	return cfg, nil
}

func (s *InterviewService) SaveUserAIConfig(ctx context.Context, req SaveAIConfigRequest) (model.UserAIProviderConfig, error) {
	if req.UserID == 0 {
		return model.UserAIProviderConfig{}, errors.New("userId is required")
	}
	if strings.TrimSpace(req.Provider) == "" {
		return model.UserAIProviderConfig{}, errors.New("provider is required")
	}

	existing, err := s.repo.GetUserAIConfig(ctx, req.UserID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return model.UserAIProviderConfig{}, err
	}

	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" && !errors.Is(err, repository.ErrNotFound) {
		apiKey = existing.APIKey
	}
	if req.IsEnabled && apiKey == "" {
		return model.UserAIProviderConfig{}, errors.New("apiKey is required when config is enabled")
	}

	cfg := model.UserAIProviderConfig{
		UserID:       req.UserID,
		ProviderCode: strings.TrimSpace(req.Provider),
		APIKey:       apiKey,
		APIKeyMasked: maskAPIKey(apiKey),
		Model:        defaultProviderModel(req.Provider, req.Model),
		BaseURL:      defaultProviderBaseURL(req.Provider, req.BaseURL),
		IsEnabled:    req.IsEnabled,
	}
	if !errors.Is(err, repository.ErrNotFound) {
		cfg.ID = existing.ID
		cfg.CreatedAt = existing.CreatedAt
		cfg.LastTestOK = existing.LastTestOK
		cfg.LastTestedAt = existing.LastTestedAt
	}

	saved, saveErr := s.repo.UpsertUserAIConfig(ctx, cfg)
	if saveErr != nil {
		return model.UserAIProviderConfig{}, saveErr
	}

	saved.APIKey = ""
	return saved, nil
}

func (s *InterviewService) TestUserAIConfig(ctx context.Context, req TestAIConfigRequest) (TestAIConfigResponse, error) {
	if req.UserID == 0 {
		return TestAIConfigResponse{}, errors.New("userId is required")
	}

	provider, err := s.providerFactory.BuildFromUserInput(req.Provider, req.APIKey, req.Model, req.BaseURL)
	if err != nil {
		return TestAIConfigResponse{}, err
	}

	_, callErr := provider.GenerateQuestions(ctx, buildPingQuestionsInput())
	if callErr != nil {
		return TestAIConfigResponse{
			OK:      false,
			Message: callErr.Error(),
		}, nil
	}

	now := time.Now()
	existing, getErr := s.repo.GetUserAIConfig(ctx, req.UserID)
	if getErr == nil {
		existing.LastTestOK = true
		existing.LastTestedAt = &now
		if _, err := s.repo.UpsertUserAIConfig(ctx, existing); err != nil {
			return TestAIConfigResponse{}, err
		}
	}

	return TestAIConfigResponse{
		OK:      true,
		Message: "连接成功",
	}, nil
}

func buildPingQuestionsInput() ai.GenerateQuestionsInput {
	return ai.GenerateQuestionsInput{
		JobTitle:         "Go后端工程师",
		RoundType:        "technical_1",
		InterviewMode:    "text",
		DifficultyLevel:  "medium",
		InterviewerStyle: "balanced",
		QuestionCount:    1,
	}
}

func maskAPIKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:6] + "****" + value[len(value)-4:]
}

func defaultProviderModel(provider string, value string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}

	switch strings.TrimSpace(provider) {
	case "deepseek":
		return "deepseek-v4-flash"
	case "openai":
		return "gpt-5.5"
	default:
		return value
	}
}

func defaultProviderBaseURL(provider string, value string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}

	switch strings.TrimSpace(provider) {
	case "deepseek":
		return "https://api.deepseek.com"
	case "openai":
		return "https://api.openai.com/v1/responses"
	default:
		return value
	}
}
