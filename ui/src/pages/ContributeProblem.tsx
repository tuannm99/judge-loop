import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { contributeProblem, listProblemLabels } from '@/api/client'
import type { Difficulty, Language, Provider } from '@/api/types'
import {
  Alert,
  Button,
  Checkbox,
  Group,
  MultiSelect,
  NativeSelect,
  NumberInput,
  Paper,
  Stack,
  Text,
  TextInput,
  Textarea,
  Title
} from '@mantine/core'
import { IconPlus, IconTrash } from '@tabler/icons-react'

type DraftTestCase = {
  input: string
  expected: string
  is_hidden: boolean
}

const DEFAULT_STARTER_CODE: Record<Language, string> = {
  python: 'class Solution:\n    pass\n',
  go: 'package main\n\nfunc main() {\n\n}\n'
}

export function ContributeProblem() {
  const navigate = useNavigate()
  const qc = useQueryClient()

  const [provider, setProvider] = useState<Provider>('leetcode')
  const [externalID, setExternalID] = useState('')
  const [slug, setSlug] = useState('')
  const [title, setTitle] = useState('')
  const [difficulty, setDifficulty] = useState<Difficulty>('easy')
  const [tags, setTags] = useState<string[]>([])
  const [patternTags, setPatternTags] = useState<string[]>([])
  const [sourceURL, setSourceURL] = useState('')
  const [estimatedTime, setEstimatedTime] = useState<number>(15)
  const [version, setVersion] = useState<number>(1)
  const [pythonStarter, setPythonStarter] = useState(
    DEFAULT_STARTER_CODE.python
  )
  const [goStarter, setGoStarter] = useState(DEFAULT_STARTER_CODE.go)
  const [testCases, setTestCases] = useState<DraftTestCase[]>([
    { input: '', expected: '', is_hidden: false }
  ])

  const { mutate, isPending, error, data } = useMutation({
    mutationFn: contributeProblem,
    onSuccess: (problem) => {
      void qc.invalidateQueries({ queryKey: ['problems'] })
      void qc.invalidateQueries({ queryKey: ['problem', problem.slug] })
      navigate(`/problems/${problem.slug}`)
    }
  })
  const { data: labels } = useQuery({
    queryKey: ['problem-labels'],
    queryFn: listProblemLabels
  })

  function updateTestCase(index: number, patch: Partial<DraftTestCase>) {
    setTestCases((current) =>
      current.map((item, i) => (i === index ? { ...item, ...patch } : item))
    )
  }

  function addTestCase() {
    setTestCases((current) => [
      ...current,
      { input: '', expected: '', is_hidden: false }
    ])
  }

  function removeTestCase(index: number) {
    setTestCases((current) =>
      current.length === 1 ? current : current.filter((_, i) => i !== index)
    )
  }

  function handleSubmit() {
    mutate({
      provider,
      external_id: externalID,
      slug,
      title,
      difficulty,
      tags,
      pattern_tags: patternTags,
      source_url: sourceURL,
      estimated_time: estimatedTime || 0,
      starter_code: {
        python: pythonStarter,
        go: goStarter
      },
      version: version || 1,
      test_cases: testCases
    })
  }

  return (
    <Stack p="lg" gap="lg" maw={960}>
      <div>
        <Title order={2}>Contribute Problem</Title>
        <Text size="sm" c="dimmed">
          Add a runnable problem with starter code and judge test cases.
        </Text>
      </div>

      {error && (
        <Alert color="red" title="Submit failed">
          {error.message}
        </Alert>
      )}

      {data && (
        <Alert color="teal" title="Saved">
          Problem saved as {data.slug}
        </Alert>
      )}

      <Paper withBorder p="lg">
        <Stack gap="md">
          <Group grow align="start">
            <NativeSelect
              label="Provider"
              value={provider}
              onChange={(e) => setProvider(e.target.value as Provider)}
              data={[
                { value: 'leetcode', label: 'LeetCode' },
                { value: 'neetcode', label: 'NeetCode' },
                { value: 'hackerrank', label: 'HackerRank' }
              ]}
            />
            <TextInput
              label="External ID"
              value={externalID}
              onChange={(e) => setExternalID(e.currentTarget.value)}
            />
            <NativeSelect
              label="Difficulty"
              value={difficulty}
              onChange={(e) => setDifficulty(e.target.value as Difficulty)}
              data={[
                { value: 'easy', label: 'Easy' },
                { value: 'medium', label: 'Medium' },
                { value: 'hard', label: 'Hard' }
              ]}
            />
          </Group>

          <Group grow align="start">
            <TextInput
              label="Slug"
              value={slug}
              onChange={(e) => setSlug(e.currentTarget.value)}
            />
            <TextInput
              label="Title"
              value={title}
              onChange={(e) => setTitle(e.currentTarget.value)}
            />
          </Group>

          <TextInput
            label="Source URL"
            value={sourceURL}
            onChange={(e) => setSourceURL(e.currentTarget.value)}
          />

          <Group grow align="start">
            <MultiSelect
              label="Tags"
              description="Select from predefined tags in the common schema"
              value={tags}
              onChange={setTags}
              data={(labels?.tags ?? []).map((value) => ({
                value,
                label: value
              }))}
              searchable
              clearable
            />
            <MultiSelect
              label="Pattern Tags"
              description="Select from predefined patterns in the common schema"
              value={patternTags}
              onChange={setPatternTags}
              data={(labels?.patterns ?? []).map((value) => ({
                value,
                label: value
              }))}
              searchable
              clearable
            />
          </Group>

          <Group grow align="start">
            <NumberInput
              label="Estimated Time"
              suffix=" min"
              min={0}
              value={estimatedTime}
              onChange={(value) => setEstimatedTime(Number(value) || 0)}
            />
            <NumberInput
              label="Version"
              min={1}
              value={version}
              onChange={(value) => setVersion(Number(value) || 1)}
            />
          </Group>
        </Stack>
      </Paper>

      <Paper withBorder p="lg">
        <Stack gap="md">
          <Title order={4}>Starter Code</Title>
          <Textarea
            label="Python"
            minRows={8}
            autosize
            value={pythonStarter}
            onChange={(e) => setPythonStarter(e.currentTarget.value)}
          />
          <Textarea
            label="Go"
            minRows={8}
            autosize
            value={goStarter}
            onChange={(e) => setGoStarter(e.currentTarget.value)}
          />
        </Stack>
      </Paper>

      <Paper withBorder p="lg">
        <Stack gap="md">
          <Group justify="space-between">
            <Title order={4}>Test Cases</Title>
            <Button
              variant="default"
              leftSection={<IconPlus size={14} />}
              onClick={addTestCase}
            >
              Add Test Case
            </Button>
          </Group>

          {testCases.map((testCase, index) => (
            <Paper key={index} withBorder p="md">
              <Stack gap="sm">
                <Group justify="space-between">
                  <Text fw={600}>Case {index + 1}</Text>
                  <Button
                    color="red"
                    variant="subtle"
                    leftSection={<IconTrash size={14} />}
                    onClick={() => removeTestCase(index)}
                  >
                    Remove
                  </Button>
                </Group>
                <Textarea
                  label="Input"
                  minRows={4}
                  autosize
                  value={testCase.input}
                  onChange={(e) =>
                    updateTestCase(index, { input: e.currentTarget.value })
                  }
                />
                <Textarea
                  label="Expected Output"
                  minRows={3}
                  autosize
                  value={testCase.expected}
                  onChange={(e) =>
                    updateTestCase(index, { expected: e.currentTarget.value })
                  }
                />
                <Checkbox
                  label="Hidden test case"
                  checked={testCase.is_hidden}
                  onChange={(e) =>
                    updateTestCase(index, { is_hidden: e.currentTarget.checked })
                  }
                />
              </Stack>
            </Paper>
          ))}
        </Stack>
      </Paper>

      <Group justify="flex-end">
        <Button loading={isPending} onClick={handleSubmit}>
          Save Problem
        </Button>
      </Group>
    </Stack>
  )
}
