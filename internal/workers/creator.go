package workers

import (
	"context"
	"time"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/queue"
	"github.com/theluckymadman/memotogo/internal/repository"
	"go.uber.org/zap"
)

type TaskCreator struct {
	queue        *queue.Queue
	transcriptor Transcriptor
	meetsRepo    *repository.Storage[model.Meeting]
	taskRepo     *repository.TranscriptionTaskRepo
	deps         *deps.Deps
}

func NewTaskCreator(
	queue *queue.Queue,
	transcriptor Transcriptor,
	meetsRepo *repository.Storage[model.Meeting],
	taskRepo *repository.TranscriptionTaskRepo,
	deps *deps.Deps,
) *TaskCreator {
	return &TaskCreator{
		queue:        queue,
		transcriptor: transcriptor,
		meetsRepo:    meetsRepo,
		taskRepo:     taskRepo,
		deps:         deps,
	}
}

func (t *TaskCreator) Start(doneCtx context.Context) {
	errMsg := t.deps.ErrMsg
	logger := t.deps.Logger
	logger.Info("Transcription task creator is starting")

	createJob := func(parentDoneCtx context.Context, job queue.SpeechJob) {
		ctx, stop := context.WithTimeout(parentDoneCtx, time.Minute*1)
		defer stop()

		task, err := t.transcriptor.CreateTask(ctx, job.Audio)
		if err != nil {
			logger.Error(
				"Transcription task creator: create transcription task",
				zap.Error(err),
			)
			select {
			case job.RcvChan <- errMsg:
			case <-doneCtx.Done():
			}
			return
		}

		task.TaskStatus = model.TaskINCOMPLETE
		task.ChatID = job.ChatID
		task.Duration = job.AudioDuration

		if err = t.taskRepo.InsertTask(ctx, task.TranscriptTaskID, task.TranscriptStatus, task.TaskStatus, task.ChatID); err != nil {
			logger.Error(
				"Transcription task creator: insert transcription task to DB",
				zap.Error(err),
			)
			select {
			case job.RcvChan <- errMsg:
			case <-doneCtx.Done():
			}
			return
		}

		select {
		case job.RcvChan <- "Awesome:) Wait a bit, I'll respond soon.":
		case <-doneCtx.Done():
			return
		}
	}

	for {
		select {
		case <-doneCtx.Done():
			logger.Info("TaskCreator: stopped gracefully")
			return
		case job, ok := <-t.queue.Jobs:
			if !ok {
				logger.Info("queue closed")
				return
			}
			createJob(doneCtx, job)
		}
	}
}
