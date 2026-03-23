package tool

import (
	"context"
	"encoding/json"

	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
)

type ListMeeting struct{}

func (t ListMeeting) Name() string {
	return "list_meeting"
}

func (t ListMeeting) Description() string {
	return "Returns the list of meetings"
}

func (t ListMeeting) Parameters() llmtype.JSONSchema {
	return llmtype.JSONSchema{
		Type: "object",
		Properties: map[string]llmtype.Property{
			"userID": {
				Type:        "string",
				Description: "User identifier",
			},
		},
		Required: []string{"userID"},
	}
}

func (t ListMeeting) ReturnParameters() llmtype.JSONSchema {
	return llmtype.JSONSchema{
		Type: "object",
		Properties: map[string]llmtype.Property{
			"result": {
				Type:        "string",
				Description: "Result of listing meetings in JSON Format",
			},
		},
	}
}

func (t ListMeeting) Call(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		UserID string `json:"userID"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	return []map[string]any{
		{"id": 1, "name": "the first meeting"},
		{"id": 2, "name": "the second meeting"},
	}, nil
}
