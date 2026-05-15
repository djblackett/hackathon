package config

import (
	"os"
)

type Config struct {
	OpenAIKey string
	ServerURL string
	OllaHost  string
	TikaURL   string
}

func FromEnv() Config {

	openAIKey := os.Getenv("OPENAI_API_KEY")
	return Config{
		OpenAIKey: openAIKey,
		OllaHost:  os.Getenv("OLLAMA_HOST"),
		ServerURL: os.Getenv("AI_SERVER_URL"),
		TikaURL:   os.Getenv("TIKA_URL"),
	}
}
