import { Badge as MantineBadge } from '@mantine/core'
import type { Difficulty, SubmissionStatus } from '@/api/types'

const DIFFICULTY_COLOR: Record<Difficulty, string> = {
  easy: 'teal',
  medium: 'yellow',
  hard: 'red'
}

const STATUS_COLOR: Record<SubmissionStatus, string> = {
  pending: 'gray',
  running: 'blue',
  accepted: 'teal',
  wrong_answer: 'red',
  compile_error: 'orange',
  runtime_error: 'orange',
  time_limit_exceeded: 'orange'
}

export { MantineBadge as Badge }

export function DifficultyBadge({ difficulty }: { difficulty: Difficulty }) {
  return (
    <MantineBadge
      color={DIFFICULTY_COLOR[difficulty]}
      variant="light"
      size="sm"
    >
      {difficulty}
    </MantineBadge>
  )
}

export function StatusBadge({
  status,
  label
}: {
  status: SubmissionStatus
  label?: string
}) {
  return (
    <MantineBadge color={STATUS_COLOR[status]} variant="light" size="sm">
      {label ?? status.replace(/_/g, ' ')}
    </MantineBadge>
  )
}
