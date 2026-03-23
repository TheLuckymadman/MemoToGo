package handler

import (
	"context"
	"log"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/llm"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/queue"
	"github.com/theluckymadman/memotogo/internal/repository"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

type Handler struct {
	queue    *queue.Queue
	deps     *deps.Deps
	bot      *telebot.Bot
	llmSvc   *llm.LLMService
	userRepo *repository.Storage[model.User]
	meetRepo *repository.Storage[model.Meeting]
}

func NewHandler(
	deps *deps.Deps,
	queue *queue.Queue,
	bot *telebot.Bot,
	llmSvc *llm.LLMService,
	userRepo *repository.Storage[model.User],
	meetRepo *repository.Storage[model.Meeting],
) *Handler {
	return &Handler{queue: queue, deps: deps, bot: bot, llmSvc: llmSvc, userRepo: userRepo, meetRepo: meetRepo}
}

func (h *Handler) OnText(c telebot.Context) error {
	logger := h.deps.Logger
	logger.Info("handler event", zap.String("text", c.Text()))
	errMsg := h.deps.ErrMsg
	log.Printf("user info: %v", c.Sender())

	res, err := h.llmSvc.Chat(context.Background(), c.Text())
	if err != nil {
		logger.Error(
			"Handler.OnText: get file from server",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Error(err),
		)
		return c.Send(errMsg)
	}
	return c.Send(res)
}
