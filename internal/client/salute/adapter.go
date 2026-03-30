package salute

import (
	"context"
	"fmt"
	"strings"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/model"
	"go.uber.org/zap"
)

type SaluteAdapter struct {
	salute *Salute
	deps   *deps.Deps
}

func NewSaluteAdapter(salute *Salute, deps *deps.Deps) *SaluteAdapter {
	return &SaluteAdapter{salute: salute, deps: deps}
}

func (s *SaluteAdapter) GetTranscriptionSync(ctx context.Context, audio []byte) (*model.Transcription, error) {
	errPrefix := "SaluteAdapter.GetTranscriptionSync:"
	logger := s.deps.Logger
	logger.Info(fmt.Sprintf("%s", errPrefix))

	t, err := s.salute.GetTranscriptionSync(ctx, audio)
	if err != nil {
		logger.Error(
			fmt.Sprintf("%s salute response", errPrefix),
			zap.Error(err),
		)
		return nil, fmt.Errorf("SaluteAdapter.GetTranscriptionSync: salute response: %w", err)
	}
	transcriptionText := strings.Join(t.Result, "\n")
	transcriptRes := model.Transcription{Result: transcriptionText}

	return &transcriptRes, nil
}

func (s *SaluteAdapter) CreateTask(ctx context.Context, audio []byte) (*model.TranscriptionTask, error) {
	errPrefix := "SaluteAdapter.CreateTask:"
	logger := s.deps.Logger
	logger.Info(fmt.Sprintf("%s", errPrefix))

	fileID, err := s.salute.UploadFile(ctx, audio)
	if err != nil {
		logger.Error(
			fmt.Sprintf("%s salute response", errPrefix),
			zap.Error(err),
		)
		return nil, fmt.Errorf("SaluteAdapter.GetTranscriptionSync: salute response: %w", err)
	}

	taskID, err := s.salute.NewTranscriptionTask(ctx, fileID)
	if err != nil {
		logger.Error(
			fmt.Sprintf("%s salute response", errPrefix),
			zap.Error(err),
		)
		return nil, fmt.Errorf("SaluteAdapter.GetTranscriptionSync: salute response: %w", err)
	}

	transcriptTask := model.TranscriptionTask{TranscriptTaskID: taskID}

	return &transcriptTask, nil
}

func (s *SaluteAdapter) GetTaskStatus(ctx context.Context, task model.TranscriptionTask) (*model.TranscriptionTask, error) {
	errPrefix := "SaluteAdapter.GetTaskStatus:"
	logger := s.deps.Logger
	logger.Info(fmt.Sprintf("%s", errPrefix))

	saluteTaskStatusResp, err := s.salute.GetTaskStatus(ctx, task.TranscriptTaskID)
	if err != nil {
		logger.Error(
			fmt.Sprintf("%s salute response", errPrefix),
			zap.Error(err),
		)
		return nil, fmt.Errorf("SaluteAdapter.GetTaskStatus: salute response: %w", err)
	}

	transcriptTask := model.TranscriptionTask{
		TranscriptTaskID: task.TranscriptTaskID,
		TranscriptStatus: saluteTaskStatusResp.Result.Status,
	}

	if saluteTaskStatusResp.Result.Status == "DONE" {
		transcriptTask.DownloadFileID = saluteTaskStatusResp.Result.ResponseFileID
	}

	return &transcriptTask, nil
}

func (s *SaluteAdapter) GetTranscription(ctx context.Context, task model.TranscriptionTask) (*model.Transcription, error) {
	errPrefix := "SaluteAdapter.GetTranscription:"
	logger := s.deps.Logger
	logger.Info(fmt.Sprintf("%s", errPrefix))

	transcriptionResp, err := s.salute.DownloadFile(ctx, task.DownloadFileID)
	if err != nil {
		logger.Error(
			fmt.Sprintf("%s salute response", errPrefix),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: salute response: %w", errPrefix, err)
	}

	var buildStr strings.Builder

	for _, resp := range transcriptionResp.Result {
		for _, res := range resp.Results {
			if res.NormalizedText != "" {
				buildStr.WriteString(res.NormalizedText)
				buildStr.WriteString("\n")
			}
		}
	}
	transcriptionText := buildStr.String()

	transcription := model.Transcription{
		Result: transcriptionText,
	}

	return &transcription, nil
}
