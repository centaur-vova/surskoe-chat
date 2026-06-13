package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	url string
}

func New(modelName string, apiKey string) *Client {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, apiKey)

	return &Client{
		url: url,
	}
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
func (c *Client) Ask(systemPrompt, userPrompt string, search bool) (string, error) {
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

	resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Google API errors
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
