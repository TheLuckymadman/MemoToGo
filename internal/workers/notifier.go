package workers

import (
	"context"
	"time"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/repository"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

type Notifier struct {
	bot      *telebot.Bot
	deps     *deps.Deps
	userRepo *repository.Storage[model.User]
}

func NewNotifier(bot *telebot.Bot, deps *deps.Deps, userRepo *repository.Storage[model.User]) *Notifier {
	return &Notifier{bot: bot, deps: deps, userRepo: userRepo}
}

func (n *Notifier) Start(ctxDone context.Context) {
	logger := n.deps.Logger
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctxDone.Done():
			logger.Info("Notifier: stopped gracefully")
			return
		case <-ticker.C:
			n.notifyOnce(ctxDone, logger)
		}
	}
}

func (n *Notifier) notifyOnce(ctxDone context.Context, logger *zap.Logger) {
	ctx, stop := context.WithTimeout(ctxDone, time.Minute*2)
	defer stop()
	users, err := n.userRepo.GetRowsBy(ctx, "approved=$1 AND welcome_msg_sent = TIMESTAMP '0001-01-01 00:00:00'", true)
	if err != nil {
		logger.Error(
			"Notifier.Start: list unapproved users",
			zap.Error(err),
		)
	}
	for _, user := range users {
		chat := telebot.Chat{ID: user.ChatID}
		_, err := n.bot.Send(&chat, "We have accepted your request. Thank you for your interest in our project")
		if err != nil {
			logger.Error(
				"Notifier.Start: notify user",
				zap.Int64("User ID", user.ID),
				zap.Int64("Chat ID", user.ChatID),
				zap.Error(err),
			)
		}
		err = n.userRepo.UpdateByID(ctx, user.ID, "welcome_msg_sent=$1", time.Now())
		if err != nil {
			logger.Error(
				"Notifier.Start: update user welcome_msg_sent timestamp",
				zap.Int64("User ID", user.ID),
				zap.Int64("Chat ID", user.ChatID),
				zap.Error(err),
			)
		}
	}
}
