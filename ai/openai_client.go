package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	cl    *openai.Client
	model string
}

type reasoning struct {
	Topic string `json:"topic"`
	Year  string `json:"year,omitempty"`
}

func NewOpenAIClient(key, model string) *OpenAIClient {
	return &OpenAIClient{
		cl:    openai.NewClient(key),
		model: model,
	}
}

func (o *OpenAIClient) SuggestFilename(content string) (string, error) {
	ctx := context.Background()
	reasonPrompt := fmt.Sprintf(`Identify the main subject of this text in ≤5 words.
If a clear year (e.g., 1959, 2022) appears, include it.
Return JSON exactly like {"topic":"...", "year":""} with empty year if none.

TEXT:
"""
%s
"""`, content)

	step1, err := o.cl.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo0125,
		MaxTokens:   32,
		Temperature: 0.2,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "You are a structured data extractor."},
			{Role: "user", Content: reasonPrompt},
		},
	})
	if err != nil {
		return "", err
	}

	var r reasoning
	if err := json.Unmarshal([]byte(step1.Choices[0].Message.Content), &r); err != nil {
		return "", fmt.Errorf("parse step1 JSON: %w (raw=%s)", err, step1.Choices[0].Message.Content)
	}

	builder := strings.TrimSpace(r.Topic)
	if r.Year != "" {
		builder = builder + " " + r.Year
	}

	formatPrompt := fmt.Sprintf(`Create a single filename (5–10 words) from: %q
Format: lowercase, words separated by dashes, no extension, no generic words like "document" or "file".
Respond with the filename only.`, builder)

	step2, err := o.cl.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo0125,
		MaxTokens:   20,
		Temperature: 0.3,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "You are a file‑naming assistant."},
			{Role: "user", Content: formatPrompt},
		},
	})
	if err != nil {
		return "", err
	}
	// return first line trimmed – post‑processing will sanitize
	filename := strings.SplitN(step2.Choices[0].Message.Content, "\n", 2)[0]

	return strings.TrimSpace(filename), nil
}
