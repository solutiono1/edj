#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

if [ -z "${EDJ_API_BASE:-}" ]; then
  echo "Set EDJ_API_BASE to the Pagor API URL"
  exit 1
fi

printf "window.EDJ_API_BASE = '%s';\n" "$EDJ_API_BASE" > web/static/config.js
npx wrangler pages deploy web --project-name pagor
