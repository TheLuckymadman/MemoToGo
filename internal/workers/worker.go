package workers

import (
	"context"
	"strings"
	"time"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/llm"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/queue"
	"github.com/theluckymadman/memotogo/internal/repository"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
)

type Worker struct {
	queue        *queue.Queue
	bot          *telebot.Bot
	transcriptor Transcriptor
	llm          LLM
	meetsRepo    *repository.Storage[model.Meeting]
	deps         *deps.Deps
}

func NewWorker(
	queue *queue.Queue,
	bot *telebot.Bot,
	transcriptor Transcriptor,
	llm LLM,
	meetsRepo *repository.Storage[model.Meeting],
	deps *deps.Deps,
) *Worker {
	return &Worker{
		queue:        queue,
		bot:          bot,
		transcriptor: transcriptor,
		llm:          llm,
		meetsRepo:    meetsRepo,
		deps:         deps,
	}
}

func (w *Worker) Start() {
	logger := w.deps.Logger
	errMsg := w.deps.ErrMsg
	logger.Info("Worker is starting")
	for job := range w.queue.Jobs {
		logger.Info("Worker: job is being processed", zap.Int64("ChatID", job.ChatID))
		ctx, stop := context.WithTimeout(context.Background(), time.Minute*5)
		defer stop()

		transcriptionResp, err := w.transcriptor.GetTranscription(ctx, job.Audio)
		if err != nil {
			logger.Error(
				"Worker: send to transcription",
				zap.Error(err),
			)
			job.RcvChan <- errMsg
			//w.bot.Send(&telebot.Chat{ID: job.ChatID}, "Something is going wrong, but we already know about it and are working on fixing the issue")
			continue
		}
		logger.Info("Worker: transcription received", zap.Int64("ChatID", job.ChatID), zap.String("transcription", transcriptionResp.Result[0]+"..."))

		transcriptionText := strings.Join(transcriptionResp.Result, "\n")

		llmResp, err := w.llm.Chat(
			ctx,
			[]llm.Message{
				{
					Role:    "system",
					Content: llm.SystemPrompt,
				},
				{
					Role:    "user",
					Content: transcriptionText,
				},
			},
			[]llm.Tool{},
		)
		if err != nil {
			logger.Error(
				"Worker: send to LLM",
				zap.Error(err),
			)
			job.RcvChan <- errMsg
			//w.bot.Send(&telebot.Chat{ID: job.ChatID}, "Something is going wrong, but we already know about it and are working on fixing the issue")
			continue
		}
		var resp string
		if llmResp == nil {
			logger.Info("Worker: empty llm response received", zap.Int64("ChatID", job.ChatID))
			resp = "empty LLM answer"
		}
		if len(llmResp.Choices) > 0 && llmResp.Choices[0].Message.Role == "assistant" {
			logger.Info("Worker: llm response received", zap.Int64("ChatID", job.ChatID), zap.String("response", llmResp.Choices[0].Message.Content[:100]+"..."))
			resp = llmResp.Choices[0].Message.Content
		} else {
			logger.Info("Worker: icorrect llm response received", zap.Int64("ChatID", job.ChatID), zap.Any("response", llmResp))
			resp = "incorrect LLM answer"
		}

		err = w.meetsRepo.Create(
			ctx,
			model.Meeting{
				Date:          time.Now(),
				Duration:      job.AudioDuration,
				Transcription: transcriptionText,
				Summary:       resp,
				Emotions:      transcriptionResp.Emotions,
				Topics:        model.StringList{},
				UserID:        job.ChatID,
			},
		)
		if err != nil {
			logger.Error(
				"Worker: save meeting to DB",
				zap.Error(err),
			)
			job.RcvChan <- errMsg
			continue
		}
		job.RcvChan <- resp
		//w.bot.Send(&telebot.Chat{ID: job.ChatID}, fmt.Sprintf("Here is the transcription: %v", llmResp.Choices[0].Message.Content))
	}
}
