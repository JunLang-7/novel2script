package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Generate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{{Text: `{"name": "韩立", "role": "protagonist"}`}},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{InputTokens: 100, OutputTokens: 50},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{
		Provider:    "anthropic",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  0,
		MaxParallel: 1,
	})

	result, err := client.Generate(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result.Usage.InputTokens != 100 {
		t.Errorf("input tokens: %d", result.Usage.InputTokens)
	}
	if result.Usage.OutputTokens != 50 {
		t.Errorf("output tokens: %d", result.Usage.OutputTokens)
	}
}

func TestClient_Generate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Error: &struct {
				Message string `json:"message"`
			}{Message: "invalid request"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{
		Provider:    "anthropic",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  0,
		MaxParallel: 1,
	})

	_, err := client.Generate(context.Background(), "system", "user")
	if err == nil {
		t.Error("expected error for API error response")
	}
}

func TestClient_Generate_RateLimit(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{{Text: `{}`}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{
		Provider:    "anthropic",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  2,
		MaxParallel: 1,
	})

	result, err := client.Generate(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("Generate should succeed after retry: %v", err)
	}
	if result == nil {
		t.Error("result should not be nil")
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (1 fail + 1 success), got %d", callCount)
	}
}

func TestClient_Generate_ServerError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{{Text: `{}`}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{
		Provider:    "anthropic",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  3,
		MaxParallel: 1,
	})

	_, err := client.Generate(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("Generate should succeed after retries: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls (2 fail + 1 success), got %d", callCount)
	}
}

func TestClient_StructuredGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{{Text: `{"name": "韩立", "age": 16}`}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{
		Provider:    "anthropic",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  0,
		MaxParallel: 1,
	})

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	person, _, err := StructuredGenerate[Person](context.Background(), client, "system", "user")
	if err != nil {
		t.Fatalf("StructuredGenerate failed: %v", err)
	}
	if person.Name != "韩立" {
		t.Errorf("name: %s", person.Name)
	}
	if person.Age != 16 {
		t.Errorf("age: %d", person.Age)
	}
}

func TestClient_Generate_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{
		Provider:    "anthropic",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  0,
		MaxParallel: 1,
	})

	_, err := client.Generate(context.Background(), "system", "user")
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestClient_Generate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json at all"))
	}))
	defer server.Close()

	client := NewClient(Config{
		Provider:    "anthropic",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  1,
		MaxParallel: 1,
	})

	// Should retry and eventually fail
	_, err := client.Generate(context.Background(), "system", "user")
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestClient_OpenAIProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openaiResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: `{"result": "ok"}`}}},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			}{PromptTokens: 10, CompletionTokens: 5},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(Config{
		Provider:    "openai",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  0,
		MaxParallel: 1,
	})

	result, err := client.Generate(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result.RawJSON != `{"result": "ok"}` {
		t.Errorf("unexpected raw JSON: %s", result.RawJSON)
	}
}
