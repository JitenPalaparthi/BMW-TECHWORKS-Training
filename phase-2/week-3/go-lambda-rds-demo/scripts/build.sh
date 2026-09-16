#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap ./cmd/lambda
rm -f function.zip
zip -q function.zip bootstrap
rm bootstrap
echo "Built $ROOT/function.zip"
