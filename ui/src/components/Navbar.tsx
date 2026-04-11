import { Link, useLocation } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getProgressToday } from '@/api/client'
import { Group, Text, Anchor } from '@mantine/core'
import { IconCode, IconFlame } from '@tabler/icons-react'

const links = [
  { to: '/', label: 'Problems' },
  { to: '/problems/contribute', label: 'New Problem' },
  { to: '/problem-labels', label: 'Labels' },
  { to: '/dashboard', label: 'Dashboard' }
]

export function Navbar() {
  const { pathname } = useLocation()
  const qc = useQueryClient()

  const { data } = useQuery({
    queryKey: ['progress-today'],
    queryFn: getProgressToday,
    refetchInterval: 60_000
  })

  return (
    <header
      style={{
        borderBottom: '1px solid var(--mantine-color-dark-4)',
        padding: '10px 24px'
      }}
    >
      <Group justify="space-between">
        <Group gap="lg">
          <Group gap="xs">
            <IconCode size={18} color="var(--mantine-color-teal-5)" />
            <Text fw={700} c="teal.4">
              judge-loop
            </Text>
          </Group>

          {links.map((l) => (
            <Anchor
              key={l.to}
              component={Link}
              to={l.to}
              onClick={() => {
                void qc.invalidateQueries()
              }}
              c={pathname === l.to ? 'white' : 'dimmed'}
              size="sm"
              underline="never"
            >
              {l.label}
            </Anchor>
          ))}
        </Group>

        {data && (
          <Group gap="md">
            <Text size="sm" c="dimmed">
              <Text span c="white" fw={500}>
                {data.solved}
              </Text>{' '}
              solved today
            </Text>
            <Group gap={4}>
              <IconFlame size={14} color="var(--mantine-color-orange-4)" />
              <Text size="sm" c="orange.4" fw={500}>
                {data.streak}
              </Text>
            </Group>
          </Group>
        )}
      </Group>
    </header>
  )
}
