// Package gemini provides a simple client for Google's Gemini API.
//
// It supports text generation with optional system instructions and
// Google Search grounding. The client is minimal and focused on
// the specific needs of the Telegram bot.
package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/centaur-vova/telegram-go-chat-skeleton/internal/logger"
)

// Client is a Gemini API client configured with model and API key.
type Client struct {
	url string
}

// New creates a new Gemini client for the given model name and API key.
//
// Example:
//
//	client := gemini.New("gemini-flash-latest", os.Getenv("GEMINI_API_KEY"))
func New(modelName string, apiKey string) *Client {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, apiKey)

	return &Client{
		url: url,
	}
}

// / googleSearch enables grounding with Google Search when included in Tools.
type googleSearch struct{}

// Tool represents a tool that Gemini can use, such as Google Search.

type tool struct {
	GoogleSearch *googleSearch `json:"googleSearch,omitempty"`
}

// geminiPart represents a single part of the message content (text, image, etc.).
type geminiPart struct {
	Text string `json:"text"`
}

// GeminiContent represents a message content consisting of multiple parts.

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

// SystemInstruction holds the system-level instructions for the model.

type systemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

// GeminiRequest is the request payload for the Gemini generateContent endpoint.

type geminiRequest struct {
	SystemInstruction *systemInstruction `json:"systemInstruction,omitempty"`
	Contents          []geminiContent    `json:"contents"`
	Tools             []tool             `json:"tools,omitempty"`
}

// GeminiResponse is the response payload from the Gemini generateContent endpoint.

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Ask sends a prompt to Gemini API and returns generated text
func (c *Client) Ask(systemPrompt, userPrompt string, search bool) (string, error) {
	var tools []tool
	if search {
		tools = []tool{
			{GoogleSearch: &googleSearch{}},
		}
	}
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: userPrompt}}},
		},
		SystemInstruction: &systemInstruction{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Tools: tools,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Error closing body", "error", err)
		}

	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Google API errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google API error (Status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response received")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
