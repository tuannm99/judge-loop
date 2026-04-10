#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="${TMPDIR:-/tmp}/judge-loop-leetcode-registry"
OUT_FILE="$ROOT_DIR/registry/providers/leetcode.json"
INDEX_FILE="$ROOT_DIR/registry/index.json"
VERSION="${1:-$(date -u +%Y-%m-%d)-leetcode-free}"
UPDATED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

mkdir -p "$TMP_DIR"
rm -f "$TMP_DIR"/page-*.json

query='query problemsetQuestionList($categorySlug: String, $limit: Int, $skip: Int, $filters: QuestionListFilterInput) { problemsetQuestionList: questionList(categorySlug: $categorySlug, limit: $limit, skip: $skip, filters: $filters) { total: totalNum questions: data { difficulty frontendQuestionId: questionFrontendId paidOnly: isPaidOnly title titleSlug topicTags { name slug } } } }'

total=1
skip=0
limit=100

while [ "$skip" -lt "$total" ]; do
  payload="$(jq -n \
    --arg query "$query" \
    --argjson skip "$skip" \
    --argjson limit "$limit" \
    '{query:$query, variables:{categorySlug:"", skip:$skip, limit:$limit, filters:{}}}')"

  curl -sS --max-time 30 "https://leetcode.com/graphql/" \
    -H "Content-Type: application/json" \
    --data "$payload" \
    -o "$TMP_DIR/page-$skip.json"

  total="$(jq -r '.data.problemsetQuestionList.total' "$TMP_DIR/page-$skip.json")"
  skip=$((skip + limit))
done

jq -S -s '
  {
    problems: ([
      .[].data.problemsetQuestionList.questions[]
      | select(.paidOnly == false)
      | {
          provider: "leetcode",
          external_id: .frontendQuestionId,
          slug: .titleSlug,
          title: .title,
          difficulty: (.difficulty | ascii_downcase),
          tags: ([.topicTags[]?.slug]),
          pattern_tags: ([.topicTags[]?.slug]),
          source_url: ("https://leetcode.com/problems/" + .titleSlug + "/"),
          estimated_time: (if .difficulty == "Easy" then 15 elif .difficulty == "Medium" then 30 else 45 end),
          starter_code: {},
          version: 1
        }
    ] | sort_by(.external_id | tonumber))
  }
' "$TMP_DIR"/page-*.json > "$OUT_FILE"

checksum="$(sha256sum "$OUT_FILE" | awk '{print $1}')"
tmp_index="$TMP_DIR/index.json"

jq \
  --arg version "$VERSION" \
  --arg updated_at "$UPDATED_AT" \
  --arg checksum "sha256:$checksum" \
  '.version = $version
   | .updated_at = $updated_at
   | .manifests = (.manifests | map(if .name == "leetcode" then .checksum = $checksum else . end))' \
  "$INDEX_FILE" > "$tmp_index"
mv "$tmp_index" "$INDEX_FILE"

count="$(jq -r '.problems | length' "$OUT_FILE")"
echo "Updated $OUT_FILE with $count free LeetCode problems"
echo "Updated $INDEX_FILE version=$VERSION checksum=sha256:$checksum"
