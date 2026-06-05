package text

import (
	"fmt"
	"regexp"
	"strings"
)

// Chapter 代表检测到的一个章节。
type Chapter struct {
	Number   int
	Title    string
	StartPos int // 在原文中的字节偏移
	EndPos   int
	Content  string
}

var chapterPatterns = []*regexp.Regexp{
	// 第X章 / 第X回 / 第X节 / 第X部
	regexp.MustCompile(`(?m)^第[零一二三四五六七八九十百千万\d]+[章回节部卷].*$`),
	// 第\d+章（最常见的中文网文格式）
	regexp.MustCompile(`(?m)^第\d+章\s+.*$`),
	// Chapter X
	regexp.MustCompile(`(?im)^Chapter\s+\d+.*$`),
	// Markdown 标题格式: # 1. xxx / ## 第一章
	regexp.MustCompile(`(?m)^#{1,3}\s+\d+[\.\s、].*$`),
	regexp.MustCompile(`(?m)^#{1,3}\s*第[零一二三四五六七八九十百千万\d]+[章回节部卷].*$`),
	// Volume / Vol.
	regexp.MustCompile(`(?im)^Vol(?:ume)?\.?\s*\d+.*$`),
}

// Chunk 代表一组连续的章节，作为一次 LLM 调用的输入单元。
type Chunk struct {
	ID           string
	ChapterStart int
	ChapterEnd   int
	Text         string
	CharCount    int
	TokenEst     int
}

// SplitChapters 检测章节边界，返回有序章节列表。
func SplitChapters(raw string) ([]Chapter, error) {
	if len(strings.TrimSpace(raw)) == 0 {
		return nil, fmt.Errorf("输入文本为空")
	}

	var boundaries []int
	for _, pat := range chapterPatterns {
		locs := pat.FindAllStringIndex(raw, -1)
		for _, loc := range locs {
			boundaries = append(boundaries, loc[0])
		}
	}

	if len(boundaries) == 0 {
		return fallbackSplit(raw), nil
	}

	boundaries = dedupAndSort(boundaries)

	var chapters []Chapter
	for i := 0; i < len(boundaries); i++ {
		start := boundaries[i]
		end := len(raw)
		if i+1 < len(boundaries) {
			end = boundaries[i+1]
		}

		ch := Chapter{
			Number:   i + 1,
			StartPos: start,
			EndPos:   end,
			Content:  strings.TrimSpace(raw[start:end]),
		}
		ch.Title = extractChapterTitle(ch.Content)
		chapters = append(chapters, ch)
	}

	return chapters, nil
}

// fallbackSplit 在无章节标记时，按段落数等分为大致相当的伪章节。
func fallbackSplit(raw string) []Chapter {
	paragraphs := strings.Split(raw, "\n\n")
	if len(paragraphs) < 3 {
		// 文本太短，当作一个章节
		return []Chapter{{Number: 1, StartPos: 0, EndPos: len(raw), Content: strings.TrimSpace(raw)}}
	}

	// 每约 100 个段落作为一个伪章节
	paraPerChapter := max(100, len(paragraphs)/3)
	var chapters []Chapter
	offset := 0
	for i := 0; i < len(paragraphs); i += paraPerChapter {
		end := min(i+paraPerChapter, len(paragraphs))
		content := strings.Join(paragraphs[i:end], "\n\n")
		chapters = append(chapters, Chapter{
			Number:   len(chapters) + 1,
			StartPos: offset,
			EndPos:   offset + len(content),
			Content:  strings.TrimSpace(content),
			Title:    fmt.Sprintf("第%d节", len(chapters)+1),
		})
		offset += len(content) + 2
	}
	return chapters
}

// GroupIntoChunks 将章节按 token 限制分组为 Chunk 批次。
func GroupIntoChunks(chapters []Chapter, tokensPerChunk int) []Chunk {
	if tokensPerChunk <= 0 {
		tokensPerChunk = 15000 // 默认：约5章 × 3000字/章 ÷ 1.5字/token
	}

	var chunks []Chunk
	chunkIdx := 0
	buf := new(strings.Builder)
	startChapter := 1
	currentTokens := 0

	for i, ch := range chapters {
		chTokens := EstimateTokens(ch.Content)
		// 如果加入当前章节会超出限制且 buffer 非空，则先闭合当前 chunk
		if currentTokens+chTokens > tokensPerChunk && buf.Len() > 0 {
			chunks = append(chunks, Chunk{
				ID:           fmt.Sprintf("chunk_%03d", chunkIdx),
				ChapterStart: startChapter,
				ChapterEnd:   chapters[i-1].Number,
				Text:         buf.String(),
				CharCount:    len([]rune(buf.String())),
				TokenEst:     currentTokens,
			})
			chunkIdx++
			buf.Reset()
			startChapter = ch.Number
			currentTokens = 0
		}

		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(ch.Content)
		currentTokens += chTokens
	}

	if buf.Len() > 0 {
		chunks = append(chunks, Chunk{
			ID:           fmt.Sprintf("chunk_%03d", chunkIdx),
			ChapterStart: startChapter,
			ChapterEnd:   chapters[len(chapters)-1].Number,
			Text:         buf.String(),
			CharCount:    len([]rune(buf.String())),
			TokenEst:     currentTokens,
		})
	}

	return chunks
}

// extractChapterTitle 从章节内容的第一行提取标题。
func extractChapterTitle(content string) string {
	lines := strings.SplitN(content, "\n", 2)
	if len(lines) == 0 {
		return ""
	}
	title := strings.TrimSpace(lines[0])
	if len(title) > 100 {
		title = title[:100]
	}
	return title
}

func dedupAndSort(boundaries []int) []int {
	seen := make(map[int]bool)
	var result []int
	for _, b := range boundaries {
		if !seen[b] {
			seen[b] = true
		}
	}
	// 保持出现顺序（即原始文本中的位置顺序）
	for _, b := range boundaries {
		if seen[b] {
			result = append(result, b)
			delete(seen, b)
		}
	}
	// 排序以确保顺序正确
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}
