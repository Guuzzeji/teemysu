-- init.sql: Database initialization script
-- Creates tables and the sqlite-vec virtual table if they do not already exist.
-- Embeddings use 768 dimensions to match the embeddinggemma model.

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- Table 1: tag_table
-- Stores tag-related data with content location and type.
CREATE TABLE IF NOT EXISTS tag_table (
    tag_id      INTEGER PRIMARY KEY AUTOINCREMENT,
    content_id  INTEGER NOT NULL,
    content_loc TEXT    NOT NULL,
    type        TEXT    NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tag_content_id  ON tag_table(content_id);
CREATE INDEX IF NOT EXISTS idx_tag_content_loc ON tag_table(content_loc);

-- Table 2: mark_table
-- Stores bookmark/mark data with Discord message references.
CREATE TABLE IF NOT EXISTS mark_table (
    mark_id            INTEGER PRIMARY KEY AUTOINCREMENT,
    msg_content        TEXT    NOT NULL,
    discord_msg_id     TEXT    NOT NULL UNIQUE,
    channel_id         TEXT,
    msg_reference_id   INTEGER,
    created_at         DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_mark_discord_msg_id ON mark_table(discord_msg_id);
CREATE INDEX IF NOT EXISTS idx_mark_reference_id   ON mark_table(msg_reference_id);

-- Table 3: chat_summary_table
-- Stores summaries for each chat session.
-- discord_thread_id links a Discord thread to its RAG chat session so
-- plain messages inside the thread can continue the conversation.
CREATE TABLE IF NOT EXISTS chat_summary_table (
    session_id          INTEGER PRIMARY KEY AUTOINCREMENT,
    discord_thread_id   TEXT UNIQUE,
    last_msg_id         INTEGER,
    summary             TEXT,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_chat_summary_last_msg_id ON chat_summary_table(last_msg_id);

-- Table 4: chat_msg_table
-- Stores individual chat messages within a session.
CREATE TABLE IF NOT EXISTS chat_msg_table (
    msg_id      INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  INTEGER NOT NULL,
    msg_content TEXT    NOT NULL,
    role        TEXT    NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES chat_summary_table(session_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_msg_session_id ON chat_msg_table(session_id);

-- Table 5: vec_table
-- sqlite-vec virtual table for vector search.
CREATE VIRTUAL TABLE IF NOT EXISTS vec_table USING vec0(
    embedding    float[768],
    +content_loc TEXT
);
