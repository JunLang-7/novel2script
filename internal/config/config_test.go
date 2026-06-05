package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	oldKey := os.Getenv("NOVEL2SCRIPT_API_KEY")
	oldModel := os.Getenv("NOVEL2SCRIPT_MODEL")
	oldProvider := os.Getenv("NOVEL2SCRIPT_PROVIDER")
	oldParallel := os.Getenv("NOVEL2SCRIPT_PARALLEL")
	defer func() {
		restoreEnv("NOVEL2SCRIPT_API_KEY", oldKey)
		restoreEnv("NOVEL2SCRIPT_MODEL", oldModel)
		restoreEnv("NOVEL2SCRIPT_PROVIDER", oldProvider)
		restoreEnv("NOVEL2SCRIPT_PARALLEL", oldParallel)
	}()

	// Set to empty instead of Unsetenv to prevent .env from re-populating
	os.Setenv("NOVEL2SCRIPT_API_KEY", "")
	os.Setenv("NOVEL2SCRIPT_MODEL", "")
	os.Setenv("NOVEL2SCRIPT_PROVIDER", "")
	os.Setenv("NOVEL2SCRIPT_PARALLEL", "")

	cfg := Load()
	if cfg.Provider != "anthropic" {
		t.Errorf("default provider: %q", cfg.Provider)
	}
	if cfg.Model != "claude-sonnet-4-20250514" {
		t.Errorf("default model: %q", cfg.Model)
	}
	if cfg.APIKey != "" {
		t.Errorf("default API key should be empty: %q", cfg.APIKey)
	}
	if cfg.Parallel != 5 {
		t.Errorf("default parallel: %d", cfg.Parallel)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("NOVEL2SCRIPT_API_KEY", "sk-test-key")
	os.Setenv("NOVEL2SCRIPT_MODEL", "gpt-4")
	os.Setenv("NOVEL2SCRIPT_PROVIDER", "openai")
	os.Setenv("NOVEL2SCRIPT_PARALLEL", "10")
	defer func() {
		os.Unsetenv("NOVEL2SCRIPT_API_KEY")
		os.Unsetenv("NOVEL2SCRIPT_MODEL")
		os.Unsetenv("NOVEL2SCRIPT_PROVIDER")
		os.Unsetenv("NOVEL2SCRIPT_PARALLEL")
	}()

	cfg := Load()
	if cfg.APIKey != "sk-test-key" {
		t.Errorf("APIKey: %q", cfg.APIKey)
	}
	if cfg.Model != "gpt-4" {
		t.Errorf("Model: %q", cfg.Model)
	}
	if cfg.Provider != "openai" {
		t.Errorf("Provider: %q", cfg.Provider)
	}
	if cfg.Parallel != 10 {
		t.Errorf("Parallel: %d", cfg.Parallel)
	}
}

func TestEnv_ReturnsDefault(t *testing.T) {
	val := Env("NONEXISTENT_ENV_VAR_12345", "default-value")
	if val != "default-value" {
		t.Errorf("expected default, got %q", val)
	}
}

func TestEnv_ReturnsEnv(t *testing.T) {
	os.Setenv("NOVEL2SCRIPT_TEST_VAR", "test-value")
	defer os.Unsetenv("NOVEL2SCRIPT_TEST_VAR")

	val := Env("NOVEL2SCRIPT_TEST_VAR", "default")
	if val != "test-value" {
		t.Errorf("expected test-value, got %q", val)
	}
}

func TestEnv_EmptyStringUsesDefault(t *testing.T) {
	os.Setenv("NOVEL2SCRIPT_TEST_EMPTY", "")
	defer os.Unsetenv("NOVEL2SCRIPT_TEST_EMPTY")

	val := Env("NOVEL2SCRIPT_TEST_EMPTY", "fallback")
	if val != "fallback" {
		t.Errorf("expected fallback for empty env, got %q", val)
	}
}

