#!/bin/sh
set -eu

UI_DIR="${UI_DIR:-/workspace/ui}"
LOCKFILE_STAMP="node_modules/.package-lock.json"

cd "$UI_DIR"

if [ ! -d node_modules ] || [ ! -f "$LOCKFILE_STAMP" ] || ! cmp -s package-lock.json "$LOCKFILE_STAMP"; then
  npm ci
  cp package-lock.json "$LOCKFILE_STAMP"
fi

exec "$@"
