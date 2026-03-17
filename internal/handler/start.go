package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/theluckymadman/memotogo/internal/model"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

func (h *Handler) Start(c telebot.Context) error {
	logger := h.deps.Logger
	errMsg := h.deps.ErrMsg
	logger.Info("Handler.Start", zap.Int64("User ID", c.Sender().ID), zap.String("Usename", c.Sender().Username))
	ctx, stop := context.WithTimeout(context.Background(), time.Second*60)
	defer stop()

	user, err := h.userRepo.GetByID(ctx, c.Sender().ID)
	if err != nil {
		logger.Error(
			"Handler.Start: get user by id",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Error(err),
		)
		return c.Send(errMsg)
	}
	if user != nil {
		logger.Info(
			"Handler.Start: user already exists",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Bool("Approved", user.Approved),
		)
		if user.Approved {
			return c.Send(fmt.Sprintf("Hi %s, welcome again:)", c.Sender().Username))
		} else {
			return nil
		}
	}
	newUser := model.User{
		ID:         c.Sender().ID,
		Username:   c.Sender().Username,
		ChatID:     c.Chat().ID,
		Created_at: time.Now(),
		Approved:   false,
	}
	err = h.userRepo.Create(ctx, newUser)
	if err != nil {
		logger.Error(
			"Handler.Start: create user",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Error(err),
		)
		return c.Send(errMsg)
	}
	logger.Info(
		"Handler.Start: user cerated and waits for an approvement",
		zap.Int64("User ID", c.Sender().ID),
		zap.String("Usename", c.Sender().Username),
		zap.Bool("Approved", newUser.Approved),
	)
	msg := fmt.Sprintf(
		"Hello %s, thank you for your interest in our awesome bot. We are already processing your request. Please wait a bit :)",
		c.Sender().Username,
	)

	return c.Send(msg)
}
