package analyzers

import (
	"context"
	"fmt"
	"strings"

	"github.com/JunLang-7/novel2script/internal/llm"
	"github.com/JunLang-7/novel2script/internal/models"
	"github.com/JunLang-7/novel2script/internal/text"

	"golang.org/x/sync/errgroup"
)

// CharacterAnalyzer 负责从小说文本中提取角色信息。
type CharacterAnalyzer struct {
	client  *llm.Client
	maxParallel int
}

// NewCharacterAnalyzer 创建角色提取器。
func NewCharacterAnalyzer(client *llm.Client, maxParallel int) *CharacterAnalyzer {
	if maxParallel <= 0 {
		maxParallel = 3
	}
	return &CharacterAnalyzer{client: client, maxParallel: maxParallel}
}

// Analyze 执行多级渐进式角色提取。
//
// 策略：
//   Pass 1 — 前3章精细提取（核心角色列表）
//   Pass 2 — 后续每50章批量提取新角色（并行）
//   Pass 3 — 合并去重
func (a *CharacterAnalyzer) Analyze(ctx context.Context, rawText string) ([]models.Character, error) {
	chapters, err := text.SplitChapters(rawText)
	if err != nil {
		return nil, fmt.Errorf("章节检测失败: %w", err)
	}

	if len(chapters) == 0 {
		return nil, nil
	}

	merger := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
	}

	// Pass 1: 前3章精细提取
	pass1End := min(3, len(chapters))
	pass1Text := joinChapters(chapters[:pass1End])
	if pass1Text != "" {
		prompt := strings.Replace(llm.CharacterExtractionPrompt, "{text}", pass1Text, 1)
		chars, _, err := llm.StructuredGenerate[[]models.Character](ctx, a.client, llm.SystemPrompt, prompt)
		if err != nil {
			return nil, fmt.Errorf("Pass1角色提取失败: %w", err)
		}
		merger.add(chars)
	}

	// Pass 2: 后续每50章并行提取（仅对大型小说）
	if len(chapters) > 50 {
		batchSize := 50
		var batches [][]models.Character
		var batchTexts []string

		for start := 50; start < len(chapters); start += batchSize {
			end := min(start+batchSize, len(chapters))
			batchTexts = append(batchTexts, joinChapters(chapters[start:end]))
		}

		batches = make([][]models.Character, len(batchTexts))

		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(a.maxParallel)

		for i, bt := range batchTexts {
			g.Go(func() error {
				prompt := strings.Replace(llm.CharacterExtractionPrompt, "{text}", bt, 1)
				chars, _, err := llm.StructuredGenerate[[]models.Character](ctx, a.client, llm.SystemPrompt, prompt)
				if err == nil {
					batches[i] = chars
				}
				return nil // 单个批次失败不影响整体
			})
		}
		g.Wait()

		for _, batch := range batches {
			merger.add(batch)
		}
	}

	result := merger.result()
	// 重排序号
	for i := range result {
		result[i].ImportanceRank = i + 1
	}
	return result, nil
}

// characterMerger 角色去重合并器。
type characterMerger struct {
	nameIndex  map[string]int
	aliasIndex map[string]int
	chars      []models.Character
}

func (m *characterMerger) add(characters []models.Character) {
	for _, ch := range characters {
		idx, exists := m.findExisting(&ch)
		if exists {
			m.mergeInto(idx, &ch)
			m.indexCharacter(idx, &ch)
		} else {
			idx = len(m.chars)
			m.chars = append(m.chars, ch)
			m.indexCharacter(idx, &ch)
		}
	}
}

func (m *characterMerger) result() []models.Character {
	return m.chars
}

func (m *characterMerger) findExisting(ch *models.Character) (int, bool) {
	if idx, ok := m.nameIndex[ch.Name]; ok {
		return idx, true
	}
	for _, alias := range ch.Aliases {
		if idx, ok := m.aliasIndex[alias]; ok {
			return idx, true
		}
	}
	if idx, ok := m.aliasIndex[ch.Name]; ok {
		return idx, true
	}
	for _, alias := range ch.Aliases {
		if idx, ok := m.nameIndex[alias]; ok {
			return idx, true
		}
	}
	return -1, false
}

func (m *characterMerger) mergeInto(idx int, ch *models.Character) {
	existing := &m.chars[idx]

	aliasSet := make(map[string]bool)
	for _, a := range existing.Aliases {
		aliasSet[a] = true
	}
	for _, a := range ch.Aliases {
		if !aliasSet[a] {
			existing.Aliases = append(existing.Aliases, a)
			aliasSet[a] = true
		}
	}

	traitSet := make(map[string]bool)
	for _, t := range existing.Traits {
		traitSet[t] = true
	}
	for _, t := range ch.Traits {
		if !traitSet[t] {
			existing.Traits = append(existing.Traits, t)
			traitSet[t] = true
		}
	}

	relSet := make(map[string]bool)
	for _, r := range existing.Relationships {
		relSet[r.TargetID] = true
	}
	for _, r := range ch.Relationships {
		if !relSet[r.TargetID] {
			existing.Relationships = append(existing.Relationships, r)
		}
	}

	if len(ch.Description) > len(existing.Description) {
		existing.Description = ch.Description
	}
	if len(ch.CharacterArc) > len(existing.CharacterArc) {
		existing.CharacterArc = ch.CharacterArc
	}
	if ch.FirstAppearanceChapter > 0 &&
		(existing.FirstAppearanceChapter == 0 || ch.FirstAppearanceChapter < existing.FirstAppearanceChapter) {
		existing.FirstAppearanceChapter = ch.FirstAppearanceChapter
	}
}

func (m *characterMerger) indexCharacter(idx int, ch *models.Character) {
	m.nameIndex[ch.Name] = idx
	for _, alias := range ch.Aliases {
		if _, ok := m.aliasIndex[alias]; !ok {
			m.aliasIndex[alias] = idx
		}
	}
}

func joinChapters(chapters []text.Chapter) string {
	parts := make([]string, len(chapters))
	for i, ch := range chapters {
		parts[i] = ch.Content
	}
	return strings.Join(parts, "\n\n")
}
