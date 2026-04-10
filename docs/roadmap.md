# Roadmap

## Milestone 1 — Bootstrap ✅

- Monorepo directory structure (`cmd/`, `internal/`)
- Documentation
- Docker Compose (PostgreSQL + Redis)
- DB schema + seed problems
- Domain models
- Single Go module (`github.com/tuannm99/judge-loop`, Go 1.26.0)

## Milestone 2 — API Server MVP ✅

- `GET /api/problems` — list with filters
- `GET /api/problems/:id` — detail
- `GET /api/problems/suggest` — get suggested problem
- `POST /api/submissions` — submit code (mock verdict)
- `GET /api/submissions/:id` — poll status
- `GET /api/submissions/history` — paginated history
- `GET /api/progress/today` — daily summary
- `GET /api/streak` — current streak count
- `POST/GET /api/timers/start|stop|current` — timer management
- `GET /api/reviews/today` — spaced repetition due list

## Milestone 3 — Golang Architecture ✅

- Single-module Go layout under `cmd/` and `internal/`
- Ports-and-adapters package structure
- Split inbound and outbound ports
- Fx-based dependency wiring for runtime processes
- GORM + goose PostgreSQL persistence

## Milestone 4 — Local Agent MVP ✅

- `GET /local/status/today` — practiced today?
- `GET /local/problems` — proxy problem list to api-server
- `GET /local/problems/:id` — proxy problem detail to api-server
- `GET /local/problems/suggest` — proxy suggestion to api-server
- `GET /local/timer/current` — active timer
- `POST /local/timer/start` — begin timed session
- `POST /local/timer/stop` — end timed session
- `POST /local/submit` — proxy to api-server
- `POST /local/sync` — sync registry from server

## Milestone 5 — Neovim Plugin MVP ✅

- On-startup check: call local agent, show reminder if no practice
- `:JudgeUI` — floating workflow panel
- `:JudgeProblems` — browse problems
- `:JudgeSuggest` — pick a suggested problem
- `:JudgeLanguage` — switch solve buffer language
- `:JudgeStart` — open a problem solve buffer or start timer
- `:JudgeStop` — stop timer
- `:JudgeSubmit` — submit current buffer
- `:JudgeStatus` — show today's status
- Cached solve buffers under `~/.judgeloopcache`

## Milestone 6 — Vertical Slice ✅

- End-to-end: Neovim → agent → server → judge → verdict back to plugin
- Timer shown in status line
- Streak incremented on first daily solve

## Milestone 7 — Judge Worker ✅

- Docker sandbox execution
- Python support
- Go support
- Verdicts: AC, WA, CE, RE, TLE
- Resource limits: CPU, memory, time

## Milestone 8 — Registry ✅

- `index.json` with versioned manifest list
- Provider manifests: leetcode
- Track manifests: blind75, neetcode150
- Local agent sync command
- Problem import into local bank

## Milestone 9 — Personalization ✅

- Daily mission generation
- Performance snapshot (avg solve time, attempts)
- Weak pattern detection
- Self-comparison over time
- Spaced repetition review scheduling

## Current Focus — Editor Workflow

- Tighten the Neovim floating UI into the primary daily-use surface
- Improve structured error reporting across plugin → local-agent → api-server calls
- Keep local-agent as the plugin's single local API boundary
- Keep domain judge logic independent from sandbox infrastructure types
