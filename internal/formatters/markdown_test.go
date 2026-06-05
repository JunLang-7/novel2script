package formatters

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/JunLang-7/novel2script/internal/models"
)

func TestWriteMarkdown_BasicScript(t *testing.T) {
	script := &models.Script{
		ScriptTitle:  "测试剧本",
		SourceNovel:  "测试原著",
		SourceAuthor: "作者",
		Adaptor:      "novel2script v0.1.0",
		GeneratedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Version:      "1.0.0",
		Metadata: models.Metadata{
			Genre:              []string{"仙侠"},
			TotalNovelChapters: 3,
			TotalNovelChars:    4500,
			Synopsis:           "测试梗概",
		},
		Characters: []models.Character{
			{ID: "char_1", Name: "韩立", Role: models.RoleProtagonist, Traits: []string{"谨慎"}, Relationships: []models.Relationship{
				{TargetID: "char_2", Type: "师徒", Description: "师父"},
			}},
			{ID: "char_2", Name: "墨居仁", Role: models.RoleSupporting, Traits: []string{"严厉"}},
		},
		Acts: []models.Act{
			{
				ID:      "act_1",
				Title:   "第一幕",
				Summary: "故事开始",
				Scenes: []models.Scene{
					{
						ID:       "scene_1",
						Title:    "开场",
						Sequence: 1,
						Setting:  models.SceneSetting{Location: "山村", TimeOfDay: "清晨"},
						CharactersPresent: []models.CharacterPresence{
							{ID: "char_1", State: "修炼"},
						},
						Elements: []models.ScriptElement{
							{ID: "e1", Type: models.ElemAction, Content: "雾气笼罩。"},
							{ID: "e2", Type: models.ElemDialogue, SpeakerID: "char_1", SpeakerName: "韩立", Content: "开始了。"},
							{ID: "e3", Type: models.ElemInternalMonologue, SpeakerID: "char_1", SpeakerName: "韩立", Content: "必须突破。"},
						},
						Transition: &models.Transition{Type: models.TransitionFadeTo},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteMarkdown(&buf, script)
	if err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}

	output := buf.String()
	checks := []string{
		"# 测试剧本",
		"原著: 测试原著",
		"## 基本信息",
		"## 角色表",
		"韩立",
		"墨居仁",
		"## 剧本正文",
		"## 第一幕",
		"### 场景 1: 开场",
		"雾气笼罩",
		"韩立:",
		"[转场: fade_to]",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("missing expected content: %q", check)
		}
	}
}

func TestWriteMarkdown_EmptyFields(t *testing.T) {
	script := &models.Script{
		ScriptTitle: "测试",
		SourceNovel: "原著",
		Characters:  []models.Character{},
		Acts:        []models.Act{},
	}

	var buf bytes.Buffer
	err := WriteMarkdown(&buf, script)
	if err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}
}

func TestWriteMarkdown_TitleCard(t *testing.T) {
	script := &models.Script{
		ScriptTitle: "测试",
		SourceNovel: "原著",
		Acts: []models.Act{
			{
				ID:    "act_1",
				Title: "第一幕",
				Scenes: []models.Scene{
					{
						ID:      "scene_1",
						Title:   "开场",
						Setting: models.SceneSetting{Location: "某地"},
						Elements: []models.ScriptElement{
							{ID: "e1", Type: models.ElemTitleCard, Content: "第一章"},
							{ID: "e2", Type: models.ElemNarration, Content: "传说..."},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	WriteMarkdown(&buf, script)
	output := buf.String()

	if !strings.Contains(output, "[字幕] 第一章") {
		t.Error("missing title card")
	}
	if !strings.Contains(output, "[旁白] 传说...") {
		t.Error("missing narration")
	}
}

func TestFormatChars(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0字"},
		{500, "500字"},
		{1500, "1千字"},
		{10000, "1万字"},
		{25000, "2万字"},
	}

	for _, tt := range tests {
		got := formatChars(tt.n)
		if got != tt.want {
			t.Errorf("formatChars(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
