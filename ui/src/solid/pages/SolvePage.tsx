import { For, Show, createEffect, createMemo, createSignal, onCleanup } from 'solid-js'
import { createStore } from 'solid-js/store'
import {
  createSubmission,
  getProblem,
  getProblemTestCases,
  getSubmission,
  listSubmissions
} from '@/api/client'
import type { Language, Problem, Submission } from '@/api/types'
import {
  EmptyBlock,
  ErrorAlert,
  InfoAlert,
  LoadingBlock,
  LoadingInline,
  WarningAlert
} from '../components/common/Feedback'
import { PageShell } from '../components/common/PageShell'
import { Badge, Button, Card, CodeBlock, CodeEditor, SelectField, Tabs } from '../components/common'
import { DifficultyBadge, StatusBadge } from '../components/problems/Badges'
import { DEFAULT_CODE, LANGUAGE_OPTIONS } from '../shared/constants'
import type { NavigateFn, SolveTestCase } from '../shared/types'
import {
  classes,
  formatElapsed,
  formatError,
  formatDateTime,
  isPending,
  resolveStarterCode
} from '../shared/utils'

function preferredLanguage(problem: Problem): Language {
  return (
    LANGUAGE_OPTIONS.find(({ value }) => problem.starter_code[value]?.trim())?.value ?? 'python'
  )
}

