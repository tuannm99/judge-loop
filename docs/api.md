# API Specs

This spec is based on the current implementation in `internal/adapter/http/apiserver` and `internal/adapter/http/localagent`.

Purpose:

- manual testing with `curl` or Postman
- input for API test cases
- reflect the code as it exists now, not the older roadmap/docs

## Test Setup

### api-server

- Base URL: `http://localhost:8080`
- Health: `GET /health`

### local-agent

- Base URL: `http://localhost:7070`
- Health: `GET /health`

### Notes

- There is no runtime auth yet; `api-server` uses `USER_ID` from server config.
- All JSON bodies should use `Content-Type: application/json`.
- Validation and parse errors usually return:

```json
{
  "error": "..."
}
```

- Internal server errors usually return HTTP `500`.
- Proxy failures from `local-agent` to `api-server` usually return HTTP `502`.

## api-server

### `GET /health`

Response `200`:

```json
{
  "status": "ok"
}
```

### `GET /api/problems`

List problems.

Query params:

- `difficulty`: `easy` | `medium` | `hard`
- `tag`: string
- `pattern`: string
- `provider`: `leetcode` | `neetcode` | `hackerrank`
- `limit`: integer > 0
- `offset`: integer >= 0

Response `200`:

```json
{
  "problems": [
    {
      "id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e",
      "slug": "two-sum",
      "title": "Two Sum",
      "difficulty": "easy",
      "tags": ["array", "hash-table"],
      "pattern_tags": ["lookup"],
      "provider": "leetcode",
      "external_id": "1",
      "source_url": "https://leetcode.com/problems/two-sum",
      "estimated_time": 15,
      "starter_code": {
        "python": "class Solution:\n    pass\n",
        "go": "package main\n\nfunc main() {}\n"
      }
    }
  ],
  "total": 1
}
```

Example:

```bash
curl "http://localhost:8080/api/problems?difficulty=easy&limit=10&offset=0"
```

### `GET /api/problems/:id`

Get a problem by UUID or slug.

Response `200`:

```json
{
  "id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e",
  "slug": "two-sum",
  "title": "Two Sum",
  "difficulty": "easy",
  "tags": ["array", "hash-table"],
  "pattern_tags": ["lookup"],
  "provider": "leetcode",
  "external_id": "1",
  "source_url": "https://leetcode.com/problems/two-sum",
  "estimated_time": 15,
  "starter_code": {
    "python": "class Solution:\n    pass\n",
    "go": "package main\n\nfunc main() {}\n"
  }
}
```

Response `404`:

```json
{
  "error": "problem not found"
}
```

### `GET /api/problems/suggest`

Get a suggested problem for the current user.

Response `200`: same shape as `GET /api/problems/:id`

Response `404`:

```json
{
  "error": "no unsolved problems available"
}
```

### `POST /api/problems/contribute`

Add or update one problem in the question bank.

Request body:

```json
{
  "provider": "leetcode",
  "external_id": "1",
  "slug": "two-sum",
  "title": "Two Sum",
  "difficulty": "easy",
  "tags": ["array", "hash-table"],
  "pattern_tags": ["lookup"],
  "source_url": "https://leetcode.com/problems/two-sum",
  "estimated_time": 15,
  "starter_code": {
    "python": "class Solution:\n    pass\n",
    "go": "package main\n\nfunc main() {}\n"
  },
  "version": 1,
  "test_cases": [
    {
      "input": "2 7\n9",
      "expected": "0 1"
    },
    {
      "input": "3 2 4\n6",
      "expected": "1 2",
      "is_hidden": true
    }
  ]
}
```

Response `201`:

```json
{
  "id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e",
  "slug": "two-sum",
  "title": "Two Sum",
  "difficulty": "easy",
  "tags": ["array", "hash-table"],
  "pattern_tags": ["lookup"],
  "provider": "leetcode",
  "external_id": "1",
  "source_url": "https://leetcode.com/problems/two-sum",
  "estimated_time": 15,
  "starter_code": {
    "python": "class Solution:\n    pass\n",
    "go": "package main\n\nfunc main() {}\n"
  }
}
```

Example:

```bash
curl -X POST http://localhost:8080/api/problems/contribute \
  -H 'Content-Type: application/json' \
  -d '{
    "provider": "leetcode",
    "external_id": "1",
    "slug": "two-sum",
    "title": "Two Sum",
    "difficulty": "easy",
    "tags": ["array", "hash-table"],
    "pattern_tags": ["lookup"],
    "source_url": "https://leetcode.com/problems/two-sum",
    "estimated_time": 15,
    "starter_code": {
      "python": "class Solution:\n    pass\n",
      "go": "package main\n\nfunc main() {}\n"
    },
    "version": 1,
    "test_cases": [
      {
        "input": "2 7\n9",
        "expected": "0 1"
      },
      {
        "input": "3 2 4\n6",
        "expected": "1 2",
        "is_hidden": true
      }
    ]
  }'
```

