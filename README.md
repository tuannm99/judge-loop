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

| Layer      | Tech              |
| ---------- | ----------------- |
| Backend    | Go, Gin           |
| Database   | PostgreSQL (GORM) |
| Migrations | goose             |
| Queue      | Redis + asynq     |
| Sandbox    | Docker            |
| Plugin     | Lua (Neovim)      |

## Structure

```
judge-loop/                        # module: github.com/tuannm99/judge-loop
  go.mod                           # single module — no go.work, no per-package go.mod
  cmd/                             # executable entry points (main packages only)
    migrate/                       # Run goose migrations manually
    api-server/                    # REST API server
    judge-worker/                  # Async submission evaluator
    local-agent/                   # Local HTTP daemon (runs on dev machine)
  internal/                        # private application code
    domain/                        # core domain types and pure domain logic
      judge/                       # verdict evaluation logic
    application/                   # use cases and orchestration
    port/
      in/                          # inbound ports exposed to adapters, split by capability
      out/                         # outbound ports required by application, split by dependency
    adapter/                       # delivery and integration adapters
      http/                        # Gin handlers for api-server and local-agent
      queue/                       # asynq-facing adapters
      sandbox/                     # code runner adapter
    infrastructure/                # concrete technical implementations
      postgres/                    # PostgreSQL persistence via GORM + goose; implements outbound repository ports
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
make infra          # 1. start PostgreSQL + Redis
make migrate        # 2. run migrations (required before first start)
make api-server     # 3. start API server
make judge-worker   # 4. start judge worker
make local-agent    # 5. start local agent
```

Or without Make:

```bash
# 1. Start infrastructure (PostgreSQL + Redis)
docker compose -f deploy/compose/docker-compose.yml up -d

# 2. Run database migrations (must be done before first start, and after pulling new migrations)
DATABASE_URL=postgres://judgeloop:judgeloop@localhost:5432/judgeloop?sslmode=disable \
  go run ./cmd/migrate

# 3. Start api-server
go run ./cmd/api-server

# 4. Start judge-worker
go run ./cmd/judge-worker

# 5. Start local-agent
go run ./cmd/local-agent

# Optional: seed development problems + test cases
psql $DATABASE_URL < scripts/seed_problems.sql
```

Migrations are **not** run automatically on startup. Run `cmd/migrate` manually before the first start and whenever new migration files are added.

### Environment variables

**`cmd/migrate` and `cmd/api-server` and `cmd/judge-worker`**

| Variable       | Default                                                                   | Description    |
| -------------- | ------------------------------------------------------------------------- | -------------- |
| `DATABASE_URL` | `postgres://judgeloop:judgeloop@localhost:5432/judgeloop?sslmode=disable` | PostgreSQL DSN |

**`cmd/api-server`**

| Variable    | Default                                | Description                     |
| ----------- | -------------------------------------- | ------------------------------- |
| `REDIS_URL` | `localhost:6379`                       | Redis address for the job queue |
| `PORT`      | `8080`                                 | HTTP listen port                |
| `USER_ID`   | `00000000-0000-0000-0000-000000000001` | UUID of the local user account  |

**`cmd/judge-worker`**

| Variable          | Default          | Description                           |
| ----------------- | ---------------- | ------------------------------------- |
| `REDIS_URL`       | `localhost:6379` | Redis address                         |
| `CONCURRENCY`     | `2`              | Number of parallel evaluation workers |
| `TIME_LIMIT_SECS` | `10`             | Per-submission execution time limit   |

**`cmd/local-agent`**

| Variable              | Default                                | Description                        |
| --------------------- | -------------------------------------- | ---------------------------------- |
| `JUDGE_SERVER_URL`    | `http://localhost:8080`                | URL of the api-server              |
| `JUDGE_PORT`          | `7070`                                 | Local agent listen port            |
| `JUDGE_USER_ID`       | `00000000-0000-0000-0000-000000000001` | UUID of the local user account     |
| `JUDGE_REGISTRY_PATH` | `./registry`                           | Path to the local problem registry |
| `JUDGE_DATA_DIR`      | OS-specific user data dir              | Directory for local agent state    |

## Architecture

The codebase now follows a ports-and-adapters style:

- `domain` contains core business types and pure logic
- `application` contains use cases
- `application` owns concrete use-case services and mission-generation helpers
- `port/in` defines what adapters can call
- `port/out` defines what the application needs from infrastructure
- `adapter` contains HTTP, queue, and sandbox adapters
- `infrastructure` contains concrete PostgreSQL, queue, sandbox, registry, and local timer implementations

The dependency rule is simple: outer layers depend inward. `cmd/*` wires the graph; business logic stays out of entrypoints.

## Milestones

| #   | Name              | Status  |
| --- | ----------------- | ------- |
| 1   | Bootstrap         | ✅ done |
| 2   | API server MVP    | ✅ done |
| 3   | Local agent MVP   | pending |
| 4   | Neovim plugin MVP | pending |
| 5   | Vertical slice    | pending |
| 6   | Judge worker      | pending |
| 7   | Registry          | pending |
| 8   | Personalization   | pending |

## Docs

- [Architecture](docs/architecture.md)
- [API Reference](docs/api.md)
- [Local Agent](docs/local-agent.md)
- [Neovim Plugin](docs/nvim-plugin.md)
- [Problem Registry](docs/registry.md)
- [Personalization Engine](docs/personalization.md)
- [Roadmap](docs/roadmap.md)
