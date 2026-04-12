import { For, Show, createEffect, createMemo, createSignal, onCleanup, onMount } from 'solid-js'
import { createStore } from 'solid-js/store'
import { listProblemLabels, listProblems, suggestProblem } from '@/api/client'
import type { Difficulty, Problem } from '@/api/types'
import { EmptyBlock, ErrorAlert, LoadingBlock, WarningAlert } from '../components/common/Feedback'
import { PageShell } from '../components/common/PageShell'
import { Pagination } from '../components/common/Pagination'
import { Badge, Button, Card, SearchInputField, SelectField } from '../components/common/Primitives'
import { SectionLead } from '../components/common/SectionLead'
import { SectionTitle } from '../components/common/SectionTitle'
import { DifficultyBadge } from '../components/problems/Badges'
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
    tags: [] as string[]
  })

  const pageSize = 12
  const [difficulty, setDifficulty] = createSignal('')
  const [sortMode, setSortMode] = createSignal<SortMode>('default')
  const [currentPage, setCurrentPage] = createSignal(1)
  const [showTags, setShowTags] = createSignal(false)
  const [tagQuery, setTagQuery] = createSignal('')

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
    void tagKey

    let active = true
    setState({ loading: true, error: '' })

    void (async () => {
      try {
        const result = await listProblems({
          difficulty: (selectedDifficulty as Difficulty | '') || undefined,
          tags: [...state.tags],
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
  const filteredTags = createMemo(() => {
    const selected = new Set(state.tags)
    const query = tagQuery().trim().toLowerCase()

    return [...state.labels.tags]
      .filter((tag) => query === '' || tag.toLowerCase().includes(query))
      .sort((left, right) => {
        const leftSelected = selected.has(left)
        const rightSelected = selected.has(right)
        if (leftSelected !== rightSelected) {
          return leftSelected ? -1 : 1
        }
        return left.localeCompare(right)
      })
  })
  const visibleTags = createMemo(() => {
    const tags = filteredTags()
    return showTags() ? tags : tags.slice(0, 14)
  })
  const hiddenTagCount = createMemo(() => Math.max(0, filteredTags().length - visibleTags().length))
  const shouldShowTagPanel = createMemo(
    () => showTags() || state.tags.length > 0 || tagQuery().trim().length > 0
  )

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Training floor"
          title="Browse problems in Judge Loop"
          copy="The board stays fast, while the filter bar stays tight even when the tag list gets large."
        />
        <div class="flex flex-wrap gap-2">
          <Badge content={`${state.total} total problems`} color="blue" />
          <Badge content={`page ${currentPage()} / ${totalPages()}`} color="indigo" />
        </div>
      </Card>

      <Card class="space-y-4">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="space-y-1">
            <div class="text-sm font-semibold text-gray-900">Filters</div>
            <p class="text-sm text-gray-500">
              Keep the bar compact, search tags fast, and expand the tray only when needed.
            </p>
          </div>
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

        <div class="grid gap-3 lg:grid-cols-[minmax(0,220px)_minmax(0,220px)_minmax(0,1fr)] lg:items-end">
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
          <div class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
            <SearchInputField
              label="Find tag"
              value={tagQuery()}
              placeholder="Search tags"
              onInput={(event) => {
                setTagQuery(event.currentTarget.value)
                if (event.currentTarget.value.trim()) {
                  setShowTags(true)
                }
              }}
            />
            <Button
              pill
              color="light"
              class="sm:self-end"
              onClick={() => setShowTags((open) => !open)}
            >
              {showTags() ? 'Hide tags' : 'Browse tags'}
            </Button>
          </div>
        </div>

        <Show when={state.tags.length > 0}>
          <div class="rounded-2xl border border-blue-100 bg-blue-50/60 px-3 py-3">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-xs font-semibold uppercase tracking-[0.18em] text-blue-700">
                Selected tags
              </span>
              <For each={state.tags}>
                {(tag) => (
                  <Button
                    pill
                    size="xs"
                    color="blue"
                    onClick={() => {
                      setState('tags', toggleValue(state.tags, tag))
                      setCurrentPage(1)
                    }}
                  >
                    {tag} ×
                  </Button>
                )}
              </For>
              <Button
                pill
                size="xs"
                color="light"
                onClick={() => {
                  setState('tags', [])
                  setCurrentPage(1)
                }}
              >
                Clear all
              </Button>
            </div>
          </div>
        </Show>

        <Show when={!state.labelsLoading} fallback={<LoadingBlock label="Loading tags..." />}>
          <Show when={shouldShowTagPanel()}>
            <div class="rounded-2xl border border-gray-200 bg-gray-50/80 p-3">
              <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
                <div class="text-sm font-medium text-gray-700">Tags</div>
                <div class="flex flex-wrap items-center gap-2">
                  <Badge content={`${filteredTags().length} matches`} color="dark" />
                  <Show when={hiddenTagCount() > 0}>
                    <Button pill size="xs" color="light" onClick={() => setShowTags(true)}>
                      Show {hiddenTagCount()} more
                    </Button>
                  </Show>
                </div>
              </div>

              <Show
                when={visibleTags().length > 0}
                fallback={<p class="text-sm text-gray-500">No tags match the current search.</p>}
              >
                <div class="flex max-h-44 flex-wrap gap-2 overflow-y-auto pr-1">
                  <For each={visibleTags()}>
                    {(tag) => (
                      <Button
                        pill
                        size="xs"
                        aria-pressed={state.tags.includes(tag)}
                        color={state.tags.includes(tag) ? 'blue' : 'alternative'}
                        onClick={() => {
                          setState('tags', toggleValue(state.tags, tag))
                          setCurrentPage(1)
                        }}
                      >
                        {state.tags.includes(tag) ? `✓ ${tag}` : tag}
                      </Button>
                    )}
                  </For>
                </div>
              </Show>
            </div>
          </Show>
        </Show>

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
