# Local Agent

## Purpose

The local agent is a lightweight HTTP daemon running on the developer's machine. It acts as a bridge between the Neovim plugin and the remote api-server.

It does NOT spy on the user. It only tracks what the user explicitly configures:
- Timer sessions (started by the user)
- Submission events (triggered by the user)
- Daily session state (did I practice today?)

No file watching outside configured workspace. No keylogging. No network traffic outside of explicit sync.

## Responsibilities

1. **Session gating** — On startup check, determine if the user practiced today
2. **Timer management** — Start/stop/query timed practice sessions (local state + synced to server)
3. **Submit proxy** — Attach session context to submissions before forwarding to api-server
4. **Registry sync** — Download and cache the problem registry locally
5. **Offline support** — Timer works even when server is unreachable

## Configuration

Config file: `~/.config/judge-loop/agent.yaml`

```yaml
server_url: http://localhost:8080
listen_port: 7070
workspace: ~/code/practice      # only this directory is "watched"
user_id: uuid                    # set on first init
data_dir: ~/.local/share/judge-loop
```

## Local state

The agent keeps minimal local state in `~/.local/share/judge-loop/`:

```
agent.db         # SQLite (timer sessions, offline queue)
registry/        # cached registry manifests
  index.json
  providers/
    leetcode.json
    neetcode.json
```

## Startup behavior

When the Neovim plugin calls `GET /local/status/today`:

1. Agent checks local DB for today's sessions
2. If server is reachable, also checks `GET /api/progress/today`
3. Returns `practiced: false` if no solved submission today
4. Plugin displays a reminder notification

## Timer flow

```
POST /local/timer/start
  → write TimerSession to local SQLite (started_at, problem_id)
  → attempt to POST /api/timers/start (best-effort, non-blocking)
  → return { ok: true }

POST /local/timer/stop
  → update local TimerSession (ended_at, elapsed)
  → attempt to POST /api/timers/stop
  → return elapsed_seconds
```

Timer continues locally even if server is unreachable. Synced on next successful connection.

## Submit flow

```
POST /local/submit
  → read active timer session
  → attach session_id to submission payload
  → POST /api/submissions
  → return submission_id for polling
```

## Registry sync

```
POST /local/sync
  → GET /api/registry/index.json from server
  → compare version with local cache
  → download changed manifests
  → import new problems into local bank
  → return sync summary
```

## Running the agent

```bash
# First time init
judge-agent init

# Start daemon
judge-agent serve

# Check status
judge-agent status
```

The agent is designed to be started by Neovim on first use (via the plugin) or by the user's shell profile.
