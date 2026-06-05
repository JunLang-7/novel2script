package models

import "time"

// Script 是完整的剧本，为 YAML 序列化的顶层结构。
type Script struct {
	ScriptTitle  string      `yaml:"script_title" validate:"required"`
	SourceNovel  string      `yaml:"source_novel" validate:"required"`
	SourceAuthor string      `yaml:"source_author,omitempty"`
	Adaptor      string      `yaml:"adaptor,omitempty"`
	GeneratedAt  time.Time   `yaml:"generated_at"`
	Version      string      `yaml:"version"`
	Metadata     Metadata    `yaml:"metadata"`
	Characters   []Character `yaml:"characters" validate:"required,dive"`
	Acts         []Act       `yaml:"acts" validate:"required,dive"`
}

// Metadata 包含剧本的元信息。
type Metadata struct {
	Genre                []string `yaml:"genre,omitempty"`
	OriginalLanguage     string   `yaml:"original_language,omitempty"`
	TargetFormat         string   `yaml:"target_format,omitempty"`
	TotalNovelChapters   int      `yaml:"total_novel_chapters,omitempty"`
	TotalNovelChars      int      `yaml:"total_novel_chars,omitempty"`
	AdaptationCoverage   string   `yaml:"adaptation_coverage,omitempty"`
	Synopsis             string   `yaml:"synopsis,omitempty"`
	EstimatedTotalScenes int      `yaml:"estimated_total_scenes,omitempty"`
}

// Act 代表剧本中的一幕。
type Act struct {
	ID           string  `yaml:"id" validate:"required"`
	Title        string  `yaml:"title" validate:"required"`
	Summary      string  `yaml:"summary,omitempty"`
	ChapterRange string  `yaml:"chapter_range,omitempty"`
	Scenes       []Scene `yaml:"scenes" validate:"required,dive"`
}

// SceneType 枚举场景的类型。
type SceneType string

const (
	SceneTypeScene    SceneType = "scene"
	SceneTypeMontage  SceneType = "montage"
	SceneTypeFlashback SceneType = "flashback"
	SceneTypeInterlude SceneType = "interlude"
)

// Scene 代表剧本中的一个场景。
type Scene struct {
	ID                 string              `yaml:"id" validate:"required"`
	Type               SceneType           `yaml:"type"`
	Title              string              `yaml:"title" validate:"required"`
	Sequence           int                 `yaml:"sequence"`
	Setting            SceneSetting        `yaml:"setting" validate:"required"`
	Summary            string              `yaml:"summary,omitempty"`
	ChapterSource      int                 `yaml:"chapter_source,omitempty"`
	EstimatedDuration  string              `yaml:"estimated_duration,omitempty"`
	Mood               string              `yaml:"mood,omitempty"`
	CharactersPresent  []CharacterPresence `yaml:"characters_present,omitempty"`
	Elements           []ScriptElement     `yaml:"elements" validate:"required,dive"`
	Transition         *Transition         `yaml:"transition,omitempty"`
}

// SceneSetting 描述场景的环境设置。
type SceneSetting struct {
	Location     string `yaml:"location" validate:"required"`
	LocationType string `yaml:"location_type,omitempty"`
	TimeOfDay    string `yaml:"time_of_day,omitempty"`
	Atmosphere   string `yaml:"atmosphere,omitempty"`
	Weather      string `yaml:"weather,omitempty"`
	Era          string `yaml:"era,omitempty"`
}

// TransitionType 枚举转场类型。
type TransitionType string

const (
	TransitionCutTo      TransitionType = "cut_to"
	TransitionFadeTo     TransitionType = "fade_to"
	TransitionDissolveTo TransitionType = "dissolve_to"
	TransitionMatchCut   TransitionType = "match_cut"
)

// Transition 描述场景间的转场。
type Transition struct {
	Type           TransitionType `yaml:"type"`
	NextSceneHint  string         `yaml:"next_scene_hint,omitempty"`
	Duration       string         `yaml:"duration,omitempty"`
}
