package ai

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

type fakeAPI struct {
	chatCalls  []fakeChatCall
	chatReply  string
	chatErr    error
	embedVals  []float64
	embedErr   error
	embedCalls int
	embedModel string
}

type fakeChatCall struct {
	model string
	msgs  []Message
}

func (f *fakeAPI) Chat(_ context.Context, model string, msgs []Message) (string, error) {
	f.chatCalls = append(f.chatCalls, fakeChatCall{model: model, msgs: msgs})
	if f.chatErr != nil {
		return "", f.chatErr
	}
	return f.chatReply, nil
}

func (f *fakeAPI) Embed(_ context.Context, model, text string) ([]float64, error) {
	f.embedCalls++
	f.embedModel = model
	if f.embedErr != nil {
		return nil, f.embedErr
	}
	return f.embedVals, nil
}

func TestResolveModel(t *testing.T) {
	tests := []struct {
		name    string
		override string
		envVal  string
		want    string
		wantErr string
	}{
		{
			name:     "override wins over env",
			override: "custom-model",
			envVal:   "env-model",
			want:     "custom-model",
		},
		{
			name:    "env used when override empty",
			envVal:  "env-model",
			want:    "env-model",
		},
		{
			name:    "both empty returns error",
			wantErr: "not set",
		},
		{
			name:     "whitespace override trimmed",
			override: "  model-x  ",
			want:     "model-x",
		},
		{
			name:    "whitespace env trimmed",
			envVal:  "  model-y  ",
			want:    "model-y",
		},
		{
			name:     "whitespace-only override falls through",
			override: "   ",
			envVal:   "env-model",
			want:     "env-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveModel(tt.override, tt.envVal)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewMissingEnv(t *testing.T) {
	t.Setenv("OPENAI_CHAT_MODEL", "")
	t.Setenv("OPENAI_EMBED_MODEL", "")
	os.Unsetenv("OPENAI_CHAT_MODEL")
	os.Unsetenv("OPENAI_EMBED_MODEL")

	_, err := New()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "OPENAI_CHAT_MODEL") {
		t.Errorf("error %q should mention OPENAI_CHAT_MODEL", err.Error())
	}
}

func TestNewEnv(t *testing.T) {
	t.Setenv("OPENAI_CHAT_MODEL", "test-chat")
	t.Setenv("OPENAI_EMBED_MODEL", "test-embed")

	c, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ChatModel() != "test-chat" {
		t.Errorf("ChatModel() = %q, want %q", c.ChatModel(), "test-chat")
	}
	if c.EmbedModel() != "test-embed" {
		t.Errorf("EmbedModel() = %q, want %q", c.EmbedModel(), "test-embed")
	}
}

func TestNewWithModels(t *testing.T) {
	// Override wins over env
	t.Setenv("OPENAI_CHAT_MODEL", "env-chat")
	t.Setenv("OPENAI_EMBED_MODEL", "env-embed")

	c, err := NewWithModels("override-chat", "override-embed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ChatModel() != "override-chat" {
		t.Errorf("ChatModel() = %q, want %q", c.ChatModel(), "override-chat")
	}
	if c.EmbedModel() != "override-embed" {
		t.Errorf("EmbedModel() = %q, want %q", c.EmbedModel(), "override-embed")
	}

	// Empty override falls through to env
	c2, err := NewWithModels("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c2.ChatModel() != "env-chat" {
		t.Errorf("ChatModel() = %q, want %q", c2.ChatModel(), "env-chat")
	}

	// Both empty, no env → error
	os.Unsetenv("OPENAI_CHAT_MODEL")
	os.Unsetenv("OPENAI_EMBED_MODEL")
	_, err = NewWithModels("", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetters(t *testing.T) {
	c := &Client{chatModel: "chat-abc", embedModel: "embed-xyz"}
	if c.ChatModel() != "chat-abc" {
		t.Errorf("ChatModel() = %q, want %q", c.ChatModel(), "chat-abc")
	}
	if c.EmbedModel() != "embed-xyz" {
		t.Errorf("EmbedModel() = %q, want %q", c.EmbedModel(), "embed-xyz")
	}
}

func TestChat(t *testing.T) {
	fake := &fakeAPI{chatReply: "hello"}
	c := &Client{api: fake, chatModel: "gpt-test"}

	got, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
	if len(fake.chatCalls) != 1 {
		t.Fatalf("expected 1 chat call, got %d", len(fake.chatCalls))
	}
}

func TestChatEmptyChoices(t *testing.T) {
	fake := &fakeAPI{chatReply: ""}
	c := &Client{api: fake, chatModel: "gpt-test"}

	got, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestChatError(t *testing.T) {
	wantErr := errors.New("api down")
	fake := &fakeAPI{chatErr: wantErr}
	c := &Client{api: fake, chatModel: "gpt-test"}

	_, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("got err %v, want %v", err, wantErr)
	}
}

func TestChatUnknownRole(t *testing.T) {
	sdk := &sdkClient{client: nil}

	_, err := sdk.Chat(context.Background(), "model", []Message{
		{Role: "dragon", Content: "roar"},
	})
	if err == nil {
		t.Fatal("expected error for unknown role, got nil")
	}
	if !strings.Contains(err.Error(), "unknown role") {
		t.Errorf("error %q should contain 'unknown role'", err.Error())
	}
}

func TestChatUsesConfiguredModel(t *testing.T) {
	fake := &fakeAPI{chatReply: "ok"}
	c := &Client{api: fake, chatModel: "my-model"}

	_, err := c.Chat(context.Background(), []Message{
		{Role: RoleUser, Content: "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.chatCalls[0].model != "my-model" {
		t.Errorf("model = %q, want %q", fake.chatCalls[0].model, "my-model")
	}
}

func TestEmbed(t *testing.T) {
	want := []float64{0.1, 0.2, 0.3}
	fake := &fakeAPI{embedVals: want}
	c := &Client{api: fake, embedModel: "test-model"}

	got, err := c.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d floats, want %d", len(got), len(want))
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("got[%d] = %v, want %v", i, v, want[i])
		}
	}
	if fake.embedCalls != 1 {
		t.Errorf("embedCalls = %d, want 1", fake.embedCalls)
	}
}

func TestEmbedEmptyData(t *testing.T) {
	fake := &fakeAPI{embedVals: nil}
	c := &Client{api: fake, embedModel: "test-model"}

	got, err := c.Embed(context.Background(), "hello")
	if err == nil && got == nil {
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d floats, want empty", len(got))
	}
}

func TestEmbedError(t *testing.T) {
	wantErr := errors.New("embed failed")
	fake := &fakeAPI{embedErr: wantErr}
	c := &Client{api: fake, embedModel: "test-model"}

	got, err := c.Embed(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("got err %v, want %v", err, wantErr)
	}
	if len(got) != 0 {
		t.Errorf("got %d floats on error, want 0", len(got))
	}
}

func TestEmbedUsesConfiguredModel(t *testing.T) {
	fake := &fakeAPI{embedVals: []float64{1.0}}
	c := &Client{api: fake, embedModel: "custom-embed-model"}

	_, err := c.Embed(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.embedCalls != 1 {
		t.Fatalf("embedCalls = %d, want 1", fake.embedCalls)
	}
	if fake.embedModel != "custom-embed-model" {
		t.Errorf("model = %q, want %q", fake.embedModel, "custom-embed-model")
	}
}
