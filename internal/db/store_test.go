package db

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestNewCreatesDB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	s, err := New(path)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	_ = s.Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("db file not created: %v", err)
	}
}

func TestMarkCRUD(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	ref := int64(7)
	id, err := s.SaveMark(ctx, "hello world", "msg-1", "ch-1", &ref)
	if err != nil {
		t.Fatalf("save mark: %v", err)
	}

	m, err := s.GetMark(ctx, id)
	if err != nil {
		t.Fatalf("get mark: %v", err)
	}
	if m.MsgContent != "hello world" || m.DiscordMsgID != "msg-1" || m.ChannelID != "ch-1" || m.MsgReferenceID == nil || *m.MsgReferenceID != 7 {
		t.Fatalf("unexpected mark: %+v", m)
	}

	m.MsgContent = "updated"
	if err := s.UpdateMark(ctx, m); err != nil {
		t.Fatalf("update mark: %v", err)
	}

	m2, err := s.GetMark(ctx, id)
	if err != nil {
		t.Fatalf("get updated mark: %v", err)
	}
	if m2.MsgContent != "updated" {
		t.Errorf("content = %q, want updated", m2.MsgContent)
	}
}

func TestChatSession(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sid, err := s.CreateChatSession(ctx, "thread-1", "initial summary")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	_, err = s.SaveChatMessage(ctx, sid, "user", "hi")
	if err != nil {
		t.Fatalf("save message: %v", err)
	}
	mid2, err := s.SaveChatMessage(ctx, sid, "assistant", "hello")
	if err != nil {
		t.Fatalf("save message: %v", err)
	}

	msgs, err := s.GetChatMessages(ctx, sid)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[1].Role != "assistant" {
		t.Errorf("unexpected order/roles: %+v", msgs)
	}

	if err := s.UpdateChatSummary(ctx, sid, &mid2, "final summary"); err != nil {
		t.Fatalf("update summary: %v", err)
	}

	sum, err := s.GetChatSummary(ctx, sid)
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if sum.Summary != "final summary" || sum.LastMsgID == nil || *sum.LastMsgID != mid2 {
		t.Errorf("unexpected summary: %+v", sum)
	}
}

func TestMarkCRUDByDiscordID(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.SaveMark(ctx, "hello", "disc-1", "ch-9", nil)
	if err != nil {
		t.Fatalf("save mark: %v", err)
	}

	m, err := s.GetMarkByDiscordID(ctx, "disc-1")
	if err != nil {
		t.Fatalf("get mark by discord id: %v", err)
	}
	if m.MarkID != id || m.MsgContent != "hello" || m.ChannelID != "ch-9" {
		t.Fatalf("unexpected mark: %+v", m)
	}

	if _, err := s.GetMarkByDiscordID(ctx, "missing"); err == nil {
		t.Fatalf("expected error for unknown discord id")
	}
}

func TestChatSessionThreadLookup(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sid, err := s.CreateChatSession(ctx, "thread-alpha", "summary alpha")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := s.CreateChatSession(ctx, "thread-beta", "summary beta"); err != nil {
		t.Fatalf("create session: %v", err)
	}

	cs, err := s.GetChatSessionByThread(ctx, "thread-alpha")
	if err != nil {
		t.Fatalf("get by thread: %v", err)
	}
	if cs.SessionID != sid || cs.Summary != "summary alpha" {
		t.Fatalf("unexpected summary: %+v", cs)
	}

	if _, err := s.GetChatSessionByThread(ctx, "missing"); err == nil {
		t.Fatalf("expected error for unknown thread")
	}
}

func TestTag(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.SaveTag(ctx, 1, "mark/1", "todo")
	if err != nil {
		t.Fatalf("save tag: %v", err)
	}
	_, err = s.SaveTag(ctx, 1, "mark/1", "important")
	if err != nil {
		t.Fatalf("save tag: %v", err)
	}

	tags, err := s.GetTagsByLoc(ctx, "mark/1")
	if err != nil {
		t.Fatalf("get tags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("got %d tags, want 2", len(tags))
	}
}

func TestTagBaseline(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	m1, err := s.SaveMark(ctx, "building an ftp server in go", "msg-1", "ch-1", nil)
	if err != nil {
		t.Fatalf("save mark 1: %v", err)
	}
	m2, err := s.SaveMark(ctx, "bookmarks app roadmap", "msg-2", "ch-1", nil)
	if err != nil {
		t.Fatalf("save mark 2: %v", err)
	}
	for _, tag := range []string{"golang", "ftp"} {
		if _, err := s.SaveTag(ctx, m1, "msg-1", tag); err != nil {
			t.Fatalf("save tag %q: %v", tag, err)
		}
	}
	if _, err := s.SaveTag(ctx, m2, "msg-2", "golang"); err != nil {
		t.Fatalf("save tag golang: %v", err)
	}

	baseline, err := s.GetTagBaseline(ctx)
	if err != nil {
		t.Fatalf("get tag baseline: %v", err)
	}
	if len(baseline) != 2 {
		t.Fatalf("got %d baseline tags, want 2 unique", len(baseline))
	}
	byTag := map[string]string{}
	for _, ex := range baseline {
		byTag[ex.Tag] = ex.Example
	}
	if byTag["ftp"] != "building an ftp server in go" {
		t.Errorf("ftp example = %q", byTag["ftp"])
	}
	if byTag["golang"] == "" {
		t.Errorf("golang example missing")
	}
}

func TestVectorSearch(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	const dim = 768
	a := make([]float32, dim)
	a[0] = 1

	b := make([]float32, dim)
	b[1] = 1

	if err := s.SaveVector(ctx, "loc-a", a); err != nil {
		t.Fatalf("save vector a: %v", err)
	}
	if err := s.SaveVector(ctx, "loc-b", b); err != nil {
		t.Fatalf("save vector b: %v", err)
	}

	results, err := s.SearchVectors(ctx, a, 2)
	if err != nil {
		t.Fatalf("search vectors: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].ContentLoc != "loc-a" {
		t.Errorf("first result = %q, want loc-a", results[0].ContentLoc)
	}
	if math.Abs(results[0].Distance) > 1e-6 {
		t.Errorf("distance to self = %v, want ~0", results[0].Distance)
	}
	if math.Abs(results[1].Distance-math.Sqrt(2)) > 1e-3 {
		t.Errorf("distance to b = %v, want ~sqrt(2)", results[1].Distance)
	}
}
