package config

import (
	"fmt"
	"os"
)

type Config struct {
	OpenAIKey string
	// OllaHost  string
	// Model     string
}

func FromEnv() Config {

	openAIKey := os.Getenv("OPENAI_API_KEY")
	fmt.Println("OpenAI Key:", openAIKey)
	return Config{
		OpenAIKey: openAIKey,
		// OllaHost:  os.Getenv("OLLAMA_HOST"), // optional
		// Model:     os.Getenv("MODEL"),       // can be empty; CLI flag wins
	}
}
