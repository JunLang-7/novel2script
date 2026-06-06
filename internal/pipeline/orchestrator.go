package pipeline

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JunLang-7/novel2script/internal/llm"
	"github.com/JunLang-7/novel2script/internal/models"
	"github.com/JunLang-7/novel2script/internal/text"
)

// Orchestrator 协调整个小说转剧本的管道流程。
type Orchestrator struct {
	client *llm.Client
	cfg    OrchestratorConfig
}

// OrchestratorConfig 配置管道行为。
type OrchestratorConfig struct {
	TokensPerChunk int
	Parallelism    int
	Verbose        bool
}

// DefaultOrchestratorConfig 返回默认配置。
func DefaultOrchestratorConfig() OrchestratorConfig {
	return OrchestratorConfig{
		TokensPerChunk: 15000,
		Parallelism:    5,
		Verbose:        false,
	}
}

// NewOrchestrator 创建管道协调器。
func NewOrchestrator(client *llm.Client, cfg OrchestratorConfig) *Orchestrator {
	return &Orchestrator{client: client, cfg: cfg}
}

// PipelineStats 记录管道执行的统计信息。
type PipelineStats struct {
	TotalChapters     int
	TotalChars        int
	NumChunks         int
	NumLLMCalls       int
	TotalInputTokens  int
	TotalOutputTokens int
	Duration          time.Duration
}

// Run 执行完整的转换管道。
func (o *Orchestrator) Run(ctx context.Context, rawText string) (*models.Script, *PipelineStats, error) {
	start := time.Now()
	stats := &PipelineStats{}

	// Step 1: 章节检测与分块
	o.log("检测章节边界...")
	chapters, err := text.SplitChapters(rawText)
	if err != nil {
		return nil, nil, fmt.Errorf("章节检测失败: %w", err)
	}
	stats.TotalChapters = len(chapters)
	stats.TotalChars = len([]rune(rawText))
	o.log("检测到 %d 个章节，总计约 %s", stats.TotalChapters, text.FormatCharCount(stats.TotalChars))

	if len(chapters) < 3 {
		o.warn("警告: 检测到的章节数不足3个，改编效果可能不理想")
	}

	chunks := text.GroupIntoChunks(chapters, o.cfg.TokensPerChunk)
	stats.NumChunks = len(chunks)
	o.log("分为 %d 个处理批次", len(chunks))

	// Step 2: 角色提取
	o.log("提取角色信息...")
	characters, charUsage, err := o.extractCharacters(ctx, chunks)
	if err != nil {
		o.warn("警告: 角色提取失败: %v", err)
	}
	stats.NumLLMCalls++
	stats.TotalInputTokens += charUsage.InputTokens
	stats.TotalOutputTokens += charUsage.OutputTokens
	o.log("提取到 %d 个角色", len(characters))

	// 后处理：补全关系中的 target_id
	fillTargetIDs(characters)

	// Step 2.5: 提取元数据（作者、类型、梗概）
	o.log("提取元数据...")
	sourceAuthor, genre, synopsis, metaUsage, err := o.extractMetadata(ctx, chunks)
	if err != nil {
		o.warn("警告: 元数据提取失败: %v", err)
	}
	stats.NumLLMCalls++
	stats.TotalInputTokens += metaUsage.InputTokens
	stats.TotalOutputTokens += metaUsage.OutputTokens
	o.log("元数据: 作者=%s, 类型=%v", sourceAuthor, genre)

	// Step 3: 场景分割
	o.log("分割场景...")
	sceneMerger := NewSceneMerger()
	for _, chunk := range chunks {
		scenes, usage, err := o.analyzeScenes(ctx, chunk, characters)
		if err != nil {
			o.warn("警告: 场景分割失败(chunk %s): %v", chunk.ID, err)
			continue
		}
		sceneMerger.Add(scenes)
		stats.NumLLMCalls++
		stats.TotalInputTokens += usage.InputTokens
		stats.TotalOutputTokens += usage.OutputTokens
	}
	scenes := sceneMerger.Result()
	o.log("分割出 %d 个场景", len(scenes))

	// Step 4: 剧本转换
	o.log("转换剧本元素...")
	for i := range scenes {
		elements, usage, err := o.convertScene(ctx, &scenes[i], characters)
		if err != nil {
			o.warn("警告: 场景转换失败(%s): %v", scenes[i].ID, err)
			continue
		}
		scenes[i].Elements = elements
		stats.NumLLMCalls++
		stats.TotalInputTokens += usage.InputTokens
		stats.TotalOutputTokens += usage.OutputTokens
	}
	o.log("转换完成")

	// Step 5: 构建分幕结构
	o.log("构建分幕结构...")
	actBuilder := &ActBuilder{}
	acts := actBuilder.Build(scenes, nil)

	// Step 6: 组装 Script
	script := &models.Script{
		ScriptTitle:  detectScriptTitle(chapters),
		SourceNovel:  "",
		SourceAuthor: sourceAuthor,
		Adaptor:      "novel2script v0.1.0",
		GeneratedAt:  time.Now(),
		Version:      "0.1.0",
		Metadata: models.Metadata{
			Genre:                genre,
			OriginalLanguage:     "zh-CN",
			TargetFormat:         "screenplay",
			TotalNovelChapters:   stats.TotalChapters,
			TotalNovelChars:      stats.TotalChars,
			AdaptationCoverage:   fmt.Sprintf("1-%d章", stats.TotalChapters),
			Synopsis:             synopsis,
			EstimatedTotalScenes: len(scenes),
		},
		Characters: characters,
		Acts:       acts,
	}

	stats.Duration = time.Since(start)
	o.log("管道完成，耗时 %v", stats.Duration.Round(time.Millisecond))

	return script, stats, nil
}

