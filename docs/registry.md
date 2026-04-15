# Problem Registry

## Overview

The problem registry is a versioned index of problem metadata from multiple providers.
It is inspired by Mason's registry pattern: a central `index.json` points to provider and track manifests.

**Important:** Provider statements are not fetched automatically. The registry stays metadata-first, but a problem may include an optional local `description_markdown` field for an author-written prompt.

The bundled LeetCode provider data is still metadata-first and split into free and premium manifests. The free manifest is the bank imported by sync; the premium manifest is retained as metadata only. Free entries include Python, Go, JavaScript, TypeScript, and Rust starter snippets when LeetCode exposes them; entries without snippets keep `starter_code: {}` so editor clients can fall back to local templates. Optional `description_markdown` content is intended for locally authored prompts, not scraped provider statements.

## Registry structure

```
registry/
  index.json              # versioned list of manifests
  providers/
    leetcode/
      free/problems.json   # Free LeetCode problem manifests used as the bank
      premium/problems.json # Premium LeetCode metadata, not imported into the bank
    judge-ready/
      blind75.json          # Local authored runnable overlays with test cases
    neetcode.json          # NeetCode curated list
    hackerrank.json        # HackerRank problem manifests
  tracks/
    blind75.json           # Blind 75 track
    neetcode150.json       # NeetCode 150 track
    patterns.json          # Pattern-based track
  roadmaps/
    interview-prep.json    # Interview prep roadmap
    dsa-foundations.json   # DSA foundations roadmap
```

## index.json

```json
{
  "version": "1.0.3",
  "updated_at": "2026-01-01T00:00:00Z",
  "manifests": [
    {
      "name": "leetcode",
      "path": "providers/leetcode/free/problems.json",
      "checksum": "sha256:..."
    },
    {
      "name": "neetcode",
      "path": "providers/neetcode.json",
      "checksum": "sha256:..."
    },
    {
      "name": "blind75",
      "path": "tracks/blind75.json",
      "checksum": "sha256:..."
    }
  ]
}
```

## ProblemManifest

Each entry in a provider manifest:

```json
{
  "provider": "leetcode",
  "external_id": "1",
  "slug": "two-sum",
  "title": "Two Sum",
  "difficulty": "easy",
  "tags": ["array", "hash-table", "lookup", "two-pointer"],
  "source_url": "https://leetcode.com/problems/two-sum",
  "estimated_time": 15,
  "description_markdown": "## Two Sum\n\nReturn indices for the pair that adds to target.",
  "starter_code": {
    "python": "import json\nimport sys\n\n..."
  },
  "test_cases": [
    {
      "input": "{\"nums\":[2,7,11,15],\"target\":9}",
      "expected": "[0,1]"
    },
    {
      "input": "{\"nums\":[3,2,4],\"target\":6}",
      "expected": "[1,2]",
      "is_hidden": true
    }
  ],
  "version": 1
}
```

Fields:

- `provider` — source platform
- `external_id` — provider's own problem ID
- `slug` — URL-safe identifier, unique within provider
- `title` — display name
- `difficulty` — easy | medium | hard
- `tags` — combined topic and problem-solving tags (e.g. array, sliding-window, two-pointer)
- `source_url` — link to original problem
- `estimated_time` — minutes, rough estimate
- `description_markdown` — optional local Markdown prompt authored in judge-loop
- `starter_code` — optional language-specific starter code
- `test_cases` — optional locally authored stdin/stdout cases imported into the judge store during registry sync
- `version` — manifest version, incremented on metadata changes

For backward compatibility, registry ingestion still accepts legacy `pattern_tags` and merges them into `tags`.

For judge-ready simple-mode problems, prefer JSON strings in `test_cases.input` and `test_cases.expected`. The starter program reads JSON from stdin and prints JSON to stdout; the judge compares valid JSON semantically.

## Track manifest

```json
{
  "name": "blind75",
  "title": "Blind 75",
  "description": "75 essential interview problems",
  "problems": [
    { "provider": "leetcode", "slug": "two-sum", "order": 1 },
    {
      "provider": "leetcode",
      "slug": "best-time-to-buy-and-sell-stock",
      "order": 2
    }
  ]
}
```

## Roadmap manifest

```json
{
  "name": "interview-prep",
  "title": "Interview Prep",
  "stages": [
    {
      "name": "arrays-and-hashing",
      "title": "Arrays & Hashing",
      "track": "neetcode150",
      "problems": ["two-sum", "contains-duplicate"]
    }
  ]
}
```

## Sync flow

1. `POST /local/sync` triggers the local agent
2. Agent fetches `index.json` from server
3. Compares versions: skip manifests with matching checksum
4. Downloads changed manifests
5. Upserts problems into local problem bank (SQLite)
6. Returns sync summary

## Updating LeetCode metadata

Use:

```bash
scripts/update_leetcode_registry.sh
```

The updater pages LeetCode public metadata, splits free and premium entries, fetches starter snippets for Python, Go, JavaScript, TypeScript, and Rust for free entries in slow batches, writes `registry/providers/leetcode/free/problems.json` and `registry/providers/leetcode/premium/problems.json`, and updates the LeetCode checksum in `registry/index.json` for the free bank manifest.

It does not fetch problem statements, editorials, solutions, or test cases. Paid-only entries returned by the listing endpoint are written to the premium metadata manifest only; that premium manifest is not referenced by `registry/index.json`, so local registry sync does not import those problems into the bank.

Test cases are intentionally not bulk-imported from LeetCode. The public detail metadata exposes example inputs, not expected outputs suitable for local judging, and the current sandbox executes whole programs from stdin rather than LeetCode-style function/class snippets. Curated problems can still add local judge cases through `POST /api/problems/contribute`.

## Updating curated tracks

Use:

```bash
scripts/update_curated_tracks.mjs
```

The updater writes Blind 75 and NeetCode 150 track manifests from canonical slug lists, filters them against the local free LeetCode provider manifest, and updates track checksums in `registry/index.json`.

Because the provider manifest is free-only, premium-only track entries are omitted instead of creating broken references. Current available counts are Blind 75: 70/75 and NeetCode 150: 143/150.

## Judge-ready overlays

Use:

```bash
node scripts/update_judge_ready_seed.mjs
node scripts/validate_judge_ready.mjs
```

`update_judge_ready_seed.mjs` writes the local authored Blind 75 overlay under `registry/providers/judge-ready/blind75.json` and adds it to `registry/index.json` after the free LeetCode bank. During sync, later overlay entries upsert the same LeetCode `(provider, external_id)` rows and import their local `test_cases`.

`validate_judge_ready.mjs` checks that every judge-ready overlay references a free LeetCode slug, includes all supported starter languages, and stores valid JSON strings in every test case input and expected output.

## RegistryVersion

Tracks which version of the registry the user has locally:

```go
type RegistryVersion struct {
    Version    string
    UpdatedAt  time.Time
    Manifests  []ManifestRef
}
```
