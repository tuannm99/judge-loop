# judge-loop

A discipline-driven coding practice system for algorithm and interview training.

## What it is

`judge-loop` is not just an online judge. It combines:

- **Judge server** — evaluates code submissions in isolated Docker sandboxes
- **Local agent** — runs on your machine, tracks sessions, exposes a local HTTP API
- **Neovim plugin** — integrates into your editor for reminders, timers, and submission
- **Timer / streak system** — enforces a daily training loop
- **Problem registry** — versioned index of problems from LeetCode, NeetCode, HackerRank
- **Personalization engine** — adapts daily missions based on your performance

## Core flow

1. Open Neovim → local agent checks if you practiced today → reminds you if not
2. Start a timed session from Neovim
3. Solve a problem → submit via the plugin
4. Judge server evaluates → verdict stored
5. System adapts tomorrow's tasks based on your performance

## Tech stack

| Layer | Tech |
|-------|------|
| Backend | Go, Gin |
| Database | PostgreSQL (pgx) |
| Queue | Redis + asynq |
| Sandbox | Docker |
| Plugin | Lua (Neovim) |

## Structure

```
judge-loop/                        # module: github.com/tuannm99/judge-loop
  go.mod                           # single module — no go.work, no per-package go.mod
  cmd/                             # executable entry points (main packages only)
    api-server/                    # REST API server
    judge-worker/                  # Async submission evaluator
    local-agent/                   # Local HTTP daemon (runs on dev machine)
  internal/                        # private packages — enforced by Go compiler
    domain/                        # shared domain structs (no logic, no DB)
    storage/                       # PostgreSQL query layer (pgx)
    queue/                         # asynq job type definitions
    judge/                         # verdict evaluation logic
    sandbox/                       # Docker container lifecycle
    timer/                         # timer session management
    events/                        # activity event types
    problemset/                    # problem bank queries
    registry/                      # problem registry sync
    personalization/               # daily mission + performance
    recommendation/                # suggestion engine
  plugins/
    nvim-judge-loop/               # Neovim Lua plugin
    vscode-judge-loop/             # VS Code extension (future)
  deploy/
    compose/                       # docker-compose files
    docker/                        # Dockerfiles
  docs/                            # architecture and design docs
  scripts/                         # utility scripts
```

## Quickstart

```bash
# Start infrastructure
docker compose -f deploy/compose/docker-compose.yml up -d

# Run migrations + seed
psql $DATABASE_URL < internal/storage/migrations/001_init.sql
psql $DATABASE_URL < scripts/seed_problems.sql

# Start api-server
go run ./cmd/api-server

# Start local-agent
go run ./cmd/local-agent
```

## Milestones

| # | Name | Status |
|---|------|--------|
| 1 | Bootstrap | ✅ done |
| 2 | API server MVP | ✅ done |
| 3 | Local agent MVP | pending |
| 4 | Neovim plugin MVP | pending |
| 5 | Vertical slice | pending |
| 6 | Judge worker | pending |
| 7 | Registry | pending |
| 8 | Personalization | pending |

## Docs

- [Architecture](docs/architecture.md)
- [API Reference](docs/api.md)
- [Local Agent](docs/local-agent.md)
- [Neovim Plugin](docs/nvim-plugin.md)
- [Problem Registry](docs/registry.md)
- [Personalization Engine](docs/personalization.md)
- [Roadmap](docs/roadmap.md)
