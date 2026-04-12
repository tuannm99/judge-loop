#!/usr/bin/env node
import { createHash } from 'node:crypto'
import { mkdir, readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'

const rootDir = path.resolve(import.meta.dirname, '..')
const providerPath = path.join(rootDir, 'registry/providers/leetcode.json')
const indexPath = path.join(rootDir, 'registry/index.json')
const endpoint = 'https://leetcode.com/graphql/'
const pageSize = 100
const detailBatchSize = 25
const version =
  process.argv[2] || `${new Date().toISOString().slice(0, 10)}-leetcode-free`

const listQuery = `
query problemsetQuestionList($categorySlug: String, $limit: Int, $skip: Int, $filters: QuestionListFilterInput) {
  problemsetQuestionList: questionList(categorySlug: $categorySlug, limit: $limit, skip: $skip, filters: $filters) {
    total: totalNum
    questions: data {
      difficulty
      frontendQuestionId: questionFrontendId
      paidOnly: isPaidOnly
      title
      titleSlug
      topicTags { name slug }
    }
  }
}`

async function graphql(query, variables) {
  const response = await fetch(endpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'User-Agent': 'judge-loop-registry-updater'
    },
    body: JSON.stringify({ query, variables })
  })

  if (!response.ok) {
    throw new Error(
      `LeetCode GraphQL ${response.status}: ${await response.text()}`
    )
  }

  const payload = await response.json()
  if (payload.errors?.length) {
    throw new Error(`LeetCode GraphQL error: ${JSON.stringify(payload.errors)}`)
  }
  return payload.data
}

async function loadQuestions() {
  let total = 1
  let skip = 0
  const questions = []

  while (skip < total) {
    const data = await graphql(listQuery, {
      categorySlug: '',
      skip,
      limit: pageSize,
      filters: {}
    })
    const page = data.problemsetQuestionList
    total = page.total
    questions.push(...page.questions)
    skip += pageSize
    process.stderr.write(
      `Fetched ${Math.min(skip, total)}/${total} problem metadata\r`
    )
  }
  process.stderr.write('\n')
  return questions
}

function quoteGraphQLString(value) {
  return JSON.stringify(value)
}

async function loadStarterCodeBySlug(slugs) {
  const out = new Map()

  for (let i = 0; i < slugs.length; i += detailBatchSize) {
    const batch = slugs.slice(i, i + detailBatchSize)
    const fields = batch
      .map((slug, idx) => {
        return `q${idx}: question(titleSlug: ${quoteGraphQLString(slug)}) { titleSlug isPaidOnly codeSnippets { langSlug code } }`
      })
      .join('\n')
    const data = await graphql(`query questionBatch {\n${fields}\n}`, {})

    for (const item of Object.values(data)) {
      if (!item || item.isPaidOnly) {
        continue
      }
      const starter = {}
      const snippets = item.codeSnippets || []
      const python =
        snippets.find((s) => s.langSlug === 'python3') ||
        snippets.find((s) => s.langSlug === 'python')
      const go = snippets.find((s) => s.langSlug === 'golang')

      if (python?.code) {
        starter.python = python.code
      }
      if (go?.code) {
        starter.go = go.code
      }
      out.set(item.titleSlug, starter)
    }
    process.stderr.write(
      `Fetched starter code ${Math.min(i + detailBatchSize, slugs.length)}/${slugs.length}\r`
    )
  }
  process.stderr.write('\n')
  return out
}

function estimateMinutes(difficulty) {
  if (difficulty === 'Easy') {
    return 15
  }
  if (difficulty === 'Medium') {
    return 30
  }
  return 45
}

function toManifest(question, starterCodeBySlug) {
  const tags = (question.topicTags || []).map((tag) => tag.slug).filter(Boolean)
  return {
    provider: 'leetcode',
    external_id: question.frontendQuestionId,
    slug: question.titleSlug,
    title: question.title,
    difficulty: question.difficulty.toLowerCase(),
    tags,
    source_url: `https://leetcode.com/problems/${question.titleSlug}/`,
    estimated_time: estimateMinutes(question.difficulty),
    starter_code: starterCodeBySlug.get(question.titleSlug) || {},
    version: 1
  }
}

async function updateIndex(checksum) {
  const index = JSON.parse(await readFile(indexPath, 'utf8'))
  index.version = version
  index.updated_at = new Date().toISOString()
  index.manifests = index.manifests.map((manifest) => {
    if (manifest.name !== 'leetcode') {
      return manifest
    }
    return { ...manifest, checksum: `sha256:${checksum}` }
  })
  await writeFile(indexPath, `${JSON.stringify(index, null, 2)}\n`)
}

const questions = await loadQuestions()
const freeQuestions = questions
  .filter((question) => question.paidOnly === false)
  .sort((a, b) => Number(a.frontendQuestionId) - Number(b.frontendQuestionId))
const starterCodeBySlug = await loadStarterCodeBySlug(
  freeQuestions.map((question) => question.titleSlug)
)
const manifest = {
  problems: freeQuestions.map((question) =>
    toManifest(question, starterCodeBySlug)
  )
}

await mkdir(path.dirname(providerPath), { recursive: true })
const providerJSON = `${JSON.stringify(manifest, null, 2)}\n`
await writeFile(providerPath, providerJSON)

const checksum = createHash('sha256').update(providerJSON).digest('hex')
await updateIndex(checksum)

const withStarter = manifest.problems.filter(
  (problem) => Object.keys(problem.starter_code).length > 0
).length
console.log(`Updated ${providerPath}`)
console.log(`Free problems: ${manifest.problems.length}`)
console.log(`Problems with starter code: ${withStarter}`)
console.log(`Index version: ${version}`)
console.log(`Checksum: sha256:${checksum}`)
