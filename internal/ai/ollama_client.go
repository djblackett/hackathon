package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type ollamaReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}
type ollamaResp struct {
	Response string `json:"response"`
}

type OllamaClient struct {
	model string
	url   string
}

func NewOllamaClient(model string) *OllamaClient {
	// Get OLLAMA_HOST from environment, fallback to localhost for local development
	url := os.Getenv("OLLAMA_HOST")
	if url == "" {
		url = "http://localhost:11434/api/generate"
	}
	return &OllamaClient{model: model, url: url}
}

func (o *OllamaClient) SuggestFilename(content string) (string, error) {
	prompt := buildPrompt(content)
	body, _ := json.Marshal(ollamaReq{Model: o.model, Prompt: prompt})

	cl := http.Client{Timeout: 300 * time.Second} // Increased to 5 minutes for model loading
	resp, err := cl.Post(o.url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out ollamaResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Response, nil
}
