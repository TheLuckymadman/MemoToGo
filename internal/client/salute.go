package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/theluckymadman/memotogo/internal/client/oauth"
	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/transcription"
	"go.uber.org/zap"
)

type Salute struct {
	URL   string
	Oauth *oauth.OAuth
	deps  *deps.Deps
}

func NewSalute(
	url string,
	oauth *oauth.OAuth,
	deps *deps.Deps,
) *Salute {
	return &Salute{
		URL:   url,
		Oauth: oauth,
		deps:  deps,
	}
}

func (s *Salute) GetTranscriptionSync(ctx context.Context, audio []byte) (*transcription.TranscriptionResponse, error) {
	logger := s.deps.Logger
	logger.Info("Salute.GetTranscriptionSync")

	reqBody := bytes.NewBuffer(audio)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL+"/speech:recognize", reqBody)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTranscriptionSync: new http request: %w", err)
	}
	token, err := s.Oauth.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTranscriptionSync: get token: %w", err)
	}
	req.Header.Add("Content-Type", "audio/ogg;codecs=opus")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token.Token)

	res, err := s.deps.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTranscriptionSync: http response: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Salute.GetTranscriptionSync: salute API error: status=%d body=%s", res.StatusCode, body)
	}

	var transcriptionResp transcription.TranscriptionResponse
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTranscriptionSync: read response body: %w", err)
	}
	if err = json.Unmarshal(respBody, &transcriptionResp); err != nil {
		return nil, fmt.Errorf("Salute.GetTranscriptionSync: unmarshal transcription data: %w", err)
	}
	bodyStr := string(respBody)
	if len(bodyStr) > 100 {
		bodyStr = bodyStr[:100]
	}
	logger.Info(
		"Salute.GetTranscriptionSync: result",
		zap.Int("status", res.StatusCode),
		zap.String("body", bodyStr),
		zap.Any("transcription response", transcriptionResp),
	)
	return &transcriptionResp, nil
}

func (s *Salute) UploadFile(ctx context.Context, audio []byte) (string, error) {
	logger := s.deps.Logger
	logger.Info("Salute.UploadFile")

	payload := bytes.NewReader(audio)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL+"/data:upload", payload)
	if err != nil {
		return "", fmt.Errorf("Salute.UploadFile: new http request: %w", err)
	}
	token, err := s.Oauth.GetToken(ctx)
	if err != nil {
		return "", fmt.Errorf("Salute.UploadFile: get token: %w", err)
	}
	req.Header.Add("Content-Type", "audio/ogg;codecs=opus")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token.Token)

	res, err := s.deps.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Salute.UploadFile: http response: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("Salute.UploadFile: salute API error: status=%d body=%s", res.StatusCode, body)
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Salute.UploadFile: read response body: %w", err)
	}
	var transcriptionUpload transcription.TranscriptionUploadResponse
	if err = json.Unmarshal(respBody, &transcriptionUpload); err != nil {
		return "", fmt.Errorf("Salute.UploadFile: unmarshal data: %w", err)
	}
	bodyStr := string(respBody)
	if len(bodyStr) > 100 {
		bodyStr = bodyStr[:100]
	}
	logger.Info(
		"Salute.UploadFile: result",
		zap.Int("status", res.StatusCode),
		zap.String("body", bodyStr),
		zap.Any("response", transcriptionUpload),
	)
	return transcriptionUpload.Result.RequestFileID, nil
}

