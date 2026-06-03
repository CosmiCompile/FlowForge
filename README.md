# FlowForge

Local-first, AI-native automation platform.

## Goals
- Local-first + offline capable
- Self-hosted (Docker)
- Workflows as code (YAML/JSON/TypeScript)
- Extensible plugin system

## Repo layout (WIP)
- `cmd/flowforge-server` — Go API + scheduler + coordinator
- `cmd/flowforge-worker` — Go worker process (WIP; server runs an in-process worker for MVP)

## Running
### Local
```bash
go run ./cmd/flowforge-server
curl -sS localhost:8080/health
```

### Manual run (end-to-end MVP slice)
```bash
curl -sS -X POST localhost:8080/api/v1/runs/manual \
  -H 'content-type: application/json' \
  -d '{"name":"hello"}'

# then fetch status:
# curl -sS localhost:8080/api/v1/runs/<runId>
```

### Docker
```bash
docker compose up --build
```

## Status
Early scaffold + DB + minimal job queue + manual run endpoint.
