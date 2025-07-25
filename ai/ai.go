package ai

import (
	"errors"
	"fmt"

	"github.com/djblackett/bootdev-hackathon/config"
)

type Client interface {
	SuggestFilename(content string) (string, error)
}

func NewClient(cfg config.Config, local bool, model string) (Client, error) {
	fmt.Println("Creating AI client...")
	fmt.Println("Model:", model)
	fmt.Println("Local:", local)
	if local {
		fmt.Println("Using local Ollama client with modeal:", model)
		return NewOllamaClient(model), nil
	}
	if apiKey := cfg.OpenAIKey; apiKey != "" {
		fmt.Println("Using OpenAI client with web API")
		return NewOpenAIClient(apiKey, model), nil
	}
	return nil, errors.New("no AI backend configured")
}
