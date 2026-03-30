package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/repository"
	"go.uber.org/zap"
)

type GetMeeting struct {
	meetRepo *repository.Storage[model.Meeting]
	deps     *deps.Deps
}

func NewGetMeeting(meetRepo *repository.Storage[model.Meeting], deps *deps.Deps) *GetMeeting {
	return &GetMeeting{meetRepo: meetRepo, deps: deps}
}

func (t GetMeeting) Name() string {
	return "get_meeting"
}

func (t GetMeeting) Description() string {
	return "Returns the meeting by id"
}

func (t GetMeeting) Parameters() llmtype.JSONSchema {
	return llmtype.JSONSchema{
		Type: "object",
		Properties: map[string]llmtype.Property{
			"userID": {
				Type:        "integer",
				Description: "User identifier",
			},
			"meetingID": {
				Type:        "integer",
				Description: "Meeting identifier",
			},
		},
		Required: []string{"userID", "meetingID"},
	}
}

func (t GetMeeting) ReturnParameters() llmtype.JSONSchema {
	return llmtype.JSONSchema{
		Type: "object",
		Properties: map[string]llmtype.Property{
			"result": {
				Type:        "string",
				Description: "Meeting information",
			},
			"error": {
				Type:        "string",
				Description: "error",
			},
		},
	}
}

func (t GetMeeting) Call(ctx context.Context, args json.RawMessage) (any, error) {
	logger := t.deps.Logger
	var input struct {
		UserID    int64 `json:"userID"`
		MeetingID int64 `json:"meetingID"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	meets, err := t.meetRepo.GetRowsBy(ctx, "user_id=$1 AND id=$2", input.UserID, input.MeetingID)
	if err != nil {
		logger.Error(
			"GetMeeting.Call: get the meeting from DB",
			zap.Int64("User ID", input.UserID),
			zap.Int64("Meeting ID", input.MeetingID),
			zap.Error(err),
		)
		return map[string]any{
			"result": "",
			"error":  "get_meeting error",
		}, nil
	}

	var res string
	if len(meets) == 0 {
		errMsg := fmt.Sprintf("get_meeting error: there is no meeting with this meetingID %d", input.MeetingID)
		logger.Error(
			"GetMeeting.Call: meeting not found",
			zap.Int64("User ID", input.UserID),
			zap.Int64("Meeting ID", input.MeetingID),
			zap.String("Error", errMsg),
		)
		return map[string]any{
			"result": "",
			"error":  errMsg,
		}, nil
	}

	meet := meets[0]
	var summary string
	if meet.Summary != nil {
		summary = *meet.Summary
	}
	res = fmt.Sprintf(
		"Date: %v\nDuration: %d sec\nSummary:\n%s",
		meet.Date,
		meet.Duration,
		summary,
	)

	return map[string]any{
		"result": res,
		"error":  "",
	}, nil
}
