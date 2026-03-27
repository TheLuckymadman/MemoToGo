package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/llm/llmtype"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/repository"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

type Summarizer struct {
	runInterval  time.Duration
	bot          *telebot.Bot
	llm          LLM
	systemPrompt string
	meetsRepo    *repository.Storage[model.Meeting]
	summaryRepo  *repository.SummaryRepo
	deps         *deps.Deps
}

func NewSummarizer(
	runInterval time.Duration,
	bot *telebot.Bot,
	llm LLM,
	systemPrompt string,
	meetsRepo *repository.Storage[model.Meeting],
	summaryRepo *repository.SummaryRepo,
	deps *deps.Deps,
) *Summarizer {
	return &Summarizer{
		runInterval:  runInterval,
		bot:          bot,
		llm:          llm,
		systemPrompt: systemPrompt,
		meetsRepo:    meetsRepo,
		summaryRepo:  summaryRepo,
		deps:         deps,
	}
}

func (s *Summarizer) Start(ctxDone context.Context) {
	logger := s.deps.Logger
	logger.Info("Task Summarizer is starting")
	ticker := time.NewTicker(s.runInterval)
	defer ticker.Stop()

	summary := func() {
		logger.Info("Summarizer: check and complete task")
		ctx, stop := context.WithTimeout(ctxDone, time.Minute*1)
		defer stop()
		meetIDs, err := s.summaryRepo.EnqueueByStatus(ctx, model.TaskINCOMPLETE, model.TaskPICKEDUP)
		if err != nil {
			logger.Error(
				"Summarizer",
				zap.Error(err),
			)
			return
		}

		var wg sync.WaitGroup
		var errMeetIDs []int64
		var mu sync.Mutex
		reqLimit := make(chan int, 1)
		for _, meetID := range meetIDs {
			wg.Add(1)
			go func(meetID int64) {
				localCtx, stop := context.WithTimeout(ctx, time.Minute*3)
				defer stop()

				reqLimit <- 1
				defer func() {
					<-reqLimit
					wg.Done()
				}()

				meeting, err := s.meetsRepo.GetByID(localCtx, meetID)
				if err != nil {
					logger.Error(
						"Summarizer",
						zap.Error(err),
					)
					mu.Lock()
					errMeetIDs = append(errMeetIDs, meetID)
					mu.Unlock()
					return
				}

				messages := []llmtype.Message{
					{
						Role:    "system",
						Content: s.systemPrompt,
					},
					{
						Role:    "user",
						Content: meeting.Transcription,
					},
				}
				res, err := s.llm.Chat(localCtx, messages, []llmtype.Tool{})
				if err != nil {
					logger.Error(
						"Summarizer",
						zap.Error(err),
					)
					mu.Lock()
					errMeetIDs = append(errMeetIDs, meetID)
					mu.Unlock()
					return
				}

				err = s.summaryRepo.UpdSuccessStatusByID(localCtx, meetID, model.TaskCOMPLETE, res.Choices[0].Message.Content)
				if err != nil {
					logger.Error(
						"Summarizer",
						zap.Error(err),
					)
					mu.Lock()
					errMeetIDs = append(errMeetIDs, meetID)
					mu.Unlock()
					return
				}

				_, err = s.bot.Send(&telebot.Chat{ID: meeting.UserID}, fmt.Sprintf("Here is the transcription: %v", res.Choices[0].Message.Content))
				if err != nil {
					logger.Error(
						"Summarizer",
						zap.Error(err),
					)
					mu.Lock()
					errMeetIDs = append(errMeetIDs, meetID)
					mu.Unlock()
					return
				}
			}(meetID)
		}

		wg.Wait()

		if len(errMeetIDs) > 0 {
			for _, errMeetID := range errMeetIDs {
				err = s.summaryRepo.UpdFailureStatusByID(ctx, errMeetID, model.TaskINCOMPLETE)
				if err != nil {
					logger.Error(
						"Summarizer: cannot update summary status task",
						zap.Error(err),
					)
				}
			}
		}
	}
	for {
		select {
		case <-ctxDone.Done():
			logger.Info("Summarizer: stopped gracefully")
			return
		case <-ticker.C:
			summary()
		}
	}

}
