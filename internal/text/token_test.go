package text

import "testing"

func TestFormatCharCount(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{0, "0字"},
		{500, "500字"},
		{999, "999字"},
		{1000, "1千字"},
		{1500, "1千字"},
		{9999, "9千字"},
		{10000, "1万字"},
		{50000, "5万字"},
		{12345, "1万字"},
		{100000, "10万字"},
	}

	for _, tt := range tests {
		got := FormatCharCount(tt.count)
		if got != tt.expected {
			t.Errorf("FormatCharCount(%d) = %q, want %q", tt.count, got, tt.expected)
		}
	}
}

func TestFormatCharCount_Zero(t *testing.T) {
	got := FormatCharCount(0)
	if got != "0字" {
		t.Errorf("expected %q, got %q", "0字", got)
	}
}

func TestFormatCharCount_Wan(t *testing.T) {
	got := FormatCharCount(20000)
	if got != "2万字" {
		t.Errorf("expected %q, got %q", "2万字", got)
	}
}

func TestEstimateTokens_AsianChars(t *testing.T) {
	// Additional edge cases not covered by existing tests
	count := EstimateTokens("修仙小说")
	if count <= 0 {
		t.Error("should estimate > 0 tokens for valid CJK text")
	}
}
