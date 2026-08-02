# ci-local.ps1 — single command that mirrors CI. Run this before pushing.
#
# Backend:   fmtcheck + lint + unit + e2e (requires docker compose up)
# Frontend:  fmtcheck + lint + unit
#
# Usage:  powershell -File scripts/ci-local.ps1
$ErrorActionPreference = 'Stop'
Set-Location -Path (Split-Path -Parent $PSScriptRoot)

function Bold($s)  { Write-Host "$s" -ForegroundColor Cyan }
function Pass($s)  { Write-Host "✓ $s" -ForegroundColor Green }
function Fail($s)  { Write-Host "✗ $s" -ForegroundColor Red; exit 1 }

# ----- Backend -------------------------------------------------------------
Bold "[backend] fmtcheck"
Push-Location backend
$bad = gofumpt -l . 2>&1 | Where-Object { $_ }
if ($bad) { Fail "gofumpt would reformat:`n$bad" }
Pass "gofumpt clean"

$bad = goimports -l -local github.com/admin-template/backend . 2>&1 | Where-Object { $_ }
if ($bad) { Fail "goimports would reformat:`n$bad" }
Pass "goimports clean"

Bold "[backend] lint"
& golangci-lint run ./...
if ($LASTEXITCODE -ne 0) { Fail "lint failed" }
Pass "lint clean"

Bold "[backend] unit tests"
& go test -shuffle=on -count=1 ./internal/...
if ($LASTEXITCODE -ne 0) { Fail "unit tests failed" }
Pass "unit"

Bold "[backend] e2e tests (requires docker compose up)"
& go test -tags=e2e -count=1 -timeout=120s ./e2e/...
if ($LASTEXITCODE -ne 0) { Fail "e2e failed" }
Pass "e2e"
Pop-Location

# ----- Frontend ------------------------------------------------------------
Bold "[frontend] fmtcheck"
Push-Location frontend
& npm run fmtcheck
if ($LASTEXITCODE -ne 0) { Fail "fmtcheck failed" }
Pass "fmtcheck"

Bold "[frontend] lint"
& npm run lint
if ($LASTEXITCODE -ne 0) { Fail "lint failed" }
Pass "lint"

Bold "[frontend] unit tests"
& npm run test
if ($LASTEXITCODE -ne 0) { Fail "unit tests failed" }
Pass "unit"

Bold "[frontend] e2e tests"
Write-Host "    skipped — run manually: cd frontend && npm run test:e2e" -ForegroundColor DarkGray
Pop-Location

Write-Host ""
Bold "all gates passed"