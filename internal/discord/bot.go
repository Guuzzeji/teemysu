package discord

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Guuzzeji/teemysu/internal/ai"
	"github.com/Guuzzeji/teemysu/internal/db"
	"github.com/bwmarrin/discordgo"
)

var logger = log.New(log.Writer(), "[discord] ", log.LstdFlags|log.Lmicroseconds)

const (
	tagSystemPrompt  = "You are a tagging engine. Given the text below, respond with 1 to 5 concise lowercase tags separated by commas. Respond with ONLY the tags, no explanation."
	ragSystemPrompt  = "You are a personal knowledge base assistant. Use the numbered context below to answer the user's question. If the answer isn't in the context, just say so. Reply like a quick Discord text message: clear, direct, and no fluff. Keep your response to 5 sentences or less unless the user explicitly asks for more. Avoid polite AI filler. Context:\n"
	searchResultK    = 10
	ragContextK      = 5
	chatHistoryLimit = 10
)

type Bot struct {
	store *db.Store
	ai    *ai.Client
	s     *discordgo.Session
}

func New(token string, store *db.Store, aiClient *ai.Client) (*Bot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	s.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentMessageContent

	b := &Bot{store: store, ai: aiClient, s: s}
	s.AddHandler(b.onMessageCreate)
	return b, nil
}

func (b *Bot) Open() error  { return b.s.Open() }
func (b *Bot) Close() error { return b.s.Close() }

func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.ID == s.State.User.ID {
		return
	}
	ctx := context.Background()

	if cs, err := b.store.GetChatSessionByThread(ctx, m.ChannelID); err == nil {
		logger.Printf("thread msg: channel=%s session=%d author=%s", m.ChannelID, cs.SessionID, m.Author.Username)
		b.chatInThread(ctx, m, cs)
		return
	}

	cmd, args, ok := parseCommand(m.Content)
	if !ok {
		return
	}
	logger.Printf("cmd=%s author=%s channel=%s args=%q", cmd, m.Author.Username, m.ChannelID, args)
	switch cmd {
	case "!b", "!bookmark":
		b.bookmark(ctx, m, args)
	case "!b-auto", "!bookmark-auto":
		b.bookmarkAuto(ctx, m, args)
	case "!bi", "!bookmark-inhert", "!bookmark-inherit":
		b.bookmarkInherit(ctx, m, args)
	case "!s", "!search":
		b.search(ctx, m, args)
	case "!chat":
		b.chat(ctx, m, args)
	case "!h", "!help":
		b.help(m)
	}
}

func (b *Bot) help(m *discordgo.MessageCreate) {
	_, _ = b.s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Title: "Commands",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "!b / !bookmark", Value: "Save a bookmark with comma-separated tags.\nExample: `!b project,idea,golang http://example.com my notes`"},
			{Name: "!b-auto / !bookmark-auto", Value: "Save a bookmark; the AI picks 1-5 tags for you.\nExample: `!b-auto http://golang.com/example/ftp`"},
			{Name: "!bi / !bookmark-inherit", Value: "Reply to a bookmark and your message is saved, linked to it, and inherits all its tags.\nExample: reply to a bookmark with: `!bi this is a good website to keep in mind`"},
			{Name: "!s / !search", Value: "Semantic search over your bookmarks, top 10 results with links.\nExample: `!s golang ftp libraries`"},
			{Name: "!chat", Value: "Start a RAG chat thread. Reply to the thread to keep chatting.\nExample: `!chat how do I build a golang ftp server?`"},
			{Name: "Reply to a bookmark", Value: "Replying to a `!b`/`!b-auto` message inherits its tags and links the two bookmarks."},
			{Name: "!h / !help", Value: "This list."},
		},
	})
}

// bookmark handles !b / !bookmark: manual comma-separated tags + text to embed.
func (b *Bot) bookmark(ctx context.Context, m *discordgo.MessageCreate, args string) {
	tags, text, err := parseBookmarkArgs(args)
	if err != nil {
		b.reply(m, "usage: !b tag1,tag2 <text>")
		return
	}
	b.saveBookmark(ctx, m, text, tags, false)
}

// bookmarkAuto handles !b-auto / !bookmark-auto: LLM picks 1-5 tags for the text.
func (b *Bot) bookmarkAuto(ctx context.Context, m *discordgo.MessageCreate, args string) {
	text := strings.TrimSpace(args)
	if text == "" {
		b.reply(m, "usage: !b-auto <text>")
		return
	}
	tags, err := b.generateTags(ctx, text)
	if err != nil {
		b.reply(m, "failed to generate tags: "+err.Error())
		return
	}
	b.saveBookmark(ctx, m, text, tags, false)
}

