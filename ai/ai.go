package ai

import "errors"

type Client interface {
	SuggestFilename(content string) (string, error)
}

func NewClient(local bool, model string) (Client, error) {
	if local {
		return NewOllamaClient(model), nil
	}
	if apiKey := getenvOpenAI(); apiKey != "" {
		return NewOpenAIClient(apiKey, model), nil
	}
	return nil, errors.New("no AI backend configured")
}

func getenvOpenAI() string { /* read env var or config */ return "" }
