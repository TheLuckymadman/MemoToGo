package middleware

import (
	"context"
	"time"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/repository"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

type Middleware struct {
	userRepo *repository.Storage[model.User]
	deps     *deps.Deps
}

func NewMiddleware(userRepo *repository.Storage[model.User], deps *deps.Deps) *Middleware {
	return &Middleware{userRepo: userRepo, deps: deps}
}

func (m *Middleware) Auth(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		logger := m.deps.Logger
		errMsg := m.deps.ErrMsg
		logger.Info("Middleware.Auth", zap.Int64("User ID", c.Sender().ID), zap.String("Usename", c.Sender().Username))
		ctx, stop := context.WithTimeout(context.Background(), time.Second*60)
		defer stop()
		users, err := m.userRepo.GetRowsBy(ctx, "id=$1 AND approved=true", c.Sender().ID)
		if err != nil {
			logger.Error(
				"Middleware.Auth: get approved user by id",
				zap.Int64("User ID", c.Sender().ID),
				zap.String("Usename", c.Sender().Username),
				zap.Error(err),
			)
			return c.Send(errMsg)
		}
		if len(users) == 0 {
			logger.Warn(
				"Middleware.Auth: no such user in DB, access denied",
				zap.Int64("User ID", c.Sender().ID),
				zap.String("Usename", c.Sender().Username),
			)
			return c.Send("Ups, we didn't find you among our users. Please type the command /start.")
		}
		logger.Warn(
			"Middleware.Auth: access allowed",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
		)
		return next(c)
	}
}
