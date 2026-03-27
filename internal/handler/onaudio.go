package handler

import (
	"io"

	"github.com/theluckymadman/memotogo/internal/audio"
	"github.com/theluckymadman/memotogo/internal/queue"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

func (h *Handler) OnAudio(c telebot.Context) error {
	logger := h.deps.Logger
	logger.Info("Handler.OnAudio: is triggered", zap.Int64("User ID", c.Sender().ID), zap.String("Usename", c.Sender().Username))
	errMsg := h.deps.ErrMsg

	audioMsg := c.Message().Audio
	if audioMsg == nil {
		logger.Error(
			"Handler.OnAudio",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.String("Error", "No audio meessage provided"),
		)
		return c.Send(errMsg)
	}
	format := audioMsg.MIME // e.g. "audio/mpeg", "audio/wav"
	logger.Info(
		"Handler.OnAudio: audio file received",
		zap.Int64("User ID", c.Sender().ID),
		zap.String("Usename", c.Sender().Username),
		zap.String("FileID", audioMsg.File.FileID),
		zap.String("Format", format),
	)
	file, err := h.bot.File(&audioMsg.File)
	if err != nil {
		logger.Error(
			"Handler.OnAudio: get file from server",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Error(err),
		)
		return c.Send(errMsg)
	}
	targetAudio, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Handler.OnAudio: failed to read audio", zap.Error(err))
		return c.Send(errMsg)
	}
	if format != "audio/ogg" {
		targetAudio, err = audio.CnvToOGG(targetAudio)
		if err != nil {
			logger.Error(
				"Handler.OnAudio: convert audio to audio/ogg",
				zap.Int64("User ID", c.Sender().ID),
				zap.String("Usename", c.Sender().Username),
				zap.Error(err),
			)
			return c.Send(errMsg)
		}
	}

	rcvChan := make(chan string, 1)
	job := queue.SpeechJob{
		ChatID:        c.Chat().ID,
		FileID:        audioMsg.FileID,
		Audio:         targetAudio,
		AudioDuration: audioMsg.Duration,
		RcvChan:       rcvChan,
	}
	h.queue.Jobs <- job
	logger.Info(
		"Handler.onVoice: send job to chan",
		zap.Int64("User ID", c.Sender().ID),
		zap.String("Usename", c.Sender().Username),
		zap.String("FileID", audioMsg.File.FileID),
	)
	//c.Send("Awesome:) Wait a bit, I'll respond soon.")
	resp := <-rcvChan
	logger.Info(
		"Handler.onVoice: got respone from chan",
		zap.Int64("User ID", c.Sender().ID),
		zap.String("Usename", c.Sender().Username),
		zap.String("Response", resp),
	)
	return c.Send(resp)
}
