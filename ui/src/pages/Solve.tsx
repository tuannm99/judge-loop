import { useEffect, useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Panel, Group as PanelGroup, Separator as PanelResizeHandle } from 'react-resizable-panels'
import Editor from '@monaco-editor/react'
import {
  getProblem, createSubmission, getSubmission,
  startTimer, stopTimer, listSubmissions,
} from '@/api/client'
import type { Language } from '@/api/types'
import { DifficultyBadge, StatusBadge } from '@/components/ui'
import {
  Button, NativeSelect, Tabs, Badge, Text, Group, Stack,
  Loader, Paper, Code, Anchor,
} from '@mantine/core'
import {
  IconArrowLeft, IconClock, IconCircleCheck, IconCircleX,
} from '@tabler/icons-react'

const DEFAULT_CODE: Record<Language, string> = {
  python: '# Write your solution here\n\n',
  go: 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println()\n}\n',
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

  useEffect(() => {
    if (!problemId) return
    startTimer(problemId)
      .then(() => {
        startedAt.current = Date.now()
        tick.current = setInterval(
          () => setElapsed(Math.floor((Date.now() - startedAt.current!) / 1000)),
          1000,
        )
      })
      .catch(() => {})
    return () => {
      if (tick.current) clearInterval(tick.current)
      stopTimer().catch(() => {})
    }
  }, [problemId])

  return elapsed
}

