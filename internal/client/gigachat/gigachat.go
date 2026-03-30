package gigachat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/theluckymadman/memotogo/internal/client/oauth"
	"github.com/theluckymadman/memotogo/internal/deps"
	"go.uber.org/zap"
)

type GigaChat struct {
	URL   string
	Model string
	Oauth *oauth.OAuth
	deps  *deps.Deps
}

func NewGigaChat(url string, model string, oauth *oauth.OAuth, deps *deps.Deps) *GigaChat {
	return &GigaChat{URL: url, Model: model, Oauth: oauth, deps: deps}
}

func (g *GigaChat) Chat(ctx context.Context, messages []ReqMessage, funcs []Function) (*ChatResponse, error) {
	logger := g.deps.Logger
	logger.Info("GigaChat.Chat")

	token, err := g.Oauth.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("GigaChat.Chat: get token: %w", err)
	}

	chatReq := ChatRequest{
		Model:          g.Model,
		Messages:       messages,
		Stream:         false,
		UpdateInterval: 0,
	}

	if len(funcs) > 0 {
		logger.Info("GigaChat.Chat: add tools", zap.Any("tools", funcs))
		chatReq.Functions = funcs
	}

	var reqBody bytes.Buffer
	e := json.NewEncoder(&reqBody)
	err = e.Encode(&chatReq)
	if err != nil {
		return nil, fmt.Errorf("GigaChat.Chat: encode body: %w", err)
	}

	logger.Info("GigaChat.Chat: prepared request to LLM", zap.String("body", reqBody.String()))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.URL+"/chat/completions", &reqBody)
	if err != nil {
		return nil, fmt.Errorf("GigaChat.Chat: create request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token.Token)

	resp, err := g.deps.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GigaChat.Chat: get response: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logger.Info("GigaChat response", zap.String("body", string(respBody)))

	if resp.StatusCode != http.StatusOK {
		respBody, _ = io.ReadAll(resp.Body)
		logger.Info("GigaChat", zap.String("not ok response", string(respBody)))
		return nil, fmt.Errorf("GigaChat.Chat: status %d: %s", resp.StatusCode, respBody)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	logger.Info("GigaChat", zap.Any("ok response", chatResp.Choices[0].Message.Content))

	return &chatResp, nil
}
