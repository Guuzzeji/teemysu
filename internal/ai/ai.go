package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role
	Content string
}

type api interface {
	Chat(ctx context.Context, model string, msgs []Message) (string, error)
	Embed(ctx context.Context, model, text string) ([]float64, error)
}

type Client struct {
	api        api
	chatModel  string
	embedModel string
}

func resolveModel(override, envVal string) (string, error) {
	if v := strings.TrimSpace(override); v != "" {
		return v, nil
	}
	if v := strings.TrimSpace(envVal); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("model not set")
}

func New() (*Client, error) {
	chatEnv := os.Getenv("OPENAI_CHAT_MODEL")
	embedEnv := os.Getenv("OPENAI_EMBED_MODEL")

	chat, err := resolveModel("", chatEnv)
	if err != nil {
		return nil, fmt.Errorf("OPENAI_CHAT_MODEL not set")
	}
	embed, err := resolveModel("", embedEnv)
	if err != nil {
		return nil, fmt.Errorf("OPENAI_EMBED_MODEL not set")
	}

	c := openai.NewClient()
	return &Client{api: &sdkClient{client: &c}, chatModel: chat, embedModel: embed}, nil
}

func NewWithModels(chatModel, embedModel string) (*Client, error) {
	chatEnv := os.Getenv("OPENAI_CHAT_MODEL")
	embedEnv := os.Getenv("OPENAI_EMBED_MODEL")

	chat, err := resolveModel(chatModel, chatEnv)
	if err != nil {
		return nil, fmt.Errorf("OPENAI_CHAT_MODEL not set")
	}
	embed, err := resolveModel(embedModel, embedEnv)
	if err != nil {
		return nil, fmt.Errorf("OPENAI_EMBED_MODEL not set")
	}

	c := openai.NewClient()
	return &Client{api: &sdkClient{client: &c}, chatModel: chat, embedModel: embed}, nil
}

func (c *Client) ChatModel() string  { return c.chatModel }
func (c *Client) EmbedModel() string { return c.embedModel }

type sdkClient struct {
	client *openai.Client
}

func (s *sdkClient) Chat(ctx context.Context, model string, msgs []Message) (string, error) {
	sdkMsgs := make([]openai.ChatCompletionMessageParamUnion, len(msgs))
	for i, m := range msgs {
		switch m.Role {
		case RoleSystem:
			sdkMsgs[i] = openai.SystemMessage(m.Content)
		case RoleUser:
			sdkMsgs[i] = openai.UserMessage(m.Content)
		case RoleAssistant:
			sdkMsgs[i] = openai.AssistantMessage(m.Content)
		default:
			return "", fmt.Errorf("unknown role: %s", m.Role)
		}
	}

	resp, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: sdkMsgs,
		Model:    model,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}

func (s *sdkClient) Embed(ctx context.Context, model, text string) ([]float64, error) {
	resp, err := s.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Model: model,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}
	return resp.Data[0].Embedding, nil
}
