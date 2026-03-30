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

func (s *Salute) GetTranscriptionSync(ctx context.Context, audio []byte) (*TranscriptionResponseSync, error) {
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

	var transcriptionResp TranscriptionResponseSync
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
