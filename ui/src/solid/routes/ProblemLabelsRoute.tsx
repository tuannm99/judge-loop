import { useNavigate } from '@tanstack/solid-router'
import { ProblemLabelsPage } from '../pages/ProblemLabelsPage'

export default function ProblemLabelsRoute() {
  const navigate = useNavigate({ from: '/problem-labels' })

  return <ProblemLabelsPage navigate={(to) => void navigate({ to })} />
}
