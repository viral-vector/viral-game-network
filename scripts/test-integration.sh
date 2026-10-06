#!/bin/sh
set -eu

cd "$(dirname "$0")/.."
if [ -z "${VGN_TEST_STORE_ENDPOINT:-}" ]; then
  # Separate from production Compose; no volumes or production credentials.
  compose() { docker compose -p vgn-tests -f docker-compose.test.yaml "$@"; }
  trap 'compose down --volumes' EXIT
  compose up --detach --wait
  export VGN_TEST_STORE_ENDPOINT=ws://127.0.0.1:18000/rpc
fi

export VNET_KEY=test-signing-key
export VNET_TOKEN_EXPIRE=60
"${GO:-go}" test -tags integration "$@" ./src/... ./tests/...
