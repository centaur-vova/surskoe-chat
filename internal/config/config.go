// Package config provides configuration loading from .env and YAML files.
//
// It handles environment variables for Telegram bot token, Gemini API keys,
// chat ID, poll timeout, and log level. Prompts for commands are loaded
// from a separate YAML file (prompts.yaml).
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"gopkg.in/yaml.v3"
)

const promptsFile = "prompts.yaml"

// Config holds all configuration for the bot.
type Config struct {
	LogLevel string

	GeminiKey     string
	GeminiModel   string
	TelegramToken string
	PollTimeout   int
	ChatID        telego.ChatID

	Prompts *PromptsConfig
}

// Load reads .env file, parses environment variables, and loads prompts from YAML.
// It exits with fatal log if required variables are missing or invalid.
func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		log.Fatalf("GEMINI_API_KEY is not set")
	}

	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		log.Fatalf("GEMINI_MODEL is not set")
	}

	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if telegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		log.Fatal("TELEGRAM_CHAT_ID is not set")
	}

	chatIDInt, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatal("Invalid TELEGRAM_CHAT_ID format")
	}
	chatID := telego.ChatID{ID: chatIDInt}

	pollTimeoutStr := os.Getenv("POLL_TIMEOUT")
	if pollTimeoutStr == "" {
		log.Fatal("POLL_TIMEOUT is not set")
	}
	pollTimeout, err := strconv.Atoi(pollTimeoutStr)
	if err != nil {
		log.Fatalf("Invalid POLL_TIMEOUT format: %v", err)
	}

	// Load prompts
	prompts, err := loadPrompts(promptsFile)
	if err != nil {
		log.Fatalf("Failed to load prompts: %v", err)
	}

	return Config{
		// App
		LogLevel: os.Getenv("LOG_LEVEL"),

		// Keys/Config
		GeminiKey:     geminiKey,
		GeminiModel:   geminiModel,
		TelegramToken: telegramToken,
		ChatID:        chatID,

		// Settings
		PollTimeout: pollTimeout,
		Prompts:     prompts,
	}
}

// PromptsConfig contains all prompt templates loaded from YAML.
type PromptsConfig struct {
	Messages  promptsConfigMessages   `yaml:"messages"`
	News      struct{ System string } `yaml:"news"`
	Interview struct{ System string } `yaml:"interview"`
}

// PromptsConfigMessages contains message templates for bot responses.
type promptsConfigMessages struct {
	Recording string `yaml:"recording"`
	Error     string `yaml:"error"`
	Welcome   string `yaml:"welcome"`
}

// loadPrompts reads and parses the prompts YAML file at the given path.
func loadPrompts(path string) (*PromptsConfig, error) {
	// #nosec G304 — prompts file path is configured by the developer
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompts file: %w", err)
	}

	var cfg PromptsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse prompts file: %w", err)
	}

	return &cfg, nil
}
