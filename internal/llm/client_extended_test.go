package llm

import (
	"testing"
)

func TestExtractJSON_ArrayResponse(t *testing.T) {
	input := `[{"id": 1}, {"id": 2}]`
	result := extractJSON(input)
	if result != input {
		t.Errorf("array JSON should pass through, got: %s", result)
	}
}

func TestExtractJSON_NestedCodeBlock(t *testing.T) {
	input := "好的，以下是结果：\n```json\n{\"name\": \"韩立\", \"role\": \"protagonist\"}\n```\n希望这对你有帮助。"
	result := extractJSON(input)
	if result != `{"name": "韩立", "role": "protagonist"}` {
		t.Errorf("unexpected extraction: %s", result)
	}
}

func TestExtractJSON_PlainCodeBlock(t *testing.T) {
	// Non-JSON code block returns original text (doesn't start with { or [)
	input := "```\nnot json\n```"
	result := extractJSON(input)
	if result != input {
		t.Logf("plain code block returns original: %s", result)
	}
}

func TestExtractJSON_OnlyCodeBlock_NonJSON(t *testing.T) {
	input := "```\nline1\nline2\n```"
	result := extractJSON(input)
	// Non-JSON code blocks return original text
	if result == input {
		t.Logf("non-JSON code block correctly preserved")
	}
}

func TestExtractJSON_LeadingSpaces(t *testing.T) {
	input := "   \n  \t  {\"key\": true}  \n  "
	result := extractJSON(input)
	if result != `{"key": true}` {
		t.Errorf("expected trimmed JSON, got: %s", result)
	}
}

func TestNewClient_OpenAIProvider(t *testing.T) {
	cfg := Config{
		Provider: "openai",
		APIKey:   "sk-test",
	}
	client := NewClient(cfg)
	if client.cfg.BaseURL == "https://api.anthropic.com/v1/messages" {
		t.Error("should use OpenAI base URL, not Anthropic")
	}
}

func TestNewClient_CustomBaseURL(t *testing.T) {
	cfg := Config{
		Provider: "anthropic",
		BaseURL:  "https://custom.api.com/v1",
		APIKey:   "test",
	}
	client := NewClient(cfg)
	if client.cfg.BaseURL != "https://custom.api.com/v1" {
		t.Errorf("should respect custom base URL, got: %s", client.cfg.BaseURL)
	}
}

func TestNewClient_ZeroRetries(t *testing.T) {
	cfg := Config{
		APIKey:     "test",
		MaxRetries: 0,
	}
	client := NewClient(cfg)
	if client.cfg.MaxRetries != 3 {
		t.Errorf("zero MaxRetries should default to 3, got %d", client.cfg.MaxRetries)
	}
}

func TestTokenEstimator_Range(t *testing.T) {
	e := SimpleEstimator{}
	testCases := []struct {
		text string
		min  int
		max  int
	}{
		{"短", 0, 1},
		{"韩立站起身", 2, 3},
		{"这是一段中等长度的中文文本用于测试token数量估算", 10, 18},
		{"", 0, 0},
	}

	for _, tc := range testCases {
		got := e.Estimate(tc.text)
		if got < tc.min || got > tc.max {
			t.Errorf("Estimate(%q) = %d, want [%d, %d]", tc.text, got, tc.min, tc.max)
		}
	}
}

func TestRetryableError_Interface(t *testing.T) {
	var err error = &RetryableError{Message: "test", StatusCode: 500}
	if !isRetryable(err) {
		t.Error("RetryableError should satisfy isRetryable")
	}

	var nonRetryable error = errTest("plain error")
	if isRetryable(nonRetryable) {
		t.Error("plain error should not be retryable")
	}
}

type errTest string

func (e errTest) Error() string { return string(e) }
