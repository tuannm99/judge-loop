# Problem Registry

## Overview

The problem registry is a versioned index of problem metadata from multiple providers.
It is inspired by Mason's registry pattern: a central `index.json` points to provider and track manifests.

**Important:** Problem statements are NOT stored. Only metadata (manifest) is stored. Full problem descriptions remain on the provider's platform.

The bundled LeetCode provider manifest is metadata-only and filters out paid-only problems. It intentionally stores `starter_code: {}` for bulk-imported entries; editor clients provide local fallback templates when provider starter snippets are unavailable.

## Registry structure

```
registry/
  index.json              # versioned list of manifests
  providers/
    leetcode.json          # LeetCode problem manifests
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
      "path": "providers/leetcode.json",
      "checksum": "sha256:..."
    },
    {
      "name": "neetcode",
      "path": "providers/neetcode.json",
      "checksum": "sha256:..."
    },
    {"name": "blind75", "path": "tracks/blind75.json", "checksum": "sha256:..."}
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
  "tags": ["array", "hash-table"],
  "pattern_tags": ["lookup", "two-pointer"],
  "source_url": "https://leetcode.com/problems/two-sum",
  "estimated_time": 15,
  "version": 1
}
```

Fields:

- `provider` — source platform
- `external_id` — provider's own problem ID
- `slug` — URL-safe identifier, unique within provider
- `title` — display name
- `difficulty` — easy | medium | hard
- `tags` — data structure / algorithm tags
- `pattern_tags` — problem-solving pattern tags (e.g. sliding-window, two-pointer)
- `source_url` — link to original problem
- `estimated_time` — minutes, rough estimate
- `version` — manifest version, incremented on metadata changes

## Track manifest

```json
{
  "name": "blind75",
  "title": "Blind 75",
  "description": "75 essential interview problems",
  "problems": [
    {"provider": "leetcode", "slug": "two-sum", "order": 1},
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

The updater pages LeetCode public metadata, keeps only entries where `paidOnly` is false, writes `registry/providers/leetcode.json`, and updates the LeetCode checksum in `registry/index.json`.

It does not fetch problem statements, editorials, solutions, or test cases. Paid-only entries returned by the listing endpoint are filtered out and are not written to the local registry.

## RegistryVersion

Tracks which version of the registry the user has locally:

```go
type RegistryVersion struct {
    Version    string
    UpdatedAt  time.Time
    Manifests  []ManifestRef
}
```
