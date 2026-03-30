package salute

import (
	"time"

	"github.com/theluckymadman/memotogo/internal/client/oauth"
	"github.com/theluckymadman/memotogo/internal/deps"
)

type Salute struct {
	URL     string
	Oauth   *oauth.OAuth
	PollInt time.Duration
	deps    *deps.Deps
}

func NewSalute(
	url string,
	oauth *oauth.OAuth,
	deps *deps.Deps,
) *Salute {
	return &Salute{
		URL:   url,
		Oauth: oauth,
		deps:  deps,
	}
}
