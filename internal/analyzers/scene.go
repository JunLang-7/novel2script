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

// SceneAnalyzer 负责将小说文本分割为独立的戏剧场景。
type SceneAnalyzer struct {
	client      *llm.Client
	maxParallel int
}

// NewSceneAnalyzer 创建场景分割器。
func NewSceneAnalyzer(client *llm.Client, maxParallel int) *SceneAnalyzer {
	if maxParallel <= 0 {
		maxParallel = 5
	}
	return &SceneAnalyzer{client: client, maxParallel: maxParallel}
}

// llmScene LLM 返回的场景结构（临时类型）。
type llmScene struct {
	ID                string                     `json:"id"`
	Title             string                     `json:"title"`
	Sequence          int                        `json:"sequence"`
	Location          string                     `json:"location"`
	LocationType      string                     `json:"location_type"`
	TimeOfDay         string                     `json:"time_of_day"`
	Atmosphere        string                     `json:"atmosphere"`
	Summary           string                     `json:"summary"`
	ChapterSource     int                        `json:"chapter_source"`
	Mood              string                     `json:"mood"`
	CharactersPresent []models.CharacterPresence `json:"characters_present"`
}

// Analyze 对分块后的小说文本并行执行场景分割。
func (a *SceneAnalyzer) Analyze(ctx context.Context, rawText string) ([]models.Scene, error) {
	chapters, err := text.SplitChapters(rawText)
	if err != nil {
		return nil, fmt.Errorf("章节检测失败: %w", err)
	}

	chunks := text.GroupIntoChunks(chapters, 15000)

	type chunkResult struct {
		scenes []models.Scene
		err    error
	}
	results := make([]chunkResult, len(chunks))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(a.maxParallel)

	for i, chunk := range chunks {
		g.Go(func() error {
			scenes, err := a.analyzeChunk(ctx, chunk)
			results[i] = chunkResult{scenes: scenes, err: err}
			return nil
		})
	}
	g.Wait()

	var allScenes []models.Scene
	for i, r := range results {
		if r.err != nil {
			return nil, fmt.Errorf("场景分割失败(chunk %d): %w", i, r.err)
		}
		offset := len(allScenes)
		for j := range r.scenes {
			r.scenes[j].Sequence = offset + j + 1
		}
		allScenes = append(allScenes, r.scenes...)
	}

	return allScenes, nil
}

func (a *SceneAnalyzer) analyzeChunk(ctx context.Context, chunk text.Chunk) ([]models.Scene, error) {
	prompt := strings.Replace(llm.SceneSegmentationPrompt, "{text}", chunk.Text, 1)

	rawScenes, _, err := llm.StructuredGenerate[[]llmScene](ctx, a.client, llm.SystemPrompt, prompt)
	if err != nil {
		return nil, err
	}

	scenes := make([]models.Scene, len(rawScenes))
	for i, s := range rawScenes {
		scenes[i] = models.Scene{
			ID:    s.ID,
			Type:  models.SceneTypeScene,
			Title: s.Title,
			Setting: models.SceneSetting{
				Location:     s.Location,
				LocationType: s.LocationType,
				TimeOfDay:    s.TimeOfDay,
				Atmosphere:   s.Atmosphere,
			},
			Summary:           s.Summary,
			ChapterSource:     s.ChapterSource,
			Mood:              s.Mood,
			CharactersPresent: s.CharactersPresent,
		}
	}

	return scenes, nil
}