### `POST /api/submissions`

Create a submission and enqueue evaluation on a best-effort basis.

Request body:

```json
{
  "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e",
  "language": "python",
  "code": "print(1)",
  "session_id": "2b6300da-739d-416a-9f60-aaf6ca4b3859"
}
```

`session_id` is optional.

Response `201`:

```json
{
  "submission_id": "59e4673e-b9af-4393-8465-433b63f89f74",
  "status": "pending"
}
```

Response `400`:

```json
{
  "error": "invalid problem_id"
}
```

Or a JSON bind/validation error:

```json
{
  "error": "Key: 'createSubmissionRequest.Code' Error:Field validation for 'Code' failed on the 'required' tag"
}
```

### `GET /api/submissions/:id`

Get the current submission state.

Response `200`:

```json
{
  "id": "59e4673e-b9af-4393-8465-433b63f89f74",
  "user_id": "d290f1ee-6c54-4b01-90e6-d701748f0851",
  "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e",
  "session_id": "2b6300da-739d-416a-9f60-aaf6ca4b3859",
  "language": "python",
  "code": "print(1)",
  "status": "accepted",
  "verdict": "Accepted",
  "passed_cases": 2,
  "total_cases": 2,
  "runtime_ms": 10,
  "error_message": "",
  "submitted_at": "2026-04-03T10:00:00Z",
  "evaluated_at": "2026-04-03T10:00:03Z"
}
```

Status values:

- `pending`
- `running`
- `accepted`
- `wrong_answer`
- `compile_error`
- `runtime_error`
- `time_limit_exceeded`

Response `400`:

```json
{
  "error": "invalid submission id"
}
```

Response `404`:

```json
{
  "error": "submission not found"
}
```

### `GET /api/submissions/history`

List submissions for the current user.

Query params:

- `problem_id`: optional UUID filter

Response `200`:

```json
{
  "submissions": [
    {
      "id": "59e4673e-b9af-4393-8465-433b63f89f74",
      "user_id": "d290f1ee-6c54-4b01-90e6-d701748f0851",
      "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e",
      "language": "python",
      "code": "print(1)",
      "status": "accepted",
      "verdict": "Accepted",
      "passed_cases": 2,
      "total_cases": 2,
      "runtime_ms": 10,
      "error_message": "",
      "submitted_at": "2026-04-03T10:00:00Z",
      "evaluated_at": "2026-04-03T10:00:03Z"
    }
  ]
}
```

Note: the current handler uses fixed `limit=20` and `offset=0`.

### `GET /api/progress/today`

Response `200`:

```json
{
  "date": "2026-04-03",
  "solved": 1,
  "attempted": 2,
  "time_spent_minutes": 25,
  "streak": 3
}
```

### `GET /api/streak`

Response `200`:

```json
{
  "current": 3,
  "longest": 5,
  "last_practiced": "2026-04-03T00:00:00Z"
}
```

### `POST /api/timers/start`

Body is optional.

Request body:

```json
{
  "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e"
}
```

Response `200`:

```json
{
  "id": "2b6300da-739d-416a-9f60-aaf6ca4b3859",
  "started_at": "2026-04-03T10:00:00Z",
  "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e"
}
```

If `problem_id` cannot be parsed, the current server ignores it and still starts the timer with `problem_id = null`.

### `POST /api/timers/stop`

Body is not required.

Response `200` when a timer exists:

```json
{
  "elapsed_seconds": 320
}
```

Response `200` when no timer exists:

```json
{
  "active": false,
  "elapsed_seconds": 0
}
```

### `GET /api/timers/current`

Response `200` when a timer exists:

```json
{
  "active": true,
  "id": "2b6300da-739d-416a-9f60-aaf6ca4b3859",
  "started_at": "2026-04-03T10:00:00Z",
  "elapsed_seconds": 320,
  "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e"
}
```

Response `200` when no timer exists:

```json
{
  "active": false
}
```

### `GET /api/reviews/today`

Response `200`:

```json
{
  "reviews": [
    {
      "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e",
      "slug": "two-sum",
      "title": "Two Sum",
      "difficulty": "easy",
      "days_overdue": 0
    }
  ]
}
```

### `POST /api/registry/sync`

Upsert registry version and problem manifests.

Request body:

