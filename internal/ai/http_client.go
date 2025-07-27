package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPClient struct {
	baseURL string
	model   string
	client  *http.Client
}

type filenameRequest struct {
	Content string `json:"content"`
	Model   string `json:"model"`
}

type filenameResponse struct {
	Filename string `json:"filename"`
	Error    string `json:"error,omitempty"`
}

func NewHTTPClient(baseURL, model string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

func (c *HTTPClient) SuggestFilename(content string) (string, error) {
	req := filenameRequest{
		Content: content,
		Model:   c.model,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Post(c.baseURL+"/suggest-filename", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response filenameResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if response.Error != "" {
		return "", fmt.Errorf("server error: %s", response.Error)
	}

	return response.Filename, nil
}
