package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/CosmiCompile/FlowForge/internal/queue"
	"github.com/CosmiCompile/FlowForge/internal/runtime"
	"github.com/CosmiCompile/FlowForge/internal/util"
	"github.com/go-chi/chi/v5"
)

type API struct {
	db    *sql.DB
	queue *queue.Queue
}

func New(db *sql.DB) *API {
	return &API{db: db, queue: queue.New(db)}
}

func (a *API) Router() http.Handler {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/runs/manual", a.handleManualRun)
		r.Get("/runs/{runID}", a.handleGetRun)
	})

	return r
}

type manualRunRequest struct {
	// For now, just a name to label the run; later this will reference workflow/version.
	Name string `json:"name"`
}

type manualRunResponse struct {
	RunID string `json:"runId"`
	JobID string `json:"jobId"`
}

func (a *API) handleManualRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req manualRunRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Name == "" {
		req.Name = "manual"
	}

	runID := util.NewID("run")
	now := time.Now().UTC().Format(time.RFC3339Nano)

	// Minimal run row; workflow_version_id is stubbed for now.
	_, err := a.db.ExecContext(ctx, `
INSERT INTO runs(id, workflow_version_id, status, created_at, started_at, correlation_id)
VALUES(?, ?, ?, ?, ?, ?)
`, runID, "v0", runtime.RunQueued, now, now, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload, _ := json.Marshal(map[string]any{"runId": runID})
	jobID, err := a.queue.Enqueue(ctx, "run.execute", string(payload), time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusAccepted, manualRunResponse{RunID: runID, JobID: jobID})
}

func (a *API) handleGetRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	runID := chi.URLParam(r, "runID")

	row := a.db.QueryRowContext(ctx, `SELECT id, status, error, started_at, finished_at, correlation_id FROM runs WHERE id = ?`, runID)
	var id, status string
	var errStr, startedAt, finishedAt, corr sql.NullString
	if err := row.Scan(&id, &status, &errStr, &startedAt, &finishedAt, &corr); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":            id,
		"status":        status,
		"error":         nullStr(errStr),
		"startedAt":     nullStr(startedAt),
		"finishedAt":    nullStr(finishedAt),
		"correlationId": nullStr(corr),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func nullStr(ns sql.NullString) any {
	if ns.Valid {
		return ns.String
	}
	return nil
}

// For future use
func withTimeout(ctx context.Context, d time.Duration) (context.Context, func()) {
	return context.WithTimeout(ctx, d)
}
