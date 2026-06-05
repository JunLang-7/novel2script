package config

import (
	"os"
	"strconv"
)

// Config 全局配置。
type Config struct {
	Provider   string
	BaseURL    string
	APIKey     string
	Model      string
	Parallel   int
	CacheDir   string
}

// Load 从环境变量和 CLI 参数加载配置。
func Load() Config {
	return Config{
		Provider: Env("NOVEL2SCRIPT_PROVIDER", "anthropic"),
		BaseURL:  Env("NOVEL2SCRIPT_BASE_URL", ""),
		APIKey:   Env("NOVEL2SCRIPT_API_KEY", ""),
		Model:    Env("NOVEL2SCRIPT_MODEL", "claude-sonnet-4-20250514"),
		Parallel: EnvInt("NOVEL2SCRIPT_PARALLEL", 5),
		CacheDir: Env("NOVEL2SCRIPT_CACHE_DIR", expandHome("~/.novel2script/cache")),
	}
}

// Env 获取环境变量，不存在则返回默认值。
func Env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

// EnvInt 获取整数型环境变量。
func EnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return n
		}
	}
	return def
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + path[1:]
		}
	}
	return path
}
