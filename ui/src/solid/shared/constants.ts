import type { Difficulty, Language, ProblemLabels } from '@/api/types'
import type { DraftLabel, DraftTestCase } from './types'

export const DEFAULT_CODE: Record<Language, string> = {
  python: '# Write your solution here\n\n',
  go: 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println()\n}\n'
}

export const DEFAULT_STARTER_CODE: Record<Language, string> = {
  python: 'class Solution:\n    pass\n',
  go: 'package main\n\nfunc main() {\n\n}\n'
}

export const EMPTY_LABELS: ProblemLabels = {
  tags: [],
  patterns: []
}

export const EMPTY_TEST_CASE: DraftTestCase = {
  input: '',
  expected: '',
  is_hidden: false
}

export const EMPTY_DRAFT_LABEL: DraftLabel = {
  slug: '',
  name: ''
}

export const DIFFICULTY_ORDER: Record<Difficulty, number> = {
  easy: 1,
  medium: 2,
  hard: 3
}

export const DIFFICULTY_OPTIONS = [
  { name: 'All difficulties', value: '' },
  { name: 'Easy', value: 'easy' },
  { name: 'Medium', value: 'medium' },
  { name: 'Hard', value: 'hard' }
]

export const SORT_OPTIONS = [
  { name: 'Default order', value: 'default' },
  { name: 'Title A-Z', value: 'title' },
  { name: 'Difficulty', value: 'difficulty' },
  { name: 'Longest estimate', value: 'time-desc' },
  { name: 'Provider', value: 'provider' }
]

export const LANGUAGE_OPTIONS = [
  { name: 'Python', value: 'python' },
  { name: 'Go', value: 'go' }
]

export const PROVIDER_OPTIONS = [
  { name: 'LeetCode', value: 'leetcode' },
  { name: 'NeetCode', value: 'neetcode' },
  { name: 'HackerRank', value: 'hackerrank' }
]
