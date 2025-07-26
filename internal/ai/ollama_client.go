package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
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
	return &OllamaClient{model: model, url: "http://localhost:11434/api/generate"}
}

func (o *OllamaClient) SuggestFilename(content string) (string, error) {
	prompt := buildPrompt(content)
	body, _ := json.Marshal(ollamaReq{Model: o.model, Prompt: prompt})

	cl := http.Client{Timeout: 60 * time.Second}
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
