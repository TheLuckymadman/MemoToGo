package workers

import (
	"context"

	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
	"github.com/theluckymadman/memotogo/internal/model"
)

type Transcriptor interface {
	GetTranscriptionSync(ctx context.Context, audio []byte) (*model.Transcription, error)
	CreateTask(ctx context.Context, audio []byte) (*model.TranscriptionTask, error)
	GetTaskStatus(ctx context.Context, task model.TranscriptionTask) (*model.TranscriptionTask, error)
	GetTranscription(ctx context.Context, task model.TranscriptionTask) (*model.Transcription, error)
}

type LLM interface {
	Chat(ctx context.Context, messages []llmtype.Message, tools []llmtype.Tool) (*llmtype.ChatResponse, error)
}
