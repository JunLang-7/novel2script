package llm

import (
	"strings"
	"testing"
)

func TestSystemPrompt_NotEmpty(t *testing.T) {
	if SystemPrompt == "" {
		t.Error("SystemPrompt should not be empty")
	}
	if !strings.Contains(SystemPrompt, "剧本改编") {
		t.Error("SystemPrompt should mention 剧本改编")
	}
}

func TestCharacterExtractionPrompt_Placeholders(t *testing.T) {
	if !strings.Contains(CharacterExtractionPrompt, "{text}") {
		t.Error("CharacterExtractionPrompt should contain {text} placeholder")
	}
	if !strings.Contains(CharacterExtractionPrompt, "角色") {
		t.Error("CharacterExtractionPrompt should mention 角色")
	}
}

func TestSceneSegmentationPrompt_Placeholders(t *testing.T) {
	if !strings.Contains(SceneSegmentationPrompt, "{text}") {
		t.Error("SceneSegmentationPrompt should contain {text} placeholder")
	}
	if !strings.Contains(SceneSegmentationPrompt, "场景") {
		t.Error("SceneSegmentationPrompt should mention 场景")
	}
}

func TestScriptConversionPrompt_Placeholders(t *testing.T) {
	placeholders := []string{
		"{character_context}",
		"{scene_title}",
		"{location}",
		"{time}",
		"{characters_present}",
		"{text}",
	}
	for _, ph := range placeholders {
		if !strings.Contains(ScriptConversionPrompt, ph) {
			t.Errorf("ScriptConversionPrompt missing placeholder: %s", ph)
		}
	}
}

func TestSynopsisPrompt_NotEmpty(t *testing.T) {
	if SynopsisPrompt == "" {
		t.Error("SynopsisPrompt should not be empty")
	}
	if !strings.Contains(SynopsisPrompt, "{text}") {
		t.Error("SynopsisPrompt should contain {text}")
	}
}

func TestPlotAnalysisPrompt_NotEmpty(t *testing.T) {
	if PlotAnalysisPrompt == "" {
		t.Error("PlotAnalysisPrompt should not be empty")
	}
	if !strings.Contains(PlotAnalysisPrompt, "{text}") {
		t.Error("PlotAnalysisPrompt should contain {text}")
	}
}

func TestPromptTemplate_Substitution(t *testing.T) {
	// Verify that string replacement works correctly for all templates
	templates := map[string]string{
		"character_extraction": CharacterExtractionPrompt,
		"scene_segmentation":   SceneSegmentationPrompt,
		"script_conversion":    ScriptConversionPrompt,
	}

	testText := "<<TEST_INPUT>>"

	for name, tmpl := range templates {
		result := strings.Replace(tmpl, "{text}", testText, 1)
		if !strings.Contains(result, testText) {
			t.Errorf("%s: substitution failed", name)
		}
		if strings.Contains(result, "{text}") && name != "script_conversion" {
			// For script_conversion, it's OK to still have {text} if there are multiple
			// For others, there's only one {text}
			t.Logf("%s: {text} still present after substitution", name)
		}
	}
}
