package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
	"github.com/theluckymadman/memotogo/internal/llm/tool"

	"go.uber.org/zap"
)

type LLMService struct {
	llm          LLM
	systemPrompt string
	deps         *deps.Deps
	state        *State
	toolRegistry tool.Registry
	maxLLMIter   int
}

func NewLLMService(llm LLM, systemPrompt string, deps *deps.Deps, state *State, toolRegistry tool.Registry, maxLLMIter int) *LLMService {
	return &LLMService{llm: llm, systemPrompt: systemPrompt, deps: deps, state: state, toolRegistry: toolRegistry, maxLLMIter: maxLLMIter}
}

func (l *LLMService) Chat(ctx context.Context, userID int64, userName string, question string) (string, error) {
	logger := l.deps.Logger
	logger.Info("LLMService.Chat", zap.String("User:", question))

	toolList := []llmtype.Tool{}
	for _, t := range l.toolRegistry {
		toolList = append(toolList, t)
	}
	finalSystemPrompt := fmt.Sprintf(`
		%s
		Information about a user: 
		1) UserID is %d 
		2) User name is %s \n
		Use this information in the conversation, greetings and for calling the functions`, l.systemPrompt, userID, userName)
	messages := []llmtype.Message{
		{
			Role:    "system",
			Content: finalSystemPrompt,
		},
		{
			Role:    "user",
			Content: question,
		},
	}

	l.state.AddMessage(userID, messages)

	var iterCnt int
	for iterCnt < l.maxLLMIter {
		res, err := l.llm.Chat(ctx, messages, toolList)
		if err != nil {
			logger.Error(
				"LLMService.Chat LLM responds error",
				zap.Int64("UserID", userID),
				zap.Error(err),
			)
			return "", fmt.Errorf("LLMService.Chat LLM responds error: %w", err)
		}

		l.state.AddMessage(userID, []llmtype.Message{res.Choices[0].Message})

		call := tool.ExtractToolCall(res)
		if call == nil {
			logger.Info(
				"LLMService.Chat no tools are being called, LLM responds to user",
				zap.Int64("UserID", userID),
				zap.String("LLM response", res.Choices[0].Message.Content),
				zap.Any("User state", l.state.store[userID]),
			)
			return res.Choices[0].Message.Content, nil
		}

		logger.Info(
			"LLMService.Chat tool called",
			zap.Int64("UserID", userID),
			zap.String("LLM response", res.Choices[0].Message.Content),
			zap.String("Tool name", call.Name),
			zap.String("Tool args", string(call.Arguments)),
			zap.Any("User state", l.state.store[userID]),
		)

		tool, ok := l.toolRegistry[call.Name]
		if !ok {
			return "", fmt.Errorf("tool not found: %s", call.Name)
		}
		callRes, err := tool.Call(ctx, call.Arguments)
		if err != nil {
			return "", fmt.Errorf("tool execution: %w ", err)
		}
		callResObj := map[string]any{
			"result": callRes,
		}
		resOut, err := json.Marshal(callResObj)
		if err != nil {
			return "", fmt.Errorf("tool result marshalling: %w ", err)
		}
		messages = append(messages,
			res.Choices[0].Message,
			llmtype.Message{
				Role:    "function",
				Content: string(resOut),
			})
		l.state.AddMessage(userID, messages)
	}
	return "", fmt.Errorf("iteration limit reached %d", iterCnt)
}
