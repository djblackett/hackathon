package ai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	cl    *openai.Client
	model string
}

func NewOpenAIClient(key, model string) *OpenAIClient {
	return &OpenAIClient{
		cl:    openai.NewClient(key),
		model: model,
	}
}

func (o *OpenAIClient) SuggestFilename(content string) (string, error) {
	resp, err := o.cl.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: "You are a file recovery assistant."},
			{Role: "user", Content: buildPrompt(content)},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
