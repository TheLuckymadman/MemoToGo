package oauth

import (
	"sync"
	"time"

	"github.com/theluckymadman/memotogo/internal/deps"
)

type OAuth struct {
	OAuthURL              string
	AuthKey               string
	Scope                 string
	RefreshTokenBeforeExp time.Duration
	token                 Token
	deps                  *deps.Deps
	mu                    sync.Mutex
}

type OauthResp struct {
	Token     string `json:"access_token"`
	ExpiresAt int64  `json:"expires_at"`
}

type Token struct {
	Token     string    `json:"access_token"`
	ExpiresAt time.Time `json:"expires_at"`
}
