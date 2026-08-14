<p align="center">

<img src="./img/Tiramisu-Cake-2.jpg" alt="Teemysu">

</p>

<h1 align="center">Teemysu</h1>

Discord bot for bookmarking and semantic search. Drop links, notes, or questions. Bot handles organization.

## How It Works

1. **Save bookmarks** with manual or AI-generated tags
2. **Search** using natural language (vector embeddings)
3. **Chat** with RAG - bot answers from your saved bookmarks

SQLite stores everything. OpenAI-compatible API handles embeddings and chat.

## Setup

### Prerequisites

- Go 1.26+
- OpenAI-compatible API (Ollama, llama.cpp, OpenAI)
- Discord bot token (see below)

### Discord Bot Setup

1. Go to [discord.com/developers/applications](https://discord.com/developers/applications) → **New Application** → name it → **Create**
2. Bot user enabled by default. Copy **Application ID** from General Information
3. Left sidebar → **Bot** → **Reset Token** → copy token (store in `.env`, never commit)

**Enable Privileged Intent** (required to read message content):

On the **Bot** page → **Privileged Gateway Intents** → toggle **Message Content Intent** ON.

Without this, message content arrives empty and the bot can't read your commands.

**Generate invite URL**:

Go to **OAuth2** → **URL Generator** → select:

- Scopes: `bot`, `applications.commands`
- Bot Permissions: `View Channels`, `Send Messages`, `Read Message History`, `Embed Links`, `Attach Files`, `Add Reactions`

Copy the generated URL → open in browser → add bot to your server.

### Configure

```bash
cp .env.example .env
```

Edit `.env`:

```env
DISCORD_BOT_TOKEN=your_token_here
OPENAI_BASE_URL=http://localhost:11434/v1  # Ollama default
OPENAI_API_KEY=ollama
OPENAI_CHAT_MODEL=gemma3:1b
OPENAI_EMBED_MODEL=embeddinggemma
DATABASE_PATH=./data.db
```

### Running

**Local** (builds binary to `dist/bot`):

```bash
make run-local
```

Or step by step:

```bash
make build-local
make build-run
```

**Docker**:

```bash
make docker-up
```

Logs: `make docker-logs` | Stop: `make docker-down`

Data persists in a named volume. Wipe: `docker compose down -v`.

## Commands

| Command             | Description                      |
| ------------------- | -------------------------------- |
| `!b tag1,tag2 text` | Save bookmark with manual tags   |
| `!b-auto text`      | Save bookmark, AI picks tags     |
| `!bi text`          | Reply to bookmark: inherits tags |
| `!s query`          | Semantic search (top 10)         |
| `!chat question`    | Start RAG chat thread            |
| `!h`                | Show commands                    |

## Dependencies

- [discordgo](https://github.com/bwmarrin/discordgo) - Discord API
- [openai-go](https://github.com/openai/openai-go) - OpenAI SDK
- [sqlite-vec](https://github.com/asg017/sqlite-vec) - Vector search
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver

## License

MIT
