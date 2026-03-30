package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/theluckymadman/memotogo/internal/utils"
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
----------------------------------------------
`,
				"ID", meet.ID,
				"Data", meet.Date,
				"Duration", fmt.Sprintf("%d sec", meet.Duration),
			)
		}
		logger.Info(
			"Handler.OnList: list meetings result",
			zap.Int64("User ID", c.Sender().ID),
			zap.String("Usename", c.Sender().Username),
			zap.Int("Meeting count", len(meets)),
		)
	}

	chunks := utils.SplitText(resp, 400)

	for _, chunk := range chunks {
		if err := c.Send(chunk); err != nil {
			logger.Error("failed to send chunk", zap.Error(err))
			return err
		}
	}

	return nil
}
