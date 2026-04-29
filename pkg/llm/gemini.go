package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type geminiClient struct {
	name     string
	provider string
	apiKey   string
	endpoint string
}

func NewGeminiClient(name, provider, apiKey, endpoint string) LLMClient {
	if endpoint == "" {
		endpoint = "https://generativelanguage.googleapis.com"
	}
	return &geminiClient{
		name:     name,
		provider: provider,
		apiKey:   apiKey,
		endpoint: endpoint,
	}
}

func (c *geminiClient) Name() string     { return c.name }
func (c *geminiClient) Provider() string { return c.provider }

type gemContent struct {
	Role  string `json:"role"`
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
}

type gemRequest struct {
	Contents         []gemContent `json:"contents"`
	SystemInstruction *struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"system_instruction,omitempty"`
	GenerationConfig struct {
		Temperature float64 `json:"temperature"`
		MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
	} `json:"generationConfig"`
}

type gemResponse struct {
	Candidates []struct {
		Content gemContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *geminiClient) Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	var systemParts []struct {
		Text string `json:"text"`
	}
	contents := make([]gemContent, 0, len(req.Messages))

	for _, m := range req.Messages {
		role := string(m.Role)
		if role == "assistant" {
			role = "model"
		}
		
		if role == string(RoleSystem) {
			systemParts = append(systemParts, struct {
				Text string `json:"text"`
			}{Text: m.Content})
			continue
		}

		content := gemContent{Role: role}
		content.Parts = append(content.Parts, struct {
			Text string `json:"text"`
		}{Text: m.Content})
		contents = append(contents, content)
	}

	gemReq := gemRequest{
		Contents: contents,
	}
	if len(systemParts) > 0 {
		gemReq.SystemInstruction = &struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{Parts: systemParts}
	}
	gemReq.GenerationConfig.Temperature = req.Temperature
	gemReq.GenerationConfig.MaxOutputTokens = req.MaxTokens

	jsonData, err := json.Marshal(gemReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", c.endpoint, c.name, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var gemErr gemResponse
		json.Unmarshal(body, &gemErr)
		return nil, fmt.Errorf("gemini error (%d): %s", resp.StatusCode, gemErr.Error.Message)
	}

	var gemResp gemResponse
	if err := json.Unmarshal(body, &gemResp); err != nil {
		return nil, err
	}

	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content returned from gemini")
	}

	return &CompletionResponse{
		Content: gemResp.Candidates[0].Content.Parts[0].Text,
		Usage: Usage{
			PromptTokens:     gemResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: gemResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      gemResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}
