package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// Load 从 .env 文件和环境变量加载配置。
// .env 中的值不会覆盖已设置的环境变量。
func Load() Config {
	loadDotEnv()

	return Config{
		Provider: Env("NOVEL2SCRIPT_PROVIDER", "anthropic"),
		BaseURL:  Env("NOVEL2SCRIPT_BASE_URL", ""),
		APIKey:   Env("NOVEL2SCRIPT_API_KEY", ""),
		Model:    Env("NOVEL2SCRIPT_MODEL", "claude-sonnet-4-20250514"),
		Parallel: EnvInt("NOVEL2SCRIPT_PARALLEL", 5),
		CacheDir: Env("NOVEL2SCRIPT_CACHE_DIR", expandHome("~/.novel2script/cache")),
	}
}

func loadDotEnv() {
	// 从当前工作目录向上查找 .env 文件（最多向上 3 级）
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for range 3 {
		envPath := filepath.Join(dir, ".env")
		if f, err := os.Open(envPath); err == nil {
			parseEnvFile(f)
			f.Close()
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}

func parseEnvFile(f *os.File) {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		// 移除行内注释（# 前必须有空格或是在值之后）
		if idx := strings.IndexByte(line, '#'); idx > 0 {
			line = strings.TrimSpace(line[:idx])
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// 去掉引号
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		// 只在环境变量未设置时才写入（显式环境变量优先）
		if _, ok := os.LookupEnv(key); !ok && key != "" {
			os.Setenv(key, val)
		}
	}
	_ = scanner.Err()
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
