package text

import (
	"os"
	"testing"
)

func TestDetectAndReadFile_UTF8Content(t *testing.T) {
	// Create a temp UTF-8 file
	content := "测试中文内容 Hello World"
	tmpFile, err := os.CreateTemp("", "test_utf8_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("cannot write temp file: %v", err)
	}
	tmpFile.Close()

	read, err := DetectAndReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("DetectAndReadFile failed: %v", err)
	}
	if read != content {
		t.Errorf("content mismatch: got %q, want %q", read, content)
	}
}

func TestEstimateTokens_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		min, max int
	}{
		{"empty", "", 0, 0},
		{"single_ascii", "a", 0, 1},
		{"single_cjk", "仙", 0, 1},
		{"mixed", "hello 世界", 5, 10},
		{"long_cjk", "修仙之路漫漫无期唯有坚持方能成功", 10, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.text)
			if got < tt.min || got > tt.max {
				t.Errorf("EstimateTokens(%q) = %d, want [%d, %d]", tt.text, got, tt.min, tt.max)
			}
		})
	}
}

func TestSplitChapters_EmptyContent(t *testing.T) {
	text := "第一章 空\n\n\n第二章 也空\n\n"
	chapters, err := SplitChapters(text)
	if err != nil {
		t.Fatalf("SplitChapters failed: %v", err)
	}
	if len(chapters) != 2 {
		t.Errorf("expected 2 chapters, got %d", len(chapters))
	}
}

func TestSplitChapters_ChineseNumberFormat(t *testing.T) {
	text := "第一章 一\n内容一\n第二章 二\n内容二"
	chapters, err := SplitChapters(text)
	if err != nil {
		t.Fatalf("SplitChapters failed: %v", err)
	}
	if len(chapters) != 2 {
		t.Errorf("expected 2 chapters ('第X章' format), got %d", len(chapters))
	}
}

func TestSplitChapters_ChapterPart(t *testing.T) {
	text := "第一部 初入修仙\n内容中...\n第二部 元婴大成\n内容下..."
	chapters, err := SplitChapters(text)
	if err != nil {
		t.Fatalf("SplitChapters failed: %v", err)
	}
	if len(chapters) != 2 {
		t.Errorf("expected 2 parts ('第X部' format), got %d", len(chapters))
	}
}

func TestSplitChapters_ChapterHui(t *testing.T) {
	text := "第一回 开篇\n内容...\n第二回 发展\n内容..."
	chapters, err := SplitChapters(text)
	if err != nil {
		t.Fatalf("SplitChapters failed: %v", err)
	}
	if len(chapters) != 2 {
		t.Errorf("expected 2 chapters ('第X回' format), got %d", len(chapters))
	}
}
