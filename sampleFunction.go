package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ollamaURL = "http://localhost:11434/api/generate"
const ollamaModel = "mistral" // or "tinyllama", "phi", etc. mistral "deepseek-r1:8b"

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"` // we set this to false for simplicity
}

type OllamaResponse struct {
	Response string `json:"response"` // the actual generated text
	Done     bool   `json:"done"`
}

// GenerateFilenameFromText sends file content to the local LLM and returns a suggested filename.
func GenerateFilenameFromText(content string) (string, error) {
	prompt := fmt.Sprintf(`You are a file recovery assistant. A document was recovered from a damaged hard drive, but its filename was lost.

Your task is to infer a new, meaningful filename based on the document’s contents.

Here is the content of the file:
"""
%s
"""

Please generate **a single best filename** that describes the file as a whole. Use your judgment to decide what is most important or representative of the file's content.

**Formatting rules:**
- Use lowercase letters, numbers, dashes or underscores
- Do not include a file extension
- Limit the name to 5–10 words
- Respond with the filename only — no quotes, punctuation, or explanation

If the content is fragmented or noisy, summarize what it's *likely* to be based on keywords, structure, or clues.

Respond with only one filename.
`, content)

	payload := OllamaRequest{
		Model:  ollamaModel,
		Prompt: prompt,
		Stream: false,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	client := http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(ollamaURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error: %s", string(body))
	}

	var result OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Response, nil
}
