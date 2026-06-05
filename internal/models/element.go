package models

// ElementType 枚举剧本元素的类型（联合类型的鉴别字段）。
type ElementType string

const (
	ElemAction            ElementType = "action"
	ElemDialogue          ElementType = "dialogue"
	ElemInternalMonologue ElementType = "internal_monologue"
	ElemNarration         ElementType = "narration"
	ElemTitleCard         ElementType = "title_card"
)

// ScriptElement 代表剧本中的一个最小叙事单元。
// 通过 Type 字段区分不同元素类型，不同 Type 携带不同的可选字段子集。
//
// 字段关联规则（由 validator 保证）：
//   - Type == "dialogue": SpeakerID, SpeakerName 必填
//   - Type == "internal_monologue": SpeakerID, SpeakerName 必填
//   - Type == "action": 仅 Content, VisualCue, SourceParagraph 有值
type ScriptElement struct {
	Type            ElementType `yaml:"type" validate:"required,oneof=action dialogue internal_monologue narration title_card"`
	ID              string      `yaml:"id" validate:"required"`
	Content         string      `yaml:"content" validate:"required"`

	// 对话/独白专有字段
	SpeakerID      string `yaml:"speaker_id,omitempty"`
	SpeakerName    string `yaml:"speaker_name,omitempty"`
	Tone           string `yaml:"tone,omitempty"`
	Delivery       string `yaml:"delivery,omitempty"`
	LanguageStyle  string `yaml:"language_style,omitempty"`

	// 内心独白专有字段
	Visibility string `yaml:"visibility,omitempty"`

	// 动作/画面专有字段
	VisualCue       string `yaml:"visual_cue,omitempty"`
	SourceParagraph int    `yaml:"source_paragraph,omitempty"`
}

// IsSpoken 判断元素是否包含角色台词（对话或独白）。
func (e *ScriptElement) IsSpoken() bool {
	return e.Type == ElemDialogue || e.Type == ElemInternalMonologue
}

// NeedsSpeaker 判断元素是否必须关联角色。
func (e *ScriptElement) NeedsSpeaker() bool {
	return e.IsSpoken()
}
