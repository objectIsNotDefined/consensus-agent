package llm

import (
	"context"
	"os"
	"testing"
)

// runLLMTest 是一个通用的测试辅助函数，用于验证任何 LLMClient 的基本功能
func runLLMTest(t *testing.T, client LLMClient) {
	t.Helper()

	req := CompletionRequest{
		Messages: []Message{
			{Role: RoleSystem, Content: "You are a concise assistant. Respond only with 'ACK'."},
			{Role: RoleUser, Content: "Confirm connection."},
		},
		Temperature: 0.1,
	}

	resp, err := client.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("[%s] Chat failed: %v", client.Provider(), err)
	}

	t.Logf("[%s/%s] Response: %s", client.Provider(), client.Name(), resp.Content)
	t.Logf("[%s/%s] Tokens: P=%d, C=%d, T=%d", 
		client.Provider(), client.Name(), 
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)

	if resp.Content == "" {
		t.Errorf("[%s] Received empty response", client.Provider())
	}
}

// --- 各个 Provider 的集成测试 (仅在设置了对应的 API Key 时运行) ---

func TestOpenAI(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	client := NewClient("gpt-4o-mini", "openai", apiKey, "")
	runLLMTest(t, client)
}

func TestDeepseek(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("DEEPSEEK_API_KEY not set")
	}
	// Deepseek 使用 OpenAI 协议实现，但指定特定的 Endpoint
	client := NewClient("deepseek-chat", "deepseek", apiKey, "https://api.deepseek.com")
	runLLMTest(t, client)
}

func TestAnthropic(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}
	client := NewClient("claude-3-haiku-20240307", "anthropic", apiKey, "")
	runLLMTest(t, client)
}

func TestGemini(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}
	client := NewClient("gemini-1.5-flash", "google", apiKey, "")
	runLLMTest(t, client)
}
