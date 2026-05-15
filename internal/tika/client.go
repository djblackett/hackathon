package tika

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultTimeout = 15 * time.Second

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Extraction struct {
	Text     string
	Metadata map[string]string
	Warnings []string
}

func NewClient(baseURL string) (*Client, error) {
	return NewClientWithHTTPClient(baseURL, &http.Client{Timeout: defaultTimeout})
}

func NewClientWithHTTPClient(baseURL string, httpClient *http.Client) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("empty tika url")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid tika url %q: %w", baseURL, err)
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}

func (c *Client) ExtractFile(ctx context.Context, path string) (Extraction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Extraction{}, err
	}

	text, textErr := c.put(ctx, "/tika", data, "text/plain")
	metadata, metaErr := c.extractMetadata(ctx, data)
	if textErr != nil && metaErr != nil {
		return Extraction{}, fmt.Errorf("tika text extraction failed: %v; metadata extraction failed: %w", textErr, metaErr)
	}
	if textErr != nil {
		return Extraction{
			Metadata: metadata,
			Warnings: []string{
				"tika text extraction failed: " + textErr.Error(),
			},
		}, nil
	}
	if metaErr != nil {
		return Extraction{
			Text: text,
			Warnings: []string{
				"tika metadata extraction failed: " + metaErr.Error(),
			},
		}, nil
	}
	return Extraction{Text: text, Metadata: metadata}, nil
}

func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/tika", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("tika health returned %s", resp.Status)
	}
	return nil
}

func (c *Client) extractMetadata(ctx context.Context, data []byte) (map[string]string, error) {
	body, err := c.put(ctx, "/meta", data, "application/json")
	if err != nil {
		return nil, err
	}

	var object map[string]any
	if err := json.Unmarshal([]byte(body), &object); err == nil {
		return flattenMetadata(object), nil
	}

	var list []map[string]any
	if err := json.Unmarshal([]byte(body), &list); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return map[string]string{}, nil
	}
	return flattenMetadata(list[0]), nil
}

func (c *Client) put(ctx context.Context, endpoint string, data []byte, accept string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", accept)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("tika %s returned %s: %s", endpoint, resp.Status, strings.TrimSpace(string(limited)))
	}

	limited, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return "", err
	}
	return string(limited), nil
}

func flattenMetadata(values map[string]any) map[string]string {
	out := map[string]string{}
	for key, value := range values {
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				out[key] = strings.TrimSpace(v)
			}
		case []any:
			parts := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
					parts = append(parts, strings.TrimSpace(s))
				}
			}
			if len(parts) > 0 {
				out[key] = strings.Join(parts, ", ")
			}
		default:
			if value != nil {
				out[key] = strings.TrimSpace(fmt.Sprint(value))
			}
		}
	}
	return out
}
