import { useNavigate } from '@tanstack/solid-router'
import { ProblemListPage } from '../pages/ProblemListPage'

export default function ProblemsRoute() {
  const navigate = useNavigate({ from: '/' })

  return <ProblemListPage navigate={(to) => void navigate({ to })} />
}
