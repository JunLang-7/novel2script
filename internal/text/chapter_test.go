package text

import (
	"strings"
	"testing"
)

func TestSplitChapters_StandardFormat(t *testing.T) {
	input := `第一章 山村少年

天南国，青牛镇。

清晨的薄雾还没散尽，韩立已经背着药篓从山上下来了。

第二章 黑风岭遇险

第二天凌晨，天还没亮，韩立揣着两个干饼、一把采药镰刀，悄悄出了门。

第三章 意外的机缘

韩立被老者拉上安全的地方后，才发现对方虽然是六七十岁的老人，力道却大得惊人。`

	chapters, err := SplitChapters(input)
	if err != nil {
		t.Fatalf("SplitChapters failed: %v", err)
	}

	if len(chapters) != 3 {
		t.Fatalf("expected 3 chapters, got %d", len(chapters))
	}

	if !strings.Contains(chapters[0].Content, "天南国") {
		t.Error("chapter 1 should contain '天南国'")
	}
	if !strings.Contains(chapters[1].Content, "黑风岭") {
		t.Error("chapter 2 should contain '黑风岭'")
	}
	if chapters[0].Number != 1 {
		t.Errorf("chapter 1 number = %d, want 1", chapters[0].Number)
	}
}

func TestSplitChapters_ChapterXFormat(t *testing.T) {
	input := `第1章 开始

故事从这里开始。

第2章 发展

故事继续发展。

第3章 结束

故事到此结束。`

	chapters, err := SplitChapters(input)
	if err != nil {
		t.Fatalf("SplitChapters failed: %v", err)
	}

	if len(chapters) != 3 {
		t.Fatalf("expected 3 chapters, got %d", len(chapters))
	}
}

func TestSplitChapters_NoChapterMarks(t *testing.T) {
	input := "这是一段没有任何章节标记的文本。\n\n" + strings.Repeat("内容填充文本。\n\n", 500)

	chapters, err := SplitChapters(input)
	if err != nil {
		t.Fatalf("SplitChapters failed: %v", err)
	}

	if len(chapters) == 0 {
		t.Fatal("expected at least 1 fallback chapter")
	}
}

func TestSplitChapters_EmptyInput(t *testing.T) {
	_, err := SplitChapters("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestGroupIntoChunks_BasicGrouping(t *testing.T) {
	input := `第一章 测试1

短文内容一。

第二章 测试2

短文内容二。

第三章 测试3

短文内容三。

第四章 测试4

短文内容四。

第五章 测试5

短文内容五。`

	chapters, err := SplitChapters(input)
	if err != nil {
		t.Fatalf("SplitChapters failed: %v", err)
	}

	chunks := GroupIntoChunks(chapters, 500)
	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}

	// 所有章节都应被覆盖
	totalCovered := 0
	for _, chunk := range chunks {
		totalCovered += chunk.ChapterEnd - chunk.ChapterStart + 1
	}
	if totalCovered != len(chapters) {
		t.Errorf("not all chapters covered: covered %d, total %d", totalCovered, len(chapters))
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"", 0},
		{"你好", 1},      // 2字 → ~1 token
		{"Hello", 3},     // 5字符 * 2/3 = 3
		{"你好世界", 2},   // 4字 * 2/3 = 2
	}

	for _, tt := range tests {
		got := EstimateTokens(tt.text)
		if got < tt.expected {
			t.Errorf("EstimateTokens(%q) = %d, want at least %d", tt.text, got, tt.expected)
		}
	}
}