// extractCharacters 执行多级渐进式角色提取。
func (o *Orchestrator) extractCharacters(ctx context.Context, chunks []text.Chunk) ([]models.Character, llm.Usage, error) {
	merger := NewCharacterMerger()
	totalUsage := llm.Usage{}

	// Pass 1: 前3章精细提取
	if len(chunks) > 0 {
		firstChunk := chunks[0]
		// 只取前3章的内容
		pass1Text := firstChunk.Text
		prompt := strings.Replace(llm.CharacterExtractionPrompt, "{text}", pass1Text, 1)

		characters, err := llm.StructuredGenerate[[]models.Character](ctx, o.client, llm.SystemPrompt, prompt)
		if err == nil {
			merger.Add(characters)
		} else {
			return nil, totalUsage, fmt.Errorf("Pass1角色提取失败: %w", err)
		}
	}

	// Pass 2: 后续每50章批量提取新角色（此处简化为每个chunk）
	// 实际上对于小于50章的小说，Pass1已经足够
	if len(chunks) > 1 {
		for _, chunk := range chunks[1:] {
			prompt := strings.Replace(llm.CharacterExtractionPrompt, "{text}", chunk.Text, 1)
			characters, err := llm.StructuredGenerate[[]models.Character](ctx, o.client, llm.SystemPrompt, prompt)
			if err == nil {
				merger.Add(characters)
			}
		}
	}

	return merger.Result(), totalUsage, nil
}

// extractMetadata 从小说开头提取元数据（作者、类型、梗概）。
func (o *Orchestrator) extractMetadata(ctx context.Context, chunks []text.Chunk) (sourceAuthor string, genre []string, synopsis string, usage llm.Usage, err error) {
	if len(chunks) == 0 {
		return "", nil, "", llm.Usage{}, nil
	}

	text := chunks[0].Text
	prompt := strings.Replace(llm.MetadataExtractionPrompt, "{text}", text, 1)

	type llmMetadata struct {
		SourceAuthor string   `json:"source_author"`
		Genre        []string `json:"genre"`
		Synopsis     string   `json:"synopsis"`
	}

	meta, err := llm.StructuredGenerate[llmMetadata](ctx, o.client, llm.SystemPrompt, prompt)
	if err != nil {
		return "", nil, "", llm.Usage{}, err
	}

	return meta.SourceAuthor, meta.Genre, meta.Synopsis, llm.Usage{}, nil
}

// analyzeScenes 对一个chunk进行场景分割。
func (o *Orchestrator) analyzeScenes(ctx context.Context, chunk text.Chunk, characters []models.Character) ([]models.Scene, llm.Usage, error) {
	prompt := strings.Replace(llm.SceneSegmentationPrompt, "{character_context}", buildCharacterContext(characters), 1)
	prompt = strings.Replace(prompt, "{text}", chunk.Text, 1)

	type llmScene struct {
		ID                string                      `json:"id"`
		Title             string                      `json:"title"`
		Sequence          int                         `json:"sequence"`
		Location          string                      `json:"location"`
		LocationType      string                      `json:"location_type"`
		TimeOfDay         string                      `json:"time_of_day"`
		Atmosphere        string                      `json:"atmosphere"`
		Summary           string                      `json:"summary"`
		ChapterSource     int                         `json:"chapter_source"`
		Mood              string                      `json:"mood"`
		CharactersPresent []models.CharacterPresence  `json:"characters_present"`
	}

	rawScenes, err := llm.StructuredGenerate[[]llmScene](ctx, o.client, llm.SystemPrompt, prompt)
	if err != nil {
		return nil, llm.Usage{}, err
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
			SourceText:        chunk.Text,
			ChapterSource:     s.ChapterSource,
			Mood:              s.Mood,
			CharactersPresent: s.CharactersPresent,
		}
	}

	return scenes, llm.Usage{}, nil
}

