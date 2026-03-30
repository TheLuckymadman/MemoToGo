package model

import "time"

type TranscriptionTask struct {
	ID               int64     `json:"id" db:"id,omitempty"`
	ChatID           int64     `json:"chat_id" db:"chat_id"`
	TranscriptTaskID string    `json:"transcription_task_id" db:"transcription_task_id"`
	TranscriptStatus string    `json:"transcription_status" db:"transcription_status"`
	TaskStatus       string    `json:"task_status" db:"task_status"`
	RetryCnt         int64     `json:"retry_cnt" db:"retry_cnt"`
	UploadFileID     string    `json:"upload_file_id" db:"upload_file_id"`
	DownloadFileID   string    `json:"download_file_id" db:"download_file_id"`
	Duration         int       `json:"duration" db:"duration"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	MeetingID        time.Time `json:"meeting_id" db:"meeting_id"`
}

type Transcription struct {
	Result string `json:"result"`
}

const (
	TranscriptionNEW      = "NEW"
	TranscriptionRUNNING  = "RUNNING"
	TranscriptionCANCELED = "CANCELED"
	TranscriptionDONE     = "DONE"
	TranscriptionERROR    = "ERROR"
	TaskINCOMPLETE        = "INCOMPLETE"
	TaskPICKEDUP          = "PICKEDUP"
	TaskCOMPLETE          = "COMPLETE"
)
