package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CosmiCompile/FlowForge/internal/queue"
	"github.com/CosmiCompile/FlowForge/internal/runtime"
)

type Worker struct {
	id    string
	db    *sql.DB
	queue *queue.Queue
}

func New(id string, db *sql.DB) *Worker {
	return &Worker{id: id, db: db, queue: queue.New(db)}
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		job, err := w.queue.LockNext(ctx, w.id)
		if err != nil {
			if err == queue.ErrNoJob {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return err
		}

		if err := w.handleJob(ctx, job); err != nil {
			_ = w.queue.MarkFailed(ctx, job.ID, err.Error(), 1*time.Second)
			continue
		}
		_ = w.queue.MarkDone(ctx, job.ID)
	}
}

func (w *Worker) handleJob(ctx context.Context, job *runtime.Job) error {
	switch job.Type {
	case "run.execute":
		return w.execRun(ctx, job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

func (w *Worker) execRun(ctx context.Context, job *runtime.Job) error {
	var payload struct {
		RunID string `json:"runId"`
	}
	if err := json.Unmarshal([]byte(job.PayloadJSON), &payload); err != nil {
		return err
	}
	if payload.RunID == "" {
		return fmt.Errorf("missing runId")
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := w.db.ExecContext(ctx, `UPDATE runs SET status = ?, started_at = ? WHERE id = ?`, runtime.RunRunning, now, payload.RunID)
	if err != nil {
		return err
	}

	// MVP: no steps yet; mark successful.
	finished := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = w.db.ExecContext(ctx, `UPDATE runs SET status = ?, finished_at = ? WHERE id = ?`, runtime.RunSuccess, finished, payload.RunID)
	return err
}
