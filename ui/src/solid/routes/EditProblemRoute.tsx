import { useNavigate, useParams } from '@tanstack/solid-router'
import { ContributeProblemPage } from '../pages/ContributeProblemPage'

export default function EditProblemRoute() {
  const navigate = useNavigate({ from: '/problems/$slug/edit' })
  const params = useParams({ from: '/problems/$slug/edit' })

  return <ContributeProblemPage navigate={(to) => void navigate({ to })} slug={params().slug} />
}