func TestEnvInt_Default(t *testing.T) {
	val := EnvInt("NONEXISTENT_ENV_INT_12345", 42)
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestEnvInt_FromEnv(t *testing.T) {
	os.Setenv("NOVEL2SCRIPT_TEST_INT", "99")
	defer os.Unsetenv("NOVEL2SCRIPT_TEST_INT")

	val := EnvInt("NOVEL2SCRIPT_TEST_INT", 1)
	if val != 99 {
		t.Errorf("expected 99, got %d", val)
	}
}

func TestEnvInt_InvalidUsesDefault(t *testing.T) {
	os.Setenv("NOVEL2SCRIPT_TEST_BAD_INT", "not-a-number")
	defer os.Unsetenv("NOVEL2SCRIPT_TEST_BAD_INT")

	val := EnvInt("NOVEL2SCRIPT_TEST_BAD_INT", 7)
	if val != 7 {
		t.Errorf("expected default 7, got %d", val)
	}
}

func TestEnvInt_ZeroUsesDefault(t *testing.T) {
	os.Setenv("NOVEL2SCRIPT_TEST_ZERO", "0")
	defer os.Unsetenv("NOVEL2SCRIPT_TEST_ZERO")

	val := EnvInt("NOVEL2SCRIPT_TEST_ZERO", 10)
	if val != 10 {
		t.Errorf("expected default 10 for zero, got %d", val)
	}
}

func TestEnvInt_NegativeUsesDefault(t *testing.T) {
	os.Setenv("NOVEL2SCRIPT_TEST_NEG", "-5")
	defer os.Unsetenv("NOVEL2SCRIPT_TEST_NEG")

	val := EnvInt("NOVEL2SCRIPT_TEST_NEG", 10)
	if val != 10 {
		t.Errorf("expected default 10 for negative, got %d", val)
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	result := expandHome("~/")
	if result != home+"/" {
		t.Errorf("expandHome: %q != %q", result, home+"/")
	}
}

func TestExpandHome_NoPrefix(t *testing.T) {
	result := expandHome("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("expected unchanged, got %q", result)
	}
}

func TestExpandHome_Empty(t *testing.T) {
	result := expandHome("")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestParseEnvFile_Basic(t *testing.T) {
	content := "KEY1=value1\nKEY2=hello world\n"
	f := tempEnvFile(t, content)
	defer f.Close()

	parseEnvFile(f)

	if v := os.Getenv("KEY1"); v != "value1" {
		t.Errorf("KEY1 = %q, want %q", v, "value1")
	}
	if v := os.Getenv("KEY2"); v != "hello world" {
		t.Errorf("KEY2 = %q, want %q", v, "hello world")
	}
	os.Unsetenv("KEY1")
	os.Unsetenv("KEY2")
}

func TestParseEnvFile_QuotedValues(t *testing.T) {
	content := `DOUBLE="quoted value"
SINGLE='single quoted'
`
	f := tempEnvFile(t, content)
	defer f.Close()

	parseEnvFile(f)

	if v := os.Getenv("DOUBLE"); v != "quoted value" {
		t.Errorf("DOUBLE = %q, want %q", v, "quoted value")
	}
	if v := os.Getenv("SINGLE"); v != "single quoted" {
		t.Errorf("SINGLE = %q, want %q", v, "single quoted")
	}
	os.Unsetenv("DOUBLE")
	os.Unsetenv("SINGLE")
}

func TestParseEnvFile_CommentsAndBlankLines(t *testing.T) {
	content := `# 这是注释行

# 这是另一个注释
REAL_KEY=real_value
# 行内注释
`
	f := tempEnvFile(t, content)
	defer f.Close()

	parseEnvFile(f)

	if v := os.Getenv("REAL_KEY"); v != "real_value" {
		t.Errorf("REAL_KEY = %q, want %q", v, "real_value")
	}
	os.Unsetenv("REAL_KEY")
}

func TestParseEnvFile_InlineComment(t *testing.T) {
	content := "KEY=value # 这是行内注释\n"
	f := tempEnvFile(t, content)
	defer f.Close()

	parseEnvFile(f)

	if v := os.Getenv("KEY"); v != "value" {
		t.Errorf("KEY = %q, want %q", v, "value")
	}
	os.Unsetenv("KEY")
}

func TestParseEnvFile_ExistingEnvWins(t *testing.T) {
	os.Setenv("EXISTING_KEY", "already_set")
	defer os.Unsetenv("EXISTING_KEY")

	content := "EXISTING_KEY=from_file\n"
	f := tempEnvFile(t, content)
	defer f.Close()

	parseEnvFile(f)

	if v := os.Getenv("EXISTING_KEY"); v != "already_set" {
		t.Errorf("existing env should not be overridden: got %q", v)
	}
}

func TestLoad_WithDotEnv(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	// Clear relevant env vars
	oldVars := saveEnvVars("NOVEL2SCRIPT_API_KEY", "NOVEL2SCRIPT_PROVIDER", "NOVEL2SCRIPT_PARALLEL")
	defer restoreEnvVars(oldVars)

	// Write .env file in temp directory
	envContent := "NOVEL2SCRIPT_API_KEY=sk-dotenv-key\nNOVEL2SCRIPT_PROVIDER=openai\nNOVEL2SCRIPT_PARALLEL=8\n"
	os.WriteFile(".env", []byte(envContent), 0644)

	cfg := Load()
	if cfg.APIKey != "sk-dotenv-key" {
		t.Errorf("expected sk-dotenv-key from .env, got %q", cfg.APIKey)
	}
	if cfg.Provider != "openai" {
		t.Errorf("expected openai from .env, got %q", cfg.Provider)
	}
	if cfg.Parallel != 8 {
		t.Errorf("expected parallel 8 from .env, got %d", cfg.Parallel)
	}
}

func TestLoad_ExistingEnvOverridesDotEnv(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	os.Setenv("NOVEL2SCRIPT_API_KEY", "sk-from-env")
	defer os.Unsetenv("NOVEL2SCRIPT_API_KEY")

	os.WriteFile(".env", []byte("NOVEL2SCRIPT_API_KEY=sk-from-file\n"), 0644)

	cfg := Load()
	if cfg.APIKey != "sk-from-env" {
		t.Errorf("env var should override .env: got %q", cfg.APIKey)
	}
}

func saveEnvVars(keys ...string) map[string]string {
	vars := make(map[string]string)
	for _, k := range keys {
		vars[k] = os.Getenv(k)
		os.Unsetenv(k)
	}
	return vars
}

func restoreEnvVars(vars map[string]string) {
	for k, v := range vars {
		if v != "" {
			os.Setenv(k, v)
		} else {
			os.Unsetenv(k)
		}
	}
}

func tempEnvFile(t *testing.T, content string) *os.File {
	t.Helper()
	f, err := os.CreateTemp("", "env_test_*.env")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("seek temp file: %v", err)
	}
	return f
}

func restoreEnv(key, val string) {
	if val != "" {
		os.Setenv(key, val)
	} else {
		os.Unsetenv(key)
	}
}
