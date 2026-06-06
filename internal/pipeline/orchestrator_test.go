package pipeline

import (
	"context"
	"testing"

	"github.com/JunLang-7/novel2script/internal/llm"
	"github.com/JunLang-7/novel2script/internal/models"
	"github.com/JunLang-7/novel2script/internal/storage"
)

// mockCache 实现 storage.Cache 接口，用于测试断点续传逻辑。
type mockCache struct {
	chars    map[string][]models.Character
	chunks   map[string]*storage.ChunkResult
	complete map[string]bool
}

func newMockCache() *mockCache {
	return &mockCache{
		chars:    make(map[string][]models.Character),
		chunks:   make(map[string]*storage.ChunkResult),
		complete: make(map[string]bool),
	}
}

func (m *mockCache) PutChunkResult(ctx context.Context, chunkID string, result *storage.ChunkResult) error {
	m.chunks[chunkID] = result
	return nil
}

func (m *mockCache) GetChunkResult(ctx context.Context, chunkID string) (*storage.ChunkResult, bool, error) {
	r, ok := m.chunks[chunkID]
	return r, ok, nil
}

func (m *mockCache) PutCharacterRegistry(ctx context.Context, hash string, chars []models.Character) error {
	m.chars[hash] = chars
	return nil
}

func (m *mockCache) GetCharacterRegistry(ctx context.Context, hash string) ([]models.Character, bool, error) {
	c, ok := m.chars[hash]
	return c, ok, nil
}

func (m *mockCache) MarkComplete(ctx context.Context, chunkID string) error {
	m.complete[chunkID] = true
	return nil
}

func (m *mockCache) IsComplete(ctx context.Context, chunkID string) (bool, error) {
	return m.complete[chunkID], nil
}

func (m *mockCache) ListCompleted(ctx context.Context) ([]string, error) {
	var ids []string
	for id := range m.complete {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *mockCache) Clear(ctx context.Context) error {
	m.chars = make(map[string][]models.Character)
	m.chunks = make(map[string]*storage.ChunkResult)
	m.complete = make(map[string]bool)
	return nil
}

func (m *mockCache) Close() error { return nil }

func TestOrchestrator_NewOrchestrator_WithCache(t *testing.T) {
	cache := newMockCache()
	orch := NewOrchestrator(nil, OrchestratorConfig{Cache: cache})
	if orch.cache != cache {
		t.Error("cache not set")
	}
}

func TestOrchestrator_NewOrchestrator_WithoutCache(t *testing.T) {
	orch := NewOrchestrator(nil, OrchestratorConfig{})
	if orch.cache != nil {
		t.Error("cache should be nil when not set")
	}
}

func TestOrchestrator_loadCharacters_Hit(t *testing.T) {
	cache := newMockCache()
	hash := "abc123"
	expected := []models.Character{{Name: "韩立", ID: "char_hanli"}}
	cache.PutCharacterRegistry(context.Background(), hash, expected)

	orch := NewOrchestrator(nil, OrchestratorConfig{Cache: cache})
	chars, ok := orch.loadCharacters(context.Background(), hash)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(chars) != 1 || chars[0].Name != "韩立" {
		t.Errorf("unexpected characters: %+v", chars)
	}
}

func TestOrchestrator_loadCharacters_Miss(t *testing.T) {
	cache := newMockCache()
	orch := NewOrchestrator(nil, OrchestratorConfig{Cache: cache})
	_, ok := orch.loadCharacters(context.Background(), "nonexistent")
	if ok {
		t.Error("expected cache miss")
	}
}

func TestOrchestrator_loadCharacters_NoCache(t *testing.T) {
	orch := NewOrchestrator(nil, OrchestratorConfig{})
	_, ok := orch.loadCharacters(context.Background(), "hash")
	if ok {
		t.Error("expected false when no cache configured")
	}
}

func TestOrchestrator_saveCharacters(t *testing.T) {
	cache := newMockCache()
	hash := "abc123"
	chars := []models.Character{{Name: "韩立"}}
	orch := NewOrchestrator(nil, OrchestratorConfig{Cache: cache})
	orch.saveCharacters(context.Background(), hash, chars)

	cached, ok, _ := cache.GetCharacterRegistry(context.Background(), hash)
	if !ok {
		t.Fatal("characters not saved")
	}
	if len(cached) != 1 {
		t.Errorf("expected 1 character, got %d", len(cached))
	}
}

func TestOrchestrator_saveCharacters_NoCache(t *testing.T) {
	orch := NewOrchestrator(nil, OrchestratorConfig{})
	orch.saveCharacters(context.Background(), "hash", nil)
	// should not panic
}

func TestOrchestrator_isChunkCached(t *testing.T) {
	cache := newMockCache()
	cache.MarkComplete(context.Background(), "chunk_001")
	orch := NewOrchestrator(nil, OrchestratorConfig{Cache: cache})

	if !orch.isChunkCached(context.Background(), "chunk_001") {
		t.Error("chunk_001 should be cached")
	}
	if orch.isChunkCached(context.Background(), "chunk_002") {
		t.Error("chunk_002 should not be cached")
	}
}

func TestOrchestrator_isChunkCached_NoCache(t *testing.T) {
	orch := NewOrchestrator(nil, OrchestratorConfig{})
	if orch.isChunkCached(context.Background(), "chunk_001") {
		t.Error("should return false when no cache")
	}
}

func TestOrchestrator_loadChunkScenes(t *testing.T) {
	cache := newMockCache()
	cache.PutChunkResult(context.Background(), "chunk_001", &storage.ChunkResult{
		Scenes: []models.Scene{{Title: "测试场景"}},
	})
	orch := NewOrchestrator(nil, OrchestratorConfig{Cache: cache})

	scenes, ok := orch.loadChunkScenes(context.Background(), "chunk_001")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(scenes) != 1 || scenes[0].Title != "测试场景" {
		t.Errorf("unexpected scenes: %+v", scenes)
	}
}

func TestOrchestrator_saveChunkResult(t *testing.T) {
	cache := newMockCache()
	orch := NewOrchestrator(nil, OrchestratorConfig{Cache: cache})
	scenes := []models.Scene{{Title: "保存的场景"}}
	usage := llm.Usage{InputTokens: 100, OutputTokens: 50}

	orch.saveChunkResult(context.Background(), "chunk_001", scenes, usage)

	// 验证场景已保存
	cached, ok, _ := cache.GetChunkResult(context.Background(), "chunk_001")
	if !ok {
		t.Fatal("chunk result not saved")
	}
	if len(cached.Scenes) != 1 || cached.Scenes[0].Title != "保存的场景" {
		t.Error("scenes mismatch")
	}
	if cached.InputTokens != 100 || cached.OutputTokens != 50 {
		t.Error("token usage mismatch")
	}

	// 验证已标记完成
	complete, _ := cache.IsComplete(context.Background(), "chunk_001")
	if !complete {
		t.Error("chunk should be marked complete")
	}
}

func TestOrchestrator_saveChunkResult_NoCache(t *testing.T) {
	orch := NewOrchestrator(nil, OrchestratorConfig{})
	orch.saveChunkResult(context.Background(), "chunk_001", nil, llm.Usage{})
	// should not panic
}
