package analyzers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JunLang-7/novel2script/internal/llm"
)

func TestSceneAnalyzer_Analyze_WithMockLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scenes := []llmScene{
			{
				ID:            "scene_1",
				Title:         "神手谷",
				Sequence:      1,
				Location:      "神手谷",
				LocationType:  "山谷",
				TimeOfDay:     "黄昏",
				Atmosphere:    "宁静",
				Summary:       "韩立在神手谷修炼",
				ChapterSource: 1,
				Mood:          "平缓",
			},
		}
		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{{Text: toJSON(scenes)}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClient(llm.Config{
		Provider:    "anthropic",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  0,
		MaxParallel: 1,
	})

	analyzer := NewSceneAnalyzer(client, 2)

	input := "第一章 初入修仙\n韩立站在神手谷中，望着远处的云雾。\n\n" +
		"第二章 药园初见\n墨大夫打量着韩立。\n\n" +
		"第三章 长春功\n韩立盘膝坐在床上修炼。\n"

	scenes, err := analyzer.Analyze(context.Background(), input)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("expected 1 scene, got %d", len(scenes))
	}
	if scenes[0].Title != "神手谷" {
		t.Errorf("scene title: %q", scenes[0].Title)
	}
	if scenes[0].Setting.Location != "神手谷" {
		t.Errorf("location: %q", scenes[0].Setting.Location)
	}
	if scenes[0].Sequence != 1 {
		t.Errorf("sequence: %d", scenes[0].Sequence)
	}
}

func TestSceneAnalyzer_Analyze_EmptyInput(t *testing.T) {
	client := llm.NewClient(llm.Config{
		Provider:   "anthropic",
		BaseURL:    "http://localhost",
		APIKey:     "test-key",
		MaxRetries: 0,
	})

	analyzer := NewSceneAnalyzer(client, 2)
	_, err := analyzer.Analyze(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestCharacterAnalyzer_Analyze_WithMockLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		characters := []map[string]any{
			{
				"name":              "韩立",
				"role":              "protagonist",
				"description":       "山村少年",
				"importance_rank":   1,
				"traits":            []string{"谨慎", "坚韧"},
				"first_appearance_chapter": 1,
			},
		}
		resp := anthropicResponse{
			Content: []struct {
				Text string `json:"text"`
			}{{Text: toJSON(characters)}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClient(llm.Config{
		Provider:    "anthropic",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		MaxRetries:  0,
		MaxParallel: 1,
	})

	analyzer := NewCharacterAnalyzer(client, 2)

	input := "第一章 初入修仙\n韩立站在神手谷中，望着远处的云雾。\n\n" +
		"第二章 药园初见\n墨大夫打量着韩立。\n\n" +
		"第三章 长春功\n韩立盘膝坐在床上修炼。\n"

	characters, err := analyzer.Analyze(context.Background(), input)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if len(characters) != 1 {
		t.Fatalf("expected 1 character, got %d", len(characters))
	}
	if characters[0].Name != "韩立" {
		t.Errorf("character name: %q", characters[0].Name)
	}
}

func TestCharacterAnalyzer_Analyze_EmptyInput(t *testing.T) {
	client := llm.NewClient(llm.Config{
		Provider:   "anthropic",
		BaseURL:    "http://localhost",
		APIKey:     "test-key",
		MaxRetries: 0,
	})

	analyzer := NewCharacterAnalyzer(client, 2)
	_, err := analyzer.Analyze(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func toJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
