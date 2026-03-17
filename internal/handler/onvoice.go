package handler

import (
	"io"

	"github.com/theluckymadman/memotogo/internal/queue"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

func (h *Handler) OnVoice(c telebot.Context) error {
	logger := h.deps.Logger
	logger.Info("Handler.onVoice: is triggered", zap.Int64("User ID", c.Sender().ID), zap.String("Usename", c.Sender().Username))
	errMsg := h.deps.ErrMsg

	voice := c.Message().Voice
	if voice == nil {
		logger.Error(
			"Handler.onVoice",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.String("Error", "No voice meessage provided"),
		)
		return c.Send(errMsg)
	}
	logger.Info(
		"Handler.onVoice: voice file received",
		zap.Int64("User ID", c.Sender().ID),
		zap.String("Usename", c.Sender().Username),
		zap.String("FileID", voice.File.FileID),
	)
	file, err := h.bot.File(&voice.File)
	if err != nil {
		logger.Error(
			"Handler.onVoice: get file from server",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Error(err),
		)
		return c.Send(errMsg)
	}
	audio, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Handler.onVoice: failed to read audio", zap.Error(err))
		return c.Send(errMsg)
	}

	rcvChan := make(chan string, 1)
	job := queue.SpeechJob{
		ChatID:        c.Chat().ID,
		FileID:        c.Message().Voice.FileID,
		Audio:         audio,
		AudioDuration: c.Message().Voice.Duration,
		RcvChan:       rcvChan,
	}
	h.queue.Jobs <- job
	logger.Info(
		"Handler.onVoice: send job to chan",
		zap.Int64("User ID", c.Sender().ID),
		zap.String("Usename", c.Sender().Username),
		zap.String("FileID", voice.File.FileID),
	)
	c.Send("Awesome:) Wait a bit, I'll respond soon.")
	resp := <-rcvChan
	logger.Info(
		"Handler.onVoice: got respone from chan",
		zap.Int64("User ID", c.Sender().ID),
		zap.String("Usename", c.Sender().Username),
		zap.String("Response", resp),
	)
	return c.Send(resp)
}
