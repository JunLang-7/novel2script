package formatters

import (
	"testing"

	"github.com/JunLang-7/novel2script/internal/models"
)

func TestFormatElement_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		element  models.ScriptElement
		contains []string
	}{
		{"action", models.ScriptElement{Type: models.ElemAction, Content: "他走了一步。"}, []string{"[动作]", "他走了一步"}},
		{"dialogue_with_tone", models.ScriptElement{Type: models.ElemDialogue, SpeakerName: "韩立", Tone: "怒", Content: "放肆！"}, []string{"韩立（怒）", "放肆"}},
		{"dialogue_no_tone", models.ScriptElement{Type: models.ElemDialogue, SpeakerName: "张三", Content: "你好。"}, []string{"张三: \"你好"}},
		{"monologue", models.ScriptElement{Type: models.ElemInternalMonologue, SpeakerName: "李四", Visibility: "画外音", Content: "怎么办..."}, []string{"李四（内心独白·画外音）"}},
		{"narration", models.ScriptElement{Type: models.ElemNarration, Content: "很久以前..."}, []string{"[旁白]", "很久以前"}},
		{"title_card", models.ScriptElement{Type: models.ElemTitleCard, Content: "三天后"}, []string{"[字幕]", "三天后"}},
		{"unknown_type", models.ScriptElement{Type: "unknown", Content: "内容"}, []string{"内容"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatElement(tt.element)
			for _, c := range tt.contains {
				if !containsStr(result, c) {
					t.Errorf("FormatElement() = %q, want contains %q", result, c)
				}
			}
		})
	}
}

func TestValidateScript_EmptyScene(t *testing.T) {
	script := &models.Script{
		ScriptTitle: "测试",
		SourceNovel: "原著",
		Characters: []models.Character{
			{ID: "char_1", Name: "韩立", Role: models.RoleProtagonist},
		},
		Acts: []models.Act{
			{
				ID:    "act_1",
				Title: "第一幕",
				Scenes: []models.Scene{
					{
						ID:      "scene_1",
						Title:   "空场景",
						Setting: models.SceneSetting{Location: "某地"},
						Elements: []models.ScriptElement{},
					},
				},
			},
		},
	}

	warnings := ValidateScript(script)
	if len(warnings) == 0 {
		t.Error("expected warning for empty elements")
	}
}

func TestValidateScript_MultipleIssues(t *testing.T) {
	script := &models.Script{
		ScriptTitle: "测试",
		SourceNovel: "原著",
		Characters: []models.Character{
			{ID: "char_1", Name: "韩立", Role: models.RoleProtagonist},
		},
		Acts: []models.Act{
			{
				ID:    "act_1",
				Title: "第一幕",
				Scenes: []models.Scene{
					{
						ID:      "scene_1",
						Title:   "问题场景",
						Setting: models.SceneSetting{Location: "某地"},
						Elements: []models.ScriptElement{
							{ID: "e1", Type: models.ElemDialogue, SpeakerID: "char_999", Content: "台词"},
							{ID: "e2", Type: models.ElemInternalMonologue, SpeakerID: "", Content: "独白"},
						},
					},
				},
			},
		},
	}

	warnings := ValidateScript(script)
	foundUnknownChar := false
	for _, w := range warnings {
		if containsStr(w, "char_999") {
			foundUnknownChar = true
		}
	}
	if !foundUnknownChar {
		t.Error("expected warning about unknown character char_999")
	}
}

func TestBuildCharacterIndex(t *testing.T) {
	characters := []models.Character{
		{ID: "char_1", Name: "韩立"},
		{ID: "char_2", Name: "墨居仁"},
		{ID: "char_3", Name: "王婶"},
	}

	idx := BuildCharacterIndex(characters)
	if len(idx) != 3 {
		t.Errorf("expected 3 entries, got %d", len(idx))
	}
	if idx["char_1"].Name != "韩立" {
		t.Error("char_1 lookup failed")
	}
	if idx["char_99"] != nil {
		t.Error("non-existent character should return nil")
	}
}

func TestValidateScript_NoWarnings_ValidScript(t *testing.T) {
	script := &models.Script{
		ScriptTitle: "测试",
		SourceNovel: "原著",
		Characters: []models.Character{
			{ID: "char_1", Name: "韩立", Role: models.RoleProtagonist},
		},
		Acts: []models.Act{
			{
				ID:    "act_1",
				Title: "第一幕",
				Scenes: []models.Scene{
					{
						ID:                "scene_1",
						Title:             "场景",
						Setting:           models.SceneSetting{Location: "某地"},
						CharactersPresent: []models.CharacterPresence{{ID: "char_1"}},
						Elements: []models.ScriptElement{
							{ID: "e1", Type: models.ElemAction, Content: "动作"},
							{ID: "e2", Type: models.ElemDialogue, SpeakerID: "char_1", Content: "台词"},
						},
					},
				},
			},
		},
	}

	warnings := ValidateScript(script)
	if len(warnings) > 0 {
		t.Errorf("expected 0 warnings, got: %v", warnings)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
