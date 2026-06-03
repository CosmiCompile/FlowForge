package runtime

import "time"

type RunStatus string

type StepStatus string

type JobStatus string

const (
	RunQueued   RunStatus = "queued"
	RunRunning  RunStatus = "running"
	RunSuccess  RunStatus = "success"
	RunFailed   RunStatus = "failed"
	RunCanceled RunStatus = "canceled"
)

const (
	StepQueued  StepStatus = "queued"
	StepRunning StepStatus = "running"
	StepSuccess StepStatus = "success"
	StepFailed  StepStatus = "failed"
	StepSkipped StepStatus = "skipped"
)

const (
	JobPending JobStatus = "pending"
	JobLocked  JobStatus = "locked"
	JobDone    JobStatus = "done"
	JobFailed  JobStatus = "failed"
)

type Job struct {
	ID          string
	Type        string
	PayloadJSON string
	Status      JobStatus
	LockedAt    *time.Time
	LockedBy    *string
	AvailableAt time.Time
	Attempt     int
	LastError   *string
	CreatedAt   time.Time
}
