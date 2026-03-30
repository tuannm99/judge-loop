# Architecture

## Overview

`judge-loop` is a monorepo containing three services, multiple shared packages, and a Neovim plugin. The system is designed for a single user running the local agent on their development machine, connected to a (potentially self-hosted) server.

```
┌─────────────────────────────────────────────────────────────┐
│ Developer Machine                                           │
│                                                             │
│  ┌──────────────┐     ┌───────────────────────────────┐     │
│  │ Neovim       │────▶│ local-agent (:7070)           │     │
│  │ (Lua plugin) │◀────│  - session status             │     │
│  └──────────────┘     │  - timer start/stop           │     │
│                       │  - submit proxy               │     │
│                       │  - registry sync              │     │
│                       └────────────┬──────────────────┘     │
└────────────────────────────────────│────────────────────────┘
                                     │ HTTP
                        ┌────────────▼──────────────────┐
                        │ api-server (:8080)             │
                        │  - problems CRUD              │
                        │  - submission intake          │
                        │  - progress / streak          │
                        │  - timer persistence          │
                        │  - daily mission              │
                        └────────────┬──────────────────┘
                                     │
                    ┌────────────────┼────────────────────┐
                    │                │                    │
             ┌──────▼──────┐  ┌──────▼──────┐   ┌────────▼──────┐
             │ PostgreSQL  │  │ Redis       │   │ judge-worker  │
             │ (state)     │  │ (queue/TTL) │   │ (async eval)  │
             └─────────────┘  └─────────────┘   └───────┬───────┘
                                                        │
                                                 ┌──────▼──────┐
                                                 │ Docker      │
                                                 │ sandbox     │
                                                 └─────────────┘
```

## Services

### api-server

- Gin HTTP server on port 8080
- Handles all persistent state reads/writes
- Enqueues submissions to Redis via asynq
- Returns submission status via polling

### judge-worker

- Consumes submission jobs from Redis queue
- Spawns Docker containers for isolated execution
- Writes verdict back to PostgreSQL
- Notifies api-server (or client polls)

### local-agent

- Lightweight HTTP daemon on port 7070 (localhost only)
- Proxies submissions to api-server
- Maintains local timer state
- Syncs problem registry from server
- Checks daily practice status on startup

## Module layout

Single Go module: `github.com/tuannm99/judge-loop`

```
cmd/           ← main packages (entry points only, no business logic)
internal/      ← all business logic (Go compiler prevents external import)
```

| Package (`internal/...`) | Responsibility                                 |
| ------------------------ | ---------------------------------------------- |
| `domain`                 | Shared Go structs — no logic, no DB            |
| `storage`                | PostgreSQL queries via pgx, migrations         |
| `queue`                  | asynq job type definitions                     |
| `judge`                  | Verdict scoring logic (language-agnostic)      |
| `sandbox`                | Docker container lifecycle for code execution  |
| `timer`                  | Timer session tracking and persistence         |
| `events`                 | ActivityEvent definitions and publishing       |
| `problemset`             | Problem bank queries and filtering             |
| `registry`               | Registry index sync and manifest parsing       |
| `personalization`        | Daily mission generation, performance analysis |
| `recommendation`         | Problem suggestion based on profile            |

## Data flow: submission

```
User (Neovim)
  → POST /local/submit (local-agent)
    → validate + attach session context
    → POST /api/submissions (api-server)
      → write submission row (status=pending)
      → enqueue job to Redis (asynq)
      → return submission_id
  → poll GET /api/submissions/:id
    judge-worker picks up job
      → pull language image
      → run code in Docker (timeout, no network)
      → compare output vs test cases
      → write verdict
  → poll returns verdict
```

## Database

PostgreSQL is the source of truth for all user data, submissions, and progress.
See `internal/storage/migrations/001_init.sql` for the full schema.

## Queue

Redis + asynq for async submission evaluation.

- Queue: `submissions`
- Retry: 3 times with exponential backoff
- Dead letter: stored in Redis for inspection

## Assumptions

- Single-user MVP: no multi-tenancy, no auth in v1
- Local agent runs on developer's machine (localhost only)
- Judge worker and api-server can run locally or on a remote server
- Problem statements are NOT stored — only metadata (manifest)
- Docker must be available on the judge-worker host
