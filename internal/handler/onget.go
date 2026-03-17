package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

func (h *Handler) OnGet(c telebot.Context) error {
	logger := h.deps.Logger
	errMsg := h.deps.ErrMsg
	logger.Info("Handler.OnGet: is triggered", zap.Int64("User ID", c.Sender().ID), zap.String("Usename", c.Sender().Username))
	//errMsg := h.deps.ErrMsg
	ctx, stop := context.WithTimeout(context.Background(), time.Second*60)
	defer stop()

	text := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/get"))
	id, err := strconv.Atoi(text)
	if err != nil {
		logger.Error(
			"Handler.OnGet: convert meeting ID to int",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Error(err),
		)
		return c.Send(fmt.Sprintf("It looks like the meeting ID your provided (%s) is incorrect.\nPlease ensure that you use only digitals", text))
	}

	meets, err := h.meetRepo.GetRowsBy(ctx, "user_id=$1 AND id=$2", c.Sender().ID, id)
	if err != nil {
		logger.Error(
			"Handler.OnGet: get meessage in DB",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Error(err),
		)
		return c.Send(errMsg)
	}
	var resp string
	if len(meets) == 0 {
		resp = fmt.Sprintf("I can't find any meetings with the ID %d", id)
	} else {
		resp = "Here is the meeting you requested:\n"
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
			"Handler.OnGet: get meeting result",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Int("Meeting count", len(meets)),
		)
	}
	return c.Send(resp)
}
