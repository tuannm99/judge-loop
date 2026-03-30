# API Reference

## api-server (port 8080)

Base URL: `http://localhost:8080`

---

### Problems

#### `GET /api/problems`

List problems with optional filters.

Query params:

- `difficulty` — easy | medium | hard
- `tag` — algorithm tag (e.g. `array`, `dp`)
- `pattern` — pattern tag (e.g. `sliding-window`)
- `provider` — leetcode | neetcode | hackerrank
- `limit` — default 20
- `offset` — default 0

Response:

```json
{
  "problems": [
    {
      "id": "uuid",
      "slug": "two-sum",
      "title": "Two Sum",
      "difficulty": "easy",
      "tags": ["array", "hash-table"],
      "pattern_tags": ["lookup"],
      "provider": "leetcode",
      "source_url": "https://leetcode.com/problems/two-sum",
      "estimated_time": 15
    }
  ],
  "total": 100
}
```

#### `GET /api/problems/:id`

Get a single problem by UUID or slug.

#### `GET /api/problems/suggest`

Get a suggested problem based on user profile.

Response: single problem object.

---

### Submissions

#### `POST /api/submissions`

Submit code for evaluation.

Body:

```json
{
  "problem_id": "uuid",
  "language": "python",
  "code": "def twoSum(...):\n  ...",
  "session_id": "uuid (optional)"
}
```

Response:

```json
{
  "submission_id": "uuid",
  "status": "pending"
}
```

#### `GET /api/submissions/:id`

Poll submission status.

Response:

```json
{
  "id": "uuid",
  "status": "accepted",
  "verdict": "Accepted",
  "passed_cases": 10,
  "total_cases": 10,
  "runtime_ms": 42,
  "error_message": null,
  "submitted_at": "2026-01-01T10:00:00Z"
}
```

Status values: `pending` | `running` | `accepted` | `wrong_answer` | `compile_error` | `runtime_error` | `time_limit_exceeded`

#### `GET /api/submissions/history`

Get submission history.

Query params:

- `problem_id` — filter by problem
- `limit` — default 20

---

### Progress

#### `GET /api/progress/today`

Get today's practice summary.

Response:

```json
{
  "date": "2026-01-01",
  "solved": 2,
  "attempted": 3,
  "time_spent_minutes": 45,
  "streak": 7
}
```

#### `GET /api/streak`

Get current streak info.

Response:

```json
{
  "current": 7,
  "longest": 14,
  "last_practiced": "2026-01-01"
}
```

---

### Timers

#### `POST /api/timers/start`

Start a timer session.

Body:

```json
{
  "problem_id": "uuid (optional)"
}
```

#### `POST /api/timers/stop`

Stop current timer. Body: `{}`.

#### `GET /api/timers/current`

Get active timer.

Response:

```json
{
  "active": true,
  "started_at": "2026-01-01T10:00:00Z",
  "elapsed_seconds": 300,
  "problem_id": "uuid or null"
}
```

---

### Reviews

#### `GET /api/reviews/today`

Get problems due for spaced repetition review today.

Response:

```json
{
  "reviews": [
    {
      "problem_id": "uuid",
      "slug": "two-sum",
      "last_solved": "2025-12-25",
      "days_overdue": 2
    }
  ]
}
```

---

## local-agent (port 7070)

Base URL: `http://localhost:7070`

All endpoints are localhost-only.

---

#### `GET /local/status/today`

Check if user has practiced today.

Response:

```json
{
  "practiced": false,
  "solved_count": 0,
  "active_timer": false,
  "message": "No practice yet today. Start a session!"
}
```

#### `GET /local/timer/current`

Get current timer state.

#### `POST /local/timer/start`

Start a timer. Body: `{ "problem_id": "uuid or empty" }`.

#### `POST /local/timer/stop`

Stop active timer. Body: `{}`.

#### `POST /local/submit`

Proxy submission to api-server.

Same body as `POST /api/submissions`.

Returns same response as api-server.

#### `POST /local/sync`

Trigger registry sync from server.

Response:

```json
{
  "synced": true,
  "problems_imported": 12,
  "registry_version": "1.0.3"
}
```
