package service

import (
	"errors"
	"strings"
	"time"

	"ai-interview/backend/internal/ai"
	aideepseek "ai-interview/backend/internal/ai/deepseek"
	aimock "ai-interview/backend/internal/ai/mock"
	openaiai "ai-interview/backend/internal/ai/openai"
	"ai-interview/backend/internal/config"
)

type ProviderFactory struct {
	config          config.Config
	mockProvider    ai.Provider
	defaultProvider ai.Provider
}

func NewProviderFactory(cfg config.Config) *ProviderFactory {
	mockProvider := aimock.NewProvider()

	var defaultProvider ai.Provider = mockProvider
	if cfg.DeepSeek.APIKey != "" {
		deepSeekProvider, err := aideepseek.NewProvider(aideepseek.Config{
			APIKey:  cfg.DeepSeek.APIKey,
			BaseURL: cfg.DeepSeek.BaseURL,
			Model:   cfg.DeepSeek.Model,
			Timeout: 90 * time.Second,
		})
		if err == nil {
			defaultProvider = ai.NewFallbackProvider(deepSeekProvider, mockProvider)
		}
	} else if cfg.OpenAI.APIKey != "" {
		openAIProvider, err := openaiai.NewProvider(openaiai.Config{
			APIKey:  cfg.OpenAI.APIKey,
			BaseURL: cfg.OpenAI.BaseURL,
			Model:   cfg.OpenAI.Model,
			Timeout: 90 * time.Second,
		})
		if err == nil {
			defaultProvider = ai.NewFallbackProvider(openAIProvider, mockProvider)
		}
	}

	return &ProviderFactory{
		config:          cfg,
		mockProvider:    mockProvider,
		defaultProvider: defaultProvider,
	}
}

func (f *ProviderFactory) Default() ai.Provider {
	return f.defaultProvider
}

func (f *ProviderFactory) BuildFromUserInput(provider string, apiKey string, model string, baseURL string) (ai.Provider, error) {
	switch strings.TrimSpace(provider) {
	case "deepseek":
		return aideepseek.NewProvider(aideepseek.Config{
			APIKey:  apiKey,
			BaseURL: baseURL,
			Model:   model,
			Timeout: 90 * time.Second,
		})
	case "openai":
		return openaiai.NewProvider(openaiai.Config{
			APIKey:  apiKey,
			BaseURL: baseURL,
			Model:   model,
			Timeout: 90 * time.Second,
		})
	default:
		return nil, errors.New("unsupported provider")
	}
}
