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

		requestFileID, err := w.transcriptor.UploadFile(ctx, job.Audio)
		if err != nil {
			logger.Error(
				"Worker: upload transcription",
				zap.Error(err),
			)
			job.RcvChan <- errMsg
			continue
		}

		taskID, err := w.transcriptor.NewTranscriptionTask(ctx, requestFileID)
		if err != nil {
			logger.Error(
				"Worker: new transcription task",
				zap.Error(err),
			)
			job.RcvChan <- errMsg
			continue
		}

		taskStatus, err := w.transcriptor.GetTaskStatus(ctx, taskID) // NEW, RUNNING, CANCELED, DONE, ERROR
		if err != nil {
			logger.Error(
				"Worker: get transcription status task",
				zap.Error(err),
			)
			job.RcvChan <- errMsg
			continue
		}
		for taskStatus.Result.Status == "NEW" || taskStatus.Result.Status == "RUNNING" {
			time.Sleep(time.Second * 10)
			taskStatus, err = w.transcriptor.GetTaskStatus(ctx, taskID) // NEW, RUNNING, CANCELED, DONE, ERROR
			if err != nil {
				logger.Error(
					"Worker: get transcription status task",
					zap.Error(err),
				)
			}
		}
		if taskStatus.Result.Status != "DONE" {
			logger.Error(
				"Worker: failed transcription status task",
				zap.Error(err),
				zap.Any("status", taskStatus),
			)
			job.RcvChan <- errMsg
			continue
		}

		transcriptionResp, err := w.transcriptor.DownloadFile(ctx, taskStatus.Result.ResponseFileID)
		if err != nil {
			logger.Error(
				"Worker: download transcription",
				zap.Error(err),
			)
			job.RcvChan <- errMsg
			continue
		}

		var buildStr strings.Builder
		// slices.SortFunc(transcriptionResp, func(i, j transcription.DownloadItem) int {
		// 	if i.ProcessedAudioStart < j.ProcessedAudioStart {
		// 		return -1
		// 	}
		// 	if i.ProcessedAudioStart > j.ProcessedAudioStart {
		// 		return 1
		// 	}
		// 	return 0
		// })

		for _, resp := range transcriptionResp {
			for _, res := range resp.Results {
				if res.NormalizedText != "" {
					buildStr.WriteString(res.NormalizedText)
					buildStr.WriteString("\n")
				}
			}
		}
		transcriptionText := buildStr.String()
		// transcriptionResp, err := w.transcriptor.GetTranscriptionSync(ctx, job.Audio)
		// if err != nil {
		// 	logger.Error(
		// 		"Worker: send to transcription",
		// 		zap.Error(err),
		// 	)
		// 	job.RcvChan <- errMsg
		// 	//w.bot.Send(&telebot.Chat{ID: job.ChatID}, "Something is going wrong, but we already know about it and are working on fixing the issue")
		// 	continue
		// }
		//logger.Info("Worker: transcription received", zap.Int64("ChatID", job.ChatID), zap.String("transcription", transcriptionResp.Result[0]+"..."))

		//transcriptionText := strings.Join(transcriptionResp.Result, "\n")

		llmResp, err := w.llm.Chat(
			ctx,
			[]model.ReqMessage{
				{
					Role:    "system",
					Content: llm.SystemPrompt,
				},
				{
					Role:    "user",
					Content: transcriptionText,
				},
			},
			[]model.Function{},
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
				//Emotions:      transcriptionResp.Emotions,
				Topics: model.StringList{},
				UserID: job.ChatID,
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
