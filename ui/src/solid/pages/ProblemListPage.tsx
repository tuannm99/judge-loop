import { For, Show, createEffect, createMemo, createSignal, onCleanup, onMount } from 'solid-js'
import { createStore } from 'solid-js/store'
import { listProblemLabels, listProblems, suggestProblem } from '@/api/client'
import type { Difficulty, Problem } from '@/api/types'
import { EmptyBlock, ErrorAlert, LoadingBlock, WarningAlert } from '../components/common/Feedback'
import { PageShell } from '../components/common/PageShell'
import { Pagination } from '../components/common/Pagination'
import { Button, Card, Badge, SelectField } from '../components/common/Primitives'
import { SectionLead } from '../components/common/SectionLead'
import { SectionTitle } from '../components/common/SectionTitle'
import { DifficultyBadge } from '../components/problems/Badges'
import { LabelButtonRow } from '../components/problems/LabelButtonRow'
import { DIFFICULTY_OPTIONS, EMPTY_LABELS, SORT_OPTIONS } from '../shared/constants'
import type { NavigateFn, SortMode } from '../shared/types'
import { formatError, sortProblems, toggleValue } from '../shared/utils'

export function ProblemListPage(props: { navigate: NavigateFn }) {
  const [state, setState] = createStore({
    loading: true,
    labelsLoading: true,
    error: '',
    labelsError: '',
    suggesting: false,
    total: 0,
    problems: [] as Problem[],
    labels: EMPTY_LABELS,
    tags: [] as string[],
    patterns: [] as string[]
  })

  const pageSize = 12
  const [difficulty, setDifficulty] = createSignal('')
  const [sortMode, setSortMode] = createSignal<SortMode>('default')
  const [currentPage, setCurrentPage] = createSignal(1)

  onMount(() => {
    let active = true

    void (async () => {
      setState({ labelsLoading: true, labelsError: '' })
      try {
        const labels = await listProblemLabels()
        if (active) {
          setState('labels', labels)
        }
      } catch (error) {
        if (active) {
          setState({
            labels: EMPTY_LABELS,
            labelsError: formatError(error)
          })
        }
      } finally {
        if (active) {
          setState('labelsLoading', false)
        }
      }
    })()

    onCleanup(() => {
      active = false
    })
  })

  createEffect(() => {
    const page = currentPage()
    const selectedDifficulty = difficulty()
    const tagKey = state.tags.join('|')
    const patternKey = state.patterns.join('|')
    void tagKey
    void patternKey

    let active = true
    setState({ loading: true, error: '' })

    void (async () => {
      try {
        const result = await listProblems({
          difficulty: (selectedDifficulty as Difficulty | '') || undefined,
          tags: [...state.tags],
          patterns: [...state.patterns],
          limit: pageSize,
          offset: (page - 1) * pageSize
        })

        if (active) {
          setState({
            problems: result.problems,
            total: result.total
          })
        }
      } catch (error) {
        if (active) {
          setState({
            error: formatError(error),
            problems: [],
            total: 0
          })
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

  const sortedProblems = createMemo(() => sortProblems(state.problems, sortMode()))
  const totalPages = createMemo(() => Math.max(1, Math.ceil(state.total / pageSize)))

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Training floor"
          title="Browse problems in Judge Loop"
          copy="The UI is now rendered by Solid with Flowbite-styled building blocks and the existing API surface."
        />
        <div class="flex flex-wrap gap-2">
          <Badge content={`${state.total} total problems`} color="blue" />
          <Badge content={`page ${currentPage()} / ${totalPages()}`} color="indigo" />
        </div>
      </Card>

      <Card class="space-y-6">
        <SectionTitle title="Filters" subtitle="Difficulty and labels still hit the same API." />

        <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
          <SelectField
            label="Difficulty"
            value={difficulty()}
            options={DIFFICULTY_OPTIONS}
            onChange={(event) => {
              setDifficulty(event.currentTarget.value)
              setCurrentPage(1)
            }}
          />
          <SelectField
            label="Sort current page"
            value={sortMode()}
            options={SORT_OPTIONS}
            onChange={(event) => {
              setSortMode(event.currentTarget.value as SortMode)
            }}
          />
          <div class="flex items-end">
            <Button
              pill
              loading={state.suggesting}
              onClick={async () => {
                setState({ suggesting: true, error: '' })
                try {
                  const problem = await suggestProblem()
                  props.navigate(`/problems/${problem.slug}`)
                } catch (error) {
                  setState('error', formatError(error))
                } finally {
                  setState('suggesting', false)
                }
              }}
            >
              Suggest one
            </Button>
          </div>
        </div>

        <LabelButtonRow
          title="Tags"
          helperText="Select multiple tags to combine problem filters."
          values={state.labels.tags}
          selected={state.tags}
          loading={state.labelsLoading}
          activeColor="blue"
          onToggle={(value) => {
            setState('tags', toggleValue(state.tags, value))
            setCurrentPage(1)
          }}
          onClear={() => {
            setState('tags', [])
            setCurrentPage(1)
          }}
        />

        <LabelButtonRow
          title="Patterns"
          helperText="Select multiple patterns to narrow the board further."
          values={state.labels.patterns}
          selected={state.patterns}
          loading={state.labelsLoading}
          activeColor="indigo"
          onToggle={(value) => {
            setState('patterns', toggleValue(state.patterns, value))
            setCurrentPage(1)
          }}
          onClear={() => {
            setState('patterns', [])
            setCurrentPage(1)
          }}
        />

        <Show when={state.error}>
          <ErrorAlert>{state.error}</ErrorAlert>
        </Show>
        <Show when={state.labelsError}>
          <WarningAlert>{state.labelsError}</WarningAlert>
        </Show>
      </Card>

      <Card class="space-y-6">
        <SectionTitle
          title="Problem board"
          subtitle="Pick a problem, open the source, or edit the registry entry."
        />

        <Show when={!state.loading} fallback={<LoadingBlock label="Loading problems..." />}>
          <Show
            when={sortedProblems().length > 0}
            fallback={
              <EmptyBlock
                title="No problems match the current filter."
                copy="Clear a chip or sync the local registry."
              />
            }
          >
            <div class="space-y-6">
              <div class="grid gap-4 lg:grid-cols-2 xl:grid-cols-3">
                <For each={sortedProblems()}>
                  {(problem) => (
                    <Card class="flex h-full flex-col gap-4">
                      <div class="flex items-start justify-between gap-4">
                        <div class="space-y-1">
                          <p class="text-xs uppercase tracking-wide text-gray-500">
                            {problem.provider}
                            {problem.external_id ? ` · #${problem.external_id}` : ''}
                          </p>
                          <h3 class="text-lg font-semibold text-gray-900">{problem.title}</h3>
                        </div>
                        <DifficultyBadge difficulty={problem.difficulty} />
                      </div>

                      <div class="flex flex-wrap gap-2 text-sm text-gray-500">
                        <span>{problem.estimated_time} min</span>
                        <span>•</span>
                        <span>{problem.slug}</span>
                      </div>

                      <div class="flex flex-wrap gap-2">
                        <For each={problem.tags}>
                          {(tag) => <Badge content={tag} color="blue" />}
                        </For>
                        <For each={problem.pattern_tags}>
                          {(pattern) => <Badge content={pattern} color="indigo" />}
                        </For>
                      </div>

                      <div class="mt-auto flex flex-wrap gap-2">
                        <Button
                          pill
                          href={`/problems/${problem.slug}`}
                          onClick={(event) => {
                            event.preventDefault()
                            props.navigate(`/problems/${problem.slug}`)
                          }}
                        >
                          Solve
                        </Button>
                        <Button
                          pill
                          color="alternative"
                          href={`/problems/${problem.slug}/edit`}
                          onClick={(event) => {
                            event.preventDefault()
                            props.navigate(`/problems/${problem.slug}/edit`)
                          }}
                        >
                          Edit
                        </Button>
                        <Show when={problem.source_url}>
                          <Button pill color="light" href={problem.source_url} target="_blank">
                            Source
                          </Button>
                        </Show>
                      </div>
                    </Card>
                  )}
                </For>
              </div>

              <Show when={state.total > pageSize}>
                <Pagination
                  page={currentPage()}
                  pageSize={pageSize}
                  total={state.total}
                  onPageChange={(page) => setCurrentPage(page)}
                />
              </Show>
            </div>
          </Show>
        </Show>
      </Card>
    </PageShell>
  )
}
