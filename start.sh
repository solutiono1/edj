#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

# Load optional secrets (GOOGLE_CLIENT_ID) from .env, never committed.
if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

export EDJ_RECEIVE_ADDRESS="${EDJ_RECEIVE_ADDRESS:-bitcoincash:qrzp600axt2j5e92ardqg8sas9tvdn7h4sgxftwggj}"
export EDJ_TOKEN_ADDRESS="${EDJ_TOKEN_ADDRESS:-bitcoincash:zrzp600axt2j5e92ardqg8sas9tvdn7h4s0v64qwhp}"
export ADDR="${ADDR:-:8080}"

exec /tmp/opencode/edj-server