// convertScene 将一个场景转换为剧本元素序列。
func (o *Orchestrator) convertScene(ctx context.Context, scene *models.Scene, characters []models.Character) ([]models.ScriptElement, llm.Usage, error) {
	// 构建角色上下文
	charCtx := buildCharacterContext(characters)
	charList := buildCharacterList(scene.CharactersPresent)

	prompt := llm.ScriptConversionPrompt
	prompt = strings.Replace(prompt, "{character_context}", charCtx, 1)
	prompt = strings.Replace(prompt, "{scene_title}", scene.Title, 1)
	prompt = strings.Replace(prompt, "{scene_summary}", scene.Summary, 1)
	prompt = strings.Replace(prompt, "{location}", scene.Setting.Location, 1)
	prompt = strings.Replace(prompt, "{time}", scene.Setting.TimeOfDay, 1)
	prompt = strings.Replace(prompt, "{characters_present}", charList, 1)
	sceneText := scene.Summary
	if sceneText == "" {
		sceneText = scene.SourceText
	}
	prompt = strings.Replace(prompt, "{text}", sceneText, 1)

	type llmElement struct {
		ID           string `json:"id"`
		Type         string `json:"type"`
		Content      string `json:"content"`
		SpeakerID    string `json:"speaker_id,omitempty"`
		SpeakerName  string `json:"speaker_name,omitempty"`
		Tone         string `json:"tone,omitempty"`
		Delivery     string `json:"delivery,omitempty"`
		VisualCue    string `json:"visual_cue,omitempty"`
		Visibility   string `json:"visibility,omitempty"`
	}

	rawElements, err := llm.StructuredGenerate[[]llmElement](ctx, o.client, llm.SystemPrompt, prompt)
	if err != nil {
		return nil, llm.Usage{}, err
	}

	elements := make([]models.ScriptElement, len(rawElements))
	for i, e := range rawElements {
		elements[i] = models.ScriptElement{
			Type:         models.ElementType(e.Type),
			ID:           e.ID,
			Content:      e.Content,
			SpeakerID:    e.SpeakerID,
			SpeakerName:  e.SpeakerName,
			Tone:         e.Tone,
			Delivery:     e.Delivery,
			VisualCue:    e.VisualCue,
			Visibility:   e.Visibility,
		}
	}

	return elements, llm.Usage{}, nil
}

func buildCharacterContext(characters []models.Character) string {
	var b strings.Builder
	for _, ch := range characters {
		fmt.Fprintf(&b, "- %s (id: %s, role: %s): %s\n", ch.Name, ch.ID, ch.Role, ch.Description)
	}
	return b.String()
}

func buildCharacterList(present []models.CharacterPresence) string {
	ids := make([]string, len(present))
	for i, p := range present {
		ids[i] = p.ID
	}
	return strings.Join(ids, ", ")
}

func detectScriptTitle(chapters []text.Chapter) string {
	if len(chapters) == 0 {
		return "未命名剧本"
	}
	// 取第一章标题作为剧本名
	return chapters[0].Title + "·剧本改编"
}

func (o *Orchestrator) log(format string, args ...any) {
	if o.cfg.Verbose {
		log.Printf("[novel2script] "+format, args...)
	}
}

// warn 总是输出警告信息到 stderr。
func (o *Orchestrator) warn(format string, args ...any) {
	log.Printf("[novel2script] "+format, args...)
}

// fillTargetIDs 根据角色名和别名自动补全关系中的 target_id。
func fillTargetIDs(characters []models.Character) {
	// 构建 name/alias → id 的索引
	index := make(map[string]string)
	for _, ch := range characters {
		index[ch.Name] = ch.ID
		for _, alias := range ch.Aliases {
			if _, exists := index[alias]; !exists {
				index[alias] = ch.ID
			}
		}
	}

	for i := range characters {
		ch := &characters[i]
		for j := range ch.Relationships {
			rel := &ch.Relationships[j]
			if rel.TargetID != "" {
				continue
			}
			// 在描述中查找已知角色名（排除自己避免自引用）
			rel.TargetID = findTargetID(rel.Description, index, ch.ID)
		}
	}
}

// findTargetID 在文本中查找已索引的角色名，返回对应的 ID。
// excludeID 用于排除当前角色自身，避免自引用。
func findTargetID(text string, index map[string]string, excludeID string) string {
	var bestMatch string
	bestLen := 0
	for name, id := range index {
		if id == excludeID {
			continue
		}
		if len(name) > bestLen && strings.Contains(text, name) {
			bestMatch = id
			bestLen = len(name)
		}
	}
	return bestMatch
}
