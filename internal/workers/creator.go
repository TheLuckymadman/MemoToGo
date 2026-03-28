package workers

import (
	"context"
	"fmt"
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

func (t *TaskCreator) Start(doneCtx context.Context, workerID int) {
	errMsg := t.deps.ErrMsg
	logger := t.deps.Logger
	msgPrefix := fmt.Sprintf("Transcription task creator id: %d:", workerID)
	logger.Info(fmt.Sprintf("%s starting", msgPrefix))

	createJob := func(parentDoneCtx context.Context, job queue.SpeechJob) {
		ctx, stop := context.WithTimeout(parentDoneCtx, time.Minute*1)
		defer stop()

		task, err := t.transcriptor.CreateTask(ctx, job.Audio)
		if err != nil {
			logger.Error(
				fmt.Sprintf("%s create transcription task", msgPrefix),
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

		if err = t.taskRepo.InsertTask(ctx, task.TranscriptTaskID, task.TranscriptStatus, task.TaskStatus, task.ChatID, task.Duration); err != nil {
			logger.Error(
				fmt.Sprintf("%s insert transcription task to DB", msgPrefix),
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
			for {
				select {
				case job, ok := <-t.queue.Jobs:
					if !ok {
						break
					}
					select {
					case job.RcvChan <- errMsg:
					case <-doneCtx.Done():
					}
				default:
					logger.Info(fmt.Sprintf("%s stopped gracefully", msgPrefix))
					return
				}
			}
		case job, ok := <-t.queue.Jobs:
			if !ok {
				logger.Info(fmt.Sprintf("%s queue closed", msgPrefix))
				return
			}
			createJob(doneCtx, job)
		}
	}
}
