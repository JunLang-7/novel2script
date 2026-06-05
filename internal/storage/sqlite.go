package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JunLang-7/novel2script/internal/models"

	_ "modernc.org/sqlite"
)

// SQLiteCache 基于 SQLite 的缓存实现。
type SQLiteCache struct {
	db *sql.DB
}

// NewSQLiteCache 创建 SQLite 缓存。
func NewSQLiteCache(cacheDir string) (*SQLiteCache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	dbPath := filepath.Join(cacheDir, "novel2script_cache.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("打开SQLite失败: %w", err)
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteCache{db: db}, nil
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS chunk_results (
		chunk_id TEXT PRIMARY KEY,
		scenes_json TEXT,
		characters_json TEXT,
		input_tokens INTEGER DEFAULT 0,
		output_tokens INTEGER DEFAULT 0,
		completed INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS character_registry (
		novel_hash TEXT PRIMARY KEY,
		characters_json TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_chunk_completed ON chunk_results(completed);
	`
	_, err := db.Exec(schema)
	return err
}

func (c *SQLiteCache) PutChunkResult(ctx context.Context, chunkID string, result *ChunkResult) error {
	scenesJSON, err := json.Marshal(result.Scenes)
	if err != nil {
		return err
	}
	charsJSON, err := json.Marshal(result.Characters)
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO chunk_results (chunk_id, scenes_json, characters_json, input_tokens, output_tokens)
		 VALUES (?, ?, ?, ?, ?)`,
		chunkID, string(scenesJSON), string(charsJSON), result.InputTokens, result.OutputTokens,
	)
	return err
}

func (c *SQLiteCache) GetChunkResult(ctx context.Context, chunkID string) (*ChunkResult, bool, error) {
	row := c.db.QueryRowContext(ctx,
		`SELECT scenes_json, characters_json, input_tokens, output_tokens
		 FROM chunk_results WHERE chunk_id = ?`, chunkID)

	var scenesJSON, charsJSON string
	var result ChunkResult
	err := row.Scan(&scenesJSON, &charsJSON, &result.InputTokens, &result.OutputTokens)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	if err := json.Unmarshal([]byte(scenesJSON), &result.Scenes); err != nil {
		return nil, false, err
	}
	if err := json.Unmarshal([]byte(charsJSON), &result.Characters); err != nil {
		return nil, false, err
	}

	return &result, true, nil
}

func (c *SQLiteCache) PutCharacterRegistry(ctx context.Context, novelHash string, characters []models.Character) error {
	data, err := json.Marshal(characters)
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO character_registry (novel_hash, characters_json) VALUES (?, ?)`,
		novelHash, string(data),
	)
	return err
}

func (c *SQLiteCache) GetCharacterRegistry(ctx context.Context, novelHash string) ([]models.Character, bool, error) {
	row := c.db.QueryRowContext(ctx,
		`SELECT characters_json FROM character_registry WHERE novel_hash = ?`, novelHash)

	var data string
	err := row.Scan(&data)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var chars []models.Character
	if err := json.Unmarshal([]byte(data), &chars); err != nil {
		return nil, false, err
	}
	return chars, true, nil
}

func (c *SQLiteCache) MarkComplete(ctx context.Context, chunkID string) error {
	_, err := c.db.ExecContext(ctx,
		`UPDATE chunk_results SET completed = 1 WHERE chunk_id = ?`, chunkID)
	return err
}

func (c *SQLiteCache) IsComplete(ctx context.Context, chunkID string) (bool, error) {
	row := c.db.QueryRowContext(ctx,
		`SELECT completed FROM chunk_results WHERE chunk_id = ?`, chunkID)

	var completed int
	err := row.Scan(&completed)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return completed == 1, nil
}

func (c *SQLiteCache) ListCompleted(ctx context.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx,
		`SELECT chunk_id FROM chunk_results WHERE completed = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (c *SQLiteCache) Clear(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, `DELETE FROM chunk_results`)
	if err != nil {
		return err
	}
	_, err = c.db.ExecContext(ctx, `DELETE FROM character_registry`)
	return err
}

func (c *SQLiteCache) Close() error {
	return c.db.Close()
}

