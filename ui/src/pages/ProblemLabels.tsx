import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createProblemLabel,
  deleteProblemLabel,
  listProblemLabelRecords,
  updateProblemLabel
} from '@/api/client'
import type { ProblemLabel } from '@/api/types'
import {
  ActionIcon,
  Alert,
  Button,
  Group,
  Loader,
  Paper,
  Stack,
  Table,
  Tabs,
  Text,
  TextInput,
  Title
} from '@mantine/core'
import { IconCheck, IconPencil, IconPlus, IconTrash, IconX } from '@tabler/icons-react'

type LabelKind = 'tag' | 'pattern'

type DraftState = {
  slug: string
  name: string
}

function emptyDraft(): DraftState {
  return { slug: '', name: '' }
}

function LabelSection({ kind, title }: { kind: LabelKind; title: string }) {
  const qc = useQueryClient()
  const [draft, setDraft] = useState<DraftState>(emptyDraft())
  const [editing, setEditing] = useState<Record<string, DraftState>>({})

  const { data, isLoading, error } = useQuery({
    queryKey: ['problem-label-records', kind],
    queryFn: () => listProblemLabelRecords(kind)
  })

  const labels = useMemo(() => data?.labels ?? [], [data])

  const invalidate = async () => {
    await qc.invalidateQueries({ queryKey: ['problem-label-records', kind] })
    await qc.invalidateQueries({ queryKey: ['problem-labels'] })
    await qc.invalidateQueries({ queryKey: ['problems'] })
  }

  const createMutation = useMutation({
    mutationFn: () =>
      createProblemLabel({
        kind,
        slug: draft.slug.trim(),
        name: draft.name.trim() || draft.slug.trim()
      }),
    onSuccess: async () => {
      setDraft(emptyDraft())
      await invalidate()
    }
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, value }: { id: string; value: DraftState }) =>
      updateProblemLabel(id, {
        slug: value.slug.trim(),
        name: value.name.trim() || value.slug.trim()
      }),
    onSuccess: async (_, vars) => {
      setEditing((current) => {
        const next = { ...current }
        delete next[vars.id]
        return next
      })
      await invalidate()
    }
  })

  const deleteMutation = useMutation({
    mutationFn: deleteProblemLabel,
    onSuccess: invalidate
  })

  function startEdit(label: ProblemLabel) {
    setEditing((current) => ({
      ...current,
      [label.id]: { slug: label.slug, name: label.name }
    }))
  }

  function cancelEdit(id: string) {
    setEditing((current) => {
      const next = { ...current }
      delete next[id]
      return next
    })
  }

  return (
    <Stack gap="md">
      <Paper withBorder p="lg">
        <Stack gap="md">
          <div>
            <Title order={4}>{title}</Title>
            <Text size="sm" c="dimmed">
              These are the predefined options used by problem contribution.
            </Text>
          </div>

          {(createMutation.error || updateMutation.error || deleteMutation.error || error) && (
            <Alert color="red" title="Label update failed">
              {String(
                createMutation.error?.message ||
                  updateMutation.error?.message ||
                  deleteMutation.error?.message ||
                  error?.message
              )}
            </Alert>
          )}

          <Group align="end">
            <TextInput
              label="Slug"
              placeholder={kind === 'tag' ? 'array' : 'two-pointers'}
              value={draft.slug}
              onChange={(e) => {
                const value = e.currentTarget.value
                setDraft((current) => ({ ...current, slug: value }))
              }}
            />
            <TextInput
              label="Name"
              placeholder={kind === 'tag' ? 'Array' : 'Two Pointers'}
              value={draft.name}
              onChange={(e) => {
                const value = e.currentTarget.value
                setDraft((current) => ({ ...current, name: value }))
              }}
            />
            <Button
              leftSection={<IconPlus size={14} />}
              onClick={() => createMutation.mutate()}
              loading={createMutation.isPending}
              disabled={!draft.slug.trim()}
            >
              Add
            </Button>
          </Group>
        </Stack>
      </Paper>

      <Paper withBorder p="lg">
        {isLoading ? (
          <Group justify="center" py="lg">
            <Loader size="sm" />
          </Group>
        ) : (
          <Table striped highlightOnHover withTableBorder>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Slug</Table.Th>
                <Table.Th>Name</Table.Th>
                <Table.Th></Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {labels.length === 0 ? (
                <Table.Tr>
                  <Table.Td colSpan={3}>
                    <Text size="sm" c="dimmed" ta="center" py="md">
                      No {kind}s configured yet.
                    </Text>
                  </Table.Td>
                </Table.Tr>
              ) : (
                labels.map((label) => {
                  const editValue = editing[label.id]
                  const isEditing = Boolean(editValue)
                  return (
                    <Table.Tr key={label.id}>
                      <Table.Td>
                        {isEditing ? (
                          <TextInput
                            value={editValue.slug}
                            onChange={(e) => {
                              const value = e.currentTarget.value
                              setEditing((current) => ({
                                ...current,
                                [label.id]: {
                                  ...(current[label.id] ?? editValue),
                                  slug: value
                                }
                              }))
                            }}
                          />
                        ) : (
                          label.slug
                        )}
                      </Table.Td>
                      <Table.Td>
                        {isEditing ? (
                          <TextInput
                            value={editValue.name}
                            onChange={(e) => {
                              const value = e.currentTarget.value
                              setEditing((current) => ({
                                ...current,
                                [label.id]: {
                                  ...(current[label.id] ?? editValue),
                                  name: value
                                }
                              }))
                            }}
                          />
                        ) : (
                          label.name
                        )}
                      </Table.Td>
                      <Table.Td>
                        <Group justify="flex-end" gap="xs">
                          {isEditing ? (
                            <>
                              <ActionIcon
                                color="teal"
                                variant="light"
                                onClick={() =>
                                  updateMutation.mutate({ id: label.id, value: editValue })
                                }
                              >
                                <IconCheck size={16} />
                              </ActionIcon>
                              <ActionIcon
                                color="gray"
                                variant="light"
                                onClick={() => cancelEdit(label.id)}
                              >
                                <IconX size={16} />
                              </ActionIcon>
                            </>
                          ) : (
                            <ActionIcon
                              variant="light"
                              onClick={() => startEdit(label)}
                            >
                              <IconPencil size={16} />
                            </ActionIcon>
                          )}
                          <ActionIcon
                            color="red"
                            variant="light"
                            onClick={() => deleteMutation.mutate(label.id)}
                          >
                            <IconTrash size={16} />
                          </ActionIcon>
                        </Group>
                      </Table.Td>
                    </Table.Tr>
                  )
                })
              )}
            </Table.Tbody>
          </Table>
        )}
      </Paper>
    </Stack>
  )
}

export function ProblemLabelsPage() {
  return (
    <Stack p="lg" gap="lg" maw={1100}>
      <div>
        <Title order={2}>Problem Labels</Title>
        <Text size="sm" c="dimmed">
          Manage shared tags and patterns from the common schema.
        </Text>
      </div>

      <Tabs defaultValue="tag">
        <Tabs.List>
          <Tabs.Tab value="tag">Tags</Tabs.Tab>
          <Tabs.Tab value="pattern">Patterns</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="tag" pt="md">
          <LabelSection kind="tag" title="Tags" />
        </Tabs.Panel>

        <Tabs.Panel value="pattern" pt="md">
          <LabelSection kind="pattern" title="Patterns" />
        </Tabs.Panel>
      </Tabs>
    </Stack>
  )
}
