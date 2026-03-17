package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	ID             int64     `db:"id"`
	Username       string    `db:"username"`
	ChatID         int64     `db:"chat_id"`
	Created_at     time.Time `db:"created_at"`
	Approved       bool      `db:"approved"`
	WelcomeMsgSent time.Time `db:"welcome_msg_sent"`
}

type Meeting struct {
	ID             int64       `db:"id,omitempty"`
	Date           time.Time   `db:"date"`
	Duration       int         `db:"duration"`
	Transcription  string      `db:"transcription"`
	Summary        string      `db:"summary"`
	Emotions       EmotionList `db:"emotions"`
	Topics         StringList  `db:"topics"`
	UserID         int64       `db:"user_id"`
	SearchVectorEn string      `db:"search_vector_en"`
	SearchVectorRu string      `db:"search_vector_ru"`
}

type Emotion struct {
	Neutral    float64 `db:"neutral"`
	Positive   float64 `db:"positive"`
	Negative   float64 `db:"negative"`
	Neutral_a  float64 `db:"neutral_a"`
	Positive_a float64 `db:"positive_a"`
	Negative_a float64 `db:"negative_a"`
	Neutral_t  float64 `db:"neutral_t"`
	Positive_t float64 `db:"positive_t"`
	Negative_t float64 `db:"negative_t"`
}

type StringList []string

func (s *StringList) Scan(value interface{}) error {
	s, err := UnmarshalWrap[StringList](value)
	if err != nil {
		return err
	}
	return nil
}

type EmotionList []Emotion

func (e *EmotionList) Scan(value interface{}) error {
	res, err := UnmarshalWrap[EmotionList](value)
	if err != nil {
		return err
	}
	*e = *res
	return nil
}

func UnmarshalWrap[E any](value interface{}) (*E, error) {
	var e E
	bytes, ok := value.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to scan")
	}
	if err := json.Unmarshal(bytes, &e); err != nil {
		return nil, fmt.Errorf("failed unmarshal")
	}
	return &e, nil
}
