package formatters

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/JunLang-7/novel2script/internal/models"
)

func TestWriteYAML(t *testing.T) {
	script := &models.Script{
		ScriptTitle:  "测试剧本",
		SourceNovel:  "测试原著",
		SourceAuthor: "测试作者",
		Adaptor:      "novel2script v0.1.0",
		GeneratedAt:  time.Now(),
		Version:      "0.1.0",
		Metadata: models.Metadata{
			Genre:              []string{"仙侠"},
			OriginalLanguage:   "zh-CN",
			TargetFormat:       "screenplay",
			TotalNovelChapters: 3,
			TotalNovelChars:    4500,
		},
		Characters: []models.Character{
			{ID: "char_1", Name: "韩立", Role: models.RoleProtagonist},
		},
		Acts: []models.Act{
			{
				ID:    "act_1",
				Title: "第一幕",
				Scenes: []models.Scene{
					{
						ID:       "scene_1_1",
						Title:    "开场",
						Sequence: 1,
						Setting:  models.SceneSetting{Location: "山村"},
						Elements: []models.ScriptElement{
							{
								ID:      "elem_1",
								Type:    models.ElemAction,
								Content: "清晨，雾气笼罩着山村。",
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteYAML(&buf, script)
	if err != nil {
		t.Fatalf("WriteYAML failed: %v", err)
	}

	output := buf.String()

	// 验证 header 注释
	if !strings.Contains(output, "# novel2script 剧本格式") {
		t.Error("missing schema comment header")
	}

	// 验证关键字段
	if !strings.Contains(output, "script_title: 测试剧本") {
		t.Error("missing script_title")
	}
	if !strings.Contains(output, "source_novel: 测试原著") {
		t.Error("missing source_novel")
	}
	if !strings.Contains(output, "char_1") {
		t.Error("missing character id")
	}
}

func TestFormatElement(t *testing.T) {
	tests := []struct {
		name     string
		element  models.ScriptElement
		contains string
	}{
		{
			name:     "action",
			element:  models.ScriptElement{Type: models.ElemAction, Content: "他站了起来。"},
			contains: "[动作] 他站了起来。",
		},
		{
			name:     "dialogue",
			element:  models.ScriptElement{Type: models.ElemDialogue, SpeakerName: "韩立", Content: "你好。"},
			contains: "韩立: \"你好。\"",
		},
		{
			name:     "dialogue with tone",
			element:  models.ScriptElement{Type: models.ElemDialogue, SpeakerName: "韩立", Tone: "笑", Content: "好的。"},
			contains: "韩立（笑）",
		},
		{
			name:     "narration",
			element:  models.ScriptElement{Type: models.ElemNarration, Content: "传说中..."},
			contains: "[旁白] 传说中...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatElement(tt.element)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("FormatElement = %q, want contains %q", result, tt.contains)
			}
		})
	}
}

func TestEstimateDuration(t *testing.T) {
	scene := &models.Scene{
		Elements: []models.ScriptElement{
			{Content: strings.Repeat("文", 500)},
		},
	}

	duration := EstimateDuration(scene)
	if duration != "2min" {
		t.Errorf("expected '2min', got '%s'", duration)
	}

	shortScene := &models.Scene{Elements: nil}
	duration = EstimateDuration(shortScene)
	if duration != "<1min" {
		t.Errorf("expected '<1min', got '%s'", duration)
	}
}

func TestValidateScript(t *testing.T) {
	script := &models.Script{
		ScriptTitle: "测试",
		SourceNovel: "测试",
		Characters: []models.Character{
			{ID: "char_1", Name: "韩立", Role: models.RoleProtagonist},
		},
		Acts: []models.Act{
			{
				ID:    "act_1",
				Title: "第一幕",
				Scenes: []models.Scene{
					{
						ID:                "scene_1_1",
						Title:             "开场",
						Setting:           models.SceneSetting{Location: "山村"},
						CharactersPresent: []models.CharacterPresence{{ID: "char_1"}},
						Elements: []models.ScriptElement{
							{ID: "elem_1", Type: models.ElemAction, Content: "动作。"},
							{ID: "elem_2", Type: models.ElemDialogue, SpeakerID: "char_1", Content: "台词。"},
						},
					},
				},
			},
		},
	}

	warnings := ValidateScript(script)
	if len(warnings) > 0 {
		t.Errorf("expected 0 warnings for valid script, got: %v", warnings)
	}
}

func TestValidateScript_UnknownCharacter(t *testing.T) {
	script := &models.Script{
		ScriptTitle: "测试",
		SourceNovel: "测试",
		Characters: []models.Character{
			{ID: "char_1", Name: "韩立", Role: models.RoleProtagonist},
		},
		Acts: []models.Act{
			{
				ID:    "act_1",
				Title: "第一幕",
				Scenes: []models.Scene{
					{
						ID:      "scene_1_1",
						Title:   "开场",
						Setting: models.SceneSetting{Location: "山村"},
						Elements: []models.ScriptElement{
							{ID: "elem_1", Type: models.ElemDialogue, SpeakerID: "char_unknown", Content: "台词。"},
						},
					},
				},
			},
		},
	}

	warnings := ValidateScript(script)
	if len(warnings) == 0 {
		t.Error("expected warnings for unknown character reference")
	}
	t.Logf("warnings: %v", warnings)
}
