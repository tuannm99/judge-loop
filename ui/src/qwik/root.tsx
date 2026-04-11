import {
  $,
  component$,
  type QRL,
  Slot,
  useOnWindow,
  useSignal,
  useStore,
  useVisibleTask$
} from '@builder.io/qwik'
import {
  Alert,
  Badge,
  Button,
  Card,
  Checkbox,
  FlowbiteProvider,
  Input,
  Navbar,
  Pagination,
  Select,
  Spinner,
  Table,
  Textarea
} from 'flowbite-qwik'
import {
  contributeProblem,
  createProblemLabel,
  createSubmission,
  deleteProblemLabel,
  getProblem,
  getProblemTestCases,
  getProgressToday,
  getReviewsToday,
  getStreak,
  getSubmission,
  listProblemLabelRecords,
  listProblemLabels,
  listProblems,
  listSubmissions,
  suggestProblem,
  updateProblem,
  updateProblemLabel
} from '@/api/client'
import type {
  Difficulty,
  Language,
  Problem,
  ProblemLabel,
  ProblemLabels,
  ProgressToday,
  Provider,
  ReviewItem,
  Streak,
  Submission,
  SubmissionStatus
} from '@/api/types'
import { matchRoute, type RouteMatch } from './router'

type NavigateFn = QRL<(to: string) => void>
type LabelKind = 'tag' | 'pattern'
type SortMode = 'default' | 'title' | 'difficulty' | 'time-desc' | 'provider'

type DraftTestCase = {
  input: string
  expected: string
  is_hidden: boolean
}

type DraftLabel = {
  slug: string
  name: string
}

const DEFAULT_CODE: Record<Language, string> = {
  python: '# Write your solution here\n\n',
  go: 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println()\n}\n'
}

const DEFAULT_STARTER_CODE: Record<Language, string> = {
  python: 'class Solution:\n    pass\n',
  go: 'package main\n\nfunc main() {\n\n}\n'
}

const EMPTY_LABELS: ProblemLabels = {
  tags: [],
  patterns: []
}

const EMPTY_TEST_CASE: DraftTestCase = {
  input: '',
  expected: '',
  is_hidden: false
}

const EMPTY_DRAFT_LABEL: DraftLabel = {
  slug: '',
  name: ''
}

const DIFFICULTY_ORDER: Record<Difficulty, number> = {
  easy: 1,
  medium: 2,
  hard: 3
}

const DIFFICULTY_OPTIONS = [
  { name: 'All difficulties', value: '' },
  { name: 'Easy', value: 'easy' },
  { name: 'Medium', value: 'medium' },
  { name: 'Hard', value: 'hard' }
]

const SORT_OPTIONS = [
  { name: 'Default order', value: 'default' },
  { name: 'Title A-Z', value: 'title' },
  { name: 'Difficulty', value: 'difficulty' },
  { name: 'Longest estimate', value: 'time-desc' },
  { name: 'Provider', value: 'provider' }
]

const LANGUAGE_OPTIONS = [
  { name: 'Python', value: 'python' },
  { name: 'Go', value: 'go' }
]

const PROVIDER_OPTIONS = [
  { name: 'LeetCode', value: 'leetcode' },
  { name: 'NeetCode', value: 'neetcode' },
  { name: 'HackerRank', value: 'hackerrank' }
]

function classes(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(' ')
}

function formatError(error: unknown) {
  if (error instanceof Error && error.message) return error.message
  return 'Request failed'
}

function formatElapsed(totalSeconds: number) {
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return `${minutes}:${String(seconds).padStart(2, '0')}`
}

function formatDate(value: string | undefined) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString()
}

function formatDateTime(value: string | undefined) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleString()
}

function resolveStarterCode(
  starterCode: Partial<Record<Language, string>> | undefined,
  fallback: Record<Language, string> = DEFAULT_CODE
) {
  return {
    python: starterCode?.python || fallback.python,
    go: starterCode?.go || fallback.go
  }
}

function toggleValue(values: string[], value: string) {
  return values.includes(value)
    ? values.filter((item) => item !== value)
    : [...values, value]
}

function isPending(status: SubmissionStatus) {
  return status === 'pending' || status === 'running'
}

function sortProblems(problems: Problem[], mode: SortMode) {
  const items = [...problems]

  if (mode === 'title') {
    items.sort((a, b) => a.title.localeCompare(b.title))
  } else if (mode === 'difficulty') {
    items.sort(
      (a, b) => DIFFICULTY_ORDER[a.difficulty] - DIFFICULTY_ORDER[b.difficulty]
    )
  } else if (mode === 'time-desc') {
    items.sort((a, b) => b.estimated_time - a.estimated_time)
  } else if (mode === 'provider') {
    items.sort((a, b) => a.provider.localeCompare(b.provider))
  }

  return items
}

function navKey(route: RouteMatch) {
  if (
    route.name === 'problems' ||
    route.name === 'solve-problem' ||
    route.name === 'contribute-problem' ||
    route.name === 'edit-problem'
  ) {
    return 'problems'
  }
  if (route.name === 'problem-labels') return 'labels'
  if (route.name === 'dashboard') return 'dashboard'
  return ''
}

function difficultyBadgeType(difficulty: Difficulty) {
  if (difficulty === 'easy') return 'green'
  if (difficulty === 'medium') return 'yellow'
  return 'pink'
}

function statusBadgeType(status: SubmissionStatus) {
  if (status === 'accepted') return 'green'
  if (status === 'pending' || status === 'running') return 'blue'
  return 'pink'
}

export default component$(() => {
  const route = useSignal<RouteMatch>(
    matchRoute(globalThis.location?.pathname ?? '/')
  )

  const navigate = $((to: string) => {
    const next = matchRoute(to)
    if (window.location.pathname !== next.path) {
      window.history.pushState(null, '', next.path)
    }
    route.value = next
    window.scrollTo({ top: 0, behavior: 'smooth' })
  })

  useOnWindow(
    'popstate',
    $(() => {
      route.value = matchRoute(window.location.pathname)
    })
  )

  return (
    <FlowbiteProvider theme="blue">
      <div class="min-h-screen bg-gray-50 text-gray-900">
        <div class="mx-auto max-w-7xl px-4 py-4 sm:px-6 lg:px-8">
          <AppNavbar route={route.value} navigate={navigate} />

          <main class="mt-6">
            {route.value.name === 'problems' && (
              <ProblemListPage navigate={navigate} />
            )}
            {route.value.name === 'solve-problem' && (
              <SolvePage navigate={navigate} slug={route.value.slug} />
            )}
            {route.value.name === 'contribute-problem' && (
              <ContributeProblemPage navigate={navigate} />
            )}
            {route.value.name === 'edit-problem' && (
              <ContributeProblemPage
                navigate={navigate}
                slug={route.value.slug}
              />
            )}
            {route.value.name === 'problem-labels' && (
              <ProblemLabelsPage navigate={navigate} />
            )}
            {route.value.name === 'dashboard' && (
              <DashboardPage navigate={navigate} />
            )}
            {route.value.name === 'not-found' && (
              <NotFoundPage navigate={navigate} path={route.value.path} />
            )}
          </main>
        </div>
      </div>
    </FlowbiteProvider>
  )
})

