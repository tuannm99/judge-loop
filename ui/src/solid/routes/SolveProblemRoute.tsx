import { useNavigate, useParams } from '@tanstack/solid-router'
import { SolvePage } from '../pages/SolvePage'

export default function SolveProblemRoute() {
  const navigate = useNavigate({ from: '/problems/$slug' })
  const params = useParams({ from: '/problems/$slug' })

  return <SolvePage navigate={(to) => void navigate({ to })} slug={params().slug} />
}