// bookmarkInherit handles !bi / !bookmark-inherit: reply to a bookmark,
// save the reply text and inherit every tag from the replied-to bookmark.
func (b *Bot) bookmarkInherit(ctx context.Context, m *discordgo.MessageCreate, args string) {
	text := strings.TrimSpace(args)
	if text == "" {
		b.reply(m, "usage: reply to a bookmark with: !bi <text to save>")
		return
	}
	b.saveBookmark(ctx, m, text, nil, true)
}

func (b *Bot) saveBookmark(ctx context.Context, m *discordgo.MessageCreate, text string, tags []string, requireInherit bool) {
	var refID *int64
	inherited := []string{}
	if m.ReferencedMessage != nil {
		if parent, err := b.store.GetMarkByDiscordID(ctx, m.ReferencedMessage.ID); err == nil {
			refID = &parent.MarkID
			if ptags, err := b.store.GetTagsByLoc(ctx, m.ReferencedMessage.ID); err == nil {
				for _, t := range ptags {
					inherited = append(inherited, t.Type)
				}
			}
		}
	}
	if requireInherit && refID == nil {
		b.reply(m, "!bi: reply to a bookmark message to inherit its tags")
		return
	}
	allTags := mergeTags(inherited, tags)

	markID, err := b.store.SaveMark(ctx, text, m.ID, m.ChannelID, refID)
	if err != nil {
		b.reply(m, "failed to save bookmark: "+err.Error())
		return
	}
	for _, t := range allTags {
		if _, err := b.store.SaveTag(ctx, markID, m.ID, t); err != nil {
			b.reply(m, "failed to save tag: "+err.Error())
			return
		}
	}
	if vec, err := b.embedText(ctx, text); err == nil {
		_ = b.store.SaveVector(ctx, m.ID, vec)
	} else {
		logger.Printf("saveBookmark: embed failed: %v", err)
	}
	_ = b.s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
	logger.Printf("bookmark saved: mark=%d msg=%s tags=%v inherited=%v ref=%v", markID, m.ID, allTags, inherited, refID)
}

// search handles !s / !search: vector search, reply with the top 10 as jump links.
func (b *Bot) search(ctx context.Context, m *discordgo.MessageCreate, args string) {
	query := strings.TrimSpace(args)
	if query == "" {
		b.reply(m, "usage: !search <query>")
		return
	}
	vec, err := b.embedText(ctx, query)
	if err != nil {
		b.reply(m, "search failed: "+err.Error())
		return
	}
	results, err := b.store.SearchVectors(ctx, vec, searchResultK)
	if err != nil {
		b.reply(m, "search failed: "+err.Error())
		return
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(results))
	for _, r := range results {
		mark, err := b.store.GetMarkByDiscordID(ctx, r.ContentLoc)
		if err != nil {
			continue
		}
		tags, _ := b.store.GetTagsByLoc(ctx, r.ContentLoc)
		value := b.jumpLink(m, r.ContentLoc, mark.ChannelID) + "\n" + truncate(mark.MsgContent, 800)
		if len(tags) > 0 {
			value = "tags: " + strings.Join(tagTypes(tags), ", ") + "\n" + value
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%d. score %.3f", len(fields)+1, r.Distance),
			Value: value,
		})
	}
	if len(fields) == 0 {
		b.reply(m, "no results found")
		logger.Printf("search: query=%q results=0", query)
		return
	}
	_, err = b.s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Title:  "Search: " + query,
		Fields: fields,
		Color:  0x5865F2,
	})
	if err != nil {
		b.reply(m, "failed to send results: "+err.Error())
	}
	logger.Printf("search: query=%q results=%d", query, len(fields))
}

// chat handles !chat: creates a thread, bookmarks it, and starts a RAG conversation.
func (b *Bot) chat(ctx context.Context, m *discordgo.MessageCreate, args string) {
	query := strings.TrimSpace(args)
	if query == "" {
		b.reply(m, "usage: !chat <your question>")
		return
	}

	thread, err := b.s.MessageThreadStart(m.ChannelID, m.ID, truncate(query, 100), 60)
	if err != nil {
		b.reply(m, "failed to create thread: "+err.Error())
		return
	}

	sessionID, err := b.store.CreateChatSession(ctx, thread.ID, query)
	if err != nil {
		b.reply(m, "failed to create chat session: "+err.Error())
		return
	}
	logger.Printf("chat session: thread=%s session=%d query=%q", thread.ID, sessionID, query)

	_ = b.saveChatMessage(ctx, sessionID, "user", query)
	reply, err := b.ragAnswer(ctx, query, nil)
	if err != nil {
		reply = "sorry, I could not answer: " + err.Error()
		logger.Printf("chat reply failed: %v", err)
	}
	_, _ = b.s.ChannelMessageSend(thread.ID, reply)
	_ = b.saveChatMessage(ctx, sessionID, "assistant", reply)

	markID, err := b.store.SaveMark(ctx, query, thread.ID, m.ChannelID, nil)
	if err == nil {
		for _, t := range tagsFromQuery(query) {
			_, _ = b.store.SaveTag(ctx, markID, thread.ID, t)
		}
	}
	if vec, err := b.embedText(ctx, query); err == nil {
		_ = b.store.SaveVector(ctx, thread.ID, vec)
	}
}

