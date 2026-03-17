package handler

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

func (h *Handler) OnList(c telebot.Context) error {
	logger := h.deps.Logger
	errMsg := h.deps.ErrMsg
	logger.Info("Handler.OnList: is triggered", zap.Int64("User ID", c.Sender().ID), zap.String("Usename", c.Sender().Username))
	//errMsg := h.deps.ErrMsg
	ctx, stop := context.WithTimeout(context.Background(), time.Second*60)
	defer stop()

	//text := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/list"))

	meets, err := h.meetRepo.GetRowsBy(ctx, "user_id=$1", c.Sender().ID)
	if err != nil {
		logger.Error(
			"Handler.OnList: list meessages in DB",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Error(err),
		)
		return c.Send(errMsg)
	}
	var resp string
	if len(meets) == 0 {
		resp = "I can't find any meetings yet:("
	} else {
		resp = "Here is the meetings I found:\n"
		for _, meet := range meets {
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
				meet.Summary,
			)
		}
		logger.Info(
			"Handler.OnList: list meetings result",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Int("Meeting count", len(meets)),
		)
	}
	return c.Send(resp)
}
