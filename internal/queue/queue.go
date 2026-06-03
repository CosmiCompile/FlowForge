package queue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/CosmiCompile/FlowForge/internal/runtime"
	"github.com/CosmiCompile/FlowForge/internal/util"
)

type Queue struct {
	db *sql.DB
}

func New(db *sql.DB) *Queue {
	return &Queue{db: db}
}

func (q *Queue) Enqueue(ctx context.Context, jobType string, payloadJSON string, availableAt time.Time) (string, error) {
	id := util.NewID("job")
	_, err := q.db.ExecContext(ctx, `
INSERT INTO jobs(id, type, payload_json, status, available_at, attempt, created_at)
VALUES(?, ?, ?, ?, ?, 0, ?)
`, id, jobType, payloadJSON, runtime.JobPending, availableAt.UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return "", err
	}
	return id, nil
}

var ErrNoJob = errors.New("no job available")

// LockNext finds one available job and locks it for this worker.
// SQLite strategy: BEGIN IMMEDIATE + select + update; single-writer is fine for MVP.
func (q *Queue) LockNext(ctx context.Context, workerID string) (*runtime.Job, error) {
	tx, err := q.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Ensure we have a write lock.
	if _, err := tx.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	row := tx.QueryRowContext(ctx, `
SELECT id, type, payload_json, attempt, available_at
FROM jobs
WHERE status = ? AND available_at <= ?
ORDER BY available_at ASC
LIMIT 1
`, runtime.JobPending, now)

	var id, typ, payload, availableAtStr string
	var attempt int
	if err := row.Scan(&id, &typ, &payload, &attempt, &availableAtStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoJob
		}
		return nil, err
	}

	lockedAt := time.Now().UTC()
	_, err = tx.ExecContext(ctx, `
UPDATE jobs
SET status = ?, locked_at = ?, locked_by = ?
WHERE id = ? AND status = ?
`, runtime.JobLocked, lockedAt.Format(time.RFC3339Nano), workerID, id, runtime.JobPending)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	avail, _ := time.Parse(time.RFC3339Nano, availableAtStr)
	wid := workerID
	return &runtime.Job{
		ID:          id,
		Type:        typ,
		PayloadJSON: payload,
		Status:      runtime.JobLocked,
		LockedAt:    &lockedAt,
		LockedBy:    &wid,
		AvailableAt: avail,
		Attempt:     attempt,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (q *Queue) MarkDone(ctx context.Context, jobID string) error {
	_, err := q.db.ExecContext(ctx, `UPDATE jobs SET status = ? WHERE id = ?`, runtime.JobDone, jobID)
	return err
}

func (q *Queue) MarkFailed(ctx context.Context, jobID string, errMsg string, retryAfter time.Duration) error {
	availableAt := time.Now().Add(retryAfter).UTC().Format(time.RFC3339Nano)
	_, err := q.db.ExecContext(ctx, `
UPDATE jobs
SET status = ?, last_error = ?, attempt = attempt + 1, locked_at = NULL, locked_by = NULL, available_at = ?
WHERE id = ?
`, runtime.JobPending, fmt.Sprintf("%s", errMsg), availableAt, jobID)
	return err
}
