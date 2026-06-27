import { useNavigate } from '@tanstack/solid-router'
import { ContributeProblemPage } from '../pages/ContributeProblemPage'

export default function ContributeProblemRoute() {
  const navigate = useNavigate({ from: '/problems/new' })

  return <ContributeProblemPage navigate={(to) => void navigate({ to })} />
}