export function SolvePage(props: { navigate: NavigateFn; slug: string }) {
  const [state, setState] = createStore({
    loadingProblem: true,
    problemError: '',
    actionError: '',
    problem: null as Problem | null,
    submissions: [] as Submission[],
    visibleTestCases: [] as SolveTestCase[],
    hiddenTestCaseCount: 0,
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

  const [language, setLanguage] = createSignal<Language>('python')

  createEffect(() => {
    const slug = props.slug
    let active = true

    setState({
      loadingProblem: true,
      problemError: '',
      actionError: '',
      problem: null,
      submissions: [],
      visibleTestCases: [],
      hiddenTestCaseCount: 0,
      verdict: null,
      selectedSubmission: null,
      selectedSubmissionId: '',
      codeByLanguage: { ...DEFAULT_CODE },
      tab: 'description',
      submitting: false,
      pollingId: '',
      timerActive: false,
      startedAt: 0,
      elapsed: 0
    })
    setLanguage('python')

    void (async () => {
      try {
        const problem = await getProblem(slug)
        const history = await listSubmissions(problem.id)
        let testCases: SolveTestCase[] = []

        try {
          testCases = (await getProblemTestCases(problem.id)).test_cases
        } catch (error) {
          if (active) {
            setState('actionError', formatError(error))
          }
        }

        const visibleTestCases = testCases.filter((testCase) => !testCase.is_hidden)
        const initialLanguage = preferredLanguage(problem)

        if (active) {
          setState({
            problem,
            codeByLanguage: resolveStarterCode(problem.starter_code),
            submissions: history.submissions,
            visibleTestCases,
            hiddenTestCaseCount: testCases.length - visibleTestCases.length,
            selectedSubmissionId: history.submissions[0]?.id ?? ''
          })
          setLanguage(initialLanguage)
        }
      } catch (error) {
        if (active) {
          setState('problemError', formatError(error))
        }
      } finally {
        if (active) {
          setState('loadingProblem', false)
        }
      }
    })()

    onCleanup(() => {
      active = false
    })
  })

  createEffect(() => {
    const id = state.selectedSubmissionId
    if (!id) {
      setState('selectedSubmission', null)
      return
    }

    let active = true
    setState('selectedLoading', true)

    void (async () => {
      try {
        const submission = await getSubmission(id)
        if (active) {
          setState({
            selectedSubmission: submission,
            actionError: ''
          })
        }
      } catch (error) {
        if (active) {
          setState('actionError', formatError(error))
        }
      } finally {
        if (active) {
          setState('selectedLoading', false)
        }
      }
    })()

    onCleanup(() => {
      active = false
    })
  })

  createEffect(() => {
    const id = state.pollingId
    if (!id) return

    let active = true

    const load = async () => {
      try {
        const submission = await getSubmission(id)
        if (!active) return

        setState({
          verdict: submission,
          selectedSubmission: submission,
          selectedSubmissionId: submission.id
        })

        if (!isPending(submission.status)) {
          setState('pollingId', '')
          if (state.problem?.id) {
            const history = await listSubmissions(state.problem.id)
            if (active) {
              setState('submissions', history.submissions)
            }
          }
        }
      } catch (error) {
        if (active) {
          setState({
            actionError: formatError(error),
            pollingId: ''
          })
        }
      }
    }

    void load()
    const timer = window.setInterval(() => {
      void load()
    }, 1500)

    onCleanup(() => {
      active = false
      clearInterval(timer)
    })
  })

  createEffect(() => {
    const timerActive = state.timerActive
    const startedAt = state.startedAt
    if (!timerActive || !startedAt) return

    setState('elapsed', Math.floor((Date.now() - startedAt) / 1000))

    const timer = window.setInterval(() => {
      setState('elapsed', Math.floor((Date.now() - startedAt) / 1000))
    }, 1000)

    onCleanup(() => {
      clearInterval(timer)
    })
  })

  const focusedSubmission = createMemo(
    () =>
      state.selectedSubmission ??
      (state.verdict && state.verdict.id === state.selectedSubmissionId ? state.verdict : null) ??
      state.verdict
  )
  const completedSubmission = createMemo(() => {
    const submission = focusedSubmission()
    return submission && !isPending(submission.status) ? submission : null
  })

  return (
    <PageShell>
      <Card class="space-y-4">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-3">
            <div class="text-xs uppercase tracking-[0.2em] text-gray-500">Live run</div>
            <div class="flex flex-wrap items-center gap-2">
              <h1 class="text-3xl font-semibold tracking-tight text-gray-900">
                {state.problem?.title ?? 'Loading problem'}
              </h1>
              <Show when={state.problem}>
                {(problem) => <DifficultyBadge difficulty={problem().difficulty} />}
              </Show>
            </div>
            <p class="max-w-3xl text-sm text-gray-500">
              This solve screen now runs in Solid while keeping the same backend interaction model.
            </p>
          </div>

          <div class="flex flex-wrap gap-2">
            <Button pill color="alternative" onClick={() => props.navigate('/')}>
              Back to board
            </Button>
            <Show when={state.problem?.source_url}>
              <Button pill color="light" href={state.problem?.source_url} target="_blank">
                Open source
              </Button>
            </Show>
          </div>
        </div>
      </Card>

      <Show when={state.problemError}>
        <ErrorAlert>{state.problemError}</ErrorAlert>
      </Show>
      <Show when={state.actionError}>
        <WarningAlert>{state.actionError}</WarningAlert>
      </Show>

      <Show
        when={!state.loadingProblem}
        fallback={
          <Card>
            <LoadingBlock label="Loading problem and submission history..." />
          </Card>
        }
      >
        <Show when={state.problem}>
          {(problem) => (
            <div class="grid gap-6 xl:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
              <Card class="space-y-5">
                <div class="border-b border-gray-100 pb-4">
                  <Tabs
                    value={state.tab}
                    items={[
                      { value: 'description', label: 'Description' },
                      {
                        value: 'submissions',
                        label: `Submissions (${state.submissions.length})`
                      }
                    ]}
                    onChange={(value) => setState('tab', value as 'description' | 'submissions')}
                  />
                </div>

                <Show
                  when={state.tab === 'description'}
                  fallback={
                    <div class="space-y-4">
                      <Show when={state.submitting || state.pollingId}>
                        <InfoAlert>Submission is being evaluated.</InfoAlert>
                      </Show>

                      <Show when={completedSubmission()}>
                        {(submission) => (
                          <Card
                            class={classes(
                              'space-y-4 border',
                              submission().status === 'accepted'
                                ? 'border-green-200 bg-green-50'
                                : 'border-red-200 bg-red-50'
                            )}
                          >
                            <div class="flex flex-wrap items-center justify-between gap-2">
                              <StatusBadge
                                status={submission().status}
                                label={submission().verdict || undefined}
                              />
                              <span class="text-sm text-gray-500">
                                {formatDateTime(submission().submitted_at)}
                              </span>
                            </div>
                            <div class="flex flex-wrap gap-3 text-sm text-gray-600">
                              <span>
                                Passed {submission().passed_cases}/{submission().total_cases}
                              </span>
                              <span>
                                {submission().runtime_ms > 0
                                  ? `${submission().runtime_ms} ms`
                                  : 'pending runtime'}
                              </span>
                            </div>
                            <Show when={submission().error_message}>
                              <CodeBlock label="Error output" code={submission().error_message} />
                            </Show>
                            <CodeBlock label="Submitted code" code={submission().code} />
                          </Card>
                        )}
                      </Show>

                      <Show when={state.selectedLoading}>
                        <LoadingInline label="Loading submission details..." />
                      </Show>

                      <Show
                        when={state.submissions.length > 0}
                        fallback={
                          <EmptyBlock
                            title="No submissions yet."
                            copy="Ship one from the editor and the history feed will populate here."
                          />
                        }
                      >
                        <div class="space-y-2">
                          <For each={state.submissions}>
                            {(submission) => (
                              <button
                                class={classes(
                                  'flex w-full items-center justify-between rounded-2xl border px-4 py-3 text-left transition',
                                  state.selectedSubmissionId === submission.id
                                    ? 'border-blue-200 bg-blue-50'
                                    : 'border-gray-200 bg-white hover:border-gray-300'
                                )}
                                onClick={() => {
                                  setState({
                                    selectedSubmissionId: submission.id,
                                    tab: 'submissions'
                                  })
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
                            )}
                          </For>
                        </div>
                      </Show>
                    </div>
                  }
                >
                  <div class="space-y-4">
                    <div class="flex flex-wrap gap-2">
                      <For each={problem().tags}>
                        {(tag) => <Badge content={tag} color="blue" />}
                      </For>
                    </div>

                    <div class="grid gap-3 rounded-2xl bg-gray-50 p-4 text-sm text-gray-500 md:grid-cols-3">
                      <span>provider: {problem().provider}</span>
                      <span>slug: {problem().slug}</span>
                      <span>estimate: {problem().estimated_time} min</span>
                    </div>

                    <Card class="space-y-4 border">
                      <div class="space-y-2">
                        <h3 class="text-lg font-semibold text-gray-900">Problem details</h3>
                        <p class="text-sm text-gray-500">
                          Full statements are not stored in the local bank yet. Use the provider
                          prompt for the full description and solve from the examples below.
                        </p>
                      </div>

                      <div class="flex flex-wrap items-center gap-2">
                        <Show
                          when={problem().source_url}
                          fallback={<Badge content="No source link saved" color="dark" />}
                        >
                          <Button
                            pill
                            size="sm"
                            color="blue"
                            href={problem().source_url}
                            target="_blank"
                          >
                            Read full prompt
                          </Button>
                        </Show>
                        <Show when={state.hiddenTestCaseCount > 0}>
                          <Badge
                            content={`${state.hiddenTestCaseCount} hidden judge case${
                              state.hiddenTestCaseCount === 1 ? '' : 's'
                            }`}
                            color="dark"
                          />
                        </Show>
                      </div>

                      <Show
                        when={state.visibleTestCases.length > 0}
                        fallback={
                          <div class="rounded-2xl border border-dashed border-gray-200 bg-gray-50 p-4 text-sm text-gray-600">
                            No visible example cases are stored for this problem yet.
                          </div>
                        }
                      >
                        <div class="space-y-3">
                          <div class="text-sm font-medium text-gray-700">Visible examples</div>
                          <div class="grid gap-3">
                            <For each={state.visibleTestCases}>
                              {(testCase, index) => (
                                <Card class="space-y-3 border">
                                  <div class="text-sm font-semibold text-gray-900">
                                    Example {index() + 1}
                                  </div>
                                  <div class="grid gap-3 lg:grid-cols-2">
                                    <div class="space-y-2">
                                      <CodeBlock label="Input" code={testCase.input} />
                                    </div>
                                    <div class="space-y-2">
                                      <CodeBlock label="Expected output" code={testCase.expected} />
                                    </div>
                                  </div>
                                </Card>
                              )}
                            </For>
                          </div>
                        </div>
                      </Show>
                    </Card>
                  </div>
                </Show>
              </Card>

              <Card class="space-y-5">
                <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
                  <div class="grid gap-4 sm:grid-cols-[minmax(0,220px)_auto] sm:items-end">
                    <SelectField
                      label="Language"
                      value={language()}
                      options={LANGUAGE_OPTIONS}
                      onChange={(event) => setLanguage(event.currentTarget.value as Language)}
                    />
                    <Badge content={`timer ${formatElapsed(state.elapsed)}`} color="dark" />
                  </div>

                  <Button
                    pill
                    loading={state.submitting}
                    disabled={!state.problem}
                    onClick={async () => {
                      if (!state.problem) return

                      setState({
                        submitting: true,
                        actionError: ''
                      })

                      try {
                        const selectedLanguage = language()
                        const result = await createSubmission({
                          problem_id: state.problem.id,
                          language: selectedLanguage,
                          code: state.codeByLanguage[selectedLanguage]
                        })

                        setState({
                          tab: 'submissions',
                          timerActive: false,
                          startedAt: 0,
                          elapsed: 0,
                          pollingId: result.submission_id,
                          selectedSubmissionId: result.submission_id
                        })
                      } catch (error) {
                        setState('actionError', formatError(error))
                      } finally {
                        setState('submitting', false)
                      }
                    }}
                  >
                    Submit
                  </Button>
                </div>

                <CodeEditor
                  label="Code"
                  language={language()}
                  rows={24}
                  value={state.codeByLanguage[language()]}
                  onInput={(value) => {
                    const selectedLanguage = language()
                    if (!state.timerActive) {
                      setState({
                        timerActive: true,
                        startedAt: Date.now(),
                        elapsed: 0
                      })
                    }

                    setState('codeByLanguage', selectedLanguage, value)
                  }}
                />
              </Card>
            </div>
          )}
        </Show>
      </Show>
    </PageShell>
  )
}
