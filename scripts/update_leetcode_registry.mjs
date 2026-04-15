#!/usr/bin/env node
import { createHash } from 'node:crypto'
import { mkdir, readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'

const rootDir = path.resolve(import.meta.dirname, '..')
const providerRoot = path.join(rootDir, 'registry/providers/leetcode')
const freeProviderPath = path.join(providerRoot, 'free/problems.json')
const premiumProviderPath = path.join(providerRoot, 'premium/problems.json')
const indexPath = path.join(rootDir, 'registry/index.json')
const endpoint = 'https://leetcode.com/graphql/'
const pageSize = 100
const detailBatchSize = Number(process.env.LEETCODE_DETAIL_BATCH_SIZE || 10)
const listDelayMS = Number(process.env.LEETCODE_LIST_DELAY_MS || 500)
const detailDelayMS = Number(process.env.LEETCODE_DETAIL_DELAY_MS || 1200)
const maxDetailProblems = Number(process.env.LEETCODE_MAX_DETAIL_PROBLEMS || 0)
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

function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms)
  })
}

async function graphql(query, variables, attempt = 1) {
  const response = await fetch(endpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'User-Agent': 'judge-loop-registry-updater'
    },
    body: JSON.stringify({ query, variables })
  })

  if (!response.ok) {
    const body = await response.text()
    if (attempt < 4 && [429, 500, 502, 503, 504].includes(response.status)) {
      const wait = 2000 * attempt
      process.stderr.write(
        `LeetCode GraphQL ${response.status}; retrying in ${wait}ms\n`
      )
      await sleep(wait)
      return graphql(query, variables, attempt + 1)
    }
    throw new Error(`LeetCode GraphQL ${response.status}: ${body}`)
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
    if (skip < total) {
      await sleep(listDelayMS)
    }
  }
  process.stderr.write('\n')
  return questions
}

function quoteGraphQLString(value) {
  return JSON.stringify(value)
}

async function loadStarterCodeBySlug(slugs) {
  const out = new Map()
  const detailSlugs =
    maxDetailProblems > 0 ? slugs.slice(0, maxDetailProblems) : slugs

  for (let i = 0; i < detailSlugs.length; i += detailBatchSize) {
    const batch = detailSlugs.slice(i, i + detailBatchSize)
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
      const languageMap = {
        python: ['python3', 'python'],
        go: ['golang'],
        javascript: ['javascript'],
        typescript: ['typescript'],
        rust: ['rust']
      }

      for (const [language, langSlugs] of Object.entries(languageMap)) {
        const snippet = snippets.find((s) => langSlugs.includes(s.langSlug))
        if (snippet?.code) {
          starter[language] = snippet.code
        }
      }
      out.set(item.titleSlug, starter)
    }
    process.stderr.write(
      `Fetched starter code ${Math.min(i + detailBatchSize, detailSlugs.length)}/${detailSlugs.length}\r`
    )
    if (i + detailBatchSize < detailSlugs.length) {
      await sleep(detailDelayMS)
    }
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

function compareQuestionID(a, b) {
  return String(a.frontendQuestionId).localeCompare(String(b.frontendQuestionId), undefined, {
    numeric: true,
    sensitivity: 'base'
  })
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
    pattern_tags: tags,
    source_url: `https://leetcode.com/problems/${question.titleSlug}/`,
    estimated_time: estimateMinutes(question.difficulty),
    description_markdown: '',
    starter_code: starterCodeBySlug.get(question.titleSlug) || {},
    version: 1
  }
}

function toPremiumManifest(question) {
  const tags = (question.topicTags || []).map((tag) => tag.slug).filter(Boolean)
  return {
    provider: 'leetcode',
    external_id: question.frontendQuestionId,
    slug: question.titleSlug,
    title: question.title,
    difficulty: question.difficulty.toLowerCase(),
    tags,
    pattern_tags: tags,
    paid_only: true,
    source_url: `https://leetcode.com/problems/${question.titleSlug}/`,
    estimated_time: estimateMinutes(question.difficulty),
    description_markdown: '',
    starter_code: {},
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
    return {
      ...manifest,
      path: 'providers/leetcode/free/problems.json',
      checksum: `sha256:${checksum}`
    }
  })
  await writeFile(indexPath, `${JSON.stringify(index, null, 2)}\n`)
}

const questions = await loadQuestions()
const freeQuestions = questions
  .filter((question) => question.paidOnly === false)
  .sort(compareQuestionID)
const premiumQuestions = questions
  .filter((question) => question.paidOnly === true)
  .sort(compareQuestionID)
const starterCodeBySlug = await loadStarterCodeBySlug(
  freeQuestions.map((question) => question.titleSlug)
)
const now = new Date().toISOString()
const freeManifest = {
  provider: 'leetcode',
  kind: 'free',
  updated_at: now,
  problems: freeQuestions.map((question) =>
    toManifest(question, starterCodeBySlug)
  )
}
const premiumManifest = {
  provider: 'leetcode',
  kind: 'premium',
  updated_at: now,
  problems: premiumQuestions.map(toPremiumManifest)
}

await mkdir(path.dirname(freeProviderPath), { recursive: true })
await mkdir(path.dirname(premiumProviderPath), { recursive: true })
const freeProviderJSON = `${JSON.stringify(freeManifest, null, 2)}\n`
const premiumProviderJSON = `${JSON.stringify(premiumManifest, null, 2)}\n`
await writeFile(freeProviderPath, freeProviderJSON)
await writeFile(premiumProviderPath, premiumProviderJSON)

const checksum = createHash('sha256').update(freeProviderJSON).digest('hex')
const premiumChecksum = createHash('sha256').update(premiumProviderJSON).digest('hex')
await updateIndex(checksum)

const languageCounts = ['python', 'go', 'javascript', 'typescript', 'rust'].map(
  (language) => [
    language,
    freeManifest.problems.filter((problem) => problem.starter_code[language]).length
  ]
)
const withStarter = freeManifest.problems.filter(
  (problem) => Object.keys(problem.starter_code).length > 0
).length
console.log(`Updated ${freeProviderPath}`)
console.log(`Updated ${premiumProviderPath}`)
console.log(`Free problems: ${freeManifest.problems.length}`)
console.log(`Premium problems: ${premiumManifest.problems.length}`)
console.log(`Problems with starter code: ${withStarter}`)
for (const [language, count] of languageCounts) {
  console.log(`${language} starter code: ${count}`)
}
console.log(`Index version: ${version}`)
console.log(`Checksum: sha256:${checksum}`)
console.log(`Premium checksum: sha256:${premiumChecksum}`)
