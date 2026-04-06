# Architecture

## Overview

`judge-loop` now uses a clear ports-and-adapters structure inside a single Go module.

The main runtime pieces are:

- `api-server` for HTTP reads/writes and queue submission
- `judge-worker` for async evaluation
- `local-agent` for editor-facing local workflows
- PostgreSQL for persistent state
- Redis/asynq for background jobs
- Docker sandbox for code execution

```
Neovim plugin
  -> local-agent
  -> api-server
  -> Redis/asynq
  -> judge-worker
  -> Docker sandbox
  -> PostgreSQL
```

## Diagrams

### System Context

```mermaid
flowchart LR
  Developer[Developer]
  Editor[Neovim Plugin]
  LocalAgent[local-agent]
  APIServer[api-server]
  Queue[Redis / asynq]
  Worker[judge-worker]
  Sandbox[Docker sandbox]
  Postgres[(PostgreSQL)]
  RegistryFiles[Registry files on disk]

  Developer --> Editor
  Editor --> LocalAgent
  LocalAgent --> APIServer
  LocalAgent --> RegistryFiles
  APIServer --> Queue
  APIServer --> Postgres
  Worker --> Queue
  Worker --> Sandbox
  Worker --> Postgres
```

### C4-Style Container View

```mermaid
flowchart TB
  subgraph Client["Person + Client"]
    User[User]
    Plugin[Neovim plugin]
  end

  subgraph Host["Local machine"]
    Agent[local-agent\nHTTP proxy + timer + registry sync]
    Registry[registry/]
  end

  subgraph Platform["judge-loop backend"]
    API[api-server\nHTTP API + submission creation]
    Redis[Redis / asynq\njob queue]
    Worker[judge-worker\nasync evaluator]
    DB[(PostgreSQL)]
    Sandbox[Docker sandbox]
  end

  User --> Plugin
  Plugin --> Agent
  Agent --> API
  Agent --> Registry
  API --> DB
  API --> Redis
  Worker --> Redis
  Worker --> DB
  Worker --> Sandbox
```

### Submission Sequence

```mermaid
sequenceDiagram
  autonumber
  participant Plugin as Neovim plugin
  participant Agent as local-agent
  participant API as api-server
  participant DB as PostgreSQL
  participant Q as Redis/asynq
  participant Worker as judge-worker
  participant Sandbox as Docker sandbox

  Plugin->>Agent: POST /local/submit
  Agent->>API: POST /api/submissions
  API->>DB: insert submission(status=pending)
  API->>Q: enqueue evaluate-submission
  API-->>Agent: 201 {submission_id, status=pending}
  Agent-->>Plugin: submission_id
  Plugin->>Agent: poll /local/submissions/:id
  Agent->>API: GET /api/submissions/:id
  Q-->>Worker: deliver job
  Worker->>DB: load submission + test cases
  Worker->>DB: claim submission(status=running)
  Worker->>Sandbox: run code against test case input
  Sandbox-->>Worker: output / stderr / runtime
  Worker->>DB: update final verdict
  API-->>Agent: submission status
  Agent-->>Plugin: verdict
```

### Fallback Sequence

This is the intended resilient path when no external worker claims the queued job in time.

```mermaid
sequenceDiagram
  autonumber
  participant API as api-server
  participant Q as Redis/asynq
  participant DB as PostgreSQL
  participant Eval as in-process evaluation fallback
  participant Sandbox as Docker sandbox

  API->>DB: insert submission(status=pending)
  API->>Q: enqueue evaluate-submission
  API->>Eval: schedule delayed fallback
  Eval->>DB: try claim submission(status: pending -> running)
  alt claim succeeds
    Eval->>Sandbox: run user code
    Sandbox-->>Eval: execution result
    Eval->>DB: write final verdict
  else already claimed by worker
    Eval-->>API: no-op
  end
```

## Dependency Rule

The package graph is organized so dependencies point inward:

- `cmd/*` wires dependencies and starts processes
- `adapter/*` translates external inputs/outputs
- `application/*` contains use-case orchestration
- `domain/*` contains core domain types and pure logic
- `port/in` exposes application capabilities to adapters
- `port/out` defines infrastructure dependencies required by the application
- `infrastructure/*` provides concrete implementations of outbound ports

Adapters and infrastructure should not own business rules. Entry points should not contain use-case logic.

## Package Layout

Single Go module: `github.com/tuannm99/judge-loop`

