package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type anthropicClient struct {
	name     string
	provider string
	apiKey   string
	endpoint string
}

func NewAnthropicClient(name, provider, apiKey, endpoint string) LLMClient {
	if endpoint == "" {
		endpoint = "https://api.anthropic.com"
	}
	return &anthropicClient{
		name:     name,
		provider: provider,
		apiKey:   apiKey,
		endpoint: endpoint,
	}
}

func (c *anthropicClient) Name() string     { return c.name }
func (c *anthropicClient) Provider() string { return c.provider }

type antMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type antRequest struct {
	Model       string       `json:"model"`
	System      string       `json:"system,omitempty"`
	Messages    []antMessage `json:"messages"`
	MaxTokens   int          `json:"max_tokens"`
	Temperature float64      `json:"temperature"`
}

type antResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *anthropicClient) Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	var system string
	messages := make([]antMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == RoleSystem {
			system = m.Content
			continue
		}
		messages = append(messages, antMessage{Role: string(m.Role), Content: m.Content})
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	antReq := antRequest{
		Model:       c.name,
		System:      system,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
	}

	jsonData, err := json.Marshal(antReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/messages", c.endpoint)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var antErr antResponse
		json.Unmarshal(body, &antErr)
		return nil, fmt.Errorf("anthropic error (%d): %s", resp.StatusCode, antErr.Error.Message)
	}

	var antResp antResponse
	if err := json.Unmarshal(body, &antResp); err != nil {
		return nil, err
	}

	if len(antResp.Content) == 0 {
		return nil, fmt.Errorf("no content returned from anthropic")
	}

	return &CompletionResponse{
		Content: antResp.Content[0].Text,
		Usage: Usage{
			PromptTokens:     antResp.Usage.InputTokens,
			CompletionTokens: antResp.Usage.OutputTokens,
			TotalTokens:      antResp.Usage.InputTokens + antResp.Usage.OutputTokens,
		},
	}, nil
}
