import { For, Show, createEffect, createSignal, onCleanup, onMount } from 'solid-js'
import { createStore } from 'solid-js/store'
import {
  contributeProblem,
  getProblem,
  getProblemTestCases,
  listProblemLabels,
  updateProblem
} from '@/api/client'
import type { Difficulty, Provider } from '@/api/types'
import { ErrorAlert, LoadingBlock, WarningAlert } from '../components/common/Feedback'
import { PageShell } from '../components/common/PageShell'
import {
  Button,
  CheckboxField,
  Card,
  InputField,
  SelectField,
  TextareaField
} from '../components/common'
import { SectionLead } from '../components/common/SectionLead'
import { SectionTitle } from '../components/common/SectionTitle'
import { LabelButtonRow } from '../components/problems/LabelButtonRow'
import {
  DEFAULT_STARTER_CODE,
  DIFFICULTY_OPTIONS,
  EMPTY_LABELS,
  EMPTY_TEST_CASE,
  PROVIDER_OPTIONS
} from '../shared/constants'
import type { DraftTestCase, NavigateFn } from '../shared/types'
import { formatError, resolveStarterCode, toggleValue } from '../shared/utils'

export function ContributeProblemPage(props: { navigate: NavigateFn; slug?: string }) {
  const isEditMode = () => Boolean(props.slug)
  const [state, setState] = createStore({
    loading: isEditMode(),
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

  const [provider, setProvider] = createSignal<Provider>('leetcode')
  const [difficulty, setDifficulty] = createSignal<Difficulty>('easy')

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
          setState('labelsError', formatError(error))
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
    const slug = props.slug ?? ''
    if (!slug) {
      setState('loading', false)
      return
    }

    let active = true
    setState({
      loading: true,
      error: ''
    })

    void (async () => {
      try {
        const problem = await getProblem(slug)
        const testCases = await getProblemTestCases(problem.id)

        if (active) {
          setState({
            problemID: problem.id,
            external_id: problem.external_id,
            slug: problem.slug,
            title: problem.title,
            tags: [...problem.tags],
            pattern_tags: [...problem.pattern_tags],
            source_url: problem.source_url,
            estimated_time: problem.estimated_time,
            starter_code: resolveStarterCode(problem.starter_code, DEFAULT_STARTER_CODE),
            test_cases:
              testCases.test_cases.length > 0
                ? testCases.test_cases.map((testCase) => ({
                    input: testCase.input,
                    expected: testCase.expected,
                    is_hidden: Boolean(testCase.is_hidden)
                  }))
                : [{ ...EMPTY_TEST_CASE }]
          })
          setProvider(problem.provider)
          setDifficulty(problem.difficulty)
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

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Registry editor"
          title={isEditMode() ? 'Update a stored problem.' : 'Add a new problem.'}
          copy="This form is now rendered in Solid with Flowbite-styled controls."
        />
      </Card>

      <Show when={state.error}>
        <ErrorAlert>{state.error}</ErrorAlert>
      </Show>
      <Show when={state.labelsError}>
        <WarningAlert>{state.labelsError}</WarningAlert>
      </Show>

      <Show
        when={!state.loading}
        fallback={
          <Card>
            <LoadingBlock label="Loading problem metadata..." />
          </Card>
        }
      >
        <div class="space-y-6">
          <Card class="space-y-6">
            <SectionTitle
              title="Problem metadata"
              subtitle="Provider, difficulty, tags, and source metadata."
            />

            <div class="grid gap-4 md:grid-cols-3">
              <SelectField
                label="Provider"
                value={provider()}
                options={PROVIDER_OPTIONS}
                onChange={(event) => setProvider(event.currentTarget.value as Provider)}
              />
              <InputField
                label="External ID"
                value={state.external_id}
                onInput={(event) => setState('external_id', event.currentTarget.value)}
              />
              <SelectField
                label="Difficulty"
                value={difficulty()}
                options={DIFFICULTY_OPTIONS.slice(1)}
                onChange={(event) => setDifficulty(event.currentTarget.value as Difficulty)}
              />
            </div>

            <div class="grid gap-4 md:grid-cols-2">
              <InputField
                label="Slug"
                value={state.slug}
                onInput={(event) => setState('slug', event.currentTarget.value)}
              />
              <InputField
                label="Title"
                value={state.title}
                onInput={(event) => setState('title', event.currentTarget.value)}
              />
            </div>

            <InputField
              label="Source URL"
              value={state.source_url}
              onInput={(event) => setState('source_url', event.currentTarget.value)}
            />

            <div class="grid gap-4 md:grid-cols-2">
              <InputField
                label="Estimated time (minutes)"
                type="number"
                value={String(state.estimated_time)}
                onInput={(event) =>
                  setState('estimated_time', Number(event.currentTarget.value) || 0)
                }
              />
              <Show when={!isEditMode()}>
                <InputField
                  label="Version"
                  type="number"
                  value={String(state.version)}
                  onInput={(event) =>
                    setState('version', Math.max(1, Number(event.currentTarget.value) || 1))
                  }
                />
              </Show>
            </div>

            <LabelButtonRow
              title="Tags"
              helperText="Select every tag that applies to this problem."
              values={state.labels.tags}
              selected={state.tags}
              loading={state.labelsLoading}
              activeColor="blue"
              onToggle={(value) => setState('tags', toggleValue(state.tags, value))}
              onClear={() => setState('tags', [])}
            />

            <LabelButtonRow
              title="Pattern tags"
              helperText="Pattern tags are multi-select too."
              values={state.labels.patterns}
              selected={state.pattern_tags}
              loading={state.labelsLoading}
              activeColor="indigo"
              onToggle={(value) => setState('pattern_tags', toggleValue(state.pattern_tags, value))}
              onClear={() => setState('pattern_tags', [])}
            />
          </Card>

          <Card class="space-y-6">
            <SectionTitle
              title="Starter code"
              subtitle="Language drafts are saved with the registry entry."
            />

            <div class="grid gap-4 xl:grid-cols-2">
              <TextareaField
                label="Python"
                rows={12}
                value={state.starter_code.python}
                onInput={(event) => setState('starter_code', 'python', event.currentTarget.value)}
              />
              <TextareaField
                label="Go"
                rows={12}
                value={state.starter_code.go}
                onInput={(event) => setState('starter_code', 'go', event.currentTarget.value)}
              />
            </div>
          </Card>

          <Card class="space-y-6">
            <SectionTitle
              title="Judge test cases"
              subtitle="Add visible or hidden cases before shipping the problem."
            />

            <div class="space-y-4">
              <For each={state.test_cases}>
                {(testCase, index) => (
                  <Card class="space-y-4 border">
                    <div class="flex flex-wrap items-center justify-between gap-2">
                      <h3 class="text-base font-semibold text-gray-900">Case {index() + 1}</h3>
                      <Button
                        pill
                        color="light"
                        disabled={state.test_cases.length === 1}
                        onClick={() => {
                          if (state.test_cases.length === 1) return
                          setState(
                            'test_cases',
                            state.test_cases.filter((_, itemIndex) => itemIndex !== index())
                          )
                        }}
                      >
                        Remove
                      </Button>
                    </div>

                    <div class="grid gap-4 xl:grid-cols-2">
                      <TextareaField
                        label="Input"
                        rows={8}
                        value={testCase.input}
                        onInput={(event) =>
                          setState('test_cases', index(), 'input', event.currentTarget.value)
                        }
                      />
                      <TextareaField
                        label="Expected output"
                        rows={8}
                        value={testCase.expected}
                        onInput={(event) =>
                          setState('test_cases', index(), 'expected', event.currentTarget.value)
                        }
                      />
                    </div>

                    <CheckboxField
                      label="Hidden test case"
                      checked={testCase.is_hidden}
                      onChange={(event) =>
                        setState('test_cases', index(), 'is_hidden', event.currentTarget.checked)
                      }
                    />
                  </Card>
                )}
              </For>

              <div class="flex flex-wrap justify-between gap-3">
                <Button
                  pill
                  color="alternative"
                  onClick={() =>
                    setState('test_cases', [...state.test_cases, { ...EMPTY_TEST_CASE }])
                  }
                >
                  Add test case
                </Button>

                <Button
                  pill
                  loading={state.saving}
                  disabled={!state.slug.trim() || !state.title.trim()}
                  onClick={async () => {
                    setState({ saving: true, error: '' })

                    try {
                      const payload = {
                        provider: provider(),
                        external_id: state.external_id,
                        slug: state.slug.trim(),
                        title: state.title.trim(),
                        difficulty: difficulty(),
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
                      setState('error', formatError(error))
                    } finally {
                      setState('saving', false)
                    }
                  }}
                >
                  {isEditMode() ? 'Update problem' : 'Save problem'}
                </Button>
              </div>
            </div>
          </Card>
        </div>
      </Show>
    </PageShell>
  )
}
