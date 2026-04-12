import { For, Show, createSignal, onCleanup, onMount } from 'solid-js'
import { createStore } from 'solid-js/store'
import {
  createProblemLabel,
  deleteProblemLabel,
  listProblemLabelRecords,
  updateProblemLabel
} from '@/api/client'
import type { ProblemLabel } from '@/api/types'
import { EmptyBlock, ErrorAlert, LoadingBlock } from '../components/common/Feedback'
import { PageShell } from '../components/common/PageShell'
import {
  Button,
  Card,
  DataTable,
  InputField,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
  Tabs
} from '../components/common'
import { SectionLead } from '../components/common/SectionLead'
import { EMPTY_DRAFT_LABEL } from '../shared/constants'
import type { DraftLabel, LabelKind, NavigateFn } from '../shared/types'
import { formatDate, formatError } from '../shared/utils'

export function ProblemLabelsPage(props: { navigate: NavigateFn }) {
  const [state, setState] = createStore({
    loading: true,
    saving: false,
    error: '',
    draft: { ...EMPTY_DRAFT_LABEL },
    editing: {} as Record<string, DraftLabel>,
    tags: [] as ProblemLabel[],
    patterns: [] as ProblemLabel[]
  })
  const [activeKind, setActiveKind] = createSignal<LabelKind>('tag')

  onMount(() => {
    let active = true

    void (async () => {
      setState('loading', true)
      try {
        const [tags, patterns] = await Promise.all([
          listProblemLabelRecords('tag'),
          listProblemLabelRecords('pattern')
        ])
        if (active) {
          setState({
            tags: tags.labels,
            patterns: patterns.labels,
            error: ''
          })
        }
      } catch (error) {
        if (active) {
          setState('error', formatError(error))
        }
      } finally {
        if (active) {
          setState('loading', false)
        }
      }
    })()

    onCleanup(() => {
      active = false
    })
  })

  const records = () => (activeKind() === 'tag' ? state.tags : state.patterns)

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Shared taxonomy"
          title="Manage labels in Solid."
          copy="Tags and patterns still save to the same endpoints, now with a simpler Solid component tree."
        />
      </Card>

      <Show when={state.error}>
        <ErrorAlert>{state.error}</ErrorAlert>
      </Show>

      <Card class="space-y-6">
        <Tabs
          value={activeKind()}
          items={[
            { value: 'tag', label: 'Tags' },
            { value: 'pattern', label: 'Patterns' }
          ]}
          onChange={(value) => setActiveKind(value as LabelKind)}
        />

        <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
          <InputField
            label="Slug"
            value={state.draft.slug}
            onInput={(event) => setState('draft', 'slug', event.currentTarget.value)}
          />
          <InputField
            label="Name"
            value={state.draft.name}
            onInput={(event) => setState('draft', 'name', event.currentTarget.value)}
          />
          <div class="flex items-end">
            <Button
              pill
              loading={state.saving}
              disabled={!state.draft.slug.trim()}
              onClick={async () => {
                setState({ saving: true, error: '' })
                try {
                  await createProblemLabel({
                    kind: activeKind(),
                    slug: state.draft.slug.trim(),
                    name: state.draft.name.trim() || state.draft.slug.trim()
                  })

                  const refreshed = await listProblemLabelRecords(activeKind())
                  if (activeKind() === 'tag') {
                    setState('tags', refreshed.labels)
                  } else {
                    setState('patterns', refreshed.labels)
                  }
                  setState('draft', { ...EMPTY_DRAFT_LABEL })
                } catch (error) {
                  setState('error', formatError(error))
                } finally {
                  setState('saving', false)
                }
              }}
            >
              Add label
            </Button>
          </div>
        </div>

        <Show when={!state.loading} fallback={<LoadingBlock label="Loading label records..." />}>
          <Show
            when={records().length > 0}
            fallback={
              <EmptyBlock
                title={`No ${activeKind()}s yet.`}
                copy="Create the first entry and it will become available across the rest of the UI."
              />
            }
          >
            <DataTable>
              <TableHead>
                <TableRow>
                  <TableHeaderCell>Slug</TableHeaderCell>
                  <TableHeaderCell>Name</TableHeaderCell>
                  <TableHeaderCell>Updated</TableHeaderCell>
                  <TableHeaderCell>Actions</TableHeaderCell>
                </TableRow>
              </TableHead>
              <TableBody>
                <For each={records()}>
                  {(label) => {
                    const editing = () => state.editing[label.id]
                    const isEditing = () => Boolean(editing())

                    return (
                      <TableRow>
                        <TableCell>
                          <Show when={isEditing()} fallback={<span>{label.slug}</span>}>
                            <input
                              class="w-full rounded-xl border border-gray-300 px-3 py-2 text-sm"
                              value={editing()?.slug ?? ''}
                              onInput={(event) =>
                                setState('editing', label.id, {
                                  ...(editing() ?? {
                                    slug: label.slug,
                                    name: label.name
                                  }),
                                  slug: event.currentTarget.value
                                })
                              }
                            />
                          </Show>
                        </TableCell>
                        <TableCell>
                          <Show when={isEditing()} fallback={<span>{label.name}</span>}>
                            <input
                              class="w-full rounded-xl border border-gray-300 px-3 py-2 text-sm"
                              value={editing()?.name ?? ''}
                              onInput={(event) =>
                                setState('editing', label.id, {
                                  ...(editing() ?? {
                                    slug: label.slug,
                                    name: label.name
                                  }),
                                  name: event.currentTarget.value
                                })
                              }
                            />
                          </Show>
                        </TableCell>
                        <TableCell>{formatDate(label.updated_at)}</TableCell>
                        <TableCell>
                          <div class="flex flex-wrap gap-2">
                            <Show
                              when={isEditing()}
                              fallback={
                                <Button
                                  pill
                                  size="xs"
                                  color="alternative"
                                  onClick={() =>
                                    setState('editing', label.id, {
                                      slug: label.slug,
                                      name: label.name
                                    })
                                  }
                                >
                                  Edit
                                </Button>
                              }
                            >
                              <Button
                                pill
                                size="xs"
                                color="green"
                                onClick={async () => {
                                  setState({ saving: true, error: '' })
                                  try {
                                    await updateProblemLabel(label.id, {
                                      slug: editing()?.slug.trim() ?? label.slug,
                                      name:
                                        editing()?.name.trim() ||
                                        editing()?.slug.trim() ||
                                        label.name
                                    })

                                    const refreshed = await listProblemLabelRecords(activeKind())
                                    if (activeKind() === 'tag') {
                                      setState('tags', refreshed.labels)
                                    } else {
                                      setState('patterns', refreshed.labels)
                                    }

                                    const nextEditing = { ...state.editing }
                                    delete nextEditing[label.id]
                                    setState('editing', nextEditing)
                                  } catch (error) {
                                    setState('error', formatError(error))
                                  } finally {
                                    setState('saving', false)
                                  }
                                }}
                              >
                                Save
                              </Button>
                              <Button
                                pill
                                size="xs"
                                color="alternative"
                                onClick={() => {
                                  const nextEditing = { ...state.editing }
                                  delete nextEditing[label.id]
                                  setState('editing', nextEditing)
                                }}
                              >
                                Cancel
                              </Button>
                            </Show>
                            <Button
                              pill
                              size="xs"
                              color="red"
                              onClick={async () => {
                                setState({ saving: true, error: '' })
                                try {
                                  await deleteProblemLabel(label.id)
                                  const refreshed = await listProblemLabelRecords(activeKind())
                                  if (activeKind() === 'tag') {
                                    setState('tags', refreshed.labels)
                                  } else {
                                    setState('patterns', refreshed.labels)
                                  }
                                } catch (error) {
                                  setState('error', formatError(error))
                                } finally {
                                  setState('saving', false)
                                }
                              }}
                            >
                              Delete
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    )
                  }}
                </For>
              </TableBody>
            </DataTable>
          </Show>
        </Show>
      </Card>

      <div class="flex justify-end">
        <Button pill color="alternative" onClick={() => props.navigate('/')}>
          Back to problems
        </Button>
      </div>
    </PageShell>
  )
}
