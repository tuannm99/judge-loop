import type { Difficulty, SubmissionStatus } from '@/api/types'
import { difficultyBadgeType, statusBadgeType } from '../../shared/utils'
import { Badge } from '../common/Primitives'

export function DifficultyBadge(props: { difficulty: Difficulty }) {
  return <Badge content={props.difficulty} color={difficultyBadgeType(props.difficulty)} />
}

export function StatusBadge(props: { status: SubmissionStatus; label?: string }) {
  return (
    <Badge
      content={props.label ?? props.status.replace(/_/g, ' ')}
      color={statusBadgeType(props.status)}
    />
  )
}
