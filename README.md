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
| Database | PostgreSQL (GORM) |
| Migrations | goose |
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
  internal/                        # private application code
    domain/                        # core domain types and pure domain logic
      judge/                       # verdict evaluation logic
    application/                   # use cases and orchestration
      personalization/             # daily mission generation / weak-pattern logic
    port/
      in/                          # inbound ports exposed to adapters
      out/                         # outbound ports required by application
    adapter/                       # delivery and integration adapters
      http/                        # Gin handlers for api-server and local-agent
      queue/                       # asynq-facing adapters
      sandbox/                     # code runner adapter
      storage/                     # repository adapters over postgres stores
    infrastructure/                # concrete technical implementations
      postgres/                    # PostgreSQL persistence via GORM + goose
      queue/                       # asynq jobs and queue wiring
      sandbox/                     # Docker sandbox execution
      registry/                    # local registry manifest loading
      localtimer/                  # in-memory local-agent timer
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

# Start api-server (runs goose migrations on boot)
go run ./cmd/api-server

# Start judge-worker (also runs goose migrations on boot)
go run ./cmd/judge-worker

# Start local-agent
go run ./cmd/local-agent

# Optional: seed development problems + test cases
psql $DATABASE_URL < scripts/seed_problems.sql
```

`internal/infrastructure/postgres/migrations/` is managed by goose and embedded into the Go binaries. You do not need to run the schema SQL manually for normal local startup.

## Architecture

The codebase now follows a ports-and-adapters style:

- `domain` contains core business types and pure logic
- `application` contains use cases
- `port/in` defines what adapters can call
- `port/out` defines what the application needs from infrastructure
- `adapter` contains HTTP, queue, sandbox, and repository adapters
- `infrastructure` contains concrete PostgreSQL, queue, sandbox, registry, and local timer implementations

The dependency rule is simple: outer layers depend inward. `cmd/*` wires the graph; business logic stays out of entrypoints.

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
