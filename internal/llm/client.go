package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// Client 封装 LLM API 调用，支持 Anthropic 和 OpenAI 兼容接口。
type Client struct {
	httpClient *http.Client
	cfg        Config
	semaphore  chan struct{}
}

// NewClient 创建 LLM 客户端。
func NewClient(cfg Config) *Client {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.MaxParallel <= 0 {
		cfg.MaxParallel = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 120 * time.Second
	}
	if cfg.BaseURL == "" {
		switch cfg.Provider {
		case "openai":
			cfg.BaseURL = "https://api.openai.com/v1/chat/completions"
		default:
			cfg.BaseURL = "https://api.anthropic.com/v1/messages"
		}
	}

	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		cfg:        cfg,
		semaphore:  make(chan struct{}, cfg.MaxParallel),
	}
}

// anthropicRequest 定义 Anthropic Messages API 的请求体。
type anthropicRequest struct {
	Model       string            `json:"model"`
	MaxTokens   int               `json:"max_tokens"`
	System      string            `json:"system"`
	Messages    []anthropicMessage `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// openaiRequest 定义 OpenAI Chat Completions API 的请求体。
type openaiRequest struct {
	Model       string         `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Temperature float64        `json:"temperature,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Generate 发送 prompt 到 LLM，返回原始 JSON 响应和 token 用量。
func (c *Client) Generate(ctx context.Context, sysPrompt, userPrompt string) (*Result, error) {
	// 获取并发信号量
	select {
	case c.semaphore <- struct{}{}:
		defer func() { <-c.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避
			backoff := min(time.Duration(1<<uint(attempt-1))*time.Second, 30*time.Second)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		result, err := c.doGenerate(ctx, sysPrompt, userPrompt, attempt)
		if err == nil {
			return result, nil
		}

		// 429 或其他可重试错误
		if isRetryable(err) {
			lastErr = err
			continue
		}
		return nil, err
	}

	return nil, fmt.Errorf("请求失败，已重试%d次: %w", c.cfg.MaxRetries, lastErr)
}

func (c *Client) doGenerate(ctx context.Context, sysPrompt, userPrompt string, retryCount int) (*Result, error) {
	switch c.cfg.Provider {
	case "openai":
		return c.callOpenAI(ctx, sysPrompt, userPrompt, retryCount)
	default:
		return c.callAnthropic(ctx, sysPrompt, userPrompt, retryCount)
	}
}

func (c *Client) callAnthropic(ctx context.Context, sysPrompt, userPrompt string, retryCount int) (*Result, error) {
	formatHint := ""
	if retryCount > 0 {
		formatHint = "\n\n请严格返回合法 JSON 格式，不要包含任何额外说明文字。"
	}

	req := anthropicRequest{
		Model:     c.cfg.Model,
		MaxTokens: 8192,
		System:    sysPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt + formatHint},
		},
		Temperature: 0.1,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.cfg.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode == 429 {
		return nil, &RetryableError{Message: "速率限制", StatusCode: 429}
	}
	if resp.StatusCode >= 500 {
		return nil, &RetryableError{Message: fmt.Sprintf("服务器错误(%d)", resp.StatusCode), StatusCode: resp.StatusCode}
	}

	var ar anthropicResponse
	if err := json.Unmarshal(respBody, &ar); err != nil {
		return nil, &RetryableError{Message: fmt.Sprintf("JSON解析失败: %v, 响应: %s", err, truncate(string(respBody), 200))}
	}

	if resp.StatusCode != 200 {
		if ar.Error != nil {
			return nil, fmt.Errorf("%s", ar.Error.Message)
		}
		return nil, fmt.Errorf("API返回错误(%d): %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	if ar.Error != nil {
		return nil, fmt.Errorf("API错误: %s", ar.Error.Message)
	}

	if len(ar.Content) == 0 {
		return nil, &RetryableError{Message: "响应内容为空"}
	}

	return &Result{
		RawJSON: extractJSON(ar.Content[0].Text),
		Usage: Usage{
			InputTokens:  ar.Usage.InputTokens,
			OutputTokens: ar.Usage.OutputTokens,
		},
	}, nil
}

func (c *Client) callOpenAI(ctx context.Context, sysPrompt, userPrompt string, retryCount int) (*Result, error) {
	formatHint := ""
	if retryCount > 0 {
		formatHint = "\n\n请严格返回合法 JSON 格式，不要包含任何额外说明文字。"
	}

	req := openaiRequest{
		Model: c.cfg.Model,
		Messages: []openaiMessage{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userPrompt + formatHint},
		},
		Temperature: 0.1,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode == 429 {
		return nil, &RetryableError{Message: "速率限制", StatusCode: 429}
	}
	if resp.StatusCode >= 500 {
		return nil, &RetryableError{Message: fmt.Sprintf("服务器错误(%d)", resp.StatusCode), StatusCode: resp.StatusCode}
	}

	var or openaiResponse
	if err := json.Unmarshal(respBody, &or); err != nil {
		return nil, &RetryableError{Message: fmt.Sprintf("JSON解析失败: %v, 响应: %s", err, truncate(string(respBody), 200))}
	}

	if resp.StatusCode != 200 {
		if or.Error != nil {
			return nil, fmt.Errorf("%s", or.Error.Message)
		}
		return nil, fmt.Errorf("API返回错误(%d): %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	if or.Error != nil {
		return nil, fmt.Errorf("API错误: %s", or.Error.Message)
	}

	if len(or.Choices) == 0 {
		return nil, &RetryableError{Message: "响应内容为空"}
	}

	return &Result{
		RawJSON: extractJSON(or.Choices[0].Message.Content),
		Usage: Usage{
			InputTokens:  or.Usage.PromptTokens,
			OutputTokens: or.Usage.CompletionTokens,
		},
	}, nil
}

// ParallelGenerate 并行调用 LLM，通过 errgroup 聚合错误。
func (c *Client) ParallelGenerate(ctx context.Context, prompts []PromptPair) ([]*Result, []error) {
	results := make([]*Result, len(prompts))
	errs := make([]error, len(prompts))

	g, ctx := errgroup.WithContext(ctx)

	for i, p := range prompts {
		g.Go(func() error {
			result, err := c.Generate(ctx, p.System, p.User)
			results[i] = result
			errs[i] = err
			// 不返回错误，让所有请求都能完成（部分失败不影响其他）
			return nil
		})
	}

	g.Wait()
	return results, errs
}

// StructuredGenerate 泛型方法：发送 prompt，将 JSON 响应解析到指定类型。
func StructuredGenerate[T any](ctx context.Context, c *Client, sysPrompt, userPrompt string) (T, error) {
	var zero T
	result, err := c.Generate(ctx, sysPrompt, userPrompt)
	if err != nil {
		return zero, fmt.Errorf("LLM调用失败: %w", err)
	}

	var parsed T
	if err := json.Unmarshal([]byte(result.RawJSON), &parsed); err != nil {
		return zero, fmt.Errorf("JSON解析失败: %w\n原始响应: %s", err, truncate(result.RawJSON, 500))
	}

	return parsed, nil
}

// RetryableError 表示可以重试的错误。
type RetryableError struct {
	Message    string
	StatusCode int
}

func (e *RetryableError) Error() string {
	return e.Message
}

func isRetryable(err error) bool {
	_, ok := err.(*RetryableError)
	return ok
}

// extractJSON 从 LLM 响应中提取 JSON 内容。
// 处理 LLM 可能包裹在 ```json ... ``` 或 ``` ... ``` 中返回的情况。
func extractJSON(text string) string {
	text = strings.TrimSpace(text)

	// 尝试提取 ```json ... ``` 包裹的内容
	if idx := strings.Index(text, "```json"); idx != -1 {
		start := idx + len("```json")
		if end := strings.Index(text[start:], "```"); end != -1 {
			return strings.TrimSpace(text[start : start+end])
		}
	}
	if idx := strings.Index(text, "```"); idx != -1 {
		start := idx + len("```")
		if end := strings.Index(text[start:], "```"); end != -1 {
			inner := strings.TrimSpace(text[start : start+end])
			if strings.HasPrefix(inner, "{") || strings.HasPrefix(inner, "[") {
				return inner
			}
		}
	}

	// 直接返回（假设就是纯 JSON）
	return text
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