// chatInThread answers plain messages inside a known chat thread without any command.
func (b *Bot) chatInThread(ctx context.Context, m *discordgo.MessageCreate, cs db.ChatSummary) {
	content := strings.TrimSpace(m.Content)
	if content == "" || strings.HasPrefix(content, "!") {
		return
	}
	history, err := b.store.GetChatMessages(ctx, cs.SessionID)
	if err != nil {
		b.reply(m, "failed to load chat history: "+err.Error())
		return
	}
	reply, err := b.ragAnswer(ctx, content, history)
	if err != nil {
		b.reply(m, "sorry, I could not answer: "+err.Error())
		logger.Printf("chatInThread reply failed: %v", err)
		return
	}
	_ = b.saveChatMessage(ctx, cs.SessionID, "user", content)
	_, _ = b.s.ChannelMessageSend(m.ChannelID, reply)
	_ = b.saveChatMessage(ctx, cs.SessionID, "assistant", reply)
	logger.Printf("chatInThread answered: session=%d content=%q", cs.SessionID, truncate(content, 80))
}

func (b *Bot) ragAnswer(ctx context.Context, query string, history []db.ChatMessage) (string, error) {
	contextStr, _ := b.ragContext(ctx, query)
	msgs := []ai.Message{{Role: ai.RoleSystem, Content: ragSystemPrompt + contextStr}}
	for _, h := range lastN(history, chatHistoryLimit) {
		role := ai.RoleUser
		if h.Role == "assistant" {
			role = ai.RoleAssistant
		}
		msgs = append(msgs, ai.Message{Role: role, Content: h.MsgContent})
	}
	msgs = append(msgs, ai.Message{Role: ai.RoleUser, Content: query})
	return b.ai.Chat(ctx, msgs)
}

func (b *Bot) ragContext(ctx context.Context, query string) (string, error) {
	vec, err := b.embedText(ctx, query)
	if err != nil {
		return "", err
	}
	results, err := b.store.SearchVectors(ctx, vec, ragContextK)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for i, r := range results {
		mark, err := b.store.GetMarkByDiscordID(ctx, r.ContentLoc)
		if err != nil {
			continue
		}
		tags, _ := b.store.GetTagsByLoc(ctx, r.ContentLoc)
		fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, strings.Join(tagTypes(tags), ","), mark.MsgContent)
	}
	return sb.String(), nil
}

func (b *Bot) generateTags(ctx context.Context, text string) ([]string, error) {
	system := tagSystemPrompt
	if baseline, err := b.store.GetTagBaseline(ctx); err == nil && len(baseline) > 0 {
		var sb strings.Builder
		sb.WriteString("\n\nPrefer reusing these existing tags when they fit; only create new tags if none fit.\nExisting tags with a sample of the text tagged with each:\n")
		for _, ex := range baseline {
			fmt.Fprintf(&sb, "- %s: %s\n", ex.Tag, truncate(ex.Example, 100))
		}
		system += sb.String()
	}

	resp, err := b.ai.Chat(ctx, []ai.Message{
		{Role: ai.RoleSystem, Content: system},
		{Role: ai.RoleUser, Content: text},
	})
	if err != nil {
		return nil, err
	}
	return parseTagResponse(resp), nil
}

func (b *Bot) embedText(ctx context.Context, text string) ([]float32, error) {
	emb, err := b.ai.Embed(ctx, text)
	if err != nil {
		return nil, err
	}
	vec := make([]float32, len(emb))
	for i, f := range emb {
		vec[i] = float32(f)
	}
	return vec, nil
}

func (b *Bot) saveChatMessage(ctx context.Context, sessionID int64, role, content string) error {
	_, err := b.store.SaveChatMessage(ctx, sessionID, role, content)
	return err
}

func (b *Bot) jumpLink(m *discordgo.MessageCreate, msgID, channelID string) string {
	guildID := m.GuildID
	if ch, err := b.s.State.Channel(channelID); err == nil && ch.GuildID != "" {
		guildID = ch.GuildID
	}
	if guildID == "" {
		guildID = "@me"
	}
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, msgID)
}

func (b *Bot) reply(m *discordgo.MessageCreate, content string) {
	_, _ = b.s.ChannelMessageSend(m.ChannelID, content)
}
