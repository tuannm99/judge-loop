#!/usr/bin/env node
import { createHash } from 'node:crypto'
import { mkdir, readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'

const rootDir = path.resolve(import.meta.dirname, '..')
const providerPath = path.join(rootDir, 'registry/providers/leetcode/free/problems.json')
const indexPath = path.join(rootDir, 'registry/index.json')
const reportPath = path.join(rootDir, 'registry/reports/leetcode-example-coverage.json')
const cachePath =
  process.env.LEETCODE_EXAMPLE_CACHE || '/tmp/judge-loop-leetcode-example-details.json'
const endpoint = 'https://leetcode.com/graphql/'
const batchSize = Number(process.env.LEETCODE_EXAMPLE_BATCH_SIZE || 8)
const delayMS = Number(process.env.LEETCODE_EXAMPLE_DELAY_MS || 800)
const maxProblems = Number(process.env.LEETCODE_EXAMPLE_MAX_PROBLEMS || 0)
const languages = ['python', 'go', 'javascript', 'typescript', 'rust']
const version = process.argv[2]

const unorderedArrays = new Set([
  'two-sum',
  'top-k-frequent-elements',
  'generate-parentheses',
  'binary-tree-paths',
  'find-all-numbers-disappeared-in-an-array',
  'find-all-duplicates-in-an-array',
  'intersection-of-two-arrays',
  'intersection-of-two-arrays-ii'
])
const unorderedNestedArrays = new Set([
  '3sum',
  '4sum',
  'combination-sum',
  'combination-sum-ii',
  'combination-sum-iii',
  'combinations',
  'permutations',
  'permutations-ii',
  'subsets',
  'subsets-ii',
  'group-anagrams',
  'palindrome-partitioning'
])
const comparatorBySlug = new Map([
  ['course-schedule-ii', { kind: 'valid_topological_order' }]
])

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

async function graphql(query, attempt = 1) {
  const response = await fetch(endpoint, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'User-Agent': 'judge-loop-registry-updater'
    },
    body: JSON.stringify({ query })
  })
  if (!response.ok) {
    if (attempt < 5 && [429, 500, 502, 503, 504].includes(response.status)) {
      await sleep(attempt * 2000)
      return graphql(query, attempt + 1)
    }
    throw new Error(`LeetCode GraphQL ${response.status}: ${await response.text()}`)
  }
  const payload = await response.json()
  if (payload.errors?.length) {
    throw new Error(JSON.stringify(payload.errors))
  }
  return payload.data
}

async function loadCache() {
  try {
    return JSON.parse(await readFile(cachePath, 'utf8'))
  } catch (error) {
    if (error.code === 'ENOENT') return {}
    throw error
  }
}

async function fetchDetails(problems) {
  const cache = await loadCache()
  const targets = maxProblems > 0 ? problems.slice(0, maxProblems) : problems
  const missing = targets.filter((problem) => !cache[problem.slug])

  for (let offset = 0; offset < missing.length; offset += batchSize) {
    const batch = missing.slice(offset, offset + batchSize)
    const fields = batch
      .map(
        (problem, index) =>
          `q${index}: question(titleSlug: ${JSON.stringify(problem.slug)}) {
            titleSlug isPaidOnly metaData exampleTestcaseList content
          }`
      )
      .join('\n')
    const data = await graphql(`query registryExampleBatch {\n${fields}\n}`)
    for (const detail of Object.values(data)) {
      if (detail?.titleSlug && !detail.isPaidOnly) {
        cache[detail.titleSlug] = detail
      }
    }
    await writeFile(cachePath, `${JSON.stringify(cache)}\n`)
    process.stderr.write(
      `Fetched example details ${Math.min(offset + batch.length, missing.length)}/${missing.length}\r`
    )
    if (offset + batch.length < missing.length) await sleep(delayMS)
  }
  if (missing.length > 0) process.stderr.write('\n')
  return cache
}

