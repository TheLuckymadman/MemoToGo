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

type ListMeetings struct {
	meetRepo *repository.Storage[model.Meeting]
	deps     *deps.Deps
}

func NewListMeetings(meetRepo *repository.Storage[model.Meeting], deps *deps.Deps) *ListMeetings {
	return &ListMeetings{meetRepo: meetRepo, deps: deps}
}

func (t ListMeetings) Name() string {
	return "list_meetings"
}

func (t ListMeetings) Description() string {
	return "Returns the list of meetings"
}

func (t ListMeetings) Parameters() llmtype.JSONSchema {
	return llmtype.JSONSchema{
		Type: "object",
		Properties: map[string]llmtype.Property{
			"userID": {
				Type:        "integer",
				Description: "User identifier",
			},
		},
		Required: []string{"userID"},
	}
}

func (t ListMeetings) ReturnParameters() llmtype.JSONSchema {
	return llmtype.JSONSchema{
		Type: "object",
		Properties: map[string]llmtype.Property{
			"result": {
				Type: "array",
				Items: &llmtype.JSONSchema{
					Type: "object",
					Properties: map[string]llmtype.Property{
						"id":          {Type: "integer"},
						"description": {Type: "string"},
					},
				},
			},
			"error": {
				Type:        "string",
				Description: "error",
			},
		},
	}
}

func (t ListMeetings) Call(ctx context.Context, args json.RawMessage) (any, error) {
	logger := t.deps.Logger
	var input struct {
		UserID int64 `json:"userID"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	meets, err := t.meetRepo.GetRowsBy(ctx, "user_id=$1", input.UserID)
	if err != nil {
		logger.Error(
			"ListMeetings.Call: list meetings from DB",
			zap.Int64("User ID", input.UserID),
			zap.Error(err),
		)
		return map[string]any{
			"result": nil,
			"error":  "list_meetings error",
		}, nil
	}

	type meet struct {
		ID          int64  `json:"id"`
		Description string `json:"description"`
	}
	var res []meet

	for _, m := range meets {
		res = append(res, meet{
			ID:          m.ID,
			Description: fmt.Sprintf("Date: %v, Duration: %d sec", m.Date, m.Duration),
		})
	}

	return map[string]any{
		"result": res,
		"error":  "",
	}, nil
}
