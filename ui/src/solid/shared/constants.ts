import type { Difficulty, Language, ProblemLabels } from '@/api/types'
import type { DraftLabel, DraftTestCase } from './types'

export const DEFAULT_CODE: Record<Language, string> = {
  python: '# Write your solution here\n\n',
  go: [
    'package main',
    '',
    'import (',
    '\t"fmt"',
    '\t"io"',
    '\t"os"',
    ')',
    '',
    'func solve(input string) string {',
    '\treturn ""',
    '}',
    '',
    'func main() {',
    '\tdata, _ := io.ReadAll(os.Stdin)',
    '\tfmt.Print(solve(string(data)))',
    '}',
    ''
  ].join('\n'),
  javascript: [
    "const fs = require('node:fs')",
    '',
    'function solve(input) {',
    "  return ''",
    '}',
    '',
    "const input = fs.readFileSync(0, 'utf8')",
    'const output = solve(input)',
    'if (output !== undefined) {',
    '  process.stdout.write(String(output))',
    '}',
    ''
  ].join('\n'),
  typescript: [
    'function solve(input: string): string {',
    "  return ''",
    '}',
    '',
    'const input = await new Response(Deno.stdin.readable).text()',
    'const output = solve(input)',
    'if (output !== undefined) {',
    '  await Deno.stdout.write(new TextEncoder().encode(String(output)))',
    '}',
    ''
  ].join('\n'),
  rust: [
    'use std::io::{self, Read};',
    '',
    'fn solve(input: &str) -> String {',
    '    String::new()',
    '}',
    '',
    'fn main() {',
    '    let mut input = String::new();',
    '    io::stdin().read_to_string(&mut input).unwrap();',
    '    print!("{}", solve(&input));',
    '}',
    ''
  ].join('\n')
}

export const DEFAULT_STARTER_CODE: Record<Language, string> = {
  python: 'class Solution:\n    pass\n',
  go: 'func solve(input string) string {\n\treturn ""\n}\n',
  javascript: "function solve(input) {\n  return ''\n}\n",
  typescript: 'function solve(input: string): string {\n  return ""\n}\n',
  rust: 'fn solve(input: &str) -> String {\n    String::new()\n}\n'
}

export const EMPTY_LABELS: ProblemLabels = {
  tags: []
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

export const LANGUAGE_OPTIONS: Array<{ name: string; value: Language }> = [
  { name: 'Python', value: 'python' },
  { name: 'Go', value: 'go' },
  { name: 'JavaScript', value: 'javascript' },
  { name: 'TypeScript', value: 'typescript' },
  { name: 'Rust', value: 'rust' }
]

export const PROVIDER_OPTIONS = [
  { name: 'LeetCode', value: 'leetcode' },
  { name: 'NeetCode', value: 'neetcode' },
  { name: 'HackerRank', value: 'hackerrank' }
]
