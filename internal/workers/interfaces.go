package workers

import (
	"context"

	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/transcription"
)

type Transcriptor interface {
	GetTranscriptionSync(ctx context.Context, audio []byte) (*transcription.TranscriptionResponse, error)
	UploadFile(ctx context.Context, audio []byte) (string, error)
	NewTranscriptionTask(ctx context.Context, requestFileID string) (string, error)
	GetTaskStatus(ctx context.Context, taskID string) (*transcription.GetTaskStatusResponse, error)
	DownloadFile(ctx context.Context, responseFileID string) (transcription.DownloadResponse, error)
}

type LLM interface {
	Chat(ctx context.Context, messages []model.ReqMessage, tools []model.Function) (*model.ChatResponse, error)
}
