package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

func (h *Handler) OnFind(c telebot.Context) error {
	logger := h.deps.Logger
	errMsg := h.deps.ErrMsg
	logger.Info("Handler.onFind: is triggered", zap.Int64("User ID", c.Sender().ID), zap.String("Usename", c.Sender().Username))
	//errMsg := h.deps.ErrMsg
	ctx, stop := context.WithTimeout(context.Background(), time.Second*60)
	defer stop()

	text := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/find"))

	meets, err := h.meetRepo.Search(ctx, text, 0)
	if err != nil {
		logger.Error(
			"Handler.onFind: Search meessage in DB",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Error(err),
		)
		return c.Send(errMsg)
	}
	var resp string
	if len(meets) == 0 {
		resp = "I can't find any meetings by the provided you keyword:("
	} else {
		resp = "Here is the meetings I found:\n"
		for _, meet := range meets {
			var summary string
			if meet.Summary != nil {
				summary = *meet.Summary
			}
			resp += fmt.Sprintf(
				`%-10s: %v
%-10s: %v
%-10s: %v
%s
----------------------------------------------
`,
				"ID", meet.ID,
				"Data", meet.Date,
				"Duration", fmt.Sprintf("%d sec", meet.Duration),
				summary,
			)
		}
		logger.Info(
			"Handler.onFind: searchresult",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Int("Meeting count", len(meets)),
		)
	}
	return c.Send(resp)
}
