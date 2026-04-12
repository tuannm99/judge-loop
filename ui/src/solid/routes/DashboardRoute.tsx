import { useNavigate } from '@tanstack/solid-router'
import { DashboardPage } from '../pages/DashboardPage'

export default function DashboardRoute() {
  const navigate = useNavigate({ from: '/dashboard' })

  return <DashboardPage navigate={(to) => void navigate({ to })} />
}