```json
{
  "version": "2026-04-03",
  "updated_at": "2026-04-03T08:00:00Z",
  "problems": [
    {
      "slug": "two-sum",
      "title": "Two Sum",
      "difficulty": "easy",
      "provider": "leetcode",
      "source_url": "https://leetcode.com/problems/two-sum",
      "tags": ["array", "hash-table"],
      "pattern_tags": ["lookup"],
      "estimated_time": 15
    }
  ],
  "manifests": [
    {
      "kind": "provider",
      "name": "leetcode",
      "path": "providers/leetcode.json"
    }
  ]
}
```

Response `200`:

```json
{
  "version": "2026-04-03",
  "synced": 1
}
```

Response `400` if `version` or `problems` is missing.

### `GET /api/registry/version`

Response `200` when no sync has happened yet:

```json
{
  "version": "none",
  "synced_at": null
}
```

Response `200` after sync:

```json
{
  "version": "2026-04-03",
  "synced_at": "2026-04-03T08:00:00Z"
}
```

## local-agent

### `GET /health`

Response `200`:

```json
{
  "status": "ok"
}
```

### `GET /local/status/today`

If `api-server` is reachable:

Response `200`:

```json
{
  "practiced": true,
  "solved_count": 1,
  "active_timer": false,
  "streak": 3,
  "message": "Good work today! Come back tomorrow."
}
```

If `api-server` is unreachable:

Response `200`:

```json
{
  "practiced": false,
  "solved_count": 0,
  "active_timer": false,
  "message": "No practice yet today. Start a session!",
  "server_error": "Get \"http://localhost:8080/api/progress/today\": dial tcp ..."
}
```

### `GET /local/timer/current`

If a local timer exists:

```json
{
  "active": true,
  "id": "2b6300da-739d-416a-9f60-aaf6ca4b3859",
  "started_at": "2026-04-03T10:00:00Z",
  "elapsed_seconds": 320,
  "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e"
}
```

If no local timer exists and the server is unreachable:

```json
{
  "active": false
}
```

If no local timer exists but the server is reachable, the response is proxied from `GET /api/timers/current`.

### `POST /local/timer/start`

Body is optional:

```json
{
  "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e"
}
```

Response `200`:

```json
{
  "id": "2b6300da-739d-416a-9f60-aaf6ca4b3859",
  "started_at": "2026-04-03T10:00:00Z",
  "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e"
}
```

Note: `local-agent` starts the local timer first, then syncs to `api-server` on a best-effort background call.

### `POST /local/timer/stop`

Response `200` when a timer exists:

```json
{
  "elapsed_seconds": 320
}
```

Response `200` when no timer exists:

```json
{
  "active": false,
  "elapsed_seconds": 0
}
```

### `POST /local/submit`

Proxy submission creation to `api-server`.

Request body:

```json
{
  "problem_id": "6b0ef3c5-d73d-4064-a992-f34bb0e18f8e",
  "language": "python",
  "code": "print(1)"
}
```

If a local timer is active, `local-agent` automatically attaches `session_id`.

Response `201`:

```json
{
  "submission_id": "59e4673e-b9af-4393-8465-433b63f89f74",
  "status": "pending"
}
```

Response `400`: body validation error.

Response `502`:

```json
{
  "error": "api-server unreachable: ..."
}
```

### `GET /local/submissions/:id`

Proxy status from `GET /api/submissions/:id`.

Response `200`:

```json
{
  "id": "59e4673e-b9af-4393-8465-433b63f89f74",
  "status": "accepted",
  "verdict": "Accepted",
  "passed_cases": 2,
  "total_cases": 2,
  "runtime_ms": 10,
  "error_message": ""
}
```

Response `502`:

```json
{
  "error": "api-server GET /api/submissions/:id: 404 Not Found"
}
```

### `POST /local/sync`

Load the local registry from disk, then push it to `api-server`.

Response `200`:

```json
{
  "synced": true,
  "version": "2026-04-03",
  "problems": 12,
  "message": "Registry synced: 12 problems (version 2026-04-03)"
}
```

Response `503` when `registry_path` is not configured:

```json
{
  "synced": false,
  "message": "registry_path not configured"
}
```

Response `500` when loading the local index/manifests fails.

Response `502` when sync to `api-server` fails.

## Suggested Smoke Test Order

```bash
curl http://localhost:8080/health
curl http://localhost:7070/health
curl http://localhost:8080/api/problems
curl -X POST http://localhost:7070/local/timer/start -H 'Content-Type: application/json' -d '{}'
curl -X POST http://localhost:7070/local/submit -H 'Content-Type: application/json' -d '{"problem_id":"<uuid>","language":"python","code":"print(1)"}'
curl http://localhost:7070/local/submissions/<submission_id>
curl -X POST http://localhost:7070/local/timer/stop -H 'Content-Type: application/json' -d '{}'
```
