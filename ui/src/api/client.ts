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
} from './types'

const BASE = '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error((body as { error?: string }).error ?? `HTTP ${res.status}`)
  }
  return res.json() as Promise<T>
}

// Problems
export interface ListProblemsParams {
  difficulty?: Difficulty
  tag?: string
  pattern?: string
  provider?: Provider
  limit?: number
  offset?: number
}

export function listProblems(params: ListProblemsParams = {}) {
  const q = new URLSearchParams()
  if (params.difficulty) q.set('difficulty', params.difficulty)
  if (params.tag) q.set('tag', params.tag)
  if (params.pattern) q.set('pattern', params.pattern)
  if (params.provider) q.set('provider', params.provider)
  if (params.limit) q.set('limit', String(params.limit))
  if (params.offset) q.set('offset', String(params.offset))
  const qs = q.toString()
  return request<{ problems: Problem[]; total: number }>(`/problems${qs ? `?${qs}` : ''}`)
}

export function getProblem(idOrSlug: string) {
  return request<Problem>(`/problems/${idOrSlug}`)
}

export function suggestProblem() {
  return request<Problem>('/problems/suggest')
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
    body: JSON.stringify(payload),
  })
}

export function getSubmission(id: string) {
  return request<Submission>(`/submissions/${id}`)
}

export function listSubmissions(problemId?: string) {
  const q = problemId ? `?problem_id=${problemId}` : ''
  return request<{ submissions: Submission[] }>(`/submissions/history${q}`)
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
  return request<{ reviews: ReviewItem[] }>('/reviews/today')
}

// Timers
export function startTimer(problemId?: string) {
  return request<{ id: string; started_at: string; problem_id?: string }>('/timers/start', {
    method: 'POST',
    body: JSON.stringify(problemId ? { problem_id: problemId } : {}),
  })
}

export function stopTimer() {
  return request<{ elapsed_seconds: number; active?: boolean }>('/timers/stop', {
    method: 'POST',
    body: JSON.stringify({}),
  })
}

export function currentTimer() {
  return request<Timer>('/timers/current')
}
