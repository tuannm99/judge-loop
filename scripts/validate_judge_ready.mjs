#!/usr/bin/env node
import { readdir, readFile } from 'node:fs/promises'
import path from 'node:path'

const rootDir = path.resolve(import.meta.dirname, '..')
const freePath = path.join(rootDir, 'registry/providers/leetcode/free/problems.json')
const judgeReadyDir = path.join(rootDir, 'registry/providers/judge-ready')
const languages = ['python', 'go', 'javascript', 'typescript', 'rust']

async function jsonFiles(dir) {
  const entries = await readdir(dir, { withFileTypes: true }).catch((err) => {
    if (err.code === 'ENOENT') {
      return []
    }
    throw err
  })

  const files = []
  for (const entry of entries) {
    const child = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      files.push(...(await jsonFiles(child)))
    } else if (entry.isFile() && entry.name.endsWith('.json')) {
      files.push(child)
    }
  }
  return files
}

function parseJSONField(value, label, errors) {
  try {
    JSON.parse(value)
  } catch (err) {
    errors.push(`${label} is not valid JSON: ${err.message}`)
  }
}

const free = JSON.parse(await readFile(freePath, 'utf8'))
const freeBySlug = new Map(free.problems.map((problem) => [problem.slug, problem]))
const files = await jsonFiles(judgeReadyDir)
const errors = []
const seen = new Map()
let problemCount = 0
let testCaseCount = 0

for (const file of files) {
  const manifest = JSON.parse(await readFile(file, 'utf8'))
  if (!Array.isArray(manifest.problems)) {
    errors.push(`${path.relative(rootDir, file)}: problems must be an array`)
    continue
  }

  for (const problem of manifest.problems) {
    problemCount += 1
    const where = `${path.relative(rootDir, file)}:${problem.slug || '<missing slug>'}`

    if (!problem.slug) {
      errors.push(`${where}: slug is required`)
      continue
    }
    if (!freeBySlug.has(problem.slug)) {
      errors.push(`${where}: slug is not present in the free LeetCode bank`)
    }
    if (seen.has(problem.slug)) {
      errors.push(`${where}: duplicate slug also found in ${seen.get(problem.slug)}`)
    }
    seen.set(problem.slug, path.relative(rootDir, file))

    for (const field of ['provider', 'external_id', 'title', 'difficulty', 'source_url']) {
      if (!problem[field]) {
        errors.push(`${where}: ${field} is required`)
      }
    }
    if (!problem.description_markdown?.trim()) {
      errors.push(`${where}: description_markdown is required for judge-ready overlays`)
    }

    for (const language of languages) {
      if (!problem.starter_code?.[language]?.trim()) {
        errors.push(`${where}: starter_code.${language} is required`)
      }
    }

    if (!Array.isArray(problem.test_cases) || problem.test_cases.length === 0) {
      errors.push(`${where}: at least one test case is required`)
      continue
    }

    problem.test_cases.forEach((testCase, index) => {
      testCaseCount += 1
      const caseLabel = `${where}:test_cases[${index}]`
      if (typeof testCase.input !== 'string' || testCase.input.trim() === '') {
        errors.push(`${caseLabel}: input is required`)
      } else {
        parseJSONField(testCase.input, `${caseLabel}.input`, errors)
      }
      if (typeof testCase.expected !== 'string' || testCase.expected.trim() === '') {
        errors.push(`${caseLabel}: expected is required`)
      } else {
        parseJSONField(testCase.expected, `${caseLabel}.expected`, errors)
      }
    })
  }
}

if (errors.length > 0) {
  console.error(errors.join('\n'))
  process.exit(1)
}

console.log(`Judge-ready manifests: ${files.length}`)
console.log(`Judge-ready problems: ${problemCount}/${free.problems.length}`)
console.log(`Judge-ready test cases: ${testCaseCount}`)
console.log(`Remaining free problems: ${free.problems.length - problemCount}`)
