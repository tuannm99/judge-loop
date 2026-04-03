import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { getProgressToday, getStreak, getReviewsToday } from '@/api/client'
import { DifficultyBadge } from '@/components/ui'
import {
  Card, Text, Group, Stack, SimpleGrid, Center, Loader, Paper, Badge,
} from '@mantine/core'
import {
  IconFlame, IconCalendarCheck, IconClock, IconTarget, IconRefresh,
} from '@tabler/icons-react'

function StatCard({
  label, value, icon,
}: {
  label: string
  value: string | number
  icon: React.ReactNode
}) {
  return (
    <Card withBorder>
      <Group justify="space-between" mb="xs">
        <Text size="xs" tt="uppercase" fw={600} c="dimmed">{label}</Text>
        {icon}
      </Group>
      <Text size="2rem" fw={700}>{value}</Text>
    </Card>
  )
}

export function Dashboard() {
  const navigate = useNavigate()

  const { data: progress, isLoading } = useQuery({
    queryKey: ['progress-today'],
    queryFn: getProgressToday,
    refetchInterval: 30_000,
  })
  const { data: streak } = useQuery({ queryKey: ['streak'], queryFn: getStreak })
  const { data: reviews } = useQuery({ queryKey: ['reviews-today'], queryFn: getReviewsToday })

  if (isLoading) return <Center h={200}><Loader /></Center>

  return (
    <Stack p="lg" gap="xl" maw={900}>
      <div>
        <Text size="xl" fw={600}>Dashboard</Text>
        <Text size="sm" c="dimmed">{progress?.date ?? '—'}</Text>
      </div>

      <SimpleGrid cols={{ base: 2, md: 4 }}>
        <StatCard label="Solved today" value={progress?.solved ?? 0} icon={<IconTarget size={16} />} />
        <StatCard label="Attempted" value={progress?.attempted ?? 0} icon={<IconCalendarCheck size={16} />} />
        <StatCard label="Time today" value={`${progress?.time_spent_minutes ?? 0}m`} icon={<IconClock size={16} />} />
        <StatCard label="Streak" value={streak?.current ?? 0} icon={<IconFlame size={16} color="var(--mantine-color-orange-4)" />} />
      </SimpleGrid>

      {streak && (
        <Card withBorder>
          <Text size="sm" fw={600} c="dimmed" mb="md">Streak</Text>
          <Group gap="xl">
            <div>
              <Group gap="xs">
                <IconFlame size={18} color="var(--mantine-color-orange-4)" />
                <Text size="xl" fw={700}>{streak.current}</Text>
              </Group>
              <Text size="xs" c="dimmed">current</Text>
            </div>
            <div>
              <Text size="xl" fw={700} c="dimmed">{streak.longest}</Text>
              <Text size="xs" c="dimmed">longest</Text>
            </div>
            <Text size="xs" c="dimmed">
              Last practiced: {streak.last_practiced
                ? new Date(streak.last_practiced).toLocaleDateString()
                : '—'}
            </Text>
          </Group>
        </Card>
      )}

      <div>
        <Group gap="xs" mb="sm">
          <IconRefresh size={14} />
          <Text size="sm" fw={500}>Due for review today</Text>
          {reviews?.reviews && <Badge size="sm" variant="default">{reviews.reviews.length}</Badge>}
        </Group>

        {reviews?.reviews?.length === 0 && (
          <Text size="sm" c="dimmed">No reviews due — you're all caught up.</Text>
        )}

        <Stack gap="xs">
          {reviews?.reviews?.map((r) => (
            <Paper
              key={r.problem_id}
              withBorder
              p="sm"
              style={{ cursor: 'pointer' }}
              onClick={() => navigate(`/problems/${r.slug}`)}
            >
              <Group justify="space-between">
                <Group gap="sm">
                  <Text size="sm">{r.title}</Text>
                  <DifficultyBadge difficulty={r.difficulty} />
                </Group>
                {r.days_overdue > 0 && (
                  <Text size="xs" c="red">{r.days_overdue}d overdue</Text>
                )}
              </Group>
            </Paper>
          ))}
        </Stack>
      </div>
    </Stack>
  )
}
