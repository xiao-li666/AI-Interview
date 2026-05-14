package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	MySQL    MySQLConfig
	DeepSeek DeepSeekConfig
	OpenAI   OpenAIConfig
}

type ServerConfig struct {
	Port int
}

type MySQLConfig struct {
	DSN      string
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

type DeepSeekConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Port: intFromEnv("APP_PORT", 8080),
		},
		MySQL: MySQLConfig{
			DSN:      stringFromEnv("MYSQL_DSN", ""),
			Host:     stringFromEnv("MYSQL_HOST", "127.0.0.1"),
			Port:     intFromEnv("MYSQL_PORT", 3306),
			User:     stringFromEnv("MYSQL_USER", "root"),
			Password: stringFromEnv("MYSQL_PASSWORD", "6204222lL@"),
			Database: stringFromEnv("MYSQL_DATABASE", "ai_interview"),
		},
		DeepSeek: DeepSeekConfig{
			APIKey:  stringFromEnv("DEEPSEEK_API_KEY", ""),
			BaseURL: stringFromEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
			Model:   stringFromEnv("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		},
		OpenAI: OpenAIConfig{
			APIKey:  stringFromEnv("OPENAI_API_KEY", ""),
			BaseURL: stringFromEnv("OPENAI_BASE_URL", ""),
			Model:   stringFromEnv("OPENAI_MODEL", "gpt-5.5"),
		},
	}
}

func stringFromEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func intFromEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
