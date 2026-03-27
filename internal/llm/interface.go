package llm

import (
	"context"

	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
)

type LLM interface {
	Chat(ctx context.Context, messages []llmtype.Message, tools []llmtype.Tool) (*llmtype.ChatResponse, error)
}

