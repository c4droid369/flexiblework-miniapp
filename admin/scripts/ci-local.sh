#!/usr/bin/env bash
# ci-local.sh — single command that mirrors CI. Run this before pushing.
#
# Backend:   fmtcheck + lint + unit + e2e (requires docker compose up)
# Frontend:  fmtcheck + lint + unit
#
# Usage:
#   chmod +x scripts/ci-local.sh   # one-time
#   ./scripts/ci-local.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

bold() { printf "\033[1m%s\033[0m\n" "$*"; }
pass() { printf "\033[32m✓ %s\033[0m\n" "$*"; }
fail() { printf "\033[31m✗ %s\033[0m\n" "$*" >&2; exit 1; }

# ----- Backend -------------------------------------------------------------
bold "[backend] fmtcheck"
( cd backend && gofumpt -l . | grep -q . ) && fail "gofumpt would reformat" || pass "gofumpt clean"
( cd backend && goimports -l -local github.com/admin-template/backend . | grep -q . ) && fail "goimports would reformat" || pass "goimports clean"

bold "[backend] lint"
( cd backend && golangci-lint run ./... ) && pass "lint clean"

bold "[backend] unit tests"
( cd backend && go test -shuffle=on -count=1 ./internal/... ) && pass "unit"

bold "[backend] e2e tests (requires docker compose up)"
( cd backend && go test -tags=e2e -count=1 -timeout=120s ./e2e/... ) && pass "e2e"

# ----- Frontend ------------------------------------------------------------
bold "[frontend] fmtcheck"
( cd frontend && npm run fmtcheck ) && pass "fmtcheck"

bold "[frontend] lint"
( cd frontend && npm run lint ) && pass "lint"

bold "[frontend] unit tests"
( cd frontend && npm run test ) && pass "unit"

# Frontend e2e needs a running frontend container — skip in CI gate.
bold "[frontend] e2e tests"
echo "    skipped — run manually: cd frontend && npm run test:e2e" >&2

echo
bold "all gates passed"