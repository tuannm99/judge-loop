import { useEffect, useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Panel,
  Group as PanelGroup,
  Separator as PanelResizeHandle
} from 'react-resizable-panels'
import Editor from '@monaco-editor/react'
import {
  getProblem,
  createSubmission,
  getSubmission,
  listSubmissions
} from '@/api/client'
import type { Language } from '@/api/types'
import { DifficultyBadge, StatusBadge } from '@/components/ui'
import {
  Button,
  NativeSelect,
  Tabs,
  Badge,
  Text,
  Group,
  Stack,
  Loader,
  Paper,
  Code,
  Anchor
} from '@mantine/core'
import {
  IconArrowLeft,
  IconClock,
  IconCircleCheck,
  IconCircleX
} from '@tabler/icons-react'

const DEFAULT_CODE: Record<Language, string> = {
  python: '# Write your solution here\n\n',
  go: 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println()\n}\n'
}

function resolveStarterCode(
  starterCode: Partial<Record<Language, string>> | undefined
) {
  return {
    python: starterCode?.python || DEFAULT_CODE.python,
    go: starterCode?.go || DEFAULT_CODE.go
  } satisfies Record<Language, string>
}

function formatElapsed(s: number) {
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${m}:${String(sec).padStart(2, '0')}`
}

function useElapsedTimer(problemId: string | undefined) {
  const [elapsed, setElapsed] = useState(0)
  const startedAt = useRef<number | null>(null)
  const tick = useRef<ReturnType<typeof setInterval> | null>(null)
  const active = useRef(false)

  function clearTick() {
    if (tick.current) {
      clearInterval(tick.current)
      tick.current = null
    }
  }

  async function begin() {
    if (!problemId || active.current) return
    active.current = true
    startedAt.current = Date.now()
    setElapsed(0)
    clearTick()
    tick.current = setInterval(() => {
      if (startedAt.current == null) return
      setElapsed(Math.floor((Date.now() - startedAt.current) / 1000))
    }, 1000)
  }

  async function end() {
    if (!active.current) return

    clearTick()
    active.current = false
    setElapsed(0)
    startedAt.current = null
  }

  useEffect(
    () => () => {
      clearTick()
      active.current = false
    },
    []
  )

  return { elapsed, begin, end, isActive: active.current }
}

export function Solve() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const [language, setLanguage] = useState<Language>('python')
  const [codeByLanguage, setCodeByLanguage] =
    useState<Record<Language, string>>(DEFAULT_CODE)
  const [pollingId, setPollingId] = useState<string | null>(null)
  const [selectedSubmissionId, setSelectedSubmissionId] = useState<
    string | null
  >(null)
  const [tab, setTab] = useState<string>('description')

  const { data: problem } = useQuery({
    queryKey: ['problem', slug],
    queryFn: () => getProblem(slug!),
    enabled: !!slug
  })

  const { data: history } = useQuery({
    queryKey: ['submissions', problem?.id],
    queryFn: () => listSubmissions(problem!.id),
    enabled: !!problem?.id
  })

  const {
    elapsed,
    begin: beginTimer,
    end: endTimer,
    isActive: timerActive
  } = useElapsedTimer(problem?.id)
  const code = codeByLanguage[language]

  const { data: verdict, isFetching: polling } = useQuery({
    queryKey: ['submission-verdict', pollingId],
    queryFn: () => getSubmission(pollingId!),
    enabled: !!pollingId,
    refetchInterval: (query) => {
      const s = query.state.data?.status
      return !s || s === 'pending' || s === 'running' ? 1500 : false
    }
  })

  const { data: selectedSubmission, isFetching: loadingSelectedSubmission } =
    useQuery({
      queryKey: ['submission-selected', selectedSubmissionId],
      queryFn: () => getSubmission(selectedSubmissionId!),
      enabled: !!selectedSubmissionId
    })

  const { mutate: submit, isPending: submitting } = useMutation({
    mutationFn: () =>
      createSubmission({ problem_id: problem!.id, language, code }),
    onSuccess: async (res) => {
      await endTimer()
      setPollingId(res.submission_id)
      setSelectedSubmissionId(res.submission_id)
      setTab('submissions')
    }
  })

  useEffect(() => {
    if (
      verdict &&
      verdict.status !== 'pending' &&
      verdict.status !== 'running'
    ) {
      void qc.invalidateQueries({ queryKey: ['submissions', problem?.id] })
      void qc.invalidateQueries({ queryKey: ['progress-today'] })
    }
  }, [verdict, problem?.id, qc])

  useEffect(() => {
    if (!history?.submissions?.length) return
    setSelectedSubmissionId((current) => current ?? history.submissions[0].id)
  }, [history])

  useEffect(() => {
    if (!problem) return
    setCodeByLanguage(resolveStarterCode(problem.starter_code))
  }, [problem])

  const isLoading = submitting || polling
  const reviewedSubmission = selectedSubmission ?? verdict
  const accepted = reviewedSubmission?.status === 'accepted'

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: 'calc(100vh - 53px)'
      }}
    >
      {/* Top bar */}
      <Paper
        withBorder
        radius={0}
        px="md"
        py="xs"
        style={{ borderTop: 0, borderLeft: 0, borderRight: 0 }}
      >
        <Group justify="space-between">
          <Group gap="sm">
            <Button
              variant="subtle"
              size="xs"
              px={4}
              onClick={() => navigate('/')}
              leftSection={<IconArrowLeft size={14} />}
            >
              Back
            </Button>
            {problem && (
              <>
                <Text size="sm" fw={500}>
                  {problem.title}
                </Text>
                <DifficultyBadge difficulty={problem.difficulty} />
              </>
            )}
          </Group>

          <Group gap="sm">
            <Group gap={4}>
              <IconClock size={12} />
              <Text size="xs" c="dimmed">
                {formatElapsed(elapsed)}
              </Text>
            </Group>

            <NativeSelect
              size="xs"
              value={language}
              onChange={(e) => {
                const lang = e.target.value as Language
                setLanguage(lang)
              }}
              data={[
                { value: 'python', label: 'Python' },
                { value: 'go', label: 'Go' }
              ]}
            />

            <Button
              size="xs"
              loading={isLoading}
              disabled={!problem}
              onClick={() => submit()}
            >
              Submit
            </Button>
          </Group>
        </Group>
      </Paper>

      {/* Split panes */}
      <PanelGroup
        orientation="horizontal"
        style={{ flex: 1, overflow: 'hidden' }}
      >
        <Panel
          defaultSize={40}
          minSize={25}
          style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden' }}
        >
          <Tabs
            value={tab}
            onChange={(v) => setTab(v ?? 'description')}
            style={{ display: 'flex', flexDirection: 'column', height: '100%' }}
          >
            <Tabs.List px="sm" pt={4}>
              <Tabs.Tab value="description">Description</Tabs.Tab>
              <Tabs.Tab value="submissions">
                Submissions
                {history?.submissions?.length ? (
                  <Badge size="xs" ml={6} variant="default">
                    {history.submissions.length}
                  </Badge>
                ) : null}
              </Tabs.Tab>
            </Tabs.List>

            <Tabs.Panel
              value="description"
              p="md"
              style={{ overflowY: 'auto', flex: 1 }}
            >
              {problem ? (
                <Stack gap="sm">
                  <Group gap={6}>
                    {problem.tags?.map((t) => (
                      <Badge key={t} size="xs" variant="default">
                        {t}
                      </Badge>
                    ))}
                    {problem.pattern_tags?.map((p) => (
                      <Badge key={p} size="xs" color="violet" variant="light">
                        {p}
                      </Badge>
                    ))}
                  </Group>
                  <Text size="xs" c="dimmed">
                    No description stored — open the original via the{' '}
                    {problem.source_url && (
                      <Anchor
                        href={problem.source_url}
                        target="_blank"
                        size="xs"
                      >
                        source link
                      </Anchor>
                    )}
                    .
                  </Text>
                </Stack>
              ) : (
                <Loader size="sm" />
              )}
            </Tabs.Panel>

            <Tabs.Panel
              value="submissions"
              p="md"
              style={{ overflowY: 'auto', flex: 1 }}
            >
              <Stack gap="sm">
                {/* Latest verdict */}
                {reviewedSubmission &&
                  reviewedSubmission.status !== 'pending' &&
                  reviewedSubmission.status !== 'running' && (
                    <Paper withBorder p="md" bg={accepted ? 'teal.9' : 'red.9'}>
                      <Group gap="xs" mb="xs">
                        {accepted ? (
                          <IconCircleCheck
                            size={16}
                            color="var(--mantine-color-teal-4)"
                          />
                        ) : (
                          <IconCircleX
                            size={16}
                            color="var(--mantine-color-red-4)"
                          />
                        )}
                        <Text
                          size="sm"
                          fw={600}
                          c={accepted ? 'teal.3' : 'red.3'}
                        >
                          {reviewedSubmission.verdict}
                        </Text>
                      </Group>
                      <Text size="xs" c="dimmed">
                        Passed: {reviewedSubmission.passed_cases}/
                        {reviewedSubmission.total_cases}
                        {reviewedSubmission.runtime_ms > 0 &&
                          ` · ${reviewedSubmission.runtime_ms}ms`}
                      </Text>
                      {reviewedSubmission.error_message && (
                        <Code block mt="xs" style={{ fontSize: 11 }}>
                          {reviewedSubmission.error_message}
                        </Code>
                      )}
                      <Code
                        block
                        mt="xs"
                        style={{ fontSize: 11, whiteSpace: 'pre-wrap' }}
                      >
                        {reviewedSubmission.code}
                      </Code>
                    </Paper>
                  )}

                {loadingSelectedSubmission && (
                  <Group gap="xs">
                    <Loader size="xs" />
                    <Text size="xs" c="dimmed">
                      Loading submission details…
                    </Text>
                  </Group>
                )}

                {(submitting || polling) && (
                  <Group gap="xs">
                    <Loader size="xs" />
                    <Text size="xs" c="dimmed">
                      Evaluating…
                    </Text>
                  </Group>
                )}

                {history?.submissions?.map((s) => (
                  <Group
                    key={s.id}
                    justify="space-between"
                    px="xs"
                    py={6}
                    style={{
                      borderBottom: '1px solid var(--mantine-color-dark-4)',
                      cursor: 'pointer',
                      background:
                        selectedSubmissionId === s.id
                          ? 'var(--mantine-color-dark-6)'
                          : undefined
                    }}
                    onClick={() => setSelectedSubmissionId(s.id)}
                  >
                    <StatusBadge
                      status={s.status}
                      label={s.verdict || undefined}
                    />
                    <Text size="xs" c="dimmed">
                      {new Date(s.submitted_at).toLocaleString()}
                    </Text>
                  </Group>
                ))}

                {!submitting && !history?.submissions?.length && (
                  <Text size="xs" c="dimmed">
                    No submissions yet.
                  </Text>
                )}
              </Stack>
            </Tabs.Panel>
          </Tabs>
        </Panel>

        <PanelResizeHandle
          style={{
            width: 4,
            background: 'var(--mantine-color-dark-4)',
            cursor: 'col-resize'
          }}
        />

        <Panel defaultSize={60} minSize={40}>
          <Editor
            height="100%"
            language={language === 'go' ? 'go' : 'python'}
            value={code}
            onChange={(v) => {
              if (!timerActive) {
                void beginTimer()
              }
              setCodeByLanguage((current) => ({
                ...current,
                [language]: v ?? ''
              }))
            }}
            theme="vs-dark"
            options={{
              fontSize: 14,
              minimap: { enabled: false },
              scrollBeyondLastLine: false,
              padding: { top: 12 },
              fontFamily: 'ui-monospace, Consolas, monospace'
            }}
          />
        </Panel>
      </PanelGroup>
    </div>
  )
}
