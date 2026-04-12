export type NavigateFn = (to: string) => void
export type LabelKind = 'tag' | 'pattern'
export type SortMode = 'default' | 'title' | 'difficulty' | 'time-desc' | 'provider'

export type DraftTestCase = {
  input: string
  expected: string
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