const AppNavbar = component$(
  (props: { route: RouteMatch; navigate: NavigateFn }) => {
    const state = useStore({
      progress: null as ProgressToday | null,
      error: ''
    })

    useVisibleTask$(({ cleanup }) => {
      let active = true

      const load = async () => {
        try {
          const progress = await getProgressToday()
          if (active) {
            state.progress = progress
            state.error = ''
          }
        } catch (error) {
          if (active) {
            state.error = formatError(error)
          }
        }
      }

      void load()
      const timer = window.setInterval(() => {
        void load()
      }, 60_000)

      cleanup(() => {
        active = false
        clearInterval(timer)
      })
    })

    const current = navKey(props.route)

    return (
      <Navbar
        border
        rounded
        fullWidth
        class="sticky top-4 z-20 rounded-2xl border border-gray-200 bg-white/95 px-2 shadow-sm backdrop-blur"
      >
        <Navbar.Brand
          href="/"
          onClick$={(event) => {
            event.preventDefault()
            props.navigate('/')
          }}
        >
          <div class="flex items-center gap-3">
            <span class="inline-flex size-10 items-center justify-center rounded-xl bg-blue-600 text-sm font-bold uppercase tracking-wide text-white">
              jl
            </span>
            <div>
              <div class="text-sm font-semibold uppercase tracking-[0.22em] text-gray-900">
                judge-loop
              </div>
              <div class="text-xs text-gray-500">Qwik + Flowbite</div>
            </div>
          </div>
        </Navbar.Brand>

        <div class="flex items-center gap-2 md:order-2">
          {state.progress ? (
            <>
              <Badge content={`${state.progress.solved} solved`} type="green" />
              <Badge
                content={`streak ${state.progress.streak}`}
                type="indigo"
              />
            </>
          ) : (
            <Badge content={state.error || 'syncing'} type="dark" />
          )}
          <Navbar.Toggle />
        </div>

        <Navbar.Collapse>
          <Navbar.Link
            href="/"
            active={current === 'problems'}
            onClick$={(event) => {
              event.preventDefault()
              props.navigate('/')
            }}
          >
            Problems
          </Navbar.Link>
          <Navbar.Link
            href="/problems/contribute"
            active={props.route.name === 'contribute-problem'}
            onClick$={(event) => {
              event.preventDefault()
              props.navigate('/problems/contribute')
            }}
          >
            New Problem
          </Navbar.Link>
          <Navbar.Link
            href="/problem-labels"
            active={current === 'labels'}
            onClick$={(event) => {
              event.preventDefault()
              props.navigate('/problem-labels')
            }}
          >
            Labels
          </Navbar.Link>
          <Navbar.Link
            href="/dashboard"
            active={current === 'dashboard'}
            onClick$={(event) => {
              event.preventDefault()
              props.navigate('/dashboard')
            }}
          >
            Dashboard
          </Navbar.Link>
        </Navbar.Collapse>
      </Navbar>
    )
  }
)