func (s *Salute) NewTranscriptionTask(ctx context.Context, requestFileID string) (string, error) {
	logger := s.deps.Logger
	logger.Info("Salute.NewTranscriptionTask")

	task := transcription.TranscriptionTaskRequest{
		Options: transcription.Options{
			Model:         "general",
			AudioEncoding: "OPUS",
			SampleRate:    16000,
			Language:      "ru-RU",
			// EnableProfanityFilter: false,
			// HypothesesCount:       1,
			// NoSpeechTimeout:       "0s",
			// MaxSpeechTimeout:      "0s",
			// Hints: transcription.Hints{
			// 	Words:         []string{},
			// 	EnableLetters: false,
			// 	EouTimeout:    "0s",
			// },
			ChannelsCount: 1,
			// SpeakerSeparation: transcription.SpeakerSeparationOptions{
			// 	Enable:                false,
			// 	EnableOnlyMainSpeaker: false,
			// 	Count:                 1,
			// },
			// InsightModels: []string{"csi", "call_features"},
		},
		RequestFileID: requestFileID,
	}

	var payload bytes.Buffer
	encoder := json.NewEncoder(&payload)
	encoder.Encode(task)
	req, err := http.NewRequest(http.MethodPost, s.URL+"/speech:async_recognize", &payload)
	if err != nil {
		return "", fmt.Errorf("Salute.NewTranscriptionTask: unmarshal data: %w", err)
	}

	token, err := s.Oauth.GetToken(ctx)
	if err != nil {
		return "", fmt.Errorf("Salute.NewTranscriptionTask: get token: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token.Token)

	res, err := s.deps.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Salute.NewTranscriptionTask: http response: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("Salute.NewTranscriptionTask: salute API error: status=%d body=%s", res.StatusCode, body)
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Salute.NewTranscriptionTask: read response body: %w", err)
	}
	var transcriptionNewTask transcription.TranscriptionNewTaskResponse
	if err = json.Unmarshal(respBody, &transcriptionNewTask); err != nil {
		return "", fmt.Errorf("Salute.NewTranscriptionTask: unmarshal data: %w", err)
	}
	bodyStr := string(respBody)
	if len(bodyStr) > 100 {
		bodyStr = bodyStr[:100]
	}
	logger.Info(
		"Salute.NewTranscriptionTaskadFile: result",
		zap.Int("status", res.StatusCode),
		zap.String("body", bodyStr),
		zap.Any("response", transcriptionNewTask),
	)
	return transcriptionNewTask.Result.ID, nil
}

func (s *Salute) GetTaskStatus(ctx context.Context, taskID string) (*transcription.GetTaskStatusResponse, error) {
	logger := s.deps.Logger
	logger.Info("Salute.GetTaskStatus")

	var payload bytes.Buffer
	encoder := json.NewEncoder(&payload)
	encoder.Encode(transcription.GetTaskStatusRequest{
		ID: taskID,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+"/task:get?id="+taskID, &payload)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTaskStatus: new http request: %w", err)
	}
	token, err := s.Oauth.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTaskStatus: get token: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token.Token)

	res, err := s.deps.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTaskStatus: http response: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Salute.GetTaskStatus: salute API error: status=%d body=%s", res.StatusCode, body)
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTaskStatus: read response body: %w", err)
	}
	var getTaskStatusResponse transcription.GetTaskStatusResponse
	if err = json.Unmarshal(respBody, &getTaskStatusResponse); err != nil {
		logger.Error(
			"Salute.GetTaskStatus: unmarshal",
			zap.Int("status", res.StatusCode),
			zap.String("body", string(respBody)),
		)
		return nil, fmt.Errorf("Salute.GetTaskStatus: unmarshal data: %w", err)
	}
	bodyStr := string(respBody)
	if len(bodyStr) > 100 {
		bodyStr = bodyStr[:100]
	}
	logger.Info(
		"Salute.GetTaskStatus: result",
		zap.Int("status", res.StatusCode),
		zap.String("body", bodyStr),
		zap.Any("response", getTaskStatusResponse),
	)
	return &getTaskStatusResponse, nil
}

func (s *Salute) DownloadFile(ctx context.Context, responseFileID string) (transcription.DownloadResponse, error) {
	logger := s.deps.Logger
	logger.Info("Salute.DownloadFile")

	req, err := http.NewRequest(http.MethodGet, s.URL+"/data:download?response_file_id="+responseFileID, nil)
	if err != nil {
		return nil, fmt.Errorf("Salute.DownloadFile: new http request: %w", err)
	}
	token, err := s.Oauth.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("Salute.DownloadFile: get token: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token.Token)

	res, err := s.deps.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Salute.DownloadFile: http response: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Salute.DownloadFile: salute API error: status=%d body=%s", res.StatusCode, body)
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Salute.DownloadFile: read response body: %w", err)
	}
	var downloadResponce transcription.DownloadResponse
	if err = json.Unmarshal(respBody, &downloadResponce); err != nil {
		return nil, fmt.Errorf("Salute.DownloadFile: unmarshal data: %w", err)
	}
	bodyStr := string(respBody)
	if len(bodyStr) > 100 {
		bodyStr = bodyStr[:100]
	}
	logger.Info(
		"Salute.DownloadFile: result",
		zap.Int("status", res.StatusCode),
		//zap.String("body", bodyStr),
		zap.Any("response", downloadResponce),
	)
	return downloadResponce, nil
}
