export type NavigateFn = (to: string) => void
export type SortMode = 'default' | 'title' | 'difficulty' | 'time-desc' | 'provider'

export type DraftTestCase = {
  name?: string
  input: string
  expected: string
  input_json?: unknown
  expected_json?: unknown
  metadata?: unknown
  is_hidden: boolean
}

export type SolveTestCase = {
  input: string
  expected: string
  is_hidden?: boolean
}

export type DraftLabel = {
  slug: string
  name: string
}
