package llm

import (
	"context"
)

// Role defines the role of a message in a chat completion.
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// Message represents a single turn in a conversation.
type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
}

// CompletionRequest defines the input for the LLM.
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	Stream      bool      `json:"stream"`
}

// CompletionResponse defines the output from the LLM.
type CompletionResponse struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
}

// Usage tracks token consumption.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// LLMClient is the common interface for all providers.
type LLMClient interface {
	Name() string
	Provider() string
	Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}

// NewClient is a factory for creating LLM clients based on provider.
func NewClient(name, provider, apiKey, endpoint string) LLMClient {
	switch provider {
	case "openai", "deepseek":
		return NewOpenAIClient(name, provider, apiKey, endpoint)
	case "anthropic":
		return NewAnthropicClient(name, provider, apiKey, endpoint)
	case "google", "gemini":
		return NewGeminiClient(name, provider, apiKey, endpoint)
	default:
		// Fallback to OpenAI compatible for unknown providers
		return NewOpenAIClient(name, provider, apiKey, endpoint)
	}
}
