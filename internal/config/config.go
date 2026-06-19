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
	"gopkg.in/yaml.v3"
)

const configFile = "config.yaml"

// SecretsConfig holds sensitive credentials loaded from environment variables.
// These are never written to config.yaml and are excluded from version control.
type SecretsConfig struct {
	GeminiKey     string
	TelegramToken string
}

// NATSConfig holds sensitive NATS connection credentials
type NATSConfig struct {
	Host  string
	Port  int
	Token string
}

// TelegramConfig holds non-sensitive Telegram settings from config.yaml.
type TelegramConfig struct {
	ChatID int64 `yaml:"chat_id"`
}

// GeminiConfig holds configuration for Google Gemini AI.
type GeminiConfig struct {
	Model string `yaml:"model"`
}

// BotConfig holds core bot runtime settings.
type BotConfig struct {
	LogLevel    string `yaml:"log_level"`
	PollTimeout int    `yaml:"poll_timeout"`
}

// PluginsConfig aggregates all plugin configurations.
type PluginsConfig struct {
	// Profanity filter plugin settings.
	Profanity ProfanityConfig `yaml:"profanity"`
}

// ProfanityConfig holds configuration for the profanity filter plugin.
type ProfanityConfig struct {
	Enabled  bool     `yaml:"enabled"`
	BadWords []string `yaml:"bad_words"`
	Action   string   `yaml:"action"` // delete, warn, mute
}

// Config holds all configuration for the bot.
type Config struct {
	Secrets  SecretsConfig  `yaml:"-"`
	NATS     NATSConfig     `yaml:"-"`
	Bot      BotConfig      `yaml:"bot"`
	Telegram TelegramConfig `yaml:"telegram"`
	Gemini   GeminiConfig   `yaml:"gemini"`
	Prompts  *PromptsConfig `yaml:"prompts"`
	Plugins  PluginsConfig  `yaml:"plugins"`
}

// Load reads .env file, parses environment variables, and loads prompts from YAML.
// It exits with fatal log if required variables are missing or invalid.
func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load config from file
	cfg, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Append from env
	cfg.Secrets = SecretsConfig{
		GeminiKey:     os.Getenv("GEMINI_API_KEY"),
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
	}

	cfg.NATS = NATSConfig{
		Host:  os.Getenv("NATS_HOST"),
		Port:  parsePort(os.Getenv("NATS_PORT")),
		Token: os.Getenv("NATS_TOKEN"),
	}

	// Validation
	if cfg.Secrets.TelegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	return cfg
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

// loadConfig reads and parses the config YAML file at the given path.
func loadConfig(path string) (*Config, error) {
	// #nosec G304 — prompts file path is configured by the developer
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func parsePort(s string) int {
	p, _ := strconv.Atoi(s)
	return p
}
