import type {
  Problem,
  Submission,
  ProgressToday,
  Streak,
  ReviewItem,
  Timer,
  Language,
  Difficulty,
  Provider,
  ProblemLabel
} from './types'

const BASE = '/api'

function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : []
}

function normalizeProblem(problem: Problem): Problem {
  return {
    ...problem,
    tags: asArray(problem.tags),
    pattern_tags: asArray(problem.pattern_tags),
    starter_code: problem.starter_code ?? {}
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error((body as { error?: string }).error ?? `HTTP ${res.status}`)
  }
  if (res.status === 204) {
    return undefined as T
  }
  return res.json() as Promise<T>
}

// Problems
export interface ListProblemsParams {
  difficulty?: Difficulty
  tags?: string[]
  patterns?: string[]
  provider?: Provider
  limit?: number
  offset?: number
}

export function listProblems(params: ListProblemsParams = {}) {
  const q = new URLSearchParams()
  if (params.difficulty) q.set('difficulty', params.difficulty)
  for (const tag of params.tags ?? []) q.append('tag', tag)
  for (const pattern of params.patterns ?? []) q.append('pattern', pattern)
  if (params.provider) q.set('provider', params.provider)
  if (params.limit) q.set('limit', String(params.limit))
  if (params.offset) q.set('offset', String(params.offset))
  const qs = q.toString()
  return request<{ problems: Problem[] | null; total: number }>(
    `/problems${qs ? `?${qs}` : ''}`
  ).then((result) => ({
    total: result.total,
    problems: asArray(result.problems).map(normalizeProblem)
  }))
}

export function getProblem(idOrSlug: string) {
  return request<Problem>(`/problems/${idOrSlug}`).then(normalizeProblem)
}

export function listProblemLabels() {
  return request<{ tags: string[] | null; patterns: string[] | null }>(
    '/problems/labels'
  ).then((result) => ({
    tags: asArray(result.tags),
    patterns: asArray(result.patterns)
  }))
}

export function listProblemLabelRecords(kind: 'tag' | 'pattern') {
  return request<{ labels: ProblemLabel[] | null }>(
    `/problem-labels?kind=${kind}`
  ).then((result) => ({
    labels: asArray(result.labels)
  }))
}

export function createProblemLabel(payload: {
  kind: 'tag' | 'pattern'
  slug: string
  name?: string
}) {
  return request<ProblemLabel>('/problem-labels', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export function updateProblemLabel(
  id: string,
  payload: { slug: string; name?: string }
) {
  return request<ProblemLabel>(`/problem-labels/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload)
  })
}

export function deleteProblemLabel(id: string) {
  return request<void>(`/problem-labels/${id}`, {
    method: 'DELETE'
  })
}

export function suggestProblem() {
  return request<Problem>('/problems/suggest').then(normalizeProblem)
}

export interface ContributeProblemPayload {
  provider: Provider
  external_id: string
  slug: string
  title: string
  difficulty: Difficulty
  tags: string[]
  pattern_tags: string[]
  source_url: string
  estimated_time: number
  starter_code: Partial<Record<Language, string>>
  version: number
  test_cases: Array<{
    input: string
    expected: string
    is_hidden?: boolean
  }>
}

export function contributeProblem(payload: ContributeProblemPayload) {
  return request<Problem>('/problems/contribute', {
    method: 'POST',
    body: JSON.stringify(payload)
  }).then(normalizeProblem)
}

export function updateProblem(
  id: string,
  payload: Omit<ContributeProblemPayload, 'version'>
) {
  return request<Problem>(`/problems/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload)
  }).then(normalizeProblem)
}

export function getProblemTestCases(id: string) {
  return request<{
    test_cases:
      | Array<{
          input: string
          expected: string
          is_hidden?: boolean
        }>
      | null
  }>(`/problems/${id}/test-cases`).then((result) => ({
    test_cases: asArray(result.test_cases)
  }))
}

// Submissions
export function createSubmission(payload: {
  problem_id: string
  language: Language
  code: string
  session_id?: string
}) {
  return request<{ submission_id: string; status: string }>('/submissions', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export function getSubmission(id: string) {
  return request<Submission>(`/submissions/${id}`)
}

export function listSubmissions(problemId?: string) {
  const q = problemId ? `?problem_id=${problemId}` : ''
  return request<{ submissions: Submission[] | null }>(
    `/submissions/history${q}`
  ).then((result) => ({
    submissions: asArray(result.submissions)
  }))
}

// Progress
export function getProgressToday() {
  return request<ProgressToday>('/progress/today')
}

export function getStreak() {
  return request<Streak>('/streak')
}

// Reviews
export function getReviewsToday() {
  return request<{ reviews: ReviewItem[] | null }>('/reviews/today').then(
    (result) => ({
      reviews: asArray(result.reviews)
    })
  )
}

// Timers
export function startTimer(problemId?: string) {
  return request<{ id: string; started_at: string; problem_id?: string }>(
    '/timers/start',
    {
      method: 'POST',
      body: JSON.stringify(problemId ? { problem_id: problemId } : {})
    }
  )
}

export function stopTimer() {
  return request<{ elapsed_seconds: number; active?: boolean }>('/timers/stop', {
    method: 'POST',
    body: JSON.stringify({})
  })
}

export function currentTimer() {
  return request<Timer>('/timers/current')
}
