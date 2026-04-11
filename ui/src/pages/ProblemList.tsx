import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  createColumnHelper,
  type SortingState
} from '@tanstack/react-table'
import { listProblemLabels, listProblems, suggestProblem } from '@/api/client'
import type { Problem, Difficulty } from '@/api/types'
import { DifficultyBadge, Badge } from '@/components/ui'
import {
  Button,
  MultiSelect,
  NativeSelect,
  Table,
  Text,
  Group,
  Stack,
  Center,
  Loader,
  Anchor,
  Pagination
} from '@mantine/core'
import {
  IconArrowsShuffle,
  IconExternalLink,
  IconPencil
} from '@tabler/icons-react'

const col = createColumnHelper<Problem>()

export function ProblemList() {
  const pageSize = 25
  const navigate = useNavigate()
  const [difficulty, setDifficulty] = useState<Difficulty | ''>('')
  const [tags, setTags] = useState<string[]>([])
  const [patterns, setPatterns] = useState<string[]>([])
  const [sorting, setSorting] = useState<SortingState>([])
  const [page, setPage] = useState(1)

  const columns = [
    col.accessor('title', {
      header: 'Title',
      cell: (i) => (
        <Text size="sm" fw={500}>
          {i.getValue()}
        </Text>
      )
    }),
    col.accessor('difficulty', {
      header: 'Difficulty',
      cell: (i) => <DifficultyBadge difficulty={i.getValue()} />
    }),
    col.accessor('tags', {
      header: 'Tags',
      enableSorting: false,
      cell: (i) => (
        <Group gap={4}>
          {i.getValue()?.map((t) => (
            <Badge key={t} size="xs" variant="default">
              {t}
            </Badge>
          ))}
        </Group>
      )
    }),
    col.accessor('pattern_tags', {
      header: 'Patterns',
      enableSorting: false,
      cell: (i) => (
        <Group gap={4}>
          {i.getValue()?.map((p) => (
            <Badge key={p} size="xs" color="violet" variant="light">
              {p}
            </Badge>
          ))}
        </Group>
      )
    }),
    col.accessor('provider', {
      header: 'Provider',
      cell: (i) => (
        <Text size="xs" c="dimmed">
          {i.getValue()}
        </Text>
      )
    }),
    col.accessor('estimated_time', {
      header: 'Est.',
      cell: (i) => (
        <Text size="xs" c="dimmed">
          {i.getValue()}m
        </Text>
      )
    }),
    col.accessor('source_url', {
      header: '',
      enableSorting: false,
      cell: (i) =>
        i.getValue() ? (
          <Anchor
            href={i.getValue()}
            target="_blank"
            onClick={(e) => e.stopPropagation()}
          >
            <IconExternalLink size={14} />
          </Anchor>
        ) : null
    }),
    col.display({
      id: 'actions',
      header: '',
      enableSorting: false,
      cell: (i) => (
        <Button
          variant="subtle"
          size="compact-sm"
          leftSection={<IconPencil size={14} />}
          onClick={(e) => {
            e.stopPropagation()
            navigate(`/problems/${i.row.original.slug}/edit`)
          }}
        >
          Edit
        </Button>
      )
    })
  ]

  useEffect(() => {
    setPage(1)
  }, [difficulty, tags, patterns])

  const { data, isLoading } = useQuery({
    queryKey: ['problems', difficulty, tags, patterns, page],
    queryFn: () =>
      listProblems({
        difficulty: difficulty || undefined,
        tags,
        patterns,
        limit: pageSize,
        offset: (page - 1) * pageSize
      }),
    placeholderData: (previousData) => previousData
  })
  const { data: labels } = useQuery({
    queryKey: ['problem-labels'],
    queryFn: listProblemLabels
  })

  const table = useReactTable({
    data: data?.problems ?? [],
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel()
  })

  async function handleSuggest() {
    try {
      const p = await suggestProblem()
      navigate(`/problems/${p.slug}`)
    } catch {
      /* no problems */
    }
  }

  return (
    <Stack p="lg" gap="md">
      <Group justify="space-between">
        <Group gap="sm">
          <Text size="xl" fw={600}>
            Problems
          </Text>
          {data && (
            <Text size="sm" c="dimmed">
              {data.total} total
            </Text>
          )}
        </Group>

        <Group gap="sm">
          <NativeSelect
            value={difficulty}
            onChange={(e) => setDifficulty(e.target.value as Difficulty | '')}
            data={[
              { value: '', label: 'All difficulties' },
              { value: 'easy', label: 'Easy' },
              { value: 'medium', label: 'Medium' },
              { value: 'hard', label: 'Hard' }
            ]}
            size="sm"
          />
          <MultiSelect
            value={tags}
            onChange={setTags}
            data={(labels?.tags ?? []).map((value) => ({
              value,
              label: value
            }))}
            placeholder="Filter tags"
            searchable
            clearable
            size="sm"
            w={220}
          />
          <MultiSelect
            value={patterns}
            onChange={setPatterns}
            data={(labels?.patterns ?? []).map((value) => ({
              value,
              label: value
            }))}
            placeholder="Filter patterns"
            searchable
            clearable
            size="sm"
            w={220}
          />
          <Button
            variant="default"
            size="sm"
            leftSection={<IconArrowsShuffle size={14} />}
            onClick={() => void handleSuggest()}
          >
            Suggest
          </Button>
        </Group>
      </Group>

      {isLoading ? (
        <Center h={200}>
          <Loader />
        </Center>
      ) : (
        <Stack gap="md">
          <Table.ScrollContainer minWidth={600}>
            <Table
              striped
              highlightOnHover
              withTableBorder
              withColumnBorders={false}
            >
              <Table.Thead>
                {table.getHeaderGroups().map((hg) => (
                  <Table.Tr key={hg.id}>
                    {hg.headers.map((h) => (
                      <Table.Th
                        key={h.id}
                        style={{
                          cursor: h.column.getCanSort() ? 'pointer' : undefined
                        }}
                        onClick={h.column.getToggleSortingHandler()}
                      >
                        <Text size="xs" tt="uppercase" fw={600} c="dimmed">
                          {flexRender(h.column.columnDef.header, h.getContext())}
                          {h.column.getIsSorted() === 'asc' && ' ↑'}
                          {h.column.getIsSorted() === 'desc' && ' ↓'}
                        </Text>
                      </Table.Th>
                    ))}
                  </Table.Tr>
                ))}
              </Table.Thead>
              <Table.Tbody>
                {table.getRowModel().rows.length === 0 ? (
                  <Table.Tr>
                    <Table.Td colSpan={columns.length}>
                      <Center py="xl">
                        <Text c="dimmed" size="sm">
                          No problems found. Run local-agent sync first.
                        </Text>
                      </Center>
                    </Table.Td>
                  </Table.Tr>
                ) : (
                  table.getRowModel().rows.map((row) => (
                    <Table.Tr
                      key={row.id}
                      style={{ cursor: 'pointer' }}
                      onClick={() => navigate(`/problems/${row.original.slug}`)}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <Table.Td key={cell.id}>
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </Table.Td>
                      ))}
                    </Table.Tr>
                  ))
                )}
              </Table.Tbody>
            </Table>
          </Table.ScrollContainer>

          {!!data?.total && data.total > pageSize && (
            <Group justify="space-between" align="center">
              <Text size="sm" c="dimmed">
                Showing {(page - 1) * pageSize + 1}-
                {Math.min(page * pageSize, data.total)} of {data.total}
              </Text>
              <Pagination
                value={page}
                onChange={setPage}
                total={Math.ceil(data.total / pageSize)}
                size="sm"
              />
            </Group>
          )}
        </Stack>
      )}
    </Stack>
  )
}