| Package                              | Responsibility                                                       |
| ------------------------------------ | -------------------------------------------------------------------- |
| `cmd/api-server`                     | process entrypoint and dependency wiring for the HTTP API            |
| `cmd/judge-worker`                   | process entrypoint and dependency wiring for async evaluation        |
| `cmd/local-agent`                    | process entrypoint and dependency wiring for the local daemon        |
| `internal/domain`                    | domain entities and value types                                      |
| `internal/domain/judge`              | pure verdict evaluation logic                                        |
| `internal/application`               | application use cases, orchestration, and mission-generation helpers |
| `internal/port/in`                   | inbound ports implemented by application services                    |
| `internal/port/out`                  | outbound ports implemented by adapters/infrastructure                |
| `internal/adapter/http`              | Gin handlers for `api-server` and `local-agent`                      |
| `internal/adapter/queue`             | asynq-facing adapters for publish/consume flow                       |
| `internal/adapter/sandbox`           | code runner adapter used by application services                     |
| `internal/infrastructure/postgres`   | GORM repositories and embedded goose migrations                      |
| `internal/infrastructure/queue`      | asynq task definitions and queue setup                               |
| `internal/infrastructure/sandbox`    | Docker-based code execution                                          |
| `internal/infrastructure/registry`   | local registry manifest loading from disk                            |
| `internal/infrastructure/localtimer` | in-memory timer for the local-agent                                  |

## Service Roles

### `api-server`

- exposes the HTTP API on port `8080`
- uses application services behind `port/in`
- persists state through postgres-backed repository adapters
- publishes evaluation jobs through the queue adapter

### `judge-worker`

- consumes evaluation jobs from Redis/asynq
- calls the evaluation application service
- runs code through the sandbox adapter
- writes final verdicts back through repositories

### `local-agent`

- exposes local HTTP endpoints on port `7070`
- keeps an in-memory local timer
- proxies submissions to `api-server`
- loads registry manifests from local disk and pushes them to the server

## Submission Flow

```
Neovim
  -> POST /local/submit
  -> local-agent forwards to POST /api/submissions
  -> api-server creates submission row
  -> api-server enqueues evaluation job
  -> judge-worker consumes job
  -> evaluation service loads submission + test cases
  -> sandbox runs code
  -> domain judge logic computes verdict
  -> submission row updated in PostgreSQL
```

## Low-Level Design

### Low-Level Sequence

```mermaid
sequenceDiagram
  autonumber
  participant H as SubmissionsAPI
  participant S as SubmissionService
  participant Repo as SubmissionRepository
  participant Pub as EvaluationPublisher
  participant Eval as EvaluationService
  participant Cases as TestCaseRepository
  participant Run as CodeRunner

  H->>S: CreateSubmission(userID, problemID, language, code, sessionID)
  S->>Repo: Create(submission)
  S->>Pub: PublishEvaluation(job)
  S->>Eval: schedule fallback after delay
  Eval->>Repo: GetByID(submissionID)
  Eval->>Repo: TryStartEvaluation(submissionID)
  Eval->>Cases: GetByProblem(problemID)
  loop each test case
    Eval->>Run: Run(language, code, input)
    Run-->>Eval: RunResult
  end
  Eval->>Repo: UpdateVerdict(...)
```

### Low-Level Class Diagram

```mermaid
classDiagram
  class SubmissionsAPI {
    +CreateSubmission()
    +GetSubmission()
    +ListSubmissions()
  }

  class SubmissionService {
    -submissions SubmissionRepository
    -publisher EvaluationPublisher
    -evaluator EvaluationService
    +CreateSubmission()
    +GetSubmission()
    +ListSubmissions()
  }

  class EvaluationService {
    -submissions SubmissionRepository
    -testCases TestCaseRepository
    -reviews ReviewRepository
    -sessions SessionRepository
    -runner CodeRunner
    +EvaluateSubmission()
  }

  class SubmissionRepository {
    <<interface>>
    +Create()
    +GetByID()
    +TryStartEvaluation()
    +UpdateVerdict()
    +ListByUser()
  }

  class EvaluationPublisher {
    <<interface>>
    +PublishEvaluation()
  }

  class TestCaseRepository {
    <<interface>>
    +GetByProblem()
  }

  class CodeRunner {
    <<interface>>
    +Run()
  }

  class SubmissionRepositoryImpl
  class EvaluationPublisherAdapter
  class Runner

  SubmissionsAPI --> SubmissionService
  SubmissionService --> SubmissionRepository
  SubmissionService --> EvaluationPublisher
  SubmissionService --> EvaluationService
  EvaluationService --> SubmissionRepository
  EvaluationService --> TestCaseRepository
  EvaluationService --> CodeRunner
  SubmissionRepositoryImpl ..|> SubmissionRepository
  EvaluationPublisherAdapter ..|> EvaluationPublisher
  Runner ..|> CodeRunner
```

## Persistence and Migrations

PostgreSQL is the source of truth for submissions, sessions, reviews, and registry versions.

- ORM: GORM
- migrations: goose
- migration path: `internal/infrastructure/postgres/migrations`
- migrations run automatically on service startup

## Testing

Application-layer tests are built around split ports and generated mocks:

- mock generator: `mockery v3`
- assertion library: `testify`
- mock packages:
  - `internal/port/in/mocks`
  - `internal/port/out/mocks`

## Assumptions

- single-user MVP, no auth boundary yet
- local-agent runs on the developer machine
- problem metadata is stored, not full problem statements
- Docker must be available where `judge-worker` runs
