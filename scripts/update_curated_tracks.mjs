#!/usr/bin/env node
import { createHash } from 'node:crypto'
import { readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'

const rootDir = path.resolve(import.meta.dirname, '..')
const providerPath = path.join(rootDir, 'registry/providers/leetcode/free/problems.json')
const indexPath = path.join(rootDir, 'registry/index.json')

const tracks = {
  blind75: {
    title: 'Blind 75',
    description:
      'Blind 75 interview track, limited to entries available in the local free LeetCode registry.',
    slugs: [
      'contains-duplicate',
      'valid-anagram',
      'two-sum',
      'group-anagrams',
      'top-k-frequent-elements',
      'product-of-array-except-self',
      'valid-sudoku',
      'longest-consecutive-sequence',
      'valid-palindrome',
      '3sum',
      'container-with-most-water',
      'best-time-to-buy-and-sell-stock',
      'longest-substring-without-repeating-characters',
      'longest-repeating-character-replacement',
      'minimum-window-substring',
      'valid-parentheses',
      'find-minimum-in-rotated-sorted-array',
      'search-in-rotated-sorted-array',
      'reverse-linked-list',
      'merge-two-sorted-lists',
      'reorder-list',
      'remove-nth-node-from-end-of-list',
      'linked-list-cycle',
      'merge-k-sorted-lists',
      'invert-binary-tree',
      'maximum-depth-of-binary-tree',
      'same-tree',
      'subtree-of-another-tree',
      'lowest-common-ancestor-of-a-binary-search-tree',
      'binary-tree-level-order-traversal',
      'validate-binary-search-tree',
      'kth-smallest-element-in-a-bst',
      'construct-binary-tree-from-preorder-and-inorder-traversal',
      'binary-tree-maximum-path-sum',
      'serialize-and-deserialize-binary-tree',
      'find-median-from-data-stream',
      'combination-sum',
      'word-search',
      'implement-trie-prefix-tree',
      'design-add-and-search-words-data-structure',
      'word-search-ii',
      'number-of-islands',
      'clone-graph',
      'pacific-atlantic-water-flow',
      'course-schedule',
      'graph-valid-tree',
      'number-of-connected-components-in-an-undirected-graph',
      'alien-dictionary',
      'climbing-stairs',
      'house-robber',
      'house-robber-ii',
      'longest-palindromic-substring',
      'palindromic-substrings',
      'decode-ways',
      'coin-change',
      'maximum-product-subarray',
      'word-break',
      'longest-increasing-subsequence',
      'unique-paths',
      'longest-common-subsequence',
      'maximum-subarray',
      'jump-game',
      'insert-interval',
      'merge-intervals',
      'non-overlapping-intervals',
      'meeting-rooms',
      'meeting-rooms-ii',
      'rotate-image',
      'spiral-matrix',
      'set-matrix-zeroes',
      'sum-of-two-integers',
      'number-of-1-bits',
      'counting-bits',
      'missing-number',
      'reverse-bits'
    ]
  },
  neetcode150: {
    title: 'NeetCode 150',
    description:
      'NeetCode 150 interview track, limited to entries available in the local free LeetCode registry.',
    slugs: [
      'contains-duplicate',
      'valid-anagram',
      'two-sum',
      'group-anagrams',
      'top-k-frequent-elements',
      'product-of-array-except-self',
      'valid-sudoku',
      'encode-and-decode-strings',
      'longest-consecutive-sequence',
      'valid-palindrome',
      'two-sum-ii-input-array-is-sorted',
      '3sum',
      'container-with-most-water',
      'trapping-rain-water',
      'best-time-to-buy-and-sell-stock',
      'longest-substring-without-repeating-characters',
      'longest-repeating-character-replacement',
      'permutation-in-string',
      'minimum-window-substring',
      'sliding-window-maximum',
      'valid-parentheses',
      'min-stack',
      'evaluate-reverse-polish-notation',
      'generate-parentheses',
      'daily-temperatures',
      'car-fleet',
      'largest-rectangle-in-histogram',
      'binary-search',
      'search-a-2d-matrix',
      'koko-eating-bananas',
      'find-minimum-in-rotated-sorted-array',
      'search-in-rotated-sorted-array',
      'time-based-key-value-store',
      'median-of-two-sorted-arrays',
      'reverse-linked-list',
      'merge-two-sorted-lists',
      'reorder-list',
      'remove-nth-node-from-end-of-list',
      'copy-list-with-random-pointer',
      'add-two-numbers',
      'linked-list-cycle',
      'find-the-duplicate-number',
      'lru-cache',
      'merge-k-sorted-lists',
      'reverse-nodes-in-k-group',
      'invert-binary-tree',
      'maximum-depth-of-binary-tree',
      'diameter-of-binary-tree',
      'balanced-binary-tree',
      'same-tree',
      'subtree-of-another-tree',
      'lowest-common-ancestor-of-a-binary-search-tree',
      'binary-tree-level-order-traversal',
      'binary-tree-right-side-view',
      'count-good-nodes-in-binary-tree',
      'validate-binary-search-tree',
      'kth-smallest-element-in-a-bst',
      'construct-binary-tree-from-preorder-and-inorder-traversal',
      'binary-tree-maximum-path-sum',
      'serialize-and-deserialize-binary-tree',
      'implement-trie-prefix-tree',
      'design-add-and-search-words-data-structure',
      'word-search-ii',
      'kth-largest-element-in-a-stream',
      'last-stone-weight',
      'k-closest-points-to-origin',
      'kth-largest-element-in-an-array',
      'task-scheduler',
      'design-twitter',
      'find-median-from-data-stream',
      'subsets',
      'combination-sum',
      'permutations',
      'subsets-ii',
      'combination-sum-ii',
      'word-search',
      'palindrome-partitioning',
      'letter-combinations-of-a-phone-number',
      'n-queens',
      'number-of-islands',
      'clone-graph',
      'max-area-of-island',
      'pacific-atlantic-water-flow',
      'surrounded-regions',
      'rotting-oranges',
      'walls-and-gates',
      'course-schedule',
      'course-schedule-ii',
      'redundant-connection',
      'number-of-connected-components-in-an-undirected-graph',
      'graph-valid-tree',
      'word-ladder',
      'reconstruct-itinerary',
      'min-cost-to-connect-all-points',
      'network-delay-time',
      'swim-in-rising-water',
      'alien-dictionary',
      'cheapest-flights-within-k-stops',
      'climbing-stairs',
      'min-cost-climbing-stairs',
      'house-robber',
      'house-robber-ii',
      'longest-palindromic-substring',
      'palindromic-substrings',
      'decode-ways',
      'coin-change',
      'maximum-product-subarray',
      'word-break',
      'longest-increasing-subsequence',
      'partition-equal-subset-sum',
      'unique-paths',
      'longest-common-subsequence',
      'best-time-to-buy-and-sell-stock-with-cooldown',
      'coin-change-ii',
      'target-sum',
      'interleaving-string',
      'longest-increasing-path-in-a-matrix',
      'distinct-subsequences',
      'edit-distance',
      'burst-balloons',
      'regular-expression-matching',
      'maximum-subarray',
      'jump-game',
      'jump-game-ii',
      'gas-station',
      'hand-of-straights',
      'merge-triplets-to-form-target-triplet',
      'partition-labels',
      'valid-parenthesis-string',
      'insert-interval',
      'merge-intervals',
      'non-overlapping-intervals',
      'meeting-rooms',
      'meeting-rooms-ii',
      'minimum-interval-to-include-each-query',
      'rotate-image',
      'spiral-matrix',
      'set-matrix-zeroes',
      'happy-number',
      'plus-one',
      'powx-n',
      'multiply-strings',
      'detect-squares',
      'single-number',
      'number-of-1-bits',
      'counting-bits',
      'reverse-bits',
      'missing-number',
      'sum-of-two-integers',
      'reverse-integer'
    ]
  }
}

const provider = JSON.parse(await readFile(providerPath, 'utf8'))
const availableSlugs = new Set(provider.problems.map((problem) => problem.slug))
const index = JSON.parse(await readFile(indexPath, 'utf8'))

for (const [name, track] of Object.entries(tracks)) {
  const seen = new Set()
  const duplicates = track.slugs.filter((slug) => {
    if (!seen.has(slug)) {
      seen.add(slug)
      return false
    }
    return true
  })
  if (duplicates.length > 0) {
    throw new Error(`${name} has duplicate slugs: ${duplicates.join(', ')}`)
  }

  const missing = track.slugs.filter((slug) => !availableSlugs.has(slug))
  const available = track.slugs.filter((slug) => availableSlugs.has(slug))
  const manifest = {
    name,
    title: track.title,
    description: `${track.description} Canonical count: ${track.slugs.length}. Available count: ${available.length}. Omitted unavailable slugs: ${missing.length ? missing.join(', ') : 'none'}.`,
    problems: available.map((slug, index) => ({
      provider: 'leetcode',
      slug,
      order: index + 1
    }))
  }

  const filePath = path.join(rootDir, `registry/tracks/${name}.json`)
  const body = `${JSON.stringify(manifest, null, 2)}\n`
  await writeFile(filePath, body)

  const checksum = createHash('sha256').update(body).digest('hex')
  index.manifests = index.manifests.map((ref) => {
    if (ref.name !== name) {
      return ref
    }
    return { ...ref, checksum: `sha256:${checksum}` }
  })
  console.log(
    `${name}: wrote ${available.length}/${track.slugs.length}, missing ${missing.length}`
  )
  if (missing.length > 0) {
    console.log(`  missing: ${missing.join(', ')}`)
  }
}

await writeFile(indexPath, `${JSON.stringify(index, null, 2)}\n`)
