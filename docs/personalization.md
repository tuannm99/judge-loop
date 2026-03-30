# Personalization Engine

## Purpose

The personalization engine answers:

- What should the user do today?
- What is overdue?
- What is their weakest pattern?
- Are they improving vs their past self?
- Is solve time decreasing?

## Core models

### UserTrainingProfile

Stores the user's training preferences and identified weaknesses.

```go
type UserTrainingProfile struct {
    UserID          uuid.UUID
    Goals           []string          // e.g. "faang-interview", "competitive"
    MinutesPerDay   int               // target daily time
    DifficultyMix   DifficultyMix     // e.g. {Easy: 20, Medium: 60, Hard: 20}
    WeakPatterns    []string          // patterns needing work
    FocusPatterns   []string          // patterns to prioritize
    UpdatedAt       time.Time
}
```

### TrainingContract

Defines daily and weekly targets.

```go
type TrainingContract struct {
    UserID          uuid.UUID
    DailyProblems   int              // required solves per day
    WeeklyProblems  int              // weekly target
    FocusTime       int              // minutes per session
    ReviewEnabled   bool             // spaced repetition on/off
    ActiveFrom      time.Time
}
```

### DailyMission

Generated each day. Contains required and optional tasks.

```go
type DailyMission struct {
    UserID          uuid.UUID
    Date            time.Time
    RequiredTasks   []MissionTask     // must complete
    OptionalTasks   []MissionTask     // bonus
    ReviewTasks     []MissionTask     // spaced repetition
    GeneratedAt     time.Time
}

type MissionTask struct {
    ProblemID   uuid.UUID
    Slug        string
    Reason      string    // e.g. "weak pattern: sliding-window"
    Priority    int
}
```

### PerformanceSnapshot

Periodic snapshot of user performance.

```go
type PerformanceSnapshot struct {
    UserID          uuid.UUID
    SnapshotDate    time.Time
    AvgSolveTime    float64           // minutes
    TotalAttempts   int
    AcceptedCount   int
    HintUsageRate   float64           // 0.0–1.0
    PatternScores   map[string]float64 // pattern → score 0–1
}
```

### ProblemPerformance

Per-problem performance record.

```go
type ProblemPerformance struct {
    UserID          uuid.UUID
    ProblemID       uuid.UUID
    FirstSolveTime  *float64          // minutes, nil if unsolved
    BestSolveTime   *float64
    LatestSolveTime *float64
    Attempts        int
    Accepted        bool
    Complexity      string            // self-reported: O(n), O(n²), etc.
    Confidence      int               // self-reported: 1–5
    LastAttemptAt   time.Time
}
```

Note: `Complexity` is manually entered by the user. The system does not analyze code complexity automatically.

## Daily mission generation

Algorithm (MVP):

1. Load user's `TrainingContract` and `UserTrainingProfile`
2. Find problems not yet solved (from problem bank)
3. Find overdue review problems (spaced repetition schedule)
4. Rank unsolved problems by:
   - Matching weak patterns (high priority)
   - Target difficulty mix
   - Estimated time fitting within `FocusTime`
5. Fill required tasks up to `DailyProblems` count
6. Add optional tasks from next tier
7. Add review tasks from `ReviewSchedule`

## Review scheduling

Spaced repetition intervals (MVP, simple):

- After first solve: review in 1 day
- After second solve: review in 3 days
- After third solve: review in 7 days
- After fourth+: review in 14 days

`ReviewSchedule` stores the next review date per problem per user.

## Performance tracking

After each accepted submission:

1. Update `ProblemPerformance` (solve time, attempt count)
2. If solve time improved → update `BestSolveTime`
3. Recalculate pattern score for affected patterns
4. Check if a `PerformanceSnapshot` is due (weekly)

## Weak pattern detection

Pattern score = accepted / attempted for problems in that pattern.

A pattern is "weak" if score < 0.5 and the user has at least 3 attempts in it.

## Self-comparison

Compare latest `PerformanceSnapshot` vs one from 4 weeks ago:

- If avg solve time decreased → "Improving"
- If accepted rate increased → "Improving"
- Surface this in `GET /api/progress/today`
