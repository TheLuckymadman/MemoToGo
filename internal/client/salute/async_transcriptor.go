package salute

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// func (s *Salute) CreateTranscriptionTask(ctx context.Context, audio []byte) (string, error) {
// 	logger := s.deps.Logger
// 	logger.Info("Salute.GetTranscriptionAsync")
// 	requestFileID, err := s.uploadFile(ctx, audio)
// 	if err != nil {
// 		logger.Error(
// 			"Salute.GetTranscriptionAsync: upload transcription file",
// 			zap.Error(err),
// 		)
// 		return "", fmt.Errorf("Salute.GetTranscriptionAsync: upload transcription file: %w", err)
// 	}

// 	taskID, err := s.newTranscriptionTask(ctx, requestFileID)
// 	if err != nil {
// 		logger.Error(
// 			"Salute.GetTranscriptionAsync: new transcription task",
// 			zap.Error(err),
// 		)
// 		return "", fmt.Errorf("Salute.GetTranscriptionAsync: create transcription task: %w", err)
// 	}
// 	return taskID, nil
// }

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
	var transcriptionUpload UploadResponse
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

	task := TaskRequest{
		Options: Options{
			Model:                 "general",
			AudioEncoding:         "OPUS",
			SampleRate:            16000,
			Language:              "ru-RU",
			EnableProfanityFilter: false,
			HypothesesCount:       1,
			NoSpeechTimeout:       "0s",
			MaxSpeechTimeout:      "0s",
			Hints: Hints{
				Words:         []string{},
				EnableLetters: false,
				EouTimeout:    "0s",
			},
			ChannelsCount: 1,
			SpeakerSeparation: SpeakerSeparationOptions{
				Enable:                false,
				EnableOnlyMainSpeaker: false,
				Count:                 1,
			},
			InsightModels: []string{"csi", "call_features"},
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
	var transcriptionNewTask NewTaskResponse
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

func (s *Salute) GetTaskStatus(ctx context.Context, taskID string) (*GetTaskStatusResponse, error) {
	logger := s.deps.Logger
	logger.Info("Salute.GetTaskStatus")

	var payload bytes.Buffer
	encoder := json.NewEncoder(&payload)
	encoder.Encode(GetTaskStatusRequest{
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
	var getTaskStatusResponse GetTaskStatusResponse
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

func (s *Salute) DownloadFile(ctx context.Context, responseFileID string) (*DownloadResponse, error) {
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

	logger.Info(
		"Salute.DownloadFile: response",
		zap.Int("status", res.StatusCode),
		//zap.String("body", bodyStr),
		zap.String("body", string(respBody)),
	)

	var downloadResponce DownloadResponse
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
	return &downloadResponce, nil
}
