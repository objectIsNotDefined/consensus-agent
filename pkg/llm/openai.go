package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type openaiClient struct {
	name     string
	provider string
	apiKey   string
	endpoint string
}

// NewOpenAIClient creates a client for OpenAI or OpenAI-compatible APIs (like Deepseek).
func NewOpenAIClient(name, provider, apiKey, endpoint string) LLMClient {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}
	// Deepseek and many providers need the /v1 suffix if not provided
	if provider == "deepseek" && !strings.HasSuffix(endpoint, "/v1") && !strings.Contains(endpoint, "/v1/") {
		endpoint = strings.TrimSuffix(endpoint, "/") + "/v1"
	}

	return &openaiClient{
		name:     name,
		provider: provider,
		apiKey:   apiKey,
		endpoint: endpoint,
	}
}

func (c *openaiClient) Name() string     { return c.name }
func (c *openaiClient) Provider() string { return c.provider }

type oaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type oaRequest struct {
	Model       string      `json:"model"`
	Messages    []oaMessage `json:"messages"`
	Temperature float64     `json:"temperature"`
	MaxTokens   int         `json:"max_tokens,omitempty"`
}

type oaResponse struct {
	Choices []struct {
		Message oaMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *openaiClient) Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	messages := make([]oaMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = oaMessage{Role: string(m.Role), Content: m.Content}
	}

	oaReq := oaRequest{
		Model:       c.name,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	jsonData, err := json.Marshal(oaReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/chat/completions", c.endpoint)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var oaErr oaResponse
		json.Unmarshal(body, &oaErr)
		return nil, fmt.Errorf("openai error (%d): %s", resp.StatusCode, oaErr.Error.Message)
	}

	var oaResp oaResponse
	if err := json.Unmarshal(body, &oaResp); err != nil {
		return nil, err
	}

	if len(oaResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from openai")
	}

	return &CompletionResponse{
		Content: oaResp.Choices[0].Message.Content,
		Usage: Usage{
			PromptTokens:     oaResp.Usage.PromptTokens,
			CompletionTokens: oaResp.Usage.CompletionTokens,
			TotalTokens:      oaResp.Usage.TotalTokens,
		},
	}, nil
}
