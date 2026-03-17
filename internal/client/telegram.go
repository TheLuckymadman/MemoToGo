package client

import (
	"fmt"
	"time"

	"gopkg.in/telebot.v3"
)

func NewBot(token string) (*telebot.Bot, error) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot error: %w", err)
	}
	return b, nil
}
