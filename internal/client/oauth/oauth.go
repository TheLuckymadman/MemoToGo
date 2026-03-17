package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/theluckymadman/memotogo/internal/deps"
	"go.uber.org/zap"
)

func NewOAuth(
	oauthURL string,
	authKey string,
	scope string,
	refreshTokenBeforeExp time.Duration,
	deps *deps.Deps,
) *OAuth {
	return &OAuth{
		OAuthURL:              oauthURL,
		AuthKey:               authKey,
		Scope:                 scope,
		RefreshTokenBeforeExp: refreshTokenBeforeExp,
		deps:                  deps,
	}
}

func (o *OAuth) GetToken(ctx context.Context) (Token, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	logger := o.deps.Logger
	logger.Info("OAuth.GetToken is starting")
	if time.Until(o.token.ExpiresAt) >= o.RefreshTokenBeforeExp {
		logger.Info(
			"OAuth.GetToken: token is still valid",
			zap.Time("expired at", o.token.ExpiresAt),
			zap.Duration("time period", time.Until(o.token.ExpiresAt)),
		)
		return o.token, nil
	}

	logger.Info(
		"OAuth.GetToken: token is old, try to refresh",
		zap.Duration("time period", time.Until(o.token.ExpiresAt)),
	)
	reqBody := bytes.NewBuffer([]byte("scope=" + o.Scope))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.OAuthURL, reqBody)
	if err != nil {
		return Token{}, fmt.Errorf("new http request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("RqUID", uuid.New().String())
	req.Header.Add("Authorization", "Basic "+o.AuthKey)
	resp, err := o.deps.HTTPClient.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("OAuth.GetToken: http response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Token{}, fmt.Errorf("OAuth.GetToken: API error: status=%d body=%s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Token{}, fmt.Errorf("OAuth.GetToken: read response body: %w", err)
	}
	oauthResp := OauthResp{}
	err = json.Unmarshal(body, &oauthResp)
	if err != nil {
		return Token{}, fmt.Errorf("OAuth.GetToken: umarshal token: %w", err)
	}
	o.token.ExpiresAt = time.Unix(oauthResp.ExpiresAt, 0)
	o.token.Token = oauthResp.Token
	return o.token, nil
}
