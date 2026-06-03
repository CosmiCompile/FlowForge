# FlowForge

Local-first, AI-native automation platform.

## Goals
- Local-first + offline capable
- Self-hosted (Docker)
- Workflows as code (YAML/JSON/TypeScript)
- Extensible plugin system

## Repo layout (WIP)
- `cmd/flowforge-server` — Go API + scheduler + coordinator
- `cmd/flowforge-worker` — Go worker process
- `packages/sdk` — TypeScript SDK (authoring)
- `apps/web` — Next.js UI (visual builder)

## Status
Scaffolded. MVP implementation in progress.
