package models

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestScript_YAMLRoundTrip(t *testing.T) {
	original := &Script{
		ScriptTitle:  "测试剧本",
		SourceNovel:  "测试小说",
		SourceAuthor: "测试作者",
		Adaptor:      "novel2script v0.1.0",
		GeneratedAt:  time.Date(2026, 6, 5, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		Version:      "0.1.0",
		Metadata: Metadata{
			Genre:              []string{"仙侠"},
			OriginalLanguage:   "zh-CN",
			TargetFormat:       "screenplay",
			TotalNovelChapters: 5,
		},
		Characters: []Character{
			{ID: "c1", Name: "韩立", Role: RoleProtagonist, ImportanceRank: 1},
		},
		Acts: []Act{
			{
				ID:      "act_1",
				Title:   "第一幕",
				Summary: "开端",
				Scenes: []Scene{
					{
						ID:       "scene_1_1",
						Type:     SceneTypeScene,
						Title:    "开场",
						Sequence: 1,
						Setting: SceneSetting{
							Location: "测试地点",
						},
						Elements: []ScriptElement{
							{
								Type:    ElemAction,
								ID:      "e1",
								Content: "测试内容",
							},
						},
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var restored Script
	if err := yaml.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if restored.ScriptTitle != "测试剧本" {
		t.Errorf("script_title: %q", restored.ScriptTitle)
	}
	if restored.SourceNovel != "测试小说" {
		t.Errorf("source_novel: %q", restored.SourceNovel)
	}
	if len(restored.Characters) != 1 {
		t.Errorf("characters count: %d", len(restored.Characters))
	}
	if restored.Characters[0].Name != "韩立" {
		t.Errorf("character name: %q", restored.Characters[0].Name)
	}
	if len(restored.Acts) != 1 {
		t.Errorf("acts count: %d", len(restored.Acts))
	}
	if len(restored.Acts[0].Scenes) != 1 {
		t.Errorf("scenes count: %d", len(restored.Acts[0].Scenes))
	}
}

func TestScriptElement_IsSpoken(t *testing.T) {
	tests := []struct {
		typ      ElementType
		expected bool
	}{
		{ElemAction, false},
		{ElemDialogue, true},
		{ElemInternalMonologue, true},
		{ElemNarration, false},
		{ElemTitleCard, false},
	}

	for _, tt := range tests {
		elem := ScriptElement{Type: tt.typ}
		if got := elem.IsSpoken(); got != tt.expected {
			t.Errorf("%s.IsSpoken() = %v, want %v", tt.typ, got, tt.expected)
		}
	}
}

func TestScriptElement_NeedsSpeaker(t *testing.T) {
	tests := []struct {
		typ      ElementType
		expected bool
	}{
		{ElemAction, false},
		{ElemDialogue, true},
		{ElemInternalMonologue, true},
		{ElemNarration, false},
		{ElemTitleCard, false},
	}

	for _, tt := range tests {
		elem := ScriptElement{Type: tt.typ}
		if got := elem.NeedsSpeaker(); got != tt.expected {
			t.Errorf("%s.NeedsSpeaker() = %v, want %v", tt.typ, got, tt.expected)
		}
	}
}

func TestSceneTypes(t *testing.T) {
	if SceneTypeScene != "scene" {
		t.Errorf("SceneTypeScene: %q", SceneTypeScene)
	}
	if SceneTypeMontage != "montage" {
		t.Errorf("SceneTypeMontage: %q", SceneTypeMontage)
	}
	if SceneTypeFlashback != "flashback" {
		t.Errorf("SceneTypeFlashback: %q", SceneTypeFlashback)
	}
	if SceneTypeInterlude != "interlude" {
		t.Errorf("SceneTypeInterlude: %q", SceneTypeInterlude)
	}
}

func TestTransitionTypes(t *testing.T) {
	if TransitionCutTo != "cut_to" {
		t.Errorf("TransitionCutTo: %q", TransitionCutTo)
	}
	if TransitionFadeTo != "fade_to" {
		t.Errorf("TransitionFadeTo: %q", TransitionFadeTo)
	}
	if TransitionDissolveTo != "dissolve_to" {
		t.Errorf("TransitionDissolveTo: %q", TransitionDissolveTo)
	}
	if TransitionMatchCut != "match_cut" {
		t.Errorf("TransitionMatchCut: %q", TransitionMatchCut)
	}
}

func TestScene_SourceTextNotSerialized(t *testing.T) {
	scene := Scene{
		ID:         "s1",
		Title:      "测试",
		SourceText: "原始文本内容很长老说",
		Setting: SceneSetting{
			Location: "某地",
		},
	}

	data, err := yaml.Marshal(&scene)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if contains(string(data), "原始文本内容") {
		t.Error("SourceText should not appear in YAML output")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
