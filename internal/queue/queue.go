package queue

type Queue struct {
	Jobs chan SpeechJob
}

type SpeechJob struct {
	ChatID        int64
	FileID        string
	Audio         []byte
	AudioDuration int
	RcvChan       chan string
}

func NewQueue(size int64) *Queue {
	return &Queue{
		make(chan SpeechJob, size),
	}
}
