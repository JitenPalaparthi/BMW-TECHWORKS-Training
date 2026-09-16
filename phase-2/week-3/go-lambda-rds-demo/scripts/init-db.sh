#!/usr/bin/env bash
set -euo pipefail
: "${DB_HOST:?Set DB_HOST}"
: "${DB_NAME:?Set DB_NAME}"
: "${DB_USER:?Set DB_USER}"
: "${DB_PASSWORD:?Set DB_PASSWORD}"
export PGPASSWORD="$DB_PASSWORD"
psql "host=$DB_HOST port=${DB_PORT:-5432} dbname=$DB_NAME user=$DB_USER sslmode=${DB_SSLMODE:-require}" -f "$(dirname "$0")/../sql/schema.sql"
