package models

// CharacterRole 枚举角色定位。
type CharacterRole string

const (
	RoleProtagonist    CharacterRole = "protagonist"
	RoleDeuteragonist  CharacterRole = "deuteragonist"
	RoleAntagonist     CharacterRole = "antagonist"
	RoleSupporting     CharacterRole = "supporting"
	RoleLoveInterest   CharacterRole = "love_interest"
	RoleCameo          CharacterRole = "cameo"
)

// Character 代表小说/剧本中的一个角色。
type Character struct {
	ID                     string         `yaml:"id" validate:"required"`
	Name                   string         `yaml:"name" validate:"required"`
	Aliases                []string       `yaml:"aliases,omitempty"`
	Role                   CharacterRole  `yaml:"role" validate:"required"`
	ImportanceRank         int            `yaml:"importance_rank,omitempty"`
	Description            string         `yaml:"description,omitempty"`
	Archetype              string         `yaml:"archetype,omitempty"`
	Traits                 []string       `yaml:"traits,omitempty"`
	CharacterArc           string         `yaml:"character_arc,omitempty"`
	FirstAppearanceChapter int            `yaml:"first_appearance_chapter,omitempty"`
	Notes                  string         `yaml:"notes,omitempty"`
	Relationships          []Relationship `yaml:"relationships,omitempty"`
}

// Relationship 描述两个角色之间的关系。
type Relationship struct {
	TargetID                 string `yaml:"target_id" validate:"required"`
	Type                     string `yaml:"type" validate:"required"`
	Description              string `yaml:"description,omitempty"`
	FirstInteractionChapter  int    `yaml:"first_interaction_chapter,omitempty"`
}

// CharacterPresence 描述角色在场景中的在场状态。
type CharacterPresence struct {
	ID    string `yaml:"id" validate:"required"`
	State string `yaml:"state,omitempty"`
}
