package storage

import (
	"context"

	"github.com/JunLang-7/novel2script/internal/models"
)

// Cache 定义中间结果缓存接口，用于断点续传。
type Cache interface {
	// PutChunkResult 缓存一个 chunk 的处理结果。
	PutChunkResult(ctx context.Context, chunkID string, result *ChunkResult) error

	// GetChunkResult 获取已缓存的 chunk 处理结果。
	GetChunkResult(ctx context.Context, chunkID string) (*ChunkResult, bool, error)

	// PutCharacterRegistry 缓存合并后的角色表。
	PutCharacterRegistry(ctx context.Context, novelHash string, characters []models.Character) error

	// GetCharacterRegistry 获取已缓存的角色表。
	GetCharacterRegistry(ctx context.Context, novelHash string) ([]models.Character, bool, error)

	// MarkComplete 标记指定 chunk 处理完成。
	MarkComplete(ctx context.Context, chunkID string) error

	// IsComplete 检查指定 chunk 是否处理完成。
	IsComplete(ctx context.Context, chunkID string) (bool, error)

	// ListCompleted 列出已完成的 chunk ID 列表。
	ListCompleted(ctx context.Context) ([]string, error)

	// Clear 清除所有缓存。
	Clear(ctx context.Context) error

	// Close 关闭缓存连接。
	Close() error
}

// ChunkResult 一个 chunk 的分析结果。
type ChunkResult struct {
	Scenes      []models.Scene       `json:"scenes"`
	Characters  []models.Character   `json:"characters"`
	InputTokens  int                 `json:"input_tokens"`
	OutputTokens int                 `json:"output_tokens"`
}

// NovelHash 为小说文本计算简单哈希（用于缓存键）。
func NovelHash(text string) string {
	// 取文本开头 + 结尾 + 长度的组合作为 hash
	if len(text) <= 200 {
		return hashString(text)
	}
	return hashString(text[:100] + text[len(text)-100:] + itoa(len(text)))
}

func hashString(s string) string {
	var h uint64
	for _, c := range s {
		h = h*31 + uint64(c)
	}
	return fmtUint64(h)
}

func fmtUint64(n uint64) string {
	const hexChars = "0123456789abcdef"
	var buf [16]byte
	for i := 15; i >= 0; i-- {
		buf[i] = hexChars[n&0xf]
		n >>= 4
	}
	return string(buf[:])
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
