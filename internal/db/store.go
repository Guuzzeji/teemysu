package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"sync"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

import "C"

//go:embed scripts/init.sql
var initSQL string

var vecOnce sync.Once

type Store struct{ db *sql.DB }

type Mark struct {
	MarkID         int64
	MsgContent     string
	DiscordMsgID   string
	MsgReferenceID *int64
	CreatedAt      time.Time
}

type ChatSummary struct {
	SessionID int64
	LastMsgID *int64
	Summary   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChatMessage struct {
	MsgID      int64
	SessionID  int64
	MsgContent string
	Role       string
	CreatedAt  time.Time
}

type Tag struct {
	TagID      int64
	ContentID  int64
	ContentLoc string
	Type       string
	CreatedAt  time.Time
}

type VectorResult struct {
	ContentLoc string
	Distance   float64
}

func New(path string) (*Store, error) {
	vecOnce.Do(func() { sqlite_vec.Auto() })

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(initSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) SaveMark(ctx context.Context, content, discordMsgID string, referenceID *int64) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO mark_table (msg_content, discord_msg_id, msg_reference_id) VALUES (?, ?, ?)`,
		content, discordMsgID, referenceID,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) GetMark(ctx context.Context, markID int64) (Mark, error) {
	var m Mark
	err := s.db.QueryRowContext(ctx,
		`SELECT mark_id, msg_content, discord_msg_id, msg_reference_id, created_at FROM mark_table WHERE mark_id = ?`,
		markID,
	).Scan(&m.MarkID, &m.MsgContent, &m.DiscordMsgID, &m.MsgReferenceID, &m.CreatedAt)
	return m, err
}

func (s *Store) UpdateMark(ctx context.Context, m Mark) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE mark_table SET msg_content = ?, discord_msg_id = ?, msg_reference_id = ? WHERE mark_id = ?`,
		m.MsgContent, m.DiscordMsgID, m.MsgReferenceID, m.MarkID,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) CreateChatSession(ctx context.Context, summary string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO chat_summary_table (summary) VALUES (?)`,
		summary,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) GetChatSummary(ctx context.Context, sessionID int64) (ChatSummary, error) {
	var cs ChatSummary
	err := s.db.QueryRowContext(ctx,
		`SELECT session_id, last_msg_id, summary, created_at, updated_at FROM chat_summary_table WHERE session_id = ?`,
		sessionID,
	).Scan(&cs.SessionID, &cs.LastMsgID, &cs.Summary, &cs.CreatedAt, &cs.UpdatedAt)
	return cs, err
}

func (s *Store) UpdateChatSummary(ctx context.Context, sessionID int64, lastMsgID *int64, summary string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE chat_summary_table SET last_msg_id = ?, summary = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`,
		lastMsgID, summary, sessionID,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) SaveChatMessage(ctx context.Context, sessionID int64, role, content string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO chat_msg_table (session_id, msg_content, role) VALUES (?, ?, ?)`,
		sessionID, content, role,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) GetChatMessages(ctx context.Context, sessionID int64) ([]ChatMessage, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT msg_id, session_id, msg_content, role, created_at FROM chat_msg_table WHERE session_id = ? ORDER BY created_at ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.MsgID, &m.SessionID, &m.MsgContent, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (s *Store) SaveTag(ctx context.Context, contentID int64, contentLoc, tagType string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO tag_table (content_id, content_loc, type) VALUES (?, ?, ?)`,
		contentID, contentLoc, tagType,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) GetTagsByLoc(ctx context.Context, contentLoc string) ([]Tag, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT tag_id, content_id, content_loc, type, created_at FROM tag_table WHERE content_loc = ?`,
		contentLoc,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.TagID, &t.ContentID, &t.ContentLoc, &t.Type, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (s *Store) SaveVector(ctx context.Context, contentLoc string, embedding []float32) error {
	blob, err := sqlite_vec.SerializeFloat32(embedding)
	if err != nil {
		return fmt.Errorf("serialize embedding: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM vec_table WHERE content_loc = ?`, contentLoc); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO vec_table(content_loc, embedding) VALUES (?, ?)`,
		contentLoc, blob,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) SearchVectors(ctx context.Context, embedding []float32, k int) ([]VectorResult, error) {
	blob, err := sqlite_vec.SerializeFloat32(embedding)
	if err != nil {
		return nil, fmt.Errorf("serialize query: %w", err)
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT content_loc, distance FROM vec_table WHERE embedding MATCH ? ORDER BY distance LIMIT ?`,
		blob, k,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []VectorResult
	for rows.Next() {
		var r VectorResult
		if err := rows.Scan(&r.ContentLoc, &r.Distance); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
