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
4. **Registry sync** — Load local registry manifests and push them to the api-server
5. **Offline support** — Timer works even when server is unreachable

## Configuration

Config file: `~/.config/judge-loop/agent.yaml`

```yaml
server_url: http://localhost:8080
port: 7070
user_id: uuid
data_dir: ~/.local/share/judge-loop
registry_path: ./registry
```

## Local state

The current agent keeps only in-memory timer state plus configured access to a local registry directory:

```
registry/        # local manifest source used by POST /local/sync
  index.json
  providers/
    leetcode/
      free/problems.json
      premium/problems.json
    neetcode.json
```

Timer state is in memory in the current MVP and is lost on agent restart.

## Startup behavior

When the Neovim plugin calls `GET /local/status/today`:

1. Agent checks its in-memory local timer state
2. If server is reachable, also checks `GET /api/progress/today`
3. Returns `practiced: false` if no solved submission today
4. Plugin displays a reminder notification

## Timer flow

```
POST /local/timer/start
  → create in-memory TimerSession (started_at, problem_id)
  → attempt to POST /api/timers/start (best-effort, non-blocking)
  → return { ok: true }

POST /local/timer/stop
  → stop in-memory TimerSession and compute elapsed locally
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
  → read local registry/index.json
  → load provider manifests from local registry_path
  → POST loaded manifests to /api/registry/sync
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

## Architecture note

In the current code layout:

- HTTP handlers live under `internal/adapter/http/localagent`
- local timer implementation lives under `internal/infrastructure/localtimer`
- registry manifest loading lives under `internal/infrastructure/registry`

`cmd/local-agent` is only the entrypoint and wiring layer.
