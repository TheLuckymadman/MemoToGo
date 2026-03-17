package handler

import (
	"log"

	"github.com/theluckymadman/memotogo/internal/deps"
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
	userRepo *repository.Storage[model.User]
	meetRepo *repository.Storage[model.Meeting]
}

func NewHandler(
	deps *deps.Deps,
	queue *queue.Queue,
	bot *telebot.Bot,
	userRepo *repository.Storage[model.User],
	meetRepo *repository.Storage[model.Meeting],
) *Handler {
	return &Handler{queue: queue, deps: deps, bot: bot, userRepo: userRepo, meetRepo: meetRepo}
}

func (h *Handler) OnText(c telebot.Context) error {
	h.deps.Logger.Info("handler event", zap.String("text", c.Text()))
	log.Printf("user info: %v", c.Sender())
	return c.Send("Hello text!")
}

func (h *Handler) OnAudio(c telebot.Context) error {
	h.deps.Logger.Info("handler event", zap.String("audio", c.Data()))
	log.Printf("user info: %v", c.Sender())
	return c.Send("Hello audio!")
}
