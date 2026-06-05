package storage

import (
	"context"
	"os"
	"testing"

	"github.com/JunLang-7/novel2script/internal/models"
)

func newTestCache(t *testing.T) *SQLiteCache {
	t.Helper()
	dir, err := os.MkdirTemp("", "novel2script_test_*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	cache, err := NewSQLiteCache(dir)
	if err != nil {
		t.Fatalf("NewSQLiteCache: %v", err)
	}
	t.Cleanup(func() { cache.Close() })
	return cache
}

func TestSQLiteCache_PutAndGetChunkResult(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	result := &ChunkResult{
		Scenes: []models.Scene{
			{ID: "s1", Title: "场景一", Setting: models.SceneSetting{Location: "山村"}},
		},
		Characters: []models.Character{
			{ID: "c1", Name: "韩立", Role: models.RoleProtagonist},
		},
		InputTokens:  1000,
		OutputTokens: 200,
	}

	err := cache.PutChunkResult(ctx, "chunk_001", result)
	if err != nil {
		t.Fatalf("PutChunkResult: %v", err)
	}

	got, ok, err := cache.GetChunkResult(ctx, "chunk_001")
	if err != nil {
		t.Fatalf("GetChunkResult: %v", err)
	}
	if !ok {
		t.Fatal("expected chunk result to exist")
	}
	if len(got.Scenes) != 1 {
		t.Errorf("expected 1 scene, got %d", len(got.Scenes))
	}
	if got.Scenes[0].Title != "场景一" {
		t.Errorf("scene title mismatch: %s", got.Scenes[0].Title)
	}
	if len(got.Characters) != 1 {
		t.Errorf("expected 1 character, got %d", len(got.Characters))
	}
	if got.InputTokens != 1000 {
		t.Errorf("input tokens: %d", got.InputTokens)
	}
}

func TestSQLiteCache_GetChunkResult_NotFound(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	_, ok, err := cache.GetChunkResult(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetChunkResult: %v", err)
	}
	if ok {
		t.Error("expected not found for nonexistent chunk")
	}
}

func TestSQLiteCache_MarkComplete(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	result := &ChunkResult{}
	cache.PutChunkResult(ctx, "chunk_001", result)

	// Initially not complete
	done, err := cache.IsComplete(ctx, "chunk_001")
	if err != nil {
		t.Fatalf("IsComplete: %v", err)
	}
	if done {
		t.Error("chunk should not be complete initially")
	}

	// Mark complete
	err = cache.MarkComplete(ctx, "chunk_001")
	if err != nil {
		t.Fatalf("MarkComplete: %v", err)
	}

	done, err = cache.IsComplete(ctx, "chunk_001")
	if err != nil {
		t.Fatalf("IsComplete: %v", err)
	}
	if !done {
		t.Error("chunk should be complete after MarkComplete")
	}
}

func TestSQLiteCache_ListCompleted(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	for i := range 3 {
		id := "chunk_" + string(rune('0'+i))
		cache.PutChunkResult(ctx, id, &ChunkResult{})
		if i%2 == 0 {
			cache.MarkComplete(ctx, id)
		}
	}

	completed, err := cache.ListCompleted(ctx)
	if err != nil {
		t.Fatalf("ListCompleted: %v", err)
	}
	if len(completed) != 2 {
		t.Errorf("expected 2 completed chunks, got %d: %v", len(completed), completed)
	}
}

func TestSQLiteCache_CharacterRegistry(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	characters := []models.Character{
		{ID: "c1", Name: "韩立", Role: models.RoleProtagonist},
		{ID: "c2", Name: "南宫婉", Role: models.RoleLoveInterest},
	}

	err := cache.PutCharacterRegistry(ctx, "novel_hash_123", characters)
	if err != nil {
		t.Fatalf("PutCharacterRegistry: %v", err)
	}

	got, ok, err := cache.GetCharacterRegistry(ctx, "novel_hash_123")
	if err != nil {
		t.Fatalf("GetCharacterRegistry: %v", err)
	}
	if !ok {
		t.Fatal("expected registry to exist")
	}
	if len(got) != 2 {
		t.Errorf("expected 2 characters, got %d", len(got))
	}
	if got[0].Name != "韩立" {
		t.Errorf("first character: %s", got[0].Name)
	}
}

func TestSQLiteCache_CharacterRegistry_NotFound(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	_, ok, err := cache.GetCharacterRegistry(ctx, "nonexistent_hash")
	if err != nil {
		t.Fatalf("GetCharacterRegistry: %v", err)
	}
	if ok {
		t.Error("expected not found")
	}
}

func TestSQLiteCache_Clear(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	cache.PutChunkResult(ctx, "chunk_001", &ChunkResult{})
	cache.Clear(ctx)

	_, ok, _ := cache.GetChunkResult(ctx, "chunk_001")
	if ok {
		t.Error("chunk should be cleared")
	}
}

func TestNovelHash(t *testing.T) {
	h1 := NovelHash("测试小说内容")
	h2 := NovelHash("测试小说内容")
	h3 := NovelHash("不同的内容")

	if h1 != h2 {
		t.Errorf("same content should have same hash: %s vs %s", h1, h2)
	}
	if h1 == h3 {
		t.Error("different content should have different hash")
	}
}
