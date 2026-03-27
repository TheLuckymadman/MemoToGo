package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type SummaryRepo struct {
	db *sql.DB
}

func NewSummaryTextRepo(db *sql.DB) *SummaryRepo {
	return &SummaryRepo{db: db}
}

func (s *SummaryRepo) EnqueueByStatus(ctx context.Context, curStatus string, setStatus string) (meetingIDs []int64, err error) {
	errPrefix := "SummaryTextRepo.EnqueByStatus:"
	tx, err := s.db.BeginTx(ctx, nil)
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
	SELECT meeting_id 
	FROM transcription_tasks
	WHERE summary_status = $1 OR summary_status is NULL
	FOR UPDATE SKIP LOCKED
	`, curStatus)
	if err != nil {
		err = fmt.Errorf("%s DB query: %w", errPrefix, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task sql.NullInt64
		if err = rows.Scan(&task); err != nil {
			err = fmt.Errorf("%s scan data: %w", errPrefix, err)
			return
		}
		if task.Valid {
			meetingIDs = append(meetingIDs, task.Int64)
		}
	}

	if err = rows.Err(); err != nil {
		err = fmt.Errorf("%s scan data: %w", errPrefix, err)
		return
	}

	_, err = tx.ExecContext(ctx, `
	UPDATE transcription_tasks
	SET summary_status = $1 
	WHERE meeting_id = ANY($2)
	`, setStatus, pq.Array(meetingIDs))
	if err != nil {
		err = fmt.Errorf("%s DB query: %w", errPrefix, err)
		return
	}
	return
}

func (s *SummaryRepo) UpdFailureStatusByID(
	ctx context.Context,
	meetingID int64,
	updSummaryStatus string,
) error {
	_, err := s.db.ExecContext(ctx, `
	UPDATE transcription_tasks
	SET summary_status = $1, summary_created = now()
	WHERE meeting_id = $2
	`, updSummaryStatus, meetingID)
	if err != nil {
		return fmt.Errorf("SummaryRepo.UpdFailureStatusByID: %w", err)
	}

	return nil
}

func (s *SummaryRepo) UpdSuccessStatusByID(
	ctx context.Context,
	meetingID int64,
	updSummaryStatus string,
	summary string,
) (err error) {
	errPrefix := "TranscriptionTaskRepo.UpdFinStatusByID:"
	tx, err := s.db.BeginTx(ctx, nil)
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

	_, err = tx.ExecContext(ctx, `
	UPDATE transcription_tasks
	SET summary_status = $1, summary_created = now()
	WHERE meeting_id = $2
	`, updSummaryStatus, meetingID)
	if err != nil {
		err = fmt.Errorf("%s %w", errPrefix, err)
		return
	}

	res, err := tx.ExecContext(ctx, `
	UPDATE meetings
	SET summary = $1
	WHERE id = $2
	`, summary, meetingID)
	if err != nil {
		err = fmt.Errorf("%s DB query: %w", errPrefix, err)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("%s no transcription_task found for meeting_id=%d", errPrefix, meetingID)
	}

	return
}
