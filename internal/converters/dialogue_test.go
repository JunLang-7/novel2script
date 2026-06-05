package converters

import (
	"testing"

	"github.com/JunLang-7/novel2script/internal/models"
)

func TestNormalizeDialogue_Basic(t *testing.T) {
	elements := []models.ScriptElement{
		{
			ID:      "elem_1",
			Type:    models.ElemDialogue,
			Content: "韩立笑道：这功法果然玄妙。",
			SpeakerName: "韩立",
		},
	}

	result := NormalizeDialogue(elements)

	if result[0].Tone != "笑" {
		t.Errorf("expected tone '笑', got '%s'", result[0].Tone)
	}

	// 检查 "笑道：" 已被移除
	if contains(result[0].Content, "笑道：") {
		t.Errorf("'笑道：' should have been removed from content, got: %s", result[0].Content)
	}
}

func TestNormalizeDialogue_NonDialogue(t *testing.T) {
	elements := []models.ScriptElement{
		{
			ID:      "elem_1",
			Type:    models.ElemAction,
			Content: "韩立笑道：这功法果然玄妙。",
		},
	}

	result := NormalizeDialogue(elements)

	// action 类型不处理
	if result[0].Tone != "" {
		t.Errorf("action should not have tone set")
	}
}

func TestNormalizeDialogue_ToneAlreadySet(t *testing.T) {
	elements := []models.ScriptElement{
		{
			ID:      "elem_1",
			Type:    models.ElemDialogue,
			Content: "韩立笑道：这功法果然玄妙。",
			Tone:    "喜",
		},
	}

	result := NormalizeDialogue(elements)

	// 已有 tone 不应覆盖
	if result[0].Tone != "喜" {
		t.Errorf("existing tone should not be overwritten, got '%s'", result[0].Tone)
	}
}

func TestMergeShortActions(t *testing.T) {
	shortAction1 := models.ScriptElement{
		ID:      "elem_1",
		Type:    models.ElemAction,
		Content: "他站起身。",
	}
	shortAction2 := models.ScriptElement{
		ID:      "elem_2",
		Type:    models.ElemAction,
		Content: "走向门口。",
	}

	elements := []models.ScriptElement{shortAction1, shortAction2}
	result := MergeShortActions(elements)

	if len(result) != 1 {
		t.Errorf("expected 1 merged element, got %d: %+v", len(result), result)
	}

	if result[0].Content != "他站起身。 走向门口。" {
		t.Errorf("unexpected merged content: %s", result[0].Content)
	}
}

func TestMergeShortActions_NotBothShort(t *testing.T) {
	longAction := models.ScriptElement{
		ID:      "elem_1",
		Type:    models.ElemAction,
		Content: "他站起身，环顾四周，确认没有异常后慢慢走向门口，心中充满了警惕和不安，手心微微出汗。",
	}
	shortAction := models.ScriptElement{
		ID:      "elem_2",
		Type:    models.ElemAction,
		Content: "推开门。",
	}

	elements := []models.ScriptElement{longAction, shortAction}
	result := MergeShortActions(elements)

	if len(result) != 2 {
		t.Errorf("long+short should not merge, got %d elements", len(result))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
