package gigachat

import (
	"context"
	"fmt"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
	"go.uber.org/zap"
)

type GigaChatAdapter struct {
	giga *GigaChat
	deps *deps.Deps
}

func NewGigaChatAdapter(giga *GigaChat, deps *deps.Deps) *GigaChatAdapter {
	return &GigaChatAdapter{giga: giga, deps: deps}
}

func (a *GigaChatAdapter) Chat(ctx context.Context, messages []llmtype.Message, tools []llmtype.Tool) (*llmtype.ChatResponse, error) {
	logger := a.deps.Logger
	logger.Info("GigaChatAdapter.Chat")

	var gigaMessages []ReqMessage
	for _, m := range messages {
		msg := ReqMessage{
			Role:    m.Role,
			Content: m.Content,
		}

		if m.ToolCall != nil {
			msg.FunctionCall = &FunctionCall{
				Name:      m.ToolCall.Name,
				Arguments: m.ToolCall.Arguments,
			}
		}

		gigaMessages = append(gigaMessages, msg)
	}

	gigaRes, err := a.giga.Chat(ctx, gigaMessages, ToGigaFunctions(tools))
	if err != nil {
		logger.Error(
			"GigaChatAdapter.Chat: giga response",
			zap.Error(err),
		)
		return nil, fmt.Errorf("GigaChatAdapter.Chat: response: %w", err)
	}
	logger.Info("GigaChatAdapter.Chat: response", zap.Any("Response", gigaRes))

	var choices []llmtype.Choice

	for _, c := range gigaRes.Choices {

		msg := llmtype.Message{
			Role:    c.Message.Role,
			Content: c.Message.Content,
		}

		if c.Message.FunctionCall != nil {
			msg.ToolCall = &llmtype.ToolCall{
				Name:      c.Message.FunctionCall.Name,
				Arguments: c.Message.FunctionCall.Arguments,
			}
		}

		choices = append(choices, llmtype.Choice{
			Message:      msg,
			FinishReason: c.FinishReason,
		})
	}

	res := &llmtype.ChatResponse{
		Choices: choices,
		Model:   gigaRes.Model,
		Usage: llmtype.Usage{
			PromptTokens:          gigaRes.Usage.PromptTokens,
			CompletionTokens:      gigaRes.Usage.CompletionTokens,
			PrecachedPromptTokens: gigaRes.Usage.PrecachedPromptTokens,
			TotalTokens:           gigaRes.Usage.TotalTokens,
		},
	}

	return res, nil
}

func ToGigaFunctions(tools []llmtype.Tool) []Function {
	var funcs []Function

	if len(tools) == 0 {
		return nil
	}

	for _, t := range tools {
		funcs = append(funcs, Function{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: JSONSchema{
				Type:       t.Parameters().Type,
				Properties: convertProps(t.Parameters().Properties),
				Required:   t.Parameters().Required,
			},
			ReturnParameters: JSONSchema{
				Type:       t.ReturnParameters().Type,
				Properties: convertProps(t.ReturnParameters().Properties),
			},
		})
	}

	return funcs
}

func convertProps(props map[string]llmtype.Property) map[string]Property {
	gigaProps := make(map[string]Property)
	for k, v := range props {
		gigaProps[k] = Property{
			Type:        v.Type,
			Description: v.Description,
		}
	}
	return gigaProps
}
