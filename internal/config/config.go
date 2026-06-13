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

type Config struct {
	LogLevel string

	GeminiKey     string
	GeminiModel   string
	TelegramToken string
	PollTimeout   int
	ChatID        telego.ChatID

	Prompts *PromptsConfig
}

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

type PromptsConfigMessages struct {
	Recording string `yaml:"recording"`
	Error     string `yaml:"error"`
	Welcome   string `yaml:"welcome"`
}

type PromptsConfig struct {
	Messages  PromptsConfigMessages   `yaml:"messages"`
	News      struct{ System string } `yaml:"news"`
	Interview struct{ System string } `yaml:"interview"`
}

// loadPrompts reads and parses the prompts YAML file
func loadPrompts(path string) (*PromptsConfig, error) {
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
