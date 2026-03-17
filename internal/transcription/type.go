package transcription

import "github.com/theluckymadman/memotogo/internal/model"

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
