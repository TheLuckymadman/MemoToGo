package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type Meeting struct {
	ID             int64      `db:"id,omitempty"`
	Date           time.Time  `db:"date"`
	Duration       int        `db:"duration"`
	Transcription  string     `db:"transcription"`
	Summary        *string    `db:"summary"`
	Topics         StringList `db:"topics"`
	UserID         int64      `db:"user_id"`
	SearchVectorEn string     `db:"search_vector_en"`
	SearchVectorRu string     `db:"search_vector_ru"`
}

type StringList []string

func (s *StringList) Scan(value interface{}) error {
	res, err := UnmarshalWrap[StringList](value)
	if err != nil {
		return err
	}
	*s = *res
	return nil
}

func UnmarshalWrap[E any](value interface{}) (*E, error) {
	var e E
	// if value == nil {
	// 	return nil, nil
	// }
	bytes, ok := value.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to scan")
	}
	if err := json.Unmarshal(bytes, &e); err != nil {
		return nil, fmt.Errorf("failed unmarshal")
	}
	return &e, nil
}
