#!/usr/bin/env node
import { createHash } from 'node:crypto'
import { readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'

const rootDir = path.resolve(import.meta.dirname, '..')
const freeProviderPath = path.join(rootDir, 'registry/providers/leetcode/free/problems.json')
const indexPath = path.join(rootDir, 'registry/index.json')

const defaultLimits = {
  timeout_ms: 2000,
  memory_mb: 128
}

const manualProblems = [
  {
    slug: 'two-sum',
    description_markdown:
      'Given an integer array and a target value, return the indices of two different elements whose sum equals the target. The local judge accepts the pair in any order.',
    execution_spec: functionSpec(
      'twoSum',
      [
        ['nums', 'int[]'],
        ['target', 'int']
      ],
      'int[]',
      { kind: 'unordered_array' }
    ),
    test_cases: [
      testCase('example pair at front', [[2, 7, 11, 15], 9], [0, 1]),
      testCase('middle pair', [[3, 2, 4], 6], [1, 2]),
      testCase('duplicate values', [[3, 3], 6], [0, 1], true),
      testCase('negative values', [[-1, -2, -3, -4, -5], -8], [2, 4], true)
    ]
  },
  {
    slug: 'valid-parentheses',
    description_markdown:
      'Check whether every opening bracket is closed by the same type of bracket and in the correct order.',
    execution_spec: functionSpec('isValid', [['s', 'string']], 'bool'),
    test_cases: [
      testCase('single pair', ['()'], true),
      testCase('mixed pairs', ['()[]{}'], true),
      testCase('wrong order', ['(]'], false),
      testCase('nested mismatch', ['([)]'], false, true),
      testCase('nested valid', ['{[]}'], true, true)
    ]
  },
  {
    slug: 'maximum-subarray',
    description_markdown:
      'Find the largest possible sum of a non-empty contiguous subarray. The input may contain negative and positive numbers.',
    execution_spec: functionSpec('maxSubArray', [['nums', 'int[]']], 'int'),
    test_cases: [
      testCase('mixed signs', [[-2, 1, -3, 4, -1, 2, 1, -5, 4]], 6),
      testCase('single element', [[1]], 1),
      testCase('all positive', [[5, 4, -1, 7, 8]], 23),
      testCase('all negative', [[-8, -3, -6, -2, -5, -4]], -2, true)
    ]
  },
  {
    slug: 'best-time-to-buy-and-sell-stock',
    description_markdown:
      'Given daily prices, choose one buy day and one later sell day to maximize profit. Return zero when no profitable transaction exists.',
    execution_spec: functionSpec('maxProfit', [['prices', 'int[]']], 'int'),
    test_cases: [
      testCase('buy low sell high', [[7, 1, 5, 3, 6, 4]], 5),
      testCase('descending prices', [[7, 6, 4, 3, 1]], 0),
      testCase('short rise', [[1, 2]], 1),
      testCase('late best sell', [[2, 4, 1, 7]], 6, true)
    ]
  },
  {
    slug: 'valid-palindrome',
    description_markdown:
      'Return whether a string reads the same forward and backward after ignoring case and non-alphanumeric characters.',
    execution_spec: functionSpec('isPalindrome', [['s', 'string']], 'bool'),
    test_cases: [
      testCase('sentence palindrome', ['A man, a plan, a canal: Panama'], true),
      testCase('not palindrome', ['race a car'], false),
      testCase('empty after filtering', [' '], true),
      testCase('mixed case and punctuation', ['No lemon, no melon'], true, true)
    ]
  },
  {
    slug: 'contains-duplicate',
    description_markdown:
      'Return true when any number appears at least twice in the array; otherwise return false.',
    execution_spec: functionSpec('containsDuplicate', [['nums', 'int[]']], 'bool'),
    test_cases: [
      testCase('has duplicate', [[1, 2, 3, 1]], true),
      testCase('all unique', [[1, 2, 3, 4]], false),
      testCase('many duplicates', [[1, 1, 1, 3, 3, 4, 3, 2, 4, 2]], true, true),
      testCase('empty array', [[]], false, true)
    ]
  },
  {
    slug: 'valid-anagram',
    description_markdown:
      'Determine whether two strings contain exactly the same characters with the same frequencies.',
    execution_spec: functionSpec(
      'isAnagram',
      [
        ['s', 'string'],
        ['t', 'string']
      ],
      'bool'
    ),
    test_cases: [
      testCase('same letters', ['anagram', 'nagaram'], true),
      testCase('different counts', ['rat', 'car'], false),
      testCase('empty strings', ['', ''], true),
      testCase('same letters different frequency', ['aacc', 'ccac'], false, true)
    ]
  },
  {
    slug: 'product-of-array-except-self',
    description_markdown:
      'Return an array where each position contains the product of every input value except the one at that position.',
    execution_spec: functionSpec('productExceptSelf', [['nums', 'int[]']], 'int[]'),
    test_cases: [
      testCase('positive values', [[1, 2, 3, 4]], [24, 12, 8, 6]),
      testCase('one zero', [[-1, 1, 0, -3, 3]], [0, 0, 9, 0, 0]),
      testCase('two zeros', [[0, 4, 0]], [0, 0, 0], true),
      testCase('includes negatives', [[2, -3, 4]], [-12, 8, -6], true)
    ]
  },
  {
    slug: 'top-k-frequent-elements',
    description_markdown:
      'Return the k values that occur most often in the array. The local judge accepts those values in any order.',
    execution_spec: functionSpec(
      'topKFrequent',
      [
        ['nums', 'int[]'],
        ['k', 'int']
      ],
      'int[]',
      { kind: 'unordered_array' }
    ),
    test_cases: [
      testCase('single most frequent', [[1, 1, 1, 2, 2, 3], 2], [1, 2]),
      testCase('one value', [[1], 1], [1]),
      testCase('negative values', [[4, 1, -1, 2, -1, 2, 3], 2], [-1, 2], true),
      testCase('all same', [[5, 5, 5], 1], [5], true)
    ]
  },
  {
    slug: 'longest-substring-without-repeating-characters',
    description_markdown:
      'Return the length of the longest substring that contains no repeated characters.',
    execution_spec: functionSpec('lengthOfLongestSubstring', [['s', 'string']], 'int'),
    test_cases: [
      testCase('repeating middle', ['abcabcbb'], 3),
      testCase('all same', ['bbbbb'], 1),
      testCase('mixed repeat', ['pwwkew'], 3),
      testCase('empty string', [''], 0, true),
      testCase('space included', ['dvdf'], 3, true)
    ]
  },
  {
    slug: 'container-with-most-water',
    description_markdown:
      'Choose two vertical lines that together with the x-axis can contain the most water.',
    execution_spec: functionSpec('maxArea', [['height', 'int[]']], 'int'),
    test_cases: [
      testCase('classic case', [[1, 8, 6, 2, 5, 4, 8, 3, 7]], 49),
      testCase('two lines', [[1, 1]], 1),
      testCase('short tall pair', [[4, 3, 2, 1, 4]], 16, true),
      testCase('best inner width', [[1, 2, 1]], 2, true)
    ]
  },
  {
    slug: 'search-in-rotated-sorted-array',
    description_markdown:
      'Search a target value in a rotated sorted array with distinct values and return its index, or -1 when absent.',
    execution_spec: functionSpec(
      'search',
      [
        ['nums', 'int[]'],
        ['target', 'int']
      ],
      'int'
    ),
    test_cases: [
      testCase('found after rotation', [[4, 5, 6, 7, 0, 1, 2], 0], 4),
      testCase('missing target', [[4, 5, 6, 7, 0, 1, 2], 3], -1),
      testCase('single missing', [[1], 0], -1),
      testCase('single found', [[1], 1], 0, true)
    ]
  },
  {
    slug: 'find-minimum-in-rotated-sorted-array',
    description_markdown:
      'Return the minimum element from a sorted array that may have been rotated.',
    execution_spec: functionSpec('findMin', [['nums', 'int[]']], 'int'),
    test_cases: [
      testCase('rotated once', [[3, 4, 5, 1, 2]], 1),
      testCase('rotated near end', [[4, 5, 6, 7, 0, 1, 2]], 0),
      testCase('not rotated', [[11, 13, 15, 17]], 11),
      testCase('single value', [[2]], 2, true)
    ]
  },
  {
    slug: 'binary-search',
    description_markdown:
      'Return the index of a target in a sorted integer array, or -1 when the target is not present.',
    execution_spec: functionSpec(
      'search',
      [
        ['nums', 'int[]'],
        ['target', 'int']
      ],
      'int'
    ),
    test_cases: [
      testCase('found target', [[-1, 0, 3, 5, 9, 12], 9], 4),
      testCase('missing target', [[-1, 0, 3, 5, 9, 12], 2], -1),
      testCase('first element', [[2, 5], 2], 0, true),
      testCase('last element', [[2, 5], 5], 1, true)
    ]
  },
  {
    slug: 'climbing-stairs',
    description_markdown:
      'Count how many distinct ways there are to climb n steps when each move can climb one or two steps.',
    execution_spec: functionSpec('climbStairs', [['n', 'int']], 'int'),
    test_cases: [
      testCase('two steps', [2], 2),
      testCase('three steps', [3], 3),
      testCase('one step', [1], 1, true),
      testCase('five steps', [5], 8, true)
    ]
  },
  {
    slug: 'coin-change',
    description_markdown:
      'Return the fewest number of coins needed to make up an amount, or -1 if no combination can make it.',
    execution_spec: functionSpec(
      'coinChange',
      [
        ['coins', 'int[]'],
        ['amount', 'int']
      ],
      'int'
    ),
    test_cases: [
      testCase('standard change', [[1, 2, 5], 11], 3),
      testCase('impossible', [[2], 3], -1),
      testCase('zero amount', [[1], 0], 0),
      testCase('larger combination', [[1, 3, 4], 6], 2, true)
    ]
  },
  {
    slug: 'house-robber',
    description_markdown:
      'Maximize money robbed from a line of houses without robbing two adjacent houses.',
    execution_spec: functionSpec('rob', [['nums', 'int[]']], 'int'),
    test_cases: [
      testCase('skip middle', [[1, 2, 3, 1]], 4),
      testCase('choose separated', [[2, 7, 9, 3, 1]], 12),
      testCase('single house', [[5]], 5, true),
      testCase('two houses', [[2, 1]], 2, true)
    ]
  },
  {
    slug: 'house-robber-ii',
    description_markdown:
      'Maximize money robbed from houses arranged in a circle, where the first and last houses are adjacent.',
    execution_spec: functionSpec('rob', [['nums', 'int[]']], 'int'),
    test_cases: [
      testCase('three houses', [[2, 3, 2]], 3),
      testCase('four houses', [[1, 2, 3, 1]], 4),
      testCase('single house', [[1]], 1, true),
      testCase('larger circle', [[1, 2, 3]], 3, true)
    ]
  },
  {
    slug: 'number-of-islands',
    description_markdown:
      'Count connected groups of land cells in a grid, using horizontal and vertical adjacency.',
    execution_spec: functionSpec('numIslands', [['grid', 'string[][]']], 'int'),
    test_cases: [
      testCase(
        'one island',
        [[['1', '1', '1', '1', '0'], ['1', '1', '0', '1', '0'], ['1', '1', '0', '0', '0'], ['0', '0', '0', '0', '0']]],
        1
      ),
      testCase(
        'three islands',
        [[['1', '1', '0', '0', '0'], ['1', '1', '0', '0', '0'], ['0', '0', '1', '0', '0'], ['0', '0', '0', '1', '1']]],
        3
      ),
      testCase('empty water', [[['0']]], 0, true),
      testCase('single land', [[['1']]], 1, true)
    ]
  },
  {
    slug: 'max-area-of-island',
    description_markdown:
      'Return the largest area of a connected group of land cells in a binary grid.',
    execution_spec: functionSpec('maxAreaOfIsland', [['grid', 'int[][]']], 'int'),
    test_cases: [
      testCase('mixed grid', [[[0, 0, 1, 0, 0], [0, 1, 1, 1, 0], [0, 0, 1, 0, 0]]], 5),
      testCase('no land', [[[0, 0], [0, 0]]], 0),
      testCase('one cell', [[[1]]], 1, true),
      testCase('separate islands', [[[1, 0, 1], [1, 0, 0]]], 2, true)
    ]
  },
  {
    slug: 'flood-fill',
    description_markdown:
      'Starting from a pixel, recolor its connected component of the original color using four-direction adjacency.',
    execution_spec: functionSpec(
      'floodFill',
      [
        ['image', 'int[][]'],
        ['sr', 'int'],
        ['sc', 'int'],
        ['color', 'int']
      ],
      'int[][]'
    ),
    test_cases: [
      testCase('fills center component', [[[1, 1, 1], [1, 1, 0], [1, 0, 1]], 1, 1, 2], [[2, 2, 2], [2, 2, 0], [2, 0, 1]]),
      testCase('same color no change', [[[0, 0, 0], [0, 0, 0]], 0, 0, 0], [[0, 0, 0], [0, 0, 0]]),
      testCase('single cell', [[[1]], 0, 0, 3], [[3]], true)
    ]
  },
  {
    slug: 'rotting-oranges',
    description_markdown:
      'Return the minutes needed for rotten oranges to rot all reachable fresh oranges, or -1 if some fresh orange remains.',
    execution_spec: functionSpec('orangesRotting', [['grid', 'int[][]']], 'int'),
    test_cases: [
      testCase('all rot eventually', [[[2, 1, 1], [1, 1, 0], [0, 1, 1]]], 4),
      testCase('blocked fresh orange', [[[2, 1, 1], [0, 1, 1], [1, 0, 1]]], -1),
      testCase('no fresh oranges', [[[0, 2]]], 0),
      testCase('single fresh unreachable', [[[1]]], -1, true)
    ]
  },
  {
    slug: 'course-schedule',
    description_markdown:
      'Determine whether all courses can be completed given prerequisite pairs.',
    execution_spec: functionSpec(
      'canFinish',
      [
        ['numCourses', 'int'],
        ['prerequisites', 'int[][]']
      ],
      'bool'
    ),
    test_cases: [
      testCase('simple possible', [2, [[1, 0]]], true),
      testCase('simple cycle', [2, [[1, 0], [0, 1]]], false),
      testCase('chain possible', [4, [[1, 0], [2, 1], [3, 2]]], true, true),
      testCase('larger cycle', [3, [[0, 1], [1, 2], [2, 0]]], false, true)
    ]
  },
  {
    slug: 'course-schedule-ii',
    description_markdown:
      'Return an ordering of courses that satisfies prerequisites, or an empty array if no valid ordering exists.',
    execution_spec: functionSpec(
      'findOrder',
      [
        ['numCourses', 'int'],
        ['prerequisites', 'int[][]']
      ],
      'int[]'
    ),
    test_cases: [
      testCase('two courses', [2, [[1, 0]]], [0, 1]),
      testCase('impossible cycle', [2, [[1, 0], [0, 1]]], []),
      testCase('linear chain', [3, [[1, 0], [2, 1]]], [0, 1, 2], true)
    ]
  },
  {
    slug: 'longest-consecutive-sequence',
    description_markdown:
      'Return the length of the longest run of consecutive integer values in an unsorted array.',
    execution_spec: functionSpec('longestConsecutive', [['nums', 'int[]']], 'int'),
    test_cases: [
      testCase('classic run', [[100, 4, 200, 1, 3, 2]], 4),
      testCase('longer mixed run', [[0, 3, 7, 2, 5, 8, 4, 6, 0, 1]], 9),
      testCase('empty array', [[]], 0, true),
      testCase('duplicates only', [[1, 1, 1]], 1, true)
    ]
  },
  {
    slug: 'merge-intervals',
    description_markdown:
      'Merge all overlapping intervals and return the compact list of disjoint intervals.',
    execution_spec: functionSpec('merge', [['intervals', 'int[][]']], 'int[][]'),
    test_cases: [
      testCase('overlap first pair', [[[1, 3], [2, 6], [8, 10], [15, 18]]], [[1, 6], [8, 10], [15, 18]]),
      testCase('touching intervals', [[[1, 4], [4, 5]]], [[1, 5]]),
      testCase('single interval', [[[1, 4]]], [[1, 4]], true),
      testCase('nested interval', [[[1, 4], [2, 3]]], [[1, 4]], true)
    ]
  },
  {
    slug: 'insert-interval',
    description_markdown:
      'Insert a new interval into a sorted non-overlapping interval list and merge overlaps.',
    execution_spec: functionSpec(
      'insert',
      [
        ['intervals', 'int[][]'],
        ['newInterval', 'int[]']
      ],
      'int[][]'
    ),
    test_cases: [
      testCase('insert between', [[[1, 3], [6, 9]], [2, 5]], [[1, 5], [6, 9]]),
      testCase('merge many', [[[1, 2], [3, 5], [6, 7], [8, 10], [12, 16]], [4, 8]], [[1, 2], [3, 10], [12, 16]]),
      testCase('empty intervals', [[], [5, 7]], [[5, 7]], true),
      testCase('append interval', [[[1, 2]], [3, 4]], [[1, 2], [3, 4]], true)
    ]
  },
  {
    slug: 'non-overlapping-intervals',
    description_markdown:
      'Return the minimum number of intervals to remove so the remaining intervals do not overlap.',
    execution_spec: functionSpec('eraseOverlapIntervals', [['intervals', 'int[][]']], 'int'),
    test_cases: [
      testCase('remove one', [[[1, 2], [2, 3], [3, 4], [1, 3]]], 1),
      testCase('remove duplicates', [[[1, 2], [1, 2], [1, 2]]], 2),
      testCase('already disjoint', [[[1, 2], [2, 3]]], 0),
      testCase('nested overlaps', [[[1, 100], [11, 22], [1, 11], [2, 12]]], 2, true)
    ]
  }
]

const manualBySlug = new Map(manualProblems.map((problem) => [problem.slug, problem]))

function functionSpec(entrypoint, params, returns, comparator = { kind: 'exact' }) {
  return {
    mode: 'function',
    entrypoint,
    signature: {
      params: params.map(([name, type]) => ({ name, type })),
      returns
    },
    comparator,
    ...defaultLimits
  }
}

function testCase(name, args, expected, isHidden = false) {
  return {
    name,
    args,
    expected,
    is_hidden: isHidden
  }
}

function legacyInput(spec, args) {
  const out = {}
  for (const [index, param] of spec.signature.params.entries()) {
    out[param.name] = args[index]
  }
  return out
}

function toRegistryTestCase(problem, item) {
  const input = legacyInput(problem.execution_spec, item.args)
  return {
    name: item.name,
    input: JSON.stringify(input),
    expected: JSON.stringify(item.expected),
    input_json: { args: item.args },
    expected_json: item.expected,
    metadata: { source: 'manual' },
    is_hidden: item.is_hidden
  }
}

function inferFunctionSpec(problem) {
  const python = problem.starter_code?.python || ''
  const solutionStart = python.indexOf('class Solution:')
  if (solutionStart === -1) {
    return null
  }

  const solutionBody = python.slice(solutionStart)
  const match = solutionBody.match(
    /\n\s{4}def\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*(?:->\s*([^:]+))?:/
  )
  if (!match || match[1] === '__init__') {
    return null
  }

  const params = parsePythonParams(match[2])
  if (params.length === 0) {
    return null
  }

  return {
    mode: 'function',
    entrypoint: match[1],
    signature: {
      params,
      returns: normalizePythonType(match[3] || 'any')
    },
    comparator: { kind: 'exact' },
    ...defaultLimits
  }
}

function parsePythonParams(raw) {
  return splitTopLevel(raw, ',')
    .map((param) => param.trim())
    .filter((param) => param && param !== 'self')
    .map((param) => {
      const withoutDefault = splitTopLevel(param, '=')[0].trim()
      const [name, type = 'any'] = splitTopLevel(withoutDefault, ':').map((part) =>
        part.trim()
      )
      return { name, type: normalizePythonType(type) }
    })
    .filter((param) => param.name && /^[A-Za-z_][A-Za-z0-9_]*$/.test(param.name))
}

function splitTopLevel(value, delimiter) {
  const out = []
  let start = 0
  let depth = 0
  for (let i = 0; i < value.length; i += 1) {
    const ch = value[i]
    if (ch === '[' || ch === '(' || ch === '{') {
      depth += 1
    } else if (ch === ']' || ch === ')' || ch === '}') {
      depth -= 1
    } else if (ch === delimiter && depth === 0) {
      out.push(value.slice(start, i))
      start = i + 1
    }
  }
  out.push(value.slice(start))
  return out
}

function normalizePythonType(type) {
  const value = type.trim().replace(/^typing\./, '')
  const aliases = new Map([
    ['int', 'int'],
    ['str', 'string'],
    ['bool', 'bool'],
    ['float', 'float'],
    ['None', 'void'],
    ['NoneType', 'void'],
    ['list[int]', 'int[]'],
    ['List[int]', 'int[]'],
    ['list[str]', 'string[]'],
    ['List[str]', 'string[]'],
    ['list[bool]', 'bool[]'],
    ['List[bool]', 'bool[]'],
    ['list[list[int]]', 'int[][]'],
    ['List[List[int]]', 'int[][]'],
    ['list[list[str]]', 'string[][]'],
    ['List[List[str]]', 'string[][]']
  ])
  if (aliases.has(value)) {
    return aliases.get(value)
  }

  const optional = value.match(/^Optional\[(.*)]$/)
  if (optional) {
    return `${normalizePythonType(optional[1])}?`
  }
  return value || 'any'
}

function summaryDescription(problem, spec) {
  const tags = (problem.tags || []).slice(0, 4).join(', ')
  const params = spec?.signature?.params?.map((param) => param.name).join(', ')
  const suffix = tags ? ` Main tags: ${tags}.` : ''
  if (spec?.entrypoint && params) {
    return `Solve ${problem.title} by implementing ${spec.entrypoint}(${params}). This local registry entry records the callable signature for judging and keeps the statement summary intentionally short.${suffix}`
  }
  return `Solve ${problem.title}. This local registry entry keeps a short summary and metadata for organizing practice, while the full statement remains linked from the source URL.${suffix}`
}

function updateIndexChecksum(index, name, providerJSON) {
  const checksum = createHash('sha256').update(providerJSON).digest('hex')
  index.manifests = index.manifests.map((manifest) => {
    if (manifest.name !== name) {
      return manifest
    }
    return { ...manifest, checksum: `sha256:${checksum}` }
  })
  return checksum
}

const freeProvider = JSON.parse(await readFile(freeProviderPath, 'utf8'))

let enriched = 0
let inferredSpecs = 0
let descriptions = 0
freeProvider.problems = freeProvider.problems.map((problem) => {
  const manual = manualBySlug.get(problem.slug)
  if (manual) {
    enriched += 1
    return {
      ...problem,
      description_markdown: manual.description_markdown,
      execution_spec: manual.execution_spec,
      test_cases: manual.test_cases.map((item) => toRegistryTestCase(manual, item)),
      judge_ready: true,
      version: Math.max(problem.version || 1, 2)
    }
  }

  const inferred = inferFunctionSpec(problem)
  const next = { ...problem, judge_ready: Boolean(problem.test_cases?.length) }
  if (!next.description_markdown) {
    next.description_markdown = summaryDescription(problem, inferred)
    descriptions += 1
  }
  if (!next.execution_spec?.mode && inferred) {
    next.execution_spec = inferred
    inferredSpecs += 1
  }
  if (!next.test_cases?.length) {
    next.test_cases = []
    next.judge_ready = false
  }
  if (next.description_markdown !== problem.description_markdown || next.execution_spec !== problem.execution_spec) {
    next.version = Math.max(problem.version || 1, 2)
  }
  return next
})

freeProvider.updated_at = new Date().toISOString()

const index = JSON.parse(await readFile(indexPath, 'utf8'))
if (process.argv[2]) {
  index.version = process.argv[2]
}
index.updated_at = new Date().toISOString()

const providerJSON = `${JSON.stringify(freeProvider, null, 2)}\n`
const checksum = updateIndexChecksum(index, 'leetcode', providerJSON)

await writeFile(freeProviderPath, providerJSON)
await writeFile(indexPath, `${JSON.stringify(index, null, 2)}\n`)

console.log(`Updated ${freeProviderPath}`)
console.log(`Manually enriched LeetCode problems with test cases: ${enriched}`)
console.log(`Inferred function specs: ${inferredSpecs}`)
console.log(`Generated summaries: ${descriptions}`)
console.log(`Checksum: sha256:${checksum}`)
