package transcription

import (
	"time"

	"github.com/theluckymadman/memotogo/internal/model"
)

// type Emotion struct {
// 	Neutral   float64 `json:"neutral"`
// 	Positive  float64 `json:"positive"`
// 	Negative  float64 `json:"negative"`
// 	NeutralA  float64 `json:"neutral_a"`
// 	PositiveA float64 `json:"positive_a"`
// 	NegativeA float64 `json:"negative_a"`
// 	NeutralT  float64 `json:"neutral_t"`
// 	PositiveT float64 `json:"positive_t"`
// 	NegativeT float64 `json:"negative_t"`
// }

type TranscriptionResponse struct {
	Result         []string               `json:"result"`
	Emotions       model.EmotionList      `json:"emotions"`
	PersonIdentity map[string]interface{} `json:"person_identity"`
	Status         int                    `json:"status"`
}

type TranscriptionUploadResponse struct {
	Status int          `json:"status"`
	Result UploadResult `json:"result"`
}

type UploadResult struct {
	RequestFileID string `json:"request_file_id"`
}

type TranscriptionTaskRequest struct {
	Options       Options `json:"options"`
	RequestFileID string  `json:"request_file_id"`
}

type Options struct {
	Model         string `json:"model"`
	AudioEncoding string `json:"audio_encoding"`
	SampleRate    int    `json:"sample_rate"`
	Language      string `json:"language"`
	// EnableProfanityFilter bool                     `json:"enable_profanity_filter"`
	// HypothesesCount       int                      `json:"hypotheses_count"`
	// NoSpeechTimeout       string                   `json:"no_speech_timeout"`
	// MaxSpeechTimeout      string                   `json:"max_speech_timeout"`
	// Hints                 Hints                    `json:"hints"`
	ChannelsCount int `json:"channels_count"`
	// SpeakerSeparation     SpeakerSeparationOptions `json:"speaker_separation_options"`
	// InsightModels         []string                 `json:"insight_models"`
}

type Hints struct {
	Words         []string `json:"words"`
	EnableLetters bool     `json:"enable_letters"`
	EouTimeout    string   `json:"eou_timeout"`
}

type SpeakerSeparationOptions struct {
	Enable                bool `json:"enable"`
	EnableOnlyMainSpeaker bool `json:"enable_only_main_speaker"`
	Count                 int  `json:"count"`
}

type TranscriptionNewTaskResponse struct {
	Status int           `json:"status"`
	Result NewTaskResult `json:"result"`
}

type NewTaskResult struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"`
}

type GetTaskStatusRequest struct {
	ID string `json:"id"`
}

type GetTaskStatusResponse struct {
	Status int              `json:"status"`
	Result TaskStatusResult `json:"result"`
}

type TaskStatusResult struct {
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Status         string    `json:"status"`
	ResponseFileID string    `json:"response_file_id"`
}

type DownloadResponse []DownloadItem

type DownloadItem struct {
	Results             []Result          `json:"results"`
	Eou                 bool              `json:"eou"`
	EmotionsResult      EmotionsResult    `json:"emotions_result"`
	ProcessedAudioStart string            `json:"processed_audio_start"`
	ProcessedAudioEnd   string            `json:"processed_audio_end"`
	BackendInfo         BackendInfo       `json:"backend_info"`
	Channel             int               `json:"channel"`
	SpeakerInfo         SpeakerInfo       `json:"speaker_info"`
	EouReason           string            `json:"eou_reason"`
	Insight             string            `json:"insight"`
	PersonIdentity      PersonIdentity    `json:"person_identity"`
}

type Result struct {
	Text           string          `json:"text"`
	NormalizedText string          `json:"normalized_text"`
	Start          string          `json:"start"`
	End            string          `json:"end"`
	WordAlignments []WordAlignment `json:"word_alignments"`
}

type WordAlignment struct {
	Word  string `json:"word"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type EmotionsResult struct {
	Positive float64 `json:"positive"`
	Neutral  float64 `json:"neutral"`
	Negative float64 `json:"negative"`
}

type BackendInfo struct {
	ModelName    string `json:"model_name"`
	ModelVersion string `json:"model_version"`
	ServerVersion string `json:"server_version"`
}

type SpeakerInfo struct {
	SpeakerID             int     `json:"speaker_id"`
	MainSpeakerConfidence float64 `json:"main_speaker_confidence"`
}

type PersonIdentity struct {
	Age         string  `json:"age"`
	Gender      string  `json:"gender"`
	AgeScore    float64 `json:"age_score"`
	GenderScore float64 `json:"gender_score"`
}