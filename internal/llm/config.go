package llm

import "time"

// Config 配置 LLM 客户端。
type Config struct {
	Provider   string // "anthropic" | "openai"
	BaseURL    string
	APIKey     string
	Model      string
	MaxRetries int // 默认 3
	MaxParallel int // 默认 5
	Timeout    time.Duration // 默认 120s
}

// DefaultConfig 返回默认配置（从环境变量读取）。
func DefaultConfig() Config {
	return Config{
		Provider:   envOrDefault("NOVEL2SCRIPT_PROVIDER", "anthropic"),
		BaseURL:    envOrDefault("NOVEL2SCRIPT_BASE_URL", ""),
		APIKey:     envOrDefault("NOVEL2SCRIPT_API_KEY", ""),
		Model:      envOrDefault("NOVEL2SCRIPT_MODEL", "claude-sonnet-4-20250514"),
		MaxRetries: 3,
		MaxParallel: 5,
		Timeout:    120 * time.Second,
	}
}

// PromptPair 表示一组 system + user prompt。
type PromptPair struct {
	System string
	User   string
}

// Usage 记录一次 LLM 调用的 token 使用量。
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// Result 封装 LLM 调用结果。
type Result struct {
	RawJSON string
	Usage   Usage
}

func envOrDefault(key, def string) string {
	return def // 实际读取在 config 包中通过 os.Getenv 完成
}

// TokenEstimator 估算文本的 token 数量。
type TokenEstimator interface {
	Estimate(text string) int
}

// SimpleEstimator 使用保守的中文 token 估算：1.5字 ≈ 1 token。
type SimpleEstimator struct{}

func (e SimpleEstimator) Estimate(text string) int {
	runes := len([]rune(text))
	return runes * 2 / 3
}

// Ensure SimpleEstimator satisfies TokenEstimator.
var _ TokenEstimator = SimpleEstimator{}
