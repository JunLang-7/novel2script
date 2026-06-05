package analyzers

import (
	"testing"

	"github.com/JunLang-7/novel2script/internal/models"
	"github.com/JunLang-7/novel2script/internal/text"
)

func TestCharacterMerger_Add(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
	}

	m.add([]models.Character{
		{Name: "韩立", Role: models.RoleProtagonist, Description: "主角"},
		{Name: "墨居仁", Role: models.RoleSupporting, Description: "师父"},
	})

	result := m.result()
	if len(result) != 2 {
		t.Fatalf("expected 2 characters, got %d", len(result))
	}
	if result[0].Name != "韩立" {
		t.Errorf("first character: %q", result[0].Name)
	}
	if result[1].Name != "墨居仁" {
		t.Errorf("second character: %q", result[1].Name)
	}
}

func TestCharacterMerger_Empty(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
	}

	m.add([]models.Character{})
	m.add(nil)

	result := m.result()
	if len(result) != 0 {
		t.Errorf("expected 0 characters, got %d", len(result))
	}
}

func TestCharacterMerger_DedupByName(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
	}

	m.add([]models.Character{
		{Name: "韩立", Description: "短"},
	})
	m.add([]models.Character{
		{Name: "韩立", Description: "更长的描述信息"},
	})

	result := m.result()
	if len(result) != 1 {
		t.Fatalf("expected 1 character, got %d", len(result))
	}
	if result[0].Description != "更长的描述信息" {
		t.Errorf("expected longer description: %q", result[0].Description)
	}
}

func TestCharacterMerger_DedupByAlias(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
	}

	m.add([]models.Character{
		{Name: "韩立", Aliases: []string{"韩跑跑", "韩师叔"}},
	})
	// New character name matches existing alias
	m.add([]models.Character{
		{Name: "韩跑跑", Traits: []string{"速度"}},
	})

	result := m.result()
	if len(result) != 1 {
		t.Fatalf("expected 1 character after alias match, got %d", len(result))
	}
	if result[0].Name != "韩立" {
		t.Errorf("primary name should be 韩立: %q", result[0].Name)
	}
}

func TestCharacterMerger_FindExisting_NameMatch(t *testing.T) {
	m := &characterMerger{
		nameIndex:  map[string]int{"韩立": 0},
		aliasIndex: make(map[string]int),
	}

	ch := &models.Character{Name: "韩立"}
	idx, found := m.findExisting(ch)
	if !found {
		t.Error("should find by name")
	}
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
}

func TestCharacterMerger_FindExisting_AliasMatch(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: map[string]int{"韩跑跑": 0},
	}

	ch := &models.Character{Name: "韩立", Aliases: []string{"韩跑跑"}}
	idx, found := m.findExisting(ch)
	if !found {
		t.Error("should find by alias")
	}
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
}

func TestCharacterMerger_FindExisting_NameAsAlias(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: map[string]int{"韩立": 0},
	}

	ch := &models.Character{Name: "韩立"}
	idx, found := m.findExisting(ch)
	if !found {
		t.Error("should find when new name matches existing alias")
	}
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
}

func TestCharacterMerger_FindExisting_NotFound(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
	}

	ch := &models.Character{Name: "新角色"}
	_, found := m.findExisting(ch)
	if found {
		t.Error("should not find unknown character")
	}
}

func TestCharacterMerger_MergeInto_Traits(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
		chars: []models.Character{
			{Name: "韩立", Traits: []string{"谨慎", "坚韧"}},
		},
	}

	m.mergeInto(0, &models.Character{
		Traits: []string{"谨慎", "重情义"},
	})

	result := m.result()
	if len(result[0].Traits) != 3 {
		t.Errorf("expected 3 unique traits, got %d: %v", len(result[0].Traits), result[0].Traits)
	}
}

func TestCharacterMerger_MergeInto_Aliases(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
		chars: []models.Character{
			{Name: "韩立", Aliases: []string{"韩跑跑"}},
		},
	}

	m.mergeInto(0, &models.Character{
		Aliases: []string{"韩师叔", "韩跑跑"},
	})

	result := m.result()
	if len(result[0].Aliases) != 2 {
		t.Errorf("expected 2 unique aliases, got %d: %v", len(result[0].Aliases), result[0].Aliases)
	}
}

func TestCharacterMerger_IndexCharacter(t *testing.T) {
	m := &characterMerger{
		nameIndex:  make(map[string]int),
		aliasIndex: make(map[string]int),
	}

	ch := &models.Character{
		Name:    "韩立",
		Aliases: []string{"韩跑跑", "韩师叔"},
	}
	m.indexCharacter(0, ch)

	if idx, ok := m.nameIndex["韩立"]; !ok || idx != 0 {
		t.Error("name not indexed")
	}
	if idx, ok := m.aliasIndex["韩跑跑"]; !ok || idx != 0 {
		t.Error("alias 韩跑跑 not indexed")
	}
	if idx, ok := m.aliasIndex["韩师叔"]; !ok || idx != 0 {
		t.Error("alias 韩师叔 not indexed")
	}
}

func TestJoinChapters(t *testing.T) {
	chapters := []text.Chapter{
		{Title: "第一章", Content: "内容一"},
		{Title: "第二章", Content: "内容二"},
	}

	result := joinChapters(chapters)
	expected := "内容一\n\n内容二"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestJoinChapters_Empty(t *testing.T) {
	result := joinChapters(nil)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestJoinChapters_Single(t *testing.T) {
	chapters := []text.Chapter{
		{Title: "第一章", Content: "内容"},
	}

	result := joinChapters(chapters)
	if result != "内容" {
		t.Errorf("expected %q, got %q", "内容", result)
	}
}

func TestNewCharacterAnalyzer_Defaults(t *testing.T) {
	analyzer := NewCharacterAnalyzer(nil, 0)
	if analyzer.maxParallel != 3 {
		t.Errorf("default maxParallel: %d", analyzer.maxParallel)
	}

	analyzer2 := NewCharacterAnalyzer(nil, -5)
	if analyzer2.maxParallel != 3 {
		t.Errorf("negative maxParallel should default to 3: %d", analyzer2.maxParallel)
	}

	analyzer3 := NewCharacterAnalyzer(nil, 10)
	if analyzer3.maxParallel != 10 {
		t.Errorf("custom maxParallel: %d", analyzer3.maxParallel)
	}
}

func TestNewSceneAnalyzer_Defaults(t *testing.T) {
	analyzer := NewSceneAnalyzer(nil, 0)
	if analyzer.maxParallel != 5 {
		t.Errorf("default maxParallel: %d", analyzer.maxParallel)
	}

	analyzer2 := NewSceneAnalyzer(nil, 8)
	if analyzer2.maxParallel != 8 {
		t.Errorf("custom maxParallel: %d", analyzer2.maxParallel)
	}
}
