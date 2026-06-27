#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
node "$ROOT_DIR/scripts/update_leetcode_registry.mjs" "$@"
node "$ROOT_DIR/scripts/enrich_leetcode_registry.mjs"
exec node "$ROOT_DIR/scripts/enrich_leetcode_examples.mjs"
