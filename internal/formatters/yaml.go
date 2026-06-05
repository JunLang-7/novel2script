package formatters

import (
	"fmt"
	"io"
	"strings"

	"github.com/JunLang-7/novel2script/internal/models"

	"gopkg.in/yaml.v3"
)

// WriteYAML 将 Script 序列化为 YAML 写入 writer。
func WriteYAML(w io.Writer, script *models.Script) error {
	data, err := yaml.Marshal(script)
	if err != nil {
		return fmt.Errorf("YAML序列化失败: %w", err)
	}
	_, err = w.Write(addSchemaComment(data, script))
	return err
}

// WriteYAMLFile 将 Script 序列化为 YAML 写入文件路径。
func WriteYAMLFile(path string, script *models.Script) error {
	// 实际写入留给 formatters/yaml.go 的 WriteYAML 结合文件操作
	return nil // placeholder
}

// addSchemaComment 在 YAML 文件头添加 schema 引用注释。
func addSchemaComment(data []byte, script *models.Script) []byte {
	header := fmt.Sprintf("# novel2script 剧本格式 v%s\n", script.Version)
	header += fmt.Sprintf("# 原著: %s\n", script.SourceNovel)
	header += fmt.Sprintf("# 生成时间: %s\n", script.GeneratedAt.Format("2006-01-02 15:04:05"))
	header += "# schema: https://novel2script.dev/schemas/script-v1.json\n"
	header += "# 本文件为 AI 辅助生成的剧本初稿，可自由编辑修改。\n\n"
	return append([]byte(header), data...)
}

// FormatElement 将 ScriptElement 格式化为人类可读的字符串。
func FormatElement(e models.ScriptElement) string {
	switch e.Type {
	case models.ElemAction:
		return fmt.Sprintf("[动作] %s", e.Content)
	case models.ElemDialogue:
		if e.Tone != "" {
			return fmt.Sprintf("%s（%s）: \"%s\"", e.SpeakerName, e.Tone, e.Content)
		}
		return fmt.Sprintf("%s: \"%s\"", e.SpeakerName, e.Content)
	case models.ElemInternalMonologue:
		return fmt.Sprintf("%s（内心独白·%s）: \"%s\"", e.SpeakerName, e.Visibility, e.Content)
	case models.ElemNarration:
		return fmt.Sprintf("[旁白] %s", e.Content)
	case models.ElemTitleCard:
		return fmt.Sprintf("[字幕] %s", e.Content)
	default:
		return e.Content
	}
}

// EstimateDuration 估算场景时长（中文剧本约每分钟250-300字）。
func EstimateDuration(scene *models.Scene) string {
	totalChars := 0
	for _, elem := range scene.Elements {
		totalChars += len([]rune(elem.Content))
	}
	minutes := totalChars / 250
	if minutes < 1 {
		return "<1min"
	}
	return fmt.Sprintf("%dmin", minutes)
}

// BuildCharacterIndex 从角色表构建索引映射，用于交叉校验。
func BuildCharacterIndex(characters []models.Character) map[string]*models.Character {
	idx := make(map[string]*models.Character, len(characters))
	for i := range characters {
		idx[characters[i].ID] = &characters[i]
	}
	return idx
}

// ValidateScript 对 Script 进行交叉校验，返回警告列表。
func ValidateScript(script *models.Script) []string {
	var warnings []string
	charIdx := BuildCharacterIndex(script.Characters)

	for _, act := range script.Acts {
		for _, scene := range act.Scenes {
			// 检查场景至少有一个元素
			if len(scene.Elements) == 0 {
				warnings = append(warnings,
					fmt.Sprintf("场景 %s 没有剧本元素", scene.ID))
			}

			// 检查所有对话/独白的 speaker 存在于角色表
			for _, elem := range scene.Elements {
				if elem.IsSpoken() && elem.SpeakerID != "" {
					if _, ok := charIdx[elem.SpeakerID]; !ok {
						warnings = append(warnings,
							fmt.Sprintf("场景 %s 的元素 %s 引用了未知角色 ID: %s",
								scene.ID, elem.ID, elem.SpeakerID))
					}
				}
			}

			// 检查至少有一个角色在场
			if len(scene.CharactersPresent) == 0 {
				warnings = append(warnings,
					fmt.Sprintf("场景 %s 没有任何角色在场标注", scene.ID))
			}
		}
	}

	return warnings
}

// Ensure implement formatting interface
var _ = strings.TrimSpace
