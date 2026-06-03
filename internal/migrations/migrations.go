package migrations

// Minimal embedded migrations for MVP.
// Later: replace with a proper migration tool, but this keeps bootstrap simple.

type Migration struct {
	Version int
	UpSQL   string
}

var All = []Migration{
	{
		Version: 1,
		UpSQL: `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS workflows (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	slug TEXT NOT NULL,
	description TEXT,
	current_version_id TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS workflow_versions (
	id TEXT PRIMARY KEY,
	workflow_id TEXT NOT NULL,
	version TEXT NOT NULL,
	source_type TEXT NOT NULL,
	source_ref TEXT,
	compiled_ir_json TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (workflow_id) REFERENCES workflows(id)
);

CREATE TABLE IF NOT EXISTS triggers (
	id TEXT PRIMARY KEY,
	workflow_id TEXT NOT NULL,
	type TEXT NOT NULL,
	config_json TEXT,
	enabled INTEGER NOT NULL DEFAULT 1,
	last_fired_at TEXT,
	FOREIGN KEY (workflow_id) REFERENCES workflows(id)
);

CREATE TABLE IF NOT EXISTS runs (
	id TEXT PRIMARY KEY,
	workflow_version_id TEXT NOT NULL,
	trigger_id TEXT,
	event_id TEXT,
	status TEXT NOT NULL,
	started_at TEXT,
	finished_at TEXT,
	error TEXT,
	correlation_id TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (workflow_version_id) REFERENCES workflow_versions(id)
);

CREATE TABLE IF NOT EXISTS run_steps (
	id TEXT PRIMARY KEY,
	run_id TEXT NOT NULL,
	step_key TEXT NOT NULL,
	status TEXT NOT NULL,
	attempt INTEGER NOT NULL DEFAULT 0,
	started_at TEXT,
	finished_at TEXT,
	error TEXT,
	output_json TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (run_id) REFERENCES runs(id)
);

-- Simple in-DB queue for MVP
CREATE TABLE IF NOT EXISTS jobs (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	payload_json TEXT NOT NULL,
	status TEXT NOT NULL,
	locked_at TEXT,
	locked_by TEXT,
	available_at TEXT NOT NULL,
	attempt INTEGER NOT NULL DEFAULT 0,
	last_error TEXT,
	created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_jobs_available ON jobs(status, available_at);
CREATE INDEX IF NOT EXISTS idx_runs_created ON runs(created_at);
CREATE INDEX IF NOT EXISTS idx_run_steps_run ON run_steps(run_id);
`,
	},
}
