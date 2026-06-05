package formatters

import (
	"fmt"
	"io"
	"strings"

	"github.com/JunLang-7/novel2script/internal/models"
)

// WriteMarkdown 将 Script 格式化为可读性高的 Markdown 文本。
func WriteMarkdown(w io.Writer, script *models.Script) error {
	var b strings.Builder

	// 标题
	fmt.Fprintf(&b, "# %s\n\n", script.ScriptTitle)
	fmt.Fprintf(&b, "> 原著: %s", script.SourceNovel)
	if script.SourceAuthor != "" {
		fmt.Fprintf(&b, " | 作者: %s", script.SourceAuthor)
	}
	fmt.Fprintf(&b, "\n> 改编工具: %s | 生成时间: %s\n\n", script.Adaptor,
		script.GeneratedAt.Format("2006-01-02 15:04:05"))

	// 元数据
	b.WriteString("---\n\n## 基本信息\n\n")
	fmt.Fprintf(&b, "- **题材**: %s\n", strings.Join(script.Metadata.Genre, "、"))
	fmt.Fprintf(&b, "- **原著章节**: %d 章\n", script.Metadata.TotalNovelChapters)
	fmt.Fprintf(&b, "- **原著字数**: %s\n", formatChars(script.Metadata.TotalNovelChars))
	fmt.Fprintf(&b, "- **改编范围**: %s\n", script.Metadata.AdaptationCoverage)
	if script.Metadata.Synopsis != "" {
		fmt.Fprintf(&b, "- **故事梗概**: %s\n", script.Metadata.Synopsis)
	}

	// 角色表
	b.WriteString("\n---\n\n## 角色表\n\n")
	b.WriteString("| 角色 | 定位 | 描述 | 特征 |\n")
	b.WriteString("|------|------|------|------|\n")
	for _, ch := range script.Characters {
		fmt.Fprintf(&b, "| **%s** | %s | %s | %s |\n",
			ch.Name, ch.Role, ch.Description, strings.Join(ch.Traits, "、"))
	}

	// 角色关系
	b.WriteString("\n### 角色关系\n\n")
	for _, ch := range script.Characters {
		if len(ch.Relationships) > 0 {
			fmt.Fprintf(&b, "**%s**: \n", ch.Name)
			for _, rel := range ch.Relationships {
				// 查找关联角色名
				targetName := rel.TargetID
				for _, c := range script.Characters {
					if c.ID == rel.TargetID {
						targetName = c.Name
						break
					}
				}
				fmt.Fprintf(&b, "- %s → %s: %s\n", rel.Type, targetName, rel.Description)
			}
			b.WriteString("\n")
		}
	}

	// 剧本正文
	b.WriteString("---\n\n## 剧本正文\n\n")
	for _, act := range script.Acts {
		fmt.Fprintf(&b, "## %s\n\n", act.Title)
		if act.Summary != "" {
			fmt.Fprintf(&b, "> %s\n\n", act.Summary)
		}

		for _, scene := range act.Scenes {
			fmt.Fprintf(&b, "### 场景 %d: %s\n\n", scene.Sequence, scene.Title)

			// 场景信息
			b.WriteString("| 属性 | 值 |\n|------|------|\n")
			fmt.Fprintf(&b, "| 地点 | %s |\n", scene.Setting.Location)
			if scene.Setting.TimeOfDay != "" {
				fmt.Fprintf(&b, "| 时间 | %s |\n", scene.Setting.TimeOfDay)
			}
			if scene.Setting.Atmosphere != "" {
				fmt.Fprintf(&b, "| 氛围 | %s |\n", scene.Setting.Atmosphere)
			}
			if scene.EstimatedDuration != "" {
				fmt.Fprintf(&b, "| 预估时长 | %s |\n", scene.EstimatedDuration)
			}
			b.WriteString("\n")

			// 出场角色
			if len(scene.CharactersPresent) > 0 {
				b.WriteString("**出场角色**: ")
				names := make([]string, len(scene.CharactersPresent))
				for i, cp := range scene.CharactersPresent {
					// 查找角色名
					name := cp.ID
					for _, c := range script.Characters {
						if c.ID == cp.ID {
							name = c.Name
							break
						}
					}
					if cp.State != "" {
						names[i] = fmt.Sprintf("%s(%s)", name, cp.State)
					} else {
						names[i] = name
					}
				}
				b.WriteString(strings.Join(names, "、"))
				b.WriteString("\n\n")
			}

			// 剧本元素
			for _, elem := range scene.Elements {
				formatted := FormatElement(elem)
				switch elem.Type {
				case models.ElemAction:
					b.WriteString(formatted + "\n\n")
				case models.ElemDialogue, models.ElemInternalMonologue:
					b.WriteString(formatted + "\n\n")
				case models.ElemNarration:
					b.WriteString("> " + formatted + "\n\n")
				case models.ElemTitleCard:
					b.WriteString("### " + formatted + "\n\n")
				default:
					b.WriteString(formatted + "\n\n")
				}
			}

			// 转场
			if scene.Transition != nil {
				fmt.Fprintf(&b, "*[转场: %s]*\n\n", scene.Transition.Type)
			}
		}
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func formatChars(n int) string {
	if n >= 10000 {
		return fmt.Sprintf("%d万字", n/10000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%d千字", n/1000)
	}
	return fmt.Sprintf("%d字", n)
}
