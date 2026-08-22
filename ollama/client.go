package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const url = "http://localhost:11434/api/chat"

type Client struct {
	url        string
	model      string
	httpClient *http.Client
}

func NewClient(model string) *Client {
	return &Client{
		url:        url,
		model:      model,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) Chat(ctx context.Context, evidence []byte) (string, error) {
	requestBody := ChatRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: string(evidence)},
		},
		Stream: false,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("could not reach Ollama at %s — is it running? try: ollama serve (%w)", c.url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(response.Body)

		if response.StatusCode == http.StatusNotFound || strings.Contains(string(responseBody), "not found") {
			return "", fmt.Errorf("model %q is not pulled yet — run: ollama pull %s", c.model, c.model)
		}
		return "", fmt.Errorf("ollama returned status %d: %s", response.StatusCode, string(responseBody))
	}

	var chatResponse ChatResponse
	if err := json.NewDecoder(response.Body).Decode(&chatResponse); err != nil {
		return "", fmt.Errorf("failed to decode response body: %w", err)
	}

	if chatResponse.Message.Content == "" {
		return "", fmt.Errorf("ollama returned an empty response — is model %q pulled? try: ollama pull %s", c.model, c.model)
	}

	return chatResponse.Message.Content, nil
}