function decodeEntities(value) {
  const named = {
    amp: '&',
    apos: "'",
    gt: '>',
    lt: '<',
    nbsp: ' ',
    quot: '"'
  }
  return value.replace(/&(#x[\da-f]+|#\d+|[a-z]+);/gi, (match, entity) => {
    if (entity[0] !== '#') return named[entity.toLowerCase()] ?? match
    const hex = entity[1].toLowerCase() === 'x'
    const body = entity.slice(hex ? 2 : 1)
    return String.fromCodePoint(Number.parseInt(body, hex ? 16 : 10))
  })
}

function htmlText(value) {
  return decodeEntities(
    value
      .replace(/<br\s*\/?>/gi, '\n')
      .replace(/<\/(?:p|div|li)>/gi, '\n')
      .replace(/<[^>]+>/g, '')
  )
    .replace(/\r/g, '')
    .trim()
}

function extractOutputs(content) {
  const exampleSpans = [
    ...String(content || '').matchAll(
      /<strong[^>]*>\s*Output:?\s*<\/strong>\s*:?\s*<span[^>]*class=["'][^"']*example-io[^"']*["'][^>]*>([\s\S]*?)<\/span>/gi
    )
  ].map((match) => htmlText(match[1]))
  if (exampleSpans.length > 0) return exampleSpans

  const blocks = [...String(content || '').matchAll(/<pre[^>]*>([\s\S]*?)<\/pre>/gi)]
  const outputs = []
  for (const match of blocks) {
    const text = htmlText(match[1])
    const output = text.match(
      /(?:^|\n)\s*Output:?\s*([\s\S]*?)(?=\n\s*(?:Explanation|Input|Constraints|Follow-up):?\s*|$)/i
    )
    if (output) outputs.push(output[1].trim())
  }
  return outputs
}

function splitJSONValues(value) {
  const values = []
  let start = -1
  let depth = 0
  let string = false
  let escaped = false
  const flush = (end) => {
    if (start < 0) return
    const candidate = value.slice(start, end).trim()
    if (candidate) values.push(candidate)
    start = -1
  }
  for (let index = 0; index < value.length; index += 1) {
    const char = value[index]
    if (start < 0) {
      if (/\s/.test(char)) continue
      start = index
    }
    if (string) {
      if (escaped) escaped = false
      else if (char === '\\') escaped = true
      else if (char === '"') string = false
      continue
    }
    if (char === '"') string = true
    else if ('[{'.includes(char)) depth += 1
    else if (']}'.includes(char)) depth -= 1
    if (depth === 0 && !string && (index + 1 === value.length || /\s/.test(value[index + 1]))) {
      flush(index + 1)
    }
  }
  flush(value.length)
  return values.map((item) => JSON.parse(item))
}

function parseExpected(raw) {
  const value = raw
    .replace(/^\s*output\s*=\s*/i, '')
    .replace(/\u2212/g, '-')
    .trim()
  const candidates = [value]
  if (/^'.*'$/s.test(value)) candidates.push(JSON.stringify(value.slice(1, -1)))
  if (/^\(.*\)$/s.test(value)) candidates.push(`[${value.slice(1, -1)}]`)
  for (const candidate of candidates) {
    try {
      return { ok: true, value: JSON.parse(candidate) }
    } catch {
      // Try the next conservative normalization.
    }
  }
  return { ok: false }
}

function normalizeType(raw) {
  let value = String(raw || '').replace(/\s/g, '')
  const lower = value.toLowerCase()
  const scalar = {
    integer: 'int',
    int: 'int',
    long: 'int64',
    double: 'float',
    float: 'float',
    boolean: 'bool',
    bool: 'bool',
    string: 'string',
    character: 'string',
    void: 'void',
    listnode: 'ListNode',
    treenode: 'TreeNode',
    graphnode: 'GraphNode'
  }
  if (scalar[lower]) return scalar[lower]
  const list = value.match(/^(?:list|array)<(.+)>$/i)
  if (list) return `${normalizeType(list[1])}[]`
  while (value.endsWith('[]')) {
    return `${normalizeType(value.slice(0, -2))}[]`
  }
  if (lower === 'node') return 'Node'
  return value || 'any'
}

function signature(raw = {}) {
  const params = raw.params || []
  const names = new Set(params.map((param) => param.name))
  const generatedSizeParam = (name = '') => {
    if (/^return(?:column)?sizes?$/i.test(name)) return true
    const base = name
      .replace(/(?:row|col)size$/i, '')
      .replace(/size$/i, '')
    return base !== name && names.has(base)
  }
  return {
    params: params
      .filter((param) => param.lang !== 'c' && !generatedSizeParam(param.name))
      .map((param) => ({
        name: param.name || 'arg',
        type: normalizeType(param.type)
      })),
    returns: normalizeType(raw.return?.type || 'void')
  }
}

function comparatorFor(problem, content) {
  if (comparatorBySlug.has(problem.slug)) return comparatorBySlug.get(problem.slug)
  if (unorderedNestedArrays.has(problem.slug)) return { kind: 'unordered_nested_array' }
  if (unorderedArrays.has(problem.slug)) return { kind: 'unordered_array' }
  if (/double|float/i.test(JSON.stringify(problem.execution_spec?.signature?.returns))) {
    return { kind: 'float_epsilon', epsilon: 1e-5 }
  }
  if (/\b(?:any valid|any order|return any|multiple (?:valid )?answers)\b/i.test(htmlText(content))) {
    return { kind: 'custom' }
  }
  return { kind: 'exact' }
}

function executionSpec(problem, metadata, content) {
  const supported = languages.filter((language) => problem.starter_code?.[language]?.trim())
  const limits = { timeout_ms: 2000, memory_mb: 128, supported_languages: supported }
  if (!metadata || metadata.manual) return { mode: 'custom', ...limits }
  if (metadata.systemdesign && metadata.classname) {
    return {
      mode: 'class',
      class_name: metadata.classname.trim(),
      constructor: signature(metadata.constructor),
      methods: Object.fromEntries(
        (metadata.methods || []).map((method) => [
          method.name.trim(),
          {
            params: signature(method).params,
            returns: normalizeType(method.return?.type || 'void')
          }
        ])
      ),
      comparator: { kind: 'exact' },
      ...limits
    }
  }
  const base = signature(metadata)
  if (problem.slug === 'clone-graph') {
    base.params = base.params.map((param) => ({
      ...param,
      type: param.type === 'Node' ? 'GraphNode' : param.type
    }))
    if (base.returns === 'Node') base.returns = 'GraphNode'
  }
  if (!metadata.name || base.params.length === 0) return { mode: 'custom', ...limits }
  const outputIndex = metadata.output?.paramindex
  const mode = Number.isInteger(outputIndex) ? 'in_place' : 'function'
  return {
    mode,
    entrypoint: metadata.name.trim(),
    signature: base,
    ...(mode === 'in_place'
      ? { output: { source: 'param', param_index: outputIndex } }
      : {}),
    comparator: comparatorFor(problem, content),
    ...limits
  }
}

function supportedType(raw, allowVoid = false) {
  let value = String(raw || '').toLowerCase()
  while (value.endsWith('[]')) value = value.slice(0, -2)
  if (['int', 'int64', 'float', 'bool', 'string', 'listnode', 'treenode', 'graphnode'].includes(value)) {
    return true
  }
  return allowVoid && value === 'void'
}

function executableSpec(spec) {
  if (!['function', 'in_place', 'class'].includes(spec.mode)) return false
  if (!spec.supported_languages?.length) return false
  if (spec.comparator?.kind === 'custom') return false
  if (spec.mode === 'class') {
    return (
      spec.class_name &&
      Object.keys(spec.methods || {}).length > 0 &&
      (spec.constructor?.params || []).every((param) => supportedType(param.type)) &&
      Object.values(spec.methods || {}).every(
        (method) =>
          method.params.every((param) => supportedType(param.type)) &&
          supportedType(method.returns, true)
      )
    )
  }
  return (
    spec.entrypoint &&
    spec.signature.params.every((param) => supportedType(param.type)) &&
    supportedType(spec.signature.returns, true)
  )
}

function valueMatchesType(value, rawType) {
  let type = String(rawType || '').toLowerCase()
  if (type.endsWith('[]')) {
    if (!Array.isArray(value)) return false
    const child = type.slice(0, -2)
    return value.every((item) => valueMatchesType(item, child))
  }
  if (['listnode', 'treenode', 'graphnode'].includes(type)) {
    return value === null || Array.isArray(value)
  }
  if (['int', 'int64', 'float'].includes(type)) return typeof value === 'number'
  if (type === 'bool') return typeof value === 'boolean'
  if (type === 'string') return typeof value === 'string'
  return false
}

function exampleCases(problem, detail, metadata, spec) {
  const inputs = detail.exampleTestcaseList || []
  const outputs = extractOutputs(detail.content)
  const cases = []
  for (let index = 0; index < Math.min(inputs.length, outputs.length); index += 1) {
    let values
    let expected
    try {
      values = splitJSONValues(inputs[index])
      expected = parseExpected(outputs[index])
    } catch {
      continue
    }
    if (!expected.ok) continue
    let inputJSON
    let expectedValue = expected.value
    if (spec.mode === 'class') {
      if (values.length !== 2 || !Array.isArray(values[0]) || !Array.isArray(values[1])) continue
      const operations = values[0]
      const argumentsList = values[1]
      if (operations.length !== argumentsList.length || operations.length === 0) continue
      const constructorArgs =
        spec.constructor?.params?.length === 0 &&
        argumentsList[0]?.length === 1 &&
        argumentsList[0][0] === null
          ? []
          : argumentsList[0]
      inputJSON = {
        constructor: constructorArgs,
        calls: operations.slice(1).map((method, callIndex) => ({
          method,
          args:
            spec.methods?.[method]?.params?.length === 0 &&
            argumentsList[callIndex + 1]?.length === 1 &&
            argumentsList[callIndex + 1][0] === null
              ? []
              : argumentsList[callIndex + 1]
        }))
      }
      if (Array.isArray(expectedValue) && expectedValue.length === operations.length) {
        expectedValue = expectedValue.slice(1)
      }
    } else {
      if (values.length !== (metadata.params || []).length) continue
      if (
        !values.every((value, paramIndex) =>
          valueMatchesType(value, spec.signature?.params?.[paramIndex]?.type)
        )
      ) {
        continue
      }
      inputJSON = { args: values }
      const outputType =
        spec.mode === 'in_place'
          ? spec.signature.params[spec.output.param_index]?.type
          : spec.signature?.returns
      if (!valueMatchesType(expectedValue, outputType) && outputType !== 'void') continue
    }
    const legacyInput = Object.fromEntries(
      (metadata.params || []).map((param, paramIndex) => [param.name, values[paramIndex]])
    )
    cases.push({
      name: `leetcode example ${index + 1}`,
      input: JSON.stringify(legacyInput),
      expected: JSON.stringify(expectedValue),
      input_json: inputJSON,
      expected_json: expectedValue,
      metadata: { source: 'leetcode-example' },
      is_hidden: false
    })
  }
  return cases
}

function hasManualCases(problem) {
  return problem.test_cases?.some((testCase) => testCase.metadata?.source === 'manual')
}

const provider = JSON.parse(await readFile(providerPath, 'utf8'))
const details = await fetchDetails(provider.problems)
let classified = 0
let suites = 0
let cases = 0
let ready = 0
let noExampleInputs = 0
let noExampleOutputs = 0
let unparseableExamples = 0
let unsupportedSuites = 0
const modeCounts = {}
const coverage = {
  no_example_inputs: [],
  no_parseable_outputs: [],
  input_output_shape_not_parseable: [],
  suite_not_executable: [],
  judge_ready: []
}

provider.problems = provider.problems.map((problem) => {
  const detail = details[problem.slug]
  if (!detail) return problem
  let metadata
  try {
    metadata = JSON.parse(detail.metaData || '{}')
  } catch {
    metadata = {}
  }
  const spec = executionSpec(problem, metadata, detail.content)
  classified += 1
  modeCounts[spec.mode] = (modeCounts[spec.mode] || 0) + 1
  const preserveManual = hasManualCases(problem)
  const generatedCases = preserveManual
    ? problem.test_cases
    : exampleCases(problem, detail, metadata, spec)
  const finalSpec = preserveManual ? problem.execution_spec : spec
  const generatedReady =
    generatedCases.length > 0 &&
    executableSpec(finalSpec) &&
    generatedCases.every(
      (testCase) => testCase.input_json !== undefined && testCase.expected_json !== undefined
    )
  const judgeReady = preserveManual ? problem.judge_ready === true : generatedReady
  if (judgeReady) {
    coverage.judge_ready.push(problem.slug)
  } else if (!detail.exampleTestcaseList?.length) {
    noExampleInputs += 1
    coverage.no_example_inputs.push(problem.slug)
  } else if (!extractOutputs(detail.content).length) {
    noExampleOutputs += 1
    coverage.no_parseable_outputs.push(problem.slug)
  } else if (!generatedCases.length) {
    unparseableExamples += 1
    coverage.input_output_shape_not_parseable.push(problem.slug)
  } else {
    unsupportedSuites += 1
    coverage.suite_not_executable.push(problem.slug)
  }
  if (generatedCases.length > 0) suites += 1
  cases += generatedCases.length
  if (judgeReady) ready += 1
  return {
    ...problem,
    execution_spec: finalSpec,
    test_cases: generatedCases,
    judge_ready: preserveManual ? problem.judge_ready : judgeReady,
    version: Math.max(problem.version || 1, 3)
  }
})

provider.updated_at = new Date().toISOString()
const providerJSON = `${JSON.stringify(provider, null, 2)}\n`
const checksum = createHash('sha256').update(providerJSON).digest('hex')
const index = JSON.parse(await readFile(indexPath, 'utf8'))
if (version) index.version = version
index.updated_at = new Date().toISOString()
index.manifests = index.manifests.map((manifest) =>
  manifest.name === 'leetcode'
    ? { ...manifest, checksum: `sha256:${checksum}` }
    : manifest
)
await writeFile(providerPath, providerJSON)
await writeFile(indexPath, `${JSON.stringify(index, null, 2)}\n`)
await mkdir(path.dirname(reportPath), { recursive: true })
await writeFile(
  reportPath,
  `${JSON.stringify(
    {
      updated_at: provider.updated_at,
      total: provider.problems.length,
      classified,
      with_suites: suites,
      test_cases: cases,
      judge_ready: ready,
      execution_modes: modeCounts,
      coverage
    },
    null,
    2
  )}\n`
)

console.log(`Classified problems: ${classified}/${provider.problems.length}`)
console.log(`Problems with suites: ${suites}`)
console.log(`Generated/preserved cases: ${cases}`)
console.log(`Judge-ready problems: ${ready}`)
console.log(`Execution modes: ${JSON.stringify(modeCounts)}`)
console.log(`No example inputs: ${noExampleInputs}`)
console.log(`No parseable outputs: ${noExampleOutputs}`)
console.log(`Input/output shape not parseable: ${unparseableExamples}`)
console.log(`Suites retained but not executable: ${unsupportedSuites}`)
console.log(`Checksum: sha256:${checksum}`)
