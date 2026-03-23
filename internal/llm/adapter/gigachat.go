package adapter

import (
	"context"
	"fmt"

	"github.com/theluckymadman/memotogo/internal/client/gigachat"
	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
	"go.uber.org/zap"
)

type GigaChatAdapter struct {
	giga *gigachat.GigaChat
	deps *deps.Deps
}

func NewGigaChatAdapter() *GigaChatAdapter {
	return &GigaChatAdapter{}
}

func (a *GigaChatAdapter) Chat(ctx context.Context, messages []llmtype.Message, tools []llmtype.Tool) (*llmtype.ChatResponse, error) {
	logger := a.deps.Logger
	logger.Info("GigaChatAdapter.Chat")

	var gigaMessages []gigachat.ReqMessage
	for _, m := range messages {
		msg := gigachat.ReqMessage{
			Role:    m.Role,
			Content: m.Content,
		}

		if m.ToolCall != nil {
			msg.FunctionCall = &gigachat.FunctionCall{
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

func ToGigaFunctions(tools []llmtype.Tool) []gigachat.Function {
	var funcs []gigachat.Function

	for _, t := range tools {
		funcs = append(funcs, gigachat.Function{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: gigachat.JSONSchema{
				Type:       t.Parameters().Type,
				Properties: convertProps(t.Parameters().Properties),
				Required:   t.Parameters().Required,
			},
			ReturnParameters: gigachat.JSONSchema{
				Type:       t.ReturnParameters().Type,
				Properties: convertProps(t.ReturnParameters().Properties),
			},
		})
	}

	return funcs
}

func convertProps(props map[string]llmtype.Property) map[string]gigachat.Property {
	gigaProps := make(map[string]gigachat.Property)
	for k, v := range props {
		gigaProps[k] = gigachat.Property{
			Type:        v.Type,
			Description: v.Description,
		}
	}
	return gigaProps
}
