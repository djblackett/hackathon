package ai

import (
	"errors"
	"fmt"

	"github.com/djblackett/bootdev-hackathon/internal/config"
)

type Client interface {
	SuggestFilename(content string) (string, error)
}

func NewClient(cfg config.Config, local bool, model string) (Client, error) {
	fmt.Println("Creating AI client...")
	fmt.Println("Model:", model)
	fmt.Println("Local:", local)
	apiKey := cfg.OpenAIKey
	switch {
	case local:
		fmt.Println("Using local Ollama client with model:", model)
		return NewOllamaClient(model), nil
	case apiKey != "":
		fmt.Println("Using OpenAI client with web API")
		return NewOpenAIClient(apiKey, model), nil
	case cfg.ServerURL != "": // new: check if server URL is configured
		return NewHTTPClient(cfg.ServerURL, model), nil
	}
	return nil, errors.New("no AI backend configured")
}