const ProblemListPage = component$((props: { navigate: NavigateFn }) => {
  const state = useStore({
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
  const difficulty = useSignal('')
  const sortMode = useSignal<SortMode>('default')
  const currentPage = useSignal(1)

  useVisibleTask$(async ({ track }) => {
    track(() => currentPage.value)
    track(() => difficulty.value)
    track(() => state.tags.join('|'))
    track(() => state.patterns.join('|'))

    state.loading = true
    state.error = ''

    try {
      const result = await listProblems({
        difficulty: (difficulty.value as Difficulty | '') || undefined,
        tags: state.tags,
        patterns: state.patterns,
        limit: pageSize,
        offset: (currentPage.value - 1) * pageSize
      })

      state.problems = result.problems
      state.total = result.total
    } catch (error) {
      state.error = formatError(error)
      state.problems = []
      state.total = 0
    } finally {
      state.loading = false
    }
  })

  useVisibleTask$(async () => {
    state.labelsLoading = true
    try {
      state.labels = await listProblemLabels()
      state.labelsError = ''
    } catch (error) {
      state.labelsError = formatError(error)
      state.labels = EMPTY_LABELS
    } finally {
      state.labelsLoading = false
    }
  })

  const sortedProblems = sortProblems(state.problems, sortMode.value)
  const totalPages = Math.max(1, Math.ceil(state.total / pageSize))

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Training floor"
          title="Solve from a pre-built Qwik component shell."
          copy="This screen now leans on Flowbite Qwik primitives and Tailwind utilities instead of the old handwritten stylesheet."
        />
        <div class="flex flex-wrap gap-2">
          <Badge content={`${state.total} total problems`} type="blue" />
          <Badge
            content={`page ${currentPage.value} / ${totalPages}`}
            type="indigo"
          />
        </div>
      </Card>

      <Card class="space-y-6">
        <SectionTitle
          title="Filters"
          subtitle="Difficulty and labels still hit the same API."
        />

        <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
          <Select
            label="Difficulty"
            bind:value={difficulty}
            options={DIFFICULTY_OPTIONS}
            onChange$={() => {
              currentPage.value = 1
            }}
          />
          <Select
            label="Sort current page"
            bind:value={sortMode as unknown as typeof difficulty}
            options={SORT_OPTIONS}
          />
          <div class="flex items-end">
            <Button
              pill
              loading={state.suggesting}
              onClick$={async () => {
                state.suggesting = true
                state.error = ''
                try {
                  const problem = await suggestProblem()
                  props.navigate(`/problems/${problem.slug}`)
                } catch (error) {
                  state.error = formatError(error)
                } finally {
                  state.suggesting = false
                }
              }}
            >
              Suggest one
            </Button>
          </div>
        </div>

        <LabelButtonRow
          title="Tags"
          values={state.labels.tags}
          selected={state.tags}
          loading={state.labelsLoading}
          activeColor="blue"
          onToggle$={$((value: string) => {
            state.tags = toggleValue(state.tags, value)
            currentPage.value = 1
          })}
        />

        <LabelButtonRow
          title="Patterns"
          values={state.labels.patterns}
          selected={state.patterns}
          loading={state.labelsLoading}
          activeColor="purple"
          onToggle$={$((value: string) => {
            state.patterns = toggleValue(state.patterns, value)
            currentPage.value = 1
          })}
        />

        {state.error && <ErrorAlert>{state.error}</ErrorAlert>}
        {state.labelsError && <WarningAlert>{state.labelsError}</WarningAlert>}
      </Card>

      <Card class="space-y-6">
        <SectionTitle
          title="Problem board"
          subtitle="Pick a problem, open the source, or edit the registry entry."
        />

        {state.loading ? (
          <LoadingBlock label="Loading problems..." />
        ) : sortedProblems.length === 0 ? (
          <EmptyBlock
            title="No problems match the current filter."
            copy="Clear a chip or sync the local registry."
          />
        ) : (
          <>
            <div class="grid gap-4 lg:grid-cols-2 xl:grid-cols-3">
              {sortedProblems.map((problem) => (
                <Card key={problem.id} class="flex h-full flex-col gap-4">
                  <div class="flex items-start justify-between gap-4">
                    <div class="space-y-1">
                      <p class="text-xs uppercase tracking-wide text-gray-500">
                        {problem.provider}
                        {problem.external_id ? ` · #${problem.external_id}` : ''}
                      </p>
                      <h3 class="text-lg font-semibold text-gray-900">
                        {problem.title}
                      </h3>
                    </div>
                    <DifficultyBadge difficulty={problem.difficulty} />
                  </div>

                  <div class="flex flex-wrap gap-2 text-sm text-gray-500">
                    <span>{problem.estimated_time} min</span>
                    <span>•</span>
                    <span>{problem.slug}</span>
                  </div>

                  <div class="flex flex-wrap gap-2">
                    {problem.tags.map((tag) => (
                      <Badge key={tag} content={tag} type="blue" />
                    ))}
                    {problem.pattern_tags.map((pattern) => (
                      <Badge key={pattern} content={pattern} type="indigo" />
                    ))}
                  </div>

                  <div class="mt-auto flex flex-wrap gap-2">
                    <Button
                      pill
                      onClick$={() => props.navigate(`/problems/${problem.slug}`)}
                    >
                      Solve
                    </Button>
                    <Button
                      pill
                      color="alternative"
                      onClick$={() =>
                        props.navigate(`/problems/${problem.slug}/edit`)
                      }
                    >
                      Edit
                    </Button>
                    {problem.source_url && (
                      <Button
                        pill
                        color="light"
                        href={problem.source_url}
                        target="_blank"
                      >
                        Source
                      </Button>
                    )}
                  </div>
                </Card>
              ))}
            </div>

            {state.total > pageSize && (
              <div class="flex flex-col gap-4 border-t border-gray-100 pt-4 md:flex-row md:items-center md:justify-between">
                <p class="text-sm text-gray-500">
                  Showing {(currentPage.value - 1) * pageSize + 1}-
                  {Math.min(currentPage.value * pageSize, state.total)} of{' '}
                  {state.total}
                </p>
                <Pagination
                  currentPage={currentPage}
                  totalPages={totalPages}
                  onPageChange$={$((page: number) => {
                    currentPage.value = page
                  })}
                />
              </div>
            )}
          </>
        )}
      </Card>
    </PageShell>
  )
})

const SolvePage = component$((props: { navigate: NavigateFn; slug: string }) => {
  const state = useStore({
    loadingProblem: true,
    problemError: '',
    actionError: '',
    problem: null as Problem | null,
    submissions: [] as Submission[],
    verdict: null as Submission | null,
    selectedSubmission: null as Submission | null,
    selectedSubmissionId: '',
    selectedLoading: false,
    codeByLanguage: { ...DEFAULT_CODE },
    tab: 'description' as 'description' | 'submissions',
    submitting: false,
    pollingId: '',
    timerActive: false,
    startedAt: 0,
    elapsed: 0
  })

  const language = useSignal('python')

  useVisibleTask$(async ({ track }) => {
    const slug = track(() => props.slug)

    state.loadingProblem = true
    state.problemError = ''
    state.actionError = ''
    state.problem = null
    state.submissions = []
    state.verdict = null
    state.selectedSubmission = null
    state.selectedSubmissionId = ''
    state.codeByLanguage = { ...DEFAULT_CODE }
    state.tab = 'description'
    state.pollingId = ''
    state.timerActive = false
    state.startedAt = 0
    state.elapsed = 0
    language.value = 'python'

    try {
      const problem = await getProblem(slug)
      const history = await listSubmissions(problem.id)

      state.problem = problem
      state.codeByLanguage = resolveStarterCode(problem.starter_code)
      state.submissions = history.submissions
      state.selectedSubmissionId = history.submissions[0]?.id ?? ''
    } catch (error) {
      state.problemError = formatError(error)
    } finally {
      state.loadingProblem = false
    }
  })

  useVisibleTask$(async ({ track }) => {
    const id = track(() => state.selectedSubmissionId)
    if (!id) {
      state.selectedSubmission = null
      return
    }

    state.selectedLoading = true
    try {
      state.selectedSubmission = await getSubmission(id)
      state.actionError = ''
    } catch (error) {
      state.actionError = formatError(error)
    } finally {
      state.selectedLoading = false
    }
  })

  useVisibleTask$(({ track, cleanup }) => {
    const id = track(() => state.pollingId)
    if (!id) return

    let active = true

    const load = async () => {
      try {
        const submission = await getSubmission(id)
        if (!active) return

        state.verdict = submission
        state.selectedSubmission = submission
        state.selectedSubmissionId = submission.id

        if (!isPending(submission.status)) {
          state.pollingId = ''
          if (state.problem?.id) {
            const history = await listSubmissions(state.problem.id)
            if (active) {
              state.submissions = history.submissions
            }
          }
        }
      } catch (error) {
        if (active) {
          state.actionError = formatError(error)
          state.pollingId = ''
        }
      }
    }

    void load()
    const timer = window.setInterval(() => {
      void load()
    }, 1500)

    cleanup(() => {
      active = false
      clearInterval(timer)
    })
  })

  useVisibleTask$(({ track, cleanup }) => {
    const timerActive = track(() => state.timerActive)
    const startedAt = track(() => state.startedAt)
    if (!timerActive || !startedAt) return

    state.elapsed = Math.floor((Date.now() - startedAt) / 1000)

    const timer = window.setInterval(() => {
      state.elapsed = Math.floor((Date.now() - startedAt) / 1000)
    }, 1000)

    cleanup(() => {
      clearInterval(timer)
    })
  })

  const focusedSubmission =
    state.selectedSubmission ??
    (state.verdict && state.verdict.id === state.selectedSubmissionId
      ? state.verdict
      : null) ??
    state.verdict

  return (
    <PageShell>
      <Card class="space-y-4">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-3">
            <div class="text-xs uppercase tracking-[0.2em] text-gray-500">
              Live run
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <h1 class="text-3xl font-semibold tracking-tight text-gray-900">
                {state.problem?.title ?? 'Loading problem'}
              </h1>
              {state.problem && (
                <DifficultyBadge difficulty={state.problem.difficulty} />
              )}
            </div>
            <p class="max-w-3xl text-sm text-gray-500">
              The screen is now built from Flowbite Qwik cards, inputs, buttons,
              badges, and alerts. The backend interaction is unchanged.
            </p>
          </div>

          <div class="flex flex-wrap gap-2">
            <Button
              pill
              color="alternative"
              onClick$={() => props.navigate('/')}
            >
              Back to board
            </Button>
            {state.problem?.source_url && (
              <Button
                pill
                color="light"
                href={state.problem.source_url}
                target="_blank"
              >
                Open source
              </Button>
            )}
          </div>
        </div>
      </Card>

      {state.problemError && <ErrorAlert>{state.problemError}</ErrorAlert>}
      {state.actionError && <WarningAlert>{state.actionError}</WarningAlert>}

      {state.loadingProblem ? (
        <Card>
          <LoadingBlock label="Loading problem and submission history..." />
        </Card>
      ) : state.problem ? (
        <div class="grid gap-6 xl:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
          <Card class="space-y-5">
            <div class="flex flex-wrap items-center gap-2 border-b border-gray-100 pb-4">
              <Button
                pill
                color={state.tab === 'description' ? 'blue' : 'alternative'}
                onClick$={() => {
                  state.tab = 'description'
                }}
              >
                Description
              </Button>
              <Button
                pill
                color={state.tab === 'submissions' ? 'blue' : 'alternative'}
                onClick$={() => {
                  state.tab = 'submissions'
                }}
              >
                Submissions ({state.submissions.length})
              </Button>
            </div>

            {state.tab === 'description' ? (
              <div class="space-y-4">
                <div class="flex flex-wrap gap-2">
                  {state.problem.tags.map((tag) => (
                    <Badge key={tag} content={tag} type="blue" />
                  ))}
                  {state.problem.pattern_tags.map((pattern) => (
                    <Badge key={pattern} content={pattern} type="indigo" />
                  ))}
                </div>

                <div class="rounded-xl border border-dashed border-gray-200 bg-gray-50 p-4 text-sm text-gray-600">
                  No full description is stored locally. Use the source link for
                  the provider’s original prompt and examples.
                </div>

                <div class="grid gap-3 rounded-xl bg-gray-50 p-4 text-sm text-gray-500 md:grid-cols-3">
                  <span>provider: {state.problem.provider}</span>
                  <span>slug: {state.problem.slug}</span>
                  <span>estimate: {state.problem.estimated_time} min</span>
                </div>
              </div>
            ) : (
              <div class="space-y-4">
                {state.submitting || state.pollingId ? (
                  <InfoAlert>Submission is being evaluated.</InfoAlert>
                ) : null}

                {focusedSubmission && !isPending(focusedSubmission.status) && (
                  <Card
                    class={classes(
                      'space-y-4 border',
                      focusedSubmission.status === 'accepted'
                        ? 'border-green-200 bg-green-50'
                        : 'border-red-200 bg-red-50'
                    )}
                  >
                    <div class="flex flex-wrap items-center justify-between gap-2">
                      <StatusBadge
                        status={focusedSubmission.status}
                        label={focusedSubmission.verdict || undefined}
                      />
                      <span class="text-sm text-gray-500">
                        {formatDateTime(focusedSubmission.submitted_at)}
                      </span>
                    </div>
                    <div class="flex flex-wrap gap-3 text-sm text-gray-600">
                      <span>
                        Passed {focusedSubmission.passed_cases}/
                        {focusedSubmission.total_cases}
                      </span>
                      <span>
                        {focusedSubmission.runtime_ms > 0
                          ? `${focusedSubmission.runtime_ms} ms`
                          : 'pending runtime'}
                      </span>
                    </div>
                    {focusedSubmission.error_message && (
                      <pre class="overflow-x-auto rounded-lg bg-gray-900 p-4 text-xs text-gray-100">
                        {focusedSubmission.error_message}
                      </pre>
                    )}
                    <pre class="overflow-x-auto rounded-lg bg-gray-900 p-4 text-xs text-gray-100">
                      {focusedSubmission.code}
                    </pre>
                  </Card>
                )}

                {state.selectedLoading && (
                  <LoadingInline label="Loading submission details..." />
                )}

                {state.submissions.length === 0 ? (
                  <EmptyBlock
                    title="No submissions yet."
                    copy="Ship one from the editor and the history feed will populate here."
                  />
                ) : (
                  <div class="space-y-2">
                    {state.submissions.map((submission) => (
                      <button
                        key={submission.id}
                        class={classes(
                          'flex w-full items-center justify-between rounded-xl border px-4 py-3 text-left transition',
                          state.selectedSubmissionId === submission.id
                            ? 'border-blue-200 bg-blue-50'
                            : 'border-gray-200 bg-white hover:border-gray-300'
                        )}
                        onClick$={() => {
                          state.selectedSubmissionId = submission.id
                          state.tab = 'submissions'
                        }}
                      >
                        <StatusBadge
                          status={submission.status}
                          label={submission.verdict || undefined}
                        />
                        <span class="text-sm text-gray-500">
                          {formatDateTime(submission.submitted_at)}
                        </span>
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}
          </Card>

          <Card class="space-y-5">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div class="grid gap-4 sm:grid-cols-[minmax(0,220px)_auto] sm:items-end">
                <Select
                  label="Language"
                  bind:value={language}
                  options={LANGUAGE_OPTIONS}
                />
                <Badge
                  content={`timer ${formatElapsed(state.elapsed)}`}
                  type="dark"
                />
              </div>

              <Button
                pill
                loading={state.submitting}
                disabled={!state.problem}
                onClick$={async () => {
                  if (!state.problem) return

                  state.submitting = true
                  state.actionError = ''

                  try {
                    const selectedLanguage = language.value as Language
                    const result = await createSubmission({
                      problem_id: state.problem.id,
                      language: selectedLanguage,
                      code: state.codeByLanguage[selectedLanguage]
                    })

                    state.tab = 'submissions'
                    state.timerActive = false
                    state.startedAt = 0
                    state.elapsed = 0
                    state.pollingId = result.submission_id
                    state.selectedSubmissionId = result.submission_id
                  } catch (error) {
                    state.actionError = formatError(error)
                  } finally {
                    state.submitting = false
                  }
                }}
              >
                Submit
              </Button>
            </div>

            <Textarea
              label="Code"
              rows={24}
              class="w-full"
              value={state.codeByLanguage[language.value as Language]}
              onInput$={(value: string) => {
                const selectedLanguage = language.value as Language
                if (!state.timerActive) {
                  state.timerActive = true
                  state.startedAt = Date.now()
                  state.elapsed = 0
                }

                state.codeByLanguage[selectedLanguage] = value
              }}
            />
          </Card>
        </div>
      ) : null}
    </PageShell>
  )
})

const DashboardPage = component$((props: { navigate: NavigateFn }) => {
  const state = useStore({
    loading: true,
    error: '',
    progress: null as ProgressToday | null,
    streak: null as Streak | null,
    reviews: [] as ReviewItem[]
  })

  useVisibleTask$(({ cleanup }) => {
    let active = true

    const load = async () => {
      state.loading = true
      try {
        const [progress, streak, reviews] = await Promise.all([
          getProgressToday(),
          getStreak(),
          getReviewsToday()
        ])

        if (active) {
          state.progress = progress
          state.streak = streak
          state.reviews = reviews.reviews
          state.error = ''
        }
      } catch (error) {
        if (active) {
          state.error = formatError(error)
        }
      } finally {
        if (active) {
          state.loading = false
        }
      }
    }

    void load()
    const timer = window.setInterval(() => {
      void load()
    }, 30_000)

    cleanup(() => {
      active = false
      clearInterval(timer)
    })
  })

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Daily signal"
          title="Track progress with pre-built cards and badges."
          copy="The dashboard logic is untouched; only the presentation layer moved to Flowbite Qwik."
        />
      </Card>

      {state.error && <ErrorAlert>{state.error}</ErrorAlert>}

      {state.loading ? (
        <Card>
          <LoadingBlock label="Loading dashboard..." />
        </Card>
      ) : (
        <>
          <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <StatCard
              label="Solved today"
              value={String(state.progress?.solved ?? 0)}
            />
            <StatCard
              label="Attempted"
              value={String(state.progress?.attempted ?? 0)}
            />
            <StatCard
              label="Time today"
              value={`${state.progress?.time_spent_minutes ?? 0}m`}
            />
            <StatCard
              label="Current streak"
              value={String(state.streak?.current ?? 0)}
            />
          </div>

          <Card class="space-y-4">
            <SectionTitle
              title="Streak line"
              subtitle={
                state.progress?.date ? `As of ${state.progress.date}` : undefined
              }
            />
            <div class="grid gap-3 text-sm text-gray-600 md:grid-cols-3">
              <div class="rounded-xl bg-gray-50 p-4">
                current: {state.streak?.current ?? 0}
              </div>
              <div class="rounded-xl bg-gray-50 p-4">
                longest: {state.streak?.longest ?? 0}
              </div>
              <div class="rounded-xl bg-gray-50 p-4">
                last practiced: {formatDate(state.streak?.last_practiced)}
              </div>
            </div>
          </Card>

          <Card class="space-y-4">
            <SectionTitle
              title="Reviews due today"
              subtitle="Jump back into stale problems before the streak cools off."
            />
            {state.reviews.length === 0 ? (
              <EmptyBlock
                title="No reviews due."
                copy="You have a clean board right now."
              />
            ) : (
              <div class="space-y-3">
                {state.reviews.map((review) => (
                  <button
                    key={review.problem_id}
                    class="flex w-full items-center justify-between rounded-xl border border-gray-200 bg-white px-4 py-4 text-left transition hover:border-gray-300"
                    onClick$={() => props.navigate(`/problems/${review.slug}`)}
                  >
                    <div class="space-y-1">
                      <h3 class="text-base font-semibold text-gray-900">
                        {review.title}
                      </h3>
                      <p class="text-sm text-gray-500">slug: {review.slug}</p>
                    </div>
                    <Badge
                      content={
                        review.days_overdue > 0
                          ? `${review.days_overdue}d overdue`
                          : 'due today'
                      }
                      type={review.days_overdue > 0 ? 'pink' : 'blue'}
                    />
                  </button>
                ))}
              </div>
            )}
          </Card>
        </>
      )}
    </PageShell>
  )
})

const ContributeProblemPage = component$(
  (props: { navigate: NavigateFn; slug?: string }) => {
    const isEditMode = Boolean(props.slug)
    const state = useStore({
      loading: isEditMode,
      saving: false,
      error: '',
      labelsLoading: true,
      labelsError: '',
      problemID: '',
      external_id: '',
      slug: '',
      title: '',
      tags: [] as string[],
      pattern_tags: [] as string[],
      source_url: '',
      estimated_time: 15,
      version: 1,
      starter_code: { ...DEFAULT_STARTER_CODE },
      test_cases: [{ ...EMPTY_TEST_CASE }] as DraftTestCase[],
      labels: EMPTY_LABELS
    })

    const provider = useSignal('leetcode')
    const difficulty = useSignal('easy')

    useVisibleTask$(async () => {
      state.labelsLoading = true
      try {
        state.labels = await listProblemLabels()
        state.labelsError = ''
      } catch (error) {
        state.labelsError = formatError(error)
      } finally {
        state.labelsLoading = false
      }
    })

    useVisibleTask$(async ({ track }) => {
      const slug = track(() => props.slug ?? '')
      if (!slug) {
        state.loading = false
        return
      }

      state.loading = true
      state.error = ''

      try {
        const problem = await getProblem(slug)
        const testCases = await getProblemTestCases(problem.id)

        state.problemID = problem.id
        state.external_id = problem.external_id
        state.slug = problem.slug
        state.title = problem.title
        state.tags = [...problem.tags]
        state.pattern_tags = [...problem.pattern_tags]
        state.source_url = problem.source_url
        state.estimated_time = problem.estimated_time
        state.starter_code = resolveStarterCode(
          problem.starter_code,
          DEFAULT_STARTER_CODE
        )
        state.test_cases =
          testCases.test_cases.length > 0
            ? testCases.test_cases.map((testCase) => ({
                input: testCase.input,
                expected: testCase.expected,
                is_hidden: Boolean(testCase.is_hidden)
              }))
            : [{ ...EMPTY_TEST_CASE }]

        provider.value = problem.provider
        difficulty.value = problem.difficulty
      } catch (error) {
        state.error = formatError(error)
      } finally {
        state.loading = false
      }
    })

    return (
      <PageShell>
        <Card class="space-y-4">
          <SectionLead
            eyebrow="Registry editor"
            title={isEditMode ? 'Update a stored problem.' : 'Add a new problem.'}
            copy="This form now uses Flowbite Qwik inputs, selects, textareas, buttons, and checkboxes."
          />
        </Card>

        {state.error && <ErrorAlert>{state.error}</ErrorAlert>}
        {state.labelsError && <WarningAlert>{state.labelsError}</WarningAlert>}

        {state.loading ? (
          <Card>
            <LoadingBlock label="Loading problem metadata..." />
          </Card>
        ) : (
          <>
            <Card class="space-y-6">
              <SectionTitle
                title="Problem metadata"
                subtitle="Provider, difficulty, tags, and source metadata."
              />

              <div class="grid gap-4 md:grid-cols-3">
                <Select
                  label="Provider"
                  bind:value={provider}
                  options={PROVIDER_OPTIONS}
                />
                <Input
                  label="External ID"
                  value={state.external_id}
                  onInput$={(_, el) => {
                    state.external_id = el.value
                  }}
                />
                <Select
                  label="Difficulty"
                  bind:value={difficulty}
                  options={DIFFICULTY_OPTIONS.slice(1)}
                />
              </div>

              <div class="grid gap-4 md:grid-cols-2">
                <Input
                  label="Slug"
                  value={state.slug}
                  onInput$={(_, el) => {
                    state.slug = el.value
                  }}
                />
                <Input
                  label="Title"
                  value={state.title}
                  onInput$={(_, el) => {
                    state.title = el.value
                  }}
                />
              </div>

              <Input
                label="Source URL"
                value={state.source_url}
                onInput$={(_, el) => {
                  state.source_url = el.value
                }}
              />

              <div class="grid gap-4 md:grid-cols-2">
                <Input
                  label="Estimated time (minutes)"
                  type="number"
                  value={String(state.estimated_time)}
                  onInput$={(_, el) => {
                    state.estimated_time = Number(el.value) || 0
                  }}
                />
                {!isEditMode && (
                  <Input
                    label="Version"
                    type="number"
                    value={String(state.version)}
                    onInput$={(_, el) => {
                      state.version = Math.max(1, Number(el.value) || 1)
                    }}
                  />
                )}
              </div>

              <LabelButtonRow
                title="Tags"
                values={state.labels.tags}
                selected={state.tags}
                loading={state.labelsLoading}
                activeColor="blue"
                onToggle$={$((value: string) => {
                  state.tags = toggleValue(state.tags, value)
                })}
              />

              <LabelButtonRow
                title="Pattern tags"
                values={state.labels.patterns}
                selected={state.pattern_tags}
                loading={state.labelsLoading}
                activeColor="purple"
                onToggle$={$((value: string) => {
                  state.pattern_tags = toggleValue(state.pattern_tags, value)
                })}
              />
            </Card>

            <Card class="space-y-6">
              <SectionTitle
                title="Starter code"
                subtitle="Language drafts are saved with the registry entry."
              />

              <div class="grid gap-4 xl:grid-cols-2">
                <Textarea
                  label="Python"
                  rows={12}
                  value={state.starter_code.python}
                  onInput$={(value: string) => {
                    state.starter_code.python = value
                  }}
                />
                <Textarea
                  label="Go"
                  rows={12}
                  value={state.starter_code.go}
                  onInput$={(value: string) => {
                    state.starter_code.go = value
                  }}
                />
              </div>
            </Card>

            <Card class="space-y-6">
              <SectionTitle
                title="Judge test cases"
                subtitle="Add visible or hidden cases before shipping the problem."
              />

              <div class="space-y-4">
                {state.test_cases.map((testCase, index) => (
                  <Card key={index} class="space-y-4 border">
                    <div class="flex flex-wrap items-center justify-between gap-2">
                      <h3 class="text-base font-semibold text-gray-900">
                        Case {index + 1}
                      </h3>
                      <Button
                        pill
                        color="light"
                        disabled={state.test_cases.length === 1}
                        onClick$={() => {
                          if (state.test_cases.length === 1) return
                          state.test_cases = state.test_cases.filter(
                            (_, itemIndex) => itemIndex !== index
                          )
                        }}
                      >
                        Remove
                      </Button>
                    </div>

                    <div class="grid gap-4 xl:grid-cols-2">
                      <Textarea
                        label="Input"
                        rows={8}
                        value={testCase.input}
                        onInput$={(value: string) => {
                          state.test_cases = state.test_cases.map(
                            (item, itemIndex) =>
                              itemIndex === index
                                ? { ...item, input: value }
                                : item
                          )
                        }}
                      />
                      <Textarea
                        label="Expected output"
                        rows={8}
                        value={testCase.expected}
                        onInput$={(value: string) => {
                          state.test_cases = state.test_cases.map(
                            (item, itemIndex) =>
                              itemIndex === index
                                ? { ...item, expected: value }
                                : item
                          )
                        }}
                      />
                    </div>

                    <Checkbox
                      checked={testCase.is_hidden}
                      onChange$={(checked: boolean) => {
                        state.test_cases = state.test_cases.map(
                          (item, itemIndex) =>
                            itemIndex === index
                              ? { ...item, is_hidden: checked }
                              : item
                        )
                      }}
                    >
                      Hidden test case
                    </Checkbox>
                  </Card>
                ))}

                <div class="flex flex-wrap justify-between gap-3">
                  <Button
                    pill
                    color="alternative"
                    onClick$={() => {
                      state.test_cases = [...state.test_cases, { ...EMPTY_TEST_CASE }]
                    }}
                  >
                    Add test case
                  </Button>

                  <Button
                    pill
                    loading={state.saving}
                    disabled={!state.slug.trim() || !state.title.trim()}
                    onClick$={async () => {
                      state.saving = true
                      state.error = ''

                      try {
                        const payload = {
                          provider: provider.value as Provider,
                          external_id: state.external_id,
                          slug: state.slug.trim(),
                          title: state.title.trim(),
                          difficulty: difficulty.value as Difficulty,
                          tags: [...state.tags],
                          pattern_tags: [...state.pattern_tags],
                          source_url: state.source_url.trim(),
                          estimated_time: state.estimated_time || 0,
                          starter_code: {
                            python: state.starter_code.python,
                            go: state.starter_code.go
                          },
                          test_cases: state.test_cases.map((testCase) => ({
                            input: testCase.input,
                            expected: testCase.expected,
                            is_hidden: testCase.is_hidden
                          }))
                        }

                        const saved = props.slug
                          ? await updateProblem(state.problemID, payload)
                          : await contributeProblem({
                              ...payload,
                              version: state.version || 1
                            })

                        props.navigate(`/problems/${saved.slug}`)
                      } catch (error) {
                        state.error = formatError(error)
                      } finally {
                        state.saving = false
                      }
                    }}
                  >
                    {isEditMode ? 'Update problem' : 'Save problem'}
                  </Button>
                </div>
              </div>
            </Card>
          </>
        )}
      </PageShell>
    )
  }
)

const ProblemLabelsPage = component$((props: { navigate: NavigateFn }) => {
  const state = useStore({
    loading: true,
    saving: false,
    error: '',
    draft: { ...EMPTY_DRAFT_LABEL },
    editing: {} as Record<string, DraftLabel>,
    tags: [] as ProblemLabel[],
    patterns: [] as ProblemLabel[]
  })

  const activeKind = useSignal<LabelKind>('tag')

  useVisibleTask$(async () => {
    state.loading = true
    try {
      const [tags, patterns] = await Promise.all([
        listProblemLabelRecords('tag'),
        listProblemLabelRecords('pattern')
      ])
      state.tags = tags.labels
      state.patterns = patterns.labels
      state.error = ''
    } catch (error) {
      state.error = formatError(error)
    } finally {
      state.loading = false
    }
  })

  const records = activeKind.value === 'tag' ? state.tags : state.patterns

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Shared taxonomy"
          title="Manage labels through pre-built table and form components."
          copy="Tags and patterns still save to the same endpoints; the presentation is now Flowbite-driven."
        />
      </Card>

      {state.error && <ErrorAlert>{state.error}</ErrorAlert>}

      <Card class="space-y-6">
        <div class="flex flex-wrap gap-2">
          <Button
            pill
            color={activeKind.value === 'tag' ? 'blue' : 'alternative'}
            onClick$={() => {
              activeKind.value = 'tag'
            }}
          >
            Tags
          </Button>
          <Button
            pill
            color={activeKind.value === 'pattern' ? 'blue' : 'alternative'}
            onClick$={() => {
              activeKind.value = 'pattern'
            }}
          >
            Patterns
          </Button>
        </div>

        <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
          <Input
            label="Slug"
            value={state.draft.slug}
            onInput$={(_, el) => {
              state.draft = { ...state.draft, slug: el.value }
            }}
          />
          <Input
            label="Name"
            value={state.draft.name}
            onInput$={(_, el) => {
              state.draft = { ...state.draft, name: el.value }
            }}
          />
          <div class="flex items-end">
            <Button
              pill
              loading={state.saving}
              disabled={!state.draft.slug.trim()}
              onClick$={async () => {
                state.saving = true
                state.error = ''
                try {
                  await createProblemLabel({
                    kind: activeKind.value,
                    slug: state.draft.slug.trim(),
                    name: state.draft.name.trim() || state.draft.slug.trim()
                  })

                  const refreshed = await listProblemLabelRecords(activeKind.value)
                  if (activeKind.value === 'tag') {
                    state.tags = refreshed.labels
                  } else {
                    state.patterns = refreshed.labels
                  }
                  state.draft = { ...EMPTY_DRAFT_LABEL }
                } catch (error) {
                  state.error = formatError(error)
                } finally {
                  state.saving = false
                }
              }}
            >
              Add label
            </Button>
          </div>
        </div>

        {state.loading ? (
          <LoadingBlock label="Loading label records..." />
        ) : records.length === 0 ? (
          <EmptyBlock
            title={`No ${activeKind.value}s yet.`}
            copy="Create the first entry and it will become available across the rest of the UI."
          />
        ) : (
          <div class="overflow-x-auto">
            <Table hoverable>
              <Table.Head>
                <Table.HeadCell>Slug</Table.HeadCell>
                <Table.HeadCell>Name</Table.HeadCell>
                <Table.HeadCell>Updated</Table.HeadCell>
                <Table.HeadCell>Actions</Table.HeadCell>
              </Table.Head>
              <Table.Body class="divide-y">
                {records.map((label) => {
                  const editing = state.editing[label.id]
                  const isEditing = Boolean(editing)

                  return (
                    <Table.Row key={label.id}>
                      <Table.Cell>
                        {isEditing ? (
                          <Input
                            value={editing.slug}
                            onInput$={(_, el) => {
                              state.editing = {
                                ...state.editing,
                                [label.id]: {
                                  ...(state.editing[label.id] ?? editing),
                                  slug: el.value
                                }
                              }
                            }}
                          />
                        ) : (
                          label.slug
                        )}
                      </Table.Cell>
                      <Table.Cell>
                        {isEditing ? (
                          <Input
                            value={editing.name}
                            onInput$={(_, el) => {
                              state.editing = {
                                ...state.editing,
                                [label.id]: {
                                  ...(state.editing[label.id] ?? editing),
                                  name: el.value
                                }
                              }
                            }}
                          />
                        ) : (
                          label.name
                        )}
                      </Table.Cell>
                      <Table.Cell>{formatDate(label.updated_at)}</Table.Cell>
                      <Table.Cell>
                        <div class="flex flex-wrap gap-2">
                          {isEditing ? (
                            <>
                              <Button
                                pill
                                size="xs"
                                color="green"
                                onClick$={async () => {
                                  state.saving = true
                                  state.error = ''
                                  try {
                                    await updateProblemLabel(label.id, {
                                      slug: editing.slug.trim(),
                                      name:
                                        editing.name.trim() || editing.slug.trim()
                                    })

                                    const refreshed = await listProblemLabelRecords(
                                      activeKind.value
                                    )
                                    if (activeKind.value === 'tag') {
                                      state.tags = refreshed.labels
                                    } else {
                                      state.patterns = refreshed.labels
                                    }

                                    const next = { ...state.editing }
                                    delete next[label.id]
                                    state.editing = next
                                  } catch (error) {
                                    state.error = formatError(error)
                                  } finally {
                                    state.saving = false
                                  }
                                }}
                              >
                                Save
                              </Button>
                              <Button
                                pill
                                size="xs"
                                color="alternative"
                                onClick$={() => {
                                  const next = { ...state.editing }
                                  delete next[label.id]
                                  state.editing = next
                                }}
                              >
                                Cancel
                              </Button>
                            </>
                          ) : (
                            <Button
                              pill
                              size="xs"
                              color="alternative"
                              onClick$={() => {
                                state.editing = {
                                  ...state.editing,
                                  [label.id]: {
                                    slug: label.slug,
                                    name: label.name
                                  }
                                }
                              }}
                            >
                              Edit
                            </Button>
                          )}
                          <Button
                            pill
                            size="xs"
                            color="red"
                            onClick$={async () => {
                              state.saving = true
                              state.error = ''
                              try {
                                await deleteProblemLabel(label.id)
                                const refreshed = await listProblemLabelRecords(
                                  activeKind.value
                                )
                                if (activeKind.value === 'tag') {
                                  state.tags = refreshed.labels
                                } else {
                                  state.patterns = refreshed.labels
                                }
                              } catch (error) {
                                state.error = formatError(error)
                              } finally {
                                state.saving = false
                              }
                            }}
                          >
                            Delete
                          </Button>
                        </div>
                      </Table.Cell>
                    </Table.Row>
                  )
                })}
              </Table.Body>
            </Table>
          </div>
        )}
      </Card>

      <div class="flex justify-end">
        <Button
          pill
          color="alternative"
          onClick$={() => props.navigate('/')}
        >
          Back to problems
        </Button>
      </div>
    </PageShell>
  )
})

const NotFoundPage = component$(
  (props: { navigate: NavigateFn; path: string }) => (
    <PageShell>
      <Card class="space-y-6">
        <SectionLead
          eyebrow="Route not found"
          title="That page does not exist."
          copy={`The current path "${props.path}" is not wired into the Qwik app.`}
        />
        <div class="flex justify-start">
          <Button pill onClick$={() => props.navigate('/')}>
            Go back to problems
          </Button>
        </div>
      </Card>
    </PageShell>
  )
)

const PageShell = component$(() => (
  <div class="space-y-6">
    <Slot />
  </div>
))

const SectionLead = component$(
  (props: { eyebrow: string; title: string; copy: string }) => (
    <div class="space-y-3">
      <p class="text-xs font-semibold uppercase tracking-[0.22em] text-blue-600">
        {props.eyebrow}
      </p>
      <h1 class="text-3xl font-semibold tracking-tight text-gray-900">
        {props.title}
      </h1>
      <p class="max-w-3xl text-sm text-gray-500">{props.copy}</p>
    </div>
  )
)

const SectionTitle = component$(
  (props: { title: string; subtitle?: string }) => (
    <div class="space-y-1">
      <h2 class="text-xl font-semibold tracking-tight text-gray-900">
        {props.title}
      </h2>
      {props.subtitle && <p class="text-sm text-gray-500">{props.subtitle}</p>}
    </div>
  )
)

const LabelButtonRow = component$(
  (props: {
    title: string
    values: string[]
    selected: string[]
    loading: boolean
    activeColor: 'blue' | 'purple'
    onToggle$: QRL<(value: string) => void>
  }) => (
    <div class="space-y-3">
      <div class="text-sm font-medium text-gray-700">{props.title}</div>
      {props.loading ? (
        <LoadingInline label={`Loading ${props.title.toLowerCase()}...`} />
      ) : props.values.length === 0 ? (
        <p class="text-sm text-gray-500">No {props.title.toLowerCase()} yet.</p>
      ) : (
        <div class="flex flex-wrap gap-2">
          {props.values.map((value) => (
            <Button
              key={value}
              pill
              size="xs"
              color={
                props.selected.includes(value) ? props.activeColor : 'alternative'
              }
              onClick$={() => props.onToggle$(value)}
            >
              {value}
            </Button>
          ))}
        </div>
      )}
    </div>
  )
)

const DifficultyBadge = component$((props: { difficulty: Difficulty }) => (
  <Badge
    content={props.difficulty}
    type={difficultyBadgeType(props.difficulty)}
  />
))

const StatusBadge = component$(
  (props: { status: SubmissionStatus; label?: string }) => (
    <Badge
      content={props.label ?? props.status.replace(/_/g, ' ')}
      type={statusBadgeType(props.status)}
    />
  )
)

const StatCard = component$((props: { label: string; value: string }) => (
  <Card class="space-y-2">
    <p class="text-sm font-medium uppercase tracking-wide text-gray-500">
      {props.label}
    </p>
    <p class="text-3xl font-semibold tracking-tight text-gray-900">
      {props.value}
    </p>
  </Card>
))

const LoadingInline = component$((props: { label: string }) => (
  <div class="flex items-center gap-3 text-sm text-gray-500">
    <Spinner size="5" />
    <span>{props.label}</span>
  </div>
))

const LoadingBlock = component$((props: { label: string }) => (
  <div class="flex min-h-40 items-center justify-center">
    <LoadingInline label={props.label} />
  </div>
))

const EmptyBlock = component$((props: { title: string; copy: string }) => (
  <div class="rounded-xl border border-dashed border-gray-200 bg-gray-50 px-6 py-10 text-center">
    <h3 class="text-lg font-semibold text-gray-900">{props.title}</h3>
    <p class="mt-2 text-sm text-gray-500">{props.copy}</p>
  </div>
))

const ErrorAlert = component$(() => (
  <Alert color="failure" rounded withBorderAccent>
    <Slot />
  </Alert>
))

const WarningAlert = component$(() => (
  <Alert color="warning" rounded withBorderAccent>
    <Slot />
  </Alert>
))

const InfoAlert = component$(() => (
  <Alert color="info" rounded withBorderAccent>
    <Slot />
  </Alert>
))
