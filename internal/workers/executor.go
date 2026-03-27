package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/repository"
	"go.uber.org/zap"
)

type TaskExecutor struct {
	runInterval         time.Duration
	transcriptor        Transcriptor
	meetsRepo           *repository.Storage[model.Meeting]
	transcriptTasksRepo *repository.TranscriptionTaskRepo
	deps                *deps.Deps
}

func NewTaskExecutor(
	runInterval time.Duration,
	transcriptor Transcriptor,
	meetsRepo *repository.Storage[model.Meeting],
	transcriptTasksRepo *repository.TranscriptionTaskRepo,
	deps *deps.Deps,
) *TaskExecutor {
	return &TaskExecutor{
		runInterval:         runInterval,
		transcriptor:        transcriptor,
		meetsRepo:           meetsRepo,
		transcriptTasksRepo: transcriptTasksRepo,
		deps:                deps,
	}
}

func (t *TaskExecutor) Start(doneCtx context.Context) {
	logger := t.deps.Logger
	logger.Info("Task executor is starting")
	ticker := time.NewTicker(t.runInterval)
	defer ticker.Stop()

	checkAndComplete := func() error {
		logger.Info("TaskExecutor: check and complete task")
		ctx, stop := context.WithTimeout(doneCtx, time.Minute*5)
		defer stop()

		taskIDs, err := t.transcriptTasksRepo.EnqueueByStatus(ctx, model.TaskINCOMPLETE, model.TaskPICKEDUP)
		if err != nil {
			return fmt.Errorf("TaskExecutor: enque transcription tasks: %w", err)
		}

		var wg sync.WaitGroup
		reqLimit := make(chan int, 10)
		for _, taskID := range taskIDs {
			wg.Add(1)
			go func() {
				reqLimit <- 1
				defer func() {
					<-reqLimit
					wg.Done()
				}()
				task := model.TranscriptionTask{
					TranscriptTaskID: taskID,
				}
				resTask, err := t.transcriptor.GetTaskStatus(ctx, task)
				if err != nil {
					errUpd := t.transcriptTasksRepo.UpdStatusByID(ctx, taskID, model.TaskINCOMPLETE, model.TranscriptionERROR)
					if errUpd != nil {
						errMsg := fmt.Errorf("TaskExecutor: update transcription tasks %s: %w \nafter failure %w", taskID, errUpd, err)
						logger.Error(
							"TaskExecutor: update transcription tasks",
							zap.Error(errMsg),
						)
						return
					}
					errMsg := fmt.Errorf("TaskExecutor: get transcription tasks %s status: %w", taskID, err)
					logger.Error(
						"TaskExecutor: update transcription tasks",
						zap.Error(errMsg),
					)
					return
				}

				switch resTask.TranscriptStatus {
				case model.TranscriptionDONE:
					{
						transctiption, err := t.transcriptor.GetTranscription(ctx, *resTask)
						if err != nil {
							errUpd := t.transcriptTasksRepo.UpdStatusByID(ctx, taskID, model.TaskINCOMPLETE, model.TranscriptionERROR)
							if errUpd != nil {
								errMsg := fmt.Errorf("TaskExecutor: update transcription tasks %s: %w \nafter failure %w", taskID, errUpd, err)
								logger.Error(
									"TaskExecutor: update transcription tasks",
									zap.Error(errMsg),
								)
								return
							}
							errMsg := fmt.Errorf("TaskExecutor: get transcription file %s: %w", taskID, err)
							logger.Error(
								"TaskExecutor: update transcription tasks",
								zap.Error(errMsg),
							)
							return
						}
						err = t.transcriptTasksRepo.UpdSuccessStatusByID(
							ctx,
							resTask.TranscriptTaskID,
							model.TaskCOMPLETE,
							resTask.TranscriptStatus,
							resTask.UploadFileID,
							transctiption.Result,
						)
						if err != nil {
							errUpd := t.transcriptTasksRepo.UpdStatusByID(ctx, taskID, model.TaskINCOMPLETE, model.TranscriptionERROR)
							if errUpd != nil {
								errMsg := fmt.Errorf("TaskExecutor: update transcription tasks %s: %w \nafter failure %w", taskID, errUpd, err)
								logger.Error(
									"TaskExecutor: update transcription tasks",
									zap.Error(errMsg),
								)
								return
							}
							errMsg := fmt.Errorf("TaskExecutor: save transcription result %s: %w", taskID, err)
							logger.Error(
								"TaskExecutor: update transcription tasks",
								zap.Error(errMsg),
							)
							return
						}
					}
				case model.TranscriptionCANCELED, model.TranscriptionERROR:
					{
						err = t.transcriptTasksRepo.UpdFailureStatusByID(
							ctx,
							resTask.TranscriptTaskID,
							model.TaskINCOMPLETE,
							resTask.TranscriptStatus,
						)
						if err != nil {
							errUpd := t.transcriptTasksRepo.UpdStatusByID(ctx, taskID, model.TaskINCOMPLETE, model.TranscriptionERROR)
							if errUpd != nil {
								errMsg := fmt.Errorf("TaskExecutor: update transcription tasks %s: %w \nafter failure %w", taskID, errUpd, err)
								logger.Error(
									"TaskExecutor: update transcription tasks",
									zap.Error(errMsg),
								)
								return
							}
							errMsg := fmt.Errorf("TaskExecutor: save transcription status task %s: %w", taskID, err)
							logger.Error(
								"TaskExecutor: update transcription tasks",
								zap.Error(errMsg),
							)
							return
						}
					}
				default: //model.TranscriptionNEW, model.TranscriptionRUNNING:
					{
						errUpd := t.transcriptTasksRepo.UpdStatusByID(ctx, taskID, model.TaskINCOMPLETE, resTask.TranscriptStatus)
						if errUpd != nil {
							errMsg := fmt.Errorf("TaskExecutor: update transcription tasks %s: %w \nafter failure %w", taskID, errUpd, err)
							logger.Error(
								"TaskExecutor: update transcription tasks",
								zap.Error(errMsg),
							)
							return
						}
					}

				}
			}()
		}
		wg.Wait()
		return nil
	}

	for {
		select {
		case <-doneCtx.Done():
			logger.Info("TaskExecutor: stopped gracefully")
			return
		case <-ticker.C:
			err := checkAndComplete()
			if err != nil {
				logger.Error(
					"TaskExecutor",
					zap.Error(err),
				)
			}
		}
	}
}
