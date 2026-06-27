export type Difficulty = 'easy' | 'medium' | 'hard'
export type Provider = 'leetcode' | 'neetcode' | 'hackerrank'
export type Language = 'python' | 'go' | 'javascript' | 'typescript' | 'rust'
export type SubmissionStatus =
  | 'pending'
  | 'running'
  | 'accepted'
  | 'wrong_answer'
  | 'compile_error'
  | 'runtime_error'
  | 'time_limit_exceeded'

export interface Problem {
  id: string
  slug: string
  title: string
  difficulty: Difficulty
  tags: string[]
  provider: Provider
  external_id: string
  source_url: string
  estimated_time: number
  description_markdown: string
  starter_code: Partial<Record<Language, string>>
  execution_spec: ExecutionSpec
  judge_ready: boolean
}

export interface ExecutionSpec {
  mode: 'stdin' | 'function' | 'class' | 'in_place' | 'interactive' | 'custom' | ''
  entrypoint?: string
  class_name?: string
  signature?: ExecutionSignature
  constructor?: ExecutionSignature
  methods?: Record<string, ExecutionMethod>
  output?: { source?: 'return' | 'param'; param_index?: number }
  bindings?: Partial<Record<Language, ExecutionLanguageBinding>>
  supported_languages?: Language[]
  comparator?: { kind?: string; epsilon?: number }
  timeout_ms?: number
  memory_mb?: number
}

export interface ExecutionSignature {
  params?: Array<{ name: string; type: string }>
  returns?: string
}

export interface ExecutionMethod extends ExecutionSignature {
  bindings?: Partial<Record<Language, string>>
}

export interface ExecutionLanguageBinding {
  entrypoint?: string
  class_name?: string
  constructor?: string
}

export interface Submission {
  id: string
  user_id: string
  problem_id: string
  session_id?: string
  language: Language
  code: string
  status: SubmissionStatus
  verdict: string
  passed_cases: number
  total_cases: number
  runtime_ms: number
  error_message: string
  submitted_at: string
  evaluated_at?: string
}

export interface ProgressToday {
  date: string
  solved: number
  attempted: number
  time_spent_minutes: number
  streak: number
}

export interface Streak {
  current: number
  longest: number
  last_practiced: string
}

export interface ReviewItem {
  problem_id: string
  slug: string
  title: string
  days_overdue: number
}

export interface Timer {
  active: boolean
  id?: string
  started_at?: string
  elapsed_seconds?: number
  problem_id?: string
}

export interface ProblemLabels {
  tags: string[]
}

export interface ProblemLabel {
  id: string
  kind: 'tag'
  slug: string
  name: string
  created_at: string
  updated_at: string
}
