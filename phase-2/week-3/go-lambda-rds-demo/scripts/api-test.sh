#!/usr/bin/env bash
set -euo pipefail
URL="${1:-${FUNCTION_URL:-}}"
if [[ -z "$URL" ]]; then echo "Usage: $0 https://...lambda-url.../"; exit 1; fi
URL="${URL%/}"
echo "== Health =="
curl -fsS "$URL/health"; echo
echo "== Create user =="
EMAIL="jiten.$(date +%s)@example.com"
curl -fsS -X POST "$URL/users" -H 'content-type: application/json' -d "{\"name\":\"Jiten\",\"email\":\"$EMAIL\"}"; echo
echo "== List users =="
curl -fsS "$URL/users"; echo
