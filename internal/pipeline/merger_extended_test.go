package pipeline

import (
	"testing"

	"github.com/JunLang-7/novel2script/internal/models"
)

func TestCharacterMerger_SingleCharacter(t *testing.T) {
	m := NewCharacterMerger()
	m.Add([]models.Character{
		{Name: "韩立", Role: models.RoleProtagonist, Description: "主角"},
	})

	result := m.Result()
	if len(result) != 1 {
		t.Fatalf("expected 1 character, got %d", len(result))
	}
	if result[0].ImportanceRank != 1 {
		t.Errorf("single character should have rank 1, got %d", result[0].ImportanceRank)
	}
}

func TestCharacterMerger_EmptyAdd(t *testing.T) {
	m := NewCharacterMerger()
	m.Add([]models.Character{})
	m.Add(nil)

	result := m.Result()
	if len(result) != 0 {
		t.Errorf("expected 0 characters, got %d", len(result))
	}
}

func TestCharacterMerger_NoMergingForDifferentNames(t *testing.T) {
	m := NewCharacterMerger()
	m.Add([]models.Character{
		{Name: "韩立", Role: models.RoleProtagonist},
		{Name: "墨居仁", Role: models.RoleSupporting},
		{Name: "南宫婉", Role: models.RoleLoveInterest},
	})

	result := m.Result()
	if len(result) != 3 {
		t.Fatalf("expected 3 characters, got %d", len(result))
	}

	names := make(map[string]bool)
	for _, ch := range result {
		names[ch.Name] = true
	}
	if !names["韩立"] || !names["墨居仁"] || !names["南宫婉"] {
		t.Error("missing characters after merge")
	}
}

func TestCharacterMerger_TraitCrossBatchDedup(t *testing.T) {
	m := NewCharacterMerger()
	// Trait dedup only happens across batches (not within a single Add call)
	m.Add([]models.Character{
		{Name: "韩立", Traits: []string{"谨慎", "坚韧"}},
	})
	m.Add([]models.Character{
		{Name: "韩立", Traits: []string{"谨慎", "重情义"}},
	})
	result := m.Result()
	if len(result) != 1 {
		t.Fatalf("expected 1 character, got %d", len(result))
	}
	if len(result[0].Traits) != 3 {
		t.Errorf("expected 3 unique traits across batches, got %d: %v", len(result[0].Traits), result[0].Traits)
	}
}

func TestSceneMerger_Empty(t *testing.T) {
	m := NewSceneMerger()
	result := m.Result()
	if len(result) != 0 {
		t.Errorf("expected 0 scenes, got %d", len(result))
	}
}

func TestSceneMerger_SingleBatch(t *testing.T) {
	m := NewSceneMerger()
	m.Add([]models.Scene{
		{ID: "s1", Title: "一", Sequence: 5},
		{ID: "s2", Title: "二", Sequence: 3},
	})

	result := m.Result()
	if result[0].Sequence != 1 {
		t.Errorf("expected renumbered sequence 1, got %d", result[0].Sequence)
	}
	if result[1].Sequence != 2 {
		t.Errorf("expected renumbered sequence 2, got %d", result[1].Sequence)
	}
}

func TestActBuilder_SingleScene(t *testing.T) {
	// With 1 scene, integer division produces empty first acts.
	// This is a known limitation of the simple three-act split.
	scenes := []models.Scene{
		{ID: "s1", Title: "唯一场景", Sequence: 1},
	}

	builder := &ActBuilder{}
	acts := builder.Build(scenes, nil)

	if len(acts) != 3 {
		t.Fatalf("expected 3 acts, got %d", len(acts))
	}
	total := 0
	for _, act := range acts {
		total += len(act.Scenes)
	}
	if total != 1 {
		t.Errorf("total scenes should be 1, got %d", total)
	}
}

func TestActBuilder_FourScenes(t *testing.T) {
	scenes := []models.Scene{
		{ID: "s1"}, {ID: "s2"}, {ID: "s3"}, {ID: "s4"},
	}

	builder := &ActBuilder{}
	acts := builder.Build(scenes, nil)

	// 4/3 = 1, each act should get some scenes
	total := 0
	for _, act := range acts {
		total += len(act.Scenes)
	}
	if total != 4 {
		t.Errorf("total scenes across acts: %d, want 4", total)
	}
}

func TestGetIDPrefix(t *testing.T) {
	id := GetIDPrefix(0, 1)
	expected := "a_1"
	if id != expected {
		t.Errorf("GetIDPrefix(0,1) = %q, want %q", id, expected)
	}
}

func TestGetIDPrefix_SecondLetter(t *testing.T) {
	id := GetIDPrefix(1, 2)
	expected := "b_2"
	if id != expected {
		t.Errorf("GetIDPrefix(1,2) = %q, want %q", id, expected)
	}
}
