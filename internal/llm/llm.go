package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/llm/adapter"
	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
	"github.com/theluckymadman/memotogo/internal/llm/tool"
	"github.com/theluckymadman/memotogo/internal/model"

	"go.uber.org/zap"
)

type LLMService struct {
	gigaAdapter  *adapter.GigaChatAdapter
	systemPrompt string
	deps         *deps.Deps
	state        []llmtype.Message
	toolRegistry tool.Registry
}

func NewLLMService(gigaAdapter *adapter.GigaChatAdapter, systemPrompt string, deps *deps.Deps) *LLMService {
	return &LLMService{gigaAdapter: gigaAdapter, systemPrompt: systemPrompt, deps: deps}
}

func (l *LLMService) Chat(ctx context.Context, question string) (string, error) {
	logger := l.deps.Logger
	logger.Info("LLMService.Chat", zap.String("User:", question))

	toolList := []llmtype.Tool{
		tool.ListMeeting{},
	}

	for {
		res, _ := l.gigaAdapter.Chat(
			ctx,
			[]llmtype.Message{
				{
					Role:    "system",
					Content: l.systemPrompt,
				},
				{
					Role:    "user",
					Content: question,
				},
			},
			toolList,
		)

		call := tool.ExtractToolCall(res)
		if call == nil {
			return res.Choices[0].Message.Content, nil
		}

		call.
	}

	res, err := l.gigaAdapter.Chat(
		ctx,
		[]llmtype.Message{
			{
				Role:    "system",
				Content: l.systemPrompt,
			},
			{
				Role:    "user",
				Content: question,
			},
		},
		toolList,
	)
	if err != nil {
		logger.Error(
			"LLMService.Chat: LLM respone",
			zap.Error(err),
		)
		return "", fmt.Errorf("LLMService.Chat: giga response: %w", err)
	}

	logger.Info(
		"LLMService.Answer: got respone from LLM",
		zap.Any("Response", res),
	)

	fc := res.Choices[0].Message.ToolCall

	for fc != nil && fc.Name != "" {
		json.Unmarshal(fc.Arguments)

		arg, ok := fc.Arguments["userID"].(string)
		if !ok || arg == "" {
			break
		}
		funcRes, _ := listMeeting.Call(ctx, arg)
		funcResObj := map[string]string{
			"result": funcRes,
		}

		funcResBytes, _ := json.Marshal(funcResObj)
		res, err = l.giga.Chat(
			ctx,
			[]model.ReqMessage{
				{
					Role:    "system",
					Content: l.systemPrompt,
				},
				{
					Role:    "user",
					Content: question,
				},
				{
					Role:         "assistant",
					Content:      res.Choices[0].Message.Content,
					FunctionCall: res.Choices[0].Message.FunctionCall,
				},
				{
					Role:    "function",
					Content: string(funcResBytes),
				},
			},
			[]model.Function{
				{
					Name:        listMeeting.Name(),
					Description: listMeeting.Description(),
					Parameters: model.JSONSchema{
						Type:     "object",
						Required: []string{"userID"},
						Properties: map[string]model.Property{
							"userID": {
								Type:        "string",
								Description: "Requires userID. If userID is unknown, ask the user before calling.",
							},
						},
					},
					ReturnParameters: model.JSONSchema{
						Type: "object",
						Properties: map[string]model.Property{
							"result": {
								Type:        "string",
								Description: "Result of listing meetings in JSON Format",
							},
						},
					},
				},
			},
		)
		if err != nil {
			logger.Error(
				"LLMService.Answer: get file from server",
				zap.Error(err),
			)
			return "", fmt.Errorf("LLMService.Chat: giga response after tool calling: %w", err)
		}

		logger.Info(
			"LLMService.Answer: got respone from LLM after tool calling",
			zap.Any("Response", res),
		)
		fc = res.Choices[0].Message.FunctionCall
	}

	return res.Choices[0].Message.Content, nil
}