export function Solve() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const [language, setLanguage] = useState<Language>('python')
  const [code, setCode] = useState(DEFAULT_CODE['python'])
  const [pollingId, setPollingId] = useState<string | null>(null)
  const [tab, setTab] = useState<string>('description')

  const { data: problem } = useQuery({
    queryKey: ['problem', slug],
    queryFn: () => getProblem(slug!),
    enabled: !!slug,
  })

  const { data: history } = useQuery({
    queryKey: ['submissions', problem?.id],
    queryFn: () => listSubmissions(problem!.id),
    enabled: !!problem?.id,
  })

  const elapsed = useElapsedTimer(problem?.id)

  const { data: verdict, isFetching: polling } = useQuery({
    queryKey: ['submission-verdict', pollingId],
    queryFn: () => getSubmission(pollingId!),
    enabled: !!pollingId,
    refetchInterval: (query) => {
      const s = query.state.data?.status
      return !s || s === 'pending' || s === 'running' ? 1500 : false
    },
  })

  const { mutate: submit, isPending: submitting } = useMutation({
    mutationFn: () => createSubmission({ problem_id: problem!.id, language, code }),
    onSuccess: (res) => {
      setPollingId(res.submission_id)
      setTab('submissions')
    },
  })

  useEffect(() => {
    if (verdict && verdict.status !== 'pending' && verdict.status !== 'running') {
      void qc.invalidateQueries({ queryKey: ['submissions', problem?.id] })
      void qc.invalidateQueries({ queryKey: ['progress-today'] })
    }
  }, [verdict, problem?.id, qc])

  const isLoading = submitting || polling
  const accepted = verdict?.status === 'accepted'

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 53px)' }}>
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
                <Text size="sm" fw={500}>{problem.title}</Text>
                <DifficultyBadge difficulty={problem.difficulty} />
              </>
            )}
          </Group>

          <Group gap="sm">
            <Group gap={4}>
              <IconClock size={12} />
              <Text size="xs" c="dimmed">{formatElapsed(elapsed)}</Text>
            </Group>

            <NativeSelect
              size="xs"
              value={language}
              onChange={(e) => {
                const lang = e.target.value as Language
                setLanguage(lang)
                setCode(DEFAULT_CODE[lang])
              }}
              data={[
                { value: 'python', label: 'Python' },
                { value: 'go', label: 'Go' },
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
      <PanelGroup orientation="horizontal" style={{ flex: 1, overflow: 'hidden' }}>
        <Panel defaultSize={40} minSize={25} style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
          <Tabs value={tab} onChange={(v) => setTab(v ?? 'description')} style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
            <Tabs.List px="sm" pt={4}>
              <Tabs.Tab value="description">Description</Tabs.Tab>
              <Tabs.Tab value="submissions">
                Submissions
                {history?.submissions?.length
                  ? <Badge size="xs" ml={6} variant="default">{history.submissions.length}</Badge>
                  : null}
              </Tabs.Tab>
            </Tabs.List>

            <Tabs.Panel value="description" p="md" style={{ overflowY: 'auto', flex: 1 }}>
              {problem ? (
                <Stack gap="sm">
                  <Group gap={6}>
                    {problem.tags?.map((t) => <Badge key={t} size="xs" variant="default">{t}</Badge>)}
                    {problem.pattern_tags?.map((p) => <Badge key={p} size="xs" color="violet" variant="light">{p}</Badge>)}
                  </Group>
                  <Text size="xs" c="dimmed">
                    No description stored — open the original via the{' '}
                    {problem.source_url && (
                      <Anchor href={problem.source_url} target="_blank" size="xs">source link</Anchor>
                    )}.
                  </Text>
                </Stack>
              ) : (
                <Loader size="sm" />
              )}
            </Tabs.Panel>

            <Tabs.Panel value="submissions" p="md" style={{ overflowY: 'auto', flex: 1 }}>
              <Stack gap="sm">
                {/* Latest verdict */}
                {verdict && verdict.status !== 'pending' && verdict.status !== 'running' && (
                  <Paper withBorder p="md" bg={accepted ? 'teal.9' : 'red.9'}>
                    <Group gap="xs" mb="xs">
                      {accepted
                        ? <IconCircleCheck size={16} color="var(--mantine-color-teal-4)" />
                        : <IconCircleX size={16} color="var(--mantine-color-red-4)" />}
                      <Text size="sm" fw={600} c={accepted ? 'teal.3' : 'red.3'}>
                        {verdict.verdict}
                      </Text>
                    </Group>
                    <Text size="xs" c="dimmed">
                      Passed: {verdict.passed_cases}/{verdict.total_cases}
                      {verdict.runtime_ms > 0 && ` · ${verdict.runtime_ms}ms`}
                    </Text>
                    {verdict.error_message && (
                      <Code block mt="xs" style={{ fontSize: 11 }}>{verdict.error_message}</Code>
                    )}
                  </Paper>
                )}

                {(submitting || polling) && (
                  <Group gap="xs">
                    <Loader size="xs" />
                    <Text size="xs" c="dimmed">Evaluating…</Text>
                  </Group>
                )}

                {history?.submissions?.map((s) => (
                  <Group key={s.id} justify="space-between" px="xs" py={6}
                    style={{ borderBottom: '1px solid var(--mantine-color-dark-4)' }}>
                    <StatusBadge status={s.status} label={s.verdict || undefined} />
                    <Text size="xs" c="dimmed">
                      {new Date(s.submitted_at).toLocaleString()}
                    </Text>
                  </Group>
                ))}

                {!submitting && !history?.submissions?.length && (
                  <Text size="xs" c="dimmed">No submissions yet.</Text>
                )}
              </Stack>
            </Tabs.Panel>
          </Tabs>
        </Panel>

        <PanelResizeHandle style={{ width: 4, background: 'var(--mantine-color-dark-4)', cursor: 'col-resize' }} />

        <Panel defaultSize={60} minSize={40}>
          <Editor
            height="100%"
            language={language === 'go' ? 'go' : 'python'}
            value={code}
            onChange={(v) => setCode(v ?? '')}
            theme="vs-dark"
            options={{
              fontSize: 14,
              minimap: { enabled: false },
              scrollBeyondLastLine: false,
              padding: { top: 12 },
              fontFamily: 'ui-monospace, Consolas, monospace',
            }}
          />
        </Panel>
      </PanelGroup>
    </div>
  )
}
