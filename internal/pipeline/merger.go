package pipeline

import (
	"strings"

	"github.com/JunLang-7/novel2script/internal/models"
)

// CharacterMerger 合并来自多个 chunk 的角色提取结果，去重并合并同一角色的不同别名。
type CharacterMerger struct {
	nameIndex map[string]int   // name → index in result
	aliasIndex map[string]int  // alias → index in result
	result    []models.Character
}

// NewCharacterMerger 创建角色合并器。
func NewCharacterMerger() *CharacterMerger {
	return &CharacterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
	}
}

// Add 添加一批新提取的角色，自动合并已知角色。
func (m *CharacterMerger) Add(characters []models.Character) {
	for _, ch := range characters {
		idx, isDuplicate := m.findExisting(&ch)
		if isDuplicate {
			m.mergeInto(idx, &ch)
		} else {
			idx = len(m.result)
			m.result = append(m.result, ch)
			m.indexCharacter(idx, &ch)
		}
	}
}

// Result 返回合并后的角色列表。
func (m *CharacterMerger) Result() []models.Character {
	// 重新分配 importance_rank
	for i := range m.result {
		m.result[i].ImportanceRank = i + 1
	}
	return m.result
}

func (m *CharacterMerger) findExisting(ch *models.Character) (int, bool) {
	// 按 name 精确匹配
	if idx, ok := m.nameIndex[ch.Name]; ok {
		return idx, true
	}
	// 按别名匹配
	for _, alias := range ch.Aliases {
		if idx, ok := m.aliasIndex[alias]; ok {
			return idx, true
		}
	}
	// 检查结果中角色的别名是否包含 ch.Name
	if idx, ok := m.aliasIndex[ch.Name]; ok {
		return idx, true
	}
	return -1, false
}

func (m *CharacterMerger) mergeInto(idx int, ch *models.Character) {
	existing := &m.result[idx]

	// 合并别名（去重）
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

	// 合并特征（去重）
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

	// 合并关系（去重，按 target_id）
	relSet := make(map[string]bool)
	for _, r := range existing.Relationships {
		relSet[r.TargetID] = true
	}
	for _, r := range ch.Relationships {
		if !relSet[r.TargetID] {
			existing.Relationships = append(existing.Relationships, r)
		}
	}

	// description 优先级：保留更长的
	if len(ch.Description) > len(existing.Description) {
		existing.Description = ch.Description
	}

	// character_arc 优先级：保留更长的
	if len(ch.CharacterArc) > len(existing.CharacterArc) {
		existing.CharacterArc = ch.CharacterArc
	}

	// first_appearance 取最小值
	if ch.FirstAppearanceChapter > 0 &&
		(existing.FirstAppearanceChapter == 0 || ch.FirstAppearanceChapter < existing.FirstAppearanceChapter) {
		existing.FirstAppearanceChapter = ch.FirstAppearanceChapter
	}
}

func (m *CharacterMerger) indexCharacter(idx int, ch *models.Character) {
	m.nameIndex[ch.Name] = idx
	for _, alias := range ch.Aliases {
		if _, ok := m.aliasIndex[alias]; !ok {
			m.aliasIndex[alias] = idx
		}
	}
}

// SceneMerger 合并来自多个 chunk 的场景列表，重新编号。
type SceneMerger struct {
	scenes []models.Scene
}

// NewSceneMerger 创建场景合并器。
func NewSceneMerger() *SceneMerger {
	return &SceneMerger{}
}

// Add 添加一批场景。
func (m *SceneMerger) Add(scenes []models.Scene) {
	// 重新分配 sequence 以确保全局有序
	offset := len(m.scenes)
	for i := range scenes {
		scenes[i].Sequence = offset + i + 1
	}
	m.scenes = append(m.scenes, scenes...)
}

// Result 返回所有场景。
func (m *SceneMerger) Result() []models.Scene {
	return m.scenes
}

// ActBuilder 根据场景列表构建分幕结构。
type ActBuilder struct{}

// Build 将场景列表组织为分幕结构。
// 如果提供了 act 信息（来自情节分析），按 act 分配场景；
// 否则采用简单的等分策略。
func (a *ActBuilder) Build(scenes []models.Scene, actInfo []ActInfo) []models.Act {
	if len(scenes) == 0 {
		return nil
	}

	if len(actInfo) > 0 {
		return a.buildFromActInfo(scenes, actInfo)
	}

	// 默认：按场景数等分为 3 幕（三幕结构）
	return a.buildThreeAct(scenes)
}

func (a *ActBuilder) buildFromActInfo(scenes []models.Scene, actInfo []ActInfo) []models.Act {
	acts := make([]models.Act, len(actInfo))
	for i, info := range actInfo {
		acts[i] = models.Act{
			ID:           info.ID,
			Title:        info.Title,
			Summary:      info.Summary,
			ChapterRange: info.ChapterRange,
		}
		for _, scene := range scenes {
			if scene.ChapterSource >= info.ChapterStart && scene.ChapterSource <= info.ChapterEnd {
				acts[i].Scenes = append(acts[i].Scenes, scene)
			}
		}
	}
	return acts
}

func (a *ActBuilder) buildThreeAct(scenes []models.Scene) []models.Act {
	n := len(scenes)
	bp1 := n / 3
	bp2 := n * 2 / 3

	return []models.Act{
		{
			ID:      "act_1",
			Title:   "第一幕",
			Summary: "故事开端",
			Scenes:  scenes[:bp1],
		},
		{
			ID:      "act_2",
			Title:   "第二幕",
			Summary: "故事发展",
			Scenes:  scenes[bp1:bp2],
		},
		{
			ID:      "act_3",
			Title:   "第三幕",
			Summary: "故事结局",
			Scenes:  scenes[bp2:],
		},
	}
}

// ActInfo 描述一幕的基本信息。
type ActInfo struct {
	ID           string
	Title        string
	Summary      string
	ChapterStart int
	ChapterEnd   int
	ChapterRange string
}

// getIDPrefix 从已合并场景列表中生成唯一的 ID 前缀。
func GetIDPrefix(sceneSeq int, elemSeq int) string {
	return strings.ToLower(sceneSeqToID(sceneSeq)) + "_" + elemSeqToID(elemSeq)
}

func sceneSeqToID(seq int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	if seq < 26 {
		return string(alphabet[seq])
	}
	return string(alphabet[seq%26]) + string(alphabet[seq/26])
}

func elemSeqToID(seq int) string {
	return fmtInt(seq)
}

func fmtInt(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
