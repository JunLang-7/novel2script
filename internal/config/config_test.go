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

	os.Unsetenv("NOVEL2SCRIPT_API_KEY")
	os.Unsetenv("NOVEL2SCRIPT_MODEL")
	os.Unsetenv("NOVEL2SCRIPT_PROVIDER")
	os.Unsetenv("NOVEL2SCRIPT_PARALLEL")

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

func restoreEnv(key, val string) {
	if val != "" {
		os.Setenv(key, val)
	} else {
		os.Unsetenv(key)
	}
}
