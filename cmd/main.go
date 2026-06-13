package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"gopkg.in/yaml.v3"
)

const (
	geminiModelName = "gemini-flash-latest" // "gemini-2.5-flash"
	promptsFile     = "prompts.yaml"
)

var (
	apiKey   string
	botToken string
	chatID   telego.ChatID
	prompts  *PromptsConfig
)

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

// Google search
type GoogleSearch struct{}

type Tool struct {
	GoogleSearch *GoogleSearch `json:"googleSearch,omitempty"`
}

// Gemini API request structures
type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type SystemInstruction struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiRequest struct {
	SystemInstruction *SystemInstruction `json:"systemInstruction,omitempty"`
	Contents          []GeminiContent    `json:"contents"`
	Tools             []Tool             `json:"tools,omitempty"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// askGemini sends a prompt to Gemini API and returns generated text
func askGemini(modelName, systemPrompt, userPrompt string, search bool) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, apiKey)

	var tools []Tool
	if search {
		tools = []Tool{
			{GoogleSearch: &GoogleSearch{}},
		}
	}
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{Parts: []GeminiPart{{Text: userPrompt}}},
		},
		SystemInstruction: &SystemInstruction{
			Parts: []GeminiPart{{Text: systemPrompt}},
		},
		Tools: tools,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Обработка ошибок от самого Google (например, если ключ не подошел)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Google API error (Status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("Empty response received")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey = os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatalf("GEMINI_API_KEY is not set")
	}

	botToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
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
	chatID = telego.ChatID{ID: chatIDInt}

	// Load prompts
	prompts, err = loadPrompts(promptsFile)
	if err != nil {
		log.Fatalf("Failed to load prompts: %v", err)
	}

	// Context
	ctx := context.Background()

	// Create bot instance
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatal(err)
	}

	// Get updates channel via long polling with context
	params := &telego.GetUpdatesParams{
		Timeout: 30,
	}
	updates, err := bot.UpdatesViaLongPolling(ctx, params)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("🚀 StarPom of the Spacefleet is online! Listening for commands in chat %v...\n", chatID)

	// Process incoming updates
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Ignore messages from other chats
		if update.Message.Chat.ID != chatID.ID {
			continue
		}

		text := update.Message.Text

		// /news command — generate a news article
		if text == "/news" || text == "/news@surskoe_chat_bot" {
			msg, _ := bot.SendMessage(ctx, telegoutil.Message(chatID, prompts.Messages.Recording))

			userPrompt := `Вспомни одно реальное позитивное событие из жизни любой российской сельской глубинки за последние два года (например: благоустройство сквера, ремонт дороги, открытие почты, победа на районных соревнованиях, закупка тракторов). Перенеси это событие в р.п. Сурское и перепиши строго по инструкции`
			news, err := askGemini(geminiModelName, prompts.News.System, userPrompt, false)
			if err != nil {
				log.Printf("Gemini error: %v", err)
				bot.SendMessage(ctx, telegoutil.Message(chatID, prompts.Messages.Error))
				continue
			}

			// Delete "recording..." message and send the news
			bot.DeleteMessage(ctx, telegoutil.Delete(chatID, msg.MessageID))
			bot.SendMessage(ctx, telegoutil.Message(chatID, "📡 "+news))
		}

		// /start command — welcome message
		if text == "/start" || text == "/start@surskoe_chat_bot" {
			bot.SendMessage(ctx, telegoutil.Message(chatID, prompts.Messages.Welcome))
		}
	}
}
