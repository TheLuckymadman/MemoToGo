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

func (s *Salute) GetTranscription(ctx context.Context, audio []byte) (*transcription.TranscriptionResponse, error) {
	logger := s.deps.Logger
	logger.Info("Salute.GetTranscription")

	reqBody := bytes.NewBuffer(audio)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTransctiption: new http request: %w", err)
	}
	token, err := s.Oauth.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTransctiption: get token: %w", err)
	}
	req.Header.Add("Content-Type", "audio/ogg;codecs=opus")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token.Token)

	resp, err := s.deps.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTransctiption: http response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Salute.GetTransctiption: salute API error: status=%d body=%s", resp.StatusCode, body)
	}

	var transcriptionResp transcription.TranscriptionResponse
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Salute.GetTransctiption: read response body: %w", err)
	}
	if err = json.Unmarshal(respBody, &transcriptionResp); err != nil {
		return nil, fmt.Errorf("Salute.GetTransctiption: unmarshal transcription data: %w", err)
	}
	logger.Info(
		"Salute.GetTransctiption: result",
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)[:100]),
		zap.Any("transcription response", transcriptionResp),
	)
	return &transcriptionResp, nil
}
