# Roadmap

## Milestone 1 — Bootstrap ✅
- Monorepo directory structure (`cmd/`, `internal/`)
- Documentation
- Docker Compose (PostgreSQL + Redis)
- DB schema + seed problems
- Domain models
- Single Go module (`github.com/tuannm99/judge-loop`, Go 1.25.1)

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

## Milestone 3 — Golang Architecture
- DO nothing

## Milestone 4 — Local Agent MVP
- `GET /local/status/today` — practiced today?
- `GET /local/timer/current` — active timer
- `POST /local/timer/start` — begin timed session
- `POST /local/timer/stop` — end timed session
- `POST /local/submit` — proxy to api-server
- `POST /local/sync` — sync registry from server

## Milestone 5 — Neovim Plugin MVP
- On-startup check: call local agent, show reminder if no practice
- `:JudgeStart` — start timer
- `:JudgeStop` — stop timer
- `:JudgeSubmit` — submit current buffer
- `:JudgeStatus` — show today's status
- `:JudgeMission` — show daily mission

## Milestone 6 — Vertical Slice
- End-to-end: Neovim → agent → server → judge → verdict back to plugin
- Timer persisted and shown in status line
- Streak incremented on first daily solve

## Milestone 7 — Judge Worker
- Docker sandbox execution
- Python support
- Go support
- Verdicts: AC, WA, CE, RE, TLE
- Resource limits: CPU, memory, time

## Milestone 8 — Registry
- `index.json` with versioned manifest list
- Provider manifests: leetcode, neetcode, hackerrank
- Track manifests: blind75, neetcode150, patterns
- Local agent sync command
- Problem import into local bank

## Milestone 9 — Personalization
- Daily mission generation
- Performance snapshot (avg solve time, attempts)
- Weak pattern detection
- Self-comparison over time
- Spaced repetition review scheduling
