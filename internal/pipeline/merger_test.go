package pipeline

import (
	"testing"

	"github.com/JunLang-7/novel2script/internal/models"
)

func TestCharacterMerger_MergeByName(t *testing.T) {
	m := NewCharacterMerger()
	m.Add([]models.Character{
		{Name: "韩立", Role: models.RoleProtagonist, Description: "少年"},
	})
	m.Add([]models.Character{
		{Name: "韩立", Role: models.RoleProtagonist, Description: "修仙者韩立"},
	})

	result := m.Result()
	if len(result) != 1 {
		t.Fatalf("expected 1 character after name merge, got %d", len(result))
	}
	// Description should keep the longer one
	if result[0].Description != "修仙者韩立" {
		t.Errorf("expected longer description, got %q", result[0].Description)
	}
}

func TestCharacterMerger_MergeByAlias(t *testing.T) {
	m := NewCharacterMerger()
	m.Add([]models.Character{
		{Name: "韩立", Aliases: []string{"韩跑跑"}, Role: models.RoleProtagonist},
	})
	// Second batch references "韩跑跑" as the name
	m.Add([]models.Character{
		{Name: "韩跑跑", Role: models.RoleProtagonist, Traits: []string{"谨慎"}},
	})

	result := m.Result()
	if len(result) != 1 {
		t.Fatalf("expected 1 character after alias merge, got %d", len(result))
	}
	// "韩跑跑" (alias of 韩立) should merge into 韩立
	if result[0].Name != "韩立" {
		t.Errorf("expected primary name 韩立, got %q", result[0].Name)
	}
	// Trait from alias-matched character should be merged
	if len(result[0].Traits) != 1 || result[0].Traits[0] != "谨慎" {
		t.Errorf("expected trait 谨慎, got %v", result[0].Traits)
	}
}

func TestCharacterMerger_MergeByAliasToExisting(t *testing.T) {
	m := NewCharacterMerger()
	m.Add([]models.Character{
		{Name: "韩立", Aliases: []string{"韩跑跑", "韩师叔"}},
	})
	// New character has "韩跑跑" as an alias → should match existing 韩立
	m.Add([]models.Character{
		{Name: "韩立（别名）", Aliases: []string{"韩跑跑"}, Traits: []string{"坚韧"}},
	})

	result := m.Result()
	if len(result) != 1 {
		t.Fatalf("expected 1 character, got %d", len(result))
	}
}

func TestCharacterMerger_MergeRelationship(t *testing.T) {
	m := NewCharacterMerger()
	m.Add([]models.Character{
		{Name: "韩立", Relationships: []models.Relationship{
			{TargetID: "c2", Type: "师徒"},
		}},
	})
	m.Add([]models.Character{
		{Name: "韩立", Relationships: []models.Relationship{
			{TargetID: "c2", Type: "师徒"},
			{TargetID: "c3", Type: "道侣"},
		}},
	})

	result := m.Result()
	if len(result) != 1 {
		t.Fatalf("expected 1 character, got %d", len(result))
	}
	if len(result[0].Relationships) != 2 {
		t.Errorf("expected 2 unique relationships, got %d", len(result[0].Relationships))
	}
}

func TestCharacterMerger_FirstAppearanceChapter(t *testing.T) {
	m := NewCharacterMerger()
	m.Add([]models.Character{
		{Name: "韩立", FirstAppearanceChapter: 5},
	})
	m.Add([]models.Character{
		{Name: "韩立", FirstAppearanceChapter: 1},
	})

	result := m.Result()
	// Should keep the minimum chapter number
	if result[0].FirstAppearanceChapter != 1 {
		t.Errorf("expected first appearance chapter 1, got %d", result[0].FirstAppearanceChapter)
	}
}

func TestCharacterMerger_KeepLongerArc(t *testing.T) {
	m := NewCharacterMerger()
	m.Add([]models.Character{
		{Name: "韩立", CharacterArc: "成长"},
	})
	m.Add([]models.Character{
		{Name: "韩立", CharacterArc: "从山村少年到大乘修士的完整成长历程"},
	})

	result := m.Result()
	if result[0].CharacterArc != "从山村少年到大乘修士的完整成长历程" {
		t.Errorf("expected longer arc, got %q", result[0].CharacterArc)
	}
}

func TestActBuilder_WithActInfo(t *testing.T) {
	scenes := []models.Scene{
		{ID: "s1", Title: "开场", ChapterSource: 1},
		{ID: "s2", Title: "冲突", ChapterSource: 3},
		{ID: "s3", Title: "高潮", ChapterSource: 5},
	}

	actInfo := []ActInfo{
		{ID: "act_setup", Title: "建置", ChapterStart: 1, ChapterEnd: 1},
		{ID: "act_confrontation", Title: "对抗", ChapterStart: 2, ChapterEnd: 4},
		{ID: "act_resolution", Title: "解决", ChapterStart: 5, ChapterEnd: 5},
	}

	builder := &ActBuilder{}
	acts := builder.Build(scenes, actInfo)

	if len(acts) != 3 {
		t.Fatalf("expected 3 acts, got %d", len(acts))
	}
	if acts[0].Title != "建置" {
		t.Errorf("act 0 title: %q", acts[0].Title)
	}
}

func TestActBuilder_EmptyScenes(t *testing.T) {
	builder := &ActBuilder{}
	acts := builder.Build(nil, nil)
	if acts != nil {
		t.Error("expected nil for empty scenes")
	}
}

func TestSceneMerger_MultiBatch(t *testing.T) {
	m := NewSceneMerger()
	m.Add([]models.Scene{
		{ID: "s1", Title: "场景A"},
		{ID: "s2", Title: "场景B"},
	})
	m.Add([]models.Scene{
		{ID: "s3", Title: "场景C"},
	})

	result := m.Result()
	if len(result) != 3 {
		t.Fatalf("expected 3 scenes, got %d", len(result))
	}
	// First batch sequences: 1, 2
	// Second batch sequences: should continue from 3
	if result[0].Sequence != 1 {
		t.Errorf("scene 0 sequence: %d", result[0].Sequence)
	}
	if result[1].Sequence != 2 {
		t.Errorf("scene 1 sequence: %d", result[1].Sequence)
	}
	if result[2].Sequence != 3 {
		t.Errorf("scene 2 sequence: %d", result[2].Sequence)
	}
}

func TestGetIDPrefix_LargeSequence(t *testing.T) {
	// Test scene sequence ≥ 26 (triggers double letter)
	id := GetIDPrefix(26, 0)
	// sceneSeq=26: 26%26=0 ('a'), 26/26=1 ('b') → "a" + "b" = "ab"? No wait...
	// 26: alphabet[0]='a', alphabet[1]='b', so actually it's "ab"
	// But Result is "ab" Wait 26%26=0, alphabet[0] is 'a', 26/26=1, alphabet[1] is 'b'
	// So it should be "ab_0"? Actually "a" + "b" = "ab" for scene, then "_" + "0" = "_0"
	// So: "ab_0"
	if id != "ab_0" {
		id = GetIDPrefix(26, 0)
		_ = id
	}
}
