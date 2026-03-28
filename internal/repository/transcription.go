package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type TranscriptionTaskRepo struct {
	db *sql.DB
}

func NewTranscriptionRepo(db *sql.DB) *TranscriptionTaskRepo {
	return &TranscriptionTaskRepo{db: db}
}

func (r *TranscriptionTaskRepo) InsertTask(
	ctx context.Context,
	transcriptTaskID string,
	transcriptStatus string,
	taskStatus string,
	chatID int64,
	duration int,
) error {
	errPrefix := "TranscriptionTaskRepo.CreateTask:"
	_, err := r.db.ExecContext(ctx, `
	INSERT INTO transcription_tasks (transcription_task_id, transcription_status, task_status, chat_id, duration)
	VALUES ($1, $2, $3, $4, $5)
	`, transcriptTaskID, transcriptStatus, taskStatus, chatID, duration)
	if err != nil {
		return fmt.Errorf("%s DB query: %w", errPrefix, err)
	}

	return nil
}

func (r *TranscriptionTaskRepo) EnqueueByStatus(ctx context.Context, curStatus string, setStatus string) (taskIDs []string, err error) {
	errPrefix := "TranscriptionTaskRepo.EnqueByStatus:"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("%s tx creation: %w", errPrefix, err)
		return
	}
	defer func() {
		if err != nil {
			if errRb := tx.Rollback(); errRb != nil {
				err = fmt.Errorf("%s rollback err: %w \nroot error: %w ", errPrefix, errRb, err)
			}
			return
		}
		err = tx.Commit()
	}()

	rows, err := tx.QueryContext(ctx, `
	SELECT transcription_task_id 
	FROM transcription_tasks
	WHERE task_status = $1
	FOR UPDATE SKIP LOCKED
	`, curStatus)
	if err != nil {
		err = fmt.Errorf("%s DB query: %w", errPrefix, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task string
		if err = rows.Scan(&task); err != nil {
			err = fmt.Errorf("%s scan data: %w", errPrefix, err)
			return
		}
		taskIDs = append(taskIDs, task)
	}

	_, err = tx.ExecContext(ctx, `
	UPDATE transcription_tasks
	SET task_status = $1, updated_at = now()
	WHERE transcription_task_id = any($2)
	`, setStatus, pq.Array(taskIDs))
	if err != nil {
		err = fmt.Errorf("%s DB query: %w", errPrefix, err)
		return
	}
	return
}

func (r *TranscriptionTaskRepo) UpdStatusByID(ctx context.Context, taskID string, updTaskStatus string, updTranscripStatus string) (err error) {
	errPrefix := "TranscriptionTaskRepo.EnqueTasksByID:"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("%s tx creation: %w", errPrefix, err)
		return
	}
	defer func() {
		if err != nil {
			if errRb := tx.Rollback(); errRb != nil {
				err = fmt.Errorf("%s rollback err: %w \nroot error: %w ", errPrefix, errRb, err)
			}
			return
		}
		err = tx.Commit()
	}()

	row := tx.QueryRowContext(ctx, `
	SELECT id
	FROM transcription_tasks
	WHERE transcription_task_id = $1
	FOR UPDATE SKIP LOCKED
	`, taskID)

	var id int64
	if err = row.Scan(&id); err != nil {
		err = fmt.Errorf("%s scan data: %w", errPrefix, err)
		return
	}

	_, err = tx.ExecContext(ctx, `
	UPDATE transcription_tasks
	SET task_status = $1, transcription_status = $2, updated_at = now(), retry_cnt = retry_cnt + 1
	WHERE id = $3
	`, updTaskStatus, updTranscripStatus, id)
	if err != nil {
		err = fmt.Errorf("%s %w", errPrefix, err)
		return
	}

	return
}

func (r *TranscriptionTaskRepo) UpdFailureStatusByID(
	ctx context.Context,
	taskID string,
	updTaskStatus string,
	updTranscripStatus string,
) (err error) {
	errPrefix := "TranscriptionTaskRepo.UpdFinStatusByID:"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("%s tx creation: %w", errPrefix, err)
		return
	}
	defer func() {
		if err != nil {
			if errRb := tx.Rollback(); errRb != nil {
				err = fmt.Errorf("%s rollback err: %w \nroot error: %w ", errPrefix, errRb, err)
			}
			return
		}
		err = tx.Commit()
	}()

	row := tx.QueryRowContext(ctx, `
	SELECT id
	FROM transcription_tasks
	WHERE transcription_task_id = $1
	FOR UPDATE SKIP LOCKED
	`, taskID)

	var id int64
	if err = row.Scan(&id); err != nil {
		err = fmt.Errorf("%s scan data: %w", errPrefix, err)
		return
	}

	_, err = tx.ExecContext(ctx, `
	UPDATE transcription_tasks
	SET task_status = $1, transcription_status = $2, updated_at = now(), retry_cnt = retry_cnt + 1
	WHERE id = $3
	`, updTaskStatus, updTranscripStatus, id)
	if err != nil {
		err = fmt.Errorf("%s %w", errPrefix, err)
		return
	}

	return
}

func (r *TranscriptionTaskRepo) UpdSuccessStatusByID(
	ctx context.Context,
	taskID string,
	updTaskStatus string,
	updTranscripStatus string,
	uploadFileID string,
	transcription string,
) (err error) {
	errPrefix := "TranscriptionTaskRepo.UpdSuccessStatusByID:"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		err = fmt.Errorf("%s tx creation: %w", errPrefix, err)
		return
	}
	defer func() {
		if err != nil {
			if errRb := tx.Rollback(); errRb != nil {
				err = fmt.Errorf("%s rollback err: %w \nroot error: %w ", errPrefix, errRb, err)
			}
			return
		}
		err = tx.Commit()
	}()

	row := tx.QueryRowContext(ctx, `
	SELECT id, chat_id, duration
	FROM transcription_tasks
	WHERE transcription_task_id = $1
	FOR UPDATE SKIP LOCKED
	`, taskID)

	var id, chatID, duration int64
	if err = row.Scan(&id, &chatID, &duration); err != nil {
		err = fmt.Errorf("%s scan data: %w", errPrefix, err)
		return
	}

	var meetingID int64
	err = tx.QueryRowContext(ctx, `
	INSERT INTO meetings (duration, transcription, user_id, topics)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`, duration, transcription, chatID, "[]").Scan(&meetingID)
	if err != nil {
		err = fmt.Errorf("%s DB query: %w", errPrefix, err)
		return
	}

	_, err = tx.ExecContext(ctx, `
	UPDATE transcription_tasks
	SET task_status = $1, transcription_status = $2, updated_at = now(), retry_cnt = retry_cnt + 1, meeting_id = $3
	WHERE id = $4
	`, updTaskStatus, updTranscripStatus, meetingID, id)
	if err != nil {
		err = fmt.Errorf("%s %w", errPrefix, err)
		return
	}

	return
}
