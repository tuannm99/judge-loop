import type { Difficulty, Language, Problem, SubmissionStatus } from '@/api/types'
import { DEFAULT_CODE, DIFFICULTY_ORDER } from './constants'
import type { SortMode } from './types'

export function classes(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(' ')
}

export function formatError(error: unknown) {
  if (error instanceof Error && error.message) return error.message
  return 'Request failed'
}

export function formatElapsed(totalSeconds: number) {
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return `${minutes}:${String(seconds).padStart(2, '0')}`
}

export function formatDate(value: string | undefined) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString()
}

export function formatDateTime(value: string | undefined) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleString()
}

export function resolveStarterCode(
  starterCode: Partial<Record<Language, string>> | undefined,
  fallback: Record<Language, string> = DEFAULT_CODE
) {
  return {
    python: starterCode?.python || fallback.python,
    go: starterCode?.go || fallback.go
  }
}

export function toggleValue(values: string[], value: string) {
  return values.includes(value) ? values.filter((item) => item !== value) : [...values, value]
}

export function isPending(status: SubmissionStatus) {
  return status === 'pending' || status === 'running'
}

export function sortProblems(problems: Problem[], mode: SortMode) {
  const items = [...problems]

  if (mode === 'title') {
    items.sort((a, b) => a.title.localeCompare(b.title))
  } else if (mode === 'difficulty') {
    items.sort((a, b) => DIFFICULTY_ORDER[a.difficulty] - DIFFICULTY_ORDER[b.difficulty])
  } else if (mode === 'time-desc') {
    items.sort((a, b) => b.estimated_time - a.estimated_time)
  } else if (mode === 'provider') {
    items.sort((a, b) => a.provider.localeCompare(b.provider))
  }

  return items
}

export function navKey(pathname: string) {
  if (pathname === '/' || pathname.startsWith('/problems/')) {
    return 'problems'
  }
  if (pathname === '/problem-labels') return 'labels'
  if (pathname === '/dashboard') return 'dashboard'
  return ''
}

export function difficultyBadgeType(difficulty: Difficulty) {
  if (difficulty === 'easy') return 'green'
  if (difficulty === 'medium') return 'yellow'
  return 'pink'
}

export function statusBadgeType(status: SubmissionStatus) {
  if (status === 'accepted') return 'green'
  if (status === 'pending' || status === 'running') return 'blue'
  return 'pink'
}
