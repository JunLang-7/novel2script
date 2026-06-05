package converters

import (
	"strings"

	"github.com/JunLang-7/novel2script/internal/models"
)

// toneIndicators 中文网文中常见的语气指示词及其映射。
var toneIndicators = map[string]string{
	"笑道":   "笑",
	"怒道":   "怒",
	"冷声道": "冷",
	"厉声道": "厉",
	"叹道":   "叹",
	"哭道":   "哭",
	"急道":   "急",
	"惊道":   "惊",
	"喜道":   "喜",
	"淡淡道": "淡",
	"低声道": "低",
	"喝道":   "喝",
	"骂道":   "骂",
	"问道":   "问",
	"答道":   "答",
	"说道":   "",
}

// NormalizeDialogue 对 LLM 生成的对话元素进行后处理，归一口语化标注。
func NormalizeDialogue(elements []models.ScriptElement) []models.ScriptElement {
	for i := range elements {
		if elements[i].Type != models.ElemDialogue {
			continue
		}
		content := elements[i].Content

		// 从对话内容中移除语气后缀词（如 "韩立笑道：" → "韩立：" + tone=笑）
		for indicator, tone := range toneIndicators {
			idx := strings.Index(content, indicator+"：")
			if idx > 0 {
				// 仅在未明确标注 tone 时才推断
				if elements[i].Tone == "" && tone != "" {
					elements[i].Tone = tone
				}
				content = content[:idx] + "：" +
					content[idx+len(indicator)+len("："):]
				elements[i].Content = content
				break
			}
		}
	}
	return elements
}

// MergeShortActions 合并过短的动作序列（相邻的两个短action合并为一个）。
func MergeShortActions(elements []models.ScriptElement) []models.ScriptElement {
	if len(elements) < 2 {
		return elements
	}

	var merged []models.ScriptElement
	skipNext := false

	for i := range len(elements) {
		if skipNext {
			skipNext = false
			continue
		}

		current := elements[i]
		if current.Type == models.ElemAction &&
			i+1 < len(elements) &&
			elements[i+1].Type == models.ElemAction &&
			len([]rune(current.Content)) < 30 &&
			len([]rune(elements[i+1].Content)) < 30 {

			current.Content = current.Content + " " + elements[i+1].Content
			skipNext = true
		}
		merged = append(merged, current)
	}

	return merged
}
