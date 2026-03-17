package workers

import (
	"context"

	"github.com/theluckymadman/memotogo/internal/llm"
	"github.com/theluckymadman/memotogo/internal/transcription"
)

type Transcriptor interface {
	GetTranscription(ctx context.Context, audio []byte) (*transcription.TranscriptionResponse, error)
}

type LLM interface {
	Chat(ctx context.Context, messages []llm.Message, tools []llm.Tool) (*llm.ChatResponse, error)
}
